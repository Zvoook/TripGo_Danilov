package positions

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type scanner struct {
	repository staleRepository
	queue      positionQueue
	staleAfter time.Duration
	logger     *slog.Logger
	telemetry  *instruments
}

func (s *scanner) run(ctx context.Context, ticks ticker) {
	defer ticks.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticks.Chan():
			if err := s.scanOnce(ctx); err != nil && ctx.Err() == nil {
				s.logger.WarnContext(ctx, "scan stale trips failed", "error", err)
			}
		}
	}
}

func (s *scanner) scanOnce(ctx context.Context) error {
	ctx, span := s.telemetry.tracer.Start(ctx, "positions.scan")
	defer span.End()
	trips, err := s.repository.ListStale(ctx, s.staleAfter)
	if err != nil {
		return fmt.Errorf("scan stale trips: %w", err)
	}

	for _, trip := range trips {
		if err := ctx.Err(); err != nil {
			return err
		}
		result := s.queue.enqueue(trip, span.SpanContext())
		labels := metric.WithAttributes(attribute.String("result", string(result)))
		s.telemetry.checks.Add(ctx, 1, labels)
		if result == dropped {
			s.telemetry.requests.Add(ctx, 1, labels)
			s.logger.WarnContext(ctx, "position request dropped: queue full", "trip_id", trip.ID)
		}
	}

	return nil
}
