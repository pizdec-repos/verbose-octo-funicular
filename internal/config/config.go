package config

import (
	"fmt"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	BotID       string        `env:"BOT_ID" env-required:"true"`
	SecretKey   string        `env:"SECRET_KEY" env-required:"true"`
	TokenExpiry time.Duration `env:"TOKEN_EXPIRY_MINUTES" env-default:"10m"`
	LogLevel    string        `env:"LOG_LEVEL" env-default:"info"`
	Environment string        `env:"ENVIRONMENT" env-default:"development"`
	Port        string        `env:"PORT" env-default:"8080"`
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	return &cfg, nil
}
