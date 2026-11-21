package positions

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type realTicker struct {
	*time.Ticker
}

func (t realTicker) Chan() <-chan time.Time {
	return t.C
}

type Settings struct {
	CheckInterval  time.Duration
	StaleAfter     time.Duration
	RequestTimeout time.Duration
	Workers        int
	QueueSize      int
}

type Service struct {
	queue       *queue
	stopScanner context.CancelFunc
	stopWorkers context.CancelFunc
	scannerDone chan struct{}
	workersDone chan struct{}
}

func Start(ctx context.Context, cfg Settings, repo staleRepository, push pushGateway, logger *slog.Logger, meter metric.Meter, tracer trace.Tracer) (*Service, error) {
	q := newQueue(cfg.QueueSize)
	m, err := newInstruments(meter, tracer)
	if err != nil {
		return nil, err
	}
	if err := m.observe(meter, q); err != nil {
		return nil, err
	}
	return start(ctx, cfg, repo, push, logger, q, m, realTicker{time.NewTicker(cfg.CheckInterval)}), nil
}

func start(ctx context.Context, cfg Settings, repo staleRepository, push pushGateway, logger *slog.Logger, q *queue, m *instruments, ticks ticker) *Service {
	scanCtx, stopScanner := context.WithCancel(ctx)
	workCtx, stopWorkers := context.WithCancel(ctx)
	s := &Service{
		queue: q, stopScanner: stopScanner, stopWorkers: stopWorkers,
		scannerDone: make(chan struct{}), workersDone: make(chan struct{}),
	}
	scanner := scanner{repository: repo, queue: q, staleAfter: cfg.StaleAfter, logger: logger, telemetry: m}
	pool := workerPool{queue: q, push: push, logger: logger, requestTimeout: cfg.RequestTimeout, telemetry: m}
	go func() {
		defer close(s.scannerDone)
		defer close(q.jobs)
		scanner.run(scanCtx, ticks)
	}()
	go func() {
		defer close(s.workersDone)
		pool.run(workCtx, cfg.Workers)
	}()
	return s
}

func (s *Service) Cancel(tripID uuid.UUID) {
	s.queue.Cancel(tripID)
}

func (s *Service) Shutdown(ctx context.Context) error {
	s.stopScanner()
	defer s.stopWorkers()
	select {
	case <-s.workersDone:
		return nil
	case <-ctx.Done():
		s.stopWorkers()
		<-s.scannerDone
		<-s.workersDone
		return ctx.Err()
	}
}
