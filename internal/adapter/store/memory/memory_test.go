package memory

import (
	"sync"
	"testing"

	"github.com/ametis70/hellbot/internal/domain"
	"github.com/ametis70/hellbot/internal/testutil"
)

// --- CampaignStore tests ---

func TestLatestCampaign_EmptyStore(t *testing.T) {
	s := New()
	_, err := s.LatestCampaign()
	if err == nil {
		t.Error("expected error on empty store, got nil")
	}
}

func TestSaveAndLatestCampaign(t *testing.T) {
	s := New()
	campaign := testutil.CampaignWithActiveDefend()

	if err := s.SaveCampaign(campaign); err != nil {
		t.Fatalf("SaveCampaign returned unexpected error: %v", err)
	}

	got, err := s.LatestCampaign()
	if err != nil {
		t.Fatalf("LatestCampaign returned unexpected error: %v", err)
	}
	if got != campaign {
		t.Error("LatestCampaign did not return the saved campaign")
	}
}

func TestSaveCampaign_OverwritesPrevious(t *testing.T) {
	s := New()
	first := testutil.CampaignWithActiveDefend()
	second := testutil.CampaignWithFailedDefend()

	_ = s.SaveCampaign(first)
	_ = s.SaveCampaign(second)

	got, err := s.LatestCampaign()
	if err != nil {
		t.Fatalf("LatestCampaign returned unexpected error: %v", err)
	}
	if got != second {
		t.Error("LatestCampaign did not return the most recent campaign")
	}
}

// --- EventStore tests ---

func TestGetOngoingEvent_NotFound(t *testing.T) {
	s := New()
	_, err := s.GetOngoingEvent(1, domain.EventKindDefend)
	if err == nil {
		t.Error("expected error for missing event, got nil")
	}
}

func TestSaveAndGetOngoingEvent(t *testing.T) {
	s := New()

	if err := s.SaveOngoingEvent(42, domain.EventKindDefend); err != nil {
		t.Fatalf("SaveOngoingEvent returned unexpected error: %v", err)
	}

	got, err := s.GetOngoingEvent(42, domain.EventKindDefend)
	if err != nil {
		t.Fatalf("GetOngoingEvent returned unexpected error: %v", err)
	}
	if got.ID != 42 {
		t.Errorf("expected ID 42, got %d", got.ID)
	}
	if got.Kind != domain.EventKindDefend {
		t.Errorf("expected kind %s, got %s", domain.EventKindDefend, got.Kind)
	}
}

func TestRemoveOngoingEvent(t *testing.T) {
	s := New()
	_ = s.SaveOngoingEvent(42, domain.EventKindDefend)

	if err := s.RemoveOngoingEvent(42, domain.EventKindDefend); err != nil {
		t.Fatalf("RemoveOngoingEvent returned unexpected error: %v", err)
	}

	_, err := s.GetOngoingEvent(42, domain.EventKindDefend)
	if err == nil {
		t.Error("expected error after removal, got nil")
	}
}

func TestRemoveOngoingEvent_NotFound(t *testing.T) {
	s := New()
	err := s.RemoveOngoingEvent(99, domain.EventKindDefend)
	if err == nil {
		t.Error("expected error when removing non-existent event, got nil")
	}
}

func TestListOngoingEvents_FiltersByKind(t *testing.T) {
	s := New()
	_ = s.SaveOngoingEvent(1, domain.EventKindDefend)
	_ = s.SaveOngoingEvent(2, domain.EventKindAttack)
	_ = s.SaveOngoingEvent(3, domain.EventKindAttack)

	defends, err := s.ListOngoingEvents(domain.EventKindDefend)
	if err != nil {
		t.Fatalf("ListOngoingEvents returned unexpected error: %v", err)
	}
	if len(defends) != 1 {
		t.Errorf("expected 1 defend event, got %d", len(defends))
	}

	attacks, err := s.ListOngoingEvents(domain.EventKindAttack)
	if err != nil {
		t.Fatalf("ListOngoingEvents returned unexpected error: %v", err)
	}
	if len(attacks) != 2 {
		t.Errorf("expected 2 attack events, got %d", len(attacks))
	}
}

func TestListOngoingEvents_EmptyStore(t *testing.T) {
	s := New()
	events, err := s.ListOngoingEvents(domain.EventKindDefend)
	if err != nil {
		t.Fatalf("ListOngoingEvents returned unexpected error: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("expected 0 events, got %d", len(events))
	}
}

// --- Concurrent access tests ---

func TestConcurrentSaveCampaign(t *testing.T) {
	s := New()
	var wg sync.WaitGroup
	n := 100

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = s.SaveCampaign(testutil.CampaignWithActiveDefend())
		}()
	}
	wg.Wait()

	_, err := s.LatestCampaign()
	if err != nil {
		t.Errorf("expected campaign after concurrent writes, got error: %v", err)
	}
}

func TestConcurrentSaveAndGetOngoingEvent(t *testing.T) {
	s := New()
	var wg sync.WaitGroup
	n := 100

	for i := 0; i < n; i++ {
		wg.Add(1)
		id := i
		go func() {
			defer wg.Done()
			_ = s.SaveOngoingEvent(id, domain.EventKindAttack)
		}()
	}
	wg.Wait()

	events, err := s.ListOngoingEvents(domain.EventKindAttack)
	if err != nil {
		t.Fatalf("ListOngoingEvents returned unexpected error: %v", err)
	}
	if len(events) != n {
		t.Errorf("expected %d events, got %d", n, len(events))
	}
}
