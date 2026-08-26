package observability

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Métriques métier Lullify
var (
	ActiveStreams = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "lullify_active_streams",
		Help: "Number of currently live streams",
	})

	ActiveListeners = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "lullify_active_listeners",
		Help: "Number of currently connected listeners",
	})

	StreamDisconnections = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "lullify_stream_disconnections_total",
		Help: "Total number of stream disconnections",
	}, []string{"type"}) // type: "normal" ou "abrupt"

	AudioBitrateBytes = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "lullify_audio_bitrate_bytes",
		Help:    "Audio chunk size in bytes per stream",
		Buckets: prometheus.ExponentialBuckets(1024, 2, 10),
	}, []string{"stream_id"})

	HTTPRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "lullify_http_request_duration_seconds",
		Help:    "HTTP request duration in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path", "status"})
)

// MetricsHandler retourne le handler Prometheus à monter sur /metrics
func MetricsHandler() http.Handler {
	return promhttp.Handler()
}
