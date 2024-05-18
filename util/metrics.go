package util

import (
	"log/slog"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	duration          prometheus.Summary
	inOrderPercentage prometheus.Gauge
	reqCounter        prometheus.Counter
}

func NewMetrics() *Metrics {
	slog.Info("bruh")
	reg := prometheus.NewRegistry()
	m := &Metrics{
		duration: prometheus.NewSummary(prometheus.SummaryOpts{
			Namespace:  "conalg",
			Name:       "request_duration_sec",
			Help:       "Duration of request from proposal -> decided.",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
	),
	inOrderPercentage: prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "conalg",
		Name: "in_order_percentage",
		Help: "Percentage of requests that have arrived in correct order.",
	}),
	reqCounter: prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "conalg",
		Name: "request_counter",
		Help: "Counter of decided (completely processed) requests.",
	}),
	}
	reg.MustRegister(m.duration)
	reg.MustRegister(m.inOrderPercentage)
	reg.MustRegister(m.reqCounter)

	go runMetricsHandler(reg)

	slog.Info("Initiatied Prometheus metrics: duration . . .")
	return m
}

func runMetricsHandler(reg *prometheus.Registry) {
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	http.ListenAndServe(":9000", nil)
}
