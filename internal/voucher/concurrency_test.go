package voucher_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"voucher-system/internal/database"
	"voucher-system/internal/observability"
	"voucher-system/internal/ratelimit"
	"voucher-system/internal/voucher"

	_ "github.com/mattn/go-sqlite3"
)

func TestConcurrentRedemption(t *testing.T) {
	// Setup DB
	db, err := database.InitDB(":memory:")
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer db.Close()
	if err := database.RunMigrations(db); err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}
	db.SetMaxOpenConns(1)

	repo := voucher.NewSQLiteRepository(db)
	// Use a limiter with high capacity to not block valid concurrent attempts from separate users
	limiter := ratelimit.NewLimiter(100, time.Minute)
	service := voucher.NewService(repo, limiter)

	// 1. Create a single valid voucher
	pins, _, err := service.CreateBatch(100, 1)
	if err != nil {
		t.Fatalf("CreateBatch failed: %v", err)
	}
	validPIN := pins[0]

	// 2. Spawn N goroutines trying to redeem it
	concurrency := 10
	var wg sync.WaitGroup
	wg.Add(concurrency)

	successCount := 0
	failCount := 0
	var mu sync.Mutex

	// Reset metrics before test
	// Note: Metrics are global, so this might interfere if run in parallel with other tests.
	// But go test runs packages in parallel, not tests within a package (unless t.Parallel() is called).
	// We haven't exposed a Reset for metrics, so we just capture start values.
	startMetrics := observability.GetMetrics()

	for i := 0; i < concurrency; i++ {
		go func(id int) {
			defer wg.Done()
			user := fmt.Sprintf("user-%d", id)

			// Small delay to ensure they start roughly together?
			// Or just go.

			_, err := service.RedeemPIN(validPIN, user)

			mu.Lock()
			defer mu.Unlock()
			if err == nil {
				successCount++
			} else {
				failCount++
				// We expect "voucher is used" or concurrent update detected errors
			}
		}(i)
	}

	wg.Wait()

	// 3. Verify exactly ONE success
	if successCount != 1 {
		t.Errorf("Expected exactly 1 success, got %d", successCount)
	}
	if failCount != concurrency-1 {
		t.Errorf("Expected %d failures, got %d", concurrency-1, failCount)
	}

	// 4. Verify Metrics
	endMetrics := observability.GetMetrics()
	redeemedDiff := endMetrics.TotalRedeemed - startMetrics.TotalRedeemed
	if redeemedDiff != 1 {
		t.Errorf("Expected TotalRedeemed to increase by 1, got %d", redeemedDiff)
	}
}

func TestConcurrentGeneration(t *testing.T) {
	// Setup DB
	db, err := database.InitDB(":memory:")
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer db.Close()
	if err := database.RunMigrations(db); err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}
	db.SetMaxOpenConns(1)

	repo := voucher.NewSQLiteRepository(db)
	limiter := ratelimit.NewLimiter(100, time.Minute)
	service := voucher.NewService(repo, limiter)

	concurrency := 5
	batchSize := 20
	var wg sync.WaitGroup
	wg.Add(concurrency)

	startMetrics := observability.GetMetrics()

	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			_, _, err := service.CreateBatch(50, batchSize)
			if err != nil {
				t.Errorf("CreateBatch failed in goroutine: %v", err)
			}
		}()
	}

	wg.Wait()

	// Verify Metrics
	endMetrics := observability.GetMetrics()
	generatedDiff := endMetrics.TotalGenerated - startMetrics.TotalGenerated
	expectedGenerated := uint64(concurrency * batchSize)

	if generatedDiff != expectedGenerated {
		t.Errorf("Expected Generated to increase by %d, got %d", expectedGenerated, generatedDiff)
	}
}
