package entities

type RedeemResponse struct {
	OriginalAmount  float64  `json:"original_amount"`
	DiscountApplied *float64 `json:"discount_applied"`
	FinalAmount     *float64 `json:"final_amount"`
}
