package invitation

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/scttfrdmn/prism/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSharedTokenManager(t *testing.T) {
	manager := NewSharedTokenManager()
	require.NotNil(t, manager)
	assert.NotNil(t, manager.tokens)
}

func TestCreateSharedToken_Success(t *testing.T) {
	manager := NewSharedTokenManager()
	ctx := context.Background()

	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	req := &types.CreateSharedTokenRequest{
		ProjectID:       "proj-123",
		Name:            "Workshop Token",
		Role:            types.ProjectRoleMember,
		Message:         "Welcome to the workshop",
		CreatedBy:       "admin@example.com",
		ExpiresAt:       &expiresAt,
		RedemptionLimit: 50,
	}

	token, err := manager.CreateSharedToken(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, token)

	assert.NotEmpty(t, token.ID)
	assert.NotEmpty(t, token.Token)
	assert.Equal(t, "proj-123", token.ProjectID)
	assert.Equal(t, "Workshop Token", token.Name)
	assert.Equal(t, types.ProjectRoleMember, token.Role)
	assert.Equal(t, "Welcome to the workshop", token.Message)
	assert.Equal(t, "admin@example.com", token.CreatedBy)
	assert.Equal(t, 50, token.RedemptionLimit)
	assert.Equal(t, 0, token.RedemptionCount)
	assert.Equal(t, types.SharedTokenActive, token.Status)
	assert.Empty(t, token.Redemptions)
}

func TestCreateSharedToken_WithExpiresIn(t *testing.T) {
	manager := NewSharedTokenManager()
	ctx := context.Background()

	req := &types.CreateSharedTokenRequest{
		ProjectID:       "proj-123",
		Name:            "Workshop Token",
		Role:            types.ProjectRoleMember,
		CreatedBy:       "admin@example.com",
		ExpiresIn:       "7d",
		RedemptionLimit: 50,
	}

	token, err := manager.CreateSharedToken(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, token.ExpiresAt)

	// Verify expiration is approximately 7 days from now
	expectedExpiration := time.Now().Add(7 * 24 * time.Hour)
	assert.WithinDuration(t, expectedExpiration, *token.ExpiresAt, 5*time.Second)
}

