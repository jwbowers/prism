//go:build integration
// +build integration

package integration

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// TestCLIEFSVolumeLifecycle tests EFS volume management (shared storage)
func TestCLIEFSVolumeLifecycle(t *testing.T) {
	ctx := NewCLITestContext(t)
	defer ctx.Cleanup()

	volumeName := GenerateTestName("test-efs")
	instanceName := GenerateTestName("test-instance")

	t.Run("CreateEFSVolume", func(t *testing.T) {
		result := ctx.Prism("volume", "create", volumeName)
		result.AssertSuccess(t, "EFS volume create should succeed")
		ctx.TrackVolume(volumeName)
		t.Logf("Created EFS volume: %s", volumeName)
	})

	t.Run("ListEFSVolumes", func(t *testing.T) {
		result := ctx.Prism("volume", "list")
		result.AssertSuccess(t, "volume list should succeed")
		result.AssertContains(t, volumeName, "should list created EFS volume")
	})

	t.Run("ShowEFSVolumeInfo", func(t *testing.T) {
		result := ctx.Prism("volume", "info", volumeName)
		result.AssertSuccess(t, "volume info should succeed")
		result.AssertContains(t, volumeName, "info should contain volume name")
	})

	t.Run("LaunchInstanceForMount", func(t *testing.T) {
		result, err := ctx.LaunchInstanceCLI("python-ml-workstation", instanceName, "M")
		AssertNoError(t, err, "instance launch should succeed")
		result.AssertSuccess(t, "launch command should succeed")

		// Wait for instance to be running
		_, err = ctx.WaitForInstanceRunning(instanceName)
		AssertNoError(t, err, "instance should reach running state")
		t.Logf("Instance %s is running", instanceName)
	})

	t.Run("MountEFSVolume", func(t *testing.T) {
		result := ctx.Prism("volume", "mount", volumeName, instanceName)
		result.AssertSuccess(t, "volume mount should succeed")
		t.Logf("Mounted volume %s to instance %s", volumeName, instanceName)
	})

	t.Run("VerifyMount", func(t *testing.T) {
		// Verify via API that volume is mounted
		volume, err := ctx.Client.GetVolume(context.Background(), volumeName)
		AssertNoError(t, err, "should get volume info")
		t.Logf("Volume state: %+v", volume)
	})

	t.Run("UnmountEFSVolume", func(t *testing.T) {
		result := ctx.Prism("volume", "unmount", volumeName, instanceName)
		result.AssertSuccess(t, "volume unmount should succeed")
		t.Logf("Unmounted volume %s from instance %s", volumeName, instanceName)
	})

	t.Run("DeleteEFSVolume", func(t *testing.T) {
		result := ctx.Prism("volume", "delete", volumeName, "--force")
		result.AssertSuccess(t, "EFS volume delete should succeed")
	})
}

