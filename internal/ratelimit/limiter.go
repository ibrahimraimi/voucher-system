package ratelimit

import (
	"fmt"
	"sync"
	"time"
)

// Limiter implements a simple in-memory rate limiter with lockout.
type Limiter struct {
	mu              sync.Mutex
	attempts        map[string]int
	lockouts        map[string]time.Time
	maxAttempts     int
	lockoutDuration time.Duration
}

// NewLimiter creates a new Limiter.
func NewLimiter(maxAttempts int, lockoutDuration time.Duration) *Limiter {
	return &Limiter{
		attempts:        make(map[string]int),
		lockouts:        make(map[string]time.Time),
		maxAttempts:     maxAttempts,
		lockoutDuration: lockoutDuration,
	}
}

// Allow checks if the action is allowed for the given key.
// If allowed, it increments the attempt counter.
// If the limit is reached, it sets a lockout.
// Returns true if allowed, false if locked out.
func (l *Limiter) Allow(key string) (bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Check if partially locked out
	if lockoutTime, exists := l.lockouts[key]; exists {
		if time.Now().Before(lockoutTime) {
			remaining := time.Until(lockoutTime)
			return false, fmt.Errorf("too many attempts, try again in %v", remaining.Round(time.Second))
		}
		// Lockout expired, reset
		delete(l.lockouts, key)
		delete(l.attempts, key)
	}

	l.attempts[key]++

	if l.attempts[key] > l.maxAttempts {
		l.lockouts[key] = time.Now().Add(l.lockoutDuration)
		return false, fmt.Errorf("too many attempts, locked out for %v", l.lockoutDuration)
	}

	return true, nil
}

// Reset clears the attempts for the given key (e.g., on success).
func (l *Limiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.attempts, key)
	delete(l.lockouts, key)
}
