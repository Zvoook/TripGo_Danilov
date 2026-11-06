package finish

import (
	"context"
	"github.com/google/uuid"

	tripdomain "github.com/Zvoook/TripGo_Danilov/internal/domain/trip"
)

type finisher interface {
	Finish(context.Context, uuid.UUID) (tripdomain.Trip, error)
}
