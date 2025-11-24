// Package cli - Storage Implementation Layer
//
// ARCHITECTURE NOTE: This file contains the business logic implementation for storage commands.
// The user-facing CLI interface is defined in storage_cobra.go, which delegates to these methods.
//
// This separation follows the Facade/Adapter pattern:
//   - storage_cobra.go: CLI interface (Cobra commands, flag parsing, help text)
//   - storage_impl.go: Business logic (API calls, formatting, error handling)
//
// This architecture allows:
//   - Clean separation of concerns
//   - Reusable business logic (can be called from Cobra, TUI, or tests)
//   - Consistent API interaction patterns across all commands
//
// DO NOT REMOVE THIS FILE - it is actively used by storage_cobra.go and other components.
package cli

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/scttfrdmn/prism/pkg/api/client"
	"github.com/scttfrdmn/prism/pkg/storage"
	"github.com/scttfrdmn/prism/pkg/types"
)

// StorageCommands handles all storage management operations (implementation layer)
type StorageCommands struct {
	app *App
}

// NewStorageCommands creates storage commands handler
func NewStorageCommands(app *App) *StorageCommands {
	return &StorageCommands{app: app}
}

// Volume handles volume commands
func (sc *StorageCommands) Volume(args []string) error {
	if len(args) < 1 {
		return NewUsageError("prism volume <action> [args]", "prism volume create my-shared-data")
	}

	action := args[0]
	volumeArgs := args[1:]

	// Ensure daemon is running (auto-start if needed)
	if err := sc.app.ensureDaemonRunning(); err != nil {
		return err
	}

	switch action {
	case "create":
		return sc.volumeCreate(volumeArgs)
	case "list":
		return sc.volumeList(volumeArgs)
	case "info":
		return sc.volumeInfo(volumeArgs)
	case "delete":
		return sc.volumeDelete(volumeArgs)
	case "mount":
		return sc.volumeMount(volumeArgs)
	case "unmount":
		return sc.volumeUnmount(volumeArgs)
	default:
		return NewValidationError("volume action", action, "create, list, info, delete, mount, unmount")
	}
}

func (sc *StorageCommands) volumeCreate(args []string) error {
	if len(args) < 1 {
		return NewUsageError("prism volume create <name> [options]", "prism volume create my-shared-data --performance generalPurpose")
	}

	req := types.VolumeCreateRequest{
		Name: args[0],
	}

	// Parse options
	for i := 1; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--performance" && i+1 < len(args):
			req.PerformanceMode = args[i+1]
			i++
		case arg == "--throughput" && i+1 < len(args):
			req.ThroughputMode = args[i+1]
			i++
		case arg == "--region" && i+1 < len(args):
			req.Region = args[i+1]
			i++
		default:
			return NewValidationError("volume option", arg, "--performance, --throughput, --region")
		}
	}

	volume, err := sc.app.apiClient.CreateVolume(sc.app.ctx, req)
	if err != nil {
		return WrapAPIError("create shared storage "+req.Name, err)
	}

	fmt.Printf("%s\n", FormatSuccessMessage("Created Shared Storage", volume.Name, fmt.Sprintf("(%s)", volume.FileSystemID)))
	return nil
}

