package finish_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	tripdomain "github.com/Zvoook/TripGo_Danilov/internal/domain/trip"
	api "github.com/Zvoook/TripGo_Danilov/internal/generated"
	"github.com/Zvoook/TripGo_Danilov/internal/http/trip/finish"
)

func TestFinishTrip_OK(t *testing.T) {
	id := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	finishedAt := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	usecase := NewMockFinisher(gomock.NewController(t))
	usecase.EXPECT().Finish(gomock.Any(), id).Return(tripdomain.Trip{
		ID:         id,
		Status:     tripdomain.StatusCompleted,
		FinishedAt: &finishedAt,
	}, nil).Times(1)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/trips/"+id.String()+"/finish", nil)
	rec := httptest.NewRecorder()
	finish.NewHandler(usecase).FinishTrip(rec, req, id)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rec.Code, rec.Body.String())
	}
	var got api.Trip
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.Id != id || string(got.Status) != "completed" {
		t.Errorf("trip id=%v status=%q, want id=%v status=completed", got.Id, got.Status, id)
	}
	if got.FinishedAt == nil || !got.FinishedAt.Equal(finishedAt) {
		t.Errorf("finished_at = %v, want %v", got.FinishedAt, finishedAt)
	}
}

func TestFinishTrip_Errors(t *testing.T) {
	cases := []struct {
		name   string
		err    error
		status int
		code   string
	}{
		{"not found", tripdomain.ErrTripNotFound, http.StatusNotFound, "trip_not_found"},
		{"already completed", tripdomain.ErrTripCompleted, http.StatusConflict, "trip_completed"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			id := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
			usecase := NewMockFinisher(gomock.NewController(t))
			usecase.EXPECT().Finish(gomock.Any(), id).
				Return(tripdomain.Trip{}, tc.err).Times(1)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/trips/"+id.String()+"/finish", nil)
			rec := httptest.NewRecorder()
			finish.NewHandler(usecase).FinishTrip(rec, req, id)

			if rec.Code != tc.status {
				t.Fatalf("status = %d, want %d; body = %s", rec.Code, tc.status, rec.Body.String())
			}
			var problem api.Problem
			if err := json.Unmarshal(rec.Body.Bytes(), &problem); err != nil {
				t.Fatalf("decode problem: %v", err)
			}
			if problem.Code != tc.code {
				t.Errorf("problem code = %q, want %q", problem.Code, tc.code)
			}
		})
	}
}
