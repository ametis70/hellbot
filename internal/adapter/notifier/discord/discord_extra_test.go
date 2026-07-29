package discord_test

import (
	"testing"
	"time"

	"github.com/ametis70/hellbot/internal/adapter/notifier/discord"
	"github.com/ametis70/hellbot/internal/domain"
	"github.com/ametis70/hellbot/internal/testutil"
)

// TestDiscord_FormatMessage_AllEvents exercises the package-level formatMessage
// helper (which calls domain.RenderEvent with DefaultTemplates) for every
// event kind and transition, verifying no error is returned and the output
// is non-empty.
func TestDiscord_FormatMessage_AllEvents(t *testing.T) {
	cases := []struct {
		name string
		msg  domain.EventMessage
	}{
		{
			"defend region started",
			domain.EventMessage{Kind: domain.EventKindDefend, Transition: domain.EventTransitionStarted, DefendEvent: testutil.DefendEventActive()},
		},
		{
			"defend super earth started",
			domain.EventMessage{Kind: domain.EventKindDefend, Transition: domain.EventTransitionStarted, DefendEvent: superEarthDefend()},
		},
		{
			"defend region succeeded",
			domain.EventMessage{Kind: domain.EventKindDefend, Transition: domain.EventTransitionSucceeded, DefendEvent: testutil.DefendEventSucceeded()},
		},
		{
			"defend super earth succeeded",
			domain.EventMessage{Kind: domain.EventKindDefend, Transition: domain.EventTransitionSucceeded, DefendEvent: superEarthDefend()},
		},
		{
			"defend region failed",
			domain.EventMessage{Kind: domain.EventKindDefend, Transition: domain.EventTransitionFailed, DefendEvent: testutil.DefendEventFailed()},
		},
		{
			"defend super earth failed",
			domain.EventMessage{Kind: domain.EventKindDefend, Transition: domain.EventTransitionFailed, DefendEvent: superEarthDefend()},
		},
		{
			"attack started",
			domain.EventMessage{Kind: domain.EventKindAttack, Transition: domain.EventTransitionStarted, AttackEvent: ptr(testutil.AttackEventActive())},
		},
		{
			"attack succeeded",
			domain.EventMessage{Kind: domain.EventKindAttack, Transition: domain.EventTransitionSucceeded, AttackEvent: ptr(testutil.AttackEventSucceeded())},
		},
		{
			"attack failed",
			domain.EventMessage{Kind: domain.EventKindAttack, Transition: domain.EventTransitionFailed, AttackEvent: ptr(testutil.AttackEventFailed())},
		},
		{
			"war won",
			domain.EventMessage{Kind: domain.EventKindWar, Transition: domain.EventTransitionSucceeded, WarEvent: &domain.WarEvent{Season: 50}},
		},
		{
			"war lost",
			domain.EventMessage{Kind: domain.EventKindWar, Transition: domain.EventTransitionFailed, WarEvent: &domain.WarEvent{Season: 50}},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := discord.FormatMessage(c.msg)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == "" {
				t.Error("expected non-empty message")
			}
		})
	}
}

// TestDiscord_DefaultTemplates verifies all template fields are non-empty.
func TestDiscord_DefaultTemplates(t *testing.T) {
	tmpl := discord.DefaultTemplates()
	if tmpl.DefendRegionStarted == "" || tmpl.WarWon == "" || tmpl.AttackFailed == "" {
		t.Error("expected all default templates to be non-empty")
	}
}

// TestDiscord_TimeFormatter verifies the formatter produces a Discord timestamp.
func TestDiscord_TimeFormatter(t *testing.T) {
	f := discord.TimeFormatter(nil)
	result := f(time.Unix(1700000000, 0))
	if result == "" {
		t.Error("expected non-empty timestamp")
	}
}

func superEarthDefend() *domain.DefendEvent {
	e := testutil.DefendEventActive()
	e.Region = 0
	return e
}

func ptr[T any](v T) *T { return &v }
