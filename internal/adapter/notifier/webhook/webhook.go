package webhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/ametis70/hellbot/internal/domain"
)

// Options configures the webhook notifier.
type Options struct {
	// URL is the endpoint that will receive POST requests.
	URL string
	// SecretHeader is the name of the HTTP header used for authentication
	// (e.g. "Authorization"). Leave empty to send no auth header.
	SecretHeader string
	// SecretValue is the value of the auth header (e.g. "Bearer <token>").
	SecretValue string
	// Timeout is the HTTP client timeout. Defaults to 10s.
	Timeout time.Duration
}

// Notifier implements port.Notifier by POSTing a JSON payload to a configured URL.
type Notifier struct {
	opts   Options
	client *http.Client
	logger *slog.Logger
}

// New creates a Notifier. opts.URL must be non-empty.
func New(opts Options, logger *slog.Logger) (*Notifier, error) {
	if opts.URL == "" {
		return nil, fmt.Errorf("webhook: url is required")
	}
	if opts.Timeout == 0 {
		opts.Timeout = 10 * time.Second
	}
	return &Notifier{
		opts:   opts,
		client: &http.Client{Timeout: opts.Timeout},
		logger: logger,
	}, nil
}

// ── JSON payload types ────────────────────────────────────────────────────────

// Payload is the JSON body sent to the webhook endpoint for every event.
type Payload struct {
	Kind        string       `json:"kind"`
	Transition  string       `json:"transition"`
	DefendEvent *DefendEvent `json:"defend_event,omitempty"`
	AttackEvent *AttackEvent `json:"attack_event,omitempty"`
	WarEvent    *WarEvent    `json:"war_event,omitempty"`
}

type DefendEvent struct {
	Season         int    `json:"season"`
	ID             int    `json:"id"`
	Region         int    `json:"region"`
	RegionName     string `json:"region_name"`
	Enemy          string `json:"enemy"`
	PointsMax      int    `json:"points_max"`
	Points         int    `json:"points"`
	Status         string `json:"status"`
	PlayersAtStart int    `json:"players_at_start"`
	StartTime      string `json:"start_time"`
	EndTime        string `json:"end_time"`
	StartTimeUnix  int64  `json:"start_time_unix"`
	EndTimeUnix    int64  `json:"end_time_unix"`
}

type AttackEvent struct {
	Season         int    `json:"season"`
	ID             int    `json:"id"`
	Enemy          string `json:"enemy"`
	PointsMax      int    `json:"points_max"`
	Points         int    `json:"points"`
	Status         string `json:"status"`
	PlayersAtStart int    `json:"players_at_start"`
	StartTime      string `json:"start_time"`
	EndTime        string `json:"end_time"`
	StartTimeUnix  int64  `json:"start_time_unix"`
	EndTimeUnix    int64  `json:"end_time_unix"`
}

type WarEvent struct {
	Season int `json:"season"`
}

// ── domain → payload mappers ─────────────────────────────────────────────────

func toDefendEvent(e *domain.DefendEvent) *DefendEvent {
	return &DefendEvent{
		Season:         e.Season,
		ID:             e.ID,
		Region:         e.Region,
		RegionName:     domain.GetRegion(e.Enemy, e.Region).Name,
		Enemy:          e.Enemy.String(),
		PointsMax:      e.PointsMax,
		Points:         e.Points,
		Status:         string(e.Status),
		PlayersAtStart: e.PlayersAtStart,
		StartTime:      e.StartTime.UTC().Format(time.RFC3339),
		EndTime:        e.EndTime.UTC().Format(time.RFC3339),
		StartTimeUnix:  e.StartTime.Unix(),
		EndTimeUnix:    e.EndTime.Unix(),
	}
}

func toAttackEvent(e *domain.AttackEvent) *AttackEvent {
	return &AttackEvent{
		Season:         e.Season,
		ID:             e.ID,
		Enemy:          e.Enemy.String(),
		PointsMax:      e.PointsMax,
		Points:         e.Points,
		Status:         string(e.Status),
		PlayersAtStart: e.PlayersAtStart,
		StartTime:      e.StartTime.UTC().Format(time.RFC3339),
		EndTime:        e.EndTime.UTC().Format(time.RFC3339),
		StartTimeUnix:  e.StartTime.Unix(),
		EndTimeUnix:    e.EndTime.Unix(),
	}
}

func buildPayload(msg domain.EventMessage) Payload {
	p := Payload{
		Kind:       string(msg.Kind),
		Transition: string(msg.Transition),
	}
	if msg.DefendEvent != nil {
		p.DefendEvent = toDefendEvent(msg.DefendEvent)
	}
	if msg.AttackEvent != nil {
		p.AttackEvent = toAttackEvent(msg.AttackEvent)
	}
	if msg.WarEvent != nil {
		p.WarEvent = &WarEvent{Season: msg.WarEvent.Season}
	}
	return p
}

// ── Notifier ─────────────────────────────────────────────────────────────────

func (n *Notifier) Notify(msg domain.EventMessage) error {
	payload := buildPayload(msg)

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("webhook: marshal payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, n.opts.URL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("webhook: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if n.opts.SecretHeader != "" {
		req.Header.Set(n.opts.SecretHeader, n.opts.SecretValue)
	}

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("webhook: post to %s: %w", n.opts.URL, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook: server returned %d", resp.StatusCode)
	}

	n.logger.Info("webhook: notification sent", "url", n.opts.URL, "kind", msg.Kind, "transition", msg.Transition)
	return nil
}
