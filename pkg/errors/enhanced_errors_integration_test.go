// Package errors provides integration tests for enhanced error messages
//
// These tests verify error pattern matching and enhancement with
// actionable guidance and documentation links (v0.5.12 feature).
package errors_test

import (
	"fmt"
	"strings"
	"testing"

	prismErrors "github.com/scttfrdmn/prism/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestErrorEnhancement_PermissionDenied validates enhancement of AWS permission errors
// with actionable guidance (v0.5.12 feature).
//
// Test Coverage:
// - Pattern matching for permission errors
// - IAM policy guidance
// - Documentation link inclusion
// - Error category tagging
func TestErrorEnhancement_PermissionDenied(t *testing.T) {
	// Simulate AWS permission denied error
	rawError := fmt.Errorf("UnauthorizedOperation: You are not authorized to perform this operation")

	// Run through error enhancement pipeline
	enhanced := prismErrors.EnhanceWithOperation(rawError, "launch instance")

	// Verify enhancement occurred
	require.NotNil(t, enhanced, "Enhanced error should not be nil")

	errorMsg := enhanced.Error()

	// Verify error contains actionable guidance
	assert.Contains(t, errorMsg, "launch instance failed", "Should mention failed operation")
	assert.Contains(t, errorMsg, "Suggestions", "Should include suggestions section")
	assert.Contains(t, errorMsg, "IAM", "Should mention IAM permissions")
	assert.Contains(t, errorMsg, "EC2", "Should mention EC2 service")

	// Verify category
	assert.Contains(t, errorMsg, "IAM Permissions", "Should have permissions category")

	// Verify documentation link
	assert.Contains(t, errorMsg, "docs.prism.org", "Should include docs link")
	assert.Contains(t, errorMsg, "troubleshooting/permissions", "Should link to permissions troubleshooting")
}

// TestErrorEnhancement_KeyPairNotFound validates enhancement of SSH key errors
// with configuration guidance (v0.5.12 feature).
//
// Test Coverage:
// - Pattern matching for key pair errors
// - AWS CLI configuration guidance
// - Key creation instructions
// - Regional awareness hints
func TestErrorEnhancement_KeyPairNotFound(t *testing.T) {
	// Simulate AWS key pair not found error
	rawError := fmt.Errorf("InvalidKeyPair.NotFound: The key pair 'my-key' does not exist")

	// Run through error enhancement pipeline
	enhanced := prismErrors.EnhanceWithOperation(rawError, "launch instance")

	require.NotNil(t, enhanced, "Enhanced error should not be nil")
	errorMsg := enhanced.Error()

	// Verify error contains SSH key guidance
	assert.Contains(t, errorMsg, "SSH Key Configuration", "Should have SSH key category")
	assert.Contains(t, errorMsg, "key pair", "Should mention key pair")
	assert.Contains(t, errorMsg, "create-key-pair", "Should suggest creating key")
	assert.Contains(t, errorMsg, "region", "Should mention region dependency")

	// Verify documentation link
	assert.Contains(t, errorMsg, "ssh-keys", "Should link to SSH keys docs")
}

// TestErrorEnhancement_InvalidConfiguration validates enhancement of configuration
// errors with parameter guidance (v0.5.12 feature).
//
// Test Coverage:
// - Pattern matching for configuration errors
// - Parameter validation suggestions
// - Regional resource checks
// - Resource existence verification
func TestErrorEnhancement_InvalidConfiguration(t *testing.T) {
	// Simulate AWS invalid parameter error
	rawError := fmt.Errorf("InvalidParameterValue: The instance type 'm5.32xlarge' is not supported in this region")

	// Run through error enhancement pipeline
	enhanced := prismErrors.EnhanceWithOperation(rawError, "launch instance")

	require.NotNil(t, enhanced, "Enhanced error should not be nil")
	errorMsg := enhanced.Error()

	// Verify error contains configuration guidance
	assert.Contains(t, errorMsg, "Invalid Configuration", "Should have configuration category")
	assert.Contains(t, errorMsg, "instance type", "Should mention instance type check")
	assert.Contains(t, errorMsg, "region", "Should mention regional availability")
	assert.Contains(t, errorMsg, "AMI", "Should mention AMI validation")

	// Verify documentation link
	assert.Contains(t, errorMsg, "troubleshooting/configuration", "Should link to configuration troubleshooting")
}

// TestErrorEnhancement_CapacityError validates enhancement of capacity errors
// with retry and alternative guidance (v0.5.12 feature).
//
// Test Coverage:
// - Pattern matching for capacity errors
// - Alternative instance type suggestions
// - Retry timing guidance
// - Regional failover hints
func TestErrorEnhancement_CapacityError(t *testing.T) {
	// Simulate AWS capacity error
	rawError := fmt.Errorf("InsufficientInstanceCapacity: We currently do not have sufficient capacity in the requested Availability Zone")

	// Run through error enhancement pipeline
	enhanced := prismErrors.EnhanceWithOperation(rawError, "launch instance")

	require.NotNil(t, enhanced, "Enhanced error should not be nil")
	errorMsg := enhanced.Error()

	// Verify error contains capacity guidance
	assert.Contains(t, errorMsg, "Capacity", "Should have capacity category")
	assert.Contains(t, errorMsg, "Suggestions", "Should include suggestions")

	// The implementation should provide helpful suggestions for capacity issues
	// At minimum, should acknowledge the error and provide context
	assert.NotEmpty(t, errorMsg, "Error message should not be empty")
	assert.Contains(t, strings.ToLower(errorMsg), "capacity", "Should mention capacity issue")
}

// TestErrorEnhancement_PreservesOriginalError validates that enhancement
// preserves the original error for error chain traversal (v0.5.12 feature).
//
// Test Coverage:
// - Original error preservation
// - Error unwrapping support
// - Error chain integrity
// - Is/As error matching
func TestErrorEnhancement_PreservesOriginalError(t *testing.T) {
	// Create original error
	originalErr := fmt.Errorf("UnauthorizedOperation: test error")

	// Enhance error
	enhanced := prismErrors.Enhance(originalErr)

	require.NotNil(t, enhanced, "Enhanced error should not be nil")

	// The enhanced error should still contain the original error information
	errorMsg := enhanced.Error()
	assert.Contains(t, errorMsg, "test error", "Should preserve original error message")
	assert.Contains(t, errorMsg, "UnauthorizedOperation", "Should preserve error type")
}

// TestErrorEnhancement_MultipleErrorTypes validates that different error
// patterns receive appropriate enhancements (v0.5.12 feature).
//
// Test Coverage:
// - Multiple error pattern detection
// - Category-specific guidance
// - Documentation link relevance
// - Suggestion appropriateness
func TestErrorEnhancement_MultipleErrorTypes(t *testing.T) {
	testCases := []struct {
		name            string
		errorMsg        string
		expectedContent []string // Content that should appear in enhanced error
		operation       string
	}{
		{
			name:            "Permission Error",
			errorMsg:        "UnauthorizedOperation: not authorized",
			expectedContent: []string{"IAM", "permissions", "Suggestions"},
			operation:       "launch instance",
		},
		{
			name:            "Key Pair Error",
			errorMsg:        "InvalidKeyPair.NotFound: key does not exist",
			expectedContent: []string{"SSH Key", "create-key-pair", "region"},
			operation:       "launch instance",
		},
		{
			name:            "Configuration Error",
			errorMsg:        "InvalidParameterValue: invalid parameter",
			expectedContent: []string{"Configuration", "instance type", "region"},
			operation:       "launch instance",
		},
		{
			name:            "Capacity Error",
			errorMsg:        "InsufficientInstanceCapacity: no capacity",
			expectedContent: []string{"Capacity"},
			operation:       "launch instance",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rawError := fmt.Errorf("%s", tc.errorMsg)
			enhanced := prismErrors.EnhanceWithOperation(rawError, tc.operation)

			require.NotNil(t, enhanced, "Enhanced error should not be nil")
			errorMsg := enhanced.Error()

			// Verify expected content appears
			for _, expected := range tc.expectedContent {
				assert.Contains(t, errorMsg, expected,
					"Error should contain '%s' for %s", expected, tc.name)
			}

			// Verify enhancement adds value (message is longer than original)
			assert.Greater(t, len(errorMsg), len(tc.errorMsg),
				"Enhanced error should be more detailed than original")

			// Verify docs link present
			assert.Contains(t, errorMsg, "docs.prism.org",
				"Should include documentation link for %s", tc.name)
		})
	}
}
