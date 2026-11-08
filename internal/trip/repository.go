package trip

import (
	"context"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/Zvoook/TripGo_Danilov/internal/postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("trip not found")
var ErrDriverBusy = errors.New("driver is busy")
var ErrTripCompleted = errors.New("trip already completed")
var ErrInvalidInput = errors.New("invalid trip input")

type Repository struct {
	pool         *pgxpool.Pool
	queryTimeout time.Duration
}

func NewRepository(pool *pgxpool.Pool, timeout time.Duration) *Repository {
	return &Repository{
		pool:         pool,
		queryTimeout: timeout,
	}
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (Trip, error) {
	queryCtx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	query, args, err := sq.Select(
		"id", "user_id", "driver_id", "start_longitude",
		"end_longitude", "start_latitude", "end_latitude", "price",
		"status", "started_at", "created_at", "updated_at", "finished_at",
	).From("trips").
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return Trip{}, fmt.Errorf("build select: %w", err)
	}

	executor := postgres.ExecutorFromContext(queryCtx, r.pool)
	row := executor.QueryRow(queryCtx, query, args...)
	return scanTrip(row)
}

// Called inside Do: the lock must remain held until the update is committed.
func (r *Repository) GetForUpdate(ctx context.Context, id uuid.UUID) (Trip, error) {
	queryCtx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	query, args, err := sq.Select(
		"id", "user_id", "driver_id", "start_longitude",
		"end_longitude", "start_latitude", "end_latitude", "price",
		"status", "started_at", "created_at", "updated_at", "finished_at",
	).From("trips").
		Where(sq.Eq{"id": id}).
		Suffix("FOR UPDATE").
		PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return Trip{}, fmt.Errorf("build locked select: %w", err)
	}

	executor := postgres.ExecutorFromContext(queryCtx, r.pool)
	row := executor.QueryRow(queryCtx, query, args...)
	return scanTrip(row)
}

// Both SELECT queries return the columns in this order.
func scanTrip(row pgx.Row) (Trip, error) {
	var result Trip
	err := row.Scan(
		&result.ID,
		&result.UserID,
		&result.DriverID,
		&result.StartLongitude,
		&result.EndLongitude,
		&result.StartLatitude,
		&result.EndLatitude,
		&result.Price,
		&result.Status,
		&result.StartedAt,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.FinishedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Trip{}, ErrNotFound
	} else if err != nil {
		return Trip{}, fmt.Errorf("read trip: %w", err)
	}
	return result, nil
}

func (r *Repository) Insert(ctx context.Context, value Trip) error {
	insertCtx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	query, args, err := sq.Insert("trips").Columns(
		"id", "user_id", "driver_id", "start_longitude",
		"end_longitude", "start_latitude", "end_latitude", "price",
		"status", "started_at", "created_at", "updated_at", "finished_at",
	).Values(
		value.ID,
		value.UserID,
		value.DriverID,
		value.StartLongitude,
		value.EndLongitude,
		value.StartLatitude,
		value.EndLatitude,
		value.Price,
		string(value.Status),
		value.StartedAt,
		value.CreatedAt,
		value.UpdatedAt,
		value.FinishedAt,
	).PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return fmt.Errorf("build insert: %w", err)
	}

	executor := postgres.ExecutorFromContext(insertCtx, r.pool)
	_, err = executor.Exec(insertCtx, query, args...)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == "23505" && pgErr.ConstraintName == "one_active_trip_per_driver_idx" {
			return ErrDriverBusy
		}
	}
	if err != nil {
		return fmt.Errorf("insert trip: %w", err)
	}
	return nil
}

func (r *Repository) AddHistory(
	ctx context.Context,
	tripID uuid.UUID,
	from *Status,
	to Status,
	changedAt time.Time,
) error {
	execCtx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	// nil becomes SQL NULL for the first entry in the trip history.
	var previous any
	if from != nil {
		previous = string(*from)
	}

	query, args, err := sq.Insert("trip_status_history").Columns(
		"trip_id",
		"from_status",
		"to_status",
		"changed_at",
	).Values(
		tripID,
		previous,
		string(to),
		changedAt,
	).PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return fmt.Errorf("build history insert: %w", err)
	}

	executor := postgres.ExecutorFromContext(execCtx, r.pool)
	_, err = executor.Exec(execCtx, query, args...)
	if err != nil {
		return fmt.Errorf("insert trip history: %w", err)
	}
	return nil
}

func (r *Repository) Complete(ctx context.Context, id uuid.UUID, at time.Time) error {
	queryCtx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	query, args, err := sq.Update("trips").
		Set("status", string(StatusCompleted)).
		Set("finished_at", at).
		Set("updated_at", at).
		Where(sq.Eq{"id": id, "status": string(StatusActive)}).
		PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return fmt.Errorf("build completion: %w", err)
	}

	executor := postgres.ExecutorFromContext(queryCtx, r.pool)
	result, err := executor.Exec(queryCtx, query, args...)
	if err != nil {
		return fmt.Errorf("complete trip: %w", err)
	}
	if result.RowsAffected() != 1 {
		return fmt.Errorf("complete trip: expected one active trip")
	}
	return nil
}
