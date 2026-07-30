package account

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/nocmok/go-ledger/internal/currency"
	"github.com/nocmok/go-ledger/internal/ledger"
)

var (
	ErrLedgerNotFound = errors.New("ledger not found")
	ErrNotFound       = errors.New("not found")
)

type Service interface {
	Create(ctx context.Context, idempotencyKey uuid.UUID, ledgerId uuid.UUID, name string, currency currency.Currency, overdraftAllowed bool, metadata json.RawMessage) (Account, error)
	Get(ctx context.Context, ledgerId uuid.UUID, id uuid.UUID) (Account, error)
}

type service struct {
	repository    Repository
	ledgerService ledger.Service
}

func NewService(repository Repository, ledgerService ledger.Service) Service {
	return &service{
		repository:    repository,
		ledgerService: ledgerService,
	}
}

func (s *service) Create(ctx context.Context, idempotencyKey uuid.UUID, ledgerId uuid.UUID, name string, currency currency.Currency, overdraftAllowed bool, metadata json.RawMessage) (Account, error) {
	// todo check idempotency
	_, err := s.ledgerService.Get(ctx, ledgerId)
	if err != nil {
		if errors.Is(err, ledger.ErrNotFound) {
			return Account{}, ErrLedgerNotFound
		}
		return Account{}, err
	}
	return s.repository.Create(ctx, ledgerId, name, currency, metadata)
}

func (s *service) Get(ctx context.Context, ledgerId uuid.UUID, id uuid.UUID) (Account, error) {
	account, err := s.repository.Get(ctx, ledgerId, id)
	if err == nil {
		return account, nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return Account{}, ErrNotFound
	}
	return Account{}, err
}
