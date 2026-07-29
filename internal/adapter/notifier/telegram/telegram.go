package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/ametis70/hellbot/internal/domain"
	"github.com/ametis70/hellbot/internal/port"
)

const apiBase = "https://api.telegram.org"

// Options holds configuration for a Telegram notifier instance.
type Options struct {
	Token     string
	ChatID    string
	Timezone  *time.Location
	Templates *domain.Templates
	// APIBase overrides the Telegram Bot API base URL. Defaults to
	// "https://api.telegram.org". Intended for use in tests.
	APIBase string
}

// update is a partial Telegram Update object used for command polling.
type update struct {
	UpdateID int `json:"update_id"`
	Message  *struct {
		Text     string `json:"text"`
		Entities []struct {
			Type   string `json:"type"`
			Offset int    `json:"offset"`
			Length int    `json:"length"`
		} `json:"entities"`
	} `json:"message"`
}

// Notifier implements port.Notifier by sending messages to a Telegram chat.
// It also polls for bot commands and handles /test, /status, and /statistics.
type Notifier struct {
	opts      Options
	client    *http.Client
	logger    *slog.Logger
	templates domain.Templates
	cancel    context.CancelFunc
	done      chan struct{}
	provider  port.StatusProvider
	apiBase   string
}

// New creates a new Notifier, validates options, and starts the command polling loop.
func New(opts Options, logger *slog.Logger) (*Notifier, error) {
	if logger == nil {
		panic("telegram notifier: logger is required")
	}
	if opts.Token == "" {
		return nil, fmt.Errorf("telegram notifier: token is required")
	}
	if opts.ChatID == "" {
		return nil, fmt.Errorf("telegram notifier: chat_id is required")
	}

	if opts.Timezone == nil {
		opts.Timezone = time.UTC
	}

	templates := DefaultTemplates()
	if opts.Templates != nil {
		templates = domain.MergeTemplates(templates, *opts.Templates)
	}

	ctx, cancel := context.WithCancel(context.Background())

	n := &Notifier{
		opts:      opts,
		client:    &http.Client{Timeout: 10 * time.Second},
		logger:    logger,
		templates: templates,
		cancel:    cancel,
		done:      make(chan struct{}),
		apiBase:   apiBase,
	}
	if opts.APIBase != "" {
		n.apiBase = opts.APIBase
	}

	go n.pollCommands(ctx)

	return n, nil
}

// Close stops the command polling loop and waits for it to exit.
func (n *Notifier) Close() error {
	n.cancel()
	<-n.done
	return nil
}

// Notify sends a formatted event message to the configured Telegram chat.
func (n *Notifier) Notify(msg domain.EventMessage) error {
	text, err := domain.RenderEvent(n.templates, msg, TimeFormatter(n.opts.Timezone))
	if err != nil {
		return fmt.Errorf("telegram notifier: rendering message: %w", err)
	}

	return n.sendMessage(text)
}

// pollCommands long-polls getUpdates and dispatches recognised bot commands.
func (n *Notifier) pollCommands(ctx context.Context) {
	defer close(n.done)

	offset := 0
	for {
		updates, err := n.getUpdates(ctx, offset)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			n.logger.Error("telegram notifier: getUpdates failed", "error", err)
			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Second):
				continue
			}
		}

		for _, u := range updates {
			offset = u.UpdateID + 1
			n.handleUpdate(u)
		}
	}
}

