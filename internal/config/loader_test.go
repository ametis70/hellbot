package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "config-*.yml")
	if err != nil {
		t.Fatalf("creating temp config: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("writing temp config: %v", err)
	}
	f.Close()
	return f.Name()
}

func writeSecretFile(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "secret-*")
	if err != nil {
		t.Fatalf("creating temp secret file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("writing temp secret: %v", err)
	}
	f.Close()
	return f.Name()
}

// --- Load tests ---

func TestLoad_ValidStdout(t *testing.T) {
	path := writeConfig(t, `
poll_interval: 30s
timezone: "UTC"
notifiers:
  - id: "test-stdout"
    type: stdout
    options:
      timezone: "Europe/Lisbon"
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.PollInterval != 30*time.Second {
		t.Errorf("expected 30s poll interval, got %v", cfg.PollInterval)
	}
	if cfg.Timezone != "UTC" {
		t.Errorf("expected UTC timezone, got %s", cfg.Timezone)
	}
	if len(cfg.Notifiers) != 1 {
		t.Fatalf("expected 1 notifier, got %d", len(cfg.Notifiers))
	}
	if cfg.Notifiers[0].ID != "test-stdout" {
		t.Errorf("expected notifier id test-stdout, got %s", cfg.Notifiers[0].ID)
	}
}

func TestLoad_DefaultPollInterval(t *testing.T) {
	path := writeConfig(t, `
notifiers:
  - id: "s"
    type: stdout
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.PollInterval != defaultPollInterval {
		t.Errorf("expected default poll interval %v, got %v", defaultPollInterval, cfg.PollInterval)
	}
}

func TestLoad_DefaultTimezone(t *testing.T) {
	path := writeConfig(t, `
notifiers:
  - id: "s"
    type: stdout
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Timezone != defaultTimezone {
		t.Errorf("expected default timezone %s, got %s", defaultTimezone, cfg.Timezone)
	}
}

func TestLoad_InvalidPollInterval(t *testing.T) {
	path := writeConfig(t, `poll_interval: "notaduration"`)
	_, err := Load(path)
	if err == nil {
		t.Error("expected error for invalid poll_interval, got nil")
	}
}

func TestLoad_InvalidTimezone(t *testing.T) {
	path := writeConfig(t, `timezone: "Not/ATimezone"`)
	_, err := Load(path)
	if err == nil {
		t.Error("expected error for invalid timezone, got nil")
	}
}

func TestLoad_MissingNotifierID(t *testing.T) {
	path := writeConfig(t, `
notifiers:
  - type: stdout
`)
	_, err := Load(path)
	if err == nil {
		t.Error("expected error for missing notifier id, got nil")
	}
}

func TestLoad_DuplicateNotifierID(t *testing.T) {
	path := writeConfig(t, `
notifiers:
  - id: "dup"
    type: stdout
  - id: "dup"
    type: stdout
`)
	_, err := Load(path)
	if err == nil {
		t.Error("expected error for duplicate notifier id, got nil")
	}
}

func TestLoad_UnknownNotifierType(t *testing.T) {
	path := writeConfig(t, `
notifiers:
  - id: "x"
    type: unknown
`)
	_, err := Load(path)
	if err == nil {
		t.Error("expected error for unknown notifier type, got nil")
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/config.yml")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	path := writeConfig(t, `not: valid: yaml: [`)
	_, err := Load(path)
	if err == nil {
		t.Error("expected error for invalid YAML, got nil")
	}
}

// --- resolveEnvVars tests ---

func TestResolveEnvVars_Set(t *testing.T) {
	t.Setenv("TEST_VAR", "hello")
	result := resolveEnvVars("${TEST_VAR}")
	if result != "hello" {
		t.Errorf("expected hello, got %s", result)
	}
}

func TestResolveEnvVars_Unset(t *testing.T) {
	os.Unsetenv("UNSET_VAR")
	result := resolveEnvVars("${UNSET_VAR}")
	if result != "${UNSET_VAR}" {
		t.Errorf("expected placeholder preserved, got %s", result)
	}
}

func TestResolveEnvVars_NoPlaceholder(t *testing.T) {
	result := resolveEnvVars("plainvalue")
	if result != "plainvalue" {
		t.Errorf("expected plainvalue unchanged, got %s", result)
	}
}

// --- resolveValue tests ---

func TestResolveValue_Plain(t *testing.T) {
	t.Setenv("MY_TOKEN", "abc123")
	val, err := resolveValue("token", "${MY_TOKEN}", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "abc123" {
		t.Errorf("expected abc123, got %s", val)
	}
}

func TestResolveValue_File(t *testing.T) {
	path := writeSecretFile(t, "  secret-value\n")
	val, err := resolveValue("token", "", path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "secret-value" {
		t.Errorf("expected secret-value, got %s", val)
	}
}

func TestResolveValue_BothSet(t *testing.T) {
	_, err := resolveValue("token", "plain", "/some/file")
	if err == nil {
		t.Error("expected error when both token and token_file are set, got nil")
	}
}

func TestResolveValue_NeitherSet(t *testing.T) {
	_, err := resolveValue("token", "", "")
	if err == nil {
		t.Error("expected error when neither token nor token_file is set, got nil")
	}
}

func TestResolveValue_UnsetEnvVar(t *testing.T) {
	os.Unsetenv("MISSING_TOKEN")
	_, err := resolveValue("token", "${MISSING_TOKEN}", "")
	if err == nil {
		t.Error("expected error for unset env var, got nil")
	}
}

// --- FindConfigPath tests ---

func TestFindConfigPath_Flag(t *testing.T) {
	result := FindConfigPath("/custom/config.yml")
	if result != "/custom/config.yml" {
		t.Errorf("expected /custom/config.yml, got %s", result)
	}
}

func TestFindConfigPath_EnvVar(t *testing.T) {
	t.Setenv("HELLBOT_CONFIG", "/env/config.yml")
	result := FindConfigPath("")
	if result != "/env/config.yml" {
		t.Errorf("expected /env/config.yml, got %s", result)
	}
}

func TestFindConfigPath_Default(t *testing.T) {
	os.Unsetenv("HELLBOT_CONFIG")
	result := FindConfigPath("")
	if result != defaultConfigPath {
		t.Errorf("expected %s, got %s", defaultConfigPath, result)
	}
}

func TestFindConfigPath_FlagTakesPriorityOverEnv(t *testing.T) {
	t.Setenv("HELLBOT_CONFIG", "/env/config.yml")
	result := FindConfigPath("/flag/config.yml")
	if result != "/flag/config.yml" {
		t.Errorf("expected flag to take priority, got %s", result)
	}
}

// --- Full integration: multiple stdout notifiers ---

func TestLoad_MultipleStdoutNotifiers(t *testing.T) {
	path := writeConfig(t, `
notifiers:
  - id: "stdout-1"
    type: stdout
  - id: "stdout-2"
    type: stdout
    options:
      timezone: "Europe/Lisbon"
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Notifiers) != 2 {
		t.Errorf("expected 2 notifiers, got %d", len(cfg.Notifiers))
	}
}

// ensure temp dir usage doesn't leave artifacts
var _ = filepath.Join
