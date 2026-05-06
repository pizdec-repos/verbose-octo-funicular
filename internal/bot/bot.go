package bot

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/pizdec-repos/verbose-octo-funicular/internal/config"
	"github.com/pizdec-repos/verbose-octo-funicular/internal/token"
	"github.com/pizdec-repos/verbose-octo-funicular/pkg/express"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

type Service struct {
	config  *config.Config
	token   token.Generator
	logger  *zap.Logger
	repo    Repository
	express *express.Client
	wg      sync.WaitGroup
}

func NewService(cfg *config.Config, tokenGen token.Generator, logger *zap.Logger, repo Repository, express *express.Client) *Service {
	return &Service{
		config:  cfg,
		token:   tokenGen,
		logger:  logger,
		repo:    repo,
		express: express,
	}
}

func (s *Service) Run(ctx context.Context) error {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", s.HealthHandler)
	mux.HandleFunc("/ready", s.ReadyHandler)
	mux.HandleFunc("/webhook/grafana", s.HandleWebhook)
	mux.Handle("/metrics", promhttp.Handler())

	srv := &http.Server{
		Addr:              ":" + s.config.Port,
		Handler:           mux,
		ReadHeaderTimeout: 2 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
	}

	go func() {
		s.logger.Info("http server starting", zap.String("port", s.config.Port))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error("http server failed", zap.Error(err))
		}
	}()

	cleanupTicker := time.NewTicker(1 * time.Hour)
	defer cleanupTicker.Stop()

	var cleanupWg sync.WaitGroup
	cleanupWg.Add(1)
	go func() {
		defer cleanupWg.Done()
		for {
			select {
			case <-cleanupTicker.C:
				s.repo.Cleanup()
				s.logger.Debug("alert repository cleaned up")
			case <-ctx.Done():
				return
			}
		}
	}()

	s.logger.Info("bot started",
		zap.String("bot_id", s.config.BotID),
		zap.String("port", s.config.Port),
	)

	<-ctx.Done()
	s.logger.Info("shutting down application...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		s.logger.Error("server shutdown failed", zap.Error(err))
	}

	cleanupWg.Wait()

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		s.logger.Info("all workers finished")
	case <-time.After(15 * time.Second):
		s.logger.Warn("shutdown timed out: some workers might still be running")
	}

	s.logger.Info("bot stopped gracefully")
	return nil
}

func (s *Service) HealthHandler(w http.ResponseWriter, r *http.Request) {
	select {
	case <-r.Context().Done():
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	default:
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func (s *Service) ReadyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ready"}`))
}

func (s *Service) MetricsHandler(w http.ResponseWriter, r *http.Request) {
	stats := s.repo.Stats()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		s.logger.Error("failed to encode metrics", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
	}
}
