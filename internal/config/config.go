package config

import "time"

// NotifierType identifies the kind of notifier.
type NotifierType string

const (
	NotifierTypeStdout NotifierType = "stdout"
)

// RawOptions holds unparsed YAML options for a notifier.
type RawOptions map[string]any

// NotifierConfig represents a single notifier entry in the config file.
type NotifierConfig struct {
	ID      string       `yaml:"id"`
	Type    NotifierType `yaml:"type"`
	Options RawOptions   `yaml:"options"`
}

// StdoutOptions holds parsed options for the stdout notifier.
type StdoutOptions struct {
	Timezone string `yaml:"timezone"`
}

// Config is the top-level configuration structure.
type Config struct {
	PollInterval time.Duration
	Timezone     string           `yaml:"timezone"`
	Notifiers    []NotifierConfig `yaml:"notifiers"`
}

// rawConfig mirrors Config but keeps PollInterval as a string for YAML parsing.
type rawConfig struct {
	PollInterval string           `yaml:"poll_interval"`
	Timezone     string           `yaml:"timezone"`
	Notifiers    []NotifierConfig `yaml:"notifiers"`
}