func (sc *StorageCommands) volumeList(_ []string) error {
	volumes, err := sc.app.apiClient.ListVolumes(sc.app.ctx)
	if err != nil {
		return WrapAPIError("list shared storage", err)
	}

	if len(volumes) == 0 {
		fmt.Println("No shared storage volumes found. Create one with 'prism volume create'.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, TabWriterMinWidth, TabWriterTabWidth, TabWriterPadding, TabWriterPadChar, TabWriterFlags)
	_, _ = fmt.Fprintln(w, "NAME\tFILESYSTEM ID\tSTATE\tSIZE\tCOST/MONTH")

	for _, volume := range volumes {
		var sizeGB float64
		if volume.SizeBytes != nil {
			sizeGB = float64(*volume.SizeBytes) / BytesToGB
		}
		costMonth := sizeGB * volume.EstimatedCostGB
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%.1f GB\t$%.2f\n",
			volume.Name,
			volume.FileSystemID,
			strings.ToUpper(volume.State),
			sizeGB,
			costMonth,
		)
	}
	_ = w.Flush()

	return nil
}

func (sc *StorageCommands) volumeInfo(args []string) error {
	if len(args) < 1 {
		return NewUsageError("prism volume info <name>", "prism volume info my-shared-data")
	}

	name := args[0]
	volume, err := sc.app.apiClient.GetVolume(sc.app.ctx, name)
	if err != nil {
		return WrapAPIError("get volume info for "+name, err)
	}

	fmt.Printf("📁 Shared Storage: %s\n", volume.Name)
	fmt.Printf("   Filesystem ID: %s\n", volume.FileSystemID)
	fmt.Printf("   State: %s\n", strings.ToUpper(volume.State))
	fmt.Printf("   Region: %s\n", volume.Region)
	fmt.Printf("   Performance Mode: %s\n", volume.PerformanceMode)
	fmt.Printf("   Throughput Mode: %s\n", volume.ThroughputMode)
	if volume.SizeBytes != nil {
		sizeGB := float64(*volume.SizeBytes) / BytesToGB
		fmt.Printf("   Size: %.1f GB\n", sizeGB)
		fmt.Printf("   Cost: $%.2f/month\n", sizeGB*volume.EstimatedCostGB)
	}
	fmt.Printf("   Created: %s\n", volume.CreationTime.Format(StandardDateFormat))
	fmt.Printf("   AWS Service: %s\n", volume.GetTechnicalType())

	return nil
}

func (sc *StorageCommands) volumeDelete(args []string) error {
	if len(args) < 1 {
		return NewUsageError("prism volume delete <name>", "prism volume delete my-shared-data")
	}

	name := args[0]
	err := sc.app.apiClient.DeleteVolume(sc.app.ctx, name)
	if err != nil {
		return WrapAPIError("delete shared storage "+name, err)
	}

	fmt.Printf("%s\n", FormatProgressMessage("Deleting Shared Storage", name))
	return nil
}

func (sc *StorageCommands) volumeMount(args []string) error {
	if len(args) < 2 {
		return NewUsageError("prism volume mount <volume-name> <workspace-name> [mount-point]", "prism volume mount my-shared-data my-workspace")
	}

	volumeName := args[0]
	instanceName := args[1]

	// Default mount point
	mountPoint := DefaultMountPointPrefix + volumeName
	if len(args) >= 3 {
		mountPoint = args[2]
	}

	err := sc.app.apiClient.MountVolume(sc.app.ctx, volumeName, instanceName, mountPoint)
	if err != nil {
		return WrapAPIError("mount volume "+volumeName+" to "+instanceName, err)
	}

	fmt.Printf("%s\n", FormatProgressMessage("Mounting Shared Storage", fmt.Sprintf("'%s' to '%s' at %s", volumeName, instanceName, mountPoint)))
	return nil
}

func (sc *StorageCommands) volumeUnmount(args []string) error {
	if len(args) < 2 {
		return NewUsageError("prism volume unmount <volume-name> <workspace-name>", "prism volume unmount my-shared-data my-workspace")
	}

	volumeName := args[0]
	instanceName := args[1]

	err := sc.app.apiClient.UnmountVolume(sc.app.ctx, volumeName, instanceName)
	if err != nil {
		return WrapAPIError("unmount volume "+volumeName+" from "+instanceName, err)
	}

	fmt.Printf("%s\n", FormatProgressMessage("Unmounting Shared Storage", fmt.Sprintf("'%s' from '%s'", volumeName, instanceName)))
	return nil
}

// Storage handles storage commands
func (sc *StorageCommands) Storage(args []string) error {
	if len(args) < 1 {
		return NewUsageError("prism storage <action> [args]", "prism storage create my-data 100GB")
	}

	action := args[0]
	storageArgs := args[1:]

	// Ensure daemon is running (auto-start if needed)
	if err := sc.app.ensureDaemonRunning(); err != nil {
		return err
	}

	switch action {
	case "create":
		return sc.storageCreate(storageArgs)
	case "list":
		return sc.storageList(storageArgs)
	case "info":
		return sc.storageInfo(storageArgs)
	case "attach":
		return sc.storageAttach(storageArgs)
	case "detach":
		return sc.storageDetach(storageArgs)
	case "delete":
		return sc.storageDelete(storageArgs)
	case "upload":
		return sc.storageUpload(storageArgs)
	case "download":
		return sc.storageDownload(storageArgs)
	case "transfers":
		return sc.storageTransfersList(storageArgs)
	case "transfer":
		return sc.storageTransferStatus(storageArgs)
	case "cancel":
		return sc.storageTransferCancel(storageArgs)
	default:
		return NewValidationError("storage action", action, "create, list, info, attach, detach, delete, upload, download, transfers, transfer, cancel")
	}
}

func (sc *StorageCommands) storageCreate(args []string) error {
	if len(args) < 2 {
		return NewUsageError("prism storage create <name> <size> [type]", "prism storage create my-data 100GB gp3")
	}

	req := types.StorageCreateRequest{
		Name:       args[0],
		Size:       args[1],
		VolumeType: DefaultVolumeType, // default
	}

	// Parse volume type and options
	optionStartIndex := 2
	if len(args) > 2 && !strings.HasPrefix(args[2], "--") {
		req.VolumeType = args[2]
		optionStartIndex = 3
	}

	// Parse additional options
	for i := optionStartIndex; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--region" && i+1 < len(args):
			req.Region = args[i+1]
			i++
		default:
			return NewValidationError("storage option", arg, "--region")
		}
	}

	volume, err := sc.app.apiClient.CreateStorage(sc.app.ctx, req)
	if err != nil {
		return WrapAPIError("create workspace storage "+req.Name, err)
	}

	sizeStr := "unknown"
	if volume.SizeGB != nil {
		sizeStr = fmt.Sprintf("%d GB", *volume.SizeGB)
	}
	fmt.Printf("%s\n", FormatSuccessMessage("Created Workspace Storage", volume.Name, fmt.Sprintf("(%s) - %s %s", volume.VolumeID, sizeStr, volume.VolumeType)))
	return nil
}

