package project

import (
	"testing"
	"time"

	"github.com/scttfrdmn/prism/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCostCalculator_GetInstanceHourlyRate tests retrieving hourly rates for instance types
func TestCostCalculator_GetInstanceHourlyRate(t *testing.T) {
	calc := &CostCalculator{}

	tests := []struct {
		name         string
		instanceType string
		expectedRate float64
	}{
		{
			name:         "t3.micro known instance",
			instanceType: "t3.micro",
			expectedRate: 0.0104,
		},
		{
			name:         "c5.xlarge known instance",
			instanceType: "c5.xlarge",
			expectedRate: 0.17,
		},
		{
			name:         "r5.2xlarge known instance",
			instanceType: "r5.2xlarge",
			expectedRate: 0.504,
		},
		{
			name:         "g4dn.xlarge GPU instance",
			instanceType: "g4dn.xlarge",
			expectedRate: 0.526,
		},
		{
			name:         "unknown instance type - estimated",
			instanceType: "m5.large",
			expectedRate: 0.384, // m5 base (0.096) * large multiplier (4.0)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rate := calc.GetInstanceHourlyRate(tt.instanceType)
			assert.InDelta(t, tt.expectedRate, rate, 0.01, "hourly rate should match expected")
		})
	}
}

// TestCostCalculator_GetStorageMonthlyRate tests retrieving storage rates
func TestCostCalculator_GetStorageMonthlyRate(t *testing.T) {
	calc := &CostCalculator{}

	tests := []struct {
		name         string
		storageType  string
		expectedRate float64
	}{
		{
			name:         "gp3 storage",
			storageType:  "gp3",
			expectedRate: 0.08,
		},
		{
			name:         "gp2 storage",
			storageType:  "gp2",
			expectedRate: 0.10,
		},
		{
			name:         "EFS standard",
			storageType:  "efs-standard",
			expectedRate: 0.30,
		},
		{
			name:         "io2 provisioned IOPS",
			storageType:  "io2",
			expectedRate: 0.125,
		},
		{
			name:         "unknown storage type defaults to gp3",
			storageType:  "unknown-type",
			expectedRate: 0.08, // gp3 default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rate := calc.GetStorageMonthlyRate(tt.storageType)
			assert.Equal(t, tt.expectedRate, rate)
		})
	}
}

// TestCostCalculator_EstimateMonthlyCost tests monthly cost estimation
func TestCostCalculator_EstimateMonthlyCost(t *testing.T) {
	calc := &CostCalculator{}

	tests := []struct {
		name          string
		instanceType  string
		storageGB     int
		expectedRange [2]float64 // min, max
	}{
		{
			name:          "t3.micro with 20GB",
			instanceType:  "t3.micro",
			storageGB:     20,
			expectedRange: [2]float64{9.0, 10.0}, // ~7.49 compute + 1.6 storage
		},
		{
			name:          "c5.xlarge with 50GB",
			instanceType:  "c5.xlarge",
			storageGB:     50,
			expectedRange: [2]float64{126.0, 128.0}, // ~122.4 compute + 4.0 storage
		},
		{
			name:          "r5.2xlarge with 100GB",
			instanceType:  "r5.2xlarge",
			storageGB:     100,
			expectedRange: [2]float64{370.0, 372.0}, // ~362.88 compute + 8.0 storage
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cost := calc.EstimateMonthlyCost(tt.instanceType, tt.storageGB)
			assert.GreaterOrEqual(t, cost, tt.expectedRange[0])
			assert.LessOrEqual(t, cost, tt.expectedRange[1])
		})
	}
}

