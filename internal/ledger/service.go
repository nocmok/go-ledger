package ledger

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Service interface {
	Create(ctx context.Context, idempotencyKey uuid.UUID, name string, metadata json.RawMessage) (Ledger, error)
	Get(ctx context.Context, id uuid.UUID) (Ledger, bool, error)
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
	return s.repository.Create(ctx, idempotencyKey, name, metadata)
}

func (s *service) Get(ctx context.Context, id uuid.UUID) (Ledger, bool, error) {
	ledger, err := s.repository.Get(ctx, id)
	if err == nil {
		return ledger, true, nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return Ledger{}, false, nil
	}
	return Ledger{}, false, err
}
