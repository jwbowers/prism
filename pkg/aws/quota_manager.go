package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/servicequotas"
)

// QuotaManager handles AWS quota awareness and validation
type QuotaManager struct {
	ec2Client *ec2.Client
	sqClient  *servicequotas.Client
	region    string

	// Cache for quota values (refreshed periodically)
	quotaCache     map[string]*QuotaCacheEntry
	lastCacheCheck time.Time
	cacheTTL       time.Duration
}

// QuotaCacheEntry holds cached quota information
type QuotaCacheEntry struct {
	Value      float64
	LastUpdate time.Time
}

// QuotaInfo provides detailed quota information
type QuotaInfo struct {
	ServiceCode  string
	QuotaCode    string
	QuotaName    string
	Value        float64
	UsageValue   float64
	UsagePercent float64
	IsAdjustable bool
	Unit         string
}

// QuotaValidationResult provides detailed validation results
type QuotaValidationResult struct {
	IsValid           bool
	QuotaType         string
	CurrentUsage      int
	QuotaLimit        int
	RequiredCapacity  int
	AvailableCapacity int
	UsagePercent      float64
	Warning           string
	SuggestedAction   string
}

const (
	// Service Quotas quota codes for EC2
	QuotaCodeOnDemandStandard = "L-1216C47A" // Running On-Demand Standard (A, C, D, H, I, M, R, T, Z) instances
	QuotaCodeOnDemandF        = "L-74FC7D96" // Running On-Demand F instances
	QuotaCodeOnDemandG        = "L-DB2E81BA" // Running On-Demand G and VT instances
	QuotaCodeOnDemandInf      = "L-1945791B" // Running On-Demand Inf instances
	QuotaCodeOnDemandP        = "L-417A185B" // Running On-Demand P instances
	QuotaCodeOnDemandX        = "L-7295265B" // Running On-Demand X instances
	QuotaCodeSpotStandard     = "L-34B43A08" // All Standard (A, C, D, H, I, M, R, T, Z) Spot Instance Requests
	QuotaCodeSpotG            = "L-3819A6DF" // All G and VT Spot Instance Requests
	QuotaCodeSpotP            = "L-7212CCBC" // All P Spot Instance Requests
	QuotaCodeEBSVolumes       = "L-6B0D7BF1" // Number of EBS volumes
	QuotaCodeEBSStorage       = "L-D18FCD1D" // Storage for General Purpose SSD (gp3) volumes, in TiB
	QuotaCodeEIPAddresses     = "L-0263D0A3" // EC2-VPC Elastic IPs
)

// NewQuotaManager creates a new quota manager
func NewQuotaManager(cfg aws.Config, region string) *QuotaManager {
	return &QuotaManager{
		ec2Client:  ec2.NewFromConfig(cfg),
		sqClient:   servicequotas.NewFromConfig(cfg),
		region:     region,
		quotaCache: make(map[string]*QuotaCacheEntry),
		cacheTTL:   15 * time.Minute, // Cache quotas for 15 minutes
	}
}

// ValidateInstanceLaunch performs pre-launch quota validation
func (qm *QuotaManager) ValidateInstanceLaunch(ctx context.Context, instanceType string) (*QuotaValidationResult, error) {
	// Determine which quota code to use based on instance family
	quotaCode := qm.getQuotaCodeForInstanceType(instanceType)

	// Get vCPU requirements for the instance
	requiredVCPUs := GetInstanceTypeVCPUs(instanceType)

	// Get current vCPU usage
	currentUsage, err := qm.getCurrentVCPUUsageByFamily(ctx, quotaCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get current vCPU usage: %w", err)
	}

	// Get quota limit
	quotaLimit, err := qm.getQuotaLimit(ctx, "ec2", quotaCode)
	if err != nil {
		// If we can't get quota, use conservative defaults
		quotaLimit = qm.getDefaultQuotaForCode(quotaCode)
	}

	// Calculate available capacity
	availableCapacity := int(quotaLimit) - currentUsage
	usagePercent := (float64(currentUsage) / quotaLimit) * 100

	result := &QuotaValidationResult{
		IsValid:           availableCapacity >= requiredVCPUs,
		QuotaType:         qm.getQuotaNameForCode(quotaCode),
		CurrentUsage:      currentUsage,
		QuotaLimit:        int(quotaLimit),
		RequiredCapacity:  requiredVCPUs,
		AvailableCapacity: availableCapacity,
		UsagePercent:      usagePercent,
	}

	// Generate warnings and suggestions
	if !result.IsValid {
		deficit := requiredVCPUs - availableCapacity
		result.Warning = fmt.Sprintf(
			"Insufficient quota for %s. Required: %d vCPUs, Available: %d vCPUs, Shortfall: %d vCPUs",
			instanceType, requiredVCPUs, availableCapacity, deficit,
		)
		result.SuggestedAction = fmt.Sprintf(
			"Request a quota increase for '%s' in region %s. Use: prism admin quota request --instance-type %s",
			result.QuotaType, qm.region, instanceType,
		)
	} else if usagePercent >= 90 {
		// Proactive warning at 90% usage
		result.Warning = fmt.Sprintf(
			"High quota usage: %.1f%% of %s quota is in use. Consider requesting an increase to avoid future launch failures.",
			usagePercent, result.QuotaType,
		)
		result.SuggestedAction = fmt.Sprintf(
			"Request a proactive quota increase: prism admin quota request --instance-type %s",
			instanceType,
		)
	}

	return result, nil
}

