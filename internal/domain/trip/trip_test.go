package trip

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestFinishTrip(t *testing.T) {
	tt := []struct {
		name           string
		trip           Trip
		expectedStatus Status
		expectedError  error
	}{
		{
			name: "Finish active trip",
			trip: Trip{
				ID:     uuid.New(),
				Status: StatusActive,
			},
			expectedStatus: StatusCompleted,
			expectedError:  nil,
		},
		{
			name: "Finish completed trip",
			trip: Trip{
				ID:     uuid.New(),
				Status: StatusCompleted,
			},
			expectedStatus: StatusCompleted,
			expectedError:  ErrTripCompleted,
		},
	}
	for _, test := range tt {
		t.Run(test.name, func(t *testing.T) {
			err := test.trip.Finish(time.Now())
			if err != test.expectedError {
				t.Errorf("Expected error %v, got %v", test.expectedError, err)
			}
			if test.trip.Status != test.expectedStatus {
				t.Errorf("Expected status %v, got %v", test.expectedStatus, test.trip.Status)
			}
		})
	}
}
