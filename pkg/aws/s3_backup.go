package aws

// S3/SSM-based selective file backup support (Issue #478)
//
// This file implements Phase 1 of file-level S3 backups:
//   - CreateS3FileBackup:  launches an async SSM command (aws s3 sync) and stores
//                          metadata in SSM Parameter Store. Returns immediately.
//   - GetS3BackupInfo:     reads metadata; lazily refreshes state from SSM command.
//   - ListS3FileBackups:   enumerates all backups under /prism/backups/.
//   - DeleteS3FileBackup:  removes the SSM parameter (S3 data is intentionally kept).
//
// S3 bucket naming: prism-backups-{account-id}-{region}  (deterministic, no creation needed
// beyond the one-time IAM/infra step documented in the issue).

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	ssmtypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/scttfrdmn/prism/pkg/types"
)

const s3BackupParameterPrefix = "/prism/backups"

// S3BackupMetadata stores backup state in SSM Parameter Store.
type S3BackupMetadata struct {
	BackupName   string     `json:"backup_name"`
	InstanceName string     `json:"instance_name"`
	InstanceID   string     `json:"instance_id"`
	SSMCommandID string     `json:"ssm_command_id"`
	S3Bucket     string     `json:"s3_bucket"`
	S3Prefix     string     `json:"s3_prefix"`
	IncludePaths []string   `json:"include_paths,omitempty"`
	ExcludePaths []string   `json:"exclude_paths,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	State        string     `json:"state"` // "creating", "available", "error"
	ErrorMessage string     `json:"error_message,omitempty"`
	SizeBytes    int64      `json:"size_bytes,omitempty"`
	FileCount    int        `json:"file_count,omitempty"`
}

// GetBackupS3Bucket returns the S3 bucket name used for file-level backups.
//
// Pattern: prism-backups-{account-id}-{region}
// The bucket must exist and the instance IAM role needs s3:PutObject, s3:GetObject,
// s3:ListBucket, and s3:DeleteObject permissions.
func (m *Manager) GetBackupS3Bucket(ctx context.Context) (string, error) {
	output, err := m.sts.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return "", fmt.Errorf("failed to get AWS account ID: %w", err)
	}
	return fmt.Sprintf("prism-backups-%s-%s", *output.Account, m.region), nil
}

// CreateS3FileBackup starts an async S3 file backup on a running instance via SSM.
//
// The method launches `aws s3 sync` on the instance for each path in IncludePaths
// (defaulting to /home and /data when empty), writes backup metadata to SSM Parameter
// Store, and returns immediately with state "creating".
func (m *Manager) CreateS3FileBackup(req types.BackupCreateRequest) (*types.BackupCreateResult, error) {
	ctx := context.Background()

	// Load state to find the instance (same pattern as CreateInstanceAMISnapshot)
	st, err := m.stateManager.LoadState()
	if err != nil {
		return nil, fmt.Errorf("failed to load state: %w", err)
	}

	instanceData, exists := st.Instances[req.InstanceName]
	if !exists {
		return nil, fmt.Errorf("instance '%s' not found", req.InstanceName)
	}
	if instanceData.State != "running" {
		return nil, fmt.Errorf("instance '%s' must be running for S3 file backup (state: %s)",
			req.InstanceName, instanceData.State)
	}

	bucket, err := m.GetBackupS3Bucket(ctx)
	if err != nil {
		return nil, err
	}
	s3Prefix := fmt.Sprintf("%s/%s", req.InstanceName, req.BackupName)

	script := buildS3BackupScript(bucket, s3Prefix, req.IncludePaths, req.ExcludePaths)

	// Launch SSM command without waiting — fire-and-forget
	cmdOutput, err := m.ssm.SendCommand(ctx, &ssm.SendCommandInput{
		InstanceIds:  []string{instanceData.ID},
		DocumentName: aws.String("AWS-RunShellScript"),
		Parameters:   map[string][]string{"commands": {script}},
		Comment: aws.String(fmt.Sprintf(
			"Prism S3 backup: %s -> s3://%s/%s/", req.BackupName, bucket, s3Prefix)),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to launch backup command via SSM: %w", err)
	}
	commandID := *cmdOutput.Command.CommandId

	metadata := &S3BackupMetadata{
		BackupName:   req.BackupName,
		InstanceName: req.InstanceName,
		InstanceID:   instanceData.ID,
		SSMCommandID: commandID,
		S3Bucket:     bucket,
		S3Prefix:     s3Prefix,
		IncludePaths: req.IncludePaths,
		ExcludePaths: req.ExcludePaths,
		CreatedAt:    time.Now(),
		State:        "creating",
	}
	if err := m.putS3BackupMetadata(ctx, metadata); err != nil {
		return nil, fmt.Errorf("failed to store backup metadata: %w", err)
	}

	return &types.BackupCreateResult{
		BackupName:                 req.BackupName,
		BackupID:                   commandID,
		SourceInstance:             req.InstanceName,
		BackupType:                 "full",
		StorageType:                "s3",
		StorageLocation:            fmt.Sprintf("s3://%s/%s/", bucket, s3Prefix),
		EstimatedCompletionMinutes: 5,
		StorageCostMonthly:         0,
		CreatedAt:                  time.Now(),
		Encrypted:                  false,
		Message:                    fmt.Sprintf("S3 file backup started (SSM command: %s)", commandID),
	}, nil
}

// GetS3BackupInfo retrieves the current state of an S3 file backup.
//
// If the backup is still "creating", the method queries SSM for the command result
// and lazily updates the SSM parameter so subsequent calls are cheaper.
func (m *Manager) GetS3BackupInfo(backupName string) (*types.BackupInfo, error) {
	ctx := context.Background()

	metadata, err := m.getS3BackupMetadata(ctx, backupName)
	if err != nil {
		return nil, err
	}

	if metadata.State == "creating" {
		invocation, ssmErr := m.ssm.GetCommandInvocation(ctx, &ssm.GetCommandInvocationInput{
			CommandId:  aws.String(metadata.SSMCommandID),
			InstanceId: aws.String(metadata.InstanceID),
		})
		if ssmErr == nil {
			status := string(invocation.Status)
			switch status {
			case "Success":
				metadata.State = "available"
				now := time.Now()
				metadata.CompletedAt = &now
				_ = m.putS3BackupMetadata(ctx, metadata) // best-effort persist
			case "Failed":
				metadata.State = "error"
				if invocation.StandardErrorContent != nil {
					metadata.ErrorMessage = *invocation.StandardErrorContent
				}
				_ = m.putS3BackupMetadata(ctx, metadata)
			case "Cancelled", "TimedOut":
				metadata.State = "error"
				metadata.ErrorMessage = fmt.Sprintf("SSM command %s", strings.ToLower(status))
				_ = m.putS3BackupMetadata(ctx, metadata)
			}
		}
	}

	return s3BackupMetadataToInfo(metadata), nil
}

// ListS3FileBackups returns all S3 file backups stored in SSM Parameter Store.
func (m *Manager) ListS3FileBackups() ([]types.BackupInfo, error) {
	ctx := context.Background()

	output, err := m.ssm.GetParametersByPath(ctx, &ssm.GetParametersByPathInput{
		Path:      aws.String(s3BackupParameterPrefix + "/"),
		Recursive: aws.Bool(true),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list S3 backups: %w", err)
	}

	var backups []types.BackupInfo
	for _, param := range output.Parameters {
		var meta S3BackupMetadata
		if err := json.Unmarshal([]byte(*param.Value), &meta); err != nil {
			continue // skip malformed entries
		}
		backups = append(backups, *s3BackupMetadataToInfo(&meta))
	}
	return backups, nil
}

// DeleteS3FileBackup removes an S3 backup's SSM metadata entry.
//
// S3 objects are intentionally NOT deleted to prevent accidental data loss.
// Users can purge s3://{bucket}/{prefix}/ manually when desired.
func (m *Manager) DeleteS3FileBackup(backupName string) (*types.BackupDeleteResult, error) {
	ctx := context.Background()

	metadata, err := m.getS3BackupMetadata(ctx, backupName)
	if err != nil {
		return nil, err
	}

	if _, err := m.ssm.DeleteParameter(ctx, &ssm.DeleteParameterInput{
		Name: aws.String(s3BackupParameterName(backupName)),
	}); err != nil {
		return nil, fmt.Errorf("failed to delete backup metadata: %w", err)
	}

	return &types.BackupDeleteResult{
		BackupName:            backupName,
		BackupID:              metadata.SSMCommandID,
		StorageType:           "s3",
		StorageLocation:       fmt.Sprintf("s3://%s/%s/", metadata.S3Bucket, metadata.S3Prefix),
		DeletedSizeBytes:      metadata.SizeBytes,
		StorageSavingsMonthly: estimateS3MonthlyCost(metadata.SizeBytes),
		DeletedAt:             time.Now(),
	}, nil
}

// buildS3BackupScript generates the shell script that runs on the instance to sync files to S3.
func buildS3BackupScript(bucket, s3Prefix string, includePaths, excludePaths []string) string {
	var sb strings.Builder
	sb.WriteString("#!/bin/bash\nset -e\n\n")

	if len(includePaths) == 0 {
		// Default: back up home directories and /data if present
		includePaths = []string{"/home", "/data"}
	}

	var excludeArgs strings.Builder
	for _, p := range excludePaths {
		// Quote the path to avoid shell injection via excludePaths values
		fmt.Fprintf(&excludeArgs, " --exclude %q", p)
	}

	fmt.Fprintf(&sb, "BUCKET=%q\n", bucket)
	fmt.Fprintf(&sb, "PREFIX=%q\n", s3Prefix)
	sb.WriteString("\necho '[prism-backup] Starting S3 file backup...'\n")

	for _, srcPath := range includePaths {
		// Keep directory structure relative to filesystem root in the S3 key
		relPath := strings.TrimLeft(srcPath, "/")
		if relPath == "" {
			relPath = "root"
		}
		s3Dest := fmt.Sprintf(`s3://"$BUCKET"/"$PREFIX"/%s/`, relPath)

		fmt.Fprintf(&sb,
			"if [ -d %q ] || [ -f %q ]; then\n"+
				"  echo '[prism-backup] Syncing %s'\n"+
				"  aws s3 sync %q %s%s\n"+
				"fi\n",
			srcPath, srcPath,
			srcPath,
			srcPath, s3Dest, excludeArgs.String(),
		)
	}

	sb.WriteString("\necho '[prism-backup] Backup complete.'\n")
	return sb.String()
}

