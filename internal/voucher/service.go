package voucher

import (
	"fmt"
	"time"
	"voucher-system/internal/crypto"
	"voucher-system/internal/ratelimit"

	"github.com/google/uuid"
)

type Service struct {
	repo    Repository
	limiter *ratelimit.Limiter
}

func NewService(repo Repository, limiter *ratelimit.Limiter) *Service {
	return &Service{
		repo:    repo,
		limiter: limiter,
	}
}

// CreateBatch generates a batch of unique vouchers.
// It returns the list of plaintext PINs (for export) and the batch ID.
func (s *Service) CreateBatch(value int, quantity int) ([]string, string, error) {
	batchID := uuid.New().String()
	var vouchers []*Voucher
	var pinList []string

	// Use a map to track uniqueness within the batch generation itself
	// just in case we generate duplicates in the same run (highly unlikely but good for correctness)
	generatedInBatch := make(map[string]bool)

	for i := 0; i < quantity; i++ {
		var pin string
		var hash string
		var err error

		for {
			// 1. Generate PIN
			pin, err = crypto.GenerateSecurePIN(12) // Standard 12-digit PIN
			if err != nil {
				return nil, "", fmt.Errorf("failed to generate PIN: %w", err)
			}

			// Check if we already generated this in current batch
			if generatedInBatch[pin] {
				continue
			}

			// 2. Hash PIN
			hash = crypto.HashPIN(pin)

			// 3. Check Uniqueness against DB
			existing, err := s.repo.FindByHash(hash)
			if err != nil {
				return nil, "", fmt.Errorf("failed to check hash uniqueness: %w", err)
			}
			if existing == nil {
				// Unique
				generatedInBatch[pin] = true
				break
			}
			// Collision detected, regenerate
		}

		vouchers = append(vouchers, &Voucher{
			ID:        uuid.New().String(),
			CodeHash:  hash,
			Value:     value,
			Status:    StatusUnused,
			BatchID:   batchID,
			CreatedAt: time.Now(),
		})
		pinList = append(pinList, pin)
	}

	// 4. Save Batch
	if err := s.repo.SaveBatch(vouchers); err != nil {
		return nil, "", fmt.Errorf("failed to save batch: %w", err)
	}

	return pinList, batchID, nil
}

// RedeemPIN attempts to redeem a voucher by its PIN.
// It verifies the PIN, checks its status, and marks it as used atomically.
// Returns the voucher value if successful.
func (s *Service) RedeemPIN(pin string, user string) (int, error) {
	// 1. Check Rate Limit
	allow, err := s.limiter.Allow(user)
	if !allow {
		return 0, err
	}

	// 2. Validate Checksum
	if !crypto.ValidateLuhn(pin) {
		return 0, fmt.Errorf("invalid PIN format")
	}

	hash := crypto.HashPIN(pin)
	value, err := s.repo.Redeem(hash, user)
	if err != nil {
		return 0, err
	}

	// 2. Reset Rate Limit on Success
	s.limiter.Reset(user)

	return value, nil
}
