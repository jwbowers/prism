// Package aws provides enhanced error messages with actionable guidance
package aws

import (
	"fmt"
	"strings"
)

// EnhancedError wraps an AWS error with actionable guidance
type EnhancedError struct {
	OriginalError error
	Operation     string // e.g., "launch instance", "start instance"
	Suggestions   []string
	DocsLink      string
	Category      string // e.g., "permissions", "quota", "configuration"
}

func (e *EnhancedError) Error() string {
	var msg strings.Builder

	// Start with the original error
	msg.WriteString(fmt.Sprintf("❌ %s failed: %v\n\n", e.Operation, e.OriginalError))

	// Add category if available
	if e.Category != "" {
		msg.WriteString(fmt.Sprintf("Category: %s\n\n", e.Category))
	}

	// Add suggestions
	if len(e.Suggestions) > 0 {
		msg.WriteString("💡 Suggestions to fix this:\n")
		for i, suggestion := range e.Suggestions {
			msg.WriteString(fmt.Sprintf("   %d. %s\n", i+1, suggestion))
		}
		msg.WriteString("\n")
	}

	// Add documentation link
	if e.DocsLink != "" {
		msg.WriteString(fmt.Sprintf("📚 Learn more: %s\n", e.DocsLink))
	}

	return msg.String()
}

// EnhanceError wraps an AWS error with actionable guidance
func EnhanceError(err error, operation string) error {
	if err == nil {
		return nil
	}

	errMsg := err.Error()

	// Detect error type and provide appropriate guidance
	if strings.Contains(errMsg, "UnauthorizedOperation") || strings.Contains(errMsg, "not authorized") {
		return &EnhancedError{
			OriginalError: err,
			Operation:     operation,
			Category:      "IAM Permissions",
			Suggestions: []string{
				"Check your IAM user/role has the required EC2 permissions",
				"Verify your AWS credentials: aws sts get-caller-identity",
				"Attach the 'AmazonEC2FullAccess' policy to your IAM user/role",
				"If using assumed role, check the trust policy allows your identity",
			},
			DocsLink: "https://docs.prism.org/troubleshooting/permissions",
		}
	}

	if strings.Contains(errMsg, "InvalidKeyPair.NotFound") {
		return &EnhancedError{
			OriginalError: err,
			Operation:     operation,
			Category:      "SSH Key Configuration",
			Suggestions: []string{
				"The SSH key pair doesn't exist in your AWS region",
				"Create a key pair: aws ec2 create-key-pair --key-name prism-key",
				"Or specify an existing key: prism launch <template> <name> --ssh-key <key-name>",
				"Check available keys: aws ec2 describe-key-pairs",
			},
			DocsLink: "https://docs.prism.org/getting-started/ssh-keys",
		}
	}

	if strings.Contains(errMsg, "InvalidParameterValue") {
		return &EnhancedError{
			OriginalError: err,
			Operation:     operation,
			Category:      "Invalid Configuration",
			Suggestions: []string{
				"Check the instance type is available in your region",
				"Verify the AMI exists in your region",
				"Ensure the subnet ID is valid",
				"Check the security group ID exists",
			},
			DocsLink: "https://docs.prism.org/troubleshooting/configuration",
		}
	}

	if strings.Contains(errMsg, "InsufficientInstanceCapacity") || strings.Contains(errMsg, "capacity") {
		return &EnhancedError{
			OriginalError: err,
			Operation:     operation,
			Category:      "AWS Capacity",
			Suggestions: []string{
				"AWS has insufficient capacity for this instance type in this AZ",
				"Try a different availability zone: prism launch <template> <name> --az us-east-1b",
				"Try a different instance size: prism launch <template> <name> --size M",
				"Use spot instances as alternative: prism launch <template> <name> --spot",
				"Wait a few minutes and try again",
			},
			DocsLink: "https://docs.prism.org/troubleshooting/capacity",
		}
	}

	if strings.Contains(errMsg, "InstanceLimitExceeded") || strings.Contains(errMsg, "vcpu limit") {
		return &EnhancedError{
			OriginalError: err,
			Operation:     operation,
			Category:      "AWS Quota Limit",
			Suggestions: []string{
				"You've reached your EC2 instance quota for this region",
				"View quotas: prism admin quotas list",
				"Request quota increase: https://console.aws.amazon.com/servicequotas/",
				"Stop unused instances: prism list (then prism stop <name>)",
				"Delete terminated instances: prism delete <name>",
			},
			DocsLink: "https://docs.prism.org/troubleshooting/quotas",
		}
	}

	if strings.Contains(errMsg, "VolumeInUse") {
		return &EnhancedError{
			OriginalError: err,
			Operation:     operation,
			Category:      "EBS Volume In Use",
			Suggestions: []string{
				"The EBS volume is still attached to another instance",
				"Detach volume first: prism storage detach <volume-name>",
				"Wait for previous detachment to complete",
				"Check volume status: prism storage list",
			},
			DocsLink: "https://docs.prism.org/storage/ebs",
		}
	}

	if strings.Contains(errMsg, "InvalidSubnetID") || strings.Contains(errMsg, "subnet") {
		return &EnhancedError{
			OriginalError: err,
			Operation:     operation,
			Category:      "Network Configuration",
			Suggestions: []string{
				"Create a default VPC: aws ec2 create-default-vpc",
				"Check available subnets: aws ec2 describe-subnets",
				"Specify a valid subnet: prism launch <template> <name> --subnet subnet-xxxxx",
				"Ensure subnet is in the correct availability zone",
			},
			DocsLink: "https://docs.prism.org/networking/vpc",
		}
	}

	if strings.Contains(errMsg, "RequestLimitExceeded") || strings.Contains(errMsg, "Throttling") {
		return &EnhancedError{
			OriginalError: err,
			Operation:     operation,
			Category:      "API Rate Limit",
			Suggestions: []string{
				"AWS API rate limit exceeded (this should auto-retry)",
				"If this persists, wait a few minutes and try again",
				"Reduce number of concurrent operations",
				"Check for automation scripts making excessive API calls",
			},
			DocsLink: "https://docs.prism.org/troubleshooting/rate-limits",
		}
	}

	if strings.Contains(errMsg, "DryRunOperation") {
		return &EnhancedError{
			OriginalError: err,
			Operation:     operation,
			Category:      "Dry Run Mode",
			Suggestions: []string{
				"This was a dry run - no actual resources were created",
				"Remove --dry-run flag to actually launch the instance",
				"The dry run succeeded, meaning you have the necessary permissions",
			},
			DocsLink: "https://docs.prism.org/cli/dry-run",
		}
	}

	if strings.Contains(errMsg, "Unsupported") {
		return &EnhancedError{
			OriginalError: err,
			Operation:     operation,
			Category:      "Unsupported Configuration",
			Suggestions: []string{
				"This instance type or feature is not supported in your region",
				"Check available instance types: aws ec2 describe-instance-types",
				"Try a different region: prism profile set-region us-west-2",
				"Use a supported alternative instance type",
			},
			DocsLink: "https://docs.prism.org/troubleshooting/compatibility",
		}
	}

	// Default enhanced error for unrecognized errors
	return &EnhancedError{
		OriginalError: err,
		Operation:     operation,
		Category:      "AWS Error",
		Suggestions: []string{
			"Check AWS service health: https://status.aws.amazon.com",
			"Verify your AWS credentials: aws sts get-caller-identity",
			"Review AWS CloudTrail logs for more details",
			"Try again in a few minutes (may be transient)",
		},
		DocsLink: "https://docs.prism.org/troubleshooting",
	}
}

