package valkey_test

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"github.com/ametis70/hellbot/internal/adapter/store/valkey"
	"github.com/ametis70/hellbot/internal/domain"
	"github.com/ametis70/hellbot/internal/testutil"
)

// newStore starts a miniredis server and returns a Store backed by it.
// The server and store are both closed automatically when the test ends.
func newStore(t *testing.T) *valkey.Store {
	t.Helper()
	mr := miniredis.RunT(t)
	s, err := valkey.New(valkey.Options{Addr: mr.Addr()})
	if err != nil {
		t.Fatalf("failed to create valkey store: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

// TestValkey_New_PingFails verifies that New returns an error when the server
// is unreachable.
func TestValkey_New_PingFails(t *testing.T) {
	_, err := valkey.New(valkey.Options{Addr: "localhost:1"})
	if err == nil {
		t.Error("expected error for unreachable server, got nil")
	}
}

// --- CampaignStore ---

func TestValkey_SaveAndGetCampaign(t *testing.T) {
	s := newStore(t)
	c := testutil.CampaignWithActiveDefend()
	if err := s.SaveCampaign(c); err != nil {
		t.Fatalf("SaveCampaign: %v", err)
	}
	got, err := s.LatestCampaign()
	if err != nil {
		t.Fatalf("LatestCampaign: %v", err)
	}
	if got.DefendEvent == nil {
		t.Fatal("expected defend event to be persisted")
	}
	if got.DefendEvent.ID != c.DefendEvent.ID {
		t.Errorf("expected defend ID %d, got %d", c.DefendEvent.ID, got.DefendEvent.ID)
	}
}

func TestValkey_LatestCampaign_Empty(t *testing.T) {
	s := newStore(t)
	_, err := s.LatestCampaign()
	if err == nil {
		t.Error("expected error when no campaign stored, got nil")
	}
}

func TestValkey_SaveCampaign_Overwrites(t *testing.T) {
	s := newStore(t)
	_ = s.SaveCampaign(testutil.CampaignWithActiveDefend())
	_ = s.SaveCampaign(testutil.CampaignWithNoDefend())
	got, err := s.LatestCampaign()
	if err != nil {
		t.Fatalf("LatestCampaign: %v", err)
	}
	if got.DefendEvent != nil {
		t.Error("expected second save to overwrite first (no defend event)")
	}
}

// --- EventStore ---

func TestValkey_SaveAndGetOngoingEvent(t *testing.T) {
	s := newStore(t)
	if err := s.SaveOngoingEvent(42, domain.EventKindDefend); err != nil {
		t.Fatalf("SaveOngoingEvent: %v", err)
	}
	got, err := s.GetOngoingEvent(42, domain.EventKindDefend)
	if err != nil {
		t.Fatalf("GetOngoingEvent: %v", err)
	}
	if got.ID != 42 || got.Kind != domain.EventKindDefend {
		t.Errorf("unexpected event: %+v", got)
	}
}

func TestValkey_GetOngoingEvent_NotFound(t *testing.T) {
	s := newStore(t)
	_, err := s.GetOngoingEvent(99, domain.EventKindAttack)
	if err == nil {
		t.Error("expected error for missing event, got nil")
	}
}

func TestValkey_SaveOngoingEvent_Idempotent(t *testing.T) {
	s := newStore(t)
	_ = s.SaveOngoingEvent(1, domain.EventKindAttack)
	if err := s.SaveOngoingEvent(1, domain.EventKindAttack); err != nil {
		t.Errorf("expected idempotent save, got error: %v", err)
	}
}

func TestValkey_RemoveOngoingEvent(t *testing.T) {
	s := newStore(t)
	_ = s.SaveOngoingEvent(7, domain.EventKindDefend)
	if err := s.RemoveOngoingEvent(7, domain.EventKindDefend); err != nil {
		t.Fatalf("RemoveOngoingEvent: %v", err)
	}
	_, err := s.GetOngoingEvent(7, domain.EventKindDefend)
	if err == nil {
		t.Error("expected error after removal, got nil")
	}
}

func TestValkey_RemoveOngoingEvent_NotFound(t *testing.T) {
	s := newStore(t)
	err := s.RemoveOngoingEvent(999, domain.EventKindAttack)
	if err == nil {
		t.Error("expected error removing non-existent event, got nil")
	}
}

func TestValkey_ListOngoingEvents(t *testing.T) {
	s := newStore(t)
	_ = s.SaveOngoingEvent(1, domain.EventKindAttack)
	_ = s.SaveOngoingEvent(2, domain.EventKindAttack)
	_ = s.SaveOngoingEvent(3, domain.EventKindDefend)

	attacks, err := s.ListOngoingEvents(domain.EventKindAttack)
	if err != nil {
		t.Fatalf("ListOngoingEvents: %v", err)
	}
	if len(attacks) != 2 {
		t.Errorf("expected 2 attack events, got %d", len(attacks))
	}

	defends, err := s.ListOngoingEvents(domain.EventKindDefend)
	if err != nil {
		t.Fatalf("ListOngoingEvents: %v", err)
	}
	if len(defends) != 1 {
		t.Errorf("expected 1 defend event, got %d", len(defends))
	}
}

func TestValkey_ListOngoingEvents_Empty(t *testing.T) {
	s := newStore(t)
	events, err := s.ListOngoingEvents(domain.EventKindAttack)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("expected 0 events, got %d", len(events))
	}
}

// TestValkey_ListOngoingEvents_StaleKeyCleanup verifies that keys present in
// the events set but missing from Redis (e.g. expired) are silently skipped
// and removed from the index.
func TestValkey_ListOngoingEvents_StaleKeyCleanup(t *testing.T) {
	mr := miniredis.RunT(t)
	s, err := valkey.New(valkey.Options{Addr: mr.Addr()})
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	_ = s.SaveOngoingEvent(10, domain.EventKindAttack)

	// Directly delete the event key from miniredis, leaving the set entry stale.
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	client.Del(t.Context(), "hellbot:event:10:attack")

	// ListOngoingEvents should skip the stale key without error.
	events, err := s.ListOngoingEvents(domain.EventKindAttack)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("expected 0 events after stale key skipped, got %d", len(events))
	}
}
