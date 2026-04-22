package bot

import (
	"encoding/json"
	"net/http"

	"github.com/pizdec-repos/verbose-octo-funicular/internal/alert"
	"go.uber.org/zap"
)

func (s *Service) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var payload alert.GrafanaWebhook
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		s.logger.Error("invalid payload", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(payload.Alerts) == 0 {
		s.logger.Warn("received webhook with empty alerts list")
		w.WriteHeader(http.StatusUnprocessableEntity) // 422 или 400
		return
	}

	for _, a := range payload.Alerts {
		if s.repo.ShouldNotify(r.Context(), a.Fingerprint, a.Status) {
			s.logger.Info("alert triggered",
				zap.String("id", a.Fingerprint),
				zap.String("status", a.Status),
			)
		}
	}

	w.WriteHeader(http.StatusAccepted)
}
