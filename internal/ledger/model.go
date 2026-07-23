package ledger

import (
	"encoding/json"

	"github.com/google/uuid"
)

type Ledger struct {
	ID       uuid.UUID       `json:"id"`
	Name     string          `json:"name"`
	Metadata json.RawMessage `json:"metadata"`
}

type CreateLedgerRequest struct {
	Name     string          `json:"name"`
	Metadata json.RawMessage `json:"metadata"`
}
