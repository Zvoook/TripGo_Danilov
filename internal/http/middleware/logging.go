package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/Zvoook/TripGo_Danilov/internal/http/requestlog"
)

func Logging(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			wrapped := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)

			finished := false
			defer func() {
				status := responseStatus(wrapped.Status(), finished)

				route := unmatchedRoute
				if ctx := chi.RouteContext(r.Context()); ctx != nil {
					if pattern := ctx.RoutePattern(); pattern != "" {
						route = pattern
					}
				}

				level := slog.LevelInfo
				if status >= 500 {
					level = slog.LevelError
				}

				attrs := []slog.Attr{
					slog.String("method", r.Method),
					slog.String("route", route),
					slog.Int("status", status),
					slog.Int64("duration_ms", time.Since(start).Milliseconds()),
				}

				logger.LogAttrs(r.Context(), level, "request handled", attrs...)
			}()
			next.ServeHTTP(wrapped, r.WithContext(requestlog.WithLogger(r.Context(), logger)))
			finished = true
		})
	}
}
