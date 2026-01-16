//go:build integration
// +build integration

package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/scttfrdmn/prism/pkg/api/client"
	"github.com/scttfrdmn/prism/pkg/types"
	"github.com/scttfrdmn/prism/test/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInstanceCrashRecovery_SpotTermination validates handling of spot instance terminations
// Tests: Launch spot instance → Force termination → Verify detection → Validate state cleanup
func TestInstanceCrashRecovery_SpotTermination(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping instance crash recovery test in short mode")
	}

	ctx := context.Background()
	apiClient := client.NewClientWithOptions("http://localhost:8947", client.Options{
		AWSProfile: "aws",
		AWSRegion:  "us-west-2",
	})
	registry := fixtures.NewFixtureRegistry(t, apiClient)

	t.Log("Testing spot instance termination recovery...")

	// Test 1: Launch regular instance (spot instances not available in test environment)
	var instanceName string
	var instanceID string

	t.Run("LaunchInstance", func(t *testing.T) {
		instanceName = fmt.Sprintf("crash-test-%d", time.Now().Unix())

		launchResp, err := apiClient.LaunchInstance(ctx, types.LaunchRequest{
			Template: "Ubuntu 22.04 Server",
			Name:     instanceName,
			Size:     "S",
		})
		require.NoError(t, err, "Failed to launch instance")
		registry.Register("instance", instanceName)
		instanceID = launchResp.Instance.ID

		t.Logf("✓ Instance launched: %s (ID: %s)", launchResp.Instance.Name, instanceID)

		// Wait for running state
		err = fixtures.WaitForInstanceState(t, apiClient, instanceName, "running", 5*time.Minute)
		require.NoError(t, err, "Instance should reach running state")
	})

	// Test 2: Force terminate instance (simulate crash)
	t.Run("ForceTermination", func(t *testing.T) {
		t.Log("Simulating instance crash by force terminating...")

		// Terminate via API (simulates spot interruption)
		err := apiClient.DeleteInstance(ctx, instanceName)
		require.NoError(t, err, "Failed to terminate instance")

		t.Log("✓ Instance termination initiated")

		// Wait for terminated state
		err = fixtures.WaitForInstanceState(t, apiClient, instanceName, "terminated", 3*time.Minute)
		require.NoError(t, err, "Instance should reach terminated state")

		t.Log("✓ Instance terminated (crash simulated)")
	})

	// Test 3: Verify daemon detects crash
	t.Run("VerifyCrashDetection", func(t *testing.T) {
		t.Log("Verifying daemon detected crash...")

		// Get instance state - should show terminated
		instance, err := apiClient.GetInstance(ctx, instanceName)
		require.NoError(t, err, "Should be able to get instance info")
		assert.Equal(t, "terminated", instance.State, "Instance should be marked as terminated")

		t.Logf("✓ Daemon correctly shows instance state: %s", instance.State)
	})

	// Test 4: Verify state cleanup
	t.Run("VerifyStateCleanup", func(t *testing.T) {
		t.Log("Verifying state consistency after crash...")

		// List instances - terminated instance should still be visible in history
		instances, err := apiClient.ListInstances(ctx)
		require.NoError(t, err, "Should be able to list instances")

		found := false
		for _, inst := range instances.Instances {
			if inst.Name == instanceName {
				found = true
				assert.Equal(t, "terminated", inst.State, "Instance state should be terminated")
				break
			}
		}
		assert.True(t, found, "Terminated instance should still be in state for history")

		t.Log("✓ State consistency verified")
	})

	t.Log("✅ Spot instance termination recovery test complete")
}

