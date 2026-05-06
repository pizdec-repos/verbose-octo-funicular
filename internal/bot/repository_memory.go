package bot

import (
	"sync"
	"time"
)

type InMemoryRepo struct {
	mu     sync.Mutex
	alerts map[string]*AlertState
}

func NewInMemoryRepo() *InMemoryRepo {
	return &InMemoryRepo{
		alerts: make(map[string]*AlertState),
	}
}

func (r *InMemoryRepo) ShouldNotify(fingerprint string, status string) bool {
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

	if status == "firing" && time.Since(state.LastSentAt) > 1*time.Hour {
		state.LastSentAt = time.Now()
		return true
	}

	return false
}

func (r *InMemoryRepo) MarkFailed(fingerprint string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.alerts, fingerprint)
}

func (r *InMemoryRepo) Cleanup() {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	for fp, state := range r.alerts {
		maxAge := 24 * time.Hour
		if state.Status == "resolved" {
			maxAge = 72 * time.Hour
		}

		if now.Sub(state.LastSentAt) > maxAge {
			delete(r.alerts, fp)
		}
	}
}

func (r *InMemoryRepo) Len() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.alerts)
}

func (r *InMemoryRepo) Stats() map[string]int {
	r.mu.Lock()
	defer r.mu.Unlock()

	stats := map[string]int{
		"total":    0,
		"firing":   0,
		"resolved": 0,
	}

	for _, state := range r.alerts {
		stats["total"]++
		if state.Status == "firing" {
			stats["firing"]++
		} else {
			stats["resolved"]++
		}
	}

	return stats
}
