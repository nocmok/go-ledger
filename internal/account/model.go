package account

import "github.com/google/uuid"

type Status string

const (
	StatusActive  Status = "active"
	StatusFrozen  Status = "frozen"
	StatusBlocked Status = "blocked"
)

type Currency string

const (
	CURRENCY_USD Currency = "USD"
	CURRENCY_EUR Currency = "EUR"
	CURRENCY_BTC Currency = "BTC"
	CURRENCY_ETH Currency = "ETH"
)

type Account struct {
	ID       uuid.UUID      `json:"id"`
	LedgerID uuid.UUID      `json:"ledgerId"`
	Name     string         `json:"name"`
	Currency Currency       `json:"currency"`
	Metadata map[string]any `json:"metadata"`
	Status   Status         `json:"status"`
}

type CreateAccountRequest struct {
	LedgerID uuid.UUID      `json:"ledgerId"`
	Name     string         `json:"name"`
	Currency Currency       `json:"currency"`
	Metadata map[string]any `json:"metadata"`
}
