package crypto

import (
	"testing"
)

func TestSignPIN(t *testing.T) {
	payload := "12345678"
	secret := []byte("secret")

	sig, err := SignPIN(payload, secret)
	if err != nil {
		t.Fatalf("SignPIN failed: %v", err)
	}

	if len(sig) != 4 {
		t.Errorf("Expected 4 digit signature, got %d", len(sig))
	}

	// Verify consistency
	sig2, _ := SignPIN(payload, secret)
	if sig != sig2 {
		t.Errorf("SignPIN is not deterministic")
	}

	// Verify sensitivity to secret
	sig3, _ := SignPIN(payload, []byte("other-secret"))
	if sig == sig3 {
		t.Errorf("Signature should change with secret")
	}
}

func TestValidateSignature(t *testing.T) {
	payload := "12345678"
	secret := []byte("secret")
	sig, _ := SignPIN(payload, secret)
	fullPIN := payload + sig

	if !ValidateSignature(fullPIN, secret) {
		t.Errorf("ValidateSignature failed for valid PIN")
	}

	if ValidateSignature(fullPIN, []byte("wrong-secret")) {
		t.Errorf("ValidateSignature passed for wrong secret")
	}

	invalidPIN := payload + "0000"
	if ValidateSignature(invalidPIN, secret) {
		t.Errorf("ValidateSignature passed for invalid signature")
	}
}

func TestGenerateSecurePIN_Structure(t *testing.T) {
	pin, err := GenerateSecurePIN(16)
	if err != nil {
		t.Fatalf("GenerateSecurePIN failed: %v", err)
	}

	if len(pin) != 16 {
		t.Errorf("Expected length 16, got %d", len(pin))
	}

	// Verify structure: [Payload (12) + Signature (4)]
	// Payload: [Random (11) + Luhn (1)]

	payload := pin[:12]
	// 1. Verify Luhn on payload
	if !ValidateLuhn(payload) {
		t.Errorf("Payload %s part failed Luhn validation", payload)
	}

	// 2. Verify Signature
	if !ValidateSignature(pin, SecretKey) {
		t.Errorf("PIN %s failed signature validation", pin)
	}
}
