package project

import (
	"strings"
	"time"

	"github.com/scttfrdmn/prism/pkg/types"
)

// CostCalculator calculates AWS costs for instances and storage
type CostCalculator struct {
	// AWS pricing data - in a production system, this would be loaded from AWS Pricing API
	// For now, we use estimated rates based on common instance types and regions
}

// Instance pricing data (USD per hour) - estimated rates for us-east-1
var instancePricing = map[string]float64{
	// General Purpose
	"t3.micro":    0.0104,
	"t3.small":    0.0208,
	"t3.medium":   0.0416,
	"t3.large":    0.0832,
	"t3.xlarge":   0.1664,
	"t3.2xlarge":  0.3328,
	"t3a.micro":   0.0094,
	"t3a.small":   0.0188,
	"t3a.medium":  0.0376,
	"t3a.large":   0.0752,
	"t3a.xlarge":  0.1504,
	"t3a.2xlarge": 0.3008,

	// Compute Optimized
	"c5.large":    0.085,
	"c5.xlarge":   0.17,
	"c5.2xlarge":  0.34,
	"c5.4xlarge":  0.68,
	"c5.9xlarge":  1.53,
	"c5.12xlarge": 2.04,
	"c5.18xlarge": 3.06,
	"c5.24xlarge": 4.08,

	// Memory Optimized
	"r5.large":    0.126,
	"r5.xlarge":   0.252,
	"r5.2xlarge":  0.504,
	"r5.4xlarge":  1.008,
	"r5.8xlarge":  2.016,
	"r5.12xlarge": 3.024,
	"r5.16xlarge": 4.032,
	"r5.24xlarge": 6.048,

	// GPU Instances
	"g4dn.xlarge":   0.526,
	"g4dn.2xlarge":  0.752,
	"g4dn.4xlarge":  1.204,
	"g4dn.8xlarge":  2.176,
	"g4dn.12xlarge": 3.912,
	"g4dn.16xlarge": 4.352,
	"p3.2xlarge":    3.06,
	"p3.8xlarge":    12.24,
	"p3.16xlarge":   24.48,
	"p4d.24xlarge":  32.77,
}

// Storage pricing (USD per GB per month)
var storagePricing = map[string]float64{
	"gp3":          0.08,  // General Purpose SSD (gp3)
	"gp2":          0.10,  // General Purpose SSD (gp2)
	"io1":          0.125, // Provisioned IOPS SSD (io1)
	"io2":          0.125, // Provisioned IOPS SSD (io2)
	"st1":          0.045, // Throughput Optimized HDD
	"sc1":          0.025, // Cold HDD
	"standard":     0.05,  // Magnetic
	"efs-standard": 0.30,  // EFS Standard (us-east-1)
	"efs-ia":       0.016, // EFS Infrequent Access (updated from 0.0125)
}

// EBS provisioned-IOPS pricing (USD per IOPS per month) — io1/io2 only
const (
	ebsIO1IOPSRate       = 0.065 // $/provisioned IOPS/month
	ebsIO2IOPSRate       = 0.065 // $/provisioned IOPS/month
	ebsGP3ThroughputRate = 0.04  // $/MB/s above 125 MB/s baseline
	ebsGP3BaseIOPS       = 3000  // included IOPS for gp3
	ebsGP3BaseThroughput = 125   // included MB/s for gp3
)

// EFS additional pricing
const (
	efsProvisionedThroughputRate = 6.00 // $/MB/s/month (provisioned mode)
	efsElasticReadRate           = 0.03 // $/GB transferred (elastic mode reads)
	efsElasticWriteRate          = 0.06 // $/GB transferred (elastic mode writes)
)

// S3 storage pricing (USD per GB per month)
var s3StoragePricing = map[string]float64{
	"STANDARD":            0.023,   // S3 Standard
	"STANDARD_IA":         0.0125,  // S3 Standard-IA
	"ONEZONE_IA":          0.01,    // S3 One Zone-IA
	"INTELLIGENT_TIERING": 0.023,   // S3 Intelligent-Tiering (frequent access)
	"GLACIER":             0.004,   // S3 Glacier Instant Retrieval
	"GLACIER_IR":          0.004,   // S3 Glacier Instant Retrieval (alias)
	"DEEP_ARCHIVE":        0.00099, // S3 Glacier Deep Archive
}

