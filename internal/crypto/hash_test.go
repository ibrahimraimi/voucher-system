package crypto

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestHashPIN(t *testing.T) {
	pin := "123456789012"
	expectedHashBytes := sha256.Sum256([]byte(pin))
	expectedHash := hex.EncodeToString(expectedHashBytes[:])

	hash := HashPIN(pin)
	if hash != expectedHash {
		t.Errorf("HashPIN(%s) = %s; want %s", pin, hash, expectedHash)
	}

	// Verify determinism
	hash2 := HashPIN(pin)
	if hash != hash2 {
		t.Errorf("HashPIN is not deterministic")
	}
}

func TestCompareHash(t *testing.T) {
	pin := "123456789012"
	wrongPin := "123456789013"
	hash := HashPIN(pin)

	if !CompareHash(pin, hash) {
		t.Errorf("CompareHash failed for correct PIN")
	}

	if CompareHash(wrongPin, hash) {
		t.Errorf("CompareHash succeeded for wrong PIN")
	}

	// Test with explicit known values (sanity check)
	// Hash of "test" is 9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08
	testPin := "test"
	testHash := "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08"
	if !CompareHash(testPin, testHash) {
		t.Errorf("Sanity check failed: known hash does not match")
	}
}
