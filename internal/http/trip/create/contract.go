//go:generate go tool mockgen -source ${GOFILE} -destination mocks_test.go -package ${GOPACKAGE}_test -mock_names creator=MockCreator
package create

import (
	"context"

	tripdomain "github.com/Zvoook/TripGo_Danilov/internal/domain/trip"
)

type creator interface {
	Create(context.Context, tripdomain.CreateInput) (tripdomain.Trip, error)
}
