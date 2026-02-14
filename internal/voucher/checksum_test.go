package voucher_test

import (
	"strings"
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

	// 1. Generate a valid PIN (Luhn compliant payload + valid signature)
	validPIN, err := crypto.GenerateSecurePIN(16)
	if err != nil {
		t.Fatalf("GenerateSecurePIN failed: %v", err)
	}
	
	// Valid PIN structure: [Payload (12) + Signature (4)]
	if len(validPIN) != 16 {
		t.Fatalf("Expected length 16, got %d", len(validPIN))
	}
	payload := validPIN[:12]
	
	if !crypto.ValidateLuhn(payload) {
		t.Fatalf("Generated PIN payload %s failed Luhn validation", payload)
	}
	if !crypto.ValidateSignature(validPIN, crypto.SecretKey) {
		t.Fatalf("Generated PIN %s failed signature validation", validPIN)
	}

	// 2. Test Invalid Signature
	// Modify last digit (part of signature)
	runes := []rune(validPIN)
	runes[len(runes)-1] = 'X' // Make it invalid hex/digit or just different
	if runes[len(runes)-1] == '0' {
		runes[len(runes)-1] = '1'
	} else {
		runes[len(runes)-1] = '0'
	}
	invalidSigPIN := string(runes)

	_, err = service.RedeemPIN(invalidSigPIN, user)
	if err == nil {
		t.Fatal("Expected error for invalid signature, got nil")
	}
	if !strings.Contains(err.Error(), "invalid PIN signature") {
		t.Errorf("Expected 'invalid PIN signature' error, got: %v", err)
	}

	// 3. Test Invalid Luhn (but Valid Signature)
	// We need to construct a PIN with:
	// - Payload that fails Luhn
	// - Valid Signature for that *bad* payload
	// Because RedeemPIN checks Signature FIRST. If signature is wrong, we get "invalid PIN signature".
	// We want to reach "invalid PIN format (luhn)".
	
	badPayload := "123456789012" // 12 digits. Let's make sure it fails Luhn.
	if crypto.ValidateLuhn(badPayload) {
		// If it happens to be valid, change last digit
		badPayload = "123456789013"
	}
	
	// Sign this bad payload
	sig, err := crypto.SignPIN(badPayload, crypto.SecretKey)
	if err != nil {
		t.Fatalf("SignPIN failed: %v", err)
	}
	
	pinWithBadLuhn := badPayload + sig
	
	_, err = service.RedeemPIN(pinWithBadLuhn, user)
	if err == nil {
		t.Fatal("Expected error for invalid Luhn, got nil")
	}
	if !strings.Contains(err.Error(), "invalid PIN format") {
		t.Errorf("Expected 'invalid PIN format' error, got: %v", err)
	}
}
