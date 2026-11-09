package middleware

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

func Tracing(provider trace.TracerProvider) func(http.Handler) http.Handler {
	instrument := otelhttp.NewMiddleware("http.request",
		otelhttp.WithTracerProvider(provider),
		otelhttp.WithPropagators(propagation.TraceContext{}),
		otelhttp.WithMeterProvider(noop.NewMeterProvider()),
	)

	return func(next http.Handler) http.Handler {
		return instrument(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if recoveredCtx, ok := r.Context().Value(recoveryKey{}).(*context.Context); ok {
				*recoveredCtx = r.Context()
			}
			wrapped := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)
			finished := false
			defer func() {
				span := trace.SpanFromContext(r.Context())
				if routeCtx := chi.RouteContext(r.Context()); routeCtx != nil {
					if route := routeCtx.RoutePattern(); route != "" {
						span.SetName(r.Method + " " + route)
						span.SetAttributes(attribute.String("http.route", route))
					}
				}
				if status := responseStatus(wrapped.Status(), finished); status >= 500 {
					span.SetStatus(codes.Error, "")
					span.SetAttributes(attribute.String("error.type", strconv.Itoa(status)),
						attribute.Int("http.response.status_code", status))
				}
			}()
			next.ServeHTTP(wrapped, r)
			finished = true
		}))
	}
}
