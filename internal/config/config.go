package config

import "time"

// NotifierType identifies the kind of notifier.
type NotifierType string

const (
	NotifierTypeStdout  NotifierType = "stdout"
	NotifierTypeDiscord NotifierType = "discord"
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

// DiscordOptions holds parsed options for the discord notifier.
// Token and TokenFile are mutually exclusive — exactly one must be set.
// ChannelID and ChannelIDFile are mutually exclusive — exactly one must be set.
type DiscordOptions struct {
	Token         string `yaml:"token"`
	TokenFile     string `yaml:"token_file"`
	ChannelID     string `yaml:"channel_id"`
	ChannelIDFile string `yaml:"channel_id_file"`
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
