package finish

import (
	"context"
	"time"

	"github.com/google/uuid"

	tripdomain "github.com/Zvoook/TripGo_Danilov/internal/domain/trip"
)

type (
	tripRepository interface {
		GetForUpdate(ctx context.Context, id uuid.UUID) (tripdomain.Trip, error)
		Complete(ctx context.Context, id uuid.UUID, at time.Time) error

		AddHistory(
			ctx context.Context,
			tripID uuid.UUID,
			from *tripdomain.Status,
			to tripdomain.Status,
			changedAt time.Time,
		) error
	}

	txManager interface {
		Do(ctx context.Context, fn func(ctx context.Context) error) error
	}
)
