package voucher_test

import (
	"testing"
	"time"
	"voucher-system/internal/database"
	"voucher-system/internal/ratelimit"
	"voucher-system/internal/voucher"
)

func TestRedeemPIN(t *testing.T) {
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
	limiter := ratelimit.NewLimiter(3, time.Minute)
	service := voucher.NewService(repo, limiter)

	// 1. Create a batch of vouchers
	value := 500
	quantity := 5
	pins, _, err := service.CreateBatch(value, quantity)
	if err != nil {
		t.Fatalf("CreateBatch failed: %v", err)
	}

	targetPIN := pins[0]
	user := "user123"

	// 2. Redeem successfully
	redeemedValue, err := service.RedeemPIN(targetPIN, user)
	if err != nil {
		t.Fatalf("RedeemPIN failed: %v", err)
	}
	if redeemedValue != value {
		t.Errorf("Expected value %d, got %d", value, redeemedValue)
	}

	// 3. Attempt Double Spend
	_, err = service.RedeemPIN(targetPIN, "user456")
	if err == nil {
		t.Fatal("Expected error for double spend, got nil")
	}

	// 4. Attempt Invalid PIN
	_, err = service.RedeemPIN("invalid-pin", user)
	if err == nil {
		t.Fatal("Expected error for invalid PIN, got nil")
	}
}
