package sqlite_test

import (
	"testing"

	"github.com/ametis70/hellbot/internal/adapter/store/sqlite"
	"github.com/ametis70/hellbot/internal/domain"
	"github.com/ametis70/hellbot/internal/testutil"
)

func newStore(t *testing.T) *sqlite.Store {
	t.Helper()
	s, err := sqlite.New(sqlite.Options{Path: ":memory:"})
	if err != nil {
		t.Fatalf("failed to open sqlite store: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

// --- CampaignStore ---

func TestSQLite_SaveAndGetCampaign(t *testing.T) {
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

func TestSQLite_LatestCampaign_Empty(t *testing.T) {
	s := newStore(t)
	_, err := s.LatestCampaign()
	if err == nil {
		t.Error("expected error when no campaign stored, got nil")
	}
}

func TestSQLite_SaveCampaign_Overwrites(t *testing.T) {
	s := newStore(t)
	_ = s.SaveCampaign(testutil.CampaignWithActiveDefend())
	c2 := testutil.CampaignWithNoDefend()
	_ = s.SaveCampaign(c2)
	got, _ := s.LatestCampaign()
	if got.DefendEvent != nil {
		t.Error("expected second save to overwrite first (no defend event)")
	}
}

// --- EventStore ---

func TestSQLite_SaveAndGetOngoingEvent(t *testing.T) {
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

func TestSQLite_GetOngoingEvent_NotFound(t *testing.T) {
	s := newStore(t)
	_, err := s.GetOngoingEvent(99, domain.EventKindAttack)
	if err == nil {
		t.Error("expected error for missing event, got nil")
	}
}

func TestSQLite_SaveOngoingEvent_Idempotent(t *testing.T) {
	s := newStore(t)
	_ = s.SaveOngoingEvent(1, domain.EventKindAttack)
	if err := s.SaveOngoingEvent(1, domain.EventKindAttack); err != nil {
		t.Errorf("expected idempotent save, got error: %v", err)
	}
}

func TestSQLite_RemoveOngoingEvent(t *testing.T) {
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

func TestSQLite_RemoveOngoingEvent_NotFound(t *testing.T) {
	s := newStore(t)
	err := s.RemoveOngoingEvent(999, domain.EventKindAttack)
	if err == nil {
		t.Error("expected error removing non-existent event, got nil")
	}
}

func TestSQLite_ListOngoingEvents(t *testing.T) {
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

func TestSQLite_ListOngoingEvents_Empty(t *testing.T) {
	s := newStore(t)
	events, err := s.ListOngoingEvents(domain.EventKindAttack)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("expected 0 events, got %d", len(events))
	}
}