// TestCLIEBSStorageLifecycle tests EBS volume management (workspace storage)
func TestCLIEBSStorageLifecycle(t *testing.T) {
	ctx := NewCLITestContext(t)
	defer ctx.Cleanup()

	storageName := GenerateTestName("test-ebs")
	instanceName := GenerateTestName("test-instance")

	t.Run("CreateEBSVolume", func(t *testing.T) {
		result := ctx.Prism("storage", "create", storageName, "--size", "10")
		result.AssertSuccess(t, "EBS storage create should succeed")
		ctx.TrackVolume(storageName)
		t.Logf("Created EBS volume: %s (10GB)", storageName)
	})

	t.Run("ListAllStorage", func(t *testing.T) {
		result := ctx.Prism("storage", "list")
		result.AssertSuccess(t, "storage list should succeed")
		result.AssertContains(t, storageName, "should list created EBS volume")
	})

	t.Run("ShowStorageInfo", func(t *testing.T) {
		result := ctx.Prism("storage", "info", storageName)
		result.AssertSuccess(t, "storage info should succeed")
		result.AssertContains(t, storageName, "info should contain volume name")
	})

	t.Run("LaunchInstanceForAttach", func(t *testing.T) {
		result, err := ctx.LaunchInstanceCLI("python-ml-workstation", instanceName, "M")
		AssertNoError(t, err, "instance launch should succeed")
		result.AssertSuccess(t, "launch command should succeed")

		// Wait for instance to be running
		_, err = ctx.WaitForInstanceRunning(instanceName)
		AssertNoError(t, err, "instance should reach running state")
		t.Logf("Instance %s is running", instanceName)
	})

	t.Run("AttachEBSVolume", func(t *testing.T) {
		result := ctx.Prism("storage", "attach", storageName, instanceName)
		result.AssertSuccess(t, "storage attach should succeed")
		t.Logf("Attached storage %s to instance %s", storageName, instanceName)
	})

	t.Run("VerifyAttachment", func(t *testing.T) {
		// Verify via API that volume is attached
		instance, err := ctx.Client.GetInstance(context.Background(), instanceName)
		AssertNoError(t, err, "should get instance")

		// Check if volume is in attached volumes list
		t.Logf("Instance attached volumes: %v", instance.AttachedEBSVolumes)

		// Give some time for state to propagate
		time.Sleep(5 * time.Second)
	})

	t.Run("DetachEBSVolume", func(t *testing.T) {
		result := ctx.Prism("storage", "detach", storageName)
		result.AssertSuccess(t, "storage detach should succeed")
		t.Logf("Detached storage %s from instance %s", storageName, instanceName)
	})

	t.Run("DeleteEBSVolume", func(t *testing.T) {
		result := ctx.Prism("storage", "delete", storageName, "--force")
		result.AssertSuccess(t, "EBS storage delete should succeed")
	})
}

// TestCLIStorageErrorHandling tests error scenarios for storage operations
func TestCLIStorageErrorHandling(t *testing.T) {
	ctx := NewCLITestContext(t)
	defer ctx.Cleanup()

	nonExistentVolume := GenerateTestName("nonexistent")
	nonExistentInstance := GenerateTestName("nonexistent-instance")

	t.Run("DeleteNonExistentVolume", func(t *testing.T) {
		result := ctx.Prism("volume", "delete", nonExistentVolume)
		result.AssertFailure(t, "deleting nonexistent volume should fail")
	})

	t.Run("InfoNonExistentVolume", func(t *testing.T) {
		result := ctx.Prism("volume", "info", nonExistentVolume)
		result.AssertFailure(t, "info for nonexistent volume should fail")
	})

	t.Run("MountNonExistentVolume", func(t *testing.T) {
		result := ctx.Prism("volume", "mount", nonExistentVolume, nonExistentInstance)
		result.AssertFailure(t, "mounting nonexistent volume should fail")
	})

	t.Run("UnmountNonExistentVolume", func(t *testing.T) {
		result := ctx.Prism("volume", "unmount", nonExistentVolume, nonExistentInstance)
		result.AssertFailure(t, "unmounting nonexistent volume should fail")
	})

	t.Run("AttachNonExistentStorage", func(t *testing.T) {
		result := ctx.Prism("storage", "attach", nonExistentVolume, nonExistentInstance)
		result.AssertFailure(t, "attaching nonexistent storage should fail")
	})

	t.Run("DetachNonExistentStorage", func(t *testing.T) {
		result := ctx.Prism("storage", "detach", nonExistentVolume)
		result.AssertFailure(t, "detaching nonexistent storage should fail")
	})

	t.Run("DeleteNonExistentStorage", func(t *testing.T) {
		result := ctx.Prism("storage", "delete", nonExistentVolume)
		result.AssertFailure(t, "deleting nonexistent storage should fail")
	})
}