// S3 request pricing (USD per 1000 requests)
const (
	s3PUTRate      = 0.005  // PUT, COPY, POST, LIST per 1000
	s3GETRate      = 0.0004 // GET, SELECT per 1000
	s3EgressFirst  = 0.09   // $/GB — first 10 TB/month
	s3EgressNext   = 0.085  // $/GB — next 40 TB/month
	s3EgressCutoff = 10240  // GB — boundary between tiers
)

// CalculateInstanceCosts calculates costs for a list of instances
func (c *CostCalculator) CalculateInstanceCosts(instances []types.Instance) ([]types.InstanceCost, float64) {
	var instanceCosts []types.InstanceCost
	var totalCost float64

	for _, instance := range instances {
		cost := c.calculateSingleInstanceCost(instance)
		instanceCosts = append(instanceCosts, cost)
		totalCost += cost.TotalCost
	}

	return instanceCosts, totalCost
}

// CalculateStorageCosts calculates costs for all storage volumes
func (c *CostCalculator) CalculateStorageCosts(storageVolumes []types.StorageVolume) ([]types.StorageCost, float64) {
	var storageCosts []types.StorageCost
	var totalCost float64

	// Calculate costs for all storage volumes (EFS, EBS, S3)
	for _, volume := range storageVolumes {
		cost := c.calculateStorageVolumeCost(volume)
		storageCosts = append(storageCosts, cost)
		totalCost += cost.Cost
	}

	return storageCosts, totalCost
}

// calculateSingleInstanceCost calculates the cost for a single instance
func (c *CostCalculator) calculateSingleInstanceCost(instance types.Instance) types.InstanceCost {
	hourlyRate, exists := instancePricing[instance.InstanceType]
	if !exists {
		// Use a default rate for unknown instance types
		hourlyRate = c.estimateInstanceCost(instance.InstanceType)
	}

	// Calculate hours in different states
	now := time.Now()
	totalRuntime := now.Sub(instance.LaunchTime)
	totalHours := totalRuntime.Hours()

	// For simplicity, assume the instance has been running the entire time
	// In a real implementation, we would track state changes
	runningHours := totalHours
	hibernatedHours := 0.0
	stoppedHours := 0.0

	// Adjust based on current state
	switch strings.ToLower(instance.State) {
	case "stopped":
		// If stopped, it might have been running part of the time
		runningHours = totalHours * 0.7 // Estimate 70% uptime
		stoppedHours = totalHours * 0.3
	case "hibernated":
		runningHours = totalHours * 0.5 // Estimate 50% uptime before hibernation
		hibernatedHours = totalHours * 0.5
	}

	// Calculate compute cost (only for running hours)
	computeCost := runningHours * hourlyRate

	// Calculate storage cost (EBS root volume)
	storageCost := c.calculateInstanceStorageCost(instance)

	totalCost := computeCost + storageCost

	return types.InstanceCost{
		InstanceName:    instance.Name,
		InstanceType:    instance.InstanceType,
		ComputeCost:     computeCost,
		StorageCost:     storageCost,
		TotalCost:       totalCost,
		RunningHours:    runningHours,
		HibernatedHours: hibernatedHours,
		StoppedHours:    stoppedHours,
	}
}

// calculateInstanceStorageCost calculates the EBS storage cost for an instance
func (c *CostCalculator) calculateInstanceStorageCost(instance types.Instance) float64 {
	// Estimate root volume size based on instance type
	rootVolumeSize := c.estimateRootVolumeSize(instance.InstanceType)

	// Use gp3 pricing as default for root volumes
	pricePerGB := storagePricing["gp3"]

	// Calculate monthly cost, then pro-rate for actual runtime
	monthlyStorageCost := rootVolumeSize * pricePerGB

	// Calculate days since launch
	daysSinceLaunch := time.Since(instance.LaunchTime).Hours() / 24

	// Pro-rate the monthly cost
	return monthlyStorageCost * (daysSinceLaunch / 30.0)
}

