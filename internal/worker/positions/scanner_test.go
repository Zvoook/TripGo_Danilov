package positions

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/trace"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/goleak"
	"go.uber.org/mock/gomock"

	tripdomain "github.com/Zvoook/TripGo_Danilov/internal/domain/trip"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func testInstruments(t *testing.T) *instruments {
	t.Helper()
	m, err := newInstruments(noop.NewMeterProvider().Meter("test"), tracenoop.NewTracerProvider().Tracer("test"))
	if err != nil {
		t.Fatal(err)
	}
	return m
}

type manualTicker struct {
	ticks   chan time.Time
	stopped chan struct{}
}

func (t manualTicker) Chan() <-chan time.Time { return t.ticks }
func (t manualTicker) Stop()                  { close(t.stopped) }

func TestScannerEnqueuesOnTickAndStops(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 3*time.Second)
	defer cancel()
	q := newQueue(1)
	ticks := manualTicker{ticks: make(chan time.Time, 1), stopped: make(chan struct{})}
	trip := tripdomain.StaleTrip{ID: uuid.New()}
	repo := NewMockstaleRepository(gomock.NewController(t))
	repo.EXPECT().ListStale(gomock.Any(), time.Minute).Return([]tripdomain.StaleTrip{trip}, nil)
	s := scanner{repository: repo, queue: q, staleAfter: time.Minute, logger: slog.New(slog.DiscardHandler), telemetry: testInstruments(t)}
	go s.run(ctx, ticks)
	t.Cleanup(func() {
		cancel()
		<-ticks.stopped
	})
	ticks.ticks <- time.Now()
	select {
	case task := <-q.jobs:
		if task.trip != trip {
			t.Fatalf("queued trip = %v, want %v", task.trip, trip)
		}
	case <-ctx.Done():
		t.Fatal("scanner did not enqueue the trip")
	}
}

func TestShutdownDrainsQueueWithTwoWorkers(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 4*time.Second)
	defer cancel()
	q := newQueue(3)
	ticks := manualTicker{ticks: make(chan time.Time), stopped: make(chan struct{})}
	push := NewMockpushGateway(gomock.NewController(t))
	started := make(chan struct{}, 3)
	for range 3 {
		trip := tripdomain.StaleTrip{ID: uuid.New(), DriverID: uuid.New()}
		if q.enqueue(trip, trace.SpanContext{}) != queued {
			t.Fatal("could not seed queue")
		}
		push.EXPECT().RequestPosition(gomock.Any(), trip.DriverID, trip.ID).
			DoAndReturn(func(ctx context.Context, _, _ uuid.UUID) error {
				started <- struct{}{}
				select {
				case <-ticks.stopped:
					return nil
				case <-ctx.Done():
					return ctx.Err()
				}
			})
	}
	s := start(ctx, Settings{Workers: 2, RequestTimeout: 5 * time.Second}, nil, push, slog.New(slog.DiscardHandler), q, testInstruments(t), ticks)
	t.Cleanup(func() {
		cancel()
		<-s.scannerDone
		<-s.workersDone
	})
	for range 2 {
		select {
		case <-started:
		case <-ctx.Done():
			t.Fatal("two workers did not start while the first push was blocked")
		}
	}
	if err := s.Shutdown(ctx); err != nil {
		t.Fatal(err)
	}
	if len(started) != 1 || len(q.jobs) != 0 {
		t.Fatal("shutdown did not drain the third task")
	}
}