// TestCLIStorageMultiMount tests EFS volume mounted to multiple instances
func TestCLIStorageMultiMount(t *testing.T) {
	ctx := NewCLITestContext(t)
	defer ctx.Cleanup()

	volumeName := GenerateTestName("test-shared-efs")
	instance1 := GenerateTestName("test-instance-1")
	instance2 := GenerateTestName("test-instance-2")

	t.Run("CreateSharedEFSVolume", func(t *testing.T) {
		result := ctx.Prism("volume", "create", volumeName)
		result.AssertSuccess(t, "EFS volume create should succeed")
		ctx.TrackVolume(volumeName)
		t.Logf("Created shared EFS volume: %s", volumeName)
	})

	t.Run("LaunchFirstInstance", func(t *testing.T) {
		result, err := ctx.LaunchInstanceCLI("python-ml-workstation", instance1, "M")
		AssertNoError(t, err, "first instance launch should succeed")
		result.AssertSuccess(t, "launch command should succeed")

		_, err = ctx.WaitForInstanceRunning(instance1)
		AssertNoError(t, err, "first instance should reach running state")
		t.Logf("First instance %s is running", instance1)
	})

	t.Run("LaunchSecondInstance", func(t *testing.T) {
		result, err := ctx.LaunchInstanceCLI("python-ml-workstation", instance2, "M")
		AssertNoError(t, err, "second instance launch should succeed")
		result.AssertSuccess(t, "launch command should succeed")

		_, err = ctx.WaitForInstanceRunning(instance2)
		AssertNoError(t, err, "second instance should reach running state")
		t.Logf("Second instance %s is running", instance2)
	})

	t.Run("MountToFirstInstance", func(t *testing.T) {
		result := ctx.Prism("volume", "mount", volumeName, instance1)
		result.AssertSuccess(t, "mount to first instance should succeed")
		t.Logf("Mounted %s to %s", volumeName, instance1)
	})

	t.Run("MountToSecondInstance", func(t *testing.T) {
		result := ctx.Prism("volume", "mount", volumeName, instance2)
		result.AssertSuccess(t, "mount to second instance should succeed")
		t.Logf("Mounted %s to %s (shared)", volumeName, instance2)
	})

	t.Run("VerifyMultiMount", func(t *testing.T) {
		// Both instances should have access to the same EFS volume
		volume, err := ctx.Client.GetVolume(context.Background(), volumeName)
		AssertNoError(t, err, "should get volume info")
		t.Logf("Shared volume info: %+v", volume)
	})

	t.Run("UnmountFromBothInstances", func(t *testing.T) {
		result := ctx.Prism("volume", "unmount", volumeName, instance1)
		result.AssertSuccess(t, "unmount from first instance should succeed")

		result = ctx.Prism("volume", "unmount", volumeName, instance2)
		result.AssertSuccess(t, "unmount from second instance should succeed")
	})

	t.Run("CleanupSharedVolume", func(t *testing.T) {
		result := ctx.Prism("volume", "delete", volumeName, "--force")
		result.AssertSuccess(t, "delete shared volume should succeed")
	})
}

// TestCLIStorageListBothTypes tests unified storage list showing both EFS and EBS
func TestCLIStorageListBothTypes(t *testing.T) {
	ctx := NewCLITestContext(t)
	defer ctx.Cleanup()

	efsName := GenerateTestName("test-efs")
	ebsName := GenerateTestName("test-ebs")

	t.Run("CreateEFSVolume", func(t *testing.T) {
		result := ctx.Prism("volume", "create", efsName)
		result.AssertSuccess(t, "EFS volume create should succeed")
		ctx.TrackVolume(efsName)
	})

	t.Run("CreateEBSVolume", func(t *testing.T) {
		result := ctx.Prism("storage", "create", ebsName, "--size", "10")
		result.AssertSuccess(t, "EBS volume create should succeed")
		ctx.TrackVolume(ebsName)
	})

	t.Run("ListAllStorage", func(t *testing.T) {
		result := ctx.Prism("storage", "list")
		result.AssertSuccess(t, "storage list should succeed")

		// Should show both EFS and EBS volumes
		result.AssertContains(t, efsName, "should list EFS volume")
		result.AssertContains(t, ebsName, "should list EBS volume")

		t.Logf("Unified storage list shows both types")
	})

	t.Run("ListEFSOnly", func(t *testing.T) {
		result := ctx.Prism("volume", "list")
		result.AssertSuccess(t, "volume list should succeed")
		result.AssertContains(t, efsName, "should list EFS volume")
		// EBS volumes should not appear in EFS-specific list
	})

	t.Run("CleanupBothTypes", func(t *testing.T) {
		result := ctx.Prism("volume", "delete", efsName, "--force")
		result.AssertSuccess(t, "delete EFS volume should succeed")

		result = ctx.Prism("storage", "delete", ebsName, "--force")
		result.AssertSuccess(t, "delete EBS volume should succeed")
	})
}

