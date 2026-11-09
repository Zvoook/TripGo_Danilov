//go:generate go tool mockgen -source ${GOFILE} -destination mocks_test.go -package ${GOPACKAGE}_test -mock_names getter=MockGetter
package get

import (
	"context"

	"github.com/google/uuid"

	tripdomain "github.com/Zvoook/TripGo_Danilov/internal/domain/trip"
)

type getter interface {
	Get(context.Context, uuid.UUID) (tripdomain.Trip, error)
}
