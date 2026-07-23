package ledger

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/nocmok/go-ledger/internal/ledger/query"
)

type Repository interface {
	Create(ctx context.Context, idempotencyKey uuid.UUID, name string, metadata json.RawMessage) (Ledger, error)
	Get(ctx context.Context, id uuid.UUID) (Ledger, error)
}

type repository struct {
	q *query.Queries
}

func NewRepository(db query.DBTX) Repository {
	return &repository{q: query.New(db)}
}

func (r *repository) Create(ctx context.Context, idempotencyKey uuid.UUID, name string, metadata json.RawMessage) (Ledger, error) {
	row, err := r.q.CreateLedger(ctx, query.CreateLedgerParams{
		Name:           name,
		Metadata:       metadata,
		IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		return Ledger{}, err
	}
	return Ledger{
		ID:       row.ID,
		Name:     row.Name,
		Metadata: row.Metadata,
	}, nil
}

func (r *repository) Get(ctx context.Context, id uuid.UUID) (Ledger, error) {
	row, err := r.q.GetLedger(ctx, id)
	if err != nil {
		return Ledger{}, err
	}
	return Ledger{
		ID:       row.ID,
		Name:     row.Name,
		Metadata: row.Metadata,
	}, nil
}
