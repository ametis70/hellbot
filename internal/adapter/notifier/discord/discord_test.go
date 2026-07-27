package discord

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

func defendMsgSuperEarth(transition domain.EventTransition) domain.EventMessage {
	e := testutil.DefendEventActive()
	e.Region = domain.SuperEarthRegion
	return domain.EventMessage{
		Kind:        domain.EventKindDefend,
		Transition:  transition,
		DefendEvent: e,
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

// --- Defend event tests ---

func TestFormatMessage_DefendStarted_RegionName(t *testing.T) {
	msg := defendMsg(domain.EventTransitionStarted)
	result, err := formatMessage(msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// DefendEventActive uses EnemyIlluminate, Region 0 — but fixture uses region 0
	// which is Super Earth — let's check it contains the timestamp
	if !strings.Contains(result, "<t:") {
		t.Errorf("expected Discord timestamp, got: %s", result)
	}
}

func TestFormatMessage_DefendStarted_SuperEarth(t *testing.T) {
	msg := defendMsgSuperEarth(domain.EventTransitionStarted)
	result, err := formatMessage(msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "Super Earth") {
		t.Errorf("expected 'Super Earth' in message, got: %s", result)
	}
	if !strings.Contains(result, "<t:") {
		t.Errorf("expected Discord timestamp, got: %s", result)
	}
}

func TestFormatMessage_DefendStarted_NormalRegion(t *testing.T) {
	e := testutil.DefendEventActive()
	e.Region = 5
	e.Enemy = domain.EnemyIlluminate
	msg := domain.EventMessage{
		Kind:        domain.EventKindDefend,
		Transition:  domain.EventTransitionStarted,
		DefendEvent: e,
	}
	result, err := formatMessage(msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Region 5 for Illuminate = "Orionis Region"
	if !strings.Contains(result, "Orionis Region") {
		t.Errorf("expected region name 'Orionis Region', got: %s", result)
	}
	if !strings.Contains(result, "5/10") {
		t.Errorf("expected region position '5/10', got: %s", result)
	}
}

func TestFormatMessage_DefendSucceeded_SuperEarth(t *testing.T) {
	msg := defendMsgSuperEarth(domain.EventTransitionSucceeded)
	result, err := formatMessage(msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "Super Earth") {
		t.Errorf("expected 'Super Earth' in message, got: %s", result)
	}
	if !strings.Contains(result, "defended") {
		t.Errorf("expected 'defended' in message, got: %s", result)
	}
}

func TestFormatMessage_DefendSucceeded_NormalRegion(t *testing.T) {
	e := testutil.DefendEventActive()
	e.Region = 3
	e.Enemy = domain.EnemyCyborg
	msg := domain.EventMessage{
		Kind:        domain.EventKindDefend,
		Transition:  domain.EventTransitionSucceeded,
		DefendEvent: e,
	}
	result, err := formatMessage(msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Region 3 for Cyborg = "Pictor Sector"
	if !strings.Contains(result, "Pictor Sector") {
		t.Errorf("expected region name 'Pictor Sector', got: %s", result)
	}
}

func TestFormatMessage_DefendFailed_SuperEarth(t *testing.T) {
	msg := defendMsgSuperEarth(domain.EventTransitionFailed)
	result, err := formatMessage(msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "Super Earth") {
		t.Errorf("expected 'Super Earth' in message, got: %s", result)
	}
	if !strings.Contains(result, "fallen") {
		t.Errorf("expected 'fallen' in message, got: %s", result)
	}
}

func TestFormatMessage_DefendFailed_NormalRegion(t *testing.T) {
	e := testutil.DefendEventActive()
	e.Region = 2
	e.Enemy = domain.EnemyBug
	msg := domain.EventMessage{
		Kind:        domain.EventKindDefend,
		Transition:  domain.EventTransitionFailed,
		DefendEvent: e,
	}
	result, err := formatMessage(msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Region 2 for Bug = "Kruger System"
	if !strings.Contains(result, "Kruger System") {
		t.Errorf("expected region name 'Kruger System', got: %s", result)
	}
}

// --- Attack event tests ---

func TestFormatMessage_AttackStarted(t *testing.T) {
	msg := attackMsg(domain.EventTransitionStarted)
	result, err := formatMessage(msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "homeworld") {
		t.Errorf("expected 'homeworld' in message, got: %s", result)
	}
	if !strings.Contains(result, "<t:") {
		t.Errorf("expected Discord timestamp, got: %s", result)
	}
}

func TestFormatMessage_AttackSucceeded(t *testing.T) {
	msg := attackMsg(domain.EventTransitionSucceeded)
	result, err := formatMessage(msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "succeeded") {
		t.Errorf("expected 'succeeded' in message, got: %s", result)
	}
	if !strings.Contains(result, "defeated") {
		t.Errorf("expected 'defeated' in message, got: %s", result)
	}
}

func TestFormatMessage_AttackFailed(t *testing.T) {
	msg := attackMsg(domain.EventTransitionFailed)
	result, err := formatMessage(msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "failed") {
		t.Errorf("expected 'failed' in message, got: %s", result)
	}
	if !strings.Contains(result, "homeworld") {
		t.Errorf("expected 'homeworld' in message, got: %s", result)
	}
}

// --- Error cases ---

func TestFormatMessage_UnknownKind(t *testing.T) {
	msg := domain.EventMessage{Kind: "unknown"}
	_, err := formatMessage(msg)
	if err == nil {
		t.Error("expected error for unknown kind, got nil")
	}
}

func TestFormatMessage_NilDefendEvent(t *testing.T) {
	msg := domain.EventMessage{
		Kind:        domain.EventKindDefend,
		Transition:  domain.EventTransitionStarted,
		DefendEvent: nil,
	}
	_, err := formatMessage(msg)
	if err == nil {
		t.Error("expected error for nil defend event, got nil")
	}
}

func TestFormatMessage_NilAttackEvent(t *testing.T) {
	msg := domain.EventMessage{
		Kind:        domain.EventKindAttack,
		Transition:  domain.EventTransitionStarted,
		AttackEvent: nil,
	}
	_, err := formatMessage(msg)
	if err == nil {
		t.Error("expected error for nil attack event, got nil")
	}
}

// --- Enemy name in message ---

func TestFormatMessage_ContainsEnemyName(t *testing.T) {
	e := testutil.DefendEventActive()
	e.Region = 1
	e.Enemy = domain.EnemyIlluminate
	msg := domain.EventMessage{
		Kind:        domain.EventKindDefend,
		Transition:  domain.EventTransitionStarted,
		DefendEvent: e,
	}
	result, err := formatMessage(msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "Illuminate") {
		t.Errorf("expected enemy name in message, got: %s", result)
	}
}

// --- Discord timestamp format ---

func TestDiscordTimestamp(t *testing.T) {
	ts := time.Unix(1784501941, 0)
	result := discordTimestamp(ts.Unix())
	expected := "<t:1784501941:f>"
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}
