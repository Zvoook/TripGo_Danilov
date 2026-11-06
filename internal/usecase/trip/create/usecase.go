package create

import (
	"context"
	"time"

	tripdomain "github.com/Zvoook/TripGo_Danilov/internal/domain/trip"

	"github.com/google/uuid"
)

type Usecase struct {
	repository   tripRepository
	transactions txManager
	timeout      time.Duration
}

func New(
	repository tripRepository,
	transactions txManager,
	timeout time.Duration,
) *Usecase {
	return &Usecase{
		repository:   repository,
		transactions: transactions,
		timeout:      timeout,
	}
}

func (u *Usecase) Create(ctx context.Context, input tripdomain.CreateInput) (tripdomain.Trip, error) {
	err := input.Validate()
	if err != nil {
		return tripdomain.Trip{}, err
	}

	operationCtx, cancel := context.WithTimeout(ctx, u.timeout)
	defer cancel()

	now := time.Now().UTC().Truncate(time.Microsecond)
	value := tripdomain.Trip{
		ID:             uuid.New(),
		UserID:         input.UserID,
		DriverID:       input.DriverID,
		StartLongitude: input.StartLongitude,
		StartLatitude:  input.StartLatitude,
		EndLongitude:   input.EndLongitude,
		EndLatitude:    input.EndLatitude,
		Price:          input.Price,
		Status:         tripdomain.StatusActive,
		StartedAt:      now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	err = u.transactions.Do(operationCtx, func(txCtx context.Context) error {
		err := u.repository.Insert(txCtx, value)
		if err != nil {
			return err
		}

		err = u.repository.AddHistory(txCtx, value.ID, nil, tripdomain.StatusActive, now)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return tripdomain.Trip{}, err
	}
	return value, nil
}
