package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ametis70/hellbot/internal/adapter/api/helldivers1api"
	mockfetcher "github.com/ametis70/hellbot/internal/adapter/api/mock"
	discordnotifier "github.com/ametis70/hellbot/internal/adapter/notifier/discord"
	"github.com/ametis70/hellbot/internal/adapter/notifier/stdout"
	telegramnotifier "github.com/ametis70/hellbot/internal/adapter/notifier/telegram"
	webhooknotifier "github.com/ametis70/hellbot/internal/adapter/notifier/webhook"
	"github.com/ametis70/hellbot/internal/adapter/store/memory"
	sqlitestore "github.com/ametis70/hellbot/internal/adapter/store/sqlite"
	valkeystore "github.com/ametis70/hellbot/internal/adapter/store/valkey"
	"github.com/ametis70/hellbot/internal/app"
	"github.com/ametis70/hellbot/internal/config"
	"github.com/ametis70/hellbot/internal/port"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	configFlag := flag.String("config", "", "path to config file")
	flag.Parse()

	configPath := config.FindConfigPath(*configFlag)
	cfg, err := config.Load(configPath)
	if err != nil {
		logger.Error("failed to load config", "path", configPath, "error", err)
		os.Exit(1)
	}

	globalTZ, err := time.LoadLocation(cfg.Timezone)
	if err != nil {
		logger.Error("invalid global timezone", "timezone", cfg.Timezone, "error", err)
		os.Exit(1)
	}

	// Build notifiers from config
	var closers []func() error
	notifiers := make([]port.Notifier, 0, len(cfg.Notifiers))

	for _, n := range cfg.Notifiers {
		switch n.Type {
		case config.NotifierTypeStdout:
			opts, err := config.ResolveStdoutOptions(n.Options)
			if err != nil {
				logger.Error("invalid stdout notifier options", "id", n.ID, "error", err)
				os.Exit(1)
			}
			tz := globalTZ
			if opts.Timezone != "" {
				tz, err = time.LoadLocation(opts.Timezone)
				if err != nil {
					logger.Error("invalid timezone", "id", n.ID, "timezone", opts.Timezone, "error", err)
					os.Exit(1)
				}
			}
			notifiers = append(notifiers, stdout.New(stdout.Options{
				Timezone:  tz,
				Templates: opts.Templates,
			}))
			logger.Info("registered notifier", "id", n.ID, "type", n.Type)

		case config.NotifierTypeDiscord:
			opts, err := config.ResolveDiscordOptions(n.Options)
			if err != nil {
				logger.Error("invalid discord notifier options", "id", n.ID, "error", err)
				os.Exit(1)
			}
			dn, err := discordnotifier.New(discordnotifier.Options{
				Token:     opts.Token,
				ChannelID: opts.ChannelID,
				GuildID:   opts.GuildID,
				Templates: opts.Templates,
			}, logger)
			if err != nil {
				logger.Error("failed to create discord notifier", "id", n.ID, "error", err)
				os.Exit(1)
			}
			closers = append(closers, dn.Close)
			notifiers = append(notifiers, dn)
			logger.Info("registered notifier", "id", n.ID, "type", n.Type)

		case config.NotifierTypeTelegram:
			opts, err := config.ResolveTelegramOptions(n.Options)
			if err != nil {
				logger.Error("invalid telegram notifier options", "id", n.ID, "error", err)
				os.Exit(1)
			}
			tz := globalTZ
			if opts.Timezone != "" {
				tz, err = time.LoadLocation(opts.Timezone)
				if err != nil {
					logger.Error("invalid timezone", "id", n.ID, "timezone", opts.Timezone, "error", err)
					os.Exit(1)
				}
			}
			tn, err := telegramnotifier.New(telegramnotifier.Options{
				Token:     opts.Token,
				ChatID:    opts.ChatID,
				Timezone:  tz,
				Templates: opts.Templates,
			}, logger)
			if err != nil {
				logger.Error("failed to create telegram notifier", "id", n.ID, "error", err)
				os.Exit(1)
			}
			notifiers = append(notifiers, tn)
			closers = append(closers, tn.Close)
			logger.Info("registered notifier", "id", n.ID, "type", n.Type)
		case config.NotifierTypeWebhook:
			opts, err := config.ResolveWebhookOptions(n.Options)
			if err != nil {
				logger.Error("invalid webhook notifier options", "id", n.ID, "error", err)
				os.Exit(1)
			}
			wn, err := webhooknotifier.New(webhooknotifier.Options{
				URL:          opts.URL,
				SecretHeader: opts.SecretHeader,
				SecretValue:  opts.SecretValue,
			}, logger)
			if err != nil {
				logger.Error("failed to create webhook notifier", "id", n.ID, "error", err)
				os.Exit(1)
			}
			notifiers = append(notifiers, wn)
			logger.Info("registered notifier", "id", n.ID, "type", n.Type)
		}
	}

	if len(notifiers) == 0 {
		logger.Warn("no notifiers configured — events will be detected but not reported")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Build fetcher
	var fetcher port.Fetcher
	if cfg.Dev.MockServer {
		logger.Warn("dev.mock_server is enabled — using built-in war scenario, not the real API")
		fetcher = mockfetcher.New(cancel, logger)
	} else {
		apiOpts := helldivers1api.DefaultOptions()
		if cfg.Dev.APIURL != "" {
			apiOpts.BaseURL = cfg.Dev.APIURL
			logger.Info("dev.api_url override", "url", apiOpts.BaseURL)
		}
		fetcher = helldivers1api.New(apiOpts, logger)
	}

	// Build store
	var store interface {
		port.CampaignStore
		port.EventStore
	}

	switch cfg.Store.Type {
	case config.StoreTypeValkey:
		opts, err := config.ResolveValkeyStoreOptions(cfg.Store.Options)
		if err != nil {
			logger.Error("invalid valkey store options", "error", err)
			os.Exit(1)
		}
		vs, err := valkeystore.New(valkeystore.Options{
			Addr:     opts.Addr,
			Password: opts.Password,
			DB:       opts.DB,
		})
		if err != nil {
			logger.Error("failed to connect to valkey", "error", err)
			os.Exit(1)
		}
		closers = append(closers, vs.Close)
		store = vs
		logger.Info("store initialized", "type", "valkey", "addr", opts.Addr)
	case config.StoreTypeSQLite:
		opts, err := config.ResolveSQLiteStoreOptions(cfg.Store.Options)
		if err != nil {
			logger.Error("invalid sqlite store options", "error", err)
			os.Exit(1)
		}
		ss, err := sqlitestore.New(sqlitestore.Options{
			Path: opts.Path,
		})
		if err != nil {
			logger.Error("failed to open sqlite store", "error", err)
			os.Exit(1)
		}
		closers = append(closers, ss.Close)
		store = ss
		logger.Info("store initialized", "type", "sqlite", "path", opts.Path)
	default:
		store = memory.New()
		logger.Info("store initialized", "type", "memory")
	}

	// Register interactive commands on notifiers that support them.
	for _, n := range notifiers {
		if c, ok := n.(port.Commander); ok {
			c.RegisterCommands(store)
		}
	}

	poller := app.New(fetcher, store, store, notifiers, cfg.PollInterval, logger)

	logger.Info("hellbot starting", "config", configPath, "poll_interval", cfg.PollInterval)
	if err := poller.Run(ctx); err != nil {
		logger.Error("poller exited with error", "error", err)
		os.Exit(1)
	}

	for _, close := range closers {
		if err := close(); err != nil {
			logger.Error("error closing notifier", "error", err)
		}
	}

	logger.Info("hellbot stopped")
}
