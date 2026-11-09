package container_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/Zvoook/TripGo_Danilov/internal/http/middleware"
	"github.com/Zvoook/TripGo_Danilov/internal/infra/telemetry"
)

func TestRecover(t *testing.T) {
	recorder := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	var output bytes.Buffer
	logger := telemetry.NewLogger("info", &output)
	router := chi.NewRouter()
	router.Use(middleware.Recover(logger), middleware.Tracing(provider), middleware.Logging(logger))
	router.Get("/panic", func(http.ResponseWriter, *http.Request) {
		panic("example")
	})
	router.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/panic", nil))
	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", response.Code)
	}
	var problem struct {
		Code string
	}
	if err := json.Unmarshal(response.Body.Bytes(), &problem); err != nil || problem.Code != "internal_error" {
		t.Fatalf("panic response = %q, want internal_error", response.Body.String())
	}
	spans := recorder.Ended()
	if len(spans) != 1 {
		t.Fatalf("recorded %d spans, want one", len(spans))
	}
	traceID := spans[0].SpanContext().TraceID().String()
	if log := output.String(); !strings.Contains(log, "panic in HTTP request") ||
		!strings.Contains(log, "goroutine") || strings.Count(log, traceID) != 2 {
		t.Errorf("panic and access logs should share the trace ID: %s", log)
	}

	response = httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/health", nil))
	if response.Code != http.StatusOK {
		t.Errorf("status after panic = %d, want 200", response.Code)
	}
}
