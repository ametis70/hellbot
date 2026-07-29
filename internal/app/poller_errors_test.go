package app

import (
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/ametis70/hellbot/internal/adapter/store/memory"
	"github.com/ametis70/hellbot/internal/domain"
	"github.com/ametis70/hellbot/internal/port"
	"github.com/ametis70/hellbot/internal/testutil"
)

// --- poll error paths ---

// poll: SaveCampaign fails — logged, no panic.
func TestPoll_SaveCampaignError(t *testing.T) {
	notifier := &testutil.MockNotifier{}
	store := &testutil.ErrorStore{}
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	campaign := testutil.CampaignWithNoDefend()
	p := New(&testutil.MockFetcher{Campaign: campaign}, store, store, []port.Notifier{notifier}, time.Hour, logger)
	p.PollOnce() // first poll: LatestCampaign fails → skip diff, SaveCampaign fails → logged
	// No panic, no notification expected.
	if notifier.Count() != 0 {
		t.Errorf("expected 0 notifications, got %d", notifier.Count())
	}
}

// poll: second poll with a store that now fails on LatestCampaign — skips diff.
func TestPoll_LatestCampaignErrorSkipsDiff(t *testing.T) {
	notifier := &testutil.MockNotifier{}
	// First poll succeeds (memory store), second poll uses error store to force LatestCampaign to fail.
	// Simulate by using a fresh ErrorStore for both campaigns and events.
	store := &testutil.ErrorStore{}
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	p := New(&testutil.MockFetcher{Campaign: testutil.CampaignWithActiveAttack()}, store, store, []port.Notifier{notifier}, time.Hour, logger)
	p.PollOnce()
	p.PollOnce()
	if notifier.Count() != 0 {
		t.Errorf("expected 0 notifications when store errors, got %d", notifier.Count())
	}
}

// --- handleDefendEvent store-error paths ---

// ListOngoingEvents fails → returns false, no notification.
func TestHandleDefendEvent_ListError(t *testing.T) {
	notifier := &testutil.MockNotifier{}
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	errStore := &testutil.ErrorStore{}
	memStore := memory.New()
	p := &Poller{
		fetcher:   &testutil.MockFetcher{},
		campaigns: memStore,
		events:    errStore,
		notifiers: []port.Notifier{notifier},
		logger:    logger,
	}
	result := p.handleDefendEvent(testutil.CampaignWithActiveDefend(), testutil.CampaignWithNoDefend())
	if result {
		t.Error("expected false when ListOngoingEvents errors")
	}
	if notifier.Count() != 0 {
		t.Errorf("expected 0 notifications, got %d", notifier.Count())
	}
}

// SaveOngoingEvent fails when new defend event arrives → returns false.
func TestHandleDefendEvent_SaveOngoingEventError(t *testing.T) {
	notifier := &testutil.MockNotifier{}
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	store := &saveFailStore{inner: memory.New()}
	p := &Poller{
		fetcher:   &testutil.MockFetcher{},
		campaigns: memory.New(),
		events:    store,
		notifiers: []port.Notifier{notifier},
		logger:    logger,
	}
	result := p.handleDefendEvent(testutil.CampaignWithActiveDefend(), testutil.CampaignWithNoDefend())
	if result {
		t.Error("expected false when SaveOngoingEvent errors")
	}
}

// RemoveOngoingEvent fails when defend event ends → returns false.
func TestHandleDefendEvent_RemoveOngoingEventError(t *testing.T) {
	notifier := &testutil.MockNotifier{}
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	store := &removeFailStore{}
	// Pre-populate with the event that will "end".
	store.inner = memory.New()
	_ = store.inner.SaveOngoingEvent(testutil.DefendEventActive().ID, domain.EventKindDefend)

	p := &Poller{
		fetcher:   &testutil.MockFetcher{},
		campaigns: memory.New(),
		events:    store,
		notifiers: []port.Notifier{notifier},
		logger:    logger,
	}
	result := p.handleDefendEvent(testutil.CampaignWithFailedDefend(), testutil.CampaignWithActiveDefend())
	if result {
		t.Error("expected false when RemoveOngoingEvent errors")
	}
}

// --- handleAttackEvents store-error paths ---

