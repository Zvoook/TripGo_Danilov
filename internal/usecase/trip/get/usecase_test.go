package get_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	tripdomain "github.com/Zvoook/TripGo_Danilov/internal/domain/trip"
	"github.com/Zvoook/TripGo_Danilov/internal/usecase/trip/get"
)

func TestGetTrip(t *testing.T) {
	id := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	dbErr := errors.New("database unavailable")
	cases := []struct {
		name           string
		expectedResult tripdomain.Trip
		expectedError  error
	}{
		{"found", tripdomain.Trip{ID: id, Status: tripdomain.StatusActive}, nil},
		{"not found", tripdomain.Trip{}, tripdomain.ErrTripNotFound},
		{"repository error", tripdomain.Trip{}, dbErr},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repository := NewMockTripRepository(gomock.NewController(t))
			repository.EXPECT().GetByID(gomock.Any(), id).Return(tc.expectedResult, tc.expectedError)

			got, err := get.New(repository, time.Second).Get(t.Context(), id)
			if !errors.Is(err, tc.expectedError) {
				t.Fatalf("error = %v, want %v", err, tc.expectedError)
			}
			if got != tc.expectedResult {
				t.Errorf("trip = %+v, want %+v", got, tc.expectedResult)
			}
		})
	}
}
