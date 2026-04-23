package voucher

import (
	"context"
	"database/sql"
	"sayakaya/pkg/logger"
	"sayakaya/pkg/voucher/entities"
	"time"

	"github.com/google/uuid"

	"github.com/sirupsen/logrus"
)

type VoucherService struct {
	db          *sql.DB
	voucherRepo entities.VoucherRepository
	claimRepo   entities.ClaimRepository
}

func NewVoucherService(db *sql.DB, voucherRepo entities.VoucherRepository, claimRepo entities.ClaimRepository) *VoucherService {
	return &VoucherService{
		db:          db,
		voucherRepo: voucherRepo,
		claimRepo:   claimRepo,
	}
}

func (s *VoucherService) CreateVoucher(ctx context.Context, voucher *entities.Voucher) error {
	err := s.voucherRepo.Create(ctx, voucher)
	if err != nil {
		logger.Error(ctx, "Service: CreateVoucher failed", err, logrus.Fields{"code": voucher.Code})
	}
	return err
}

func (s *VoucherService) GetVoucher(ctx context.Context, code string) (*entities.Voucher, error) {
	voucher, err := s.voucherRepo.GetByCode(ctx, code)
	if err != nil {
		logger.Error(ctx, "Service: GetVoucher failed", err, logrus.Fields{"code": code})
		return nil, err
	}

	if voucher == nil {
		return nil, entities.ErrVoucherNotFound
	}

	return voucher, nil
}

func (s *VoucherService) ClaimVoucher(ctx context.Context, code string, userID uuid.UUID) (*entities.Claim, error) {
	voucher, err := s.GetVoucher(ctx, code)
	if err != nil {
		return nil, err
	}

	if time.Now().After(voucher.ExpiresAt) {
		return nil, entities.ErrVoucherExpired
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		logger.Error(ctx, "Service: ClaimVoucher begin tx failed", err, nil)
		return nil, err
	}
	defer tx.Rollback()

	dbSuccess, err := s.voucherRepo.IncrementTotalClaims(ctx, tx, voucher.ID)
	if err != nil {
		logger.Error(ctx, "Service: ClaimVoucher increment claims failed", err, logrus.Fields{"code": code})
		return nil, err
	}

	if !dbSuccess {
		return nil, entities.ErrVoucherExhausted
	}

	claim := &entities.Claim{
		VoucherID: voucher.ID,
		UserID:    userID,
		Status:    entities.ClaimStatusClaimed,
	}

	dbSuccess, err = s.claimRepo.Create(ctx, tx, claim)
	if err != nil {
		logger.Error(ctx, "Service: ClaimVoucher create claim failed", err, logrus.Fields{"code": code, "user_id": userID})
		return nil, err
	}
	if !dbSuccess {
		return nil, entities.ErrAlreadyClaimed
	}

	if err := tx.Commit(); err != nil {
		logger.Error(ctx, "Service: ClaimVoucher commit failed", err, nil)
		return nil, err
	}

	return claim, nil
}

func (s *VoucherService) RedeemVoucher(ctx context.Context, code string, userID uuid.UUID, amount float64) (*entities.Claim, error) {
	voucher, err := s.voucherRepo.GetByCode(ctx, code)
	if err != nil {
		logger.Error(ctx, "Service: RedeemVoucher get voucher failed", err, logrus.Fields{"code": code})
		return nil, err
	}

	if voucher == nil {
		return nil, entities.ErrVoucherNotFound
	}

	claim, err := s.claimRepo.GetByUserAndVoucher(ctx, userID, voucher.ID)
	if err != nil {
		logger.Error(ctx, "Service: RedeemVoucher get claim failed", err, logrus.Fields{"code": code, "user_id": userID})
		return nil, err
	}

	if claim == nil {
		return nil, entities.ErrClaimNotFound
	}
	if claim.Status == entities.ClaimStatusRedeemed {
		return claim, nil
	}

	if amount < voucher.MinTransactionAmount {
		return nil, entities.ErrMinAmountNotMet
	}

	var discount float64
	if voucher.DiscountType == entities.DiscountTypeFlat {
		discount = voucher.DiscountValue
	} else {
		discount = amount * (voucher.DiscountValue / 100.0)
	}

	if discount > amount {
		discount = amount
	}

	finalAmount := amount - discount
	claim.Status = entities.ClaimStatusRedeemed
	claim.TransactionAmount = &amount
	claim.DiscountApplied = &discount
	claim.FinalAmount = &finalAmount

	if err := s.claimRepo.Update(ctx, nil, claim); err != nil {
		logger.Error(ctx, "Service: RedeemVoucher update claim failed", err, logrus.Fields{"claim_id": claim.ID})
		return nil, err
	}

	return claim, nil
}

func (s *VoucherService) ListClaims(ctx context.Context, code string, status string, page, limit int) ([]entities.Claim, error) {
	voucher, err := s.voucherRepo.GetByCode(ctx, code)
	if err != nil {
		return nil, err
	}

	if voucher == nil {
		return nil, entities.ErrVoucherNotFound
	}

	return s.claimRepo.ListByVoucher(ctx, voucher.ID, status, page, limit)
}

func (s *VoucherService) GetSummaryReport(ctx context.Context) ([]entities.VoucherReport, error) {
	return s.voucherRepo.GetSummaryReport(ctx)
}
