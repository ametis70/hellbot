package config

import (
	"os"
	"testing"
)

// --- ResolveWebhookOptions ---

func TestResolveWebhookOptions_Valid(t *testing.T) {
	raw := RawOptions{"url": "https://example.com/hook"}
	opts, err := ResolveWebhookOptions(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts.URL != "https://example.com/hook" {
		t.Errorf("expected URL, got %q", opts.URL)
	}
}

func TestResolveWebhookOptions_NilOptions(t *testing.T) {
	_, err := ResolveWebhookOptions(nil)
	if err == nil {
		t.Error("expected error for nil options")
	}
}

func TestResolveWebhookOptions_MissingURL(t *testing.T) {
	_, err := ResolveWebhookOptions(RawOptions{})
	if err == nil {
		t.Error("expected error for missing url")
	}
}

func TestResolveWebhookOptions_URLEnvVar(t *testing.T) {
	t.Setenv("HOOK_URL", "https://env.example.com")
	opts, err := ResolveWebhookOptions(RawOptions{"url": "${HOOK_URL}"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts.URL != "https://env.example.com" {
		t.Errorf("expected env URL, got %q", opts.URL)
	}
}

func TestResolveWebhookOptions_SecretValueFile(t *testing.T) {
	path := writeSecretFile(t, "mysecret")
	opts, err := ResolveWebhookOptions(RawOptions{
		"url":               "https://example.com",
		"secret_value_file": path,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts.SecretValue != "mysecret" {
		t.Errorf("expected secret from file, got %q", opts.SecretValue)
	}
}

func TestResolveWebhookOptions_SecretValueEnvVar(t *testing.T) {
	t.Setenv("HOOK_SECRET", "tok123")
	opts, err := ResolveWebhookOptions(RawOptions{
		"url":          "https://example.com",
		"secret_value": "${HOOK_SECRET}",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts.SecretValue != "tok123" {
		t.Errorf("expected tok123, got %q", opts.SecretValue)
	}
}

// --- ResolveDiscordOptions ---

func TestResolveDiscordOptions_NilOptions(t *testing.T) {
	_, err := ResolveDiscordOptions(nil)
	if err == nil {
		t.Error("expected error for nil options")
	}
}

func TestResolveDiscordOptions_Valid(t *testing.T) {
	t.Setenv("DISCORD_TOKEN", "mytoken")
	opts, err := ResolveDiscordOptions(RawOptions{
		"token":      "${DISCORD_TOKEN}",
		"channel_id": "123456",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts.Token != "mytoken" {
		t.Errorf("expected mytoken, got %q", opts.Token)
	}
	if opts.ChannelID != "123456" {
		t.Errorf("expected 123456, got %q", opts.ChannelID)
	}
}

func TestResolveDiscordOptions_TokenFile(t *testing.T) {
	path := writeSecretFile(t, "filetoken")
	opts, err := ResolveDiscordOptions(RawOptions{
		"token_file": path,
		"channel_id": "abc",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts.Token != "filetoken" {
		t.Errorf("expected filetoken, got %q", opts.Token)
	}
}

func TestResolveDiscordOptions_MissingToken(t *testing.T) {
	_, err := ResolveDiscordOptions(RawOptions{"channel_id": "abc"})
	if err == nil {
		t.Error("expected error for missing token")
	}
}

func TestResolveDiscordOptions_MissingChannelID(t *testing.T) {
	_, err := ResolveDiscordOptions(RawOptions{"token": "tok"})
	if err == nil {
		t.Error("expected error for missing channel_id")
	}
}

// --- ResolveTelegramOptions ---

func TestResolveTelegramOptions_NilOptions(t *testing.T) {
	_, err := ResolveTelegramOptions(nil)
	if err == nil {
		t.Error("expected error for nil options")
	}
}

func TestResolveTelegramOptions_Valid(t *testing.T) {
	t.Setenv("TG_TOKEN", "tgtoken")
	opts, err := ResolveTelegramOptions(RawOptions{
		"token":   "${TG_TOKEN}",
		"chat_id": "-100123",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts.Token != "tgtoken" {
		t.Errorf("expected tgtoken, got %q", opts.Token)
	}
}

func TestResolveTelegramOptions_MissingToken(t *testing.T) {
	_, err := ResolveTelegramOptions(RawOptions{"chat_id": "-1"})
	if err == nil {
		t.Error("expected error for missing token")
	}
}

func TestResolveTelegramOptions_MissingChatID(t *testing.T) {
	_, err := ResolveTelegramOptions(RawOptions{"token": "tok"})
	if err == nil {
		t.Error("expected error for missing chat_id")
	}
}

// --- ResolveSQLiteStoreOptions ---

func TestResolveSQLiteStoreOptions_Defaults(t *testing.T) {
	opts, err := ResolveSQLiteStoreOptions(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts.Path != "hellbot.db" {
		t.Errorf("expected default path hellbot.db, got %q", opts.Path)
	}
}

func TestResolveSQLiteStoreOptions_CustomPath(t *testing.T) {
	opts, err := ResolveSQLiteStoreOptions(RawOptions{"path": "/data/bot.db"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts.Path != "/data/bot.db" {
		t.Errorf("expected /data/bot.db, got %q", opts.Path)
	}
}

func TestResolveSQLiteStoreOptions_EnvVar(t *testing.T) {
	t.Setenv("DB_PATH", "/tmp/test.db")
	opts, err := ResolveSQLiteStoreOptions(RawOptions{"path": "${DB_PATH}"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts.Path != "/tmp/test.db" {
		t.Errorf("expected /tmp/test.db, got %q", opts.Path)
	}
}

// --- ResolveValkeyStoreOptions ---

func TestResolveValkeyStoreOptions_Defaults(t *testing.T) {
	opts, err := ResolveValkeyStoreOptions(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts.Addr != "localhost:6379" {
		t.Errorf("expected default addr, got %q", opts.Addr)
	}
}

func TestResolveValkeyStoreOptions_Custom(t *testing.T) {
	t.Setenv("VALKEY_PASS", "secret")
	opts, err := ResolveValkeyStoreOptions(RawOptions{
		"addr":     "valkey:6379",
		"password": "${VALKEY_PASS}",
		"db":       1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts.Addr != "valkey:6379" {
		t.Errorf("expected valkey:6379, got %q", opts.Addr)
	}
	if opts.Password != "secret" {
		t.Errorf("expected secret password, got %q", opts.Password)
	}
	if opts.DB != 1 {
		t.Errorf("expected DB=1, got %d", opts.DB)
	}
}

// --- Load: store validation ---

func TestLoad_StoreSQLite(t *testing.T) {
	path := writeConfig(t, `
store:
  type: sqlite
  options:
    path: "/tmp/test.db"
notifiers:
  - id: "s"
    type: stdout
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Store.Type != StoreTypeSQLite {
		t.Errorf("expected sqlite store type, got %q", cfg.Store.Type)
	}
}

func TestLoad_StoreValkey(t *testing.T) {
	path := writeConfig(t, `
store:
  type: valkey
  options:
    addr: "localhost:6379"
notifiers:
  - id: "s"
    type: stdout
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Store.Type != StoreTypeValkey {
		t.Errorf("expected valkey store type, got %q", cfg.Store.Type)
	}
}

func TestLoad_StoreUnknown(t *testing.T) {
	path := writeConfig(t, `
store:
  type: badstore
notifiers:
  - id: "s"
    type: stdout
`)
	_, err := Load(path)
	if err == nil {
		t.Error("expected error for unknown store type")
	}
}

func TestLoad_WebhookNotifier(t *testing.T) {
	path := writeConfig(t, `
notifiers:
  - id: "wh"
    type: webhook
    options:
      url: "https://example.com/hook"
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Notifiers[0].Type != NotifierTypeWebhook {
		t.Errorf("expected webhook type")
	}
}

func TestLoad_DevMockServer(t *testing.T) {
	path := writeConfig(t, `
dev:
  mock_server: true
notifiers:
  - id: "s"
    type: stdout
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.Dev.MockServer {
		t.Error("expected dev.mock_server=true")
	}
}

func TestLoad_StdoutInvalidTimezone(t *testing.T) {
	// Load only validates the timezone is parseable for the global timezone field,
	// not per-notifier timezone — that is validated at runtime in main.
	// So an invalid stdout timezone does NOT cause Load to error.
	path := writeConfig(t, `
notifiers:
  - id: "s"
    type: stdout
    options:
      timezone: "Not/Valid"
`)
	_, err := Load(path)
	// No error expected — per-notifier timezone is not validated during Load.
	if err != nil {
		t.Errorf("unexpected error (per-notifier timezone not validated in Load): %v", err)
	}
}

// ensure os import is used
var _ = os.Getenv
