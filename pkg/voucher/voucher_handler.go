package voucher

import (
	"net/http"
	"sayakaya/pkg/logger"
	"sayakaya/pkg/ratelimiter"
	"sayakaya/pkg/voucher/entities"
	"strconv"

	"github.com/google/uuid"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

type VoucherHandler struct {
	service entities.VoucherServiceInterface
	limiter *ratelimiter.Manager
}

func NewVoucherHandler(s entities.VoucherServiceInterface, l *ratelimiter.Manager) *VoucherHandler {
	return &VoucherHandler{
		service: s,
		limiter: l,
	}
}

func (h *VoucherHandler) RegisterRoutes(e *echo.Echo) {
	e.POST("/vouchers", h.Create)
	e.GET("/vouchers/:code", h.Get)
	e.POST("/vouchers/:code/claim", h.Claim)
	e.POST("/vouchers/:code/redeem", h.Redeem)
	e.GET("/vouchers/:code/claims", h.ListClaims)
	e.GET("/vouchers/report/summary", h.Report)
}

func (h *VoucherHandler) Create(ctx echo.Context) error {
	echoCtx := ctx.Request().Context()

	voucher := new(entities.Voucher)
	if err := ctx.Bind(voucher); err != nil {
		logger.Error(echoCtx, "Failed to bind voucher", err, nil)
		return ctx.JSON(http.StatusBadRequest, entities.ErrorResponse{Error: "invalid request body"})
	}

	if err := h.service.CreateVoucher(echoCtx, voucher); err != nil {
		logger.Error(echoCtx, "Failed to create voucher", err, logrus.Fields{"code": voucher.Code})
		return ctx.JSON(http.StatusInternalServerError, entities.ErrorResponse{Error: err.Error()})
	}

	return ctx.JSON(http.StatusCreated, voucher)
}

func (h *VoucherHandler) Get(ctx echo.Context) error {
	echoCtx := ctx.Request().Context()
	code := ctx.Param("code")

	voucher, err := h.service.GetVoucher(echoCtx, code)
	if err != nil {
		if err == entities.ErrVoucherNotFound {
			return ctx.JSON(http.StatusNotFound, entities.ErrorResponse{Error: err.Error()})
		}
		logger.Error(echoCtx, "Failed to get voucher", err, logrus.Fields{"code": code})
		return ctx.JSON(http.StatusInternalServerError, entities.ErrorResponse{Error: err.Error()})
	}

	resp := entities.VoucherResponse{
		Code:                 voucher.Code,
		DiscountType:         voucher.DiscountType,
		DiscountValue:        voucher.DiscountValue,
		MaxUses:              voucher.MaxUses,
		RemainingClaims:      voucher.MaxUses - voucher.TotalClaims,
		ExpiresAt:            voucher.ExpiresAt,
		MinTransactionAmount: voucher.MinTransactionAmount,
	}

	return ctx.JSON(http.StatusOK, resp)
}

func (h *VoucherHandler) Claim(ctx echo.Context) error {
	echoCtx := ctx.Request().Context()
	code := ctx.Param("code")
	req := new(entities.ClaimRequest)

	if err := ctx.Bind(req); err != nil {
		logger.Error(echoCtx, "Failed to bind claim request", err, nil)
		return ctx.JSON(http.StatusBadRequest, entities.ErrorResponse{Error: "invalid request body"})
	}

	if req.UserID == uuid.Nil {
		return ctx.JSON(http.StatusBadRequest, entities.ErrorResponse{Error: "user_id is required"})
	}

	if !h.limiter.Allow(req.UserID.String()) {
		return ctx.JSON(http.StatusTooManyRequests, entities.ErrorResponse{Error: "rate limit exceeded"})
	}

	claim, err := h.service.ClaimVoucher(echoCtx, code, req.UserID)
	if err != nil {
		switch err {
		case entities.ErrVoucherNotFound:
			return ctx.JSON(http.StatusNotFound, entities.ErrorResponse{Error: err.Error()})
		case entities.ErrVoucherExpired, entities.ErrVoucherExhausted, entities.ErrAlreadyClaimed:
			return ctx.JSON(http.StatusUnprocessableEntity, entities.ErrorResponse{Error: err.Error()})
		default:
			logger.Error(echoCtx, "Failed to claim voucher", err, logrus.Fields{"code": code, "user_id": req.UserID})
			return ctx.JSON(http.StatusInternalServerError, entities.ErrorResponse{Error: err.Error()})
		}
	}

	return ctx.JSON(http.StatusOK, claim)
}

func (h *VoucherHandler) Redeem(ctx echo.Context) error {
	echoCtx := ctx.Request().Context()
	code := ctx.Param("code")
	req := new(entities.RedeemRequest)

	if err := ctx.Bind(req); err != nil {
		logger.Error(echoCtx, "Failed to bind redeem request", err, nil)
		return ctx.JSON(http.StatusBadRequest, entities.ErrorResponse{Error: "invalid request body"})
	}

	claim, err := h.service.RedeemVoucher(echoCtx, code, req.UserID, req.TransactionAmount)
	if err != nil {
		switch err {
		case entities.ErrVoucherNotFound, entities.ErrClaimNotFound:
			return ctx.JSON(http.StatusNotFound, entities.ErrorResponse{Error: err.Error()})
		case entities.ErrMinAmountNotMet:
			return ctx.JSON(http.StatusUnprocessableEntity, entities.ErrorResponse{Error: err.Error()})
		default:
			logger.Error(echoCtx, "Failed to redeem voucher", err, logrus.Fields{"code": code, "user_id": req.UserID})
			return ctx.JSON(http.StatusInternalServerError, entities.ErrorResponse{Error: err.Error()})
		}
	}

	resp := entities.RedeemResponse{
		OriginalAmount:  req.TransactionAmount,
		DiscountApplied: claim.DiscountApplied,
		FinalAmount:     claim.FinalAmount,
	}

	return ctx.JSON(http.StatusOK, resp)
}

func (h *VoucherHandler) ListClaims(ctx echo.Context) error {
	echoCtx := ctx.Request().Context()
	code := ctx.Param("code")
	status := ctx.QueryParam("status")
	page, _ := strconv.Atoi(ctx.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(ctx.QueryParam("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	claims, err := h.service.ListClaims(echoCtx, code, status, page, limit)
	if err != nil {
		if err == entities.ErrVoucherNotFound {
			return ctx.JSON(http.StatusNotFound, entities.ErrorResponse{Error: err.Error()})
		}
		logger.Error(echoCtx, "Failed to list claims", err, logrus.Fields{"code": code})
		return ctx.JSON(http.StatusInternalServerError, entities.ErrorResponse{Error: err.Error()})
	}

	return ctx.JSON(http.StatusOK, claims)
}

func (h *VoucherHandler) Report(ctx echo.Context) error {
	echoCtx := ctx.Request().Context()

	reports, err := h.service.GetSummaryReport(echoCtx)
	if err != nil {
		logger.Error(echoCtx, "Failed to get summary report", err, nil)
		return ctx.JSON(http.StatusInternalServerError, entities.ErrorResponse{Error: err.Error()})
	}
	return ctx.JSON(http.StatusOK, reports)
}
