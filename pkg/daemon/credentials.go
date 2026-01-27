// Package daemon provides credential error detection for graceful handling
package daemon

import (
	"strings"
)

// isCredentialError checks if an error is related to missing/invalid AWS credentials
func isCredentialError(err error) bool {
	if err == nil {
		return false
	}

	errMsg := strings.ToLower(err.Error())

	// Common credential error patterns
	credentialPatterns := []string{
		"no ec2 imds role found",
		"failed to refresh cached credentials",
		"no valid providers in chain",
		"could not find credentials",
		"unable to locate credentials",
		"credentials not found",
		"no credentials",
		"operation error ec2imds",
		"credential",
		"i/o timeout",              // Often indicates IMDS timeout
		"dial tcp 169.254.169.254", // EC2 IMDS endpoint timeout
	}

	for _, pattern := range credentialPatterns {
		if strings.Contains(errMsg, pattern) {
			return true
		}
	}

	return false
}

// getCredentialErrorMessage returns a helpful error message for credential issues
func getCredentialErrorMessage() string {
	return `AWS credentials required for this operation.

To configure AWS credentials:
  1. Run: aws configure
  2. Or set environment variables:
     export AWS_ACCESS_KEY_ID=your-key
     export AWS_SECRET_ACCESS_KEY=your-secret
     export AWS_DEFAULT_REGION=us-west-2
  3. Then create a Prism profile:
     prism profile create my-profile --aws-profile default --region us-west-2

For more information: https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html`
}

// getReducedModeBanner returns startup banner for reduced functionality mode
func getReducedModeBanner() string {
	return `⚠️  AWS credentials not found - running in reduced functionality mode

Available operations:
  ✓ Template validation and discovery
  ✓ Profile management (local profiles)
  ✓ Configuration management
  ✓ GUI/TUI interface navigation

Unavailable operations:
  ✗ Instance operations (launch, list, connect)
  ✗ Storage management (volumes, EFS)
  ✗ AWS resource queries

` + getCredentialErrorMessage()
}
