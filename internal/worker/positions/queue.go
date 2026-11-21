package positions

import (
	"sync"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"

	tripdomain "github.com/Zvoook/TripGo_Danilov/internal/domain/trip"
)

type job struct {
	trip    tripdomain.StaleTrip
	started bool
	scan    trace.SpanContext
}

type queue struct {
	jobs    chan *job
	mu      sync.Mutex
	pending map[uuid.UUID]*job
}

func newQueue(size int) *queue {
	return &queue{
		jobs:    make(chan *job, size),
		pending: make(map[uuid.UUID]*job),
	}
}

func (q *queue) enqueue(trip tripdomain.StaleTrip, scan trace.SpanContext) enqueueResult {
	q.mu.Lock()
	defer q.mu.Unlock()

	if _, exists := q.pending[trip.ID]; exists {
		return duplicate
	}

	task := &job{trip: trip, scan: scan}
	select {
	case q.jobs <- task:
		q.pending[trip.ID] = task
		return queued
	default:
		return dropped
	}
}

func (q *queue) Cancel(tripID uuid.UUID) {
	q.mu.Lock()
	defer q.mu.Unlock()

	task, exists := q.pending[tripID]
	if exists && !task.started {
		delete(q.pending, tripID)
	}
}

func (q *queue) start(task *job) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.pending[task.trip.ID] != task {
		return false
	}
	task.started = true
	return true
}

func (q *queue) finish(task *job) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.pending[task.trip.ID] == task {
		delete(q.pending, task.trip.ID)
	}
}
