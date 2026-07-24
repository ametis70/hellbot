package discord

import (
	"strings"
	"testing"

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

func TestFormatMessage_DefendStarted(t *testing.T) {
	msg := defendMsg(domain.EventTransitionStarted)
	result, err := formatMessage(msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "started") {
		t.Errorf("expected 'started' in message, got: %s", result)
	}
	if !strings.Contains(result, "<t:") {
		t.Errorf("expected Discord timestamp in message, got: %s", result)
	}
}

func TestFormatMessage_DefendSucceeded(t *testing.T) {
	msg := defendMsg(domain.EventTransitionSucceeded)
	result, err := formatMessage(msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "succeeded") {
		t.Errorf("expected 'succeeded' in message, got: %s", result)
	}
}

func TestFormatMessage_DefendFailed(t *testing.T) {
	msg := defendMsg(domain.EventTransitionFailed)
	result, err := formatMessage(msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "failed") {
		t.Errorf("expected 'failed' in message, got: %s", result)
	}
}

func TestFormatMessage_AttackStarted(t *testing.T) {
	msg := attackMsg(domain.EventTransitionStarted)
	result, err := formatMessage(msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "started") {
		t.Errorf("expected 'started' in message, got: %s", result)
	}
	if !strings.Contains(result, "<t:") {
		t.Errorf("expected Discord timestamp in message, got: %s", result)
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
}

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

func TestFormatMessage_ContainsEnemyName(t *testing.T) {
	msg := defendMsg(domain.EventTransitionStarted)
	result, err := formatMessage(msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// DefendEventActive uses EnemyIlluminate
	if !strings.Contains(result, "Illuminate") {
		t.Errorf("expected enemy name in message, got: %s", result)
	}
}
