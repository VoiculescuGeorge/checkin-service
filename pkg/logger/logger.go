package logger

import (
	"context"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/trace"
)

// Setup configures the global zerolog logger.
func Setup(isLocalDev bool) {
	// Use Unix timestamps for performance and consistency
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	if isLocalDev {
		// Pretty printing for local development
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		// Default to JSON output for production
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}

// EnrichContextWithLogger adds a zerolog logger to the context with trace information.
func EnrichContextWithLogger(ctx context.Context) context.Context {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return ctx
	}

	sCtx := span.SpanContext()
	if !sCtx.HasTraceID() {
		return ctx
	}

	l := log.With().
		Str("trace_id", sCtx.TraceID().String()).
		Str("span_id", sCtx.SpanID().String()).
		Logger()

	return l.WithContext(ctx)
}