// TestCostCalculator_EstimateHibernationSavings tests hibernation savings calculation
func TestCostCalculator_EstimateHibernationSavings(t *testing.T) {
	calc := &CostCalculator{}

	tests := []struct {
		name            string
		instanceType    string
		hibernatedHours float64
		expectedSavings float64
	}{
		{
			name:            "t3.micro hibernated 100 hours",
			instanceType:    "t3.micro",
			hibernatedHours: 100.0,
			expectedSavings: 0.936, // 0.0104 * 100 * 0.90
		},
		{
			name:            "c5.xlarge hibernated 24 hours",
			instanceType:    "c5.xlarge",
			hibernatedHours: 24.0,
			expectedSavings: 3.672, // 0.17 * 24 * 0.90
		},
		{
			name:            "no hibernation",
			instanceType:    "t3.small",
			hibernatedHours: 0.0,
			expectedSavings: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			savings := calc.EstimateHibernationSavings(tt.instanceType, tt.hibernatedHours)
			assert.InDelta(t, tt.expectedSavings, savings, 0.01)
		})
	}
}

// TestCostCalculator_estimateInstanceCost tests cost estimation for unknown instance types
func TestCostCalculator_estimateInstanceCost(t *testing.T) {
	calc := &CostCalculator{}

	tests := []struct {
		name          string
		instanceType  string
		expectedRange [2]float64 // min, max for estimated cost
	}{
		{
			name:          "m5.large - known family",
			instanceType:  "m5.large",
			expectedRange: [2]float64{0.38, 0.40}, // 0.096 * 4.0
		},
		{
			name:          "m5a.xlarge - known family and size",
			instanceType:  "m5a.xlarge",
			expectedRange: [2]float64{0.68, 0.70}, // 0.086 * 8.0
		},
		{
			name:          "c5n.2xlarge - unknown family uses fallback",
			instanceType:  "c5n.2xlarge",
			expectedRange: [2]float64{1.7, 1.8}, // 0.108 * 16.0
		},
		{
			name:          "invalid format - uses default",
			instanceType:  "invalid-format",
			expectedRange: [2]float64{0.09, 0.11}, // default fallback
		},
		{
			name:          "t3.nano - small size",
			instanceType:  "t3.nano",
			expectedRange: [2]float64{0.002, 0.003}, // 0.0104 * 0.25
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cost := calc.estimateInstanceCost(tt.instanceType)
			assert.GreaterOrEqual(t, cost, tt.expectedRange[0])
			assert.LessOrEqual(t, cost, tt.expectedRange[1])
		})
	}
}

// TestCostCalculator_estimateRootVolumeSize tests root volume size estimation
func TestCostCalculator_estimateRootVolumeSize(t *testing.T) {
	calc := &CostCalculator{}

	tests := []struct {
		name         string
		instanceType string
		expectedSize float64
	}{
		{
			name:         "standard instance",
			instanceType: "t3.medium",
			expectedSize: 20.0,
		},
		{
			name:         "GPU g4dn instance",
			instanceType: "g4dn.xlarge",
			expectedSize: 50.0,
		},
		{
			name:         "GPU p3 instance",
			instanceType: "p3.2xlarge",
			expectedSize: 50.0,
		},
		{
			name:         "GPU p4d instance",
			instanceType: "p4d.24xlarge",
			expectedSize: 50.0,
		},
		{
			name:         "compute optimized instance",
			instanceType: "c5.xlarge",
			expectedSize: 20.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			size := calc.estimateRootVolumeSize(tt.instanceType)
			assert.Equal(t, tt.expectedSize, size)
		})
	}
}

// TestCostCalculator_calculateSingleInstanceCost tests single instance cost calculation
func TestCostCalculator_calculateSingleInstanceCost(t *testing.T) {
	calc := &CostCalculator{}

	now := time.Now()
	launchTime := now.Add(-24 * time.Hour) // Launched 24 hours ago

	tests := []struct {
		name     string
		instance types.Instance
	}{
		{
			name: "running instance",
			instance: types.Instance{
				Name:         "test-instance",
				InstanceType: "t3.medium",
				State:        "running",
				LaunchTime:   launchTime,
			},
		},
		{
			name: "stopped instance",
			instance: types.Instance{
				Name:         "stopped-instance",
				InstanceType: "c5.xlarge",
				State:        "stopped",
				LaunchTime:   launchTime,
			},
		},
		{
			name: "hibernated instance",
			instance: types.Instance{
				Name:         "hibernated-instance",
				InstanceType: "r5.large",
				State:        "hibernated",
				LaunchTime:   launchTime,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cost := calc.calculateSingleInstanceCost(tt.instance)

			assert.Equal(t, tt.instance.Name, cost.InstanceName)
			assert.Equal(t, tt.instance.InstanceType, cost.InstanceType)
			assert.GreaterOrEqual(t, cost.TotalCost, 0.0)
			assert.GreaterOrEqual(t, cost.ComputeCost, 0.0)
			assert.GreaterOrEqual(t, cost.StorageCost, 0.0)
			assert.InDelta(t, cost.ComputeCost+cost.StorageCost, cost.TotalCost, 0.0001)

			// Check that hours are calculated
			assert.Greater(t, cost.RunningHours, 0.0)

			// State-specific checks
			switch tt.instance.State {
			case "stopped":
				assert.Greater(t, cost.StoppedHours, 0.0)
			case "hibernated":
				assert.Greater(t, cost.HibernatedHours, 0.0)
			}
		})
	}
}

