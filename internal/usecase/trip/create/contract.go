//go:generate go tool mockgen -source ${GOFILE} -destination mocks_test.go -package ${GOPACKAGE}_test -mock_names tripRepository=MockTripRepository,txManager=MockTxManager

package create

import (
	"context"
	"time"

	"github.com/google/uuid"

	tripdomain "github.com/Zvoook/TripGo_Danilov/internal/domain/trip"
)

type (
	tripRepository interface {
		Insert(ctx context.Context, trip tripdomain.Trip) error

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
