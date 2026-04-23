package entities

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type VoucherRepository interface {
	Create(ctx context.Context, v *Voucher) error
	GetByCode(ctx context.Context, code string) (*Voucher, error)
	IncrementTotalClaims(ctx context.Context, tx *sql.Tx, id uuid.UUID) (bool, error)
	GetSummaryReport(ctx context.Context) ([]VoucherReport, error)
}

type ClaimRepository interface {
	Create(ctx context.Context, tx *sql.Tx, c *Claim) (bool, error)
	GetByUserAndVoucher(ctx context.Context, userID uuid.UUID, voucherID uuid.UUID) (*Claim, error)
	Update(ctx context.Context, tx *sql.Tx, c *Claim) error
	ListByVoucher(ctx context.Context, voucherID uuid.UUID, status string, page, limit int) ([]Claim, error)
}

type VoucherServiceInterface interface {
	CreateVoucher(ctx context.Context, v *Voucher) error
	GetVoucher(ctx context.Context, code string) (*Voucher, error)
	ClaimVoucher(ctx context.Context, code string, userID uuid.UUID) (*Claim, error)
	RedeemVoucher(ctx context.Context, code string, userID uuid.UUID, amount float64) (*Claim, error)
	ListClaims(ctx context.Context, code string, status string, page, limit int) ([]Claim, error)
	GetSummaryReport(ctx context.Context) ([]VoucherReport, error)
}