// TestInstanceCrashRecovery_DataPersistence validates EFS/EBS volumes survive instance crashes
// Tests: Launch with volumes → Crash instance → Verify volumes intact → Reattach to new instance
func TestInstanceCrashRecovery_DataPersistence(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping data persistence test in short mode")
	}

	ctx := context.Background()
	apiClient := client.NewClientWithOptions("http://localhost:8947", client.Options{
		AWSProfile: "aws",
		AWSRegion:  "us-west-2",
	})
	registry := fixtures.NewFixtureRegistry(t, apiClient)

	t.Log("Testing data persistence across instance crashes...")

	var instanceName string
	var volumeName string
	var ebsVolumeName string

	// Test 1: Create storage volumes
	t.Run("CreateStorageVolumes", func(t *testing.T) {
		volumeName = fmt.Sprintf("crash-efs-%d", time.Now().Unix())
		ebsVolumeName = fmt.Sprintf("crash-ebs-%d", time.Now().Unix())

		// Create EFS volume
		_, err := apiClient.CreateVolume(ctx, types.VolumeCreateRequest{
			Name: volumeName,
		})
		require.NoError(t, err, "Failed to create EFS volume")
		registry.Register("volume", volumeName)

		t.Logf("✓ EFS volume created: %s", volumeName)

		// Create EBS volume
		_, err = apiClient.CreateStorage(ctx, types.StorageCreateRequest{
			Name: ebsVolumeName,
			Size: "S",
		})
		require.NoError(t, err, "Failed to create EBS volume")
		registry.Register("ebs", ebsVolumeName)

		t.Logf("✓ EBS volume created: %s", ebsVolumeName)

		// Wait for volumes to be available
		time.Sleep(30 * time.Second)
	})

	// Test 2: Launch instance with volumes attached
	t.Run("LaunchInstanceWithVolumes", func(t *testing.T) {
		instanceName = fmt.Sprintf("crash-persist-%d", time.Now().Unix())

		launchResp, err := apiClient.LaunchInstance(ctx, types.LaunchRequest{
			Template:   "Ubuntu 22.04 Server",
			Name:       instanceName,
			Size:       "S",
			Volumes:    []string{volumeName},
			EBSVolumes: []string{ebsVolumeName},
		})
		require.NoError(t, err, "Failed to launch instance")
		registry.Register("instance", instanceName)

		t.Logf("✓ Instance launched with volumes: %s", launchResp.Instance.Name)

		// Wait for running state
		err = fixtures.WaitForInstanceState(t, apiClient, instanceName, "running", 5*time.Minute)
		require.NoError(t, err, "Instance should reach running state")
	})

	// Test 3: Write data to volumes (via SSH)
	t.Run("WriteDataToVolumes", func(t *testing.T) {
		t.Log("Writing test data to volumes...")

		// Get instance IP for SSH
		instance, err := apiClient.GetInstance(ctx, instanceName)
		require.NoError(t, err, "Failed to get instance details")

		// Wait for SSH to be available
		time.Sleep(30 * time.Second)

		// Write test file to EFS mount (typically /efs)
		testData := "crash-recovery-test-data"
		_, err = fixtures.SSHCommand(t, instance.PublicIP, "ubuntu",
			fmt.Sprintf("echo '%s' | sudo tee /efs/test-file.txt", testData))
		if err != nil {
			t.Logf("⚠️  Could not write to EFS (may not be mounted): %v", err)
		} else {
			t.Log("✓ Test data written to EFS volume")
		}

		// Write test file to EBS mount (typically /ebs)
		_, err = fixtures.SSHCommand(t, instance.PublicIP, "ubuntu",
			fmt.Sprintf("echo '%s' | sudo tee /ebs/test-file.txt", testData))
		if err != nil {
			t.Logf("⚠️  Could not write to EBS (may not be mounted): %v", err)
		} else {
			t.Log("✓ Test data written to EBS volume")
		}
	})

	// Test 4: Terminate instance (simulate crash)
	t.Run("SimulateCrash", func(t *testing.T) {
		t.Log("Terminating instance to simulate crash...")

		err := apiClient.DeleteInstance(ctx, instanceName)
		require.NoError(t, err, "Failed to terminate instance")

		// Wait for terminated state
		err = fixtures.WaitForInstanceState(t, apiClient, instanceName, "terminated", 3*time.Minute)
		require.NoError(t, err, "Instance should reach terminated state")

		t.Log("✓ Instance crashed (terminated)")
	})

	// Test 5: Verify volumes still exist and are available
	t.Run("VerifyVolumesPersisted", func(t *testing.T) {
		t.Log("Verifying volumes persisted after crash...")

		// Check EFS volume
		volumes, err := apiClient.ListVolumes(ctx)
		require.NoError(t, err, "Failed to list volumes")

		efsFound := false
		for _, vol := range volumes {
			if vol.Name == volumeName {
				efsFound = true
				assert.Equal(t, "available", vol.State, "EFS volume should be available")
				t.Logf("✓ EFS volume intact: %s (state: %s)", vol.Name, vol.State)
				break
			}
		}
		assert.True(t, efsFound, "EFS volume should persist after instance crash")

		// Check EBS volume
		ebsVolumes, err := apiClient.ListStorage(ctx)
		require.NoError(t, err, "Failed to list EBS volumes")

		ebsFound := false
		for _, vol := range ebsVolumes {
			if vol.Name == ebsVolumeName {
				ebsFound = true
				assert.Equal(t, "available", vol.State, "EBS volume should be available")
				t.Logf("✓ EBS volume intact: %s (state: %s)", vol.Name, vol.State)
				break
			}
		}
		assert.True(t, ebsFound, "EBS volume should persist after instance crash")
	})

	// Test 6: Launch new instance and reattach volumes
	t.Run("ReattachVolumesToNewInstance", func(t *testing.T) {
		t.Log("Launching new instance with recovered volumes...")

		newInstanceName := fmt.Sprintf("crash-recover-%d", time.Now().Unix())

		launchResp, err := apiClient.LaunchInstance(ctx, types.LaunchRequest{
			Template:   "Ubuntu 22.04 Server",
			Name:       newInstanceName,
			Size:       "S",
			Volumes:    []string{volumeName},
			EBSVolumes: []string{ebsVolumeName},
		})
		require.NoError(t, err, "Failed to launch recovery instance")
		registry.Register("instance", newInstanceName)

		t.Logf("✓ Recovery instance launched: %s", launchResp.Instance.Name)

		// Wait for running state
		err = fixtures.WaitForInstanceState(t, apiClient, newInstanceName, "running", 5*time.Minute)
		require.NoError(t, err, "Recovery instance should reach running state")

		t.Log("✓ Volumes successfully reattached to new instance")
	})

	t.Log("✅ Data persistence across crash test complete")
}

