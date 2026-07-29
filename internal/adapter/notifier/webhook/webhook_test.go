package webhook_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ametis70/hellbot/internal/adapter/notifier/webhook"
	"github.com/ametis70/hellbot/internal/domain"
	"github.com/ametis70/hellbot/internal/testutil"
)

// captureServer records the last request received.
type captureServer struct {
	header http.Header
	body   []byte
	status int
}

func (c *captureServer) handler(w http.ResponseWriter, r *http.Request) {
	c.header = r.Header.Clone()
	c.body, _ = io.ReadAll(r.Body)
	w.WriteHeader(c.status)
}

func newCapture(status int) (*captureServer, *httptest.Server) {
	c := &captureServer{status: status}
	srv := httptest.NewServer(http.HandlerFunc(c.handler))
	return c, srv
}

func newNotifier(t *testing.T, url string, opts ...func(*webhook.Options)) *webhook.Notifier {
	t.Helper()
	o := webhook.Options{URL: url, Timeout: 5 * time.Second}
	for _, fn := range opts {
		fn(&o)
	}
	n, err := webhook.New(o, testutil.DiscardLogger())
	if err != nil {
		t.Fatalf("failed to create webhook notifier: %v", err)
	}
	return n
}

// TestWebhook_PostsJSON verifies that Notify sends a POST with Content-Type
// application/json and a correctly structured body.
func TestWebhook_PostsJSON(t *testing.T) {
	capture, srv := newCapture(http.StatusOK)
	defer srv.Close()

	n := newNotifier(t, srv.URL)

	msg := domain.EventMessage{
		Kind:        domain.EventKindAttack,
		Transition:  domain.EventTransitionStarted,
		AttackEvent: ptr(testutil.AttackEventActive()),
	}

	if err := n.Notify(msg); err != nil {
		t.Fatalf("Notify returned error: %v", err)
	}

	if ct := capture.header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}

	var payload webhook.Payload
	if err := json.Unmarshal(capture.body, &payload); err != nil {
		t.Fatalf("failed to unmarshal payload: %v", err)
	}
	if payload.Kind != "attack" {
		t.Errorf("expected kind=attack, got %q", payload.Kind)
	}
	if payload.Transition != "started" {
		t.Errorf("expected transition=started, got %q", payload.Transition)
	}
	if payload.AttackEvent == nil {
		t.Fatal("expected attack_event to be set")
	}
	if payload.AttackEvent.ID != testutil.AttackEventActive().ID {
		t.Errorf("expected attack event ID %d, got %d", testutil.AttackEventActive().ID, payload.AttackEvent.ID)
	}
	if payload.DefendEvent != nil {
		t.Error("expected defend_event to be nil")
	}
	if payload.WarEvent != nil {
		t.Error("expected war_event to be nil")
	}
}

// TestWebhook_SecretHeader verifies that the configured auth header is sent.
func TestWebhook_SecretHeader(t *testing.T) {
	capture, srv := newCapture(http.StatusNoContent)
	defer srv.Close()

	n := newNotifier(t, srv.URL, func(o *webhook.Options) {
		o.SecretHeader = "Authorization"
		o.SecretValue = "Bearer secret123"
	})

	_ = n.Notify(domain.EventMessage{
		Kind:       domain.EventKindWar,
		Transition: domain.EventTransitionSucceeded,
		WarEvent:   &domain.WarEvent{Season: 50},
	})

	if got := capture.header.Get("Authorization"); got != "Bearer secret123" {
		t.Errorf("expected Authorization header %q, got %q", "Bearer secret123", got)
	}
}

// TestWebhook_NoSecretHeader verifies that no auth header is added when not configured.
func TestWebhook_NoSecretHeader(t *testing.T) {
	capture, srv := newCapture(http.StatusOK)
	defer srv.Close()

	n := newNotifier(t, srv.URL)
	_ = n.Notify(domain.EventMessage{
		Kind:       domain.EventKindWar,
		Transition: domain.EventTransitionFailed,
		WarEvent:   &domain.WarEvent{Season: 1},
	})

	if got := capture.header.Get("Authorization"); got != "" {
		t.Errorf("expected no Authorization header, got %q", got)
	}
}

// TestWebhook_ServerError verifies that a non-2xx response returns an error.
func TestWebhook_ServerError(t *testing.T) {
	_, srv := newCapture(http.StatusInternalServerError)
	defer srv.Close()

	n := newNotifier(t, srv.URL)
	err := n.Notify(domain.EventMessage{
		Kind:       domain.EventKindWar,
		Transition: domain.EventTransitionFailed,
		WarEvent:   &domain.WarEvent{Season: 1},
	})
	if err == nil {
		t.Error("expected error for 500 response, got nil")
	}
}

// TestWebhook_DefendEventPayload verifies defend event fields are serialised correctly.
func TestWebhook_DefendEventPayload(t *testing.T) {
	capture, srv := newCapture(http.StatusOK)
	defer srv.Close()

	n := newNotifier(t, srv.URL)
	_ = n.Notify(domain.EventMessage{
		Kind:        domain.EventKindDefend,
		Transition:  domain.EventTransitionSucceeded,
		DefendEvent: testutil.DefendEventSucceeded(),
	})

	var payload webhook.Payload
	if err := json.Unmarshal(capture.body, &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if payload.DefendEvent == nil {
		t.Fatal("expected defend_event to be set")
	}
	de := payload.DefendEvent
	want := testutil.DefendEventSucceeded()
	if de.ID != want.ID {
		t.Errorf("id: want %d, got %d", want.ID, de.ID)
	}
	if de.Enemy != want.Enemy.String() {
		t.Errorf("enemy: want %q, got %q", want.Enemy.String(), de.Enemy)
	}
	if de.StartTimeUnix != want.StartTime.Unix() {
		t.Errorf("start_time_unix: want %d, got %d", want.StartTime.Unix(), de.StartTimeUnix)
	}
	if de.EndTimeUnix != want.EndTime.Unix() {
		t.Errorf("end_time_unix: want %d, got %d", want.EndTime.Unix(), de.EndTimeUnix)
	}
	// Verify RFC3339 parse round-trips.
	if _, err := time.Parse(time.RFC3339, de.StartTime); err != nil {
		t.Errorf("start_time not valid RFC3339: %v", err)
	}
}

// TestWebhook_URLRequired verifies that New returns an error when URL is empty.
func TestWebhook_URLRequired(t *testing.T) {
	_, err := webhook.New(webhook.Options{}, testutil.DiscardLogger())
	if err == nil {
		t.Error("expected error for empty URL, got nil")
	}
}

func ptr[T any](v T) *T { return &v }
