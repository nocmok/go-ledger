package transaction

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/nocmok/go-ledger/internal/currency"
	"github.com/nocmok/go-ledger/internal/db"
)

const (
	defaultTransactionTimeout time.Duration = 1 * time.Minute
)

type Service interface {
	CreateTransaction(ctx context.Context, idempotencyKey uuid.UUID, ledgerId uuid.UUID, curr currency.Currency, entries []Entry, metadata json.RawMessage) (Transaction, error)
	GetTransaction(ctx context.Context, idempotencyKey uuid.UUID, ledgerId uuid.UUID, id uuid.UUID) (Transaction, error)
}

func NewService(txManager db.TxManager, repository Repository) Service {
	return &service{
		txManager:  txManager,
		repository: repository,
	}
}

type service struct {
	txManager  db.TxManager
	repository Repository
}

func (s *service) CreateTransaction(ctx context.Context, idempotencyKey uuid.UUID, ledgerId uuid.UUID, curr currency.Currency, entries []Entry, metadata json.RawMessage) (Transaction, error) {
	// todo check idempotency
	// todo validate data
	res, err := s.txManager.RunTransactionally(ctx, func(tx pgx.Tx) (any, error) {
		transactionDeadline := time.Now().Add(defaultTransactionTimeout) // todo use timeout instead of deadline
		transaction, err := s.repository.CreateTransaction(ctx, tx, ledgerId, curr, TransactionStatusPending, transactionDeadline, metadata)
		if err != nil {
			return TransactionEntity{}, err
		}
		if err := s.repository.CreateEntries(ctx, tx, ledgerId, transaction.Id, entries); err != nil {
			return TransactionEntity{}, err
		}
		return transaction, nil
	})
	if err != nil {
		return Transaction{}, err
	}
	transaction := res.(TransactionEntity)
	return Transaction{
		LedgerId:     transaction.LedgerId,
		Id:           transaction.Id,
		Currency:     transaction.Currency,
		Status:       transaction.Status,
		ErrorCode:    transaction.ErrorCode,
		ErrorMessage: transaction.ErrorMessage,
		Entries:      entries,
		Metadata:     transaction.Metadata,
	}, nil
}

func (s *service) GetTransaction(ctx context.Context, idempotencyKey uuid.UUID, ledgerId uuid.UUID, id uuid.UUID) (Transaction, error) {
	// todo check idempotency

	return Transaction{}, errors.New("not implemented")
}
