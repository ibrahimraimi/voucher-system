# API Reference

This document describes the core methods exposed by the `VoucherService`.

## Core Service

The service is instantiated with a repository interface and a rate limiter.

```go
service := voucher.NewService(repo, limiter)
```

### Methods

#### `CreateBatch`

Generates a batch of new vouchers.

```go
func (s *Service) CreateBatch(value int, quantity int) ([]string, string, error)
```

**Parameters:**
- `value`: Integer value of each voucher (e.g., 500 = $5.00).
- `quantity`: Number of vouchers to generate.

**Returns:**
- `[]string`: List of generated plaintext PINs. **Note**: These are only available at creation time. They are not stored.
- `string`: UUID of the batch.
- `error`: If generation or persistence fails.

#### `RedeemPIN`

Redeems a voucher by its PIN.

```go
func (s *Service) RedeemPIN(pin string, user string) (int, error)
```

**Parameters:**
- `pin`: The 16-digit PIN string provided by the user.
- `user`: Identifier of the redeeming user (for rate limiting and audit logs).

**Returns:**
- `int`: The value of the redeemed voucher.
- `error`: If redemption fails. Common errors:
  - `rate_limit_exceeded`: Too many attempts.
  - `invalid PIN signature`: HMAC verification failed.
  - `invalid PIN format`: Luhn checksum failed.
  - `voucher is used/blocked`: Already redeemed or blocked.
  - `redemption_failed`: Database error.
