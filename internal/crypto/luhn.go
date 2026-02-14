package crypto

import (
	"fmt"
	"strconv"
)

// CalculateLuhnDigit calculates the checksum digit for a given numeric string using the Luhn algorithm.
func CalculateLuhnDigit(number string) (int, error) {
	sum := 0
	alt := true

	// Loop from right to left
	for i := len(number) - 1; i >= 0; i-- {
		digit, err := strconv.Atoi(string(number[i]))
		if err != nil {
			return 0, fmt.Errorf("invalid digit: %v", number[i])
		}

		if alt {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
		alt = !alt
	}

	return (10 - (sum % 10)) % 10, nil
}

// ValidateLuhn checks if the provided number string is valid according to the Luhn algorithm.
func ValidateLuhn(number string) bool {
	sum := 0
	alt := false

	for i := len(number) - 1; i >= 0; i-- {
		digit, err := strconv.Atoi(string(number[i]))
		if err != nil {
			return false
		}

		if alt {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
		alt = !alt
	}

	return sum%10 == 0
}
