package entities

import (
	"time"
	"github.com/google/uuid"
)

type DiscountType string

const (
	DiscountTypeFlat    DiscountType = "flat"
	DiscountTypePercent DiscountType = "percent"
)

type Voucher struct {
	ID                   uuid.UUID    `json:"id" db:"id"`
	Code                 string       `json:"code" db:"code"`
	DiscountType         DiscountType `json:"discount_type" db:"discount_type"`
	DiscountValue        float64      `json:"discount_value" db:"discount_value"`
	MaxUses              int          `json:"max_uses" db:"max_uses"`
	TotalClaims          int          `json:"total_claims" db:"total_claims"`
	ExpiresAt            time.Time    `json:"expires_at" db:"expires_at"`
	MinTransactionAmount float64      `json:"min_transaction_amount" db:"min_transaction_amount"`
	CreatedAt            time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time    `json:"updated_at" db:"updated_at"`
}
