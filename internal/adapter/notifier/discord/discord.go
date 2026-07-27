package discord

import (
	"fmt"
	"log/slog"

	"github.com/bwmarrin/discordgo"

	"github.com/ametis70/hellbot/internal/domain"
)

// Options holds configuration for a Discord notifier instance.
type Options struct {
	Token     string
	ChannelID string
	Templates *domain.Templates
}

// DiscordNotifier implements port.Notifier by sending messages to a Discord channel.
type DiscordNotifier struct {
	opts      Options
	session   *discordgo.Session
	logger    *slog.Logger
	templates domain.Templates
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

// Close closes the underlying Discord session.
func (n *DiscordNotifier) Close() error {
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
