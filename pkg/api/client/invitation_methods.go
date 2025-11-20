// Package client provides invitation API methods for Prism v0.5.11+
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/scttfrdmn/prism/pkg/types"
)

// GetInvitationByToken retrieves an invitation by token
// GET /api/v1/invitations/{token}
func (c *HTTPClient) GetInvitationByToken(ctx context.Context, token string) (*GetInvitationResponse, error) {
	path := fmt.Sprintf("/api/v1/invitations/%s", token)

	resp, err := c.makeRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get invitation: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, parseHTTPError(resp)
	}

	var result GetInvitationResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// AcceptInvitation accepts an invitation
// POST /api/v1/invitations/{token}/accept
func (c *HTTPClient) AcceptInvitation(ctx context.Context, token string) (*InvitationActionResponse, error) {
	path := fmt.Sprintf("/api/v1/invitations/%s/accept", token)

	resp, err := c.makeRequest(ctx, http.MethodPost, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to accept invitation: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, parseHTTPError(resp)
	}

	var result InvitationActionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// DeclineInvitation declines an invitation
// POST /api/v1/invitations/{token}/decline
func (c *HTTPClient) DeclineInvitation(ctx context.Context, token string, reason string) (*InvitationActionResponse, error) {
	path := fmt.Sprintf("/api/v1/invitations/%s/decline", token)

	var body interface{}
	if reason != "" {
		body = map[string]string{
			"reason": reason,
		}
	}

	resp, err := c.makeRequest(ctx, http.MethodPost, path, body)
	if err != nil {
		return nil, fmt.Errorf("failed to decline invitation: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, parseHTTPError(resp)
	}

	var result InvitationActionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// SendInvitation sends an invitation to collaborate on a project
// POST /api/v1/projects/{projectID}/invitations
func (c *HTTPClient) SendInvitation(ctx context.Context, projectID string, req SendInvitationRequest) (*SendInvitationResponse, error) {
	path := fmt.Sprintf("/api/v1/projects/%s/invitations", projectID)

	resp, err := c.makeRequest(ctx, http.MethodPost, path, req)
	if err != nil {
		return nil, fmt.Errorf("failed to send invitation: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, parseHTTPError(resp)
	}

	var result SendInvitationResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// SendBulkInvitation sends bulk invitations to multiple users
// POST /api/v1/projects/{projectID}/invitations/bulk
func (c *HTTPClient) SendBulkInvitation(ctx context.Context, projectID string, req *types.BulkInvitationRequest) (*types.BulkInvitationResponse, error) {
	path := fmt.Sprintf("/api/v1/projects/%s/invitations/bulk", projectID)

	resp, err := c.makeRequest(ctx, http.MethodPost, path, req)
	if err != nil {
		return nil, fmt.Errorf("failed to send bulk invitations: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, parseHTTPError(resp)
	}

	var result types.BulkInvitationResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// parseHTTPError extracts error message from HTTP response
func parseHTTPError(resp *http.Response) error {
	var errMsg struct {
		Error   string `json:"error"`
		Message string `json:"message"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&errMsg); err != nil {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	msg := errMsg.Error
	if msg == "" {
		msg = errMsg.Message
	}
	if msg == "" {
		msg = resp.Status
	}

	// Map specific HTTP statuses to typed errors
	switch resp.StatusCode {
	case http.StatusNotFound:
		if contains(msg, "invitation") {
			return types.ErrInvitationNotFound
		}
		return fmt.Errorf("not found: %s", msg)
	case http.StatusGone:
		return types.ErrInvitationExpired
	case http.StatusConflict:
		if contains(msg, "already") {
			return types.ErrInvitationAlreadyUsed
		}
		return fmt.Errorf("conflict: %s", msg)
	default:
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, msg)
	}
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
}

// CreateSharedToken creates a shared invitation token
// POST /api/v1/projects/{projectID}/invitations/shared
func (c *HTTPClient) CreateSharedToken(ctx context.Context, projectID string, req *types.CreateSharedTokenRequest) (*types.SharedInvitationToken, error) {
	path := fmt.Sprintf("/api/v1/projects/%s/invitations/shared", projectID)

	resp, err := c.makeRequest(ctx, http.MethodPost, path, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create shared token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, parseHTTPError(resp)
	}

	var result struct {
		Token           string                  `json:"token"`
		ID              string                  `json:"id"`
		Name            string                  `json:"name"`
		Role            types.ProjectRole       `json:"role"`
		RedemptionLimit int                     `json:"redemption_limit"`
		Redemptions     int                     `json:"redemptions"`
		Status          types.SharedTokenStatus `json:"status"`
		RedemptionURL   string                  `json:"redemption_url"`
		QRCodeURL       string                  `json:"qr_code_url"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert to SharedInvitationToken
	token := &types.SharedInvitationToken{
		ID:              result.ID,
		ProjectID:       projectID,
		Token:           result.Token,
		Name:            result.Name,
		Role:            result.Role,
		RedemptionLimit: result.RedemptionLimit,
		RedemptionCount: result.Redemptions,
		Status:          result.Status,
	}

	return token, nil
}

// RedeemSharedToken redeems a shared token
// POST /api/v1/invitations/shared/redeem
func (c *HTTPClient) RedeemSharedToken(ctx context.Context, req *types.RedeemSharedTokenRequest) (*types.RedeemSharedTokenResponse, error) {
	path := "/api/v1/invitations/shared/redeem"

	resp, err := c.makeRequest(ctx, http.MethodPost, path, req)
	if err != nil {
		return nil, fmt.Errorf("failed to redeem token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, parseHTTPError(resp)
	}

	var result types.RedeemSharedTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GetSharedToken retrieves shared token information
// GET /api/v1/invitations/shared/{token}
func (c *HTTPClient) GetSharedToken(ctx context.Context, token string) (*types.SharedInvitationToken, error) {
	path := fmt.Sprintf("/api/v1/invitations/shared/%s", token)

	resp, err := c.makeRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get shared token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, parseHTTPError(resp)
	}

	var result types.SharedInvitationToken
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// ListSharedTokens lists all shared tokens for a project
// GET /api/v1/projects/{projectID}/invitations/shared
func (c *HTTPClient) ListSharedTokens(ctx context.Context, projectID string) ([]*types.SharedInvitationToken, error) {
	path := fmt.Sprintf("/api/v1/projects/%s/invitations/shared", projectID)

	resp, err := c.makeRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list shared tokens: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, parseHTTPError(resp)
	}

	var result struct {
		Tokens []*types.SharedInvitationToken `json:"tokens"`
		Count  int                            `json:"count"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Tokens, nil
}

// ExtendSharedToken extends the expiration of a shared token
// PATCH /api/v1/invitations/shared/{token}/extend
func (c *HTTPClient) ExtendSharedToken(ctx context.Context, token string, addDays int) error {
	path := fmt.Sprintf("/api/v1/invitations/shared/%s/extend", token)

	body := map[string]int{
		"add_days": addDays,
	}

	resp, err := c.makeRequest(ctx, http.MethodPatch, path, body)
	if err != nil {
		return fmt.Errorf("failed to extend token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return parseHTTPError(resp)
	}

	return nil
}

// RevokeSharedToken revokes a shared token
// DELETE /api/v1/invitations/shared/{token}
func (c *HTTPClient) RevokeSharedToken(ctx context.Context, token string) error {
	path := fmt.Sprintf("/api/v1/invitations/shared/%s", token)

	resp, err := c.makeRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return fmt.Errorf("failed to revoke token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return parseHTTPError(resp)
	}

	return nil
}
