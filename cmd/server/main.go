package main

import (
	"database/sql"
	"log"
	"sayakaya/pkg/config"
	customMiddleware "sayakaya/pkg/middleware"
	"sayakaya/pkg/ratelimiter"
	"sayakaya/pkg/voucher"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"
)

func main() {
	cfg := config.Load()
	m, err := migrate.New("file://migrations", cfg.DatabaseURL)

	if err != nil {
		log.Fatalf("failed automigrate: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("failed automigrate: %v", err)
	}

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect db: %v", err)
	}

	defer db.Close()

	// dependency injection
	voucherRepo := voucher.NewPostgresVoucherRepository(db)
	claimRepo := voucher.NewPostgresClaimRepository(db)
	voucherService := voucher.NewVoucherService(db, voucherRepo, claimRepo)
	limiter := ratelimiter.NewManager(float64(cfg.RateLimitRate), float64(cfg.RateLimitBurst))
	voucherHandler := voucher.NewVoucherHandler(voucherService, limiter)

	e := echo.New()
	e.Use(middleware.RequestID())
	e.Use(customMiddleware.TraceMiddleware())
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	voucherHandler.RegisterRoutes(e)
	e.Logger.Fatal(e.Start(cfg.ServerPort))
}
