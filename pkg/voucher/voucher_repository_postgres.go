package voucher

import (
	"context"
	"database/sql"
	"sayakaya/pkg/logger"
	"sayakaya/pkg/voucher/entities"

	"github.com/google/uuid"
)

type PostgresVoucherRepository struct {
	db *sql.DB
}

func NewPostgresVoucherRepository(db *sql.DB) *PostgresVoucherRepository {
	return &PostgresVoucherRepository{db: db}
}

func (r *PostgresVoucherRepository) Create(ctx context.Context, voucher *entities.Voucher) error {
	query := `
		INSERT INTO vouchers (code, discount_type, discount_value, max_uses, expires_at, min_transaction_amount)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at`

	return r.db.QueryRowContext(ctx, query,
		voucher.Code, voucher.DiscountType, voucher.DiscountValue, voucher.MaxUses, voucher.ExpiresAt, voucher.MinTransactionAmount,
	).Scan(&voucher.ID, &voucher.CreatedAt, &voucher.UpdatedAt)
}

func (r *PostgresVoucherRepository) GetByCode(ctx context.Context, code string) (*entities.Voucher, error) {
	voucher := &entities.Voucher{}
	query := `SELECT id, code, discount_type, discount_value, max_uses, total_claims, expires_at, min_transaction_amount, created_at, updated_at FROM vouchers WHERE code = $1`

	err := r.db.QueryRowContext(ctx, query, code).Scan(
		&voucher.ID, &voucher.Code, &voucher.DiscountType, &voucher.DiscountValue, &voucher.MaxUses, &voucher.TotalClaims, &voucher.ExpiresAt, &voucher.MinTransactionAmount, &voucher.CreatedAt, &voucher.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return voucher, err
}

func (r *PostgresVoucherRepository) IncrementTotalClaims(ctx context.Context, tx *sql.Tx, id uuid.UUID) (bool, error) {
	query := `UPDATE vouchers SET total_claims = total_claims + 1 WHERE id = $1 AND total_claims < max_uses RETURNING id`

	var returnedID uuid.UUID
	var err error
	if tx != nil {
		err = tx.QueryRowContext(ctx, query, id).Scan(&returnedID)
	} else {
		err = r.db.QueryRowContext(ctx, query, id).Scan(&returnedID)
	}

	if err == sql.ErrNoRows {
		return false, nil
	}

	return err == nil, err
}

func (r *PostgresVoucherRepository) GetSummaryReport(ctx context.Context) ([]entities.VoucherReport, error) {
	query := `
		select 
			v.code, 
			count(1) total, 
			SUM(
				CASE WHEN c.status = 'REDEEMED' THEN 1 ELSE 0 END
			) total_redemptions, 
			COUNT(1) total_claims, 
			sum(discount_applied) total_discount_granted, 
			avg(transaction_amount) avg_transaction_amount 
		from 
			claims c 
		join vouchers v on v.id = c.voucher_id 
		group by 
			v.code;`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		logger.Error(ctx, "Repo: GetSummaryReport failed", err, nil)
		return nil, err
	}

	defer rows.Close()

	var reports []entities.VoucherReport
	for rows.Next() {
		var report entities.VoucherReport
		if err := rows.Scan(&report.Code, &report.TotalClaims, &report.TotalRedemptions, &report.TotalDiscountGranted, &report.AvgTransactionAmount); err != nil {
			logger.Error(ctx, "Repo: GetSummaryReport scan failed", err, nil)
			return nil, err
		}
		reports = append(reports, report)
	}

	return reports, nil
}
