// Package invitation provides shared token management for Prism v0.5.13+
//
// This file implements the shared token system for workshop and conference scenarios
// where a single token code can be redeemed by multiple users up to a limit.
package invitation

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/scttfrdmn/prism/pkg/types"
)

// SharedTokenManager manages shared invitation tokens
type SharedTokenManager struct {
	tokens map[string]*types.SharedInvitationToken // token code -> token
	mutex  sync.RWMutex
}

// NewSharedTokenManager creates a new shared token manager
func NewSharedTokenManager() *SharedTokenManager {
	return &SharedTokenManager{
		tokens: make(map[string]*types.SharedInvitationToken),
	}
}

// CreateSharedToken creates a new shared invitation token
func (m *SharedTokenManager) CreateSharedToken(ctx context.Context, req *types.CreateSharedTokenRequest) (*types.SharedInvitationToken, error) {
	// Validate request
	if req.ProjectID == "" {
		return nil, types.ErrInvalidProjectID
	}
	if req.Name == "" {
		return nil, fmt.Errorf("token name is required")
	}
	if req.Role == "" {
		return nil, types.ErrInvalidRole
	}
	if req.RedemptionLimit <= 0 {
		return nil, fmt.Errorf("redemption limit must be positive")
	}

	// Generate token ID
	tokenID := generateTokenID()

	// Generate human-readable token code
	tokenCode := generateTokenCode(req.Name)

	// Parse expiration
	var expiresAt *time.Time
	if req.ExpiresAt != nil {
		expiresAt = req.ExpiresAt
	} else if req.ExpiresIn != "" {
		duration, err := parseDuration(req.ExpiresIn)
		if err != nil {
			return nil, fmt.Errorf("invalid expires_in duration: %w", err)
		}
		exp := time.Now().Add(duration)
		expiresAt = &exp
	}

	// Create token
	token := &types.SharedInvitationToken{
		ID:              tokenID,
		ProjectID:       req.ProjectID,
		Token:           tokenCode,
		Name:            req.Name,
		Role:            req.Role,
		Message:         req.Message,
		CreatedBy:       req.CreatedBy,
		CreatedAt:       time.Now(),
		ExpiresAt:       expiresAt,
		RedemptionLimit: req.RedemptionLimit,
		RedemptionCount: 0,
		Status:          types.SharedTokenActive,
		Redemptions:     []types.SharedTokenRedemption{},
	}

	// Validate
	if err := token.Validate(); err != nil {
		return nil, err
	}

	// Store token
	m.mutex.Lock()
	m.tokens[tokenCode] = token
	m.mutex.Unlock()

	return token, nil
}