// TestCostCalculator_CalculateInstanceCosts tests batch instance cost calculation
func TestCostCalculator_CalculateInstanceCosts(t *testing.T) {
	calc := &CostCalculator{}

	now := time.Now()
	instances := []types.Instance{
		{
			Name:         "instance-1",
			InstanceType: "t3.micro",
			State:        "running",
			LaunchTime:   now.Add(-12 * time.Hour),
		},
		{
			Name:         "instance-2",
			InstanceType: "t3.small",
			State:        "running",
			LaunchTime:   now.Add(-24 * time.Hour),
		},
		{
			Name:         "instance-3",
			InstanceType: "c5.large",
			State:        "stopped",
			LaunchTime:   now.Add(-48 * time.Hour),
		},
	}

	costs, totalCost := calc.CalculateInstanceCosts(instances)

	// Verify results
	assert.Len(t, costs, 3)
	assert.Greater(t, totalCost, 0.0)

	// Verify each cost is accounted for
	var summedTotal float64
	for _, cost := range costs {
		assert.Greater(t, cost.TotalCost, 0.0)
		summedTotal += cost.TotalCost
	}
	assert.Equal(t, totalCost, summedTotal)

	// Verify instance names match
	assert.Equal(t, "instance-1", costs[0].InstanceName)
	assert.Equal(t, "instance-2", costs[1].InstanceName)
	assert.Equal(t, "instance-3", costs[2].InstanceName)
}

// TestCostCalculator_CalculateInstanceCosts_EmptyList tests empty instance list
func TestCostCalculator_CalculateInstanceCosts_EmptyList(t *testing.T) {
	calc := &CostCalculator{}

	costs, totalCost := calc.CalculateInstanceCosts([]types.Instance{})

	assert.Empty(t, costs)
	assert.Equal(t, 0.0, totalCost)
}

// TestCostCalculator_calculateInstanceStorageCost tests instance storage cost calculation
func TestCostCalculator_calculateInstanceStorageCost(t *testing.T) {
	calc := &CostCalculator{}

	now := time.Now()
	tests := []struct {
		name         string
		instance     types.Instance
		expectedCost float64 // Approximate expected cost
	}{
		{
			name: "standard instance - 1 day old",
			instance: types.Instance{
				InstanceType: "t3.medium",
				LaunchTime:   now.Add(-24 * time.Hour),
			},
			expectedCost: 0.053, // 20GB * 0.08/GB/month * (1/30)
		},
		{
			name: "GPU instance - 1 day old",
			instance: types.Instance{
				InstanceType: "g4dn.xlarge",
				LaunchTime:   now.Add(-24 * time.Hour),
			},
			expectedCost: 0.133, // 50GB * 0.08/GB/month * (1/30)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cost := calc.calculateInstanceStorageCost(tt.instance)
			assert.Greater(t, cost, 0.0)
			assert.InDelta(t, tt.expectedCost, cost, 0.02) // Allow 2 cent delta
		})
	}
}

