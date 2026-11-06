package get

import (
	"context"
	"time"

	tripdomain "github.com/Zvoook/TripGo_Danilov/internal/domain/trip"
	"github.com/google/uuid"
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
