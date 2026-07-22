package app

import (
	"log/slog"
	"os"
	"testing"

	"github.com/ametis70/hellbot/internal/adapter/store/memory"
	"github.com/ametis70/hellbot/internal/domain"
	"github.com/ametis70/hellbot/internal/port"
	"github.com/ametis70/hellbot/internal/testutil"
)

func newTestPoller(notifier *testutil.MockNotifier) *Poller {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	store := memory.New()
	return &Poller{
		fetcher:   &testutil.MockFetcher{},
		campaigns: store,
		events:    store,
		notifiers: []port.Notifier{notifier},
		logger:    logger,
	}
}

// --- handleDefendEvent tests ---

// Case 1: no defend event anywhere — no notification
func TestHandleDefendEvent_NoneAnywhere(t *testing.T) {
	notifier := &testutil.MockNotifier{}
	p := newTestPoller(notifier)

	current := testutil.CampaignWithNoDefend()
	previous := testutil.CampaignWithNoDefend()

	p.handleDefendEvent(current, previous)

	if notifier.Count() != 0 {
		t.Errorf("expected 0 notifications, got %d", notifier.Count())
	}
}

// Case 2: new active defend event, nothing stored — notify started
func TestHandleDefendEvent_NewActive(t *testing.T) {
	notifier := &testutil.MockNotifier{}
	p := newTestPoller(notifier)

	current := testutil.CampaignWithActiveDefend()
	previous := testutil.CampaignWithNoDefend()

	p.handleDefendEvent(current, previous)

	if notifier.Count() != 1 {
		t.Fatalf("expected 1 notification, got %d", notifier.Count())
	}
	msg := notifier.First()
	if msg.Transition != domain.EventTransitionStarted {
		t.Errorf("expected transition %s, got %s", domain.EventTransitionStarted, msg.Transition)
	}
	if msg.Kind != domain.EventKindDefend {
		t.Errorf("expected kind %s, got %s", domain.EventKindDefend, msg.Kind)
	}
}

// Case 3: defend event already failed on first poll — should not notify started (stale)
func TestHandleDefendEvent_StaleFailedOnFirstPoll(t *testing.T) {
	notifier := &testutil.MockNotifier{}
	p := newTestPoller(notifier)

	current := testutil.CampaignWithFailedDefend()
	previous := testutil.CampaignWithNoDefend()

	p.handleDefendEvent(current, previous)

	if notifier.Count() != 0 {
		t.Errorf("expected 0 notifications, got %d", notifier.Count())
	}
}

// Case 4: same event still active — no notification
func TestHandleDefendEvent_StillActive(t *testing.T) {
	notifier := &testutil.MockNotifier{}
	p := newTestPoller(notifier)

	current := testutil.CampaignWithActiveDefend()
	previous := testutil.CampaignWithActiveDefend()

	// store the event first
	p.events.SaveOngoingEvent(current.DefendEvent.ID, domain.EventKindDefend)

	p.handleDefendEvent(current, previous)

	if notifier.Count() != 0 {
		t.Errorf("expected 0 notifications, got %d", notifier.Count())
	}
}

// Case 5: same event ended with fail — notify failed
func TestHandleDefendEvent_EndedFailed(t *testing.T) {
	notifier := &testutil.MockNotifier{}
	p := newTestPoller(notifier)

	current := testutil.CampaignWithFailedDefend()
	previous := testutil.CampaignWithActiveDefend()

	// store the active event first
	p.events.SaveOngoingEvent(previous.DefendEvent.ID, domain.EventKindDefend)

	p.handleDefendEvent(current, previous)

	if notifier.Count() != 1 {
		t.Fatalf("expected 1 notification, got %d", notifier.Count())
	}
	msg := notifier.First()
	if msg.Transition != domain.EventTransitionFailed {
		t.Errorf("expected transition %s, got %s", domain.EventTransitionFailed, msg.Transition)
	}
}

// Case 6: same event ended with success — notify succeeded
func TestHandleDefendEvent_EndedSucceeded(t *testing.T) {
	notifier := &testutil.MockNotifier{}
	p := newTestPoller(notifier)

	current := testutil.CampaignWithSucceededDefend()
	previous := testutil.CampaignWithActiveDefend()

	p.events.SaveOngoingEvent(previous.DefendEvent.ID, domain.EventKindDefend)

	p.handleDefendEvent(current, previous)

	if notifier.Count() != 1 {
		t.Fatalf("expected 1 notification, got %d", notifier.Count())
	}
	msg := notifier.First()
	if msg.Transition != domain.EventTransitionSucceeded {
		t.Errorf("expected transition %s, got %s", domain.EventTransitionSucceeded, msg.Transition)
	}
}

