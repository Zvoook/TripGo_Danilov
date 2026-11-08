package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"runtime/debug"

	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/Zvoook/TripGo_Danilov/internal/http/domainerror"
)

type recoveryKey struct{}

func Recover(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			wrapped := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)
			defer func() {
				if failure := recover(); failure != nil {
					if failure == http.ErrAbortHandler {
						panic(failure)
					}
					logger.ErrorContext(ctx, "panic in HTTP request", "stack", string(debug.Stack()))
					if wrapped.Status() == 0 {
						domainerror.WriteProblem(wrapped, r, domainerror.Problem(domainerror.ErrInternal))
					}
				}
			}()
			// Tracing updates this local context once the server span has started.
			next.ServeHTTP(wrapped, r.WithContext(context.WithValue(ctx, recoveryKey{}, &ctx)))
		})
	}
}
