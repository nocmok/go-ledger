package account

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/nocmok/go-ledger/internal/currency"
)

type Status string

const (
	StatusActive  Status = "ACTIVE"
	StatusFrozen  Status = "FROZEN"
	StatusBlocked Status = "BLOCKED"
)

type Account struct {
	ID               uuid.UUID         `json:"id"`
	LedgerID         uuid.UUID         `json:"ledgerId"`
	Name             string            `json:"name"`
	Currency         currency.Currency `json:"currency"`
	Balance          int64             `json:"balance"`
	AvailableBalance int64             `json:"availableBalance"`
	OverdraftAllowed bool              `json:"overdraftAllowed"`
	Metadata         json.RawMessage   `json:"metadata"`
	Status           Status            `json:"status"`
}
