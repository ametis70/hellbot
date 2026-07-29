package telegram_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ametis70/hellbot/internal/adapter/notifier/telegram"
	"github.com/ametis70/hellbot/internal/domain"
	"github.com/ametis70/hellbot/internal/testutil"
)

// fakeTelegramServer stubs the Telegram Bot API.
// It responds OK to sendMessage and returns empty updates to getUpdates.
type fakeTelegramServer struct {
	sends []map[string]any
}

func (f *fakeTelegramServer) handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch {
	case len(r.URL.Path) > 0 && r.URL.Path[len(r.URL.Path)-len("getUpdates"):] == "getUpdates":
		// Long-poll: return empty result immediately so the goroutine doesn't block.
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "result": []any{}})

	case len(r.URL.Path) > 0 && r.URL.Path[len(r.URL.Path)-len("sendMessage"):] == "sendMessage":
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		f.sends = append(f.sends, body)
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})

	default:
		http.NotFound(w, r)
	}
}

func newFakeServer() (*fakeTelegramServer, *httptest.Server) {
	fs := &fakeTelegramServer{}
	srv := httptest.NewServer(http.HandlerFunc(fs.handler))
	return fs, srv
}

func newNotifier(t *testing.T, apiBase string) *telegram.Notifier {
	t.Helper()
	n, err := telegram.New(telegram.Options{
		Token:    "testtoken",
		ChatID:   "-100",
		Timezone: time.UTC,
		APIBase:  apiBase,
	}, testutil.DiscardLogger())
	if err != nil {
		t.Fatalf("failed to create telegram notifier: %v", err)
	}
	t.Cleanup(func() { _ = n.Close() })
	return n
}

// TestTelegram_Notify_AttackStarted verifies a sendMessage call is made.
func TestTelegram_Notify_AttackStarted(t *testing.T) {
	fs, srv := newFakeServer()
	defer srv.Close()

	n := newNotifier(t, srv.URL)
	ev := testutil.AttackEventActive()
	err := n.Notify(domain.EventMessage{
		Kind:        domain.EventKindAttack,
		Transition:  domain.EventTransitionStarted,
		AttackEvent: &ev,
	})
	if err != nil {
		t.Fatalf("Notify returned error: %v", err)
	}
	if len(fs.sends) != 1 {
		t.Fatalf("expected 1 sendMessage call, got %d", len(fs.sends))
	}
	text, _ := fs.sends[0]["text"].(string)
	if text == "" {
		t.Error("expected non-empty message text")
	}
}

// TestTelegram_Notify_WarWon verifies war won notification is sent.
func TestTelegram_Notify_WarWon(t *testing.T) {
	fs, srv := newFakeServer()
	defer srv.Close()

	n := newNotifier(t, srv.URL)
	err := n.Notify(domain.EventMessage{
		Kind:       domain.EventKindWar,
		Transition: domain.EventTransitionSucceeded,
		WarEvent:   &domain.WarEvent{Season: 50},
	})
	if err != nil {
		t.Fatalf("Notify returned error: %v", err)
	}
	if len(fs.sends) != 1 {
		t.Fatalf("expected 1 sendMessage call, got %d", len(fs.sends))
	}
}

// TestTelegram_Notify_DefendStarted verifies defend notification is sent.
func TestTelegram_Notify_DefendStarted(t *testing.T) {
	fs, srv := newFakeServer()
	defer srv.Close()

	n := newNotifier(t, srv.URL)
	err := n.Notify(domain.EventMessage{
		Kind:        domain.EventKindDefend,
		Transition:  domain.EventTransitionStarted,
		DefendEvent: testutil.DefendEventActive(),
	})
	if err != nil {
		t.Fatalf("Notify returned error: %v", err)
	}
	if len(fs.sends) != 1 {
		t.Fatalf("expected 1 sendMessage call, got %d", len(fs.sends))
	}
}

// TestTelegram_Notify_ServerError verifies an error is returned on non-200.
func TestTelegram_Notify_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path[len(r.URL.Path)-len("sendMessage"):] == "sendMessage" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// getUpdates: return empty
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "result": []any{}})
	}))
	defer srv.Close()

	n := newNotifier(t, srv.URL)
	ev := testutil.AttackEventActive()
	err := n.Notify(domain.EventMessage{
		Kind:        domain.EventKindAttack,
		Transition:  domain.EventTransitionStarted,
		AttackEvent: &ev,
	})
	if err == nil {
		t.Error("expected error for 500 response")
	}
}

// TestTelegram_DefaultTemplates verifies default templates are non-empty.
func TestTelegram_DefaultTemplates(t *testing.T) {
	tmpl := telegram.DefaultTemplates()
	if tmpl.DefendRegionStarted == "" {
		t.Error("expected non-empty DefendRegionStarted template")
	}
	if tmpl.WarWon == "" {
		t.Error("expected non-empty WarWon template")
	}
}

// TestTelegram_TimeFormatter verifies the formatter produces a non-empty string.
func TestTelegram_TimeFormatter(t *testing.T) {
	f := telegram.TimeFormatter(time.UTC)
	result := f(time.Now())
	if result == "" {
		t.Error("expected non-empty formatted time")
	}
}

// TestTelegram_New_MissingToken verifies New fails without token.
func TestTelegram_New_MissingToken(t *testing.T) {
	_, err := telegram.New(telegram.Options{ChatID: "-1"}, testutil.DiscardLogger())
	if err == nil {
		t.Error("expected error for missing token")
	}
}

// TestTelegram_New_MissingChatID verifies New fails without chat ID.
func TestTelegram_New_MissingChatID(t *testing.T) {
	_, err := telegram.New(telegram.Options{Token: "tok"}, testutil.DiscardLogger())
	if err == nil {
		t.Error("expected error for missing chat_id")
	}
}
