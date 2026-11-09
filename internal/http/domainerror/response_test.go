package domainerror

import (
	"errors"
	"net/http"
	"testing"

	tripdomain "github.com/Zvoook/TripGo_Danilov/internal/domain/trip"
)

func TestProblemMapping(t *testing.T) {
	cases := []struct {
		name   string
		err    error
		status int32
		code   string
	}{
		{"invalid request", tripdomain.ErrInvalidRequest, http.StatusBadRequest, "invalid_request"},
		{"trip not found", tripdomain.ErrTripNotFound, http.StatusNotFound, "trip_not_found"},
		{"driver busy", tripdomain.ErrDriverBusy, http.StatusConflict, "driver_busy"},
		{"trip completed", tripdomain.ErrTripCompleted, http.StatusConflict, "trip_completed"},
		{"internal error", errors.New("database unavailable"), http.StatusInternalServerError, "internal_error"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Problem(tc.err)
			if got.Status != tc.status || got.Code != tc.code {
				t.Errorf("got status=%d code=%q, want status=%d code=%q",
					got.Status, got.Code, tc.status, tc.code)
			}
		})
	}
}
