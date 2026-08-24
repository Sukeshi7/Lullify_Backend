package observability

import (
	"fmt"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
)

var tracer = otel.Tracer("lullify/http")

// OtelMiddleware instrumente chaque requête HTTP avec une trace OTEL
// et enregistre la durée dans Prometheus.
func OtelMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extrait le contexte de trace depuis les headers entrants
		ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))

		spanName := fmt.Sprintf("%s %s", r.Method, r.URL.Path)
		ctx, span := tracer.Start(ctx, spanName)
		defer span.End()

		span.SetAttributes(
			attribute.String("http.method", r.Method),
			attribute.String("http.path", r.URL.Path),
			attribute.String("http.remote_addr", r.RemoteAddr),
		)

		// Writer qui capture le status code
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}

		start := time.Now()
		next.ServeHTTP(rw, r.WithContext(ctx))
		duration := time.Since(start)

		span.SetAttributes(attribute.Int("http.status_code", rw.status))

		// Enregistre dans Prometheus
		HTTPRequestDuration.WithLabelValues(
			r.Method,
			r.URL.Path,
			fmt.Sprintf("%d", rw.status),
		).Observe(duration.Seconds())

		// Log structuré JSON avec trace_id
		log := FromContext(ctx)
		log.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", rw.status).
			Dur("duration_ms", duration).
			Msg("request")
	})
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}
