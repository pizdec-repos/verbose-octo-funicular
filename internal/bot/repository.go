package bot

import (
	"context"
	"time"
)

type AlertState struct {
	Fingerprint string
	Status      string
	LastSentAt  time.Time
}

type Repository interface {
	ShouldNotify(ctx context.Context, fingerprint string, status string) bool
}
