package bot

import (
	"context"
	"errors"
	"net/http"
	"sync"
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
	wg     sync.WaitGroup
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
		Addr:         ":" + s.config.Port,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		s.logger.Info("http server starting", zap.String("port", s.config.Port))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error("http server failed", zap.Error(err))
		}
	}()

	s.logger.Info("bot worker started", zap.String("bot_id", s.config.BotID))

	<-ctx.Done()
	s.logger.Info("shutting down application...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		s.logger.Error("server shutdown failed", zap.Error(err))
	}

	waitCh := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(waitCh)
	}()

	select {
	case <-waitCh:
		s.logger.Info("all workers finished")
	case <-time.After(10 * time.Second):
		s.logger.Warn("shutdown timed out: some workers might still be running")
	}

	s.logger.Info("bot stopped gracefully")
	return nil
}
