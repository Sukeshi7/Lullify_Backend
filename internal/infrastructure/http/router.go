package http

import (
	"net/http"
	"os"

	"Lullify_Backend/internal/infrastructure/observability"
)

func NewRouter(
	auth *AuthHandler,
	stream *StreamHandler,
	playlist *PlaylistHandler,
	track *TrackHandler,
	history *HistoryHandler,
	favorite *FavoriteHandler,
	admin *AdminHandler,
	health *HealthHandler,
	allowedOrigins []string,
) http.Handler {
	mux := http.NewServeMux()
	auth.Routes(mux)
	stream.Routes(mux)
	playlist.Routes(mux)
	track.Routes(mux)
	history.Routes(mux)
	favorite.Routes(mux)
	admin.Routes(mux)
	health.Routes(mux)

	// /metrics — protégé admin
	mux.Handle("GET /metrics", admin.MetricsMiddleware(observability.MetricsHandler()))

	// Path secret pour Prometheus scraping (sans auth, path non documenté)
	if secretPath := os.Getenv("METRICS_SECRET_PATH"); secretPath != "" {
		mux.Handle("GET "+secretPath, observability.MetricsHandler())
	}

	return corsMiddleware(allowedOrigins)(observability.OtelMiddleware(mux))
}

func corsMiddleware(allowedOrigins []string) func(http.Handler) http.Handler {
	allowed := make(map[string]bool, len(allowedOrigins))
	for _, o := range allowedOrigins {
		allowed[o] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			if origin != "" && allowed[origin] {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Vary", "Origin")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
