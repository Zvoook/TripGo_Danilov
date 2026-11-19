package save

import (
	"context"
	"time"

	tripdomain "github.com/Zvoook/TripGo_Danilov/internal/domain/trip"
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

func (u *Usecase) Save(ctx context.Context, position tripdomain.Position) (tripdomain.Position, error) {
	if err := position.Validate(time.Now()); err != nil {
		return tripdomain.Position{}, err
	}
	operationCtx, cancel := context.WithTimeout(ctx, u.timeout)
	defer cancel()

	var saved tripdomain.Position

	err := u.transactions.Do(operationCtx, func(txCtx context.Context) error {
		value, err := u.repository.GetForUpdate(txCtx, position.TripID)
		if err != nil {
			return err
		}

		if value.Status == tripdomain.StatusCompleted {
			return tripdomain.ErrTripCompleted
		}

		saved, err = u.repository.InsertPosition(txCtx, position)
		if err != nil {
			return err
		}
		return u.repository.UpdateLastPositionAt(txCtx, saved.TripID, saved.RecordedAt)
	})
	if err != nil {
		return tripdomain.Position{}, err
	}
	return saved, nil
}
