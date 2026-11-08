package requestlog

import (
	"context"
	"log/slog"
)

type loggerKey struct{}

func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

func Error(ctx context.Context, msg string, err error) {
	if logger, ok := ctx.Value(loggerKey{}).(*slog.Logger); ok {
		logger.ErrorContext(ctx, msg, "error", err)
	}
}