// putS3BackupMetadata writes backup metadata to SSM Parameter Store.
func (m *Manager) putS3BackupMetadata(ctx context.Context, metadata *S3BackupMetadata) error {
	data, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("marshal backup metadata: %w", err)
	}
	_, err = m.ssm.PutParameter(ctx, &ssm.PutParameterInput{
		Name:      aws.String(s3BackupParameterName(metadata.BackupName)),
		Value:     aws.String(string(data)),
		Type:      ssmtypes.ParameterTypeString,
		Overwrite: aws.Bool(true),
	})
	return err
}

// getS3BackupMetadata reads backup metadata from SSM Parameter Store.
func (m *Manager) getS3BackupMetadata(ctx context.Context, backupName string) (*S3BackupMetadata, error) {
	output, err := m.ssm.GetParameter(ctx, &ssm.GetParameterInput{
		Name: aws.String(s3BackupParameterName(backupName)),
	})
	if err != nil {
		return nil, fmt.Errorf("S3 backup '%s' not found: %w", backupName, err)
	}
	var metadata S3BackupMetadata
	if err := json.Unmarshal([]byte(*output.Parameter.Value), &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse backup metadata: %w", err)
	}
	return &metadata, nil
}

// s3BackupParameterName returns the SSM parameter path for a named backup.
func s3BackupParameterName(backupName string) string {
	return fmt.Sprintf("%s/%s", s3BackupParameterPrefix, backupName)
}

