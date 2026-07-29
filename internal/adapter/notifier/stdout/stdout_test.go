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

func TestNotify_SuccessDoesNotError(t *testing.T) {
	n := New(Options{Timezone: time.UTC})
	e := testutil.AttackEventActive()
	err := n.Notify(domain.EventMessage{
		Kind:        domain.EventKindAttack,
		Transition:  domain.EventTransitionStarted,
		AttackEvent: &e,
	})
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestNotify_RenderErrorPropagated(t *testing.T) {
	n := New(Options{Timezone: time.UTC})
	// A message with an unknown kind causes RenderEvent to return an error.
	err := n.Notify(domain.EventMessage{Kind: "bogus", Transition: "started"})
	if err == nil {
		t.Error("expected error for unrenderable message, got nil")
	}
}

func TestNew_MergesTemplates(t *testing.T) {
	custom := &domain.Templates{AttackSucceeded: "custom template"}
	n := New(Options{Templates: custom})
	if n.templates.AttackSucceeded != "custom template" {
		t.Errorf("expected custom template to be merged, got %q", n.templates.AttackSucceeded)
	}
}
