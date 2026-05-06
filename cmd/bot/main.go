package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/pizdec-repos/verbose-octo-funicular/internal/bot"
	"github.com/pizdec-repos/verbose-octo-funicular/internal/config"
	"github.com/pizdec-repos/verbose-octo-funicular/internal/token"
	"github.com/pizdec-repos/verbose-octo-funicular/pkg/express"
	"github.com/pizdec-repos/verbose-octo-funicular/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	isDev := cfg.Environment == "development"
	l, err := logger.New(cfg.LogLevel, isDev)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer func() { _ = l.Sync() }()

	l.Info("config loaded",
		zap.String("port", cfg.Port),
		zap.String("express_host", cfg.ExpressHost),
		zap.String("express_chat_id", cfg.ExpressChatID),
		zap.Duration("token_expiry", cfg.TokenExpiry),
		zap.String("environment", cfg.Environment),
	)

	tokenGenerator := token.NewGenerator(cfg)
	tokenGenerator = token.NewCachedGenerator(tokenGenerator)

	alertRepo := bot.NewInMemoryRepo()
	expressClient := express.NewClient(cfg.ExpressHost, cfg.ExpressChatID, tokenGenerator, l)
	botService := bot.NewService(cfg, tokenGenerator, l, alertRepo, expressClient)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		l.Info("received shutdown signal", zap.String("signal", sig.String()))
		cancel()
	}()

	l.Info("starting application")
	if err := botService.Run(ctx); err != nil {
		l.Fatal("application failed", zap.Error(err))
	}
}
