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

