package get

import (
	"context"

	"github.com/google/uuid"

	tripdomain "github.com/Zvoook/TripGo_Danilov/internal/domain/trip"
)

type (
	tripRepository interface {
		GetByID(ctx context.Context, id uuid.UUID) (tripdomain.Trip, error)
	}
)
