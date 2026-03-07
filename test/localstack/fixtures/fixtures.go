// Package fixtures provides test fixtures for LocalStack integration tests
package fixtures

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/efs"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/scttfrdmn/prism/pkg/aws/localstack"
)

// TestFixtures holds LocalStack test resources
type TestFixtures struct {
	Config    *localstack.Config
	EC2Client *ec2.Client
	EFSClient *efs.Client
	SSMClient *ssm.Client
}

// NewTestFixtures creates test fixtures for LocalStack tests
// This loads the LocalStack configuration and creates AWS clients
func NewTestFixtures(t *testing.T) *TestFixtures {
	t.Helper()

	if !localstack.IsLocalStackEnabled() {
		t.Skip("LocalStack not enabled (set PRISM_USE_LOCALSTACK=true)")
	}

	// Wait for LocalStack to be ready
	ctx := context.Background()
	if err := localstack.WaitForReady(ctx, 60*time.Second); err != nil {
		t.Fatalf("LocalStack not ready: %v", err)
	}

	// Load LocalStack configuration
	config, err := localstack.LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load LocalStack config: %v", err)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		t.Fatalf("LocalStack config validation failed: %v", err)
	}

	// Create AWS clients
	cfg, err := localstack.NewAWSConfig(ctx)
	if err != nil {
		t.Fatalf("Failed to create LocalStack AWS config: %v", err)
	}

	return &TestFixtures{
		Config:    config,
		EC2Client: ec2.NewFromConfig(cfg),
		EFSClient: efs.NewFromConfig(cfg),
		SSMClient: ssm.NewFromConfig(cfg),
	}
}

// VerifyNetworkSetup verifies that VPC, subnets, and security groups exist
func (f *TestFixtures) VerifyNetworkSetup(t *testing.T) {
	t.Helper()

	ctx := context.Background()

	// Verify VPC exists
	vpcOutput, err := f.EC2Client.DescribeVpcs(ctx, &ec2.DescribeVpcsInput{
		VpcIds: []string{f.Config.VPCID},
	})
	if err != nil {
		t.Fatalf("Failed to describe VPC: %v", err)
	}
	if len(vpcOutput.Vpcs) == 0 {
		t.Fatalf("VPC %s not found", f.Config.VPCID)
	}
	t.Logf("VPC verified: %s", f.Config.VPCID)

	// Verify subnets exist
	subnetOutput, err := f.EC2Client.DescribeSubnets(ctx, &ec2.DescribeSubnetsInput{
		SubnetIds: f.Config.SubnetIDs,
	})
	if err != nil {
		t.Fatalf("Failed to describe subnets: %v", err)
	}
	if len(subnetOutput.Subnets) != len(f.Config.SubnetIDs) {
		t.Fatalf("Expected %d subnets, found %d", len(f.Config.SubnetIDs), len(subnetOutput.Subnets))
	}
	t.Logf("Subnets verified: %d subnets", len(subnetOutput.Subnets))

	// Verify security group exists
	sgOutput, err := f.EC2Client.DescribeSecurityGroups(ctx, &ec2.DescribeSecurityGroupsInput{
		GroupIds: []string{f.Config.SecurityGroupID},
	})
	if err != nil {
		t.Fatalf("Failed to describe security group: %v", err)
	}
	if len(sgOutput.SecurityGroups) == 0 {
		t.Fatalf("Security group %s not found", f.Config.SecurityGroupID)
	}
	t.Logf("Security group verified: %s", f.Config.SecurityGroupID)
}

// VerifyAMIsExist verifies that all AMIs are available
func (f *TestFixtures) VerifyAMIsExist(t *testing.T) {
	t.Helper()

	ctx := context.Background()

	// Get list of AMI IDs
	var amiIDs []string
	for _, amiID := range f.Config.AMIIDs {
		amiIDs = append(amiIDs, amiID)
	}

	// Describe AMIs
	output, err := f.EC2Client.DescribeImages(ctx, &ec2.DescribeImagesInput{
		ImageIds: amiIDs,
	})
	if err != nil {
		t.Fatalf("Failed to describe AMIs: %v", err)
	}

	if len(output.Images) != len(amiIDs) {
		t.Fatalf("Expected %d AMIs, found %d", len(amiIDs), len(output.Images))
	}

	t.Logf("AMIs verified: %d images available", len(output.Images))
}

// VerifyEFSSetup verifies that EFS file system and mount targets exist
func (f *TestFixtures) VerifyEFSSetup(t *testing.T) {
	t.Helper()

	if f.Config.EFSFileSystemID == "" {
		t.Skip("EFS file system not configured in LocalStack")
	}

	ctx := context.Background()

	// Verify EFS file system exists
	fsOutput, err := f.EFSClient.DescribeFileSystems(ctx, &efs.DescribeFileSystemsInput{
		FileSystemId: aws.String(f.Config.EFSFileSystemID),
	})
	if err != nil {
		t.Fatalf("Failed to describe EFS file system: %v", err)
	}
	if len(fsOutput.FileSystems) == 0 {
		t.Fatalf("EFS file system %s not found", f.Config.EFSFileSystemID)
	}
	t.Logf("EFS file system verified: %s", f.Config.EFSFileSystemID)

	// Verify mount targets exist
	if len(f.Config.EFSMountTargetIDs) > 0 {
		mtOutput, err := f.EFSClient.DescribeMountTargets(ctx, &efs.DescribeMountTargetsInput{
			FileSystemId: aws.String(f.Config.EFSFileSystemID),
		})
		if err != nil {
			t.Fatalf("Failed to describe EFS mount targets: %v", err)
		}
		if len(mtOutput.MountTargets) != len(f.Config.EFSMountTargetIDs) {
			t.Fatalf("Expected %d mount targets, found %d", len(f.Config.EFSMountTargetIDs), len(mtOutput.MountTargets))
		}
		t.Logf("EFS mount targets verified: %d targets", len(mtOutput.MountTargets))
	}
}

