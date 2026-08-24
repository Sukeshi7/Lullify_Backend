package observability

import (
	"context"
	"os"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"
)

var Logger zerolog.Logger

// InitLogger initialise zerolog en JSON structuré.
func InitLogger(serviceName string) {
	Logger = zerolog.New(os.Stdout).
		With().
		Timestamp().
		Str("service", serviceName).
		Logger()

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
}

// FromContext retourne un logger enrichi avec le trace_id et span_id du context.
func FromContext(ctx context.Context) zerolog.Logger {
	span := trace.SpanFromContext(ctx)
	if !span.SpanContext().IsValid() {
		return Logger
	}

	return Logger.With().
		Str("trace_id", span.SpanContext().TraceID().String()).
		Str("span_id", span.SpanContext().SpanID().String()).
		Logger()
}
