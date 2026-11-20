//go:build integration

package trip_test

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	tripdomain "github.com/Zvoook/TripGo_Danilov/internal/domain/trip"
	"github.com/Zvoook/TripGo_Danilov/internal/infra/postgres"
	triprepo "github.com/Zvoook/TripGo_Danilov/internal/repository/trip"
)

func TestStaleTripsAndPositionRollback(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Fatal("DATABASE_URL is required")
	}
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	repo := triprepo.NewRepository(pool, 3*time.Second)
	tx := postgres.NewTxManager(pool, 3*time.Second, slog.New(slog.DiscardHandler))
	now := time.Now().UTC().Truncate(time.Microsecond)
	insert := func(start time.Time, status tripdomain.Status) uuid.UUID {
		t.Helper()
		value := tripdomain.Trip{
			ID: uuid.New(), DriverID: uuid.New(), UserID: uuid.New(),
			Status: status, StartedAt: start, CreatedAt: start, UpdatedAt: now,
		}
		if status == tripdomain.StatusCompleted {
			value.FinishedAt = &now
		}
		if err := repo.Insert(ctx, value); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			cleanup, stop := context.WithTimeout(context.WithoutCancel(ctx), 3*time.Second)
			defer stop()
			if _, err := pool.Exec(cleanup, "DELETE FROM trips WHERE id = $1", value.ID); err != nil {
				t.Error(err)
			}
		})
		return value.ID
	}
	stale := insert(now.Add(-5*time.Minute), tripdomain.StatusActive)
	fresh := insert(now, tripdomain.StatusActive)
	completed := insert(now.Add(-5*time.Minute), tripdomain.StatusCompleted)
	list, err := repo.ListStale(ctx, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	contains := func(id uuid.UUID) bool {
		return slices.ContainsFunc(list, func(trip tripdomain.StaleTrip) bool { return trip.ID == id })
	}
	if !contains(stale) || contains(fresh) || contains(completed) {
		t.Fatal("stale query did not distinguish stale, fresh and completed trips")
	}

	rollback := errors.New("rollback position")
	write := func(at time.Time, result error) error {
		return tx.Do(ctx, func(txCtx context.Context) error {
			saved, err := repo.InsertPosition(txCtx, tripdomain.Position{TripID: stale, RecordedAt: at})
			if err != nil {
				return err
			}
			if err := repo.UpdateLastPositionAt(txCtx, stale, saved.RecordedAt); err != nil {
				return err
			}
			return result
		})
	}
	if err := write(now, nil); err != nil {
		t.Fatal(err)
	}
	older := now.Add(-time.Minute)
	if err := write(older, nil); err != nil {
		t.Fatal(err)
	}
	if err := write(now.Add(time.Second), rollback); !errors.Is(err, rollback) {
		t.Fatalf("rollback error = %v", err)
	}
	positions, err := repo.ListPositions(ctx, stale)
	if err != nil {
		t.Fatal(err)
	}
	if len(positions) != 2 || !positions[0].RecordedAt.Equal(older) || !positions[1].RecordedAt.Equal(now) {
		t.Fatalf("unexpected route after rollback: %+v", positions)
	}
	var last time.Time
	if err := pool.QueryRow(ctx, "SELECT last_position_at FROM trips WHERE id = $1", stale).Scan(&last); err != nil {
		t.Fatal(err)
	}
	if !last.Equal(now) {
		t.Fatalf("last position = %v, want %v", last, now)
	}
}

func TestTripRepositoryTransaction(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Fatal("DATABASE_URL is required; start PostgreSQL and apply the migrations first")
	}

	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal("invalid PostgreSQL connection configuration")
	}
	t.Cleanup(pool.Close)
	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("connect to PostgreSQL: %v", err)
	}

	repo := triprepo.NewRepository(pool, 3*time.Second)
	tx := postgres.NewTxManager(pool, 3*time.Second, slog.New(slog.DiscardHandler))
	rollback := errors.New("request rollback")
	cases := []struct {
		name      string
		callback  error
		wantSaved bool
	}{
		{"commit", nil, true},
		{"rollback", rollback, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			now := time.Now().UTC().Truncate(time.Microsecond)
			value := tripdomain.Trip{
				ID: uuid.New(), UserID: uuid.New(), DriverID: uuid.New(),
				Price: 1000, Status: tripdomain.StatusActive,
				StartedAt: now, CreatedAt: now, UpdatedAt: now,
			}
			t.Cleanup(func() {
				cleanupCtx, stop := context.WithTimeout(context.WithoutCancel(t.Context()), 3*time.Second)
				defer stop()
				if _, err := pool.Exec(cleanupCtx, "DELETE FROM trips WHERE id = $1", value.ID); err != nil {
					t.Errorf("clean up trip: %v", err)
				}
			})

			err := tx.Do(ctx, func(txCtx context.Context) error {
				if err := repo.Insert(txCtx, value); err != nil {
					return err
				}
				if err := repo.AddHistory(txCtx, value.ID, nil, tripdomain.StatusActive, now); err != nil {
					return err
				}
				checkTripRecords(t, txCtx, postgres.ExecutorFromContext(txCtx, pool), value.ID, true)
				return tc.callback
			})
			if !errors.Is(err, tc.callback) {
				t.Fatalf("transaction error = %v, want %v", err, tc.callback)
			}

			checkTripRecords(t, ctx, pool, value.ID, tc.wantSaved)
			got, err := repo.GetByID(ctx, value.ID)
			if !tc.wantSaved {
				if !errors.Is(err, tripdomain.ErrTripNotFound) {
					t.Fatalf("GetByID after rollback: error = %v, want ErrTripNotFound", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("GetByID after commit: %v", err)
			}
			if got.ID != value.ID || got.Status != value.Status || got.Price != value.Price {
				t.Errorf("unexpected persisted trip: %+v", got)
			}
		})
	}
}

func checkTripRecords(t *testing.T, ctx context.Context, exec postgres.Executor, id uuid.UUID, want bool) {
	t.Helper()
	const query = "SELECT EXISTS (SELECT 1 FROM trips WHERE id = $1), " +
		"EXISTS (SELECT 1 FROM trip_status_history WHERE trip_id = $1)"
	var tripExists, historyExists bool
	if err := exec.QueryRow(ctx, query, id).Scan(&tripExists, &historyExists); err != nil {
		t.Fatalf("check trip and history: %v", err)
	}
	if tripExists != want || historyExists != want {
		t.Fatalf("trip exists = %t, history exists = %t; want both %t", tripExists, historyExists, want)
	}
}
