package crypto

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// GenerateSecurePIN generates a cryptographically secure numeric PIN of the specified length.
// The length is be between 12 and 16 digits.
func GenerateSecurePIN(length int) (string, error) {
	if length < 12 || length > 16 {
		return "", fmt.Errorf("invalid PIN length: must be between 12 and 16")
	}

	const digits = "0123456789"
	pin := make([]byte, length)
	max := big.NewInt(int64(len(digits)))

	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", fmt.Errorf("failed to generate random number: %w", err)
		}
		pin[i] = digits[num.Int64()]
	}

	return string(pin), nil
}