// GetQuotaInfo retrieves detailed information about a specific quota
func (qm *QuotaManager) GetQuotaInfo(ctx context.Context, serviceCode, quotaCode string) (*QuotaInfo, error) {
	input := &servicequotas.GetServiceQuotaInput{
		ServiceCode: aws.String(serviceCode),
		QuotaCode:   aws.String(quotaCode),
	}

	output, err := qm.sqClient.GetServiceQuota(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get service quota: %w", err)
	}

	if output.Quota == nil {
		return nil, fmt.Errorf("quota not found")
	}

	quota := output.Quota

	info := &QuotaInfo{
		ServiceCode:  aws.ToString(quota.ServiceCode),
		QuotaCode:    aws.ToString(quota.QuotaCode),
		QuotaName:    aws.ToString(quota.QuotaName),
		Value:        aws.ToFloat64(quota.Value),
		IsAdjustable: quota.Adjustable, // Already a bool, not *bool
		Unit:         aws.ToString(quota.Unit),
	}

	// Get current usage if available
	if quotaCode == QuotaCodeOnDemandStandard || quotaCode == QuotaCodeOnDemandG || quotaCode == QuotaCodeOnDemandP {
		usage, err := qm.getCurrentVCPUUsageByFamily(ctx, quotaCode)
		if err == nil {
			info.UsageValue = float64(usage)
			info.UsagePercent = (info.UsageValue / info.Value) * 100
		}
	}

	return info, nil
}

// ListRelevantQuotas returns a list of EC2 quotas relevant to Prism users
func (qm *QuotaManager) ListRelevantQuotas(ctx context.Context) ([]*QuotaInfo, error) {
	relevantQuotas := []string{
		QuotaCodeOnDemandStandard,
		QuotaCodeOnDemandG,
		QuotaCodeOnDemandP,
		QuotaCodeEBSVolumes,
		QuotaCodeEBSStorage,
		QuotaCodeEIPAddresses,
	}

	var quotas []*QuotaInfo
	for _, quotaCode := range relevantQuotas {
		info, err := qm.GetQuotaInfo(ctx, "ec2", quotaCode)
		if err != nil {
			// Skip quotas we can't retrieve (may not be available in all regions)
			continue
		}
		quotas = append(quotas, info)
	}

	return quotas, nil
}

// getQuotaCodeForInstanceType determines the appropriate quota code for an instance type
func (qm *QuotaManager) getQuotaCodeForInstanceType(instanceType string) string {
	// Instance type family is the first character(s) before the number
	// Examples: t3 -> t, m5 -> m, g4dn -> g, p3 -> p

	if len(instanceType) == 0 {
		return QuotaCodeOnDemandStandard
	}

	// GPU instances
	if instanceType[0] == 'g' || instanceType[0] == 'v' {
		return QuotaCodeOnDemandG
	}

	// P family (GPU compute)
	if instanceType[0] == 'p' {
		return QuotaCodeOnDemandP
	}

	// F family (FPGA)
	if instanceType[0] == 'f' {
		return QuotaCodeOnDemandF
	}

	// X family (memory optimized)
	if instanceType[0] == 'x' {
		return QuotaCodeOnDemandX
	}

	// Inf family (Inferentia)
	if len(instanceType) >= 3 && instanceType[:3] == "inf" {
		return QuotaCodeOnDemandInf
	}

	// Default: Standard instances (A, C, D, H, I, M, R, T, Z)
	return QuotaCodeOnDemandStandard
}

