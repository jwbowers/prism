package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// ZombieInstance represents an EC2 instance without Prism management tags
type ZombieInstance struct {
	InstanceID   string
	State        string
	InstanceType string
	LaunchTime   time.Time
	Name         string
	MonthlyCost  float64
}

// ZombieVolume represents an unattached EBS volume
type ZombieVolume struct {
	VolumeID    string
	SizeGB      int32
	CreateTime  time.Time
	MonthlyCost float64
}

// ZombieScanResult contains detected zombie resources
type ZombieScanResult struct {
	Instances           []ZombieInstance
	Volumes             []ZombieVolume
	PrismInstancesCount int
	TotalMonthlyCost    float64
	Region              string
	ScanTime            time.Time
}

// ScanZombieResources detects AWS resources without Prism management tags
func (m *Manager) ScanZombieResources(ctx context.Context) (*ZombieScanResult, error) {
	result := &ZombieScanResult{
		Instances: []ZombieInstance{},
		Volumes:   []ZombieVolume{},
		Region:    m.region,
		ScanTime:  time.Now(),
	}

	// Scan EC2 instances
	instances, prismCount, err := m.scanZombieInstances(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to scan instances: %w", err)
	}
	result.Instances = instances
	result.PrismInstancesCount = prismCount

	// Scan unattached EBS volumes
	volumes, err := m.scanZombieVolumes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to scan volumes: %w", err)
	}
	result.Volumes = volumes

	// Calculate total cost
	for _, instance := range result.Instances {
		result.TotalMonthlyCost += instance.MonthlyCost
	}
	for _, volume := range result.Volumes {
		result.TotalMonthlyCost += volume.MonthlyCost
	}

	return result, nil
}

// scanZombieInstances finds EC2 instances without Prism tags
func (m *Manager) scanZombieInstances(ctx context.Context) ([]ZombieInstance, int, error) {
	// Find all running/stopped/pending instances
	input := &ec2.DescribeInstancesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("instance-state-name"),
				Values: []string{"running", "stopped", "pending"},
			},
		},
	}

	output, err := m.ec2.DescribeInstances(ctx, input)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to describe instances: %w", err)
	}

	var zombies []ZombieInstance
	prismCount := 0

	for _, reservation := range output.Reservations {
		for _, instance := range reservation.Instances {
			if m.isPrismManaged(instance.Tags) {
				prismCount++
				continue
			}

			// This is a zombie instance
			zombie := ZombieInstance{
				InstanceID:   aws.ToString(instance.InstanceId),
				State:        string(instance.State.Name),
				InstanceType: string(instance.InstanceType),
				LaunchTime:   aws.ToTime(instance.LaunchTime),
			}

			// Extract name from tags
			for _, tag := range instance.Tags {
				if aws.ToString(tag.Key) == "Name" {
					zombie.Name = aws.ToString(tag.Value)
					break
				}
			}

			// Estimate monthly cost
			zombie.MonthlyCost = estimateInstanceMonthlyCost(string(instance.InstanceType))

			zombies = append(zombies, zombie)
		}
	}

	return zombies, prismCount, nil
}

// scanZombieVolumes finds unattached EBS volumes
func (m *Manager) scanZombieVolumes(ctx context.Context) ([]ZombieVolume, error) {
	input := &ec2.DescribeVolumesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("status"),
				Values: []string{"available"},
			},
		},
	}

	output, err := m.ec2.DescribeVolumes(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to describe volumes: %w", err)
	}

	var zombies []ZombieVolume
	for _, volume := range output.Volumes {
		zombie := ZombieVolume{
			VolumeID:   aws.ToString(volume.VolumeId),
			SizeGB:     aws.ToInt32(volume.Size),
			CreateTime: aws.ToTime(volume.CreateTime),
		}

		// EBS cost: $0.10/GB/month for gp3 (standard rate)
		zombie.MonthlyCost = float64(zombie.SizeGB) * 0.10

		zombies = append(zombies, zombie)
	}

	return zombies, nil
}

// isPrismManaged checks if an instance has Prism management tags
func (m *Manager) isPrismManaged(tags []types.Tag) bool {
	for _, tag := range tags {
		key := aws.ToString(tag.Key)
		value := aws.ToString(tag.Value)

		// Check for new namespaced tag (prism:managed)
		if key == "prism:managed" && value == "true" {
			return true
		}

		// Check for legacy tag (Prism)
		if key == "Prism" && value == "true" {
			return true
		}
	}
	return false
}

// estimateInstanceMonthlyCost provides rough monthly cost estimates for instance types
func estimateInstanceMonthlyCost(instanceType string) float64 {
	// Hourly rates (approximate) × 24 hours × 30 days
	hourlyRates := map[string]float64{
		// t3 family (general purpose)
		"t3.micro":   0.010,
		"t3.small":   0.021,
		"t3.medium":  0.042,
		"t3.large":   0.083,
		"t3.xlarge":  0.166,
		"t3.2xlarge": 0.333,

		// c7g family (compute optimized ARM)
		"c7g.medium":  0.036,
		"c7g.large":   0.072,
		"c7g.xlarge":  0.145,
		"c7g.2xlarge": 0.290,
		"c7g.4xlarge": 0.580,
		"c7g.8xlarge": 1.160,

		// m5 family (general purpose)
		"m5.large":   0.096,
		"m5.xlarge":  0.192,
		"m5.2xlarge": 0.384,
		"m5.4xlarge": 0.768,

		// r5 family (memory optimized)
		"r5.large":   0.126,
		"r5.xlarge":  0.252,
		"r5.2xlarge": 0.504,
		"r5.4xlarge": 1.008,
	}

	hourlyRate, exists := hourlyRates[instanceType]
	if !exists {
		hourlyRate = 0.10 // Default estimate
	}

	return hourlyRate * 24 * 30 // Convert to monthly
}

// CleanupZombieResources terminates zombie instances and deletes zombie volumes
func (m *Manager) CleanupZombieResources(ctx context.Context, instanceIDs []string, volumeIDs []string) error {
	// Terminate zombie instances
	if len(instanceIDs) > 0 {
		terminateInput := &ec2.TerminateInstancesInput{
			InstanceIds: instanceIDs,
		}
		_, err := m.ec2.TerminateInstances(ctx, terminateInput)
		if err != nil {
			return fmt.Errorf("failed to terminate instances: %w", err)
		}
	}

	// Delete zombie volumes
	for _, volumeID := range volumeIDs {
		deleteInput := &ec2.DeleteVolumeInput{
			VolumeId: aws.String(volumeID),
		}
		_, err := m.ec2.DeleteVolume(ctx, deleteInput)
		if err != nil {
			return fmt.Errorf("failed to delete volume %s: %w", volumeID, err)
		}
	}

	return nil
}