// s3BackupMetadataToInfo converts internal S3BackupMetadata to the public BackupInfo type.
func s3BackupMetadataToInfo(m *S3BackupMetadata) *types.BackupInfo {
	info := &types.BackupInfo{
		BackupName:         m.BackupName,
		BackupID:           m.SSMCommandID,
		SourceInstance:     m.InstanceName,
		SourceInstanceId:   m.InstanceID,
		BackupType:         "full",
		StorageType:        "s3",
		StorageLocation:    fmt.Sprintf("s3://%s/%s/", m.S3Bucket, m.S3Prefix),
		State:              m.State,
		SizeBytes:          m.SizeBytes,
		CompressedBytes:    m.SizeBytes, // aws s3 sync does not compress
		FileCount:          m.FileCount,
		IncludedPaths:      m.IncludePaths,
		ExcludedPaths:      m.ExcludePaths,
		Encrypted:          false,
		StorageCostMonthly: estimateS3MonthlyCost(m.SizeBytes),
		CreatedAt:          m.CreatedAt,
		CompletedAt:        m.CompletedAt,
	}
	if m.ErrorMessage != "" {
		info.Metadata = map[string]string{"error": m.ErrorMessage}
	}
	return info
}

// estimateS3MonthlyCost estimates the monthly S3 Standard storage cost (~$0.023/GB/month).
func estimateS3MonthlyCost(sizeBytes int64) float64 {
	return float64(sizeBytes) / (1024 * 1024 * 1024) * 0.023
}
