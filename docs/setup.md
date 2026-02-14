# Setup and Usage

## Prerequisites
- **Go**: Version 1.21 or later.
- **GCC**: Required for `go-sqlite3` (CGO).

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/ibrahimraimi/voucher-system
   cd voucher-system
   ```

2. Download dependencies:
   ```bash
   go mod download
   ```

## Running the Application

To run the demo CLI which generates vouchers and displays metrics:

```bash
go run cmd/voucher-system/main.go
```

## Running Tests

### Unit & Integration Tests
Run standard tests:
```bash
go test -v ./...
```

### Race Detection
Run tests with race detection enabled (critical for concurrency validation):
```bash
go test -race -v ./...
```

## Configuration

Currently, configuration (like DB path, Rate Limits) is defined in `main.go`. In a production environment, these should be moved to environment variables or a config file.

- **DB Path**: Defaults to `./vouchers.db`
- **Secret Key**: Defined in `internal/crypto/pin.go`. **MUST** be changed and loaded securely in production.
