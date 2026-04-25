# Sayakaya Voucher Service


## Tech Stack

- **Language:** Go
- **Web Framework:** Echo
- **Database:** PostgreSQL
- **Testing:** `dockertest` 

## Concurrency Handling

Using pesimistic locking to handle high concurrency claims without race conditions.

## How to Run

### Prerequisites
- Docker and Docker Compose
- Go 1.24+ (for local development)

### Start the Service
```bash
./run_dev.sh
```
The server will be available at `http://localhost:8080`.

## How to Test

### 1. Automated Verification
The `verify.sh` script runs the entire pipeline: unit tests, race detection, environment setup, and client integration tests.
```bash
./verify.sh
```

### 2. Manual Integration Test
```bash
./test_client.sh
```

### 3. Unit & Integration Tests
```bash
# Standard tests
go test ./...

# Race condition detection
go test -race ./...
```

### Test Result

```bash
[+] up 3/3
 ✔ Image sayakaya-app       Built                                                                                                                                             9.5s
 ✔ Container sayakaya-db-1  Running                                                                                                                                           0.0s
 ✔ Container sayakaya-app-1 Started                                                                                                                                           0.3s
--- 1. Creating a Voucher (limit: 5 uses) ---
{"id":"745a3b68-7a48-438c-a02b-5884c2bbca3b","code":"RACE_TEST_1777116214","discount_type":"flat","discount_value":10000,"max_uses":5,"total_claims":0,"expires_at":"2026-12-31T23:59:59Z","min_transaction_amount":50000,"created_at":"2026-04-25T11:23:34.182614Z","updated_at":"2026-04-25T11:23:34.182614Z"}
[PASS] Voucher creation code
--- 2. Getting Voucher Details ---
{"code":"RACE_TEST_1777116214","discount_type":"flat","discount_value":10000,"max_uses":5,"remaining_claims":5,"expires_at":"2026-12-31T23:59:59Z","min_transaction_amount":50000}
[PASS] Get voucher remaining_claims
--- 3. Single Claim Test ---
{"id":"161151ff-71a4-456a-a296-0e049f676387","voucher_id":"745a3b68-7a48-438c-a02b-5884c2bbca3b","user_id":"1b95a702-6d9f-4f06-8a68-b44d7f90cbd6","status":"CLAIMED","created_at":"2026-04-25T11:23:34.220966Z","updated_at":"2026-04-25T11:23:34.220966Z"}
[PASS] Claim user_id
--- 4. Single Redeem Test ---
{"original_amount":100000,"discount_applied":10000,"final_amount":90000}
[PASS] Redeem final_amount
--- 5. Concurrency Test with 'ab' ---
Concurrency test finished.
--- 6. Verifying Final State ---
{"code":"RACE_TEST_1777116214","discount_type":"flat","discount_value":10000,"max_uses":5,"remaining_claims":3,"expires_at":"2026-12-31T23:59:59Z","min_transaction_amount":50000}
[PASS] Voucher remaining_claims
--- 7. Exhaustion Test (Concurrent unique users) ---
Exhaustion burst finished.
--- 8. Verifying Limit is Enforced ---
{"code":"RACE_TEST_1777116214","discount_type":"flat","discount_value":10000,"max_uses":5,"remaining_claims":0,"expires_at":"2026-12-31T23:59:59Z","min_transaction_amount":50000}
[PASS] Voucher remaining_claims
All tests passed successfully!
WARN[0000] /Users/alfiankan/sayakaya/docker-compose.yml: the attribute `version` is obsolete, it will be ignored, please remove it to avoid potential confusion
[+] down 4/4
 ✔ Container sayakaya-app-1      Removed                                                                                                                                      0.1s
 ✔ Container sayakaya-db-1       Removed                                                                                                                                      0.4s
 ✔ Volume sayakaya_postgres_data Removed                                                                                                                                      0.0s
 ✔ Network sayakaya_default      Removed                                                                                                                                      0.1s
Verification complete!

~/sayakaya on master +2 ?2                                                                                                                                               took 17s
> go test ./... -v
?   	sayakaya/cmd/server	[no test files]
?   	sayakaya/pkg/config	[no test files]
?   	sayakaya/pkg/logger	[no test files]
?   	sayakaya/pkg/middleware	[no test files]
?   	sayakaya/pkg/ratelimiter	[no test files]
?   	sayakaya/pkg/voucher	[no test files]
?   	sayakaya/pkg/voucher/entities	[no test files]
=== RUN   TestTokenBucket
--- PASS: TestTokenBucket (0.15s)
=== RUN   TestManager
--- PASS: TestManager (0.00s)
PASS
ok  	sayakaya/test/ratelimiter	(cached)
2026/04/25 18:24:01 Connecting to database on url:  postgres://postgres:password@localhost:32779/promoengine_test?sslmode=disable
=== RUN   TestVoucherHandler_Claim
=== RUN   TestVoucherHandler_Claim/Success
--- PASS: TestVoucherHandler_Claim (0.02s)
    --- PASS: TestVoucherHandler_Claim/Success (0.01s)
=== RUN   TestVoucherHandler_FullIntegration
=== RUN   TestVoucherHandler_FullIntegration/Full_Lifecycle
--- PASS: TestVoucherHandler_FullIntegration (0.02s)
    --- PASS: TestVoucherHandler_FullIntegration/Full_Lifecycle (0.01s)
=== RUN   TestVoucherService_RedeemVoucher
=== RUN   TestVoucherService_RedeemVoucher/Success_Percent_Discount
=== RUN   TestVoucherService_RedeemVoucher/Min_Amount_Not_Met
--- PASS: TestVoucherService_RedeemVoucher (0.03s)
    --- PASS: TestVoucherService_RedeemVoucher/Success_Percent_Discount (0.02s)
    --- PASS: TestVoucherService_RedeemVoucher/Min_Amount_Not_Met (0.01s)
=== RUN   TestVoucherService_ClaimConcurrency_Integration
--- PASS: TestVoucherService_ClaimConcurrency_Integration (0.04s)
=== RUN   TestVoucherService_SameUserConcurrency_Integration
--- PASS: TestVoucherService_SameUserConcurrency_Integration (0.03s)
PASS
ok  	sayakaya/test/voucher	5.920s

~/sayakaya on master ?2                                                                                                                                                   took 8s
> go test -race ./... -v
?   	sayakaya/cmd/server	[no test files]
?   	sayakaya/pkg/config	[no test files]
?   	sayakaya/pkg/logger	[no test files]
?   	sayakaya/pkg/middleware	[no test files]
?   	sayakaya/pkg/ratelimiter	[no test files]
?   	sayakaya/pkg/voucher	[no test files]
?   	sayakaya/pkg/voucher/entities	[no test files]
=== RUN   TestTokenBucket
--- PASS: TestTokenBucket (0.15s)
=== RUN   TestManager
--- PASS: TestManager (0.00s)
PASS
ok  	sayakaya/test/ratelimiter	(cached)
2026/04/25 18:24:14 Connecting to database on url:  postgres://postgres:password@localhost:32780/promoengine_test?sslmode=disable
=== RUN   TestVoucherHandler_Claim
=== RUN   TestVoucherHandler_Claim/Success
--- PASS: TestVoucherHandler_Claim (0.02s)
    --- PASS: TestVoucherHandler_Claim/Success (0.02s)
=== RUN   TestVoucherHandler_FullIntegration
=== RUN   TestVoucherHandler_FullIntegration/Full_Lifecycle
--- PASS: TestVoucherHandler_FullIntegration (0.03s)
    --- PASS: TestVoucherHandler_FullIntegration/Full_Lifecycle (0.02s)
=== RUN   TestVoucherService_RedeemVoucher
=== RUN   TestVoucherService_RedeemVoucher/Success_Percent_Discount
=== RUN   TestVoucherService_RedeemVoucher/Min_Amount_Not_Met
--- PASS: TestVoucherService_RedeemVoucher (0.05s)
    --- PASS: TestVoucherService_RedeemVoucher/Success_Percent_Discount (0.02s)
    --- PASS: TestVoucherService_RedeemVoucher/Min_Amount_Not_Met (0.02s)
=== RUN   TestVoucherService_ClaimConcurrency_Integration
--- PASS: TestVoucherService_ClaimConcurrency_Integration (0.10s)
=== RUN   TestVoucherService_SameUserConcurrency_Integration
--- PASS: TestVoucherService_SameUserConcurrency_Integration (0.04s)
PASS
ok  	sayakaya/test/voucher	5.600s
```
