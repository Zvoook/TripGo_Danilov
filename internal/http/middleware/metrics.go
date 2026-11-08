package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

func Metrics(router chi.Routes, meter metric.Meter) (func(http.Handler) http.Handler, error) {
	duration, err := meter.Float64Histogram("http.server.request.duration", metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10))
	if err != nil {
		return nil, fmt.Errorf("create HTTP duration histogram: %w", err)
	}
	active, err := meter.Int64UpDownCounter("http.server.active_requests", metric.WithUnit("{request}"))
	if err != nil {
		return nil, fmt.Errorf("create active requests counter: %w", err)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			path := r.URL.RawPath
			if path == "" {
				path = r.URL.Path
			}
			route := router.Find(chi.NewRouteContext(), r.Method, path)
			if route == "" {
				route = unmatchedRoute
			}
			labels := []attribute.KeyValue{
				attribute.String("http.request.method", r.Method),
				attribute.String("http.route", route),
			}

			start := time.Now()
			wrapped := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)
			active.Add(ctx, 1, metric.WithAttributes(labels...))
			finished := false
			defer func() {
				active.Add(ctx, -1, metric.WithAttributes(labels...))
				status := responseStatus(wrapped.Status(), finished)
				duration.Record(ctx, time.Since(start).Seconds(),
					metric.WithAttributes(append(labels, attribute.Int("http.response.status_code", status))...))
			}()
			next.ServeHTTP(wrapped, r)
			finished = true
		})
	}, nil
}

const unmatchedRoute = "unmatched"

func responseStatus(status int, finished bool) int {
	if status != 0 {
		return status
	}
	if !finished {
		return http.StatusInternalServerError
	}
	return http.StatusOK
}
