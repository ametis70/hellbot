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
}

// DiscordNotifier implements port.Notifier by sending messages to a Discord channel.
type DiscordNotifier struct {
	opts    Options
	session *discordgo.Session
	logger  *slog.Logger
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

	return &DiscordNotifier{
		opts:    opts,
		session: session,
		logger:  logger,
	}, nil
}

// Close closes the underlying Discord session.
func (n *DiscordNotifier) Close() error {
	return n.session.Close()
}

// Notify sends a formatted event message to the configured Discord channel.
func (n *DiscordNotifier) Notify(msg domain.EventMessage) error {
	text, err := formatMessage(msg)
	if err != nil {
		return err
	}

	_, err = n.session.ChannelMessageSend(n.opts.ChannelID, text)
	if err != nil {
		return fmt.Errorf("discord notifier: sending message: %w", err)
	}

	return nil
}

// formatMessage builds a Discord message string for an event.
// Times use Discord's native timestamp format <t:UNIX:f> which renders
// in the viewer's local timezone automatically.
func formatMessage(msg domain.EventMessage) (string, error) {
	switch msg.Kind {
	case domain.EventKindDefend:
		if msg.DefendEvent == nil {
			return "", fmt.Errorf("discord notifier: defend event is nil")
		}
		return formatDefendMessage(msg.Transition, msg.DefendEvent), nil
	case domain.EventKindAttack:
		if msg.AttackEvent == nil {
			return "", fmt.Errorf("discord notifier: attack event is nil")
		}
		return formatAttackMessage(msg.Transition, msg.AttackEvent), nil
	}
	return "", fmt.Errorf("discord notifier: unknown event kind: %s", msg.Kind)
}

func discordTimestamp(unix int64) string {
	return fmt.Sprintf("<t:%d:f>", unix)
}

func formatDefendMessage(transition domain.EventTransition, e *domain.DefendEvent) string {
	region := domain.GetRegion(e.Enemy, e.Region)

	switch transition {
	case domain.EventTransitionStarted:
		if domain.IsSuperEarth(e.Region) {
			return fmt.Sprintf(
				"🚨 **The %s is attacking Super Earth!**\nEnds: %s",
				e.Enemy,
				discordTimestamp(e.EndTime.Unix()),
			)
		}
		return fmt.Sprintf(
			"⚔️ **The %s is attacking %s (%d/%d)!**\nEnds: %s",
			e.Enemy,
			region.Name,
			e.Region,
			domain.TotalRegions,
			discordTimestamp(e.EndTime.Unix()),
		)
	case domain.EventTransitionSucceeded:
		if domain.IsSuperEarth(e.Region) {
			return fmt.Sprintf("✅ **Super Earth has been defended against the %s!**", e.Enemy)
		}
		return fmt.Sprintf("✅ **%s (%d/%d) has been defended against the %s!**", region.Name, e.Region, domain.TotalRegions, e.Enemy)
	case domain.EventTransitionFailed:
		if domain.IsSuperEarth(e.Region) {
			return fmt.Sprintf("❌ **Super Earth has fallen to the %s.**", e.Enemy)
		}
		return fmt.Sprintf("❌ **%s (%d/%d) has fallen to the %s.**", region.Name, e.Region, domain.TotalRegions, e.Enemy)
	}
	return fmt.Sprintf("[defend] %s — %s %s", transition, e.Enemy, region.Name)
}

func formatAttackMessage(transition domain.EventTransition, e *domain.AttackEvent) string {
	switch transition {
	case domain.EventTransitionStarted:
		return fmt.Sprintf(
			"🚀 **An attack against the %s's homeworld has started!**\nEnds: %s",
			e.Enemy,
			discordTimestamp(e.EndTime.Unix()),
		)
	case domain.EventTransitionSucceeded:
		return fmt.Sprintf("✅ **Attack succeeded! The %s were defeated.**", e.Enemy)
	case domain.EventTransitionFailed:
		return fmt.Sprintf("❌ **Attack failed! The %s defended their homeworld.**", e.Enemy)
	}
	return fmt.Sprintf("[attack] %s — %s", transition, e.Enemy)
}
