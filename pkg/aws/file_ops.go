// Package aws provides AWS cloud operations for Prism.
//
// file_ops.go implements SSM-based file transfer between the local machine and
// running EC2 instances. Files are relayed through a temporary S3 bucket because
// SSM Run Command does not support direct binary transfer.
//
// Transfer flow (push):
//
//	local file → S3 temp bucket (aws s3 cp) → instance (SSM aws s3 cp) → cleanup
//
// Transfer flow (pull):
//
//	instance (SSM aws s3 cp → temp bucket) → local download (aws s3 cp) → cleanup
//
// Issue #30 / sub-issue 30a
package aws

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/google/uuid"
)

// newS3Client returns a fresh S3 client from the manager's AWS config.
func (m *Manager) newS3Client() *s3.Client {
	return s3.NewFromConfig(m.cfg)
}

// FileOpResult describes the outcome of a push or pull operation.
type FileOpResult struct {
	Path        string    `json:"path"`
	SizeBytes   int64     `json:"size_bytes,omitempty"`
	Status      string    `json:"status"` // "ok" | "error"
	Message     string    `json:"message,omitempty"`
	CompletedAt time.Time `json:"completed_at"`
}

// RemoteFileEntry describes a single file or directory on the remote instance.
type RemoteFileEntry struct {
	Path        string    `json:"path"`
	SizeBytes   int64     `json:"size_bytes"`
	IsDir       bool      `json:"is_dir"`
	ModifiedAt  time.Time `json:"modified_at"`
	Permissions string    `json:"permissions"`
}

// PushFile uploads a local file to a running instance via the S3 temp bucket.
//
// The instance must be running and have the SSM agent installed.  The instance
// IAM role must allow s3:GetObject on the temp bucket (prism-temp-{acct}-{region}).
func (m *Manager) PushFile(ctx context.Context, instanceName, localPath, remotePath string) (*FileOpResult, error) {
	st, err := m.stateManager.LoadState()
	if err != nil {
		return nil, fmt.Errorf("failed to load state: %w", err)
	}

	instanceData, exists := st.Instances[instanceName]
	if !exists {
		return nil, fmt.Errorf("instance '%s' not found", instanceName)
	}
	if instanceData.State != "running" {
		return nil, fmt.Errorf("instance '%s' must be running for file transfer (state: %s)",
			instanceName, instanceData.State)
	}

	// Open local file to get its size
	fi, err := os.Stat(localPath)
	if err != nil {
		return nil, fmt.Errorf("local file not found: %w", err)
	}

	// Determine S3 key for temp storage
	tempBucket, err := m.GetTemporaryS3Bucket(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get temp bucket: %w", err)
	}
	transferKey := fmt.Sprintf("prism-transfer/%s/%s", uuid.New().String(), filepath.Base(localPath))

	// Upload local file to S3 temp bucket
	log.Printf("[file_ops] uploading %s → s3://%s/%s", localPath, tempBucket, transferKey)
	f, err := os.Open(localPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open local file: %w", err)
	}
	defer f.Close()

	s3Client := m.newS3Client()
	_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(tempBucket),
		Key:    aws.String(transferKey),
		Body:   f,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload to S3 temp bucket: %w", err)
	}

	// Ensure cleanup of the temp object regardless of outcome
	defer func() {
		cleanCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_, delErr := s3Client.DeleteObject(cleanCtx, &s3.DeleteObjectInput{
			Bucket: aws.String(tempBucket),
			Key:    aws.String(transferKey),
		})
		if delErr != nil {
			log.Printf("[file_ops] warning: failed to delete temp object s3://%s/%s: %v",
				tempBucket, transferKey, delErr)
		}
	}()

	// Run SSM command on instance to download from S3
	script := fmt.Sprintf(`#!/bin/bash
set -e
mkdir -p "$(dirname '%s')"
aws s3 cp "s3://%s/%s" "%s"
echo "OK: $(stat -c %%s '%s') bytes"
`, remotePath, tempBucket, transferKey, remotePath, remotePath)

	output, err := m.ssm.SendCommand(ctx, &ssm.SendCommandInput{
		InstanceIds:  []string{instanceData.ID},
		DocumentName: aws.String("AWS-RunShellScript"),
		Parameters:   map[string][]string{"commands": {script}},
		Comment:      aws.String(fmt.Sprintf("Prism file push: %s", filepath.Base(localPath))),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to send SSM command: %w", err)
	}

	cmdOut, err := m.waitForSSMOutput(ctx, instanceData.ID, *output.Command.CommandId)
	if err != nil {
		return &FileOpResult{Path: remotePath, Status: "error", Message: err.Error(), CompletedAt: time.Now()}, err
	}

	return &FileOpResult{
		Path:        remotePath,
		SizeBytes:   fi.Size(),
		Status:      "ok",
		Message:     strings.TrimSpace(cmdOut),
		CompletedAt: time.Now(),
	}, nil
}

