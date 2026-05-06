package bot

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	alertsReceived = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "grafana_alerts_received_total",
		Help: "Total alerts received from Grafana",
	}, []string{"status"})

	alertsSent = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "grafana_alerts_sent_total",
		Help: "Total alerts successfully sent",
	}, []string{"status"})

	alertsFailed = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "grafana_alerts_failed_total",
		Help: "Total alerts failed to send",
	})
)

func init() {
	prometheus.MustRegister(alertsReceived, alertsSent, alertsFailed)
}
