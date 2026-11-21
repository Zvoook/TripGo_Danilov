package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/trace"
)

type queryTracer struct {
	*otelpgx.Tracer
	duration metric.Float64Histogram
}

type queryMeasurementKey struct{}

type queryMeasurement struct {
	started time.Time
	labels  []attribute.KeyValue
}

var _ pgx.QueryTracer = (*queryTracer)(nil)

func NewQueryTracer(provider trace.TracerProvider, meter metric.Meter) (pgx.QueryTracer, error) {
	duration, err := meter.Float64Histogram("db.client.operation.duration", metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 5, 10))
	if err != nil {
		return nil, fmt.Errorf("create SQL duration histogram: %w", err)
	}
	return &queryTracer{
		duration: duration,
		Tracer: otelpgx.NewTracer(
			otelpgx.WithTracerProvider(provider),
			otelpgx.WithMeterProvider(noop.NewMeterProvider()),
			otelpgx.WithDisableSQLStatementInAttributes(),
		),
	}, nil
}

func (t *queryTracer) TraceQueryStart(
	ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData,
) context.Context {
	measurement := queryMeasurement{started: time.Now()}
	operation := otelpgx.SQLOperationName(ctx, data.SQL)
	measurement.labels = append(measurement.labels, attribute.String("db.operation.name", operation))
	if table := tripTable(operation, data.SQL); table != "" {
		measurement.labels = append(measurement.labels, attribute.String("db.collection.name", table))
	}

	parentID := trace.SpanContextFromContext(ctx).SpanID()
	queryCtx := t.Tracer.TraceQueryStart(ctx, conn, data)
	span := trace.SpanFromContext(queryCtx)
	if span.IsRecording() && span.SpanContext().SpanID() != parentID {
		span.SetAttributes(measurement.labels...)
	}
	return context.WithValue(queryCtx, queryMeasurementKey{}, measurement)
}

func (t *queryTracer) TraceQueryEnd(
	ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData,
) {
	measurement, ok := ctx.Value(queryMeasurementKey{}).(queryMeasurement)
	if ok {
		if kind := queryErrorType(data.Err); kind != "" {
			attr := attribute.String("error.type", kind)
			measurement.labels = append(measurement.labels, attr)
			trace.SpanFromContext(ctx).SetAttributes(attr)
		}
		t.duration.Record(ctx, time.Since(measurement.started).Seconds(),
			metric.WithAttributes(measurement.labels...))
	}
	t.Tracer.TraceQueryEnd(ctx, conn, data)
}

func tripTable(operation, sql string) string {
	const tripsTable = "trips"
	// This matches the SQL shapes built by our trip repository.
	normalized := strings.ToLower(strings.Join(strings.Fields(sql), " "))

	switch {
	case operation == "SELECT" && strings.Contains(normalized, " from trips "):
		return tripsTable
	case operation == "UPDATE" && strings.HasPrefix(normalized, "update trips "):
		return tripsTable
	case operation == "INSERT" && strings.HasPrefix(normalized, "insert into trips "):
		return tripsTable
	case operation == "INSERT" && strings.HasPrefix(normalized, "insert into trip_status_history "):
		return "trip_status_history"
	default:
		return ""
	}
}

func queryErrorType(err error) string {
	switch {
	case err == nil, errors.Is(err, pgx.ErrNoRows):
		return ""
	case errors.Is(err, context.DeadlineExceeded):
		return "deadline_exceeded"
	case errors.Is(err, context.Canceled):
		return "canceled"
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}

	return "_OTHER"
}
