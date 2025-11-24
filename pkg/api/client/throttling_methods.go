package client

import (
	"context"
	"fmt"

	"github.com/scttfrdmn/prism/pkg/throttle"
)

// GetThrottlingStatus retrieves the current throttling status for a scope
func (c *HTTPClient) GetThrottlingStatus(ctx context.Context, scope string) (*throttle.Status, error) {
	url := "/api/v1/throttling/status"
	if scope != "" && scope != "global" {
		url = fmt.Sprintf("/api/v1/throttling/status?scope=%s", scope)
	}

	resp, err := c.makeRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var status throttle.Status
	if err := c.handleResponse(resp, &status); err != nil {
		return nil, err
	}

	return &status, nil
}

// ConfigureThrottling updates the throttling configuration
func (c *HTTPClient) ConfigureThrottling(ctx context.Context, config throttle.Config) error {
	resp, err := c.makeRequest(ctx, "POST", "/api/v1/throttling/configure", config)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Server returns a success message on successful configuration
	var result map[string]interface{}
	if err := c.handleResponse(resp, &result); err != nil {
		return err
	}

	return nil
}

// GetThrottlingRemaining retrieves remaining launch tokens for a scope
func (c *HTTPClient) GetThrottlingRemaining(ctx context.Context, scope string) (map[string]interface{}, error) {
	url := "/api/v1/throttling/remaining"
	if scope != "" && scope != "global" {
		url = fmt.Sprintf("/api/v1/throttling/remaining?scope=%s", scope)
	}

	resp, err := c.makeRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var remaining map[string]interface{}
	if err := c.handleResponse(resp, &remaining); err != nil {
		return nil, err
	}

	return remaining, nil
}

// SetProjectThrottleOverride sets a project-specific throttling override
func (c *HTTPClient) SetProjectThrottleOverride(ctx context.Context, projectID string, override throttle.Override) error {
	url := fmt.Sprintf("/api/v1/throttling/projects/%s/override", projectID)

	resp, err := c.makeRequest(ctx, "POST", url, override)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := c.handleResponse(resp, &result); err != nil {
		return err
	}

	return nil
}

// RemoveProjectThrottleOverride removes a project-specific override
func (c *HTTPClient) RemoveProjectThrottleOverride(ctx context.Context, projectID string) error {
	url := fmt.Sprintf("/api/v1/throttling/projects/%s/override", projectID)

	resp, err := c.makeRequest(ctx, "DELETE", url, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := c.handleResponse(resp, nil); err != nil {
		return err
	}

	return nil
}

// ListProjectThrottleOverrides lists all project throttling overrides
func (c *HTTPClient) ListProjectThrottleOverrides(ctx context.Context) ([]throttle.Override, error) {
	resp, err := c.makeRequest(ctx, "GET", "/api/v1/throttling/projects/overrides", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Overrides []throttle.Override `json:"overrides"`
		Count     int                 `json:"count"`
	}
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return result.Overrides, nil
}
