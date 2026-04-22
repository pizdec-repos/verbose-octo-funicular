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
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	tokenGenerator := token.NewGenerator(cfg)
	botService := bot.NewService(cfg, tokenGenerator)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down gracefully...")
		cancel()
	}()

	if err := botService.Run(ctx); err != nil {
		log.Fatalf("Bot failed: %v", err)
	}
}


