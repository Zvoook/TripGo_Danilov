package container

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/metric"

	tripdomain "github.com/Zvoook/TripGo_Danilov/internal/domain/trip"
)

type tripEvent struct {
	logger  *slog.Logger
	counter metric.Int64Counter
	message string
}

func newTripEvent(meter metric.Meter, logger *slog.Logger, name, message string) (tripEvent, error) {
	counter, err := meter.Int64Counter(name, metric.WithUnit("{trip}"))
	if err != nil {
		return tripEvent{}, fmt.Errorf("create %s counter: %w", name, err)
	}
	return tripEvent{logger: logger, counter: counter, message: message}, nil
}

func (e tripEvent) record(ctx context.Context, trip tripdomain.Trip) {
	e.counter.Add(ctx, 1)
	e.logger.InfoContext(ctx, e.message, "trip_id", trip.ID.String())
}

type observedCreate struct {
	run   func(context.Context, tripdomain.CreateInput) (tripdomain.Trip, error)
	event tripEvent
}

func (u observedCreate) Create(ctx context.Context, input tripdomain.CreateInput) (tripdomain.Trip, error) {
	trip, err := u.run(ctx, input)
	if err == nil {
		u.event.record(ctx, trip)
	}
	return trip, err
}

type observedFinish struct {
	run   func(context.Context, uuid.UUID) (tripdomain.Trip, error)
	event tripEvent
}

func (u observedFinish) Finish(ctx context.Context, id uuid.UUID) (tripdomain.Trip, error) {
	trip, err := u.run(ctx, id)
	if err == nil {
		u.event.record(ctx, trip)
	}
	return trip, err
}

func observeActiveTrips(meter metric.Meter, countActive func(context.Context) (int64, error)) error {
	_, err := meter.Int64ObservableGauge("tripgo.trips.active", metric.WithUnit("{trip}"),
		metric.WithInt64Callback(func(ctx context.Context, observer metric.Int64Observer) error {
			count, err := countActive(ctx)
			if err != nil {
				return err
			}
			observer.Observe(count)
			return nil
		}))
	if err != nil {
		return fmt.Errorf("create active trips gauge: %w", err)
	}
	return nil
}
