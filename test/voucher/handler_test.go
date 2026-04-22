package voucher_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sayakaya/pkg/ratelimiter"
	"sayakaya/pkg/voucher"
	"sayakaya/pkg/voucher/entities"
	"testing"
	"time"
	"github.com/google/uuid"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestVoucherHandler_Claim(t *testing.T) {
	cleanDB(t)
	vRepo := voucher.NewPostgresVoucherRepository(testDB)
	cRepo := voucher.NewPostgresClaimRepository(testDB)
	svc := voucher.NewVoucherService(testDB, vRepo, cRepo)
	limiter := ratelimiter.NewManager(100, 100)
	h := voucher.NewVoucherHandler(svc, limiter)
	e := echo.New()

	t.Run("Success", func(t *testing.T) {
		cleanDB(t)
		code := "PROMO"
		userID := uuid.New()

		v := &entities.Voucher{
			Code:                 code,
			DiscountType:         entities.DiscountTypeFlat,
			DiscountValue:        5000,
			MaxUses:              10,
			ExpiresAt:            time.Now().Add(1 * time.Hour),
			MinTransactionAmount: 0,
		}
		svc.CreateVoucher(context.Background(), v)

		reqBody := entities.ClaimRequest{UserID: userID}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/vouchers/"+code+"/claim", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("code")
		c.SetParamValues(code)

		if assert.NoError(t, h.Claim(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			var resp entities.Claim
			json.Unmarshal(rec.Body.Bytes(), &resp)
			assert.Equal(t, userID, resp.UserID)
		}
	})
}

func TestVoucherHandler_FullIntegration(t *testing.T) {
	cleanDB(t)
	vRepo := voucher.NewPostgresVoucherRepository(testDB)
	cRepo := voucher.NewPostgresClaimRepository(testDB)
	svc := voucher.NewVoucherService(testDB, vRepo, cRepo)
	limiter := ratelimiter.NewManager(100, 100)
	h := voucher.NewVoucherHandler(svc, limiter)
	e := echo.New()

	t.Run("Full Lifecycle", func(t *testing.T) {
		code := "API_LIFECYCLE"
		userID := uuid.New()
		vReq := entities.Voucher{
			Code:                 code,
			DiscountType:         entities.DiscountTypeFlat,
			DiscountValue:        5000,
			MaxUses:              1,
			ExpiresAt:            time.Now().Add(24 * time.Hour),
			MinTransactionAmount: 10000,
		}
		body, _ := json.Marshal(vReq)
		req := httptest.NewRequest(http.MethodPost, "/vouchers", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if assert.NoError(t, h.Create(c)) {
			assert.Equal(t, http.StatusCreated, rec.Code)
		}

		claimReq := entities.ClaimRequest{UserID: userID}
		body, _ = json.Marshal(claimReq)
		req = httptest.NewRequest(http.MethodPost, "/vouchers/"+code+"/claim", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec = httptest.NewRecorder()
		c = e.NewContext(req, rec)
		c.SetParamNames("code")
		c.SetParamValues(code)

		if assert.NoError(t, h.Claim(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}

		redeemReq := entities.RedeemRequest{
			UserID:            userID,
			TransactionAmount: 15000,
		}
		body, _ = json.Marshal(redeemReq)
		req = httptest.NewRequest(http.MethodPost, "/vouchers/"+code+"/redeem", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec = httptest.NewRecorder()
		c = e.NewContext(req, rec)
		c.SetParamNames("code")
		c.SetParamValues(code)

		if assert.NoError(t, h.Redeem(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			var resp entities.RedeemResponse
			json.Unmarshal(rec.Body.Bytes(), &resp)
			assert.Equal(t, 5000.0, *resp.DiscountApplied)
			assert.Equal(t, 10000.0, *resp.FinalAmount)
		}
	})
}
