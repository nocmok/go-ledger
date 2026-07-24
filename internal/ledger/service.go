package ledger

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
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
	return s.repository.Create(ctx, idempotencyKey, name, metadata)
}

func (s *service) Get(ctx context.Context, id uuid.UUID) (Ledger, error) {
	return s.Get(ctx, id)
}
