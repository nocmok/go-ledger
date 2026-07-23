package account

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/nocmok/go-ledger/internal/account/query"
)

type Repository interface {
	Create(ctx context.Context, idempotencyKey uuid.UUID, ledgerId uuid.UUID, name string, currency Currency, metadata json.RawMessage) (Account, error)
}

type repository struct {
	q *query.Queries
}

func NewRepository(db query.DBTX) Repository {
	return &repository{q: query.New(db)}
}

func (r *repository) Create(ctx context.Context, idempotencyKey uuid.UUID, ledgerId uuid.UUID, name string, currency Currency, metadata json.RawMessage) (Account, error) {
	row, err := r.q.CreateAccount(ctx, query.CreateAccountParams{
		LedgerID:       ledgerId,
		Name:           name,
		Currency:       string(currency),
		Metadata:       metadata,
		Status:         string(StatusActive),
		IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		return Account{}, err
	}
	return Account{
		ID:       row.ID,
		LedgerID: row.LedgerID,
		Name:     row.Name,
		Currency: Currency(row.Currency),
		Metadata: row.Metadata,
		Status:   Status(row.Status),
	}, nil
}

func (r *repository) Get(ctx context.Context, id uuid.UUID) (Account, error) {
	row, err := r.q.GetAccount(ctx, id)
	if err != nil {
		return Account{}, err
	}
	return Account{
		ID:       row.ID,
		LedgerID: row.LedgerID,
		Name:     row.Name,
		Currency: Currency(row.Currency),
		Metadata: row.Metadata,
		Status:   Status(row.Status),
	}, nil
}
