package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	defaultPollInterval = 60 * time.Second
	defaultTimezone     = "UTC"
	defaultConfigPath   = "config.yml"
)

var envVarPattern = regexp.MustCompile(`\$\{([^}]+)\}`)

// resolveEnvVars replaces ${VAR} patterns with environment variable values.
// If the variable is not set, the placeholder is left as-is.
func resolveEnvVars(s string) string {
	return envVarPattern.ReplaceAllStringFunc(s, func(match string) string {
		name := match[2 : len(match)-1]
		if val, ok := os.LookupEnv(name); ok {
			return val
		}
		return match
	})
}

// readSecretFile reads a file and returns its contents trimmed of whitespace.
func readSecretFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading secret file %q: %w", path, err)
	}
	return strings.TrimSpace(string(data)), nil
}

// resolveValue resolves a plain value (with env var interpolation) or a file value.
// plain and file are mutually exclusive — exactly one must be non-empty.
// fieldName is used in error messages (e.g. "token", "channel_id").
func resolveValue(fieldName, plain, file string) (string, error) {
	if plain != "" && file != "" {
		return "", fmt.Errorf("%s and %s_file are mutually exclusive", fieldName, fieldName)
	}
	if plain == "" && file == "" {
		return "", fmt.Errorf("one of %s or %s_file is required", fieldName, fieldName)
	}
	if file != "" {
		return readSecretFile(file)
	}
	resolved := resolveEnvVars(plain)
	if resolved == plain && envVarPattern.MatchString(plain) {
		return "", fmt.Errorf("%s references an unset environment variable: %s", fieldName, plain)
	}
	return resolved, nil
}

// parseTimezone parses a timezone string into a *time.Location.
// Falls back to UTC if the string is empty.
func parseTimezone(tz string) (*time.Location, error) {
	if tz == "" {
		return time.UTC, nil
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return nil, fmt.Errorf("invalid timezone %q: %w", tz, err)
	}
	return loc, nil
}

// ResolveStdoutOptions decodes and validates stdout notifier options.
func ResolveStdoutOptions(raw RawOptions) (StdoutOptions, error) {
	opts := StdoutOptions{}
	if raw == nil {
		return opts, nil
	}

	data, err := yaml.Marshal(raw)
	if err != nil {
		return opts, fmt.Errorf("marshaling stdout options: %w", err)
	}
	if err := yaml.Unmarshal(data, &opts); err != nil {
		return opts, fmt.Errorf("parsing stdout options: %w", err)
	}
	return opts, nil
}

// ResolveDiscordOptions decodes, validates, and resolves discord notifier options.
func ResolveDiscordOptions(raw RawOptions) (DiscordOptions, error) {
	opts := DiscordOptions{}
	if raw == nil {
		return opts, fmt.Errorf("discord options are required")
	}

	data, err := yaml.Marshal(raw)
	if err != nil {
		return opts, fmt.Errorf("marshaling discord options: %w", err)
	}
	if err := yaml.Unmarshal(data, &opts); err != nil {
		return opts, fmt.Errorf("parsing discord options: %w", err)
	}

	token, err := resolveValue("token", opts.Token, opts.TokenFile)
	if err != nil {
		return opts, fmt.Errorf("discord token: %w", err)
	}
	opts.Token = token
	opts.TokenFile = ""

	channelID, err := resolveValue("channel_id", opts.ChannelID, opts.ChannelIDFile)
	if err != nil {
		return opts, fmt.Errorf("discord channel_id: %w", err)
	}
	opts.ChannelID = channelID
	opts.ChannelIDFile = ""

	return opts, nil
}

// ResolveStdoutOptions decodes and validates stdout notifier options., resolves env vars and secrets,
// validates all notifier options, and returns a parsed Config.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file %q: %w", path, err)
	}

	raw := rawConfig{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	cfg := &Config{
		Timezone:  raw.Timezone,
		Notifiers: raw.Notifiers,
	}

	// Parse poll interval
	if raw.PollInterval == "" {
		cfg.PollInterval = defaultPollInterval
	} else {
		d, err := time.ParseDuration(raw.PollInterval)
		if err != nil {
			return nil, fmt.Errorf("invalid poll_interval %q: %w", raw.PollInterval, err)
		}
		cfg.PollInterval = d
	}

	// Apply default timezone
	if cfg.Timezone == "" {
		cfg.Timezone = defaultTimezone
	}

	// Validate global timezone is parseable
	if _, err := parseTimezone(cfg.Timezone); err != nil {
		return nil, err
	}

	// Validate all notifiers
	ids := make(map[string]struct{})
	for i, n := range cfg.Notifiers {
		if n.ID == "" {
			return nil, fmt.Errorf("notifier[%d]: id is required", i)
		}
		if _, dup := ids[n.ID]; dup {
			return nil, fmt.Errorf("notifier[%d]: duplicate id %q", i, n.ID)
		}
		ids[n.ID] = struct{}{}

		switch n.Type {
		case NotifierTypeStdout:
			if _, err := ResolveStdoutOptions(n.Options); err != nil {
				return nil, fmt.Errorf("notifier %q: %w", n.ID, err)
			}
		case NotifierTypeDiscord:
			if _, err := ResolveDiscordOptions(n.Options); err != nil {
				return nil, fmt.Errorf("notifier %q: %w", n.ID, err)
			}
		default:
			return nil, fmt.Errorf("notifier %q: unknown type %q", n.ID, n.Type)
		}
	}

	return cfg, nil
}

// FindConfigPath returns the config file path using the following priority:
// 1. Explicit path argument (from --config flag)
// 2. HELLBOT_CONFIG environment variable
// 3. Default path (./config.yml)
func FindConfigPath(flagValue string) string {
	if flagValue != "" {
		return flagValue
	}
	if env := os.Getenv("HELLBOT_CONFIG"); env != "" {
		return env
	}
	return defaultConfigPath
}
