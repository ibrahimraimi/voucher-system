package crypto

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/big"
	"strconv"
)

// SecretKey is used for HMAC signing. In production, this should be loaded from environment variables.
var SecretKey = []byte("my-secret-key-12345")

// GenerateSecurePIN generates a cryptographically secure numeric PIN of the specified length.
// The PIN includes a Luhn checksum and an HMAC signature.
// For a requested length of N:
// - The last 4 digits are the HMAC signature.
// - The remaining (N-4) digits are the payload (Random + Luhn).
func GenerateSecurePIN(length int) (string, error) {
	if length < 12 || length > 16 {
		return "", fmt.Errorf("invalid PIN length: must be between 12 and 16")
	}

	// 1. Generate Payload (Random + Luhn)
	// Payload length = Total Length - Signature Length (4)
	payloadLength := length - 4

	// Within payload, last digit is Luhn. So random digits = payloadLength - 1
	randomLength := payloadLength - 1

	const digits = "0123456789"
	pinBytes := make([]byte, randomLength)
	max := big.NewInt(int64(len(digits)))

	for i := 0; i < randomLength; i++ {
		num, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", fmt.Errorf("failed to generate random number: %w", err)
		}
		pinBytes[i] = digits[num.Int64()]
	}

	randomPart := string(pinBytes)
	checkDigit, err := CalculateLuhnDigit(randomPart)
	if err != nil {
		return "", fmt.Errorf("failed to calculate checksum: %w", err)
	}

	payload := randomPart + strconv.Itoa(checkDigit)

	// 2. Sign Payload
	signature, err := SignPIN(payload, SecretKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign PIN: %w", err)
	}

	return payload + signature, nil
}

// SignPIN computes a truncated HMAC-SHA256 signature of the payload.
// Returns a 4-digit numeric string.
func SignPIN(payload string, secret []byte) (string, error) {
	h := hmac.New(sha256.New, secret)
	h.Write([]byte(payload))
	mac := h.Sum(nil)

	// Truncate to 4 digits.
	// We can treat the first 4 bytes as a uint32 and mod 10000.
	// Or just use the last bytes.
	// RFC 4226 (HOTP) style dynamic truncation is better but a simple slice is fine for this scope.
	offset := mac[len(mac)-1] & 0x0f
	code := binary.BigEndian.Uint32(mac[offset : offset+4])
	code = code & 0x7fffffff
	code = code % 10000

	return fmt.Sprintf("%04d", code), nil
}

// ValidateSignature checks if the PIN has a valid HMAC signature.
// It assumes the last 4 digits are the signature.
func ValidateSignature(fullPIN string, secret []byte) bool {
	if len(fullPIN) < 5 {
		return false
	}
	payload := fullPIN[:len(fullPIN)-4]
	signature := fullPIN[len(fullPIN)-4:]

	expectedSignature, err := SignPIN(payload, secret)
	if err != nil {
		return false
	}

	return signature == expectedSignature
}