func TestCreateSharedToken_ValidationErrors(t *testing.T) {
	manager := NewSharedTokenManager()
	ctx := context.Background()

	tests := []struct {
		name        string
		req         *types.CreateSharedTokenRequest
		expectedErr string
	}{
		{
			name: "missing project ID",
			req: &types.CreateSharedTokenRequest{
				Name:            "Token",
				Role:            types.ProjectRoleMember,
				RedemptionLimit: 50,
			},
			expectedErr: "invalid project ID",
		},
		{
			name: "missing name",
			req: &types.CreateSharedTokenRequest{
				ProjectID:       "proj-123",
				Role:            types.ProjectRoleMember,
				RedemptionLimit: 50,
			},
			expectedErr: "token name is required",
		},
		{
			name: "missing role",
			req: &types.CreateSharedTokenRequest{
				ProjectID:       "proj-123",
				Name:            "Token",
				RedemptionLimit: 50,
			},
			expectedErr: "invalid role",
		},
		{
			name: "zero redemption limit",
			req: &types.CreateSharedTokenRequest{
				ProjectID:       "proj-123",
				Name:            "Token",
				Role:            types.ProjectRoleMember,
				RedemptionLimit: 0,
			},
			expectedErr: "redemption limit must be positive",
		},
		{
			name: "negative redemption limit",
			req: &types.CreateSharedTokenRequest{
				ProjectID:       "proj-123",
				Name:            "Token",
				Role:            types.ProjectRoleMember,
				RedemptionLimit: -1,
			},
			expectedErr: "redemption limit must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := manager.CreateSharedToken(ctx, tt.req)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestCreateSharedToken_InvalidExpiresIn(t *testing.T) {
	manager := NewSharedTokenManager()
	ctx := context.Background()

	req := &types.CreateSharedTokenRequest{
		ProjectID:       "proj-123",
		Name:            "Token",
		Role:            types.ProjectRoleMember,
		ExpiresIn:       "invalid",
		RedemptionLimit: 50,
	}

	_, err := manager.CreateSharedToken(ctx, req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid expires_in duration")
}

func TestRedeemToken_Success(t *testing.T) {
	manager := NewSharedTokenManager()
	ctx := context.Background()

	// Create token
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	createReq := &types.CreateSharedTokenRequest{
		ProjectID:       "proj-123",
		Name:            "Workshop Token",
		Role:            types.ProjectRoleMember,
		CreatedBy:       "admin@example.com",
		ExpiresAt:       &expiresAt,
		RedemptionLimit: 3,
	}

	token, err := manager.CreateSharedToken(ctx, createReq)
	require.NoError(t, err)

	// Redeem token
	redeemReq := &types.RedeemSharedTokenRequest{
		Token:      token.Token,
		RedeemedBy: "user1@example.com",
	}

	response, err := manager.RedeemToken(ctx, redeemReq)

	require.NoError(t, err)
	require.NotNil(t, response)

	assert.True(t, response.Success)
	assert.Equal(t, "proj-123", response.ProjectID)
	assert.Equal(t, types.ProjectRoleMember, response.Role)
	assert.Equal(t, 1, response.RedemptionPosition)
	assert.Equal(t, 2, response.RemainingRedemptions)
	assert.Contains(t, response.Message, "Successfully redeemed")
	assert.Contains(t, response.Message, "participant 1 of 3")

	// Verify token state
	updatedToken, err := manager.GetSharedToken(ctx, token.Token)
	require.NoError(t, err)
	assert.Equal(t, 1, updatedToken.RedemptionCount)
	assert.Len(t, updatedToken.Redemptions, 1)
	assert.Equal(t, "user1@example.com", updatedToken.Redemptions[0].RedeemedBy)
	assert.Equal(t, 1, updatedToken.Redemptions[0].Position)
}

func TestRedeemToken_MultipleRedemptions(t *testing.T) {
	manager := NewSharedTokenManager()
	ctx := context.Background()

	// Create token
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	createReq := &types.CreateSharedTokenRequest{
		ProjectID:       "proj-123",
		Name:            "Workshop Token",
		Role:            types.ProjectRoleMember,
		CreatedBy:       "admin@example.com",
		ExpiresAt:       &expiresAt,
		RedemptionLimit: 3,
	}

	token, err := manager.CreateSharedToken(ctx, createReq)
	require.NoError(t, err)

	// First redemption
	resp1, err := manager.RedeemToken(ctx, &types.RedeemSharedTokenRequest{
		Token:      token.Token,
		RedeemedBy: "user1@example.com",
	})
	require.NoError(t, err)
	assert.Equal(t, 1, resp1.RedemptionPosition)
	assert.Equal(t, 2, resp1.RemainingRedemptions)

	// Second redemption
	resp2, err := manager.RedeemToken(ctx, &types.RedeemSharedTokenRequest{
		Token:      token.Token,
		RedeemedBy: "user2@example.com",
	})
	require.NoError(t, err)
	assert.Equal(t, 2, resp2.RedemptionPosition)
	assert.Equal(t, 1, resp2.RemainingRedemptions)

	// Third redemption
	resp3, err := manager.RedeemToken(ctx, &types.RedeemSharedTokenRequest{
		Token:      token.Token,
		RedeemedBy: "user3@example.com",
	})
	require.NoError(t, err)
	assert.Equal(t, 3, resp3.RedemptionPosition)
	assert.Equal(t, 0, resp3.RemainingRedemptions)

	// Verify token is exhausted
	updatedToken, err := manager.GetSharedToken(ctx, token.Token)
	require.NoError(t, err)
	assert.Equal(t, types.SharedTokenExhausted, updatedToken.Status)
}

func TestRedeemToken_ExhaustedToken(t *testing.T) {
	manager := NewSharedTokenManager()
	ctx := context.Background()

	// Create token with limit of 1
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	createReq := &types.CreateSharedTokenRequest{
		ProjectID:       "proj-123",
		Name:            "Workshop Token",
		Role:            types.ProjectRoleMember,
		CreatedBy:       "admin@example.com",
		ExpiresAt:       &expiresAt,
		RedemptionLimit: 1,
	}

	token, err := manager.CreateSharedToken(ctx, createReq)
	require.NoError(t, err)

	// First redemption (should succeed)
	_, err = manager.RedeemToken(ctx, &types.RedeemSharedTokenRequest{
		Token:      token.Token,
		RedeemedBy: "user1@example.com",
	})
	require.NoError(t, err)

	// Second redemption (should fail - exhausted)
	_, err = manager.RedeemToken(ctx, &types.RedeemSharedTokenRequest{
		Token:      token.Token,
		RedeemedBy: "user2@example.com",
	})
	require.Error(t, err)
	assert.Equal(t, types.ErrSharedTokenExhausted, err)
}

func TestRedeemToken_AlreadyRedeemed(t *testing.T) {
	manager := NewSharedTokenManager()
	ctx := context.Background()

	// Create token
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	createReq := &types.CreateSharedTokenRequest{
		ProjectID:       "proj-123",
		Name:            "Workshop Token",
		Role:            types.ProjectRoleMember,
		CreatedBy:       "admin@example.com",
		ExpiresAt:       &expiresAt,
		RedemptionLimit: 10,
	}

	token, err := manager.CreateSharedToken(ctx, createReq)
	require.NoError(t, err)

	// First redemption
	_, err = manager.RedeemToken(ctx, &types.RedeemSharedTokenRequest{
		Token:      token.Token,
		RedeemedBy: "user1@example.com",
	})
	require.NoError(t, err)

	// Same user tries to redeem again
	_, err = manager.RedeemToken(ctx, &types.RedeemSharedTokenRequest{
		Token:      token.Token,
		RedeemedBy: "user1@example.com",
	})
	require.Error(t, err)
	assert.Equal(t, types.ErrAlreadyRedeemed, err)
}

func TestRedeemToken_AlreadyRedeemedCaseInsensitive(t *testing.T) {
	manager := NewSharedTokenManager()
	ctx := context.Background()

	// Create token
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	createReq := &types.CreateSharedTokenRequest{
		ProjectID:       "proj-123",
		Name:            "Workshop Token",
		Role:            types.ProjectRoleMember,
		CreatedBy:       "admin@example.com",
		ExpiresAt:       &expiresAt,
		RedemptionLimit: 10,
	}

	token, err := manager.CreateSharedToken(ctx, createReq)
	require.NoError(t, err)

	// First redemption
	_, err = manager.RedeemToken(ctx, &types.RedeemSharedTokenRequest{
		Token:      token.Token,
		RedeemedBy: "User1@Example.com",
	})
	require.NoError(t, err)

	// Same user with different case tries to redeem
	_, err = manager.RedeemToken(ctx, &types.RedeemSharedTokenRequest{
		Token:      token.Token,
		RedeemedBy: "user1@example.com",
	})
	require.Error(t, err)
	assert.Equal(t, types.ErrAlreadyRedeemed, err)
}

func TestRedeemToken_ExpiredToken(t *testing.T) {
	manager := NewSharedTokenManager()
	ctx := context.Background()

	// Create token that's already expired
	expiresAt := time.Now().Add(-1 * time.Hour)
	createReq := &types.CreateSharedTokenRequest{
		ProjectID:       "proj-123",
		Name:            "Workshop Token",
		Role:            types.ProjectRoleMember,
		CreatedBy:       "admin@example.com",
		ExpiresAt:       &expiresAt,
		RedemptionLimit: 10,
	}

	token, err := manager.CreateSharedToken(ctx, createReq)
	require.NoError(t, err)

	// Try to redeem expired token
	_, err = manager.RedeemToken(ctx, &types.RedeemSharedTokenRequest{
		Token:      token.Token,
		RedeemedBy: "user1@example.com",
	})
	require.Error(t, err)
	assert.Equal(t, types.ErrSharedTokenExpired, err)
}

func TestRedeemToken_RevokedToken(t *testing.T) {
	manager := NewSharedTokenManager()
	ctx := context.Background()

	// Create token
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	createReq := &types.CreateSharedTokenRequest{
		ProjectID:       "proj-123",
		Name:            "Workshop Token",
		Role:            types.ProjectRoleMember,
		CreatedBy:       "admin@example.com",
		ExpiresAt:       &expiresAt,
		RedemptionLimit: 10,
	}

	token, err := manager.CreateSharedToken(ctx, createReq)
	require.NoError(t, err)

	// Revoke token
	err = manager.RevokeSharedToken(ctx, token.Token)
	require.NoError(t, err)

	// Try to redeem revoked token
	_, err = manager.RedeemToken(ctx, &types.RedeemSharedTokenRequest{
		Token:      token.Token,
		RedeemedBy: "user1@example.com",
	})
	require.Error(t, err)
	assert.Equal(t, types.ErrSharedTokenRevoked, err)
}

func TestRedeemToken_ValidationErrors(t *testing.T) {
	manager := NewSharedTokenManager()
	ctx := context.Background()

	tests := []struct {
		name        string
		req         *types.RedeemSharedTokenRequest
		expectedErr error
	}{
		{
			name: "missing token",
			req: &types.RedeemSharedTokenRequest{
				RedeemedBy: "user@example.com",
			},
			expectedErr: types.ErrInvalidToken,
		},
		{
			name: "missing redeemed_by",
			req: &types.RedeemSharedTokenRequest{
				Token: "TOKEN-123",
			},
			expectedErr: nil, // Generic error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := manager.RedeemToken(ctx, tt.req)
			require.Error(t, err)
			if tt.expectedErr != nil {
				assert.Equal(t, tt.expectedErr, err)
			}
		})
	}
}

func TestRedeemToken_NonexistentToken(t *testing.T) {
	manager := NewSharedTokenManager()
	ctx := context.Background()

	_, err := manager.RedeemToken(ctx, &types.RedeemSharedTokenRequest{
		Token:      "NONEXISTENT-TOKEN",
		RedeemedBy: "user@example.com",
	})

	require.Error(t, err)
	assert.Equal(t, types.ErrSharedTokenNotFound, err)
}

func TestGetSharedToken_Success(t *testing.T) {
	manager := NewSharedTokenManager()
	ctx := context.Background()

	// Create token
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	createReq := &types.CreateSharedTokenRequest{
		ProjectID:       "proj-123",
		Name:            "Workshop Token",
		Role:            types.ProjectRoleMember,
		CreatedBy:       "admin@example.com",
		ExpiresAt:       &expiresAt,
		RedemptionLimit: 50,
	}

	token, err := manager.CreateSharedToken(ctx, createReq)
	require.NoError(t, err)

	// Get token
	retrieved, err := manager.GetSharedToken(ctx, token.Token)

	require.NoError(t, err)
	require.NotNil(t, retrieved)
	assert.Equal(t, token.ID, retrieved.ID)
	assert.Equal(t, token.Token, retrieved.Token)
	assert.Equal(t, token.ProjectID, retrieved.ProjectID)
}

func TestGetSharedToken_NotFound(t *testing.T) {
	manager := NewSharedTokenManager()
	ctx := context.Background()

	_, err := manager.GetSharedToken(ctx, "NONEXISTENT-TOKEN")

	require.Error(t, err)
	assert.Equal(t, types.ErrSharedTokenNotFound, err)
}

func TestListSharedTokens(t *testing.T) {
	manager := NewSharedTokenManager()
	ctx := context.Background()

	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	// Create tokens for project 1
	for i := 1; i <= 2; i++ {
		req := &types.CreateSharedTokenRequest{
			ProjectID:       "proj-1",
			Name:            "Token " + string(rune('0'+i)),
			Role:            types.ProjectRoleMember,
			CreatedBy:       "admin@example.com",
			ExpiresAt:       &expiresAt,
			RedemptionLimit: 50,
		}
		_, err := manager.CreateSharedToken(ctx, req)
		require.NoError(t, err)
	}

	// Create token for project 2
	req2 := &types.CreateSharedTokenRequest{
		ProjectID:       "proj-2",
		Name:            "Token 3",
		Role:            types.ProjectRoleMember,
		CreatedBy:       "admin@example.com",
		ExpiresAt:       &expiresAt,
		RedemptionLimit: 50,
	}
	_, err := manager.CreateSharedToken(ctx, req2)
	require.NoError(t, err)

	// List tokens for project 1
	tokens, err := manager.ListSharedTokens(ctx, "proj-1")
	require.NoError(t, err)
	assert.Len(t, tokens, 2)

	// List tokens for project 2
	tokens2, err := manager.ListSharedTokens(ctx, "proj-2")
	require.NoError(t, err)
	assert.Len(t, tokens2, 1)

	// List tokens for nonexistent project
	tokens3, err := manager.ListSharedTokens(ctx, "proj-nonexistent")
	require.NoError(t, err)
	assert.Len(t, tokens3, 0)
}

func TestRevokeSharedToken_Success(t *testing.T) {
	manager := NewSharedTokenManager()
	ctx := context.Background()

	// Create token
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	createReq := &types.CreateSharedTokenRequest{
		ProjectID:       "proj-123",
		Name:            "Workshop Token",
		Role:            types.ProjectRoleMember,
		CreatedBy:       "admin@example.com",
		ExpiresAt:       &expiresAt,
		RedemptionLimit: 50,
	}

	token, err := manager.CreateSharedToken(ctx, createReq)
	require.NoError(t, err)

	// Revoke token
	err = manager.RevokeSharedToken(ctx, token.Token)
	require.NoError(t, err)

	// Verify token is revoked
	retrieved, err := manager.GetSharedToken(ctx, token.Token)
	require.NoError(t, err)
	assert.Equal(t, types.SharedTokenRevoked, retrieved.Status)
}

func TestRevokeSharedToken_NotFound(t *testing.T) {
	manager := NewSharedTokenManager()
	ctx := context.Background()

	err := manager.RevokeSharedToken(ctx, "NONEXISTENT-TOKEN")

	require.Error(t, err)
	assert.Equal(t, types.ErrSharedTokenNotFound, err)
}

func TestExtendExpiration_Success(t *testing.T) {
	manager := NewSharedTokenManager()
	ctx := context.Background()

	// Create token
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	createReq := &types.CreateSharedTokenRequest{
		ProjectID:       "proj-123",
		Name:            "Workshop Token",
		Role:            types.ProjectRoleMember,
		CreatedBy:       "admin@example.com",
		ExpiresAt:       &expiresAt,
		RedemptionLimit: 50,
	}

	token, err := manager.CreateSharedToken(ctx, createReq)
	require.NoError(t, err)

	originalExpiration := *token.ExpiresAt

	// Extend expiration by 3 days
	err = manager.ExtendExpiration(ctx, token.Token, 3*24*time.Hour)
	require.NoError(t, err)

	// Verify expiration extended
	retrieved, err := manager.GetSharedToken(ctx, token.Token)
	require.NoError(t, err)
	require.NotNil(t, retrieved.ExpiresAt)

	expectedExpiration := originalExpiration.Add(3 * 24 * time.Hour)
	assert.WithinDuration(t, expectedExpiration, *retrieved.ExpiresAt, 5*time.Second)
}

func TestExtendExpiration_NoOriginalExpiration(t *testing.T) {
	manager := NewSharedTokenManager()
	ctx := context.Background()

	// Create token without expiration
	createReq := &types.CreateSharedTokenRequest{
		ProjectID:       "proj-123",
		Name:            "Workshop Token",
		Role:            types.ProjectRoleMember,
		CreatedBy:       "admin@example.com",
		RedemptionLimit: 50,
	}

	token, err := manager.CreateSharedToken(ctx, createReq)
	require.NoError(t, err)
	require.Nil(t, token.ExpiresAt)

	// Extend expiration (should set new expiration)
	err = manager.ExtendExpiration(ctx, token.Token, 7*24*time.Hour)
	require.NoError(t, err)

	// Verify expiration set
	retrieved, err := manager.GetSharedToken(ctx, token.Token)
	require.NoError(t, err)
	require.NotNil(t, retrieved.ExpiresAt)

	expectedExpiration := time.Now().Add(7 * 24 * time.Hour)
	assert.WithinDuration(t, expectedExpiration, *retrieved.ExpiresAt, 5*time.Second)
}

func TestExtendExpiration_RevokedToken(t *testing.T) {
	manager := NewSharedTokenManager()
	ctx := context.Background()

	// Create and revoke token
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	createReq := &types.CreateSharedTokenRequest{
		ProjectID:       "proj-123",
		Name:            "Workshop Token",
		Role:            types.ProjectRoleMember,
		CreatedBy:       "admin@example.com",
		ExpiresAt:       &expiresAt,
		RedemptionLimit: 50,
	}

	token, err := manager.CreateSharedToken(ctx, createReq)
	require.NoError(t, err)

	err = manager.RevokeSharedToken(ctx, token.Token)
	require.NoError(t, err)

	// Try to extend revoked token
	err = manager.ExtendExpiration(ctx, token.Token, 3*24*time.Hour)
	require.Error(t, err)
	assert.Equal(t, types.ErrSharedTokenRevoked, err)
}

func TestExtendExpiration_NotFound(t *testing.T) {
	manager := NewSharedTokenManager()
	ctx := context.Background()

	err := manager.ExtendExpiration(ctx, "NONEXISTENT-TOKEN", 3*24*time.Hour)

	require.Error(t, err)
	assert.Equal(t, types.ErrSharedTokenNotFound, err)
}

func TestGenerateTokenCode(t *testing.T) {
	tests := []struct {
		name     string
		expected string // Partial match (without year/suffix)
	}{
		{"NeurIPS Workshop", "NEURIPS-WORKSHOP"},
		{"ML Conference", "ML-CONFERENCE"},
		{"The AI Summit", "AI-SUMMIT"},
		{"A Tutorial", "TUTORIAL"},
		{"Workshop", "WORKSHOP"},
		{"Very Long Conference Name With Many Words", "VERY-LONG-CONFERENCE"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := generateTokenCode(tt.name)
			assert.Contains(t, code, tt.expected)
			assert.Contains(t, code, "-2026") // Current year
		})
	}
}

func TestGenerateTokenCode_Uniqueness(t *testing.T) {
	// Same name should generate different codes (due to random suffix)
	code1 := generateTokenCode("Workshop Token")
	code2 := generateTokenCode("Workshop Token")

	assert.NotEqual(t, code1, code2, "Token codes should be unique")
}

func TestGenerateTokenID_Uniqueness(t *testing.T) {
	// Generate multiple IDs and verify uniqueness
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := generateTokenID()
		assert.False(t, ids[id], "Token ID should be unique")
		ids[id] = true
		assert.Len(t, id, 32, "Token ID should be 32 characters (16 bytes hex)")
	}
}

