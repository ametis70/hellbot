package config

import (
	"time"

	"github.com/ametis70/hellbot/internal/domain"
)

// NotifierType identifies the kind of notifier.
type NotifierType string

const (
	NotifierTypeStdout   NotifierType = "stdout"
	NotifierTypeDiscord  NotifierType = "discord"
	NotifierTypeTelegram NotifierType = "telegram"
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
	Timezone  string            `yaml:"timezone"`
	Templates *domain.Templates `yaml:"templates"`
}

// DiscordOptions holds parsed options for the discord notifier.
// Token and TokenFile are mutually exclusive — exactly one must be set.
// ChannelID and ChannelIDFile are mutually exclusive — exactly one must be set.
// GuildID is optional — when set, slash commands are registered as guild commands
// (instant propagation). When empty, commands are registered globally (up to 1h delay).
type DiscordOptions struct {
	Token         string            `yaml:"token"`
	TokenFile     string            `yaml:"token_file"`
	ChannelID     string            `yaml:"channel_id"`
	ChannelIDFile string            `yaml:"channel_id_file"`
	GuildID       string            `yaml:"guild_id"`
	Templates     *domain.Templates `yaml:"templates"`
}

// TelegramOptions holds parsed options for the telegram notifier.
// Token and TokenFile are mutually exclusive — exactly one must be set.
// ChatID and ChatIDFile are mutually exclusive — exactly one must be set.
type TelegramOptions struct {
	Token      string            `yaml:"token"`
	TokenFile  string            `yaml:"token_file"`
	ChatID     string            `yaml:"chat_id"`
	ChatIDFile string            `yaml:"chat_id_file"`
	Timezone   string            `yaml:"timezone"`
	Templates  *domain.Templates `yaml:"templates"`
}

// StoreType identifies the kind of backing store.
type StoreType string

const (
	StoreTypeMemory StoreType = "memory"
	StoreTypeValkey StoreType = "valkey"
	StoreTypeSQLite StoreType = "sqlite"
)

// SQLiteStoreOptions holds the path for the SQLite store.
type SQLiteStoreOptions struct {
	// Path is the file path for the SQLite database (e.g. "./hellbot.db").
	// Defaults to "hellbot.db" in the working directory.
	Path string `yaml:"path"`
}

// ValkeyStoreOptions holds connection parameters for the Valkey/Redis store.
type ValkeyStoreOptions struct {
	// Addr is the host:port of the server (default: "localhost:6379").
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

// StoreConfig holds the store type and its raw options.
type StoreConfig struct {
	Type    StoreType  `yaml:"type"`
	Options RawOptions `yaml:"options"`
}

// Config is the top-level configuration structure.
type Config struct {
	PollInterval time.Duration
	Timezone     string           `yaml:"timezone"`
	Store        StoreConfig      `yaml:"store"`
	Notifiers    []NotifierConfig `yaml:"notifiers"`
}

// rawConfig mirrors Config but keeps PollInterval as a string for YAML parsing.
type rawConfig struct {
	PollInterval string           `yaml:"poll_interval"`
	Timezone     string           `yaml:"timezone"`
	Store        StoreConfig      `yaml:"store"`
	Notifiers    []NotifierConfig `yaml:"notifiers"`
}
