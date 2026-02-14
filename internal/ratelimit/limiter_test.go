package ratelimit

import (
	"testing"
	"time"
)

func TestLimiter_Lockout(t *testing.T) {
	maxAttempts := 3
	lockoutDuration := 100 * time.Millisecond
	limiter := NewLimiter(maxAttempts, lockoutDuration)
	key := "user1"

	// 1. Attempts up to max should succeed (Allow returns true, error nil)
	// Wait, my implementation returns true until attempts > maxAttempts.
	// Allow increments first.
	// 1st call: attempts=1. 1 <= 3. Returns true.
	// 2nd call: attempts=2. 2 <= 3. Returns true.
	// 3rd call: attempts=3. 3 <= 3. Returns true.
	// 4th call: attempts=4. 4 > 3. Sets lockout. Returns false.

	for i := 0; i < maxAttempts; i++ {
		allowed, err := limiter.Allow(key)
		if !allowed || err != nil {
			t.Fatalf("attempt %d should be allowed, got allowed=%v, err=%v", i+1, allowed, err)
		}
	}

	// 2. Next attempt should trigger lockout
	allowed, err := limiter.Allow(key)
	if allowed {
		t.Fatal("expected lockout, got allowed=true")
	}
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// 3. Waiting for expiration
	time.Sleep(lockoutDuration + 10*time.Millisecond)

	// 4. Should be allowed again
	allowed, err = limiter.Allow(key)
	if !allowed || err != nil {
		t.Fatalf("should be allowed after lockout expiration, got allowed=%v, err=%v", allowed, err)
	}
}

func TestLimiter_Reset(t *testing.T) {
	limiter := NewLimiter(3, time.Minute)
	key := "user2"

	// Make some failed attempts
	limiter.Allow(key)
	limiter.Allow(key)

	// Reset
	limiter.Reset(key)

	// Attempts should start from 0 (1 after call)
	// If reset didn't work, this would be attempt 3, and next would lock out.
	// Let's verify we can do 3 more.
	for i := 0; i < 3; i++ {
		allowed, err := limiter.Allow(key)
		if !allowed {
			t.Fatalf("attempt %d after reset should be allowed", i+1)
		}
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
}
