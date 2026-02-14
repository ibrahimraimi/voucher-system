package observability

import (
	"log"
	"sync/atomic"
)

type Metrics struct {
	TotalGenerated uint64
	TotalRedeemed  uint64
	FailedAttempts uint64
	// For velocity, we could use a sliding window, but for simplicity let's just track redemptions per minute
	// This simple counter resets every minute or just increases monotonically and we calculate rate on display
	// A better approach for a dashboard is probably "redemptions in last minute".
	// Let's stick to simple monotonic counters for now, and derive rate in the dashboard loop.
}

var (
	TotalGenerated uint64
	TotalRedeemed  uint64
	FailedAttempts uint64
)

func IncGenerated(count int) {
	atomic.AddUint64(&TotalGenerated, uint64(count))
	log.Printf("METRIC: Generated %d vouchers", count)
}

func IncRedeemed() {
	atomic.AddUint64(&TotalRedeemed, 1)
	log.Printf("METRIC: Voucher redeemed successfully")
}

func IncFailed(reason string) {
	atomic.AddUint64(&FailedAttempts, 1)
	log.Printf("METRIC: Redemption failed: %s", reason)
}

func GetMetrics() Metrics {
	return Metrics{
		TotalGenerated: atomic.LoadUint64(&TotalGenerated),
		TotalRedeemed:  atomic.LoadUint64(&TotalRedeemed),
		FailedAttempts: atomic.LoadUint64(&FailedAttempts),
	}
}
