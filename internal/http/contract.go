package http

import (
	"context"
)

type readinessChecker interface {
	Ping(ctx context.Context) error
}