// getUpdates calls the Telegram getUpdates API with long-polling (timeout=30s).
func (n *Notifier) getUpdates(ctx context.Context, offset int) ([]update, error) {
	type params struct {
		Offset  int `json:"offset"`
		Timeout int `json:"timeout"`
	}

	body, err := json.Marshal(params{Offset: offset, Timeout: 30})
	if err != nil {
		return nil, fmt.Errorf("marshaling getUpdates params: %w", err)
	}

	// HTTP client timeout must exceed the long-poll timeout.
	longPollClient := &http.Client{Timeout: 35 * time.Second}

	url := fmt.Sprintf("%s/bot%s/getUpdates", n.apiBase, n.opts.Token)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("building getUpdates request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := longPollClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("getUpdates request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			n.logger.Error("telegram notifier: closing getUpdates response body", "error", err)
		}
	}()

	var result struct {
		OK     bool     `json:"ok"`
		Result []update `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding getUpdates response: %w", err)
	}
	if !result.OK {
		return nil, fmt.Errorf("getUpdates returned ok=false")
	}

	return result.Result, nil
}

// handleUpdate dispatches a single update to the appropriate command handler.
func (n *Notifier) handleUpdate(u update) {
	if u.Message == nil {
		return
	}

	for _, entity := range u.Message.Entities {
		if entity.Type != "bot_command" {
			continue
		}

		raw := u.Message.Text[entity.Offset : entity.Offset+entity.Length]
		// Strip @botname suffix present in group chats.
		cmd := strings.SplitN(raw, "@", 2)[0]

		// Argument is the text after the command entity.
		arg := strings.TrimSpace(u.Message.Text[entity.Offset+entity.Length:])

		switch cmd {
		case "/test":
			n.handleTestCommand()
		case "/status":
			n.handleStatusCommand(arg)
		case "/statistics":
			n.handleStatisticsCommand()
		}
	}
}

// RegisterCommands implements port.Commander.
func (n *Notifier) RegisterCommands(provider port.StatusProvider) {
	n.provider = provider
}

// handleTestCommand sends a test message to the configured chat_id.
func (n *Notifier) handleTestCommand() {
	n.logger.Info("telegram notifier: /test command received, sending test message")
	err := n.sendMessage("✅ hellbot is connected and can send messages to this chat\\.")
	if err != nil {
		n.logger.Error("telegram notifier: /test failed", "error", err)
	}
}

// handleStatusCommand responds to /status [faction].
func (n *Notifier) handleStatusCommand(arg string) {
	if n.provider == nil {
		n.logger.Warn("telegram notifier: /status received but no status provider registered")
		return
	}

	var filter *domain.Enemy
	if arg != "" {
		if enemy, ok := domain.ParseEnemy(arg); ok {
			e := enemy
			filter = &e
		}
	}

	c, err := n.provider.LatestCampaign()
	if err != nil {
		n.logger.Error("telegram notifier: /status failed to fetch campaign", "error", err)
		_ = n.sendMessage("⚠️ Could not retrieve war status\\.")
		return
	}

	text := escape(domain.FormatStatus(c, filter))
	if sendErr := n.sendMessage("```\n" + text + "\n```"); sendErr != nil {
		n.logger.Error("telegram notifier: /status failed to send", "error", sendErr)
	}
}

// handleStatisticsCommand responds to /statistics.
func (n *Notifier) handleStatisticsCommand() {
	if n.provider == nil {
		n.logger.Warn("telegram notifier: /statistics received but no status provider registered")
		return
	}

	c, err := n.provider.LatestCampaign()
	if err != nil {
		n.logger.Error("telegram notifier: /statistics failed to fetch campaign", "error", err)
		_ = n.sendMessage("⚠️ Could not retrieve statistics\\.")
		return
	}

	text := escape(domain.FormatStatistics(c))
	if sendErr := n.sendMessage("```\n" + text + "\n```"); sendErr != nil {
		n.logger.Error("telegram notifier: /statistics failed to send", "error", sendErr)
	}
}

// sendMessage calls the Telegram sendMessage API with MarkdownV2 parse mode.
func (n *Notifier) sendMessage(text string) error {
	type payload struct {
		ChatID    string `json:"chat_id"`
		Text      string `json:"text"`
		ParseMode string `json:"parse_mode"`
	}

	body, err := json.Marshal(payload{
		ChatID:    n.opts.ChatID,
		Text:      text,
		ParseMode: "MarkdownV2",
	})
	if err != nil {
		return fmt.Errorf("telegram notifier: marshaling payload: %w", err)
	}

	url := fmt.Sprintf("%s/bot%s/sendMessage", n.apiBase, n.opts.Token)
	resp, err := n.client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("telegram notifier: sending message: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			n.logger.Error("telegram notifier: closing sendMessage response body", "error", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram notifier: unexpected status %d", resp.StatusCode)
	}

	return nil
}

// escape escapes special MarkdownV2 characters in plain text.
func escape(s string) string {
	specials := `\_*[]()~` + "`" + `>#+-=|{}.!`
	out := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		for j := 0; j < len(specials); j++ {
			if s[i] == specials[j] {
				out = append(out, '\\')
				break
			}
		}
		out = append(out, s[i])
	}
	return string(out)
}
