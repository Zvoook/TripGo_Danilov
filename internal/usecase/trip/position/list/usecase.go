package list

import (
	"context"
	"time"

	"github.com/google/uuid"

	tripdomain "github.com/Zvoook/TripGo_Danilov/internal/domain/trip"
)

type Usecase struct {
	repository tripRepository
	timeout    time.Duration
}

func New(
	repository tripRepository,
	timeout time.Duration,
) *Usecase {
	return &Usecase{
		repository: repository,
		timeout:    timeout,
	}
}

func (u *Usecase) List(
	ctx context.Context,
	tripID uuid.UUID,
) ([]tripdomain.Position, error) {
	operationCtx, cancel := context.WithTimeout(ctx, u.timeout)
	defer cancel()

	if _, err := u.repository.GetByID(operationCtx, tripID); err != nil {
		return nil, err
	}
	return u.repository.ListPositions(operationCtx, tripID)
}
