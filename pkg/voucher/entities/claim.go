package entities

import (
	"time"
	"github.com/google/uuid"
)

type ClaimStatus string

const (
	ClaimStatusClaimed  ClaimStatus = "CLAIMED"
	ClaimStatusRedeemed ClaimStatus = "REDEEMED"
)

type Claim struct {
	ID                uuid.UUID   `json:"id" db:"id"`
	VoucherID         uuid.UUID   `json:"voucher_id" db:"voucher_id"`
	UserID            uuid.UUID   `json:"user_id" db:"user_id"`
	Status            ClaimStatus `json:"status" db:"status"`
	TransactionAmount *float64    `json:"transaction_amount,omitempty" db:"transaction_amount"`
	DiscountApplied   *float64    `json:"discount_applied,omitempty" db:"discount_applied"`
	FinalAmount       *float64    `json:"final_amount,omitempty" db:"final_amount"`
	CreatedAt         time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time   `json:"updated_at" db:"updated_at"`
}
