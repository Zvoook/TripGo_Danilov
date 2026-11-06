package finish

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

func (u *Usecase) Finish(ctx context.Context, id uuid.UUID) (tripdomain.Trip, error) {
	operationCtx, cancel := context.WithTimeout(ctx, u.timeout)
	defer cancel()

	var result tripdomain.Trip
	err := u.transactions.Do(operationCtx, func(txCtx context.Context) error {
		value, err := u.repository.GetForUpdate(txCtx, id)
		if err != nil {
			return err
		}

		previous := value.Status
		now := time.Now().UTC().Truncate(time.Microsecond)

		if err := value.Finish(now); err != nil {
			return err
		}

		err = u.repository.Complete(txCtx, id, now)
		if err != nil {
			return err
		}

		err = u.repository.AddHistory(txCtx, id, &previous, tripdomain.StatusCompleted, now)
		if err != nil {
			return err
		}
		result = value
		return nil
	})
	if err != nil {
		return tripdomain.Trip{}, err
	}
	return result, nil
}