// calculateStorageVolumeCost calculates the cost for a unified storage volume (EFS, EBS, or S3)
func (c *CostCalculator) calculateStorageVolumeCost(volume types.StorageVolume) types.StorageCost {
	var pricePerGB float64
	var sizeGB float64
	var volumeTypeStr string

	// Handle different storage types
	if volume.IsShared() {
		// EFS (Shared Storage)
		pricePerGB = storagePricing["efs-standard"]
		// EFS size is not directly available, estimate if needed
		if volume.SizeBytes != nil {
			sizeGB = float64(*volume.SizeBytes) / (1024 * 1024 * 1024)
		} else {
			sizeGB = 10.0 // Default estimate
		}
		volumeTypeStr = "EFS"
	} else if volume.IsWorkspace() {
		// EBS (Workspace Storage)
		pricePerGB = storagePricing["gp3"] // Default to gp3 pricing

		// Use volume type specific pricing if available
		if volume.VolumeType != "" {
			if price, exists := storagePricing[volume.VolumeType]; exists {
				pricePerGB = price
			}
		}

		if volume.SizeGB != nil {
			sizeGB = float64(*volume.SizeGB)
		} else {
			sizeGB = 0.0
		}
		volumeTypeStr = volume.VolumeType
	} else {
		// S3 or other cloud storage
		pricePerGB = storagePricing["gp3"] // Fallback pricing
		sizeGB = 0.0
		volumeTypeStr = string(volume.AWSService)
	}

	// Calculate monthly cost, then pro-rate for actual time
	monthlyStorageCost := sizeGB * pricePerGB
	daysSinceCreation := time.Since(volume.CreationTime).Hours() / 24
	actualCost := monthlyStorageCost * (daysSinceCreation / 30.0)

	return types.StorageCost{
		VolumeName: volume.Name,
		VolumeType: volumeTypeStr,
		SizeGB:     sizeGB,
		Cost:       actualCost,
		CostPerGB:  pricePerGB,
	}
}

// estimateInstanceCost estimates the hourly cost for unknown instance types
func (c *CostCalculator) estimateInstanceCost(instanceType string) float64 {
	// Extract instance family and size
	parts := strings.Split(instanceType, ".")
	if len(parts) != 2 {
		return 0.10 // Default fallback rate
	}

	family := parts[0]
	size := parts[1]

	// Base rates by instance family
	familyRates := map[string]float64{
		"t3":   0.0104, // t3.micro base rate
		"t3a":  0.0094, // t3a.micro base rate
		"c5":   0.085,  // c5.large base rate
		"c5n":  0.108,  // c5n.large base rate
		"r5":   0.126,  // r5.large base rate
		"r5a":  0.113,  // r5a.large base rate
		"m5":   0.096,  // m5.large base rate
		"m5a":  0.086,  // m5a.large base rate
		"g4dn": 0.526,  // g4dn.xlarge base rate
		"p3":   3.06,   // p3.2xlarge base rate
		"p4d":  32.77,  // p4d.24xlarge base rate
	}

	baseRate, exists := familyRates[family]
	if !exists {
		baseRate = 0.10 // Default rate
	}

	// Size multipliers
	sizeMultipliers := map[string]float64{
		"nano":     0.25,
		"micro":    0.5,
		"small":    1.0,
		"medium":   2.0,
		"large":    4.0,
		"xlarge":   8.0,
		"2xlarge":  16.0,
		"3xlarge":  24.0,
		"4xlarge":  32.0,
		"6xlarge":  48.0,
		"8xlarge":  64.0,
		"9xlarge":  72.0,
		"12xlarge": 96.0,
		"16xlarge": 128.0,
		"18xlarge": 144.0,
		"24xlarge": 192.0,
		"32xlarge": 256.0,
	}

	multiplier, exists := sizeMultipliers[size]
	if !exists {
		multiplier = 4.0 // Default to large equivalent
	}

	return baseRate * multiplier
}