// Common documentation links
const (
	DocsURLBase         = "https://docs.prism.org"
	DocsPermissions     = DocsURLBase + "/troubleshooting/permissions"
	DocsSSHKeys         = DocsURLBase + "/getting-started/ssh-keys"
	DocsConfiguration   = DocsURLBase + "/troubleshooting/configuration"
	DocsCapacity        = DocsURLBase + "/troubleshooting/capacity"
	DocsQuotas          = DocsURLBase + "/troubleshooting/quotas"
	DocsStorage         = DocsURLBase + "/storage/ebs"
	DocsNetworking      = DocsURLBase + "/networking/vpc"
	DocsRateLimits      = DocsURLBase + "/troubleshooting/rate-limits"
	DocsDryRun          = DocsURLBase + "/cli/dry-run"
	DocsCompatibility   = DocsURLBase + "/troubleshooting/compatibility"
	DocsTroubleshooting = DocsURLBase + "/troubleshooting"
)

// IsRetryable returns true if the error is transient and should be retried
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	errMsg := err.Error()

	// Retryable errors
	retryablePatterns := []string{
		"Throttling",
		"RequestLimitExceeded",
		"ServiceUnavailable",
		"InternalError",
		"timeout",
		"connection reset",
		"connection refused",
		"temporarily unavailable",
	}

	for _, pattern := range retryablePatterns {
		if strings.Contains(errMsg, pattern) {
			return true
		}
	}

	return false
}
