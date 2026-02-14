package voucher_test

import (
	"strings"
	"testing"
	"time"

	"voucher-system/internal/database"
	"voucher-system/internal/ratelimit"
	"voucher-system/internal/voucher"

	_ "github.com/mattn/go-sqlite3"
)

func TestAbuseProtection(t *testing.T) {
	// Setup DB
	db, err := database.InitDB(":memory:")
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer db.Close()
	if err := database.RunMigrations(db); err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}

	repo := voucher.NewSQLiteRepository(db)

	// Create a limiter with strict rules: 3 attempts, 1 second lockout
	limiter := ratelimit.NewLimiter(3, 1*time.Second)
	service := voucher.NewService(repo, limiter)

	user := "attacker"

	// 1. Exhaust attempts with invalid PINs
	for i := 0; i < 3; i++ {
		_, err := service.RedeemPIN("invalid-pin", user)
		if err == nil {
			t.Fatalf("Expected error for invalid PIN, got nil")
		}
	}

	// 2. 4th attempt should be blocked by rate limiter
	_, err = service.RedeemPIN("invalid-pin", user)
	if err == nil {
		t.Fatal("Expected error for rate limit, got nil")
	}
	if !strings.Contains(err.Error(), "too many attempts") {
		t.Errorf("Expected 'too many attempts' error, got: %v", err)
	}

	// 3. Wait for lockout to expire
	time.Sleep(1100 * time.Millisecond)

	// 4. Should be allowed again (but fail due to invalid PIN)
	// We just want to check that it's NOT a rate limit error
	_, err = service.RedeemPIN("invalid-pin", user)
	if err == nil {
		t.Fatal("Expected error for invalid PIN, got nil")
	}
	if strings.Contains(err.Error(), "too many attempts") {
		t.Errorf("Expected invalid PIN error, got rate limit error: %v", err)
	}
}
