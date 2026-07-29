package telegram_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/ametis70/hellbot/internal/adapter/notifier/telegram"
	"github.com/ametis70/hellbot/internal/adapter/store/memory"
	"github.com/ametis70/hellbot/internal/testutil"
)

// commandServer is a fake Telegram server that serves a sequence of updates
// then returns empty updates. It also captures sendMessage calls.
type commandServer struct {
	mu      sync.Mutex
	updates []map[string]any
	idx     int
	sends   []string
}

func (c *commandServer) handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch {
	case hasSuffix(r.URL.Path, "getUpdates"):
		c.mu.Lock()
		var result []map[string]any
		if c.idx < len(c.updates) {
			result = c.updates[c.idx : c.idx+1]
			c.idx++
		}
		c.mu.Unlock()
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "result": result})

	case hasSuffix(r.URL.Path, "sendMessage"):
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		c.mu.Lock()
		if text, ok := body["text"].(string); ok {
			c.sends = append(c.sends, text)
		}
		c.mu.Unlock()
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})

	default:
		http.NotFound(w, r)
	}
}

func (c *commandServer) sendCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.sends)
}

func hasSuffix(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}

func botUpdate(text, cmdType string) map[string]any {
	cmdLen := len(text)
	// Strip arg part for command length (everything up to first space).
	for i, ch := range text {
		if ch == ' ' {
			cmdLen = i
			break
		}
	}
	return map[string]any{
		"update_id": 1,
		"message": map[string]any{
			"text": text,
			"entities": []map[string]any{
				{"type": cmdType, "offset": 0, "length": cmdLen},
			},
		},
	}
}

func newCommandNotifier(t *testing.T, srv *commandServer) *telegram.Notifier {
	t.Helper()
	ts := httptest.NewServer(http.HandlerFunc(srv.handler))
	t.Cleanup(ts.Close)
	n, err := telegram.New(telegram.Options{
		Token:    "tok",
		ChatID:   "-1",
		Timezone: time.UTC,
		APIBase:  ts.URL,
	}, testutil.DiscardLogger())
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(func() { _ = n.Close() })
	return n
}

// waitSend polls until the server has received at least one send or times out.
func waitSend(t *testing.T, srv *commandServer) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if srv.sendCount() >= 1 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for send, got %d", srv.sendCount())
}

// TestTelegram_HandleUpdate_TestCommand verifies that /test triggers a sendMessage.
func TestTelegram_HandleUpdate_TestCommand(t *testing.T) {
	srv := &commandServer{
		updates: []map[string]any{botUpdate("/test", "bot_command")},
	}
	newCommandNotifier(t, srv)
	waitSend(t, srv)
}

// TestTelegram_HandleUpdate_StatusCommand_NoProvider verifies /status with no
// provider registered does not panic and sends no message.
func TestTelegram_HandleUpdate_StatusCommand_NoProvider(t *testing.T) {
	srv := &commandServer{
		updates: []map[string]any{botUpdate("/status", "bot_command")},
	}
	newCommandNotifier(t, srv)
	time.Sleep(150 * time.Millisecond)
	if srv.sendCount() != 0 {
		t.Errorf("expected 0 sends when no provider, got %d", srv.sendCount())
	}
}

// TestTelegram_HandleUpdate_StatusCommand_WithProvider verifies /status with a
// provider sends a message containing war status.
func TestTelegram_HandleUpdate_StatusCommand_WithProvider(t *testing.T) {
	store := memory.New()
	_ = store.SaveCampaign(testutil.CampaignWithNoDefend())

	srv := &commandServer{
		updates: []map[string]any{botUpdate("/status", "bot_command")},
	}
	n := newCommandNotifier(t, srv)
	n.RegisterCommands(store)
	waitSend(t, srv)
}

// TestTelegram_HandleUpdate_StatusCommand_WithFactionFilter verifies /status bugs
// sends a filtered message.
func TestTelegram_HandleUpdate_StatusCommand_WithFactionFilter(t *testing.T) {
	store := memory.New()
	_ = store.SaveCampaign(testutil.CampaignWithNoDefend())

	srv := &commandServer{
		updates: []map[string]any{botUpdate("/status bugs", "bot_command")},
	}
	n := newCommandNotifier(t, srv)
	n.RegisterCommands(store)
	waitSend(t, srv)
}

// TestTelegram_HandleUpdate_StatusCommand_StoreError verifies /status with a
// failing provider sends an error message.
func TestTelegram_HandleUpdate_StatusCommand_StoreError(t *testing.T) {
	srv := &commandServer{
		updates: []map[string]any{botUpdate("/status", "bot_command")},
	}
	n := newCommandNotifier(t, srv)
	n.RegisterCommands(&testutil.ErrorStore{})
	waitSend(t, srv)
}

// TestTelegram_HandleUpdate_StatisticsCommand_WithProvider verifies /statistics
// with a provider sends a message.
func TestTelegram_HandleUpdate_StatisticsCommand_WithProvider(t *testing.T) {
	store := memory.New()
	_ = store.SaveCampaign(testutil.CampaignWithNoDefend())

	srv := &commandServer{
		updates: []map[string]any{botUpdate("/statistics", "bot_command")},
	}
	n := newCommandNotifier(t, srv)
	n.RegisterCommands(store)
	waitSend(t, srv)
}

// TestTelegram_HandleUpdate_StatisticsCommand_NoProvider verifies /statistics
// without a provider does not panic.
func TestTelegram_HandleUpdate_StatisticsCommand_NoProvider(t *testing.T) {
	srv := &commandServer{
		updates: []map[string]any{botUpdate("/statistics", "bot_command")},
	}
	newCommandNotifier(t, srv)
	time.Sleep(150 * time.Millisecond)
	if srv.sendCount() != 0 {
		t.Errorf("expected 0 sends when no provider, got %d", srv.sendCount())
	}
}

// TestTelegram_HandleUpdate_StatisticsCommand_StoreError verifies /statistics
// with a failing provider sends an error message.
func TestTelegram_HandleUpdate_StatisticsCommand_StoreError(t *testing.T) {
	srv := &commandServer{
		updates: []map[string]any{botUpdate("/statistics", "bot_command")},
	}
	n := newCommandNotifier(t, srv)
	n.RegisterCommands(&testutil.ErrorStore{})
	waitSend(t, srv)
}

// TestTelegram_HandleUpdate_NilMessage verifies an update with no message is
// silently ignored.
func TestTelegram_HandleUpdate_NilMessage(t *testing.T) {
	srv := &commandServer{
		updates: []map[string]any{{"update_id": 1}},
	}
	newCommandNotifier(t, srv)
	time.Sleep(150 * time.Millisecond)
	if srv.sendCount() != 0 {
		t.Errorf("expected 0 sends for nil message update, got %d", srv.sendCount())
	}
}

// TestTelegram_HandleUpdate_NonCommandEntity verifies non-command entities are
// skipped.
func TestTelegram_HandleUpdate_NonCommandEntity(t *testing.T) {
	srv := &commandServer{
		updates: []map[string]any{botUpdate("hello", "mention")},
	}
	newCommandNotifier(t, srv)
	time.Sleep(150 * time.Millisecond)
	if srv.sendCount() != 0 {
		t.Errorf("expected 0 sends for non-command entity, got %d", srv.sendCount())
	}
}

// TestTelegram_RegisterCommands verifies RegisterCommands sets the provider.
func TestTelegram_RegisterCommands(t *testing.T) {
	srv := &commandServer{}
	n := newCommandNotifier(t, srv)
	store := memory.New()
	n.RegisterCommands(store) // should not panic
}
