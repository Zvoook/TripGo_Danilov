package positions

import (
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/mock/gomock"

	tripdomain "github.com/Zvoook/TripGo_Danilov/internal/domain/trip"
)

func TestWorkerSkipsCancelledTrip(t *testing.T) {
	q := newQueue(2)

	cancelledTrip := tripdomain.StaleTrip{
		ID:       uuid.New(),
		DriverID: uuid.New(),
	}
	waitingTrip := tripdomain.StaleTrip{
		ID:       uuid.New(),
		DriverID: uuid.New(),
	}

	if got := q.enqueue(cancelledTrip, trace.SpanContext{}); got != queued {
		t.Fatalf("enqueue cancelled trip = %q, want %q", got, queued)
	}
	if got := q.enqueue(waitingTrip, trace.SpanContext{}); got != queued {
		t.Fatalf("enqueue waiting trip = %q, want %q", got, queued)
	}

	q.Cancel(cancelledTrip.ID)

	push := NewMockpushGateway(gomock.NewController(t))
	push.EXPECT().
		RequestPosition(gomock.Any(), waitingTrip.DriverID, waitingTrip.ID).
		Return(nil).
		Times(1)

	pool := workerPool{
		queue:          q,
		telemetry:      testInstruments(t),
		push:           push,
		logger:         slog.New(slog.DiscardHandler),
		requestTimeout: time.Second,
	}

	close(q.jobs)
	pool.run(t.Context(), 1)
}

func TestWorkerContinuesAfterPushError(t *testing.T) {
	q := newQueue(2)

	firstTrip := tripdomain.StaleTrip{
		ID:       uuid.New(),
		DriverID: uuid.New(),
	}
	secondTrip := tripdomain.StaleTrip{
		ID:       uuid.New(),
		DriverID: uuid.New(),
	}

	if got := q.enqueue(firstTrip, trace.SpanContext{}); got != queued {
		t.Fatalf("enqueue cancelled trip = %q, want %q", got, queued)
	}
	if got := q.enqueue(secondTrip, trace.SpanContext{}); got != queued {
		t.Fatalf("enqueue waiting trip = %q, want %q", got, queued)
	}

	push := NewMockpushGateway(gomock.NewController(t))
	firstCall := push.EXPECT().
		RequestPosition(gomock.Any(), firstTrip.DriverID, firstTrip.ID).
		Return(errors.New("push unavailable")).
		Times(1)

	push.EXPECT().
		RequestPosition(gomock.Any(), secondTrip.DriverID, secondTrip.ID).
		Return(nil).
		Times(1).
		After(firstCall)

	pool := workerPool{
		queue:          q,
		telemetry:      testInstruments(t),
		push:           push,
		logger:         slog.New(slog.DiscardHandler),
		requestTimeout: time.Second,
	}

	close(q.jobs)
	pool.run(t.Context(), 1)
}
