package crypto

import "testing"

func TestCalculateLuhnDigit(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"7992739871", 3},
		{"123456", 6}, // 123456 -> 2*6=12->3, 5, 2*4=8, 3, 2*2=4, 1. Sum: 3+5+8+3+4+1 = 24. 10 - 4 = 6.
		// Wait, let's verify manually or use a standard example.
		// Wikipedia example: 7992739871.
		// 1*2=2, 7, 8*2=16->7, 9, 3*2=6, 7, 2*2=4, 9, 9*2=18->9, 7.
		// Sum: 2+7+7+9+6+7+4+9+9+7 = 67. 67 % 10 = 7. 10 - 7 = 3. Correct.
	}

	for _, tt := range tests {
		digit, err := CalculateLuhnDigit(tt.input)
		if err != nil {
			t.Errorf("CalculateLuhnDigit(%s) returned error: %v", tt.input, err)
		}
		if digit != tt.expected {
			t.Errorf("CalculateLuhnDigit(%s) = %d; want %d", tt.input, digit, tt.expected)
		}
	}
}

func TestValidateLuhn(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"79927398713", true},
		{"79927398710", false},
		{"1234566", true}, // 123456 with check digit 6
		{"1234567", false},
		{"invalid", false},
	}

	for _, tt := range tests {
		valid := ValidateLuhn(tt.input)
		if valid != tt.valid {
			t.Errorf("ValidateLuhn(%s) = %v; want %v", tt.input, valid, tt.valid)
		}
	}
}
