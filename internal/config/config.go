package config

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	BotID       string        `env:"BOT_ID" env-required:"true"`
	SecretKey   string        `env:"SECRET_KEY" env-required:"true"`
	TokenExpiry time.Duration `env:"TOKEN_EXPIRY" env-default:"10m"`

	LogLevel    string `env:"LOG_LEVEL" env-default:"info"`
	Environment string `env:"ENVIRONMENT" env-default:"development"`
	Port        string `env:"PORT" env-default:"8080"`

	ExpressHost     string `env:"EXPRESS_HOST" env-required:"true"`
	ExpressAudience string `env:"EXPRESS_AUDIENCE" env-required:"true"`
	ExpressChatID   string `env:"EXPRESS_GROUP_CHAT_ID" env-required:"true"`

	GrafanaSecret string `env:"GRAFANA_SECRET"`
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	if _, err := uuid.Parse(cfg.ExpressChatID); err != nil {
		return nil, fmt.Errorf("EXPRESS_GROUP_CHAT_ID must be valid UUID: %w", err)
	}

	return &cfg, nil
}
