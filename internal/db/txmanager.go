package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TxManager interface {
	RunTransactionally(ctx context.Context, fn func(pgx.Tx) (any, error)) (any, error)
}

func NewTxManager(pool *pgxpool.Pool) TxManager {
	return &txManager{pool: pool}
}

type txManager struct {
	pool *pgxpool.Pool
}

func (t *txManager) RunTransactionally(ctx context.Context, fn func(pgx.Tx) (any, error)) (any, error) {
	tx, err := t.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	res, err := fn(tx)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return res, nil
}
