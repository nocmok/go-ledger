package ledger

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var (
	ErrNotFound = errors.New("not found")
)

type Service interface {
	Create(ctx context.Context, idempotencyKey uuid.UUID, name string, metadata json.RawMessage) (Ledger, error)
	Get(ctx context.Context, id uuid.UUID) (Ledger, error)
}

type service struct {
	repository Repository
}

func NewService(repository Repository) Service {
	return &service{
		repository: repository,
	}
}

func (s *service) Create(ctx context.Context, idempotencyKey uuid.UUID, name string, metadata json.RawMessage) (Ledger, error) {
	// todo check idempotency
	return s.repository.Create(ctx, name, metadata)
}

func (s *service) Get(ctx context.Context, id uuid.UUID) (Ledger, error) {
	ledger, err := s.repository.Get(ctx, id)
	if err == nil {
		return ledger, nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return Ledger{}, ErrNotFound
	}
	return Ledger{}, err
}
