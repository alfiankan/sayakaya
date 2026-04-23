package voucher

import (
	"context"
	"database/sql"
	"sayakaya/pkg/logger"
	"sayakaya/pkg/voucher/entities"
	"github.com/google/uuid"

	"github.com/sirupsen/logrus"
)

type PostgresClaimRepository struct {
	db *sql.DB
}

func NewPostgresClaimRepository(db *sql.DB) *PostgresClaimRepository {
	return &PostgresClaimRepository{db: db}
}

func (r *PostgresClaimRepository) Create(ctx context.Context, tx *sql.Tx, claim *entities.Claim) (bool, error) {
	query := `
		INSERT INTO claims (voucher_id, user_id, status)
		VALUES ($1, $2, $3)
		ON CONFLICT (voucher_id, user_id) DO NOTHING
		RETURNING id, created_at, updated_at`
	var err error
	if tx != nil {
		err = tx.QueryRowContext(ctx, query, claim.VoucherID, claim.UserID, claim.Status).Scan(&claim.ID, &claim.CreatedAt, &claim.UpdatedAt)
	} else {
		err = r.db.QueryRowContext(ctx, query, claim.VoucherID, claim.UserID, claim.Status).Scan(&claim.ID, &claim.CreatedAt, &claim.UpdatedAt)
	}
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

func (r *PostgresClaimRepository) GetByUserAndVoucher(ctx context.Context, userID uuid.UUID, voucherID uuid.UUID) (*entities.Claim, error) {
	claim := &entities.Claim{}
	query := `SELECT id, voucher_id, user_id, status, transaction_amount, discount_applied, final_amount, created_at, updated_at FROM claims WHERE user_id = $1 AND voucher_id = $2`
	err := r.db.QueryRowContext(ctx, query, userID, voucherID).Scan(
		&claim.ID, &claim.VoucherID, &claim.UserID, &claim.Status, &claim.TransactionAmount, &claim.DiscountApplied, &claim.FinalAmount, &claim.CreatedAt, &claim.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return claim, err
}

func (r *PostgresClaimRepository) Update(ctx context.Context, tx *sql.Tx, claim *entities.Claim) error {
	query := `
		UPDATE claims 
		SET status = $1, transaction_amount = $2, discount_applied = $3, final_amount = $4, updated_at = CURRENT_TIMESTAMP
		WHERE id = $5`
	var err error
	if tx != nil {
		_, err = tx.ExecContext(ctx, query, claim.Status, claim.TransactionAmount, claim.DiscountApplied, claim.FinalAmount, claim.ID)
	} else {
		_, err = r.db.ExecContext(ctx, query, claim.Status, claim.TransactionAmount, claim.DiscountApplied, claim.FinalAmount, claim.ID)
	}
	if err != nil {
		logger.Error(ctx, "Repo: Update Claim failed", err, logrus.Fields{"claim_id": claim.ID})
	}
	return err
}

func (r *PostgresClaimRepository) ListByVoucher(ctx context.Context, voucherID uuid.UUID, status string, page, limit int) ([]entities.Claim, error) {
	offset := (page - 1) * limit
	query := `
		SELECT id, voucher_id, user_id, status, transaction_amount, discount_applied, final_amount, created_at, updated_at 
		FROM claims 
		WHERE voucher_id = $1`
	args := []interface{}{voucherID}
	argCount := 1
	if status != "" {
		argCount++
		query += ` AND status = $2`
		args = append(args, status)
	}
	query += ` ORDER BY created_at DESC LIMIT $` + string(rune('0'+argCount+1)) + ` OFFSET $` + string(rune('0'+argCount+2))
	args = append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		logger.Error(ctx, "Repo: ListByVoucher failed", err, logrus.Fields{"voucher_id": voucherID})
		return nil, err
	}
	defer rows.Close()
	var claims []entities.Claim
	for rows.Next() {
		var claim entities.Claim
		if err := rows.Scan(&claim.ID, &claim.VoucherID, &claim.UserID, &claim.Status, &claim.TransactionAmount, &claim.DiscountApplied, &claim.FinalAmount, &claim.CreatedAt, &claim.UpdatedAt); err != nil {
			logger.Error(ctx, "Repo: ListByVoucher scan failed", err, nil)
			return nil, err
		}
		claims = append(claims, claim)
	}
	return claims, nil
}
