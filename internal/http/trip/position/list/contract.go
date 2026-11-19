package list

import (
	"context"

	"github.com/google/uuid"

	tripdomain "github.com/Zvoook/TripGo_Danilov/internal/domain/trip"
)

type lister interface {
	List(ctx context.Context, tripID uuid.UUID) ([]tripdomain.Position, error)
}
