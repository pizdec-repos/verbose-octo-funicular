package jwtutil

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Config struct {
	SecretKey     []byte
	SigningMethod jwt.SigningMethod
	DefaultExpiry time.Duration
}

func DefaultConfig(secretKey []byte) *Config {
	return &Config{
		SecretKey:     secretKey,
		SigningMethod: jwt.SigningMethodHS256,
		DefaultExpiry: 10 * time.Minute,
	}
}

type StandardClaims struct {
	jwt.RegisteredClaims
	JTI     string `json:"jti,omitempty"`
	Version int    `json:"version,omitempty"`
}

type Generator struct {
	config *Config
}

func NewGenerator(config *Config) *Generator {
	return &Generator{
		config: config,
	}
}

func (g *Generator) Generate(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(g.config.SigningMethod, claims)
	return token.SignedString(g.config.SecretKey)
}

func (g *Generator) GenerateStandard(issuer, audience string, expiry time.Duration) (string, *StandardClaims, error) {
	if expiry == 0 {
		expiry = g.config.DefaultExpiry
	}
	now := time.Now()

	claims := &StandardClaims{
		JTI:     uuid.New().String(),
		Version: 1,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			Audience:  jwt.ClaimStrings{audience},
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
		},
	}

	token, err := g.Generate(claims)
	return token, claims, err
}

func (g *Generator) Parse(tokenString string) (jwt.MapClaims, error) {
	parser := jwt.Parser{}
	claims := jwt.MapClaims{}
	_, _, err := parser.ParseUnverified(tokenString, claims)
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}
	return claims, nil
}

func (g *Generator) Validate(tokenString string) (jwt.MapClaims, error) {
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return g.config.SecretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to validate token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("token is invalid")
	}

	return claims, nil
}

func ValidateWithClaims[T jwt.Claims](tokenString string, secretKey []byte, claims T) (T, error) {
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secretKey, nil
	})

	if err != nil {
		return claims, fmt.Errorf("failed to validate token: %w", err)
	}

	if !token.Valid {
		return claims, fmt.Errorf("token is invalid")
	}

	return claims, nil
}

func GetExpirationTime(tokenString string) (time.Time, error) {
	parser := jwt.Parser{}
	claims := jwt.MapClaims{}
	_, _, err := parser.ParseUnverified(tokenString, claims)
	if err != nil {
		return time.Time{}, err
	}

	if exp, ok := claims["exp"].(float64); ok {
		return time.Unix(int64(exp), 0), nil
	}

	return time.Time{}, fmt.Errorf("no expiration claim found")
}

func IsExpired(tokenString string) (bool, error) {
	exp, err := GetExpirationTime(tokenString)
	if err != nil {
		return false, err
	}
	return time.Now().After(exp), nil
}

func GetJTI(tokenString string) (string, error) {
	parser := jwt.Parser{}
	claims := jwt.MapClaims{}
	_, _, err := parser.ParseUnverified(tokenString, claims)
	if err != nil {
		return "", err
	}

	if jti, ok := claims["jti"].(string); ok {
		return jti, nil
	}

	return "", fmt.Errorf("no jti claim found")
}

func (g *Generator) RefreshToken(oldToken string, newExpiry time.Duration) (string, error) {
	claims, err := g.Validate(oldToken)
	if err != nil {
		return "", fmt.Errorf("invalid old token: %w", err)
	}

	newClaims := jwt.MapClaims{}
	for k, v := range claims {
		newClaims[k] = v
	}

	now := time.Now()
	newClaims["iat"] = now.Unix()
	newClaims["nbf"] = now.Unix()
	newClaims["exp"] = now.Add(newExpiry).Unix()
	newClaims["jti"] = uuid.New().String()

	token := jwt.NewWithClaims(g.config.SigningMethod, newClaims)
	return token.SignedString(g.config.SecretKey)
}

func (g *Generator) MustGenerate(claims jwt.Claims) string {
	token, err := g.Generate(claims)
	if err != nil {
		panic(fmt.Sprintf("failed to generate token: %v", err))
	}
	return token
}