func (sc *StorageCommands) storageList(_ []string) error {
	volumes, err := sc.app.apiClient.ListStorage(sc.app.ctx)
	if err != nil {
		return WrapAPIError("list storage volumes", err)
	}

	if len(volumes) == 0 {
		fmt.Println("No storage volumes found. Create one with 'prism storage create'.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, TabWriterMinWidth, TabWriterTabWidth, TabWriterPadding, TabWriterPadChar, TabWriterFlags)
	_, _ = fmt.Fprintln(w, "NAME\tTYPE\tSTATE\tSIZE\tDETAILS\tCOST/MONTH")

	for _, volume := range volumes {
		var sizeStr, detailsStr string
		var costMonth float64

		if volume.IsWorkspace() {
			// Workspace Storage (EBS)
			if volume.SizeGB != nil {
				sizeStr = fmt.Sprintf("%d GB", *volume.SizeGB)
				costMonth = float64(*volume.SizeGB) * volume.EstimatedCostGB
			}
			if volume.VolumeType != "" {
				detailsStr = volume.VolumeType
			}
			if volume.AttachedTo != "" {
				detailsStr += fmt.Sprintf(" → %s", volume.AttachedTo)
			}
		} else if volume.IsShared() {
			// Shared Storage (EFS)
			if volume.SizeBytes != nil {
				sizeGB := float64(*volume.SizeBytes) / BytesToGB
				sizeStr = fmt.Sprintf("%.1f GB", sizeGB)
				costMonth = sizeGB * volume.EstimatedCostGB
			}
			detailsStr = volume.PerformanceMode
		}

		if detailsStr == "" {
			detailsStr = "-"
		}

		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t$%.2f\n",
			volume.Name,
			volume.GetDisplayType(),
			strings.ToUpper(volume.State),
			sizeStr,
			detailsStr,
			costMonth,
		)
	}
	_ = w.Flush()

	return nil
}

func (sc *StorageCommands) storageInfo(args []string) error {
	if len(args) < 1 {
		return NewUsageError("prism storage info <name>", "prism storage info my-data")
	}

	name := args[0]
	volume, err := sc.app.apiClient.GetStorage(sc.app.ctx, name)
	if err != nil {
		return WrapAPIError("get storage info for "+name, err)
	}

	// Display common fields
	icon := "💾"
	if volume.IsShared() {
		icon = "📁"
	}
	fmt.Printf("%s %s: %s\n", icon, volume.GetDisplayType(), volume.Name)
	fmt.Printf("   State: %s\n", strings.ToUpper(volume.State))
	fmt.Printf("   Region: %s\n", volume.Region)

	// Display type-specific fields
	if volume.IsWorkspace() {
		// Workspace Storage (EBS) fields
		fmt.Printf("   Volume ID: %s\n", volume.VolumeID)
		if volume.SizeGB != nil {
			fmt.Printf("   Size: %d GB\n", *volume.SizeGB)
		}
		fmt.Printf("   Type: %s\n", volume.VolumeType)
		if volume.IOPS != nil && *volume.IOPS > 0 {
			fmt.Printf("   IOPS: %d\n", *volume.IOPS)
		}
		if volume.Throughput != nil && *volume.Throughput > 0 {
			fmt.Printf("   Throughput: %d MB/s\n", *volume.Throughput)
		}
		if volume.AttachedTo != "" {
			fmt.Printf("   Attached to: %s\n", volume.AttachedTo)
		}
		if volume.SizeGB != nil {
			fmt.Printf("   Cost: $%.2f/month\n", float64(*volume.SizeGB)*volume.EstimatedCostGB)
		}
	} else if volume.IsShared() {
		// Shared Storage (EFS) fields
		fmt.Printf("   Filesystem ID: %s\n", volume.FileSystemID)
		if volume.SizeBytes != nil {
			sizeGB := float64(*volume.SizeBytes) / BytesToGB
			fmt.Printf("   Size: %.1f GB\n", sizeGB)
			fmt.Printf("   Cost: $%.2f/month\n", sizeGB*volume.EstimatedCostGB)
		}
		fmt.Printf("   Performance Mode: %s\n", volume.PerformanceMode)
		fmt.Printf("   Throughput Mode: %s\n", volume.ThroughputMode)
	}

	fmt.Printf("   Created: %s\n", volume.CreationTime.Format(StandardDateFormat))
	fmt.Printf("   AWS Service: %s\n", volume.GetTechnicalType())

	return nil
}

func (sc *StorageCommands) storageAttach(args []string) error {
	if len(args) < 2 {
		return NewUsageError("prism storage attach <volume> <workspace>", "prism storage attach my-data my-workspace")
	}

	volumeName := args[0]
	instanceName := args[1]

	err := sc.app.apiClient.AttachStorage(sc.app.ctx, volumeName, instanceName)
	if err != nil {
		return WrapAPIError("attach storage "+volumeName+" to "+instanceName, err)
	}

	fmt.Printf("%s\n", FormatProgressMessage("Attaching volume", fmt.Sprintf("%s to workspace %s", volumeName, instanceName)))
	return nil
}

func (sc *StorageCommands) storageDetach(args []string) error {
	if len(args) < 1 {
		return NewUsageError("prism storage detach <volume>", "prism storage detach my-data")
	}

	volumeName := args[0]

	err := sc.app.apiClient.DetachStorage(sc.app.ctx, volumeName)
	if err != nil {
		return WrapAPIError("detach storage "+volumeName, err)
	}

	fmt.Printf("%s\n", FormatProgressMessage("Detaching volume", volumeName))
	return nil
}

func (sc *StorageCommands) storageDelete(args []string) error {
	if len(args) < 1 {
		return NewUsageError("prism storage delete <name>", "prism storage delete my-data")
	}

	name := args[0]
	err := sc.app.apiClient.DeleteStorage(sc.app.ctx, name)
	if err != nil {
		return WrapAPIError("delete storage "+name, err)
	}

	fmt.Printf("%s\n", FormatProgressMessage("Deleting storage", name))
	return nil
}

// S3 Transfer operations (Issue #64)

func (sc *StorageCommands) storageUpload(args []string) error {
	if len(args) < 3 {
		return NewUsageError("prism storage upload <local-file> <s3-bucket> <s3-key>", "prism storage upload data.csv my-bucket research/data.csv")
	}

	localPath := args[0]
	s3Bucket := args[1]
	s3Key := args[2]

	req := client.TransferRequest{
		Type:      "upload",
		LocalPath: localPath,
		S3Bucket:  s3Bucket,
		S3Key:     s3Key,
		Options: &storage.TransferOptions{
			PartSize:    5 * 1024 * 1024, // 5MB chunks
			Concurrency: 5,
		},
	}

	spinner := NewSpinner(fmt.Sprintf("Uploading '%s' to s3://%s/%s", localPath, s3Bucket, s3Key))
	spinner.Start()

	response, err := sc.app.apiClient.StartTransfer(sc.app.ctx, req)
	if err != nil {
		spinner.Stop()
		return WrapAPIError("start upload", err)
	}

	// Poll for progress
	transferID := response.TransferID
	for {
		progress, err := sc.app.apiClient.GetTransferStatus(sc.app.ctx, transferID)
		if err != nil {
			spinner.Stop()
			return WrapAPIError("get transfer status", err)
		}

		if progress.Status == storage.TransferStatusCompleted {
			spinner.StopWithMessage(fmt.Sprintf("✅ Upload complete: %s (%.2f MB)", localPath, float64(progress.TransferredBytes)/(1024*1024)))
			return nil
		}

		if progress.Status == storage.TransferStatusFailed {
			spinner.Stop()
			return fmt.Errorf("upload failed: %s", progress.Error)
		}

		// Update spinner with progress
		if progress.TotalBytes > 0 {
			pct := (float64(progress.TransferredBytes) / float64(progress.TotalBytes)) * 100
			spinner.UpdateMessage(fmt.Sprintf("Uploading '%s' (%.1f%%)", localPath, pct))
		}

		time.Sleep(500 * time.Millisecond)
	}
}

func (sc *StorageCommands) storageDownload(args []string) error {
	if len(args) < 3 {
		return NewUsageError("prism storage download <s3-bucket> <s3-key> <local-file>", "prism storage download my-bucket research/data.csv data.csv")
	}

	s3Bucket := args[0]
	s3Key := args[1]
	localPath := args[2]

	req := client.TransferRequest{
		Type:      "download",
		LocalPath: localPath,
		S3Bucket:  s3Bucket,
		S3Key:     s3Key,
		Options: &storage.TransferOptions{
			PartSize:    5 * 1024 * 1024, // 5MB chunks
			Concurrency: 5,
		},
	}

	spinner := NewSpinner(fmt.Sprintf("Downloading s3://%s/%s to '%s'", s3Bucket, s3Key, localPath))
	spinner.Start()

	response, err := sc.app.apiClient.StartTransfer(sc.app.ctx, req)
	if err != nil {
		spinner.Stop()
		return WrapAPIError("start download", err)
	}

	// Poll for progress
	transferID := response.TransferID
	for {
		progress, err := sc.app.apiClient.GetTransferStatus(sc.app.ctx, transferID)
		if err != nil {
			spinner.Stop()
			return WrapAPIError("get transfer status", err)
		}

		if progress.Status == storage.TransferStatusCompleted {
			spinner.StopWithMessage(fmt.Sprintf("✅ Download complete: %s (%.2f MB)", localPath, float64(progress.TransferredBytes)/(1024*1024)))
			return nil
		}

		if progress.Status == storage.TransferStatusFailed {
			spinner.Stop()
			return fmt.Errorf("download failed: %s", progress.Error)
		}

		// Update spinner with progress
		if progress.TotalBytes > 0 {
			pct := (float64(progress.TransferredBytes) / float64(progress.TotalBytes)) * 100
			spinner.UpdateMessage(fmt.Sprintf("Downloading s3://%s/%s (%.1f%%)", s3Bucket, s3Key, pct))
		}

		time.Sleep(500 * time.Millisecond)
	}
}

func (sc *StorageCommands) storageTransfersList(_ []string) error {
	transfers, err := sc.app.apiClient.ListTransfers(sc.app.ctx)
	if err != nil {
		return WrapAPIError("list transfers", err)
	}

	if len(transfers) == 0 {
		fmt.Println("No active transfers.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, TabWriterMinWidth, TabWriterTabWidth, TabWriterPadding, TabWriterPadChar, TabWriterFlags)
	_, _ = fmt.Fprintln(w, "TRANSFER ID\tTYPE\tSTATUS\tPROGRESS\tSPEED")

	for _, t := range transfers {
		var progress string
		if t.TotalBytes > 0 {
			pct := (float64(t.TransferredBytes) / float64(t.TotalBytes)) * 100
			progress = fmt.Sprintf("%.1f%%", pct)
		} else {
			progress = "unknown"
		}

		var speed string
		if t.BytesPerSecond > 0 {
			speed = fmt.Sprintf("%.2f MB/s", float64(t.BytesPerSecond)/(1024*1024))
		} else {
			speed = "-"
		}

		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			t.TransferID,
			strings.ToUpper(string(t.Type)),
			strings.ToUpper(string(t.Status)),
			progress,
			speed,
		)
	}
	_ = w.Flush()

	return nil
}

func (sc *StorageCommands) storageTransferStatus(args []string) error {
	if len(args) < 1 {
		return NewUsageError("prism storage transfer <transfer-id>", "prism storage transfer abc123")
	}

	transferID := args[0]
	progress, err := sc.app.apiClient.GetTransferStatus(sc.app.ctx, transferID)
	if err != nil {
		return WrapAPIError("get transfer status for "+transferID, err)
	}

	fmt.Printf("📦 Transfer: %s\n", progress.TransferID)
	fmt.Printf("   Type: %s\n", strings.ToUpper(string(progress.Type)))
	fmt.Printf("   Status: %s\n", strings.ToUpper(string(progress.Status)))
	fmt.Printf("   S3 Location: s3://%s/%s\n", progress.S3Bucket, progress.S3Key)
	fmt.Printf("   Local Path: %s\n", progress.FilePath)

	if progress.TotalBytes > 0 {
		pct := (float64(progress.TransferredBytes) / float64(progress.TotalBytes)) * 100
		fmt.Printf("   Progress: %.2f MB / %.2f MB (%.1f%%)\n",
			float64(progress.TransferredBytes)/(1024*1024),
			float64(progress.TotalBytes)/(1024*1024),
			pct)
	}

	if progress.BytesPerSecond > 0 {
		fmt.Printf("   Speed: %.2f MB/s\n", float64(progress.BytesPerSecond)/(1024*1024))
	}

	if progress.Status == storage.TransferStatusFailed && progress.Error != "" {
		fmt.Printf("   Error: %s\n", progress.Error)
	}

	fmt.Printf("   Started: %s\n", progress.StartTime.Format(StandardDateFormat))
	if !progress.LastUpdate.IsZero() {
		fmt.Printf("   Last Update: %s\n", progress.LastUpdate.Format(StandardDateFormat))
	}
	if progress.Status == storage.TransferStatusCompleted || progress.Status == storage.TransferStatusFailed {
		duration := time.Since(progress.StartTime)
		fmt.Printf("   Duration: %s\n", duration.Round(time.Second))
	}

	return nil
}

func (sc *StorageCommands) storageTransferCancel(args []string) error {
	if len(args) < 1 {
		return NewUsageError("prism storage cancel <transfer-id>", "prism storage cancel abc123")
	}

	transferID := args[0]
	err := sc.app.apiClient.CancelTransfer(sc.app.ctx, transferID)
	if err != nil {
		return WrapAPIError("cancel transfer "+transferID, err)
	}

	fmt.Printf("%s\n", FormatProgressMessage("Cancelling transfer", transferID))
	return nil
}
