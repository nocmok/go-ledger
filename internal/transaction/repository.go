package transaction

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nocmok/go-ledger/internal/currency"
	"github.com/nocmok/go-ledger/internal/db"
	"github.com/nocmok/go-ledger/internal/transaction/query"
)

type Repository interface {
	CreateTransaction(ctx context.Context, dbtx db.DBTX, ledgerId uuid.UUID, curr currency.Currency, status TransactionStatus, deadline time.Time, metadata json.RawMessage) (TransactionEntity, error)
	GetTransaction(ctx context.Context, dbtx db.DBTX, ledgerId uuid.UUID, id uuid.UUID) (TransactionEntity, error)
	CreateEntries(ctx context.Context, dbtx db.DBTX, ledgerId uuid.UUID, transactionId uuid.UUID, entries []Entry) error
	GetEntries(ctx context.Context, dbtx db.DBTX, ledgerId uuid.UUID, transactionId uuid.UUID) ([]Entry, error)
}

func NewRepository() Repository {
	return &repository{}
}

type repository struct{}

func (r *repository) CreateTransaction(ctx context.Context, dbtx db.DBTX, ledgerId uuid.UUID, curr currency.Currency, status TransactionStatus, deadline time.Time, metadata json.RawMessage) (TransactionEntity, error) {
	row, err := query.New(dbtx).CreateTransaction(ctx, query.CreateTransactionParams{
		LedgerID: ledgerId,
		Currency: string(curr),
		Status:   string(status),
		Deadline: deadline,
		Metadata: metadata,
	})
	if err != nil {
		return TransactionEntity{}, err
	}
	return TransactionEntity{
		LedgerId:     row.LedgerID,
		Id:           row.ID,
		Currency:     currency.Currency(row.Currency),
		Status:       TransactionStatus(row.Status),
		ErrorCode:    row.ErrorCode,
		ErrorMessage: row.ErrorMessage,
		Metadata:     row.Metadata,
	}, nil
}

func (r *repository) GetTransaction(ctx context.Context, dbtx db.DBTX, ledgerId uuid.UUID, id uuid.UUID) (TransactionEntity, error) {
	return TransactionEntity{}, errors.New("not implemented")
}

func (r *repository) CreateEntries(ctx context.Context, dbtx db.DBTX, ledgerId uuid.UUID, transactionId uuid.UUID, entries []Entry) error {
	entriesToInsert := make([]query.CreateEntriesParams, 0, len(entries))
	for _, entry := range entries {
		entriesToInsert = append(entriesToInsert, query.CreateEntriesParams{
			LedgerID:      ledgerId,
			TransactionID: transactionId,
			AccountID:     entry.AccountId,
			Amount: pgtype.Int8{
				Int64: entry.Amount,
				Valid: true,
			},
		})
	}
	_, err := query.New(dbtx).CreateEntries(ctx, entriesToInsert)
	return err
}

func (r *repository) GetEntries(ctx context.Context, dbtx db.DBTX, ledgerId uuid.UUID, transactionId uuid.UUID) ([]Entry, error) {
	return nil, errors.New("not implemented")
}
