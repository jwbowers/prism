package aws

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

// QuotaIncreaseRequest represents a quota increase request
type QuotaIncreaseRequest struct {
	ServiceCode     string
	QuotaCode       string
	QuotaName       string
	CurrentValue    float64
	DesiredValue    float64
	InstanceType    string
	Reason          string
	Region          string
	SupportCaseURL  string
	QuickRequestURL string
}

// QuotaRequestHelper assists users with quota increase requests
type QuotaRequestHelper struct {
	manager *QuotaManager
	region  string
}

// NewQuotaRequestHelper creates a new quota request helper
func NewQuotaRequestHelper(manager *QuotaManager, region string) *QuotaRequestHelper {
	return &QuotaRequestHelper{
		manager: manager,
		region:  region,
	}
}

// DetectQuotaError analyzes an AWS error to determine if it's quota-related
func (qrh *QuotaRequestHelper) DetectQuotaError(err error) (*QuotaIncreaseRequest, bool) {
	if err == nil {
		return nil, false
	}

	errStr := err.Error()

	// Check for quota-related errors
	quotaErrorPatterns := []string{
		"VcpuLimitExceeded",
		"InstanceLimitExceeded",
		"InsufficientInstanceCapacity",
		"RequestLimitExceeded",
		"MaximumNumberOfEBSVolumesExceeded",
		"MaximumVolumeCountExceeded",
	}

	isQuotaError := false
	for _, pattern := range quotaErrorPatterns {
		if strings.Contains(errStr, pattern) {
			isQuotaError = true
			break
		}
	}

	if !isQuotaError {
		return nil, false
	}

	// Create a basic quota request
	request := &QuotaIncreaseRequest{
		ServiceCode: "ec2",
		Region:      qrh.region,
		Reason:      "Insufficient capacity for research workload launch",
	}

	// Try to extract more details from the error
	if strings.Contains(errStr, "VcpuLimitExceeded") || strings.Contains(errStr, "InstanceLimitExceeded") {
		request.QuotaCode = QuotaCodeOnDemandStandard
		request.QuotaName = "Running On-Demand Standard Instances"
	}

	return request, true
}

// GenerateQuotaIncreaseRequest creates a detailed quota increase request
func (qrh *QuotaRequestHelper) GenerateQuotaIncreaseRequest(ctx context.Context, instanceType string) (*QuotaIncreaseRequest, error) {
	// Validate instance launch to get current quota info
	validation, err := qrh.manager.ValidateInstanceLaunch(ctx, instanceType)
	if err != nil {
		return nil, fmt.Errorf("failed to validate instance launch: %w", err)
	}

	// Determine quota code for this instance type
	quotaCode := qrh.manager.getQuotaCodeForInstanceType(instanceType)

	// Get current quota info
	quotaInfo, err := qrh.manager.GetQuotaInfo(ctx, "ec2", quotaCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get quota info: %w", err)
	}

	// Calculate desired quota value (current + buffer for future growth)
	// For insufficient quota: increase by at least 50% or 32 vCPUs, whichever is greater
	// For proactive requests: increase by 50%
	currentValue := quotaInfo.Value
	minIncrease := 32.0
	percentIncrease := currentValue * 0.5

	var desiredValue float64
	if percentIncrease > minIncrease {
		desiredValue = currentValue + percentIncrease
	} else {
		desiredValue = currentValue + minIncrease
	}

	// Round to nearest power of 2 for cleaner values (32, 64, 128, 256, etc.)
	desiredValue = roundToNearestPowerOfTwo(desiredValue)

	request := &QuotaIncreaseRequest{
		ServiceCode:  "ec2",
		QuotaCode:    quotaCode,
		QuotaName:    quotaInfo.QuotaName,
		CurrentValue: currentValue,
		DesiredValue: desiredValue,
		InstanceType: instanceType,
		Region:       qrh.region,
		Reason:       qrh.generateReasonText(instanceType, validation),
	}

	// Generate URLs for quick access
	request.SupportCaseURL = qrh.generateSupportCaseURL(request)
	request.QuickRequestURL = qrh.generateQuickRequestURL(request)

	return request, nil
}

// GenerateGuidance provides step-by-step guidance for requesting a quota increase
func (qrh *QuotaRequestHelper) GenerateGuidance(request *QuotaIncreaseRequest) string {
	var builder strings.Builder

	builder.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	builder.WriteString("📋 AWS Quota Increase Request Guide\n")
	builder.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	builder.WriteString("🎯 Quota Details:\n")
	builder.WriteString(fmt.Sprintf("  • Service: %s\n", request.ServiceCode))
	builder.WriteString(fmt.Sprintf("  • Quota: %s\n", request.QuotaName))
	builder.WriteString(fmt.Sprintf("  • Region: %s\n", request.Region))
	builder.WriteString(fmt.Sprintf("  • Current Limit: %.0f vCPUs\n", request.CurrentValue))
	builder.WriteString(fmt.Sprintf("  • Recommended Increase: %.0f vCPUs\n", request.DesiredValue))
	if request.InstanceType != "" {
		builder.WriteString(fmt.Sprintf("  • Instance Type: %s\n", request.InstanceType))
	}
	builder.WriteString("\n")

	builder.WriteString("📝 Step-by-Step Process:\n\n")

	builder.WriteString("1️⃣  Open AWS Service Quotas Console:\n")
	builder.WriteString(fmt.Sprintf("   %s\n\n", request.QuickRequestURL))

	builder.WriteString("2️⃣  Click \"Request quota increase\" button\n\n")

	builder.WriteString("3️⃣  Fill in the request form:\n")
	builder.WriteString(fmt.Sprintf("   • New quota value: %.0f\n", request.DesiredValue))
	builder.WriteString("   • Use case description:\n")
	builder.WriteString(fmt.Sprintf("     %s\n\n", request.Reason))

	builder.WriteString("4️⃣  Submit the request\n")
	builder.WriteString("   • Most quota increases are approved within 24-48 hours\n")
	builder.WriteString("   • You'll receive email updates on the request status\n")
	builder.WriteString("   • For urgent requests, contact AWS Support directly\n\n")

	builder.WriteString("📞 Alternative: AWS Support Center:\n")
	builder.WriteString("   If you need faster approval or have questions:\n")
	builder.WriteString(fmt.Sprintf("   %s\n\n", request.SupportCaseURL))

	builder.WriteString("💡 Tips:\n")
	builder.WriteString("   • Be specific about your use case (research, ML training, etc.)\n")
	builder.WriteString("   • Mention if this is for academic/research purposes\n")
	builder.WriteString("   • Request slightly more than you need to account for growth\n")
	builder.WriteString("   • Most quota increases are automatically approved\n\n")

	builder.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	return builder.String()
}

