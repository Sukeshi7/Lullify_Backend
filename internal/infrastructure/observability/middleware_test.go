package observability_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"Lullify_Backend/internal/infrastructure/observability"
)

func TestInitLogger(t *testing.T) {
	observability.InitLogger("test-service")
}

func TestFromContext_NoSpan(t *testing.T) {
	observability.InitLogger("test")
	log := observability.FromContext(context.Background())
	log.Info().Msg("test log from context without span")
}

func TestInitTracer_EmptyEndpoint(t *testing.T) {
	ctx := context.Background()
	shutdown, err := observability.InitTracer(ctx, "test-service", "")
	if err != nil {
		t.Fatalf("expected no error with empty endpoint, got %v", err)
	}
	if shutdown == nil {
		t.Fatal("expected shutdown function, got nil")
	}
	if err := shutdown(ctx); err != nil {
		t.Fatalf("expected no error on shutdown, got %v", err)
	}
}

func TestInitTracer_WithEndpoint(t *testing.T) {
	ctx := context.Background()
	shutdown, err := observability.InitTracer(ctx, "test", "localhost:4318")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	_ = shutdown(ctx)
}

func TestMetricsHandler(t *testing.T) {
	handler := observability.MetricsHandler()
	if handler == nil {
		t.Fatal("expected non-nil metrics handler")
	}
}

func TestOtelMiddleware_PassThrough(t *testing.T) {
	observability.InitLogger("test")

	called := false
	handler := observability.OtelMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/streams", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if !called {
		t.Error("expected inner handler to be called")
	}
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestOtelMiddleware_WithTraceHeader(t *testing.T) {
	observability.InitLogger("test")

	handler := observability.OtelMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/streams", nil)
	req.Header.Set("traceparent", "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}
}

func TestOtelMiddleware_404(t *testing.T) {
	observability.InitLogger("test")

	handler := observability.OtelMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))

	req := httptest.NewRequest(http.MethodGet, "/not-found", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}
