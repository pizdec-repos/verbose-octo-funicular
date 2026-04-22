package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	BotID       string
	SecretKey   []byte
	TokenExpiry time.Duration
	LogLevel string
}

type Option func(*Config)

func WithTokenExpiry(expiry time.Duration) Option {
	return func(c *Config) {
		c.TokenExpiry = expiry
	}
}

func Load(opts ...Option) (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		BotID: getEnv("BOT_ID", ""),
		SecretKey: []byte(getEnv("SECRET_KEY", "")),
		TokenExpiry: getDurationEnv("TOKEN_EXPIRY_MINUTES", 10*time.Minute),
		LogLevel: getEnv("LOG_LEVEL", "info"),
	}

	for _, opt := range opts {
		opt(cfg)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.BotID == "" {
		return fmt.Errorf("BOT_ID is required")
	}

	if string(c.SecretKey) == "" {
		return fmt.Errorf("SECRET_KEY is required")
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if minutes, err := strconv.Atoi(value); err == nil {
			return time.Duration(minutes) * time.Minute
		}
	}
	return defaultValue
}
