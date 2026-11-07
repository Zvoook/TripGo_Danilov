package create_test

import (
	"context"
	"errors"
	"testing"
	"time"

	tripdomain "github.com/Zvoook/TripGo_Danilov/internal/domain/trip"
	"github.com/Zvoook/TripGo_Danilov/internal/usecase/trip/create"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

func TestCreateTrip(t *testing.T) {
	input := tripdomain.CreateInput{
		UserID:   uuid.MustParse("123e4567-e89b-12d3-a456-426614174001"),
		DriverID: uuid.MustParse("123e4567-e89b-12d3-a456-426614174002"),
		Price:    1000,
	}
	cases := []struct {
		name          string
		expectedError error
	}{
		{"created", nil},
		{"driver busy", tripdomain.ErrDriverBusy},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := NewMockTripRepository(ctrl)
			tx := NewMockTxManager(ctrl)
			tx.EXPECT().Do(gomock.Any(), gomock.Any()).
				DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
					return fn(ctx)
				})

			var inserted tripdomain.Trip
			insert := repo.EXPECT().Insert(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, value tripdomain.Trip) error {
					inserted = value
					return tc.expectedError
				})
			if tc.expectedError == nil {
				repo.EXPECT().AddHistory(gomock.Any(),
					gomock.Cond(func(id uuid.UUID) bool { return id == inserted.ID }),
					nil, tripdomain.StatusActive, gomock.Any()).
					After(insert).Return(nil)
			}

			got, err := create.New(repo, tx, time.Second).Create(t.Context(), input)
			if !errors.Is(err, tc.expectedError) {
				t.Fatalf("error = %v, want %v", err, tc.expectedError)
			}
			if tc.expectedError != nil {
				return
			}
			if got != inserted || got.ID == uuid.Nil || got.Status != tripdomain.StatusActive {
				t.Fatalf("unexpected created trip: %+v", got)
			}
			if got.UserID != input.UserID || got.DriverID != input.DriverID || got.Price != input.Price {
				t.Errorf("trip does not match input: %+v", got)
			}
		})
	}
}
