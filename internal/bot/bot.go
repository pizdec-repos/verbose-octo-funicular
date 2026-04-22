package bot

import (
	"context"
	"fmt"
	"log"

	"github.com/pizdec-repos/verbose-octo-funicular/internal/config"
	"github.com/pizdec-repos/verbose-octo-funicular/internal/token"
)

type Service struct {
	config *config.Config
	token  token.Generator
}

func NewService(cfg *config.Config, tokenGen token.Generator) *Service {
	return &Service{
		config: cfg,
		token:  tokenGen,
	}
}

func (s *Service) Run(ctx context.Context) error {
	token, err := s.token.Generate()

	if err != nil {
		return fmt.Errorf("failed to generate token: %w", err)
	}

	log.Printf("Bot started with ID: %s", s.config.BotID)
	log.Printf("Generated token: %s", token)

	<-ctx.Done()
	log.Println("Bot stopped")
	return nil
}
