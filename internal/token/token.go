package token

import (
	"fmt"
	"time"

	"github.com/pizdec-repos/verbose-octo-funicular/internal/config"
	"github.com/pizdec-repos/verbose-octo-funicular/pkg/jwtutil"
)

type Generator interface {
	Generate() (string, error)
	GenerateWithExpiry(expiry time.Duration) (string, error)
	Validate(tokenString string) (*Claims, error)
}

type Claims struct {
	BotID   string `json:"bot_id"`
	Version int    `json:"version"`
	jwtutil.StandardClaims
}

type generator struct {
	config *config.Config
	jwtGen *jwtutil.Generator
}

func NewGenerator(cfg *config.Config) Generator {
	jwtConfig := jwtutil.DefaultConfig([]byte(cfg.SecretKey))
	jwtConfig.DefaultExpiry = cfg.TokenExpiry

	return &generator{
		config: cfg,
		jwtGen: jwtutil.NewGenerator(jwtConfig),
	}
}

func (g *generator) Generate() (string, error) {
	return g.GenerateWithExpiry(0)
}

func (g *generator) GenerateWithExpiry(expiry time.Duration) (string, error) {
	if expiry == 0 {
		expiry = g.config.TokenExpiry
	}
	if expiry == 0 {
		expiry = 10 * time.Minute
	}

	token, _, err := g.jwtGen.GenerateStandard(g.config.BotID, g.config.ExpressAudience, expiry)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (g *generator) Validate(tokenString string) (*Claims, error) {
	claims := &Claims{}
	_, err := jwtutil.ValidateWithClaims(tokenString, []byte(g.config.SecretKey), claims)
	if err != nil {
		return nil, err
	}

	if claims.BotID != g.config.BotID {
		return nil, fmt.Errorf("bot_id mismatch: expected %s, got %s", g.config.BotID, claims.BotID)
	}
	return claims, nil
}
