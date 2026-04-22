package entities

import "errors"

var (
	ErrVoucherNotFound  = errors.New("voucher not found")
	ErrVoucherExpired   = errors.New("voucher expired")
	ErrVoucherExhausted = errors.New("voucher max uses reached")
	ErrAlreadyClaimed   = errors.New("user already claimed this voucher")
	ErrClaimNotFound    = errors.New("claim not found")
	ErrAlreadyRedeemed  = errors.New("claim already redeemed")
	ErrMinAmountNotMet  = errors.New("minimum transaction amount not met")
)
