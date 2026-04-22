package bot

import (
	"context"
	"sync"
	"time"
)

type InMemoryRepo struct {
	mu     sync.RWMutex
	alerts map[string]*AlertState
}

func NewInMemoryRepo() *InMemoryRepo {
	return &InMemoryRepo{
		alerts: make(map[string]*AlertState),
	}
}

func (r *InMemoryRepo) ShouldNotify(ctx context.Context, fingerprint string, status string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	state, exists := r.alerts[fingerprint]
	if !exists || state.Status != status {
		r.alerts[fingerprint] = &AlertState{
			Fingerprint: fingerprint,
			Status:      status,
			LastSentAt:  time.Now(),
		}
		return true
	}

	if time.Since(state.LastSentAt) > 1*time.Hour {
		state.LastSentAt = time.Now()
		return true
	}

	return false
}