// estimateRootVolumeSize estimates the root EBS volume size for an instance type
func (c *CostCalculator) estimateRootVolumeSize(instanceType string) float64 {
	// Most instances have 8-20 GB root volumes
	// GPU instances typically have larger root volumes
	if strings.Contains(instanceType, "g4") || strings.Contains(instanceType, "p3") || strings.Contains(instanceType, "p4") {
		return 50.0 // GPU instances often need larger root volumes for drivers
	}

	return 20.0 // Standard root volume size
}

// GetInstanceHourlyRate returns the hourly rate for an instance type
func (c *CostCalculator) GetInstanceHourlyRate(instanceType string) float64 {
	if rate, exists := instancePricing[instanceType]; exists {
		return rate
	}
	return c.estimateInstanceCost(instanceType)
}

// GetStorageMonthlyRate returns the monthly rate per GB for a storage type
func (c *CostCalculator) GetStorageMonthlyRate(storageType string) float64 {
	if rate, exists := storagePricing[storageType]; exists {
		return rate
	}
	return storagePricing["gp3"] // Default to gp3 pricing
}

// EstimateMonthlyCost estimates the monthly cost for running an instance continuously
func (c *CostCalculator) EstimateMonthlyCost(instanceType string, storageGB int) float64 {
	hourlyRate := c.GetInstanceHourlyRate(instanceType)
	storageRate := c.GetStorageMonthlyRate("gp3")

	// 24 hours * 30 days = 720 hours per month
	monthlyComputeCost := hourlyRate * 720
	monthlyStorageCost := float64(storageGB) * storageRate

	return monthlyComputeCost + monthlyStorageCost
}

// ─── EBS granular cost calculation ───────────────────────────────────────────

// EBSVolumeSpec contains configuration needed for accurate EBS pricing.
type EBSVolumeSpec struct {
	VolumeID   string
	Name       string
	SizeGB     int32
	VolumeType string // gp3, gp2, io1, io2, st1, sc1
	IOPS       int32  // provisioned IOPS (io1/io2 only)
	Throughput int32  // provisioned MB/s above baseline (gp3 only)
	Region     string
	AgeSeconds float64 // seconds since creation, for pro-rating
}

// EBSCostDetail is the output struct for a single EBS volume cost breakdown.
type EBSCostDetail struct {
	VolumeName       string  `json:"volume_name"`
	VolumeID         string  `json:"volume_id"`
	VolumeType       string  `json:"volume_type"`
	SizeGB           float64 `json:"size_gb"`
	StorageCost      float64 `json:"storage_cost"`       // base GB × monthly rate
	IOPSCost         float64 `json:"iops_cost"`          // io1/io2 provisioned IOPS
	ThroughputCost   float64 `json:"throughput_cost"`    // gp3 throughput above baseline
	TotalMonthlyCost float64 `json:"total_monthly_cost"` // pro-rated to this month
	AgeHours         float64 `json:"age_hours"`
	Region           string  `json:"region"`
}

