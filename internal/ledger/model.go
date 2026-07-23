package ledger

import "github.com/google/uuid"

type Ledger struct {
	ID       uuid.UUID      `json:"id"`
	Name     string         `json:"name"`
	Metadata map[string]any `json:"metadata"`
}

type CreateLedgerRequest struct {
	Name     string         `json:"name"`
	Metadata map[string]any `json:"metadata"`
}
