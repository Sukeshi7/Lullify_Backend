package http

import (
	"net/http"

	"Lullify_Backend/internal/infrastructure/observability"
)

func NewRouter(
	auth *AuthHandler,
	stream *StreamHandler,
	playlist *PlaylistHandler,
	track *TrackHandler,
	history *HistoryHandler,
) http.Handler {
	mux := http.NewServeMux()
	auth.Routes(mux)
	stream.Routes(mux)
	playlist.Routes(mux)
	track.Routes(mux)
	history.Routes(mux)

	// /metrics — endpoint Prometheus (non protégé en dev, à sécuriser en prod)
	mux.Handle("GET /metrics", observability.MetricsHandler())

	// Chaîne de middlewares : CORS → OTEL (traces + logs + métriques)
	return corsMiddleware(observability.OtelMiddleware(mux))
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