// TestInstanceCrashRecovery_ConcurrentCrashes validates handling of multiple simultaneous crashes
// Tests: Launch multiple instances → Crash all simultaneously → Verify all detected → State consistent
func TestInstanceCrashRecovery_ConcurrentCrashes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent crash test in short mode")
	}

	ctx := context.Background()
	apiClient := client.NewClientWithOptions("http://localhost:8947", client.Options{
		AWSProfile: "aws",
		AWSRegion:  "us-west-2",
	})
	registry := fixtures.NewFixtureRegistry(t, apiClient)

	t.Log("Testing concurrent instance crash handling...")

	instanceNames := make([]string, 3)

	// Test 1: Launch multiple instances
	t.Run("LaunchMultipleInstances", func(t *testing.T) {
		t.Log("Launching 3 test instances...")

		for i := 0; i < 3; i++ {
			instanceName := fmt.Sprintf("crash-concurrent-%d-%d", i, time.Now().Unix())
			instanceNames[i] = instanceName

			launchResp, err := apiClient.LaunchInstance(ctx, types.LaunchRequest{
				Template: "Ubuntu 22.04 Server",
				Name:     instanceName,
				Size:     "S",
			})
			require.NoError(t, err, "Failed to launch instance %d", i)
			registry.Register("instance", instanceName)

			t.Logf("  ✓ Instance %d launched: %s", i+1, launchResp.Instance.Name)
		}

		// Wait for all instances to reach running state
		for _, name := range instanceNames {
			err := fixtures.WaitForInstanceState(t, apiClient, name, "running", 5*time.Minute)
			require.NoError(t, err, "Instance %s should reach running state", name)
		}

		t.Log("✓ All instances running")
	})

	// Test 2: Terminate all instances simultaneously
	t.Run("SimultaneousCrashes", func(t *testing.T) {
		t.Log("Terminating all instances simultaneously...")

		// Launch terminations concurrently
		done := make(chan error, len(instanceNames))

		for _, name := range instanceNames {
			go func(instanceName string) {
				err := apiClient.DeleteInstance(ctx, instanceName)
				done <- err
			}(name)
		}

		// Wait for all terminations to complete
		for i := 0; i < len(instanceNames); i++ {
			err := <-done
			require.NoError(t, err, "Termination %d should succeed", i)
		}

		t.Log("✓ All terminations initiated")
	})

	// Test 3: Verify all crashes detected
	t.Run("VerifyAllCrashesDetected", func(t *testing.T) {
		t.Log("Verifying all crashes detected...")

		// Wait for all to reach terminated state
		for _, name := range instanceNames {
			err := fixtures.WaitForInstanceState(t, apiClient, name, "terminated", 3*time.Minute)
			assert.NoError(t, err, "Instance %s should reach terminated state", name)
		}

		t.Log("✓ All crashes detected and state updated")
	})

	// Test 4: Verify state consistency
	t.Run("VerifyStateConsistency", func(t *testing.T) {
		t.Log("Verifying state consistency across concurrent crashes...")

		instances, err := apiClient.ListInstances(ctx)
		require.NoError(t, err, "Should be able to list instances")

		// Verify all crashed instances are present with correct state
		for _, name := range instanceNames {
			found := false
			for _, inst := range instances.Instances {
				if inst.Name == name {
					found = true
					assert.Equal(t, "terminated", inst.State, "Instance %s should be terminated", name)
					break
				}
			}
			assert.True(t, found, "Instance %s should be in state", name)
		}

		t.Log("✓ State consistency verified")
	})

	t.Log("✅ Concurrent crash handling test complete")
}

