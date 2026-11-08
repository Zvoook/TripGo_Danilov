package http_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	api "github.com/Zvoook/TripGo_Danilov/internal/generated"
	transport "github.com/Zvoook/TripGo_Danilov/internal/http"
	"github.com/Zvoook/TripGo_Danilov/internal/http/domainerror"
)

func TestHealthEndpoints(t *testing.T) {
	cases := []struct {
		path       string
		wantPing   bool
		pingErr    error
		wantCode   int
		wantStatus string
	}{
		{"/health", false, nil, http.StatusOK, "ok"},
		{"/ready", true, nil, http.StatusOK, "ok"},
		{"/ready", true, errors.New("database unavailable"), http.StatusServiceUnavailable, "unavailable"},
	}
	for _, tc := range cases {
		t.Run(tc.path+"/"+tc.wantStatus, func(t *testing.T) {
			checker := NewMockReadinessChecker(gomock.NewController(t))
			if tc.wantPing {
				checker.EXPECT().Ping(gomock.Any()).DoAndReturn(func(ctx context.Context) error {
					if _, ok := ctx.Deadline(); !ok {
						t.Error("Ping context has no deadline")
					}
					return tc.pingErr
				})
			}
			handler := transport.NewHandler(checker, time.Second, nil, nil, nil)
			router := api.HandlerWithOptions(handler, api.ChiServerOptions{
				ErrorHandlerFunc: domainerror.HandleParameterError,
			})
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tc.path, nil))

			if rec.Code != tc.wantCode {
				t.Fatalf("status = %d, want %d; body = %s", rec.Code, tc.wantCode, rec.Body)
			}
			if got := rec.Header().Get("Content-Type"); got != "application/json" {
				t.Errorf("Content-Type = %q, want application/json", got)
			}
			var result api.HealthResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if string(result.Status) != tc.wantStatus {
				t.Errorf("response status = %q, want %q", result.Status, tc.wantStatus)
			}
		})
	}
}

func TestRouterRejectsInvalidTripID(t *testing.T) {
	// Invalid path parameters must be rejected before dispatch to the trip handlers.
	handler := transport.NewHandler(nil, time.Second, nil, nil, nil)
	router := api.HandlerWithOptions(handler, api.ChiServerOptions{
		ErrorHandlerFunc: domainerror.HandleParameterError,
	})
	cases := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/v1/trips/not-a-uuid"},
		{http.MethodPost, "/api/v1/trips/not-a-uuid/finish"},
	}
	for _, tc := range cases {
		t.Run(tc.method, func(t *testing.T) {
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, httptest.NewRequest(tc.method, tc.path, nil))

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400; body = %s", rec.Code, rec.Body)
			}
			if got := rec.Header().Get("Content-Type"); got != "application/problem+json" {
				t.Errorf("Content-Type = %q, want application/problem+json", got)
			}
			var problem api.Problem
			if err := json.Unmarshal(rec.Body.Bytes(), &problem); err != nil {
				t.Fatalf("decode problem: %v", err)
			}
			if problem.Code != "invalid_request" || problem.Status != http.StatusBadRequest {
				t.Errorf("unexpected problem: %+v", problem)
			}
			if problem.Instance == nil || *problem.Instance != tc.path {
				t.Errorf("problem instance = %v, want %q", problem.Instance, tc.path)
			}
		})
	}
}