func TestSharedTokenManager_Concurrency(t *testing.T) {
	manager := NewSharedTokenManager()
	ctx := context.Background()

	// Create token
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	createReq := &types.CreateSharedTokenRequest{
		ProjectID:       "proj-123",
		Name:            "Workshop Token",
		Role:            types.ProjectRoleMember,
		CreatedBy:       "admin@example.com",
		ExpiresAt:       &expiresAt,
		RedemptionLimit: 100,
	}

	token, err := manager.CreateSharedToken(ctx, createReq)
	require.NoError(t, err)

	// Concurrent redemptions
	var wg sync.WaitGroup
	successCount := 0
	var mu sync.Mutex

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			redeemReq := &types.RedeemSharedTokenRequest{
				Token:      token.Token,
				RedeemedBy: fmt.Sprintf("user%d@example.com", id),
			}

			_, err := manager.RedeemToken(ctx, redeemReq)
			if err == nil {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	// Verify all redemptions succeeded
	assert.Equal(t, 50, successCount)

	// Verify token state
	retrieved, err := manager.GetSharedToken(ctx, token.Token)
	require.NoError(t, err)
	assert.Equal(t, 50, retrieved.RedemptionCount)
	assert.Len(t, retrieved.Redemptions, 50)
}

func TestSharedTokenManager_ConcurrentDuplicateRedemptions(t *testing.T) {
	manager := NewSharedTokenManager()
	ctx := context.Background()

	// Create token
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	createReq := &types.CreateSharedTokenRequest{
		ProjectID:       "proj-123",
		Name:            "Workshop Token",
		Role:            types.ProjectRoleMember,
		CreatedBy:       "admin@example.com",
		ExpiresAt:       &expiresAt,
		RedemptionLimit: 100,
	}

	token, err := manager.CreateSharedToken(ctx, createReq)
	require.NoError(t, err)

	// Same user tries to redeem concurrently (race condition test)
	var wg sync.WaitGroup
	successCount := 0
	var mu sync.Mutex

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			redeemReq := &types.RedeemSharedTokenRequest{
				Token:      token.Token,
				RedeemedBy: "user@example.com",
			}

			_, err := manager.RedeemToken(ctx, redeemReq)
			if err == nil {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	// Only one redemption should succeed
	assert.Equal(t, 1, successCount)

	// Verify token state
	retrieved, err := manager.GetSharedToken(ctx, token.Token)
	require.NoError(t, err)
	assert.Equal(t, 1, retrieved.RedemptionCount)
	assert.Len(t, retrieved.Redemptions, 1)
}
