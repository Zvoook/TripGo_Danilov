package trip

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	tripdomain "github.com/Zvoook/TripGo_Danilov/internal/domain/trip"
)

func (r *Repository) ListStale(ctx context.Context, staleAfter time.Duration) ([]tripdomain.StaleTrip, error) {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	const query = `
		SELECT id, driver_id
		FROM trips
		WHERE status = 'active'
		  AND COALESCE(last_position_at, started_at) < now() - $1::interval
		ORDER BY COALESCE(last_position_at, started_at), id`

	interval := pgtype.Interval{Microseconds: staleAfter.Microseconds(), Valid: true}
	rows, err := r.pool.Query(ctx, query, interval)
	if err != nil {
		return nil, fmt.Errorf("query stale trips: %w", err)
	}
	defer rows.Close()

	var trips []tripdomain.StaleTrip
	for rows.Next() {
		var trip tripdomain.StaleTrip
		if err := rows.Scan(&trip.ID, &trip.DriverID); err != nil {
			return nil, fmt.Errorf("scan stale trip: %w", err)
		}
		trips = append(trips, trip)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read stale trips: %w", err)
	}
	return trips, nil
}
