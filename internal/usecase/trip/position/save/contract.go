//go:generate go tool mockgen -source ${GOFILE} -destination mocks_test.go -package ${GOPACKAGE}_test -mock_names tripRepository=MockTripRepository,txManager=MockTxManager
package save

import (
	"context"
	"time"

	"github.com/google/uuid"

	tripdomain "github.com/Zvoook/TripGo_Danilov/internal/domain/trip"
)

type (
	tripRepository interface {
		GetForUpdate(ctx context.Context, tripID uuid.UUID) (tripdomain.Trip, error)

		InsertPosition(ctx context.Context, position tripdomain.Position) (tripdomain.Position, error)

		UpdateLastPositionAt(
			ctx context.Context,
			tripID uuid.UUID,
			recordedAt time.Time,
		) error
	}

	txManager interface {
		Do(ctx context.Context, fn func(ctx context.Context) error) error
	}
)
