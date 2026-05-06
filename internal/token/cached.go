package token

import (
	"sync"
	"time"
)

type cachedToken struct {
	token     string
	expiresAt time.Time
}

type CachedGenerator struct {
	Generator
	mu    sync.RWMutex
	cache cachedToken
}

func NewCachedGenerator(g Generator) Generator {
	return &CachedGenerator{Generator: g}
}

func (c *CachedGenerator) Generate() (string, error) {
	c.mu.RLock()
	if c.cache.token != "" && time.Now().Before(c.cache.expiresAt.Add(-30*time.Second)) {
		token := c.cache.token
		c.mu.RUnlock()
		return token, nil
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cache.token != "" && time.Now().Before(c.cache.expiresAt.Add(-30*time.Second)) {
		return c.cache.token, nil
	}

	token, err := c.Generator.Generate()
	if err != nil {
		return "", err
	}

	c.cache = cachedToken{
		token:     token,
		expiresAt: time.Now().Add(9 * time.Minute),
	}

	return token, nil
}
