package discord

import (
	"fmt"
	"log/slog"

	"github.com/bwmarrin/discordgo"

	"github.com/ametis70/hellbot/internal/domain"
	"github.com/ametis70/hellbot/internal/port"
)

// Options holds configuration for a Discord notifier instance.
type Options struct {
	Token     string
	ChannelID string
	// GuildID is optional. When set, slash commands are registered as guild
	// commands (instant). When empty, they are registered globally (up to 1h delay).
	GuildID   string
	Templates *domain.Templates
}

// DiscordNotifier implements port.Notifier by sending messages to a Discord channel.
type DiscordNotifier struct {
	opts             Options
	session          *discordgo.Session
	logger           *slog.Logger
	templates        domain.Templates
	provider         port.StatusProvider
	registeredCmdIDs []string
}

// New creates a new DiscordNotifier, opens a Discord session, and validates connectivity.
func New(opts Options, logger *slog.Logger) (*DiscordNotifier, error) {
	if logger == nil {
		panic("discord notifier: logger is required")
	}
	if opts.Token == "" {
		return nil, fmt.Errorf("discord notifier: token is required")
	}
	if opts.ChannelID == "" {
		return nil, fmt.Errorf("discord notifier: channel_id is required")
	}

	session, err := discordgo.New("Bot " + opts.Token)
	if err != nil {
		return nil, fmt.Errorf("discord notifier: creating session: %w", err)
	}

	session.Identify.Intents = discordgo.IntentsGuildMessages

	if err := session.Open(); err != nil {
		return nil, fmt.Errorf("discord notifier: opening session: %w", err)
	}

	templates := DefaultTemplates()
	if opts.Templates != nil {
		templates = domain.MergeTemplates(templates, *opts.Templates)
	}

	return &DiscordNotifier{
		opts:      opts,
		session:   session,
		logger:    logger,
		templates: templates,
	}, nil
}

// RegisterCommands implements port.Commander. It registers /status and /statistics
// slash commands and wires the interaction handler.
func (n *DiscordNotifier) RegisterCommands(provider port.StatusProvider) {
	n.provider = provider

	factionChoice := &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionString,
		Name:        "faction",
		Description: "Filter by faction: bugs, cyborgs, illuminate",
		Required:    false,
		Choices: []*discordgo.ApplicationCommandOptionChoice{
			{Name: "Bugs", Value: "bugs"},
			{Name: "Cyborgs", Value: "cyborgs"},
			{Name: "Illuminate", Value: "illuminate"},
		},
	}

	cmds := []*discordgo.ApplicationCommand{
		{
			Name:        "status",
			Description: "Show current war progress and active events per faction",
			Options:     []*discordgo.ApplicationCommandOption{factionChoice},
		},
		{
			Name:        "statistics",
			Description: "Show cumulative war statistics (all factions summed)",
		},
	}

	appID := n.session.State.User.ID
	for _, cmd := range cmds {
		registered, err := n.session.ApplicationCommandCreate(appID, n.opts.GuildID, cmd)
		if err != nil {
			n.logger.Error("discord: failed to register slash command", "name", cmd.Name, "error", err)
			continue
		}
		n.registeredCmdIDs = append(n.registeredCmdIDs, registered.ID)
		n.logger.Info("discord: registered slash command", "name", cmd.Name)
	}

	n.session.AddHandler(n.handleInteraction)
}

func (n *DiscordNotifier) handleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	data := i.ApplicationCommandData()

	switch data.Name {
	case "status":
		n.handleStatusCommand(s, i, data)
	case "statistics":
		n.handleStatisticsCommand(s, i)
	}
}

func (n *DiscordNotifier) handleStatusCommand(s *discordgo.Session, i *discordgo.InteractionCreate, data discordgo.ApplicationCommandInteractionData) {
	var filter *domain.Enemy
	for _, opt := range data.Options {
		if opt.Name == "faction" {
			if enemy, ok := domain.ParseEnemy(opt.StringValue()); ok {
				e := enemy
				filter = &e
			}
		}
	}

	text, err := n.fetchAndFormatStatus(filter)
	if err != nil {
		text = "⚠️ Could not retrieve war status: " + err.Error()
	}

	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "```\n" + text + "\n```",
		},
	})
}

func (n *DiscordNotifier) handleStatisticsCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	text, err := n.fetchAndFormatStatistics()
	if err != nil {
		text = "⚠️ Could not retrieve statistics: " + err.Error()
	}

	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "```\n" + text + "\n```",
		},
	})
}

func (n *DiscordNotifier) fetchAndFormatStatus(filter *domain.Enemy) (string, error) {
	if n.provider == nil {
		return "", fmt.Errorf("no status provider registered")
	}
	c, err := n.provider.LatestCampaign()
	if err != nil {
		return "", err
	}
	return domain.FormatStatus(c, filter), nil
}

func (n *DiscordNotifier) fetchAndFormatStatistics() (string, error) {
	if n.provider == nil {
		return "", fmt.Errorf("no status provider registered")
	}
	c, err := n.provider.LatestCampaign()
	if err != nil {
		return "", err
	}
	return domain.FormatStatistics(c), nil
}

// Close deregisters slash commands and closes the underlying Discord session.
func (n *DiscordNotifier) Close() error {
	appID := n.session.State.User.ID
	for _, id := range n.registeredCmdIDs {
		if err := n.session.ApplicationCommandDelete(appID, n.opts.GuildID, id); err != nil {
			n.logger.Warn("discord: failed to deregister slash command", "id", id, "error", err)
		}
	}
	return n.session.Close()
}

// Notify sends a formatted event message to the configured Discord channel.
func (n *DiscordNotifier) Notify(msg domain.EventMessage) error {
	text, err := domain.RenderEvent(n.templates, msg, TimeFormatter(nil))
	if err != nil {
		return fmt.Errorf("discord notifier: rendering message: %w", err)
	}

	_, err = n.session.ChannelMessageSend(n.opts.ChannelID, text)
	if err != nil {
		return fmt.Errorf("discord notifier: sending message: %w", err)
	}

	return nil
}

// formatMessage is kept for tests — delegates to domain.RenderEvent with default templates.
func formatMessage(msg domain.EventMessage) (string, error) {
	return domain.RenderEvent(DefaultTemplates(), msg, TimeFormatter(nil))
}

func discordTimestamp(unix int64) string {
	return fmt.Sprintf("<t:%d:f>", unix)
}

