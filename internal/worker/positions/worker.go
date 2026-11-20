package positions

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type workerPool struct {
	queue          *queue
	push           pushGateway
	logger         *slog.Logger
	requestTimeout time.Duration
	telemetry      *instruments
}

func (p *workerPool) process(ctx context.Context, task *job) {
	if !p.queue.start(task) {
		return
	}
	defer p.queue.finish(task)

	p.telemetry.active.Add(1)
	defer p.telemetry.active.Add(-1)
	ctx, span := p.telemetry.tracer.Start(ctx, "positions.process", trace.WithLinks(trace.Link{SpanContext: task.scan}))
	defer span.End()

	requestCtx, cancel := context.WithTimeout(ctx, p.requestTimeout)
	defer cancel()

	requestCtx, pushSpan := p.telemetry.tracer.Start(requestCtx, "Push Service POST /api/v1/push", trace.WithSpanKind(trace.SpanKindClient))
	defer pushSpan.End()
	start := time.Now()
	err := p.push.RequestPosition(requestCtx, task.trip.DriverID, task.trip.ID)
	result := "success"
	if err != nil {
		result = "error"
		pushSpan.RecordError(err)
		pushSpan.SetStatus(codes.Error, "push request failed")
		span.SetStatus(codes.Error, "push request failed")
		p.logger.WarnContext(requestCtx, "request trip position failed", "error", err)
	}
	labels := metric.WithAttributes(attribute.String("result", result))
	p.telemetry.requests.Add(requestCtx, 1, labels)
	p.telemetry.duration.Record(requestCtx, time.Since(start).Seconds(), labels)
}

func (p *workerPool) work(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case task, ok := <-p.queue.jobs:
			if !ok || ctx.Err() != nil {
				return
			}
			p.process(ctx, task)
		}
	}
}

func (p *workerPool) run(ctx context.Context, workers int) {
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			p.work(ctx)
		}()
	}
	wg.Wait()
}
