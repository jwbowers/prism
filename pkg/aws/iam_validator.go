package aws

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamTypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// IAMValidationResult contains the result of IAM permission validation before launch.
type IAMValidationResult struct {
	Valid              bool      `json:"valid"`
	Identity           string    `json:"identity"` // Caller ARN
	AccountID          string    `json:"account_id"`
	UserID             string    `json:"user_id"`
	MissingPermissions []string  `json:"missing_permissions,omitempty"`
	Warnings           []string  `json:"warnings,omitempty"`
	ValidatedAt        time.Time `json:"validated_at"`
}

// requiredLaunchPermissions are the minimum IAM permissions needed to launch Prism instances.
var requiredLaunchPermissions = []string{
	"ec2:RunInstances",
	"ec2:DescribeInstances",
	"ec2:DescribeInstanceTypes",
	"ec2:DescribeSubnets",
	"ec2:DescribeVpcs",
	"ec2:DescribeSecurityGroups",
	"ec2:DescribeKeyPairs",
	"ec2:CreateTags",
	"ec2:TerminateInstances",
	"ec2:StopInstances",
	"ec2:StartInstances",
}

// ValidateIAMPermissions checks that the current AWS credentials are valid and have the
// permissions needed to launch instances. It uses STS GetCallerIdentity to verify
// credentials, then optionally uses IAM SimulatePrincipalPolicy to enumerate missing
// permissions (requires iam:SimulatePrincipalPolicy — falls back gracefully if absent).
func (m *Manager) ValidateIAMPermissions(ctx context.Context) (*IAMValidationResult, error) {
	result := &IAMValidationResult{
		ValidatedAt: time.Now(),
	}

	// Step 1: Verify credentials are valid and not expired.
	stsClient := sts.NewFromConfig(m.cfg)
	identity, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, fmt.Errorf("credentials invalid or expired (run 'aws login' or 'aws configure'): %w", err)
	}

	if identity.Arn != nil {
		result.Identity = *identity.Arn
	}
	if identity.Account != nil {
		result.AccountID = *identity.Account
	}
	if identity.UserId != nil {
		result.UserID = *identity.UserId
	}

	// Step 2: Attempt permission simulation (may be unavailable with limited creds).
	missing, err := m.simulateLaunchPermissions(ctx, result.Identity)
	if err != nil {
		// Simulation not available — common for role-assumed credentials without iam:Simulate*.
		log.Printf("[IAM] Policy simulation unavailable (%v) — using credential check only", err)
		result.Warnings = append(result.Warnings,
			"Full permission check requires iam:SimulatePrincipalPolicy — verified credentials only")
	} else {
		result.MissingPermissions = missing
	}

	// Step 3: Check for common configuration issues (instance profile, etc.).
	result.Warnings = append(result.Warnings, m.checkCommonIAMWarnings(ctx)...)

	result.Valid = len(result.MissingPermissions) == 0
	return result, nil
}

// simulateLaunchPermissions uses IAM SimulatePrincipalPolicy to enumerate which required
// permissions are absent for the caller's principal.
func (m *Manager) simulateLaunchPermissions(ctx context.Context, principalARN string) ([]string, error) {
	star := "*"
	input := &iam.SimulatePrincipalPolicyInput{
		PolicySourceArn: &principalARN,
		ActionNames:     requiredLaunchPermissions,
		ResourceArns:    []string{star},
	}

	out, err := m.iam.SimulatePrincipalPolicy(ctx, input)
	if err != nil {
		return nil, err
	}

	var missing []string
	for _, r := range out.EvaluationResults {
		if r.EvalDecision != iamTypes.PolicyEvaluationDecisionTypeAllowed {
			if r.EvalActionName != nil {
				missing = append(missing, *r.EvalActionName)
			}
		}
	}
	return missing, nil
}

// checkCommonIAMWarnings returns advisory warnings for common misconfigurations.
func (m *Manager) checkCommonIAMWarnings(ctx context.Context) []string {
	var warnings []string
	if !m.checkIAMInstanceProfileExists("Prism-Instance-Profile") {
		warnings = append(warnings,
			"IAM instance profile 'Prism-Instance-Profile' not found — SSM session manager and autonomous monitoring will be unavailable")
	}
	return warnings
}

// GetRequiredLaunchPermissions returns the list of IAM permissions required for launching instances.
func GetRequiredLaunchPermissions() []string {
	result := make([]string, len(requiredLaunchPermissions))
	copy(result, requiredLaunchPermissions)
	return result
}
