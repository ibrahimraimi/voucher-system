# System Architecture

The Secure Voucher System is designed with a layered architecture to separate concerns and ensure maintainability.

## component Overview

### 1. Entry Point (`cmd/voucher-system`)
- **`main.go`**: Initializes the database, services, and starts the metrics dashboard. Acts as the CLI entry point.

### 2. Service Layer (`internal/voucher`)
- **`Service`**: Orchestrates business logic.
  - Handles PIN generation (calls `crypto`).
  - Validation (Rate Limit -> Signature -> Checksum).
  - Persistence (calls `Repository`).
- **`Repository`**: Handles database interactions.
  - Abstracts SQL queries.
  - Manages transactions for atomicity.

### 3. Core Logic (`internal`)
- **`crypto`**: Contains purely functional cryptographic primitives.
  - `GenerateSecurePIN`: Random + Luhn + HMAC.
  - `HashPIN`: SHA-256 hashing.
  - `SignPIN` / `ValidateSignature`: HMAC-SHA256 operations.
  - `Luhn`: Checksum calculation.
- **`ratelimit`**: In-memory concurrency-safe rate limiter.
- **`observability`**: Atomic counters for system metrics.

### 4. Storage Layer (`internal/database`)
- **SQLite**: Used as the backing store.
- **Schema**:
  - `vouchers` table:
    - `id` (UUID)
    - `code_hash` (Unique, Indexed)
    - `value` (Integer)
    - `status` (`unused` | `used` | `blocked`)
    - `batch_id` (UUID)
    - `created_at`, `redeemed_at`, `redeemed_by`

## Data Flow

### Redemption Flow
1. **User Request**: `RedeemPIN(pin, user_id)`
2. **Rate Limit Check**: Check internal counters for `user_id`. Fail via exponential backoff if limit exceeded.
3. **Signature Validation**: Verify last 4 digits match HMAC of payload. Fail fast if invalid.
4. **Format Validation**: Verify Luhn checksum of payload. Fail fast if invalid.
5. **Hashing**: Compute SHA-256 hash of the valid PIN.
6. **DB Transaction**:
   - `SELECT status FROM vouchers WHERE code_hash = ?`
   - If `status == unused`: `UPDATE vouchers SET status='used' ...`
   - Commit transaction.
7. **Metrics**: Increment success/failure counters.
