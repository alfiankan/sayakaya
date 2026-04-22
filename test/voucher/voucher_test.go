package voucher_test

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"sayakaya/pkg/voucher"
	"sayakaya/pkg/voucher/entities"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/assert"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not construct pool: %s", err)
	}
	err = pool.Client.Ping()
	if err != nil {
		log.Fatalf("Could not connect to Docker: %s", err)
	}

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "15-alpine",
		Env: []string{
			"POSTGRES_PASSWORD=password",
			"POSTGRES_USER=postgres",
			"POSTGRES_DB=promoengine_test",
			"listen_addresses='*'",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	hostAndPort := resource.GetHostPort("5432/tcp")
	databaseUrl := fmt.Sprintf("postgres://postgres:password@%s/promoengine_test?sslmode=disable", hostAndPort)
	log.Println("Connecting to database on url: ", databaseUrl)

	resource.Expire(120)

	if err := pool.Retry(func() error {
		var err error
		testDB, err = sql.Open("postgres", databaseUrl)
		if err != nil {
			return err
		}
		return testDB.Ping()
	}); err != nil {
		log.Fatalf("Could not connect to database: %s", err)
	}

	mig, err := migrate.New("file://../../migrations", databaseUrl)
	if err != nil {
		log.Fatalf("Could not create migrate instance: %s", err)
	}
	if err := mig.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Could not run migrations: %s", err)
	}

	code := m.Run()

	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

func cleanDB(t *testing.T) {
	_, _ = testDB.Exec("DELETE FROM claims")
	_, _ = testDB.Exec("DELETE FROM vouchers")
}

func TestVoucherService_RedeemVoucher(t *testing.T) {
	vRepo := voucher.NewPostgresVoucherRepository(testDB)
	cRepo := voucher.NewPostgresClaimRepository(testDB)
	svc := voucher.NewVoucherService(testDB, vRepo, cRepo)
	ctx := context.Background()

	t.Run("Success Percent Discount", func(t *testing.T) {
		cleanDB(t)
		code := "DISCOUNT50"
		userID := uuid.New()

		v := &entities.Voucher{
			Code:                 code,
			DiscountType:         entities.DiscountTypePercent,
			DiscountValue:        50,
			MinTransactionAmount: 100000,
			MaxUses:              10,
			ExpiresAt:            time.Now().Add(1 * time.Hour),
		}
		err := svc.CreateVoucher(ctx, v)
		assert.NoError(t, err)

		claim, err := svc.ClaimVoucher(ctx, code, userID)
		assert.NoError(t, err)

		res, err := svc.RedeemVoucher(ctx, code, userID, 200000)
		assert.NoError(t, err)
		assert.Equal(t, entities.ClaimStatusRedeemed, res.Status)
		assert.Equal(t, 100000.0, *res.DiscountApplied)
		assert.Equal(t, 100000.0, *res.FinalAmount)
		assert.Equal(t, claim.ID, res.ID)
	})

	t.Run("Min Amount Not Met", func(t *testing.T) {
		cleanDB(t)
		code := "MIN_AMT"
		userID := uuid.New()

		v := &entities.Voucher{
			Code:                 code,
			DiscountType:         entities.DiscountTypeFlat,
			DiscountValue:        10000,
			MinTransactionAmount: 100000,
			MaxUses:              10,
			ExpiresAt:            time.Now().Add(1 * time.Hour),
		}
		svc.CreateVoucher(ctx, v)
		svc.ClaimVoucher(ctx, code, userID)

		res, err := svc.RedeemVoucher(ctx, code, userID, 50000)
		assert.ErrorIs(t, err, entities.ErrMinAmountNotMet)
		assert.Nil(t, res)
	})
}

func TestVoucherService_ClaimConcurrency_Integration(t *testing.T) {
	cleanDB(t)
	vRepo := voucher.NewPostgresVoucherRepository(testDB)
	cRepo := voucher.NewPostgresClaimRepository(testDB)
	svc := voucher.NewVoucherService(testDB, vRepo, cRepo)
	ctx := context.Background()

	code := "CONCURRENT_VOUCHER"
	v := &entities.Voucher{
		Code:                 code,
		DiscountType:         entities.DiscountTypeFlat,
		DiscountValue:        10000,
		MaxUses:              1,
		ExpiresAt:            time.Now().Add(1 * time.Hour),
		MinTransactionAmount: 50000,
	}

	if err := svc.CreateVoucher(ctx, v); err != nil {
		t.Fatalf("failed to create voucher: %v", err)
	}

	var wg sync.WaitGroup
	numRequests := 20
	results := make(chan error, numRequests)

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			userID := uuid.New()
			_, err := svc.ClaimVoucher(ctx, code, userID)
			results <- err
		}(i)
	}

	wg.Wait()
	close(results)

	successCount := 0
	for err := range results {
		if err == nil {
			successCount++
		}
	}

	assert.Equal(t, 1, successCount)
}

func TestVoucherService_SameUserConcurrency_Integration(t *testing.T) {
	cleanDB(t)
	vRepo := voucher.NewPostgresVoucherRepository(testDB)
	cRepo := voucher.NewPostgresClaimRepository(testDB)
	svc := voucher.NewVoucherService(testDB, vRepo, cRepo)
	ctx := context.Background()

	code := "SAME_USER_RACE"
	v := &entities.Voucher{
		Code:                 code,
		DiscountType:         entities.DiscountTypeFlat,
		DiscountValue:        10000,
		MaxUses:              10,
		ExpiresAt:            time.Now().Add(1 * time.Hour),
		MinTransactionAmount: 0,
	}

	if err := svc.CreateVoucher(ctx, v); err != nil {
		t.Fatalf("failed to create voucher: %v", err)
	}

	var wg sync.WaitGroup
	numRequests := 10
	userID := uuid.New()
	results := make(chan error, numRequests)

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := svc.ClaimVoucher(ctx, code, userID)
			results <- err
		}()
	}

	wg.Wait()
	close(results)

	successCount := 0
	alreadyClaimedCount := 0
	for err := range results {
		if err == nil {
			successCount++
		} else if err == entities.ErrAlreadyClaimed {
			alreadyClaimedCount++
		}
	}

	assert.Equal(t, 1, successCount, "Only one claim should succeed for the same user")
	assert.Equal(t, 9, alreadyClaimedCount, "All other concurrent requests should error")

	voucherAfter, _ := vRepo.GetByCode(ctx, code)
	assert.Equal(t, 1, voucherAfter.TotalClaims, "Total claims in DB should be exactly 1")
}