// ListOngoingEvents fails → returns false.
func TestHandleAttackEvents_ListError(t *testing.T) {
	notifier := &testutil.MockNotifier{}
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	p := &Poller{
		fetcher:   &testutil.MockFetcher{},
		campaigns: memory.New(),
		events:    &testutil.ErrorStore{},
		notifiers: []port.Notifier{notifier},
		logger:    logger,
	}
	result := p.handleAttackEvents(testutil.CampaignWithActiveAttack())
	if result {
		t.Error("expected false when ListOngoingEvents errors")
	}
}

// RemoveOngoingEvent fails for an ended attack → logs, continues.
func TestHandleAttackEvents_RemoveError(t *testing.T) {
	notifier := &testutil.MockNotifier{}
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	store := &removeFailStore{inner: memory.New()}
	ev := testutil.AttackEventActive()
	_ = store.inner.SaveOngoingEvent(ev.ID, domain.EventKindAttack)

	p := &Poller{
		fetcher:   &testutil.MockFetcher{},
		campaigns: memory.New(),
		events:    store,
		notifiers: []port.Notifier{notifier},
		logger:    logger,
	}
	// Attack ended (success) — remove will fail.
	p.handleAttackEvents(testutil.CampaignWithEndedAttack())
	// Should not panic, changed = false because remove failed.
}

// SaveOngoingEvent fails for a new attack → logs, continues.
func TestHandleAttackEvents_SaveError(t *testing.T) {
	notifier := &testutil.MockNotifier{}
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	p := &Poller{
		fetcher:   &testutil.MockFetcher{},
		campaigns: memory.New(),
		events:    &saveFailStore{inner: memory.New()},
		notifiers: []port.Notifier{notifier},
		logger:    logger,
	}
	p.handleAttackEvents(testutil.CampaignWithActiveAttack())
	// No panic, no notification (save failed so event not registered).
	if notifier.Count() != 0 {
		t.Errorf("expected 0 notifications when save fails, got %d", notifier.Count())
	}
}

// --- store error helpers ---

// saveFailStore delegates List/Get/Remove to inner, but fails Save operations.
type saveFailStore struct {
	inner interface {
		port.CampaignStore
		port.EventStore
	}
}

func (s *saveFailStore) SaveCampaign(c *domain.CampaignStatus) error {
	return s.inner.SaveCampaign(c)
}
func (s *saveFailStore) LatestCampaign() (*domain.CampaignStatus, error) {
	return s.inner.LatestCampaign()
}
func (s *saveFailStore) SaveOngoingEvent(_ int, _ domain.EventKind) error {
	return errStoreFailure
}
func (s *saveFailStore) RemoveOngoingEvent(id int, kind domain.EventKind) error {
	return s.inner.RemoveOngoingEvent(id, kind)
}
func (s *saveFailStore) GetOngoingEvent(id int, kind domain.EventKind) (*domain.OngoingEvent, error) {
	return s.inner.GetOngoingEvent(id, kind)
}
func (s *saveFailStore) ListOngoingEvents(kind domain.EventKind) ([]*domain.OngoingEvent, error) {
	return s.inner.ListOngoingEvents(kind)
}

// removeFailStore delegates List/Save/Get to inner, but fails Remove operations.
type removeFailStore struct {
	inner interface {
		port.CampaignStore
		port.EventStore
	}
}

func (s *removeFailStore) SaveCampaign(c *domain.CampaignStatus) error {
	return s.inner.SaveCampaign(c)
}
func (s *removeFailStore) LatestCampaign() (*domain.CampaignStatus, error) {
	return s.inner.LatestCampaign()
}
func (s *removeFailStore) SaveOngoingEvent(id int, kind domain.EventKind) error {
	return s.inner.SaveOngoingEvent(id, kind)
}
func (s *removeFailStore) RemoveOngoingEvent(_ int, _ domain.EventKind) error {
	return errStoreFailure
}
func (s *removeFailStore) GetOngoingEvent(id int, kind domain.EventKind) (*domain.OngoingEvent, error) {
	return s.inner.GetOngoingEvent(id, kind)
}
func (s *removeFailStore) ListOngoingEvents(kind domain.EventKind) ([]*domain.OngoingEvent, error) {
	return s.inner.ListOngoingEvents(kind)
}

var errStoreFailure = errors.New("store failure")
