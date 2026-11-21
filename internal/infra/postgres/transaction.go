package postgres

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	pgxpool "github.com/jackc/pgx/v5/pgxpool"
)

type TxManager struct {
	logger          *slog.Logger
	pool            *pgxpool.Pool
	rollbackTimeout time.Duration
}

func NewTxManager(pool *pgxpool.Pool, timeout time.Duration, logger *slog.Logger) *TxManager {
	return &TxManager{
		pool:            pool,
		logger:          logger,
		rollbackTimeout: timeout,
	}
}

type txKey struct{}

func (m *TxManager) Do(ctx context.Context, fn func(context.Context) error) error {
	if _, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return fn(ctx)
	}

	tx, err := m.pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.ReadCommitted,
	})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		rollbackCtx, cancel := context.WithTimeout(
			context.WithoutCancel(ctx),
			m.rollbackTimeout,
		)
		defer cancel()
		err := tx.Rollback(rollbackCtx)
		if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			m.logger.ErrorContext(rollbackCtx, "rollback failed", "error", err)
		}
	}()

	txCtx := context.WithValue(ctx, txKey{}, tx)
	err = fn(txCtx)
	if err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}
