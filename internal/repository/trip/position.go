package trip

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	tripdomain "github.com/Zvoook/TripGo_Danilov/internal/domain/trip"
	"github.com/Zvoook/TripGo_Danilov/internal/infra/postgres"
)

func (r *Repository) InsertPosition(ctx context.Context, position tripdomain.Position) (tripdomain.Position, error) {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	const query = `
		INSERT INTO trip_positions (trip_id, latitude, longitude, recorded_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, recorded_at`

	executor := postgres.ExecutorFromContext(ctx, r.pool)
	err := executor.QueryRow(
		ctx,
		query,
		position.TripID,
		position.Latitude,
		position.Longitude,
		position.RecordedAt,
	).Scan(&position.ID, &position.RecordedAt)
	if err != nil {
		return tripdomain.Position{}, fmt.Errorf("insert position: %w", err)
	}

	position.RecordedAt = position.RecordedAt.UTC()
	return position, nil
}

func (r *Repository) UpdateLastPositionAt(
	ctx context.Context,
	tripID uuid.UUID,
	recordedAt time.Time,
) error {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	const query = `
	UPDATE trips
	SET last_position_at = GREATEST(last_position_at, $2)
	WHERE id = $1`

	executor := postgres.ExecutorFromContext(ctx, r.pool)
	if _, err := executor.Exec(ctx, query, tripID, recordedAt); err != nil {
		return fmt.Errorf("update last position at: %w", err)
	}
	return nil
}

func (r *Repository) ListPositions(
	ctx context.Context,
	tripID uuid.UUID,
) ([]tripdomain.Position, error) {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	const query = `
	SELECT id, trip_id, latitude, longitude, recorded_at
	FROM trip_positions
	WHERE trip_id = $1
	ORDER BY recorded_at, id`

	rows, err := r.pool.Query(ctx, query, tripID)
	if err != nil {
		return nil, fmt.Errorf("list positions: %w", err)
	}
	defer rows.Close()

	positions := make([]tripdomain.Position, 0)
	for rows.Next() {
		var position tripdomain.Position
		if err := rows.Scan(
			&position.ID,
			&position.TripID,
			&position.Latitude,
			&position.Longitude,
			&position.RecordedAt,
		); err != nil {
			return nil, fmt.Errorf("scan position: %w", err)
		}
		position.RecordedAt = position.RecordedAt.UTC()
		positions = append(positions, position)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate positions: %w", err)
	}

	return positions, nil
}
