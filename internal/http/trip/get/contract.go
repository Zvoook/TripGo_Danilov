package get

import (
	"context"

	tripdomain "github.com/Zvoook/TripGo_Danilov/internal/domain/trip"
	"github.com/google/uuid"
)

type getter interface {
	Get(context.Context, uuid.UUID) (tripdomain.Trip, error)
}