// Case 7: different event ID — old ended (failed), new started
func TestHandleDefendEvent_NewEventReplacesOld(t *testing.T) {
	notifier := &testutil.MockNotifier{}
	p := newTestPoller(notifier)

	previous := testutil.CampaignWithFailedDefend()
	newDefend := testutil.DefendEventNewActive()
	current := testutil.CampaignWithNoDefend()
	current.DefendEvent = newDefend

	p.events.SaveOngoingEvent(previous.DefendEvent.ID, domain.EventKindDefend)

	p.handleDefendEvent(current, previous)

	if notifier.Count() != 2 {
		t.Fatalf("expected 2 notifications (ended + started), got %d", notifier.Count())
	}
	if notifier.First().Transition != domain.EventTransitionFailed {
		t.Errorf("expected first notification to be failed, got %s", notifier.First().Transition)
	}
	if notifier.Last().Transition != domain.EventTransitionStarted {
		t.Errorf("expected second notification to be started, got %s", notifier.Last().Transition)
	}
}

// --- handleAttackEvents tests ---

// Case 1: no attack events anywhere — no notification
func TestHandleAttackEvents_NoneAnywhere(t *testing.T) {
	notifier := &testutil.MockNotifier{}
	p := newTestPoller(notifier)

	current := testutil.CampaignWithNoDefend()

	p.handleAttackEvents(current)

	if notifier.Count() != 0 {
		t.Errorf("expected 0 notifications, got %d", notifier.Count())
	}
}

// Case 2: new active attack event — notify started
func TestHandleAttackEvents_NewActive(t *testing.T) {
	notifier := &testutil.MockNotifier{}
	p := newTestPoller(notifier)

	current := testutil.CampaignWithActiveAttack()

	p.handleAttackEvents(current)

	if notifier.Count() != 1 {
		t.Fatalf("expected 1 notification, got %d", notifier.Count())
	}
	msg := notifier.First()
	if msg.Transition != domain.EventTransitionStarted {
		t.Errorf("expected transition %s, got %s", domain.EventTransitionStarted, msg.Transition)
	}
	if msg.Kind != domain.EventKindAttack {
		t.Errorf("expected kind %s, got %s", domain.EventKindAttack, msg.Kind)
	}
}

// Case 3: active attack event still ongoing — no notification
func TestHandleAttackEvents_StillActive(t *testing.T) {
	notifier := &testutil.MockNotifier{}
	p := newTestPoller(notifier)

	current := testutil.CampaignWithActiveAttack()
	p.events.SaveOngoingEvent(current.AttackEvents[0].ID, domain.EventKindAttack)

	p.handleAttackEvents(current)

	if notifier.Count() != 0 {
		t.Errorf("expected 0 notifications, got %d", notifier.Count())
	}
}

// Case 4: stored attack event ended with success — notify succeeded
func TestHandleAttackEvents_EndedSucceeded(t *testing.T) {
	notifier := &testutil.MockNotifier{}
	p := newTestPoller(notifier)

	current := testutil.CampaignWithEndedAttack()
	p.events.SaveOngoingEvent(current.AttackEvents[0].ID, domain.EventKindAttack)

	p.handleAttackEvents(current)

	if notifier.Count() != 1 {
		t.Fatalf("expected 1 notification, got %d", notifier.Count())
	}
	msg := notifier.First()
	if msg.Transition != domain.EventTransitionSucceeded {
		t.Errorf("expected transition %s, got %s", domain.EventTransitionSucceeded, msg.Transition)
	}
}

// Case 5: stored attack event ended with fail — notify failed
func TestHandleAttackEvents_EndedFailed(t *testing.T) {
	notifier := &testutil.MockNotifier{}
	p := newTestPoller(notifier)

	failed := testutil.AttackEventFailed()
	current := testutil.CampaignWithNoDefend()
	current.AttackEvents = []domain.AttackEvent{failed}

	p.events.SaveOngoingEvent(failed.ID, domain.EventKindAttack)

	p.handleAttackEvents(current)

	if notifier.Count() != 1 {
		t.Fatalf("expected 1 notification, got %d", notifier.Count())
	}
	msg := notifier.First()
	if msg.Transition != domain.EventTransitionFailed {
		t.Errorf("expected transition %s, got %s", domain.EventTransitionFailed, msg.Transition)
	}
}
