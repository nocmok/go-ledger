package transaction

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/nocmok/go-ledger/internal/currency"
)

type TransactionStatus string

const (
	TransactionStatusPending TransactionStatus = "PENDING"
	TransactionStatusSuccess TransactionStatus = "SUCCESS"
	TransactionStatusFailure TransactionStatus = "FAILURE"
)

type Transaction struct {
	LedgerId     uuid.UUID         `json:"ledgerId"`
	Id           uuid.UUID         `json:"id"`
	Currency     currency.Currency `json:"currency"`
	Status       TransactionStatus `json:"status"`
	ErrorCode    *string           `json:"errorCode"`
	ErrorMessage *string           `json:"errorMessage"`
	Entries      []Entry           `json:"entries"`
	Metadata     json.RawMessage   `json:"metadata"`
}

type TransactionEntity struct {
	LedgerId     uuid.UUID         `json:"ledgerId"`
	Id           uuid.UUID         `json:"id"`
	Currency     currency.Currency `json:"currency"`
	Status       TransactionStatus `json:"status"`
	ErrorCode    *string           `json:"errorCode"`
	ErrorMessage *string           `json:"errorMessage"`
	Metadata     json.RawMessage   `json:"metadata"`
}

type Entry struct {
	AccountId uuid.UUID `json:"accountId"`
	Amount    int64     `json:"amount"`
}
