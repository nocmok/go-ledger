package account

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/nocmok/go-ledger/internal/ledger"
)

var (
	ErrLedgerDoesNotExist = errors.New("ledger doesn't exist")
)

type Service interface {
	Create(ctx context.Context, idempotencyKey uuid.UUID, ledgerId uuid.UUID, name string, currency Currency, metadata json.RawMessage) (Account, error)
	Get(ctx context.Context, id uuid.UUID) (Account, bool, error)
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

func (s *service) Create(ctx context.Context, idempotencyKey uuid.UUID, ledgerId uuid.UUID, name string, currency Currency, metadata json.RawMessage) (Account, error) {
	_, ok, err := s.ledgerService.Get(ctx, ledgerId)
	if err != nil {
		return Account{}, err
	}
	if !ok {
		return Account{}, ErrLedgerDoesNotExist
	}
	return s.repository.Create(ctx, idempotencyKey, ledgerId, name, currency, metadata)
}

func (s *service) Get(ctx context.Context, id uuid.UUID) (Account, bool, error) {
	account, err := s.repository.Get(ctx, id)
	if err == nil {
		return account, true, nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return Account{}, false, nil
	}
	return Account{}, false, err
}