// CalculateEBSCost returns a detailed cost breakdown for a single EBS volume.
func (c *CostCalculator) CalculateEBSCost(spec EBSVolumeSpec) EBSCostDetail {
	gbRate := storagePricing[spec.VolumeType]
	if gbRate == 0 {
		gbRate = storagePricing["gp3"]
	}
	storageCost := float64(spec.SizeGB) * gbRate

	var iopsCost float64
	switch spec.VolumeType {
	case "io1":
		iopsCost = float64(spec.IOPS) * ebsIO1IOPSRate
	case "io2":
		iopsCost = float64(spec.IOPS) * ebsIO2IOPSRate
	}

	var throughputCost float64
	if spec.VolumeType == "gp3" && spec.Throughput > ebsGP3BaseThroughput {
		overBaseline := float64(spec.Throughput - ebsGP3BaseThroughput)
		throughputCost = overBaseline * ebsGP3ThroughputRate
	}

	total := storageCost + iopsCost + throughputCost

	// Pro-rate to current month if age is less than a full month.
	if spec.AgeSeconds > 0 {
		ageFraction := spec.AgeSeconds / (30 * 24 * 3600)
		if ageFraction < 1.0 {
			total *= ageFraction
		}
	}

	return EBSCostDetail{
		VolumeName:       spec.Name,
		VolumeID:         spec.VolumeID,
		VolumeType:       spec.VolumeType,
		SizeGB:           float64(spec.SizeGB),
		StorageCost:      storageCost,
		IOPSCost:         iopsCost,
		ThroughputCost:   throughputCost,
		TotalMonthlyCost: total,
		AgeHours:         spec.AgeSeconds / 3600,
		Region:           spec.Region,
	}
}

// ─── EFS granular cost calculation ───────────────────────────────────────────

// EFSVolumeSpec contains configuration needed for accurate EFS pricing.
type EFSVolumeSpec struct {
	FilesystemID      string
	Name              string
	SizeBytesStandard int64   // bytes in Standard storage class
	SizeBytesIA       int64   // bytes in Infrequent Access storage class
	ThroughputMode    string  // "bursting" | "provisioned" | "elastic"
	ProvisionedMBps   float64 // only for provisioned mode
	ElasticReadGB     float64 // GB transferred (elastic mode reads this month)
	ElasticWriteGB    float64 // GB transferred (elastic mode writes this month)
	Region            string
	AgeSeconds        float64
}

// EFSCostDetail is the output struct for a single EFS filesystem cost breakdown.
type EFSCostDetail struct {
	FilesystemName        string  `json:"filesystem_name"`
	FilesystemID          string  `json:"filesystem_id"`
	StandardStorageGB     float64 `json:"standard_storage_gb"`
	IAStorageGB           float64 `json:"ia_storage_gb"`
	StandardStorageCost   float64 `json:"standard_storage_cost"`
	IAStorageCost         float64 `json:"ia_storage_cost"`
	ThroughputMode        string  `json:"throughput_mode"`
	ProvisionedThroughput float64 `json:"provisioned_throughput_mbps"`
	ThroughputCost        float64 `json:"throughput_cost"`
	TotalMonthlyCost      float64 `json:"total_monthly_cost"`
	Region                string  `json:"region"`
}

const bytesPerGB = 1024 * 1024 * 1024

// CalculateEFSCost returns a detailed cost breakdown for a single EFS filesystem.
func (c *CostCalculator) CalculateEFSCost(spec EFSVolumeSpec) EFSCostDetail {
	standardGB := float64(spec.SizeBytesStandard) / bytesPerGB
	iaGB := float64(spec.SizeBytesIA) / bytesPerGB

	standardCost := standardGB * storagePricing["efs-standard"]
	iaCost := iaGB * storagePricing["efs-ia"]

	var throughputCost float64
	switch spec.ThroughputMode {
	case "provisioned":
		throughputCost = spec.ProvisionedMBps * efsProvisionedThroughputRate
	case "elastic":
		throughputCost = spec.ElasticReadGB*efsElasticReadRate + spec.ElasticWriteGB*efsElasticWriteRate
		// bursting: no additional charge
	}

	total := standardCost + iaCost + throughputCost

	return EFSCostDetail{
		FilesystemName:        spec.Name,
		FilesystemID:          spec.FilesystemID,
		StandardStorageGB:     standardGB,
		IAStorageGB:           iaGB,
		StandardStorageCost:   standardCost,
		IAStorageCost:         iaCost,
		ThroughputMode:        spec.ThroughputMode,
		ProvisionedThroughput: spec.ProvisionedMBps,
		ThroughputCost:        throughputCost,
		TotalMonthlyCost:      total,
		Region:                spec.Region,
	}
}

// ─── S3 cost calculation ──────────────────────────────────────────────────────

