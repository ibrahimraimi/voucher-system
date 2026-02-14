package voucher

import (
	"database/sql"
	"time"
)

type VoucherStatus string

const (
	StatusUnused  VoucherStatus = "unused"
	StatusUsed    VoucherStatus = "used"
	StatusBlocked VoucherStatus = "blocked"
)

type Voucher struct {
	ID         string
	CodeHash   string
	Value      int
	Status     VoucherStatus
	BatchID    string
	CreatedAt  time.Time
	RedeemedAt sql.NullTime
	RedeemedBy sql.NullString
}