// PullFile downloads a file from a running instance to the local machine.
//
// The instance SSM role must allow s3:PutObject on the temp bucket.
func (m *Manager) PullFile(ctx context.Context, instanceName, remotePath, localPath string) (*FileOpResult, error) {
	st, err := m.stateManager.LoadState()
	if err != nil {
		return nil, fmt.Errorf("failed to load state: %w", err)
	}

	instanceData, exists := st.Instances[instanceName]
	if !exists {
		return nil, fmt.Errorf("instance '%s' not found", instanceName)
	}
	if instanceData.State != "running" {
		return nil, fmt.Errorf("instance '%s' must be running (state: %s)", instanceName, instanceData.State)
	}

	tempBucket, err := m.GetTemporaryS3Bucket(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get temp bucket: %w", err)
	}
	transferKey := fmt.Sprintf("prism-transfer/%s/%s", uuid.New().String(), filepath.Base(remotePath))

	s3Client := m.newS3Client()
	defer func() {
		cleanCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_, _ = s3Client.DeleteObject(cleanCtx, &s3.DeleteObjectInput{
			Bucket: aws.String(tempBucket),
			Key:    aws.String(transferKey),
		})
	}()

	// SSM: copy remote file to S3 temp bucket
	script := fmt.Sprintf(`#!/bin/bash
set -e
aws s3 cp "%s" "s3://%s/%s"
echo "OK"
`, remotePath, tempBucket, transferKey)

	output, err := m.ssm.SendCommand(ctx, &ssm.SendCommandInput{
		InstanceIds:  []string{instanceData.ID},
		DocumentName: aws.String("AWS-RunShellScript"),
		Parameters:   map[string][]string{"commands": {script}},
		Comment:      aws.String(fmt.Sprintf("Prism file pull: %s", filepath.Base(remotePath))),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to send SSM command: %w", err)
	}

	if _, err = m.waitForSSMOutput(ctx, instanceData.ID, *output.Command.CommandId); err != nil {
		return &FileOpResult{Path: remotePath, Status: "error", Message: err.Error(), CompletedAt: time.Now()}, err
	}

	// Download from S3 to local
	getResp, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(tempBucket),
		Key:    aws.String(transferKey),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to download from S3 temp bucket: %w", err)
	}
	defer getResp.Body.Close()

	if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create local directory: %w", err)
	}
	dest, err := os.Create(localPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create local file: %w", err)
	}
	defer dest.Close()

	var written int64
	buf := make([]byte, 32*1024)
	for {
		n, readErr := getResp.Body.Read(buf)
		if n > 0 {
			nw, writeErr := dest.Write(buf[:n])
			written += int64(nw)
			if writeErr != nil {
				return nil, fmt.Errorf("failed to write local file: %w", writeErr)
			}
		}
		if readErr != nil {
			break
		}
	}

	return &FileOpResult{
		Path:        localPath,
		SizeBytes:   written,
		Status:      "ok",
		CompletedAt: time.Now(),
	}, nil
}

