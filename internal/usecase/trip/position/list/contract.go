//go:generate go tool mockgen -source ${GOFILE} -destination mocks_test.go -package ${GOPACKAGE}_test -mock_names tripRepository=MockTripRepository,txManager=MockTxManager
package list

import (
	"context"

	"github.com/google/uuid"

	tripdomain "github.com/Zvoook/TripGo_Danilov/internal/domain/trip"
)

type tripRepository interface {
	GetByID(ctx context.Context, tripID uuid.UUID) (tripdomain.Trip, error)
	ListPositions(ctx context.Context, tripID uuid.UUID) ([]tripdomain.Position, error)
}
