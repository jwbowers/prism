//go:build integration
// +build integration

package localstack_test

import (
	"context"
	"testing"
	"time"

	"github.com/scttfrdmn/prism/pkg/aws/localstack"
	"github.com/scttfrdmn/prism/test/localstack/fixtures"
	"github.com/stretchr/testify/require"
)

// TestLocalStackHealthCheck verifies LocalStack is running and healthy
func TestLocalStackHealthCheck(t *testing.T) {
	if !localstack.IsLocalStackEnabled() {
		t.Skip("LocalStack not enabled (set PRISM_USE_LOCALSTACK=true)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Wait for LocalStack to be ready
	err := localstack.WaitForReady(ctx, 30*time.Second)
	require.NoError(t, err, "LocalStack should be healthy and ready")

	// Verify LocalStack is healthy
	healthy := localstack.IsHealthy(ctx)
	require.True(t, healthy, "LocalStack should report healthy status")

	t.Log("✓ LocalStack is healthy and ready")
}

// TestLocalStackConfiguration verifies LocalStack configuration is loaded correctly
func TestLocalStackConfiguration(t *testing.T) {
	if !localstack.IsLocalStackEnabled() {
		t.Skip("LocalStack not enabled (set PRISM_USE_LOCALSTACK=true)")
	}

	// Load configuration
	config, err := localstack.LoadConfig()
	require.NoError(t, err, "Should load LocalStack configuration")
	require.NotNil(t, config, "Configuration should not be nil")

	// Validate configuration
	err = config.Validate()
	require.NoError(t, err, "Configuration should be valid")

	// Verify VPC configuration
	require.NotEmpty(t, config.VPCID, "VPC ID should be set")
	require.NotEmpty(t, config.SecurityGroupID, "Security group ID should be set")
	require.NotEmpty(t, config.SubnetIDs, "Subnet IDs should be set")
	require.Greater(t, len(config.SubnetIDs), 0, "At least one subnet should exist")

	// Verify AMI configuration
	require.NotEmpty(t, config.AMIIDs, "AMI IDs should be set")
	require.Greater(t, len(config.AMIIDs), 0, "At least one AMI should exist")

	// Verify key pair
	require.NotEmpty(t, config.KeyPair, "Key pair should be set")

	t.Logf("✓ LocalStack configuration loaded and validated")
	t.Logf("  - VPC: %s", config.VPCID)
	t.Logf("  - Subnets: %d", len(config.SubnetIDs))
	t.Logf("  - Security Group: %s", config.SecurityGroupID)
	t.Logf("  - AMIs: %d", len(config.AMIIDs))
	t.Logf("  - Key Pair: %s", config.KeyPair)
}

// TestLocalStackServices verifies all required AWS services are available
func TestLocalStackServices(t *testing.T) {
	if !localstack.IsLocalStackEnabled() {
		t.Skip("LocalStack not enabled (set PRISM_USE_LOCALSTACK=true)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Verify all required services
	err := localstack.VerifyRequiredServices(ctx)
	require.NoError(t, err, "All required services should be available")

	// Get detailed service status
	services, err := localstack.GetServiceStatus(ctx)
	require.NoError(t, err, "Should get service status")

	t.Log("✓ LocalStack services verified:")
	for _, service := range localstack.RequiredServices {
		status := services[service]
		// LocalStack Community reports services as "running" or "available"
		require.Truef(t, status == "available" || status == "running",
			"Service %s should be available or running, got: %s", service, status)
		t.Logf("  - %s: %s", service, status)
	}
}

// TestLocalStackNetworkSetup verifies VPC, subnets, and security groups
func TestLocalStackNetworkSetup(t *testing.T) {
	if !localstack.IsLocalStackEnabled() {
		t.Skip("LocalStack not enabled (set PRISM_USE_LOCALSTACK=true)")
	}

	f := fixtures.NewTestFixtures(t)
	f.VerifyNetworkSetup(t)

	t.Log("✓ LocalStack network setup verified")
}

// TestLocalStackAMIs verifies all AMIs are available
func TestLocalStackAMIs(t *testing.T) {
	if !localstack.IsLocalStackEnabled() {
		t.Skip("LocalStack not enabled (set PRISM_USE_LOCALSTACK=true)")
	}

	f := fixtures.NewTestFixtures(t)
	f.VerifyAMIsExist(t)

	// Verify we can get specific AMIs
	testCases := []struct {
		name string
		os   string
		arch string
	}{
		{"Ubuntu 22.04 x86_64", "ubuntu-22.04", "x86_64"},
		{"Ubuntu 22.04 ARM64", "ubuntu-22.04", "arm64"},
		{"Rocky Linux 9 x86_64", "rockylinux-9", "x86_64"},
		{"Rocky Linux 9 ARM64", "rockylinux-9", "arm64"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ami, err := f.GetAMIForTemplate(tc.os, tc.arch)
			require.NoError(t, err, "Should get AMI for %s", tc.name)
			require.NotEmpty(t, ami, "AMI ID should not be empty")
			t.Logf("  %s: %s", tc.name, ami)
		})
	}

	t.Log("✓ LocalStack AMIs verified")
}

// TestLocalStackEFS verifies EFS setup
func TestLocalStackEFS(t *testing.T) {
	if !localstack.IsLocalStackEnabled() {
		t.Skip("LocalStack not enabled (set PRISM_USE_LOCALSTACK=true)")
	}
	if !localstack.HasEFSSupport() {
		t.Skip("EFS not available in LocalStack Community edition (set LOCALSTACK_EFS_SUPPORT=true for Pro)")
	}

	f := fixtures.NewTestFixtures(t)
	f.VerifyEFSSetup(t)

	t.Log("✓ LocalStack EFS setup verified")
}

// TestLocalStackSSM verifies SSM Parameter Store setup
func TestLocalStackSSM(t *testing.T) {
	if !localstack.IsLocalStackEnabled() {
		t.Skip("LocalStack not enabled (set PRISM_USE_LOCALSTACK=true)")
	}

	f := fixtures.NewTestFixtures(t)

	// Check if SSM parameters were seeded (may fail if LocalStack restricts /aws/service/ writes)
	if !f.HasSSMParameters(t) {
		t.Skip("SSM parameters not available (LocalStack may restrict /aws/service/ namespace writes)")
	}

	f.VerifySSMParameters(t)
	t.Log("✓ LocalStack SSM parameters verified")
}

// TestLocalStackInstanceLaunch verifies we can launch an EC2 instance
func TestLocalStackInstanceLaunch(t *testing.T) {
	if !localstack.IsLocalStackEnabled() {
		t.Skip("LocalStack not enabled (set PRISM_USE_LOCALSTACK=true)")
	}

	f := fixtures.NewTestFixtures(t)

	// Launch a test instance
	instanceID := f.LaunchTestInstance(t, "test-instance-basic")
	require.NotEmpty(t, instanceID, "Instance ID should not be empty")

	t.Logf("✓ Successfully launched instance: %s", instanceID)
	t.Log("✓ LocalStack instance launch verified")
}

// TestLocalStackEndToEnd performs a complete end-to-end test
func TestLocalStackEndToEnd(t *testing.T) {
	if !localstack.IsLocalStackEnabled() {
		t.Skip("LocalStack not enabled (set PRISM_USE_LOCALSTACK=true)")
	}

	ctx := context.Background()

	t.Log("=== LocalStack End-to-End Verification ===")

	// 1. Health check
	t.Log("1. Checking LocalStack health...")
	err := localstack.WaitForReady(ctx, 30*time.Second)
	require.NoError(t, err)
	t.Log("   ✓ LocalStack is healthy")

	// 2. Load configuration
	t.Log("2. Loading LocalStack configuration...")
	config, err := localstack.LoadConfig()
	require.NoError(t, err)
	err = config.Validate()
	require.NoError(t, err)
	t.Log("   ✓ Configuration loaded and validated")

	// 3. Verify services
	t.Log("3. Verifying AWS services...")
	err = localstack.VerifyRequiredServices(ctx)
	require.NoError(t, err)
	t.Log("   ✓ All required services available")

	// 4. Create fixtures
	t.Log("4. Creating test fixtures...")
	f := fixtures.NewTestFixtures(t)
	require.NotNil(t, f)
	t.Log("   ✓ Test fixtures created")

	// 5. Verify network setup
	t.Log("5. Verifying network setup...")
	f.VerifyNetworkSetup(t)
	t.Log("   ✓ Network setup verified")

	// 6. Verify AMIs
	t.Log("6. Verifying AMIs...")
	f.VerifyAMIsExist(t)
	t.Log("   ✓ AMIs verified")

	// 7. Verify EFS (Pro only — Community edition does not include EFS)
	if localstack.HasEFSSupport() {
		t.Log("7. Verifying EFS...")
		f.VerifyEFSSetup(t)
		t.Log("   ✓ EFS verified")
	} else {
		t.Log("7. Skipping EFS verification (not available in LocalStack Community edition)")
	}

	// 8. Verify SSM (optional - LocalStack Community may restrict /aws/service/ writes)
	t.Log("8. Verifying SSM parameters (optional)...")
	if f.HasSSMParameters(t) {
		f.VerifySSMParameters(t)
		t.Log("   ✓ SSM parameters verified")
	} else {
		t.Log("   ⚠ SSM parameters not available (LocalStack may restrict /aws/service/ namespace)")
	}

	// 9. Launch test instance
	t.Log("9. Launching test instance...")
	instanceID := f.LaunchTestInstance(t, "e2e-test-instance")
	require.NotEmpty(t, instanceID)
	t.Logf("   ✓ Instance launched: %s", instanceID)

	t.Log("")
	t.Log("=== ✓ LocalStack End-to-End Verification Complete ===")
	t.Log("")
	t.Log("LocalStack is fully operational and ready for Prism integration tests")
}