// ListRemoteFiles returns a directory listing from a running instance via SSM.
func (m *Manager) ListRemoteFiles(ctx context.Context, instanceName, remotePath string) ([]RemoteFileEntry, error) {
	if remotePath == "" {
		remotePath = "/home"
	}

	st, err := m.stateManager.LoadState()
	if err != nil {
		return nil, fmt.Errorf("failed to load state: %w", err)
	}

	instanceData, exists := st.Instances[instanceName]
	if !exists {
		return nil, fmt.Errorf("instance '%s' not found", instanceName)
	}
	if instanceData.State != "running" {
		return nil, fmt.Errorf("instance '%s' must be running (state: %s)", instanceName, instanceData.State)
	}

	// Use ls -la with a parseable format
	script := fmt.Sprintf(`#!/bin/bash
ls -la --time-style=+%%Y-%%m-%%dT%%H:%%M:%%S "%s" 2>/dev/null || echo "ERROR: path not found"
`, remotePath)

	output, err := m.ssm.SendCommand(ctx, &ssm.SendCommandInput{
		InstanceIds:  []string{instanceData.ID},
		DocumentName: aws.String("AWS-RunShellScript"),
		Parameters:   map[string][]string{"commands": {script}},
		Comment:      aws.String("Prism file list"),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to send SSM command: %w", err)
	}

	stdout, err := m.waitForSSMOutput(ctx, instanceData.ID, *output.Command.CommandId)
	if err != nil {
		return nil, err
	}

	return parseLsOutput(stdout, remotePath), nil
}

// parseLsOutput parses the output of `ls -la --time-style=+%Y-%m-%dT%H:%M:%S` into RemoteFileEntry slices.
func parseLsOutput(output, basePath string) []RemoteFileEntry {
	var entries []RemoteFileEntry
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "total") || strings.HasPrefix(line, "ERROR") {
			continue
		}
		// Format: permissions links owner group size date name
		fields := strings.Fields(line)
		if len(fields) < 9 {
			continue
		}
		perms := fields[0]
		isDir := strings.HasPrefix(perms, "d")

		var sizeBytes int64
		if !isDir {
			fmt.Sscanf(fields[4], "%d", &sizeBytes)
		}

		var modTime time.Time
		if t, err := time.Parse("2006-01-02T15:04:05", fields[5]); err == nil {
			modTime = t
		}

		name := strings.Join(fields[8:], " ")
		if name == "." || name == ".." {
			continue
		}

		fullPath := basePath + "/" + name
		entries = append(entries, RemoteFileEntry{
			Path:        fullPath,
			SizeBytes:   sizeBytes,
			IsDir:       isDir,
			ModifiedAt:  modTime,
			Permissions: perms,
		})
	}
	return entries
}

// waitForSSMOutput polls for SSM command completion and returns stdout.
func (m *Manager) waitForSSMOutput(ctx context.Context, instanceID, commandID string) (string, error) {
	maxAttempts := 30
	for attempt := 0; attempt < maxAttempts; attempt++ {
		inv, err := m.ssm.GetCommandInvocation(ctx, &ssm.GetCommandInvocationInput{
			CommandId:  aws.String(commandID),
			InstanceId: aws.String(instanceID),
		})
		if err != nil {
			time.Sleep(10 * time.Second)
			continue
		}

		status := string(inv.Status)
		switch status {
		case "Success":
			stdout := ""
			if inv.StandardOutputContent != nil {
				stdout = *inv.StandardOutputContent
			}
			return stdout, nil
		case "Failed":
			errMsg := "command failed"
			if inv.StandardErrorContent != nil && *inv.StandardErrorContent != "" {
				errMsg = *inv.StandardErrorContent
			} else if inv.StandardOutputContent != nil {
				errMsg = *inv.StandardOutputContent
			}
			return "", fmt.Errorf("SSM command failed: %s", strings.TrimSpace(errMsg))
		case "Cancelled", "TimedOut":
			return "", fmt.Errorf("SSM command %s: %s", status, commandID)
		default:
			time.Sleep(10 * time.Second)
			continue
		}
	}
	return "", fmt.Errorf("SSM command %s timed out", commandID)
}
