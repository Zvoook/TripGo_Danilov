//go:build integration

package trip_test

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	tripdomain "github.com/Zvoook/TripGo_Danilov/internal/domain/trip"
	"github.com/Zvoook/TripGo_Danilov/internal/infra/postgres"
	triprepo "github.com/Zvoook/TripGo_Danilov/internal/repository/trip"
)

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
