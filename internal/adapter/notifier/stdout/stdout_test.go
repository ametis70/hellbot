package stdout

import (
	"strings"
	"testing"
	"time"

	"github.com/ametis70/hellbot/internal/domain"
	"github.com/ametis70/hellbot/internal/testutil"
)

func defendMsg(transition domain.EventTransition) domain.EventMessage {
	return domain.EventMessage{
		Kind:        domain.EventKindDefend,
		Transition:  transition,
		DefendEvent: testutil.DefendEventActive(),
	}
}

func attackMsg(transition domain.EventTransition) domain.EventMessage {
	e := testutil.AttackEventActive()
	return domain.EventMessage{
		Kind:        domain.EventKindAttack,
		Transition:  transition,
		AttackEvent: &e,
	}
}

func TestFormatMessage_DefendEvent(t *testing.T) {
	msg := defendMsg(domain.EventTransitionStarted)
	result, err := formatMessage(msg, time.UTC)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(result, "[defend]") {
		t.Errorf("expected result to start with [defend], got: %s", result)
	}
	if !strings.Contains(result, string(domain.EventTransitionStarted)) {
		t.Errorf("expected result to contain transition, got: %s", result)
	}
}

func TestFormatMessage_AttackEvent(t *testing.T) {
	msg := attackMsg(domain.EventTransitionStarted)
	result, err := formatMessage(msg, time.UTC)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(result, "[attack]") {
		t.Errorf("expected result to start with [attack], got: %s", result)
	}
	if !strings.Contains(result, string(domain.EventTransitionStarted)) {
		t.Errorf("expected result to contain transition, got: %s", result)
	}
}

func TestFormatMessage_UnknownKind(t *testing.T) {
	msg := domain.EventMessage{
		Kind:       "unknown",
		Transition: domain.EventTransitionStarted,
	}
	_, err := formatMessage(msg, time.UTC)
	if err == nil {
		t.Error("expected error for unknown event kind, got nil")
	}
}

func TestFormatMessage_TimezoneApplied(t *testing.T) {
	lisbon, _ := time.LoadLocation("Europe/Lisbon")
	newYork, _ := time.LoadLocation("America/New_York")

	msg := defendMsg(domain.EventTransitionStarted)

	lisbonResult, _ := formatMessage(msg, lisbon)
	newYorkResult, _ := formatMessage(msg, newYork)

	if lisbonResult == newYorkResult {
		t.Error("expected different output for different timezones, got identical results")
	}
}

func TestFormatMessage_AllTransitions(t *testing.T) {
	transitions := []domain.EventTransition{
		domain.EventTransitionStarted,
		domain.EventTransitionSucceeded,
		domain.EventTransitionFailed,
	}

	for _, tr := range transitions {
		t.Run(string(tr), func(t *testing.T) {
			msg := defendMsg(tr)
			result, err := formatMessage(msg, time.UTC)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !strings.Contains(result, string(tr)) {
				t.Errorf("expected result to contain %s, got: %s", tr, result)
			}
		})
	}
}

func TestNew_DefaultsToUTC(t *testing.T) {
	n := New(Options{})
	if n.opts.Timezone != time.UTC {
		t.Error("expected default timezone to be UTC")
	}
}
