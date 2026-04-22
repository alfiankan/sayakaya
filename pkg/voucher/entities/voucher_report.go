package entities

type VoucherReport struct {
	Code                 string  `json:"code"`
	TotalClaims          int     `json:"total_claims"`
	TotalRedemptions     int     `json:"total_redemptions"`
	TotalDiscountGranted float64 `json:"total_discount_granted"`
	AvgTransactionAmount float64 `json:"avg_transaction_amount"`
}
