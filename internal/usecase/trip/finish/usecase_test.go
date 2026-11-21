package finish_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	tripdomain "github.com/Zvoook/TripGo_Danilov/internal/domain/trip"
	"github.com/Zvoook/TripGo_Danilov/internal/usecase/trip/finish"
)

func TestFinishTrip(t *testing.T) {
	id := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	cases := []struct {
		name           string
		expectedStatus tripdomain.Status
		expectedError  error
	}{
		{"completed", tripdomain.StatusActive, nil},
		{"already completed", tripdomain.StatusCompleted, tripdomain.ErrTripCompleted},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repository := NewMockTripRepository(ctrl)
			tx := NewMockTxManager(ctrl)
			tx.EXPECT().Do(gomock.Any(), gomock.Any()).
				DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
					return fn(ctx)
				})

			load := repository.EXPECT().GetForUpdate(gomock.Any(), id).
				Return(tripdomain.Trip{ID: id, Status: tc.expectedStatus}, nil)
			var completedAt time.Time
			if tc.expectedError == nil {
				complete := repository.EXPECT().Complete(gomock.Any(), id, gomock.Any()).
					After(load).
					DoAndReturn(func(_ context.Context, _ uuid.UUID, at time.Time) error {
						completedAt = at
						return nil
					})
				previous := tripdomain.StatusActive
				repository.EXPECT().AddHistory(gomock.Any(), id, &previous, tripdomain.StatusCompleted,
					gomock.Cond(func(at time.Time) bool { return at.Equal(completedAt) })).
					After(complete).Return(nil)
			}

			got, err := finish.New(repository, tx, time.Second).Finish(t.Context(), id)
			if !errors.Is(err, tc.expectedError) {
				t.Fatalf("error = %v, want %v", err, tc.expectedError)
			}
			if tc.expectedError != nil {
				return
			}
			if got.ID != id || got.Status != tripdomain.StatusCompleted {
				t.Errorf("unexpected completed trip: %+v", got)
			}
			if got.FinishedAt == nil || !got.FinishedAt.Equal(completedAt) {
				t.Errorf("finished_at = %v, want %v", got.FinishedAt, completedAt)
			}
		})
	}
}
