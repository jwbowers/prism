package localstack

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// WaitForReady waits for LocalStack to be ready to accept requests
// This is useful in tests and CI environments where LocalStack may still be starting
func WaitForReady(ctx context.Context, timeout time.Duration) error {
	endpoint := GetEndpoint()
	healthURL := fmt.Sprintf("%s/_localstack/health", endpoint)

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		req, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
		if err != nil {
			return fmt.Errorf("failed to create health check request: %w", err)
		}

		resp, err := client.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return nil
		}

		if resp != nil {
			resp.Body.Close()
		}

		// Wait before retrying
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
			// Continue to next iteration
		}
	}

	return fmt.Errorf("LocalStack not ready after %v (endpoint: %s)", timeout, endpoint)
}

// IsHealthy checks if LocalStack is currently healthy
func IsHealthy(ctx context.Context) bool {
	endpoint := GetEndpoint()
	healthURL := fmt.Sprintf("%s/_localstack/health", endpoint)

	client := &http.Client{
		Timeout: 3 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
	if err != nil {
		return false
	}

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// GetServiceStatus returns the status of LocalStack services
// Returns a map of service names to their status ("available", "starting", "error", etc.)
func GetServiceStatus(ctx context.Context) (map[string]string, error) {
	if !IsLocalStackEnabled() {
		return nil, fmt.Errorf("LocalStack is not enabled")
	}

	endpoint := GetEndpoint()
	healthURL := fmt.Sprintf("%s/_localstack/health", endpoint)

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("health check request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("health check returned status %d", resp.StatusCode)
	}

	// Parse health response
	// LocalStack health endpoint returns JSON like: {"services": {"ec2": "available", "efs": "available"}}
	var health struct {
		Services map[string]string `json:"services"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		return nil, fmt.Errorf("failed to parse health response: %w", err)
	}

	return health.Services, nil
}

// RequiredServices lists the AWS services that Prism requires
var RequiredServices = []string{
	"ec2",
	"efs",
	"s3",
	"ssm",
	"iam",
	"sts",
}

// VerifyRequiredServices checks that all required services are available in LocalStack
func VerifyRequiredServices(ctx context.Context) error {
	services, err := GetServiceStatus(ctx)
	if err != nil {
		return fmt.Errorf("failed to get service status: %w", err)
	}

	var unavailable []string
	for _, required := range RequiredServices {
		status, exists := services[required]
		if !exists || status != "available" {
			unavailable = append(unavailable, fmt.Sprintf("%s (%s)", required, status))
		}
	}

	if len(unavailable) > 0 {
		return fmt.Errorf("required services not available in LocalStack: %v", unavailable)
	}

	return nil
}

// ResetLocalStack resets LocalStack state (requires LocalStack Pro for full reset)
// This is a best-effort operation for the free version
func ResetLocalStack(ctx context.Context) error {
	if !IsLocalStackEnabled() {
		return fmt.Errorf("LocalStack is not enabled")
	}

	// In LocalStack free version, we can't do a full reset
	// This would require the Pro version's state management API
	// For now, just verify connectivity
	if !IsHealthy(ctx) {
		return fmt.Errorf("LocalStack is not healthy, cannot reset")
	}

	// Note: To fully reset LocalStack, restart the container:
	// docker-compose -f test/localstack/docker-compose.yml restart
	return nil
}