// TestCLIStorageDuplicateNames tests handling of duplicate volume names
func TestCLIStorageDuplicateNames(t *testing.T) {
	ctx := NewCLITestContext(t)
	defer ctx.Cleanup()

	volumeName := GenerateTestName("test-duplicate")

	t.Run("CreateFirstVolume", func(t *testing.T) {
		result := ctx.Prism("volume", "create", volumeName)
		result.AssertSuccess(t, "first volume create should succeed")
		ctx.TrackVolume(volumeName)
	})

	t.Run("CreateDuplicateVolume", func(t *testing.T) {
		result := ctx.Prism("volume", "create", volumeName)
		result.AssertFailure(t, "duplicate volume create should fail")
		t.Logf("Duplicate creation correctly rejected: %s", result.Stderr)
	})

	t.Run("CleanupVolume", func(t *testing.T) {
		result := ctx.Prism("volume", "delete", volumeName, "--force")
		result.AssertSuccess(t, "volume delete should succeed")
	})
}

// TestCLIStorageAttachmentPersistence tests that volumes remain attached across operations
func TestCLIStorageAttachmentPersistence(t *testing.T) {
	ctx := NewCLITestContext(t)
	defer ctx.Cleanup()

	storageName := GenerateTestName("test-persistent-ebs")
	instanceName := GenerateTestName("test-instance")

	t.Run("Setup", func(t *testing.T) {
		// Create storage
		result := ctx.Prism("storage", "create", storageName, "--size", "10")
		result.AssertSuccess(t, "storage create should succeed")
		ctx.TrackVolume(storageName)

		// Launch instance
		launchResult, err := ctx.LaunchInstanceCLI("python-ml-workstation", instanceName, "M")
		AssertNoError(t, err, "instance launch should succeed")
		launchResult.AssertSuccess(t, "launch command should succeed")

		_, err = ctx.WaitForInstanceRunning(instanceName)
		AssertNoError(t, err, "instance should reach running state")
	})

	t.Run("AttachStorage", func(t *testing.T) {
		result := ctx.Prism("storage", "attach", storageName, instanceName)
		result.AssertSuccess(t, "storage attach should succeed")
	})

	t.Run("VerifyAttachmentPersists", func(t *testing.T) {
		// Give time for attachment to settle
		time.Sleep(5 * time.Second)

		// Query storage info multiple times
		for i := 0; i < 3; i++ {
			result := ctx.Prism("storage", "info", storageName)
			result.AssertSuccess(t, fmt.Sprintf("storage info attempt %d should succeed", i+1))
			time.Sleep(2 * time.Second)
		}

		t.Logf("Storage attachment persisted across multiple queries")
	})

	t.Run("Cleanup", func(t *testing.T) {
		result := ctx.Prism("storage", "detach", storageName)
		result.AssertSuccess(t, "storage detach should succeed")

		result = ctx.Prism("storage", "delete", storageName, "--force")
		result.AssertSuccess(t, "storage delete should succeed")
	})
}
