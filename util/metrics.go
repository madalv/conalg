package util

import (
	"log/slog"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	duration prometheus.Summary
}

func NewMetrics() *Metrics {
	reg := prometheus.NewRegistry()
	m := &Metrics{
		duration: prometheus.NewSummary(prometheus.SummaryOpts{
			Namespace:  "conalg",
			Name:       "request_duration_sec",
			Help:       "Duration of request from proposal -> decided.",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		}),
	}
	reg.MustRegister(m.duration)

	go runMetricsHandler(reg)

	slog.Info("Initiatied Prometheus metrics: duration . . .")
	return m
}

func runMetricsHandler(reg *prometheus.Registry) {
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	http.ListenAndServe(":9000", nil)
}