// TestCostCalculator_calculateStorageVolumeCost tests storage volume cost calculation
func TestCostCalculator_calculateStorageVolumeCost(t *testing.T) {
	calc := &CostCalculator{}

	now := time.Now()
	creationTime := now.Add(-30 * 24 * time.Hour) // 30 days ago

	// Helper function to create pointer
	sizePtr := func(size int32) *int32 { return &size }
	sizeBytesPtr := func(size int64) *int64 { return &size }

	tests := []struct {
		name   string
		volume types.StorageVolume
	}{
		{
			name: "EFS shared storage",
			volume: types.StorageVolume{
				Name:         "shared-efs",
				AWSService:   types.AWSServiceEFS,
				SizeBytes:    sizeBytesPtr(10 * 1024 * 1024 * 1024), // 10GB
				CreationTime: creationTime,
			},
		},
		{
			name: "EBS workspace storage",
			volume: types.StorageVolume{
				Name:         "workspace-ebs",
				AWSService:   types.AWSServiceEBS,
				VolumeType:   "gp3",
				SizeGB:       sizePtr(100),
				CreationTime: creationTime,
			},
		},
		{
			name: "EBS with io2 type",
			volume: types.StorageVolume{
				Name:         "fast-ebs",
				AWSService:   types.AWSServiceEBS,
				VolumeType:   "io2",
				SizeGB:       sizePtr(50),
				CreationTime: creationTime,
			},
		},
		{
			name: "S3 storage",
			volume: types.StorageVolume{
				Name:         "s3-bucket",
				AWSService:   types.AWSServiceS3,
				CreationTime: creationTime,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cost := calc.calculateStorageVolumeCost(tt.volume)

			assert.Equal(t, tt.volume.Name, cost.VolumeName)
			assert.NotEmpty(t, cost.VolumeType)
			assert.GreaterOrEqual(t, cost.SizeGB, 0.0)
			assert.GreaterOrEqual(t, cost.Cost, 0.0)
			assert.Greater(t, cost.CostPerGB, 0.0)

			// Verify volume type is set correctly
			if tt.volume.IsShared() {
				assert.Equal(t, "EFS", cost.VolumeType)
			} else if tt.volume.IsWorkspace() {
				assert.Equal(t, tt.volume.VolumeType, cost.VolumeType)
			}
		})
	}
}

// TestCostCalculator_CalculateStorageCosts tests batch storage cost calculation
func TestCostCalculator_CalculateStorageCosts(t *testing.T) {
	calc := &CostCalculator{}

	now := time.Now()
	creationTime := now.Add(-30 * 24 * time.Hour)

	sizePtr := func(size int32) *int32 { return &size }

	volumes := []types.StorageVolume{
		{
			Name:         "efs-volume",
			AWSService:   types.AWSServiceEFS,
			CreationTime: creationTime,
		},
		{
			Name:         "ebs-volume-1",
			AWSService:   types.AWSServiceEBS,
			VolumeType:   "gp3",
			SizeGB:       sizePtr(50),
			CreationTime: creationTime,
		},
		{
			Name:         "ebs-volume-2",
			AWSService:   types.AWSServiceEBS,
			VolumeType:   "io2",
			SizeGB:       sizePtr(100),
			CreationTime: creationTime,
		},
	}

	costs, totalCost := calc.CalculateStorageCosts(volumes)

	// Verify results
	require.Len(t, costs, 3)
	assert.Greater(t, totalCost, 0.0)

	// Verify total cost is sum of individual costs
	var summedTotal float64
	for _, cost := range costs {
		assert.GreaterOrEqual(t, cost.Cost, 0.0)
		summedTotal += cost.Cost
	}
	assert.Equal(t, totalCost, summedTotal)

	// Verify volume names match
	assert.Equal(t, "efs-volume", costs[0].VolumeName)
	assert.Equal(t, "ebs-volume-1", costs[1].VolumeName)
	assert.Equal(t, "ebs-volume-2", costs[2].VolumeName)
}

// TestCostCalculator_CalculateStorageCosts_EmptyList tests empty storage list
func TestCostCalculator_CalculateStorageCosts_EmptyList(t *testing.T) {
	calc := &CostCalculator{}

	costs, totalCost := calc.CalculateStorageCosts([]types.StorageVolume{})

	assert.Empty(t, costs)
	assert.Equal(t, 0.0, totalCost)
}
