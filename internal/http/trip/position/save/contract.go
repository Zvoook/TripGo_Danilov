package save

import (
	"context"

	tripdomain "github.com/Zvoook/TripGo_Danilov/internal/domain/trip"
)

type saver interface {
	Save(ctx context.Context, position tripdomain.Position) (tripdomain.Position, error)
}