// HasSSMParameters returns true if the canonical Ubuntu SSM parameter was successfully seeded.
// LocalStack Community may restrict writes to the /aws/service/ namespace.
func (f *TestFixtures) HasSSMParameters(t *testing.T) bool {
	t.Helper()
	ctx := context.Background()
	_, err := f.SSMClient.GetParameter(ctx, &ssm.GetParameterInput{
		Name: aws.String("/aws/service/canonical/ubuntu/server/22.04/stable/current/amd64/hvm/ebs-gp2/ami-id"),
	})
	return err == nil
}

// VerifySSMParameters verifies that SSM parameters for AMI discovery exist
func (f *TestFixtures) VerifySSMParameters(t *testing.T) {
	t.Helper()

	ctx := context.Background()

	// List of expected SSM parameters for AMI discovery
	expectedParams := []string{
		"/aws/service/canonical/ubuntu/server/22.04/stable/current/amd64/hvm/ebs-gp2/ami-id",
		"/aws/service/canonical/ubuntu/server/22.04/stable/current/arm64/hvm/ebs-gp2/ami-id",
		"/aws/service/rockylinux/9/x86_64/ami-id",
		"/aws/service/rockylinux/9/arm64/ami-id",
		"/aws/service/ami-amazon-linux-latest/al2023-ami-kernel-default-x86_64",
		"/aws/service/ami-amazon-linux-latest/al2023-ami-kernel-default-arm64",
	}

	// Verify each parameter exists
	for _, paramName := range expectedParams {
		_, err := f.SSMClient.GetParameter(ctx, &ssm.GetParameterInput{
			Name: aws.String(paramName),
		})
		if err != nil {
			t.Fatalf("SSM parameter %s not found: %v", paramName, err)
		}
	}

	t.Logf("SSM parameters verified: %d parameters", len(expectedParams))
}

// GetAMIForTemplate returns the appropriate AMI ID for a template test
// This maps template requirements to LocalStack mock AMIs
func (f *TestFixtures) GetAMIForTemplate(templateOS, arch string) (string, error) {
	// Map template OS to LocalStack AMI keys
	amiKey := fmt.Sprintf("%s-%s", templateOS, arch)

	// Common mappings
	mappings := map[string]string{
		"ubuntu-22.04-x86_64":     "ubuntu-22.04-x86_64",
		"ubuntu-22.04-arm64":      "ubuntu-22.04-arm64",
		"rockylinux-9-x86_64":     "rockylinux-9-x86_64",
		"rockylinux-9-arm64":      "rockylinux-9-arm64",
		"amazonlinux-2023-x86_64": "amazonlinux-2023-x86_64",
		"amazonlinux-2023-arm64":  "amazonlinux-2023-arm64",
		"debian-12-x86_64":        "debian-12-x86_64",
		"debian-12-arm64":         "debian-12-arm64",
	}

	if mappedKey, exists := mappings[amiKey]; exists {
		return f.Config.GetAMI(mappedKey)
	}

	return "", fmt.Errorf("no AMI mapping found for %s", amiKey)
}

// LaunchTestInstance launches a test EC2 instance in LocalStack
// This is a low-level helper for testing EC2 operations directly
func (f *TestFixtures) LaunchTestInstance(t *testing.T, name string) string {
	t.Helper()

	ctx := context.Background()

	// Get default AMI (Ubuntu 22.04 x86_64)
	amiID, err := f.Config.GetAMI("ubuntu-22.04-x86_64")
	if err != nil {
		t.Fatalf("Failed to get AMI: %v", err)
	}

	// Get default subnet
	subnetID, err := f.Config.GetDefaultSubnetID()
	if err != nil {
		t.Fatalf("Failed to get subnet: %v", err)
	}

	// Launch instance
	output, err := f.EC2Client.RunInstances(ctx, &ec2.RunInstancesInput{
		ImageId:          aws.String(amiID),
		InstanceType:     types.InstanceTypeT3Micro,
		MinCount:         aws.Int32(1),
		MaxCount:         aws.Int32(1),
		SubnetId:         aws.String(subnetID),
		SecurityGroupIds: []string{f.Config.SecurityGroupID},
		KeyName:          aws.String(f.Config.KeyPair),
		TagSpecifications: []types.TagSpecification{
			{
				ResourceType: types.ResourceTypeInstance,
				Tags: []types.Tag{
					{Key: aws.String("Name"), Value: aws.String(name)},
					{Key: aws.String("ManagedBy"), Value: aws.String("prism-test")},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("Failed to launch instance: %v", err)
	}

	if len(output.Instances) == 0 {
		t.Fatal("No instances returned from RunInstances")
	}

	instanceID := *output.Instances[0].InstanceId
	t.Logf("Launched test instance: %s (%s)", instanceID, name)

	// Register cleanup
	t.Cleanup(func() {
		_, _ = f.EC2Client.TerminateInstances(ctx, &ec2.TerminateInstancesInput{
			InstanceIds: []string{instanceID},
		})
	})

	return instanceID
}

// TerminateTestInstance terminates a test instance (used for cleanup)
func (f *TestFixtures) TerminateTestInstance(t *testing.T, instanceID string) {
	t.Helper()

	ctx := context.Background()

	_, err := f.EC2Client.TerminateInstances(ctx, &ec2.TerminateInstancesInput{
		InstanceIds: []string{instanceID},
	})
	if err != nil {
		t.Logf("Warning: Failed to terminate instance %s: %v", instanceID, err)
	}
}
