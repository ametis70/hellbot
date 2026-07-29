package helldivers1api_test

import (
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/ametis70/hellbot/internal/adapter/api/helldivers1api"
	"github.com/ametis70/hellbot/internal/adapter/store/memory"
	"github.com/ametis70/hellbot/internal/app"
	"github.com/ametis70/hellbot/internal/domain"
	"github.com/ametis70/hellbot/internal/port"
	"github.com/ametis70/hellbot/internal/testutil"
)

// newE2EPoller wires up a real poller against a mock server URL.
func newE2EPoller(serverURL string, notifier *testutil.MockNotifier) *app.Poller {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	store := memory.New()
	fetcher := helldivers1api.New(helldivers1api.Options{
		BaseURL:     serverURL,
		Timeout:     5 * time.Second,
		InsecureTLS: false,
	}, logger)
	return app.New(fetcher, store, store, []port.Notifier{notifier}, time.Hour, logger)
}

// newE2EPollerWithStore wires up a real poller with a provided store (for restart tests).
func newE2EPollerWithStore(serverURL string, store interface {
	port.CampaignStore
	port.EventStore
}, notifier *testutil.MockNotifier) *app.Poller {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	fetcher := helldivers1api.New(helldivers1api.Options{
		BaseURL:     serverURL,
		Timeout:     5 * time.Second,
		InsecureTLS: false,
	}, logger)
	return app.New(fetcher, store, store, []port.Notifier{notifier}, time.Hour, logger)
}

// TestE2E_FullWar drives a complete war from idle to war-won through the real
// HTTP client and poller, using a mock server that serves the WarScenario
// sequence. It verifies that every expected notification is emitted in order.
func TestE2E_FullWar(t *testing.T) {
	responses := testutil.WarScenario()
	srv := testutil.NewMockServer(responses)
	defer srv.Close()

	notifier := &testutil.MockNotifier{}
	poller := newE2EPoller(srv.URL(), notifier)

	// Drive each poll manually. The poller's poll() method is unexported, so we
	// use PollOnce which is exposed for testing.
	polls := len(responses)
	for i := range polls {
		poller.PollOnce()
		t.Logf("poll %d: notifications so far = %d", i+1, notifier.Count())
	}

	// Expected notification sequence:
	// poll 2: attack started
	// poll 4: attack succeeded
	// poll 6: defend started
	// poll 8: defend succeeded
	// poll 9: war won
	expected := []struct {
		kind       domain.EventKind
		transition domain.EventTransition
	}{
		{domain.EventKindAttack, domain.EventTransitionStarted},
		{domain.EventKindAttack, domain.EventTransitionSucceeded},
		{domain.EventKindDefend, domain.EventTransitionStarted},
		{domain.EventKindDefend, domain.EventTransitionSucceeded},
		{domain.EventKindWar, domain.EventTransitionSucceeded},
	}

	if notifier.Count() != len(expected) {
		t.Fatalf("expected %d notifications, got %d", len(expected), notifier.Count())
	}

	for i, want := range expected {
		got := notifier.Messages[i]
		if got.Kind != want.kind {
			t.Errorf("notification[%d]: expected kind %s, got %s", i, want.kind, got.Kind)
		}
		if got.Transition != want.transition {
			t.Errorf("notification[%d]: expected transition %s, got %s", i, want.transition, got.Transition)
		}
	}
}

// TestE2E_RestartDiscardsStaleState simulates the bot being stopped mid-war
// (during war 50) and restarted at war 53. The restarted bot must not emit any
// notifications for events that happened during downtime — it should treat the
// first poll after restart as a fresh baseline and only notify for events it
// observes from that point forward.
func TestE2E_RestartDiscardsStaleState(t *testing.T) {
	partA, partB := testutil.RestartScenario()

	// --- Part A: bot runs, detects attack start, then stops ---

	srvA := testutil.NewMockServer(partA)
	notifierA := &testutil.MockNotifier{}
	store := memory.New()
	pollerA := newE2EPollerWithStore(srvA.URL(), store, notifierA)

	for range len(partA) {
		pollerA.PollOnce()
	}
	srvA.Close()

	// Part A should have produced exactly one notification: attack started.
	if notifierA.Count() != 1 {
		t.Fatalf("partA: expected 1 notification (attack started), got %d", notifierA.Count())
	}
	if notifierA.First().Kind != domain.EventKindAttack || notifierA.First().Transition != domain.EventTransitionStarted {
		t.Errorf("partA: expected attack/started, got %s/%s", notifierA.First().Kind, notifierA.First().Transition)
	}

	// --- Part B: bot restarts with a fresh store at war 53 ---
	// The store is intentionally reset — simulates a memory store after process restart.
	// (A persistent store like SQLite would survive, but the war number jump means
	// the poller will detect a season change and emit a war notification instead of
	// stale events — that behaviour is covered separately.)

	freshStore := memory.New()
	srvB := testutil.NewMockServer(partB)
	notifierB := &testutil.MockNotifier{}
	pollerB := newE2EPollerWithStore(srvB.URL(), freshStore, notifierB)

	for range len(partB) {
		pollerB.PollOnce()
	}
	srvB.Close()

	// Part B: first poll establishes baseline (no diff), second poll sees attack 600
	// start, third poll sees it succeed.
	// Expected: two notifications — attack 600 started, attack 600 succeeded.
	if notifierB.Count() != 2 {
		t.Fatalf("partB: expected 2 notifications (attack started + succeeded), got %d", notifierB.Count())
	}
	if notifierB.Messages[0].Kind != domain.EventKindAttack || notifierB.Messages[0].Transition != domain.EventTransitionStarted {
		t.Errorf("partB[0]: expected attack/started, got %s/%s", notifierB.Messages[0].Kind, notifierB.Messages[0].Transition)
	}
	msg := notifierB.Last()
	if msg.Kind != domain.EventKindAttack {
		t.Errorf("partB: expected kind attack, got %s", msg.Kind)
	}
	if msg.Transition != domain.EventTransitionSucceeded {
		t.Errorf("partB: expected transition succeeded, got %s", msg.Transition)
	}
	if msg.AttackEvent == nil || msg.AttackEvent.ID != 600 {
		t.Errorf("partB: expected attack event ID 600, got %v", msg.AttackEvent)
	}
}