// getQuotaNameForCode returns a human-readable name for a quota code
func (qm *QuotaManager) getQuotaNameForCode(quotaCode string) string {
	names := map[string]string{
		QuotaCodeOnDemandStandard: "Running On-Demand Standard Instances",
		QuotaCodeOnDemandF:        "Running On-Demand F Instances",
		QuotaCodeOnDemandG:        "Running On-Demand G and VT Instances",
		QuotaCodeOnDemandInf:      "Running On-Demand Inf Instances",
		QuotaCodeOnDemandP:        "Running On-Demand P Instances",
		QuotaCodeOnDemandX:        "Running On-Demand X Instances",
	}

	if name, exists := names[quotaCode]; exists {
		return name
	}
	return "Unknown Quota"
}

// getDefaultQuotaForCode returns default quota values when API is unavailable
func (qm *QuotaManager) getDefaultQuotaForCode(quotaCode string) float64 {
	defaults := map[string]float64{
		QuotaCodeOnDemandStandard: 64, // 64 vCPUs for standard instances
		QuotaCodeOnDemandF:        8,  // 8 vCPUs for F instances
		QuotaCodeOnDemandG:        32, // 32 vCPUs for GPU instances
		QuotaCodeOnDemandInf:      8,  // 8 vCPUs for Inferentia instances
		QuotaCodeOnDemandP:        32, // 32 vCPUs for P instances
		QuotaCodeOnDemandX:        64, // 64 vCPUs for X instances
	}

	if defaultValue, exists := defaults[quotaCode]; exists {
		return defaultValue
	}
	return 64 // Conservative default
}

// getCurrentVCPUUsageByFamily calculates vCPU usage for a specific instance family
func (qm *QuotaManager) getCurrentVCPUUsageByFamily(ctx context.Context, quotaCode string) (int, error) {
	input := &ec2.DescribeInstancesInput{
		Filters: []ec2types.Filter{
			{
				Name:   aws.String("instance-state-name"),
				Values: []string{"running", "pending"},
			},
		},
	}

	output, err := qm.ec2Client.DescribeInstances(ctx, input)
	if err != nil {
		return 0, fmt.Errorf("failed to describe instances: %w", err)
	}

	totalVCPUs := 0
	for _, reservation := range output.Reservations {
		for _, instance := range reservation.Instances {
			instanceType := string(instance.InstanceType)

			// Only count instances that belong to this quota family
			if qm.getQuotaCodeForInstanceType(instanceType) == quotaCode {
				vcpus := GetInstanceTypeVCPUs(instanceType)
				totalVCPUs += vcpus
			}
		}
	}

	return totalVCPUs, nil
}

// getQuotaLimit retrieves the quota limit with caching
func (qm *QuotaManager) getQuotaLimit(ctx context.Context, serviceCode, quotaCode string) (float64, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("%s:%s", serviceCode, quotaCode)

	if entry, exists := qm.quotaCache[cacheKey]; exists {
		if time.Since(entry.LastUpdate) < qm.cacheTTL {
			return entry.Value, nil
		}
	}

	// Cache miss or expired, fetch from API
	input := &servicequotas.GetServiceQuotaInput{
		ServiceCode: aws.String(serviceCode),
		QuotaCode:   aws.String(quotaCode),
	}

	output, err := qm.sqClient.GetServiceQuota(ctx, input)
	if err != nil {
		return 0, fmt.Errorf("failed to get service quota: %w", err)
	}

	if output.Quota == nil || output.Quota.Value == nil {
		return 0, fmt.Errorf("quota value not available")
	}

	value := aws.ToFloat64(output.Quota.Value)

	// Update cache
	qm.quotaCache[cacheKey] = &QuotaCacheEntry{
		Value:      value,
		LastUpdate: time.Now(),
	}

	return value, nil
}

// ClearCache clears the quota cache (useful for testing or after quota changes)
func (qm *QuotaManager) ClearCache() {
	qm.quotaCache = make(map[string]*QuotaCacheEntry)
	qm.lastCacheCheck = time.Time{}
}
