package valkey

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ametis70/hellbot/internal/domain"
	"github.com/redis/go-redis/v9"
)

const (
	campaignKey    = "hellbot:campaign"
	eventsSetKey   = "hellbot:events"
	eventKeyPrefix = "hellbot:event:"
)

// Store implements port.CampaignStore and port.EventStore using a Redis/Valkey backend.
type Store struct {
	client *redis.Client
}

// Options holds the connection parameters for the Valkey/Redis store.
type Options struct {
	// Addr is the host:port of the Redis/Valkey server (e.g. "localhost:6379").
	Addr string
	// Password is optional; leave empty if the server has no auth.
	Password string
	// DB is the Redis database index (0–15).
	DB int
}

// New creates a Store and verifies connectivity with a PING.
func New(opts Options) (*Store, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     opts.Addr,
		Password: opts.Password,
		DB:       opts.DB,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("valkey: ping failed: %w", err)
	}

	return &Store{client: client}, nil
}

// Close releases the underlying connection pool.
func (s *Store) Close() error {
	return s.client.Close()
}

// ── CampaignStore ────────────────────────────────────────────────────────────

func (s *Store) SaveCampaign(c *domain.CampaignStatus) error {
	data, err := json.Marshal(c)
	if err != nil {
		return fmt.Errorf("valkey: marshal campaign: %w", err)
	}
	if err := s.client.Set(context.Background(), campaignKey, data, 0).Err(); err != nil {
		return fmt.Errorf("valkey: save campaign: %w", err)
	}
	return nil
}

func (s *Store) LatestCampaign() (*domain.CampaignStatus, error) {
	data, err := s.client.Get(context.Background(), campaignKey).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, errors.New("no campaign stored")
	}
	if err != nil {
		return nil, fmt.Errorf("valkey: get campaign: %w", err)
	}

	var c domain.CampaignStatus
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("valkey: unmarshal campaign: %w", err)
	}
	return &c, nil
}

// ── EventStore ───────────────────────────────────────────────────────────────

func eventKey(id int, kind domain.EventKind) string {
	return fmt.Sprintf("%s%d:%s", eventKeyPrefix, id, kind)
}

func (s *Store) SaveOngoingEvent(id int, kind domain.EventKind) error {
	ev := &domain.OngoingEvent{ID: id, Kind: kind}
	data, err := json.Marshal(ev)
	if err != nil {
		return fmt.Errorf("valkey: marshal event: %w", err)
	}
	ctx := context.Background()
	key := eventKey(id, kind)
	if err := s.client.Set(ctx, key, data, 0).Err(); err != nil {
		return fmt.Errorf("valkey: save event: %w", err)
	}
	if err := s.client.SAdd(ctx, eventsSetKey, key).Err(); err != nil {
		return fmt.Errorf("valkey: index event: %w", err)
	}
	return nil
}

func (s *Store) RemoveOngoingEvent(id int, kind domain.EventKind) error {
	ctx := context.Background()
	key := eventKey(id, kind)
	deleted, err := s.client.Del(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("valkey: delete event: %w", err)
	}
	if deleted == 0 {
		return errors.New("event not found")
	}
	if err := s.client.SRem(ctx, eventsSetKey, key).Err(); err != nil {
		return fmt.Errorf("valkey: deindex event: %w", err)
	}
	return nil
}

func (s *Store) GetOngoingEvent(id int, kind domain.EventKind) (*domain.OngoingEvent, error) {
	data, err := s.client.Get(context.Background(), eventKey(id, kind)).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, errors.New("event not found")
	}
	if err != nil {
		return nil, fmt.Errorf("valkey: get event: %w", err)
	}
	var ev domain.OngoingEvent
	if err := json.Unmarshal(data, &ev); err != nil {
		return nil, fmt.Errorf("valkey: unmarshal event: %w", err)
	}
	return &ev, nil
}

func (s *Store) ListOngoingEvents(kind domain.EventKind) ([]*domain.OngoingEvent, error) {
	ctx := context.Background()
	keys, err := s.client.SMembers(ctx, eventsSetKey).Result()
	if err != nil {
		return nil, fmt.Errorf("valkey: list event keys: %w", err)
	}

	result := make([]*domain.OngoingEvent, 0)
	for _, key := range keys {
		data, err := s.client.Get(ctx, key).Bytes()
		if errors.Is(err, redis.Nil) {
			// Key expired or deleted; clean up the index entry.
			_ = s.client.SRem(ctx, eventsSetKey, key)
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("valkey: get event %s: %w", key, err)
		}
		var ev domain.OngoingEvent
		if err := json.Unmarshal(data, &ev); err != nil {
			return nil, fmt.Errorf("valkey: unmarshal event %s: %w", key, err)
		}
		if ev.Kind == kind {
			result = append(result, &ev)
		}
	}
	return result, nil
}
