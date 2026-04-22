package entities

import "time"

type VoucherResponse struct {
	Code                 string       `json:"code"`
	DiscountType         DiscountType `json:"discount_type"`
	DiscountValue        float64      `json:"discount_value"`
	MaxUses              int          `json:"max_uses"`
	RemainingClaims      int          `json:"remaining_claims"`
	ExpiresAt            time.Time    `json:"expires_at"`
	MinTransactionAmount float64      `json:"min_transaction_amount"`
}
