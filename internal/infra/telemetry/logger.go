package telemetry

import (
	"context"
	"io"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

func NewLogger(inputLevel string, out io.Writer) *slog.Logger {
	var level slog.Level
	if err := level.UnmarshalText([]byte(inputLevel)); err != nil {
		level = slog.LevelInfo
	}

	handler := slog.NewJSONHandler(out, &slog.HandlerOptions{Level: level})
	return slog.New(traceLogHandler{Handler: handler}).With("service", "trip-service")
}

type traceLogHandler struct {
	slog.Handler
}

func (h traceLogHandler) Handle(ctx context.Context, record slog.Record) error {
	span := trace.SpanContextFromContext(ctx)
	if span.IsValid() {
		record.AddAttrs(slog.String("trace_id", span.TraceID().String()), slog.String("span_id", span.SpanID().String()))
	}
	return h.Handler.Handle(ctx, record)
}

func (h traceLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return traceLogHandler{Handler: h.Handler.WithAttrs(attrs)}
}

func (h traceLogHandler) WithGroup(name string) slog.Handler {
	return traceLogHandler{Handler: h.Handler.WithGroup(name)}
}
