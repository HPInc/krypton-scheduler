package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// RegisterPrometheusMetrics - register prometheus metrics.
func RegisterPrometheusMetrics() {
	prometheus.MustRegister(MetricRestLatency)
}

func ReportLatencyMetric(metric *prometheus.SummaryVec,
	startTime time.Time, label string) {
	duration := time.Since(startTime)
	metric.WithLabelValues(label).Observe(float64(duration.Milliseconds()))
}
