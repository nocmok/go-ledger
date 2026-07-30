package account

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/nocmok/go-ledger/internal/account/query"
	"github.com/nocmok/go-ledger/internal/currency"
)

type Repository interface {
	Create(ctx context.Context, ledgerId uuid.UUID, name string, currency currency.Currency, metadata json.RawMessage) (Account, error)
	Get(ctx context.Context, ledgerId uuid.UUID, id uuid.UUID) (Account, error)
}

type repository struct {
	q *query.Queries
}

func NewRepository(db query.DBTX) Repository {
	return &repository{q: query.New(db)}
}

func (r *repository) Create(ctx context.Context, ledgerId uuid.UUID, name string, curr currency.Currency, metadata json.RawMessage) (Account, error) {
	row, err := r.q.CreateAccount(ctx, query.CreateAccountParams{
		LedgerID: ledgerId,
		Name:     name,
		Currency: string(curr),
		Metadata: metadata,
		Status:   string(StatusActive),
	})
	if err != nil {
		return Account{}, err
	}
	return Account{
		ID:               row.ID,
		LedgerID:         row.LedgerID,
		Name:             row.Name,
		Currency:         currency.Currency(row.Currency),
		Balance:          row.Balance,
		AvailableBalance: row.AvailableBalance,
		OverdraftAllowed: row.OverdraftAllowed,
		Metadata:         row.Metadata,
		Status:           Status(row.Status),
	}, nil
}

func (r *repository) Get(ctx context.Context, ledgerId uuid.UUID, id uuid.UUID) (Account, error) {
	row, err := r.q.GetAccount(ctx, query.GetAccountParams{
		LedgerID: ledgerId,
		ID:       id,
	})
	if err != nil {
		return Account{}, err
	}
	return Account{
		ID:               row.ID,
		LedgerID:         row.LedgerID,
		Name:             row.Name,
		Currency:         currency.Currency(row.Currency),
		Balance:          row.Balance,
		AvailableBalance: row.AvailableBalance,
		OverdraftAllowed: row.OverdraftAllowed,
		Metadata:         row.Metadata,
		Status:           Status(row.Status),
	}, nil
}
