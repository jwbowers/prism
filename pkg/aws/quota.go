package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/servicequotas"
)

// QuotaCheck represents the result of a quota validation check
type QuotaCheck struct {
	HasSufficientQuota bool
	RequiredVCPUs      int
	CurrentUsage       int
	QuotaLimit         int
	AvailableVCPUs     int
	WarningMessage     string
	InstanceType       string
}

// GetInstanceTypeVCPUs returns the number of vCPUs for a given instance type
func GetInstanceTypeVCPUs(instanceType string) int {
	// Instance type vCPU mapping
	// Source: https://aws.amazon.com/ec2/instance-types/
	vcpuMap := map[string]int{
		// t3 family (general purpose burstable)
		"t3.nano":    2,
		"t3.micro":   2,
		"t3.small":   2,
		"t3.medium":  2,
		"t3.large":   2,
		"t3.xlarge":  4,
		"t3.2xlarge": 8,

		// t4g family (ARM burstable)
		"t4g.nano":    2,
		"t4g.micro":   2,
		"t4g.small":   2,
		"t4g.medium":  2,
		"t4g.large":   2,
		"t4g.xlarge":  4,
		"t4g.2xlarge": 8,

		// c7g family (compute optimized ARM)
		"c7g.medium":   1,
		"c7g.large":    2,
		"c7g.xlarge":   4,
		"c7g.2xlarge":  8,
		"c7g.4xlarge":  16,
		"c7g.8xlarge":  32,
		"c7g.12xlarge": 48,
		"c7g.16xlarge": 64,

		// m5 family (general purpose)
		"m5.large":    2,
		"m5.xlarge":   4,
		"m5.2xlarge":  8,
		"m5.4xlarge":  16,
		"m5.8xlarge":  32,
		"m5.12xlarge": 48,
		"m5.16xlarge": 64,
		"m5.24xlarge": 96,

		// m6g family (ARM general purpose)
		"m6g.medium":   1,
		"m6g.large":    2,
		"m6g.xlarge":   4,
		"m6g.2xlarge":  8,
		"m6g.4xlarge":  16,
		"m6g.8xlarge":  32,
		"m6g.12xlarge": 48,
		"m6g.16xlarge": 64,

		// r5 family (memory optimized)
		"r5.large":    2,
		"r5.xlarge":   4,
		"r5.2xlarge":  8,
		"r5.4xlarge":  16,
		"r5.8xlarge":  32,
		"r5.12xlarge": 48,
		"r5.16xlarge": 64,
		"r5.24xlarge": 96,

		// g4dn family (GPU instances)
		"g4dn.xlarge":   4,
		"g4dn.2xlarge":  8,
		"g4dn.4xlarge":  16,
		"g4dn.8xlarge":  32,
		"g4dn.12xlarge": 48,
		"g4dn.16xlarge": 64,

		// p3 family (GPU compute)
		"p3.2xlarge":  8,
		"p3.8xlarge":  32,
		"p3.16xlarge": 64,
	}

	if vcpus, exists := vcpuMap[instanceType]; exists {
		return vcpus
	}

	// Default fallback: try to parse from instance type suffix
	// Most instance types follow pattern: family.size where size indicates vCPUs
	// Examples: xlarge=4, 2xlarge=8, 4xlarge=16, etc.
	// For unknown types, conservatively estimate 2 vCPUs
	return 2
}

// CheckQuotaForInvitations validates that sufficient EC2 quota exists for bulk invitations
func (m *Manager) CheckQuotaForInvitations(ctx context.Context, instanceType string, count int) (*QuotaCheck, error) {
	// Calculate required vCPUs
	vcpusPerInstance := GetInstanceTypeVCPUs(instanceType)
	requiredVCPUs := vcpusPerInstance * count

	// Get current running instances and their vCPU usage
	currentUsage, err := m.getCurrentVCPUUsage(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get current vCPU usage: %w", err)
	}

	// Get quota limit
	quotaLimit, err := m.getEC2VCPUQuota(ctx)
	if err != nil {
		// If we can't get quota, use conservative default (64 vCPUs for standard instances)
		quotaLimit = 64
	}

	// Calculate available vCPUs
	availableVCPUs := quotaLimit - currentUsage

	// Determine if quota is sufficient
	hasSufficient := availableVCPUs >= requiredVCPUs

	result := &QuotaCheck{
		HasSufficientQuota: hasSufficient,
		RequiredVCPUs:      requiredVCPUs,
		CurrentUsage:       currentUsage,
		QuotaLimit:         quotaLimit,
		AvailableVCPUs:     availableVCPUs,
		InstanceType:       instanceType,
	}

	// Generate warning message if insufficient
	if !hasSufficient {
		deficit := requiredVCPUs - availableVCPUs
		result.WarningMessage = fmt.Sprintf(
			"Insufficient EC2 vCPU quota. Required: %d vCPUs (%d × %s), Available: %d vCPUs. "+
				"Shortfall: %d vCPUs. Please request a quota increase or reduce the number of invitations.",
			requiredVCPUs, count, instanceType, availableVCPUs, deficit,
		)
	}

	return result, nil
}

// getCurrentVCPUUsage calculates current vCPU usage from running instances
func (m *Manager) getCurrentVCPUUsage(ctx context.Context) (int, error) {
	input := &ec2.DescribeInstancesInput{
		Filters: []ec2types.Filter{
			{
				Name:   aws.String("instance-state-name"),
				Values: []string{"running", "pending"},
			},
		},
	}

	output, err := m.ec2.DescribeInstances(ctx, input)
	if err != nil {
		return 0, fmt.Errorf("failed to describe instances: %w", err)
	}

	totalVCPUs := 0
	for _, reservation := range output.Reservations {
		for _, instance := range reservation.Instances {
			instanceType := string(instance.InstanceType)
			vcpus := GetInstanceTypeVCPUs(instanceType)
			totalVCPUs += vcpus
		}
	}

	return totalVCPUs, nil
}

// getEC2VCPUQuota retrieves the EC2 vCPU quota limit from AWS Service Quotas
func (m *Manager) getEC2VCPUQuota(ctx context.Context) (int, error) {
	// Create Service Quotas client
	sqClient := servicequotas.NewFromConfig(m.cfg)

	// EC2 service code
	serviceCode := "ec2"

	// Standard (On-Demand) instances quota code
	// This is the quota for running On-Demand Standard (A, C, D, H, I, M, R, T, Z) instances
	quotaCode := "L-1216C47A"

	input := &servicequotas.GetServiceQuotaInput{
		ServiceCode: aws.String(serviceCode),
		QuotaCode:   aws.String(quotaCode),
	}

	output, err := sqClient.GetServiceQuota(ctx, input)
	if err != nil {
		// If quota retrieval fails, return default
		return 64, fmt.Errorf("failed to get service quota: %w", err)
	}

	if output.Quota != nil && output.Quota.Value != nil {
		return int(*output.Quota.Value), nil
	}

	// Default fallback
	return 64, nil
}
