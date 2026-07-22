package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ametis70/hellbot/internal/adapter/api/helldivers1api"
	"github.com/ametis70/hellbot/internal/adapter/notifier/stdout"
	"github.com/ametis70/hellbot/internal/adapter/store/memory"
	"github.com/ametis70/hellbot/internal/app"
	"github.com/ametis70/hellbot/internal/port"
)

func main() {
	tz := os.Getenv("TZ")
	loc, err := time.LoadLocation(tz)
	if err != nil || tz == "" {
		loc = time.UTC
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	fetcher := helldivers1api.New(helldivers1api.DefaultOptions(), logger)

	store := memory.New()

	notifiers := []port.Notifier{
		stdout.New(stdout.Options{Timezone: loc}),
	}

	poller := app.New(
		fetcher, store, store, notifiers, 60*time.Second, logger,
	)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	logger.Info("hellbot starting")
	if err := poller.Run(ctx); err != nil {
		logger.Error("poller exited with error", "error", err)
		os.Exit(1)
	}
	logger.Info("hellbot stopped")
}
