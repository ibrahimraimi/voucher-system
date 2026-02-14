package crypto

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
)

// HashPIN computes the SHA-256 hash of the given PIN and returns the hex-encoded string.
// This function ensures that we never store plaintext PINs.
func HashPIN(pin string) string {
	hash := sha256.Sum256([]byte(pin))
	return hex.EncodeToString(hash[:])
}

// CompareHash compares an input PIN against a stored hash using constant-time comparison.
// It returns true if the input PIN matches the stored hash, false otherwise.
// This approach prevents timing attacks.
func CompareHash(inputPIN, storedHash string) bool {
	// Re-compute the hash of the input PIN
	computedHash := HashPIN(inputPIN)

	// Use ConstantTimeCompare to check if they match.
	// ConstantTimeCompare expects byte slices, so we convert the hex strings.
	// Note: In a real-world scenario where storedHash might come from a DB and could be
	// of varying lengths (though SHA-256 is fixed), we should ensure lengths match first
	// to avoid leaking length information if that was a concern, but here both are hex SHA-256.
	return subtle.ConstantTimeCompare([]byte(computedHash), []byte(storedHash)) == 1
}
