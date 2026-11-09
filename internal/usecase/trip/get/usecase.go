package get

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

func (u *Usecase) Get(
	ctx context.Context,
	id uuid.UUID,
) (tripdomain.Trip, error) {
	operationCtx, cancel := context.WithTimeout(ctx, u.timeout)
	defer cancel()

	return u.repository.GetByID(operationCtx, id)
}
