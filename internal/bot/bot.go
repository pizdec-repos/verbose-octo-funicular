package bot

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/pizdec-repos/verbose-octo-funicular/internal/config"
	"github.com/pizdec-repos/verbose-octo-funicular/internal/token"
	"go.uber.org/zap"
)

type Service struct {
	config *config.Config
	token  token.Generator
	logger *zap.Logger
	repo   Repository
}

func NewService(cfg *config.Config, tokenGen token.Generator, logger *zap.Logger, repo Repository) *Service {
	return &Service{
		config: cfg,
		token:  tokenGen,
		logger: logger,
		repo:   repo,
	}
}

func (s *Service) Run(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/webhook/grafana", s.HandleWebhook)

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		s.logger.Info("http server starting", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error("http server failed", zap.Error(err))
		}
	}()

	s.logger.Info("bot worker started", zap.String("bot_id", s.config.BotID))

	<-ctx.Done()

	s.logger.Info("shutting down http server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	s.logger.Info("bot stopped gracefully")
	return nil
}