// RedeemToken redeems a shared token (atomic operation)
func (m *SharedTokenManager) RedeemToken(ctx context.Context, req *types.RedeemSharedTokenRequest) (*types.RedeemSharedTokenResponse, error) {
	if req.Token == "" {
		return nil, types.ErrInvalidToken
	}
	if req.RedeemedBy == "" {
		return nil, fmt.Errorf("redeemed_by is required")
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Get token
	token, exists := m.tokens[req.Token]
	if !exists {
		return nil, types.ErrSharedTokenNotFound
	}

	// Update status
	token.UpdateStatus()

	// Check if token can be redeemed
	if !token.CanRedeem() {
		if token.Status == types.SharedTokenExpired {
			return nil, types.ErrSharedTokenExpired
		}
		if token.Status == types.SharedTokenRevoked {
			return nil, types.ErrSharedTokenRevoked
		}
		if token.Status == types.SharedTokenExhausted {
			return nil, types.ErrSharedTokenExhausted
		}
		return nil, types.ErrInvalidRedemption
	}

	// Check if user has already redeemed this token
	for _, redemption := range token.Redemptions {
		if strings.EqualFold(redemption.RedeemedBy, req.RedeemedBy) {
			return nil, types.ErrAlreadyRedeemed
		}
	}

	// Increment redemption count atomically
	token.RedemptionCount++
	position := token.RedemptionCount

	// Record redemption
	redemption := types.SharedTokenRedemption{
		RedeemedBy: req.RedeemedBy,
		RedeemedAt: time.Now(),
		Position:   position,
	}
	token.Redemptions = append(token.Redemptions, redemption)

	// Update status (may become exhausted)
	token.UpdateStatus()

	// Build response
	response := &types.RedeemSharedTokenResponse{
		Success:              true,
		ProjectID:            token.ProjectID,
		Role:                 token.Role,
		RedemptionPosition:   position,
		RemainingRedemptions: token.RemainingRedemptions(),
		Message: fmt.Sprintf("Successfully redeemed token '%s'. You are participant %d of %d.",
			token.Name, position, token.RedemptionLimit),
	}

	return response, nil
}

// GetSharedToken retrieves a shared token by code
func (m *SharedTokenManager) GetSharedToken(ctx context.Context, tokenCode string) (*types.SharedInvitationToken, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	token, exists := m.tokens[tokenCode]
	if !exists {
		return nil, types.ErrSharedTokenNotFound
	}

	// Update status before returning
	token.UpdateStatus()

	return token, nil
}

// ListSharedTokens lists all shared tokens for a project
func (m *SharedTokenManager) ListSharedTokens(ctx context.Context, projectID string) ([]*types.SharedInvitationToken, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var tokens []*types.SharedInvitationToken
	for _, token := range m.tokens {
		if token.ProjectID == projectID {
			// Update status before returning
			token.UpdateStatus()
			tokens = append(tokens, token)
		}
	}

	return tokens, nil
}

// RevokeSharedToken revokes a shared token
func (m *SharedTokenManager) RevokeSharedToken(ctx context.Context, tokenCode string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	token, exists := m.tokens[tokenCode]
	if !exists {
		return types.ErrSharedTokenNotFound
	}

	token.Status = types.SharedTokenRevoked
	return nil
}

// ExtendExpiration extends the expiration of a shared token
func (m *SharedTokenManager) ExtendExpiration(ctx context.Context, tokenCode string, duration time.Duration) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	token, exists := m.tokens[tokenCode]
	if !exists {
		return types.ErrSharedTokenNotFound
	}

	// Cannot extend revoked tokens
	if token.Status == types.SharedTokenRevoked {
		return types.ErrSharedTokenRevoked
	}

	// Extend expiration
	if token.ExpiresAt == nil {
		// If no expiration set, set it to now + duration
		exp := time.Now().Add(duration)
		token.ExpiresAt = &exp
	} else {
		// Extend existing expiration
		exp := token.ExpiresAt.Add(duration)
		token.ExpiresAt = &exp
	}

	// Update status (may become active again)
	token.UpdateStatus()

	return nil
}

// generateTokenID generates a unique token ID
func generateTokenID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// generateTokenCode generates a human-readable token code
// Example: "NeurIPS Workshop" -> "WORKSHOP-NEURIPS-2025"
func generateTokenCode(name string) string {
	// Extract key words from name
	words := strings.Fields(name)

	// Take up to 3 significant words
	var codeWords []string
	for _, word := range words {
		// Skip common words
		lower := strings.ToLower(word)
		if lower == "the" || lower == "a" || lower == "an" {
			continue
		}
		// Convert to uppercase
		codeWords = append(codeWords, strings.ToUpper(word))
		if len(codeWords) >= 3 {
			break
		}
	}

	// Join with dashes
	var code string
	if len(codeWords) > 0 {
		code = strings.Join(codeWords, "-")
	} else {
		code = "TOKEN"
	}

	// Add year
	year := time.Now().Year()
	code = fmt.Sprintf("%s-%d", code, year)

	// Add random suffix for uniqueness (4 chars)
	b := make([]byte, 2)
	rand.Read(b)
	suffix := strings.ToUpper(hex.EncodeToString(b))

	return fmt.Sprintf("%s-%s", code, suffix)
}