// ExplainQuotaError provides a user-friendly explanation of a quota error
func (qrh *QuotaRequestHelper) ExplainQuotaError(ctx context.Context, err error, instanceType string) string {
	var builder strings.Builder

	builder.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	builder.WriteString("❌ Launch Failed: AWS Quota Limit Reached\n")
	builder.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	// Try to validate the instance to get specific quota info
	validation, validErr := qrh.manager.ValidateInstanceLaunch(ctx, instanceType)
	if validErr == nil {
		builder.WriteString("📊 Quota Status:\n")
		builder.WriteString(fmt.Sprintf("  • Instance Type: %s\n", instanceType))
		builder.WriteString(fmt.Sprintf("  • Quota: %s\n", validation.QuotaType))
		builder.WriteString(fmt.Sprintf("  • Current Usage: %d vCPUs (%.1f%%)\n", validation.CurrentUsage, validation.UsagePercent))
		builder.WriteString(fmt.Sprintf("  • Quota Limit: %d vCPUs\n", validation.QuotaLimit))
		builder.WriteString(fmt.Sprintf("  • Required: %d vCPUs\n", validation.RequiredCapacity))
		builder.WriteString(fmt.Sprintf("  • Available: %d vCPUs\n", validation.AvailableCapacity))
		builder.WriteString("\n")
	}

	builder.WriteString("🔍 What Happened:\n")
	builder.WriteString("  AWS has limits (quotas) on the number of resources you can use.\n")
	builder.WriteString("  Your account has reached its quota limit for EC2 instances.\n\n")

	builder.WriteString("✅ Solution:\n")
	builder.WriteString("  Request a quota increase from AWS. This is:\n")
	builder.WriteString("  • Free (no charge for quota increases)\n")
	builder.WriteString("  • Fast (usually approved within 24-48 hours)\n")
	builder.WriteString("  • Simple (automated approval for most requests)\n\n")

	builder.WriteString("🚀 Quick Fix:\n")
	builder.WriteString(fmt.Sprintf("  Run: prism admin quota request --instance-type %s\n\n", instanceType))

	builder.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	return builder.String()
}

// generateReasonText creates a compelling use case description for the quota request
func (qrh *QuotaRequestHelper) generateReasonText(instanceType string, validation *QuotaValidationResult) string {
	var builder strings.Builder

	builder.WriteString("Academic research workload requiring additional EC2 capacity. ")

	if instanceType != "" {
		vcpus := GetInstanceTypeVCPUs(instanceType)
		builder.WriteString(fmt.Sprintf("Need to launch %s instances (%d vCPUs each) for computational research. ", instanceType, vcpus))
	}

	if validation != nil {
		builder.WriteString(fmt.Sprintf("Current quota of %.0f vCPUs is insufficient for project requirements. ", float64(validation.QuotaLimit)))
	}

	builder.WriteString("This increase will support ongoing research activities and ensure sufficient capacity for future workloads.")

	return builder.String()
}

// generateSupportCaseURL creates a direct link to AWS Support Center
func (qrh *QuotaRequestHelper) generateSupportCaseURL(request *QuotaIncreaseRequest) string {
	baseURL := "https://console.aws.amazon.com/support/home#/case/create"
	params := url.Values{}
	params.Add("type", "service-limit-increase")
	params.Add("service", "ec2-instances")

	return fmt.Sprintf("%s?%s", baseURL, params.Encode())
}

// generateQuickRequestURL creates a direct link to Service Quotas console
func (qrh *QuotaRequestHelper) generateQuickRequestURL(request *QuotaIncreaseRequest) string {
	// Service Quotas console URL format:
	// https://console.aws.amazon.com/servicequotas/home/services/ec2/quotas/L-1216C47A
	baseURL := fmt.Sprintf("https://console.aws.amazon.com/servicequotas/home/services/%s/quotas/%s",
		request.ServiceCode,
		request.QuotaCode,
	)

	// Add region parameter
	params := url.Values{}
	params.Add("region", request.Region)

	return fmt.Sprintf("%s?%s", baseURL, params.Encode())
}

// roundToNearestPowerOfTwo rounds a value to the nearest power of 2
func roundToNearestPowerOfTwo(value float64) float64 {
	powers := []float64{8, 16, 32, 64, 128, 256, 512, 1024, 2048, 4096}

	for _, power := range powers {
		if value <= power {
			return power
		}
	}

	// If larger than all powers, return the value rounded up to nearest 1024
	return float64(int((value+1023)/1024) * 1024)
}
