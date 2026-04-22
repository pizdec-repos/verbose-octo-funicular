package bot

import (
	"context"
	"fmt"

	"github.com/pizdec-repos/verbose-octo-funicular/internal/config"
	"github.com/pizdec-repos/verbose-octo-funicular/internal/token"
	"go.uber.org/zap"
)

type Service struct {
	config *config.Config
	token  token.Generator
	logger *zap.Logger
}

func NewService(cfg *config.Config, tokenGen token.Generator, logger *zap.Logger) *Service {
	return &Service{
		config: cfg,
		token:  tokenGen,
		logger: logger,
	}
}

func (s *Service) Run(ctx context.Context) error {
	token, err := s.token.Generate()

	if err != nil {
		s.logger.Error("failed to generate token", zap.Error(err))
		return fmt.Errorf("failed to generate token: %w", err)
	}

	s.logger.Info("bot started",
		zap.String("bot_id", s.config.BotID),
		zap.String("token", token),
	)

	<-ctx.Done()
	s.logger.Info("bot stopped")
	return nil
}
