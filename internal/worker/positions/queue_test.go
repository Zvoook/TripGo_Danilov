package positions

import (
	"testing"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"

	tripdomain "github.com/Zvoook/TripGo_Danilov/internal/domain/trip"
)

func TestTryEnqueueApprovesNewTrip(t *testing.T) {
	q := newQueue(1)
	trip := tripdomain.StaleTrip{ID: uuid.New()}
	result := q.enqueue(trip, trace.SpanContext{})
	if result != queued {
		t.Fatalf("enqueue = %q, want %q", result, queued)
	}
}

func TestTryEnqueueDropsWhenFull(t *testing.T) {
	q := newQueue(1)
	trip1 := tripdomain.StaleTrip{ID: uuid.New()}
	trip2 := tripdomain.StaleTrip{ID: uuid.New()}

	firstResult := q.enqueue(trip1, trace.SpanContext{})
	if firstResult != queued {
		t.Fatalf("first enqueue = %q, want %q", firstResult, queued)
	}

	secondResult := q.enqueue(trip2, trace.SpanContext{})
	if secondResult != dropped {
		t.Fatalf("second enqueue = %q, want %q", secondResult, dropped)
	}
}

func TestTryEnqueueRejectsDuplicate(t *testing.T) {
	q := newQueue(2)
	trip := tripdomain.StaleTrip{ID: uuid.New()}

	firstResult := q.enqueue(trip, trace.SpanContext{})
	if firstResult != queued {
		t.Fatalf("first enqueue = %q, want %q", firstResult, queued)
	}

	secondResult := q.enqueue(trip, trace.SpanContext{})
	if secondResult != duplicate {
		t.Fatalf("second enqueue = %q, want %q", secondResult, duplicate)
	}
}
