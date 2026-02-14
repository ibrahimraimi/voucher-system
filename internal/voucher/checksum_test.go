package voucher_test

import (
	"testing"
	"time"

	"voucher-system/internal/crypto"
	"voucher-system/internal/database"
	"voucher-system/internal/ratelimit"
	"voucher-system/internal/voucher"

	_ "github.com/mattn/go-sqlite3"
)

func TestChecksumValidation(t *testing.T) {
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
	limiter := ratelimit.NewLimiter(10, time.Minute)
	service := voucher.NewService(repo, limiter)

	user := "checksum_tester"

	// 1. Generate a valid PIN (Luhn compliant)
	validPIN, err := crypto.GenerateSecurePIN(12)
	if err != nil {
		t.Fatalf("GenerateSecurePIN failed: %v", err)
	}
	if !crypto.ValidateLuhn(validPIN) {
		t.Fatalf("Generated PIN %s failed Luhn validation", validPIN)
	}

	// 2. Transpose two digits to make it invalid (likely)
	// Or just change the last digit.
	runes := []rune(validPIN)
	lastDigit := runes[len(runes)-1]
	if lastDigit == '0' {
		runes[len(runes)-1] = '1'
	} else {
		runes[len(runes)-1] = '0'
	}
	invalidPIN := string(runes)

	// Verify our invalid PIN is actually invalid (Luhn can detect single digit errors)
	if crypto.ValidateLuhn(invalidPIN) {
		t.Fatalf("Failed to create an invalid PIN from valid one. PIN: %s", invalidPIN)
	}

	// 3. Attempt redeem with invalid PIN
	_, err = service.RedeemPIN(invalidPIN, user)
	if err == nil {
		t.Fatal("Expected error for invalid checksum, got nil")
	}
	if err.Error() != "invalid PIN format" {
		t.Errorf("Expected 'invalid PIN format' error, got: %v", err)
	}
}
