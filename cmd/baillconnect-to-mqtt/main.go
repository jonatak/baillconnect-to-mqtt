package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/jonatak/baillconnect-to-mqtt/internal/bootstrap"
	"github.com/jonatak/baillconnect-to-mqtt/internal/config"
)

func main() {
	configPath := flag.String("config", "", "path to configuration file")
	flag.Parse()

	slog.Info(fmt.Sprintf("Start baillconnect-to-mqtt version:%s, commit:%s, buildtime:%s", config.Version, config.CommitSHA, config.BuildTime))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	slog.Info("loading configuration")
	cfg, err := config.Load(*configPath)
	if err != nil {
		slog.Error("configuration failed", "error", err)
		os.Exit(1)
	}

	slog.Info("connecting to baillconnect")
	service, err := bootstrap.NewHVACService(ctx, cfg)

	if err != nil {
		slog.Error("baillconnect service initialization failed", "error", err)
		os.Exit(1)
	}

	slog.Info("initializing mqtt server")
	server, err := bootstrap.NewMQTTServer(ctx, service, cfg)
	if err != nil {
		slog.Error("mqtt server initialization failed", "error", err)
		os.Exit(1)
	}

	slog.Info("running mqtt processor")
	if err := server.Run(ctx); err != nil {
		if !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
			slog.Error("mqtt processor failed", "error", err)
			os.Exit(1)
		}
		slog.Info("mqtt processor stopped", "error", err)
	}
	slog.Info("exited")
}
