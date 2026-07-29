package app

import (
	"context"
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

// newFullPoller creates a Poller via the public New constructor.
func newFullPoller(fetcher port.Fetcher, notifier *testutil.MockNotifier) *Poller {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	store := memory.New()
	return New(fetcher, store, store, []port.Notifier{notifier}, time.Hour, logger)
}

// --- New / PollOnce / poll ---

func TestNew_CreatesPoller(t *testing.T) {
	notifier := &testutil.MockNotifier{}
	p := newFullPoller(&testutil.MockFetcher{}, notifier)
	if p == nil {
		t.Fatal("expected non-nil Poller from New")
	}
}

// PollOnce with a fetch error — no notification, no panic.
func TestPollOnce_FetchError(t *testing.T) {
	notifier := &testutil.MockNotifier{}
	fetcher := &testutil.MockFetcher{Err: errors.New("network error")}
	p := newFullPoller(fetcher, notifier)
	p.PollOnce() // must not panic
	if notifier.Count() != 0 {
		t.Errorf("expected 0 notifications on fetch error, got %d", notifier.Count())
	}
}

// PollOnce on first poll stores campaign but skips diffing.
func TestPollOnce_FirstPollNoDiff(t *testing.T) {
	notifier := &testutil.MockNotifier{}
	fetcher := &testutil.MockFetcher{Campaign: testutil.CampaignWithActiveAttack()}
	p := newFullPoller(fetcher, notifier)
	p.PollOnce()
	if notifier.Count() != 0 {
		t.Errorf("expected 0 notifications on first poll, got %d", notifier.Count())
	}
}

// PollOnce second poll detects a new attack event.
func TestPollOnce_SecondPollDetectsAttack(t *testing.T) {
	notifier := &testutil.MockNotifier{}
	campaign := testutil.CampaignWithActiveAttack()
	fetcher := &testutil.MockFetcher{Campaign: campaign}
	p := newFullPoller(fetcher, notifier)
	p.PollOnce() // first — baseline
	p.PollOnce() // second — diff
	if notifier.Count() != 1 {
		t.Fatalf("expected 1 notification on second poll, got %d", notifier.Count())
	}
	if notifier.First().Transition != domain.EventTransitionStarted {
		t.Errorf("expected started, got %s", notifier.First().Transition)
	}
}

// Run exits cleanly when context is cancelled.
func TestRun_CancelExits(t *testing.T) {
	notifier := &testutil.MockNotifier{}
	fetcher := &testutil.MockFetcher{Campaign: testutil.CampaignWithNoDefend()}
	p := newFullPoller(fetcher, notifier)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately
	if err := p.Run(ctx); err != nil {
		t.Errorf("expected nil error from Run after cancel, got %v", err)
	}
}

// handleEvents returns true when any sub-handler detects a change.
func TestHandleEvents_ReturnsTrueOnChange(t *testing.T) {
	notifier := &testutil.MockNotifier{}
	p := newTestPoller(notifier)
	current := testutil.CampaignWithActiveAttack()
	previous := testutil.CampaignWithNoDefend()
	changed := p.handleEvents(current, previous)
	if !changed {
		t.Error("expected handleEvents to return true when attack starts")
	}
}

// handleEvents returns false when nothing changed.
func TestHandleEvents_ReturnsFalseOnNoChange(t *testing.T) {
	notifier := &testutil.MockNotifier{}
	p := newTestPoller(notifier)
	current := testutil.CampaignWithNoDefend()
	previous := testutil.CampaignWithNoDefend()
	changed := p.handleEvents(current, previous)
	if changed {
		t.Error("expected handleEvents to return false when nothing changed")
	}
}

// notify logs errors from a failing notifier but does not propagate them.
func TestNotify_FailingNotifierLogged(t *testing.T) {
	failing := &failingNotifier{}
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	store := memory.New()
	p := &Poller{
		fetcher:   &testutil.MockFetcher{},
		campaigns: store,
		events:    store,
		notifiers: []port.Notifier{failing},
		logger:    logger,
	}
	// Must not panic.
	p.notify(domain.EventMessage{Kind: domain.EventKindWar, Transition: domain.EventTransitionSucceeded, WarEvent: &domain.WarEvent{Season: 1}})
	if failing.calls != 1 {
		t.Errorf("expected failing notifier to be called once, got %d", failing.calls)
	}
}

// --- handleWarEvents ---

func TestHandleWarEvents_NoPreviousFactions(t *testing.T) {
	notifier := &testutil.MockNotifier{}
	p := newTestPoller(notifier)
	current := testutil.CampaignWithNoDefend()
	previous := &domain.CampaignStatus{}
	result := p.handleWarEvents(current, previous)
	if result {
		t.Error("expected false when previous has no factions")
	}
	if notifier.Count() != 0 {
		t.Errorf("expected 0 notifications, got %d", notifier.Count())
	}
}

func TestHandleWarEvents_NoCurrentFactions(t *testing.T) {
	notifier := &testutil.MockNotifier{}
	p := newTestPoller(notifier)
	previous := testutil.CampaignWithNoDefend()
	current := &domain.CampaignStatus{}
	result := p.handleWarEvents(current, previous)
	if result {
		t.Error("expected false when current has no factions")
	}
}

func TestHandleWarEvents_SameSeason(t *testing.T) {
	notifier := &testutil.MockNotifier{}
	p := newTestPoller(notifier)
	current := testutil.CampaignWithNoDefend()
	previous := testutil.CampaignWithNoDefend()
	result := p.handleWarEvents(current, previous)
	if result {
		t.Error("expected false when season has not changed")
	}
	if notifier.Count() != 0 {
		t.Errorf("expected 0 notifications, got %d", notifier.Count())
	}
}

func TestHandleWarEvents_WarWon(t *testing.T) {
	notifier := &testutil.MockNotifier{}
	p := newTestPoller(notifier)

	// Previous season: all non-hidden factions defeated.
	previous := &domain.CampaignStatus{
		FactionsStatus: []domain.FactionStatus{
			{Season: 50, Enemy: domain.EnemyBug, Status: domain.FactionStatusDefeated},
			{Season: 50, Enemy: domain.EnemyCyborg, Status: domain.FactionStatusDefeated},
			{Season: 50, Enemy: domain.EnemyIlluminate, Status: domain.FactionStatusDefeated},
		},
	}
	// Current season: incremented.
	current := &domain.CampaignStatus{
		FactionsStatus: []domain.FactionStatus{
			{Season: 51, Enemy: domain.EnemyBug, Status: domain.FactionStatusActive},
		},
	}

	result := p.handleWarEvents(current, previous)
	if !result {
		t.Error("expected true when war ends")
	}
	if notifier.Count() != 1 {
		t.Fatalf("expected 1 notification, got %d", notifier.Count())
	}
	msg := notifier.First()
	if msg.Kind != domain.EventKindWar {
		t.Errorf("expected kind war, got %s", msg.Kind)
	}
	if msg.Transition != domain.EventTransitionSucceeded {
		t.Errorf("expected succeeded, got %s", msg.Transition)
	}
	if msg.WarEvent == nil || msg.WarEvent.Season != 50 {
		t.Errorf("expected war season 50, got %v", msg.WarEvent)
	}
}

func TestHandleWarEvents_WarLost(t *testing.T) {
	notifier := &testutil.MockNotifier{}
	p := newTestPoller(notifier)

	// Previous season: at least one non-hidden faction still active.
	previous := &domain.CampaignStatus{
		FactionsStatus: []domain.FactionStatus{
			{Season: 50, Enemy: domain.EnemyBug, Status: domain.FactionStatusDefeated},
			{Season: 50, Enemy: domain.EnemyCyborg, Status: domain.FactionStatusActive},
		},
	}
	current := &domain.CampaignStatus{
		FactionsStatus: []domain.FactionStatus{
			{Season: 51, Enemy: domain.EnemyBug, Status: domain.FactionStatusActive},
		},
	}

	result := p.handleWarEvents(current, previous)
	if !result {
		t.Error("expected true when war ends")
	}
	msg := notifier.First()
	if msg.Transition != domain.EventTransitionFailed {
		t.Errorf("expected failed (war lost), got %s", msg.Transition)
	}
}

func TestHandleWarEvents_HiddenFactionsIgnored(t *testing.T) {
	notifier := &testutil.MockNotifier{}
	p := newTestPoller(notifier)

	// Hidden factions should not prevent war won.
	previous := &domain.CampaignStatus{
		FactionsStatus: []domain.FactionStatus{
			{Season: 50, Enemy: domain.EnemyBug, Status: domain.FactionStatusDefeated},
			{Season: 50, Enemy: domain.EnemyCyborg, Status: domain.FactionStatusHidden},
		},
	}
	current := &domain.CampaignStatus{
		FactionsStatus: []domain.FactionStatus{
			{Season: 51, Enemy: domain.EnemyBug, Status: domain.FactionStatusActive},
		},
	}

	p.handleWarEvents(current, previous)
	msg := notifier.First()
	if msg.Transition != domain.EventTransitionSucceeded {
		t.Errorf("expected succeeded when only hidden factions remain, got %s", msg.Transition)
	}
}

// --- helpers ---

type failingNotifier struct{ calls int }

func (f *failingNotifier) Notify(_ domain.EventMessage) error {
	f.calls++
	return errors.New("notify error")
}
