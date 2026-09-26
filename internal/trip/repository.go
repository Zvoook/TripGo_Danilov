package trip

import (
	"context"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("trip not found")

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

	query, args, err := sq.Select("id", "user_id", "driver_id", "start_longitude",
		"end_longitude", "start_latitude", "end_latitude", "price",
		"status", "started_at", "created_at", "updated_at", "finished_at").
		From("trips").
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar).ToSql()

	if err != nil {
		return Trip{}, fmt.Errorf("select transaction error: %w", err)
	}
	var result Trip

	err = r.pool.QueryRow(queryCtx, query, args...).Scan(
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
		return Trip{}, fmt.Errorf("select transaction error: %w", err)
	}

	return result, nil
}
