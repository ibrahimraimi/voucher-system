package crypto

import (
	"regexp"
	"testing"
)

func TestGenerateSecurePIN_Length(t *testing.T) {
	tests := []struct {
		length    int
		expectErr bool
	}{
		{12, false},
		{16, false},
		{10, true},
		{20, true},
	}

	for _, tt := range tests {
		pin, err := GenerateSecurePIN(tt.length)
		if tt.expectErr {
			if err == nil {
				t.Errorf("expected error for length %d, got nil", tt.length)
			}
		} else {
			if err != nil {
				t.Errorf("unexpected error for length %d: %v", tt.length, err)
			}
			if len(pin) != tt.length {
				t.Errorf("expected length %d, got %d", tt.length, len(pin))
			}
		}
	}
}

func TestGenerateSecurePIN_DigitsOnly(t *testing.T) {
	pin, err := GenerateSecurePIN(16)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	match, _ := regexp.MatchString("^[0-9]+$", pin)
	if !match {
		t.Errorf("expected only digits, got: %s", pin)
	}
}

func TestGenerateSecurePIN_Uniqueness(t *testing.T) {
	// Generate a small batch to check for immediate duplicates
	// Note: True randomness can produce duplicates, but with 12-16 digits it's extremely unlikely
	count := 1000
	seen := make(map[string]bool)

	for i := 0; i < count; i++ {
		pin, err := GenerateSecurePIN(12)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if seen[pin] {
			t.Fatalf("duplicate PIN generated: %s", pin)
		}
		seen[pin] = true
	}
}
