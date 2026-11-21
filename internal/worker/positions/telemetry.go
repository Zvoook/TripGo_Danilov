package positions

import (
	"context"
	"sync/atomic"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

func (m *instruments) observe(meter metric.Meter, q *queue) error {
	_, err := meter.Int64ObservableGauge("tripgo.position_requests.queue.size",
		metric.WithInt64Callback(func(_ context.Context, observer metric.Int64Observer) error {
			observer.Observe(int64(len(q.jobs)))
			return nil
		}))
	if err != nil {
		return err
	}
	_, err = meter.Int64ObservableGauge("tripgo.position_requests.workers.active",
		metric.WithInt64Callback(func(_ context.Context, observer metric.Int64Observer) error {
			observer.Observe(m.active.Load())
			return nil
		}))
	return err
}

type instruments struct {
	checks   metric.Int64Counter
	requests metric.Int64Counter
	duration metric.Float64Histogram
	active   atomic.Int64
	tracer   trace.Tracer
}

func newInstruments(meter metric.Meter, tracer trace.Tracer) (*instruments, error) {
	m := &instruments{tracer: tracer}
	var err error
	m.checks, err = meter.Int64Counter("tripgo.stale_checks")
	if err != nil {
		return nil, err
	}
	m.requests, err = meter.Int64Counter("tripgo.position_requests")
	if err != nil {
		return nil, err
	}
	m.duration, err = meter.Float64Histogram("tripgo.position_requests.duration", metric.WithUnit("s"))
	if err != nil {
		return nil, err
	}
	return m, nil
}
