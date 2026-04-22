package bot

import (
	"encoding/json"
	"net/http"

	"github.com/pizdec-repos/verbose-octo-funicular/internal/alert"
	"go.uber.org/zap"
)

func (s *Service) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	var payload alert.GrafanaWebhook
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		s.logger.Error("invalid payload", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	for _, a := range payload.Alerts {
		if s.repo.ShouldNotify(r.Context(), a.Fingerprint, a.Status) {
			s.logger.Info("alert triggered", zap.String("id", a.Fingerprint))
		}
	}
	w.WriteHeader(http.StatusAccepted)
}
