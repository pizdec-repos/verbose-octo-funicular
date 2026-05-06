package bot

import "time"

type AlertState struct {
	Fingerprint string
	Status      string
	LastSentAt  time.Time
}

type Repository interface {
	ShouldNotify(fingerprint string, status string) bool
	MarkFailed(fingerprint string)
	Cleanup()
	Len() int
	Stats() map[string]int
}
