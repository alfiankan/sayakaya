package entities

import "github.com/google/uuid"

type RedeemRequest struct {
	UserID            uuid.UUID `json:"user_id"`
	TransactionAmount float64   `json:"transaction_amount"`
}