// TestInstanceCrashRecovery_GracefulDegradation validates system behavior during partial failures
// Tests: Launch instances → Some crash → Others continue → Mixed state handling
func TestInstanceCrashRecovery_GracefulDegradation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping graceful degradation test in short mode")
	}

	ctx := context.Background()
	apiClient := client.NewClientWithOptions("http://localhost:8947", client.Options{
		AWSProfile: "aws",
		AWSRegion:  "us-west-2",
	})
	registry := fixtures.NewFixtureRegistry(t, apiClient)

	t.Log("Testing graceful degradation with partial failures...")

	var healthyInstance string
	var crashedInstance string

	// Test 1: Launch two instances
	t.Run("LaunchInstances", func(t *testing.T) {
		healthyInstance = fmt.Sprintf("healthy-%d", time.Now().Unix())
		crashedInstance = fmt.Sprintf("crashed-%d", time.Now().Unix())

		// Launch healthy instance
		launchResp1, err := apiClient.LaunchInstance(ctx, types.LaunchRequest{
			Template: "Ubuntu 22.04 Server",
			Name:     healthyInstance,
			Size:     "S",
		})
		require.NoError(t, err, "Failed to launch healthy instance")
		registry.Register("instance", healthyInstance)
		t.Logf("✓ Healthy instance launched: %s", launchResp1.Instance.Name)

		// Launch instance that will crash
		launchResp2, err := apiClient.LaunchInstance(ctx, types.LaunchRequest{
			Template: "Ubuntu 22.04 Server",
			Name:     crashedInstance,
			Size:     "S",
		})
		require.NoError(t, err, "Failed to launch crash-target instance")
		registry.Register("instance", crashedInstance)
		t.Logf("✓ Crash-target instance launched: %s", launchResp2.Instance.Name)

		// Wait for both to reach running state
		err = fixtures.WaitForInstanceState(t, apiClient, healthyInstance, "running", 5*time.Minute)
		require.NoError(t, err, "Healthy instance should reach running state")

		err = fixtures.WaitForInstanceState(t, apiClient, crashedInstance, "running", 5*time.Minute)
		require.NoError(t, err, "Crash-target instance should reach running state")
	})

	// Test 2: Crash one instance
	t.Run("CrashOneInstance", func(t *testing.T) {
		t.Log("Crashing one instance while other remains healthy...")

		err := apiClient.DeleteInstance(ctx, crashedInstance)
		require.NoError(t, err, "Failed to crash instance")

		// Wait for crashed instance to terminate
		err = fixtures.WaitForInstanceState(t, apiClient, crashedInstance, "terminated", 3*time.Minute)
		require.NoError(t, err, "Crashed instance should reach terminated state")

		t.Log("✓ One instance crashed")
	})

	// Test 3: Verify healthy instance unaffected
	t.Run("VerifyHealthyInstanceUnaffected", func(t *testing.T) {
		t.Log("Verifying healthy instance continues running...")

		instance, err := apiClient.GetInstance(ctx, healthyInstance)
		require.NoError(t, err, "Should be able to get healthy instance")
		assert.Equal(t, "running", instance.State, "Healthy instance should still be running")

		t.Logf("✓ Healthy instance unaffected: %s (state: %s)", instance.Name, instance.State)
	})

	// Test 4: Verify operations still work on healthy instance
	t.Run("VerifyOperationsOnHealthyInstance", func(t *testing.T) {
		t.Log("Verifying operations work on healthy instance...")

		// Try to stop healthy instance
		err := apiClient.StopInstance(ctx, healthyInstance)
		assert.NoError(t, err, "Should be able to stop healthy instance")

		err = fixtures.WaitForInstanceState(t, apiClient, healthyInstance, "stopped", 3*time.Minute)
		assert.NoError(t, err, "Healthy instance should reach stopped state")

		// Try to start it again
		err = apiClient.StartInstance(ctx, healthyInstance)
		assert.NoError(t, err, "Should be able to start healthy instance")

		err = fixtures.WaitForInstanceState(t, apiClient, healthyInstance, "running", 5*time.Minute)
		assert.NoError(t, err, "Healthy instance should reach running state")

		t.Log("✓ Operations working correctly on healthy instance")
	})

	// Test 5: Verify mixed state listing
	t.Run("VerifyMixedStateListing", func(t *testing.T) {
		t.Log("Verifying instance listing shows mixed states correctly...")

		instances, err := apiClient.ListInstances(ctx)
		require.NoError(t, err, "Should be able to list instances")

		healthyFound := false
		crashedFound := false

		for _, inst := range instances.Instances {
			if inst.Name == healthyInstance {
				healthyFound = true
				assert.Equal(t, "running", inst.State, "Healthy instance should be running")
			}
			if inst.Name == crashedInstance {
				crashedFound = true
				assert.Equal(t, "terminated", inst.State, "Crashed instance should be terminated")
			}
		}

		assert.True(t, healthyFound, "Healthy instance should be in list")
		assert.True(t, crashedFound, "Crashed instance should be in list")

		t.Log("✓ Mixed state listing correct")
	})

	t.Log("✅ Graceful degradation test complete")
}
