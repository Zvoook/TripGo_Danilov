package positions

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"

	tripdomain "github.com/Zvoook/TripGo_Danilov/internal/domain/trip"
)

type staleRepository interface {
	ListStale(ctx context.Context, staleAfter time.Duration) ([]tripdomain.StaleTrip, error)
}

//go:generate go tool mockgen -destination scanner_mock_test.go -package positions . staleRepository

type enqueueResult string

const (
	queued    enqueueResult = "queued"
	duplicate enqueueResult = "duplicate"
	dropped   enqueueResult = "dropped"
)

type positionQueue interface {
	enqueue(trip tripdomain.StaleTrip, scan trace.SpanContext) enqueueResult
}

//go:generate go tool mockgen -destination push_mock_test.go -package positions . pushGateway

type pushGateway interface {
	RequestPosition(ctx context.Context, driverID, tripID uuid.UUID) error
}

type ticker interface {
	Chan() <-chan time.Time
	Stop()
}