// S3CostSpec contains S3 bucket usage data for cost calculation.
type S3CostSpec struct {
	BucketName            string
	Region                string
	StorageClassBreakdown map[string]int64 // storage_class -> bytes
	RequestCounts         map[string]int64 // "GET","PUT","LIST","DELETE" -> count
	DataTransferGB        float64          // data transferred OUT (egress)
	ReplicationGB         float64          // cross-region replication GB
}

// S3CostDetail is the output struct for a single S3 bucket cost breakdown.
type S3CostDetail struct {
	BucketName       string             `json:"bucket_name"`
	Region           string             `json:"region"`
	StorageCost      float64            `json:"storage_cost"`
	RequestCost      float64            `json:"request_cost"`
	EgressCost       float64            `json:"egress_cost"`
	ReplicationCost  float64            `json:"replication_cost"`
	TotalMonthlyCost float64            `json:"total_monthly_cost"`
	StorageBreakdown map[string]float64 `json:"storage_breakdown"` // class -> $
}

// CalculateS3Cost returns a detailed cost breakdown for a single S3 bucket.
func (c *CostCalculator) CalculateS3Cost(spec S3CostSpec) S3CostDetail {
	storageBreakdown := make(map[string]float64)
	var totalStorage float64

	for class, bytes := range spec.StorageClassBreakdown {
		gb := float64(bytes) / bytesPerGB
		rate, ok := s3StoragePricing[class]
		if !ok {
			rate = s3StoragePricing["STANDARD"]
		}
		cost := gb * rate
		storageBreakdown[class] = cost
		totalStorage += cost
	}

	// Request costs
	var requestCost float64
	for reqType, count := range spec.RequestCounts {
		thousands := float64(count) / 1000
		switch reqType {
		case "PUT", "COPY", "POST", "LIST":
			requestCost += thousands * s3PUTRate
		default: // GET, HEAD, SELECT, etc.
			requestCost += thousands * s3GETRate
		}
	}

	// Tiered egress pricing
	var egressCost float64
	if spec.DataTransferGB <= s3EgressCutoff {
		egressCost = spec.DataTransferGB * s3EgressFirst
	} else {
		egressCost = s3EgressCutoff*s3EgressFirst + (spec.DataTransferGB-s3EgressCutoff)*s3EgressNext
	}

	// Replication is charged as egress
	replicationCost := spec.ReplicationGB * s3EgressFirst

	total := totalStorage + requestCost + egressCost + replicationCost

	return S3CostDetail{
		BucketName:       spec.BucketName,
		Region:           spec.Region,
		StorageCost:      totalStorage,
		RequestCost:      requestCost,
		EgressCost:       egressCost,
		ReplicationCost:  replicationCost,
		TotalMonthlyCost: total,
		StorageBreakdown: storageBreakdown,
	}
}

// ─── Discount application ─────────────────────────────────────────────────────

// ApplyDiscount applies a DiscountConfig to a base cost for a given service type.
// serviceType should be one of: "ec2", "ebs", "efs", "s3".
func (c *CostCalculator) ApplyDiscount(baseCost float64, config *types.DiscountConfig, serviceType string) float64 {
	if config == nil {
		return baseCost
	}
	var discount float64
	switch serviceType {
	case "ec2":
		discount = config.EC2Discount
	case "ebs":
		discount = config.EBSDiscount
	case "efs":
		discount = config.EFSDiscount
	default:
		discount = config.EducationalDiscount
	}
	if discount <= 0 || discount >= 1.0 {
		return baseCost
	}
	return baseCost * (1.0 - discount)
}

// EstimateHibernationSavings estimates the cost savings from hibernating vs running
func (c *CostCalculator) EstimateHibernationSavings(instanceType string, hibernatedHours float64) float64 {
	hourlyRate := c.GetInstanceHourlyRate(instanceType)

	// Hibernation saves compute costs but storage costs continue
	// Assume hibernation saves 90% of compute costs (some overhead remains)
	return hourlyRate * hibernatedHours * 0.90
}
