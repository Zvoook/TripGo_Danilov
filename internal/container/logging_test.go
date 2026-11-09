package container_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/Zvoook/TripGo_Danilov/internal/http/middleware"
	"github.com/Zvoook/TripGo_Danilov/internal/infra/telemetry"
)

func TestLogging(t *testing.T) {
	recorder := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	t.Cleanup(func() {
		if err := provider.Shutdown(context.Background()); err != nil {
			t.Errorf("shutdown tracer: %v", err)
		}
	})
	var output bytes.Buffer
	router := chi.NewRouter()
	router.Use(middleware.Tracing(provider))
	router.Use(middleware.Logging(telemetry.NewLogger("info", &output)))

	router.Get("/api/v1/trips/{tripId}", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/trips/123", nil)
	router.ServeHTTP(httptest.NewRecorder(), req)

	t.Log(output.String())

	var entry struct {
		Level   string
		Route   string
		Status  int
		TraceID string `json:"trace_id"`
		SpanID  string `json:"span_id"`
	}
	if err := json.Unmarshal(output.Bytes(), &entry); err != nil {
		t.Fatalf("decode log: %v", err)
	}
	if entry.Level != "INFO" || entry.Route != "/api/v1/trips/{tripId}" || entry.Status != 404 {
		t.Fatalf("unexpected log entry: %+v", entry)
	}
	spans := recorder.Ended()
	if len(spans) != 1 {
		t.Fatalf("recorded %d spans, want 1", len(spans))
	}

	span := spans[0]
	if span.Name() != "GET /api/v1/trips/{tripId}" {
		t.Errorf("span name = %q", span.Name())
	}
	if entry.TraceID != span.SpanContext().TraceID().String() ||
		entry.SpanID != span.SpanContext().SpanID().String() {
		t.Fatalf("log IDs do not match the span: %+v", entry)
	}
}
