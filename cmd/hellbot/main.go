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
	discordnotifier "github.com/ametis70/hellbot/internal/adapter/notifier/discord"
	"github.com/ametis70/hellbot/internal/adapter/notifier/stdout"
	telegramnotifier "github.com/ametis70/hellbot/internal/adapter/notifier/telegram"
	"github.com/ametis70/hellbot/internal/adapter/store/memory"
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
		}
	}

	if len(notifiers) == 0 {
		logger.Warn("no notifiers configured — events will be detected but not reported")
	}

	fetcher := helldivers1api.New(helldivers1api.DefaultOptions(), logger)
	store := memory.New()

	poller := app.New(fetcher, store, store, notifiers, cfg.PollInterval, logger)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

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
