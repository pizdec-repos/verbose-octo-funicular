package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/pizdec-repos/verbose-octo-funicular/internal/alert"
	"go.uber.org/zap"
)

func (s *Service) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.logger.Error("failed to read body", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var payload alert.GrafanaWebhook
	if err := json.Unmarshal(body, &payload); err != nil {
		s.logger.Error("invalid payload", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(payload.Alerts) == 0 {
		s.logger.Warn("received webhook with empty alerts list")
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	for _, a := range payload.Alerts {
		alertsReceived.WithLabelValues(a.Status).Inc()
	}

	sendCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	const maxConcurrent = 10
	sem := make(chan struct{}, maxConcurrent)
	var accepted int64

	for _, a := range payload.Alerts {
		if !s.repo.ShouldNotify(a.Fingerprint, a.Status) {
			s.logger.Debug("alert skipped",
				zap.String("fingerprint", a.Fingerprint),
				zap.String("reason", "deduplication"))
			continue
		}

		s.wg.Add(1)
		go func(a alert.Alert) {
			defer s.wg.Done()

			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-time.After(5 * time.Second):
				s.logger.Warn("send queue full, dropping alert",
					zap.String("fingerprint", a.Fingerprint))
				s.repo.MarkFailed(a.Fingerprint)
				alertsFailed.Inc()
				return
			}

			text := alert.Format(a)

			if err := s.express.SendAlert(sendCtx, text); err != nil {
				s.logger.Error("failed to send alert",
					zap.String("fingerprint", a.Fingerprint),
					zap.Error(err))
				s.repo.MarkFailed(a.Fingerprint)
				alertsFailed.Inc()
				return
			}

			atomic.AddInt64(&accepted, 1)
			alertsSent.WithLabelValues(a.Status).Inc()

			s.logger.Info("alert sent",
				zap.String("fingerprint", a.Fingerprint),
				zap.String("status", a.Status))
		}(a)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)

	total := len(payload.Alerts)
	resp := fmt.Sprintf(`{"accepted":%d,"total":%d}`, atomic.LoadInt64(&accepted), total)
	if _, err := fmt.Fprint(w, resp); err != nil {
		s.logger.Error("failed to write response", zap.Error(err))
	}
}
