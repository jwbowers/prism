package daemon

// S3/SSM file-level backup handlers (Issue #478)
//
// Routes:
//   GET  /api/v1/backups         → handleListS3Backups
//   POST /api/v1/backups         → handleCreateS3Backup
//   GET  /api/v1/backups/{name}  → handleGetS3Backup
//   DELETE /api/v1/backups/{name} → handleDeleteS3Backup
//
// These endpoints serve S3 file-level backups, distinct from the AMI snapshot
// endpoints at /api/v1/snapshots.

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/scttfrdmn/prism/pkg/aws"
	"github.com/scttfrdmn/prism/pkg/types"
)

// handleBackups routes collection-level S3 backup operations.
func (s *Server) handleBackups(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleListS3Backups(w, r)
	case http.MethodPost:
		s.handleCreateS3Backup(w, r)
	default:
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleBackupOperations routes individual S3 backup operations.
func (s *Server) handleBackupOperations(w http.ResponseWriter, r *http.Request) {
	backupName := strings.TrimPrefix(r.URL.Path, "/api/v1/backups/")
	if backupName == "" {
		s.writeError(w, http.StatusBadRequest, "Missing backup name")
		return
	}
	// Trim any trailing slash that http.ServeMux may leave
	backupName = strings.TrimRight(backupName, "/")

	switch r.Method {
	case http.MethodGet:
		s.handleGetS3Backup(w, r, backupName)
	case http.MethodDelete:
		s.handleDeleteS3Backup(w, r, backupName)
	default:
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleCreateS3Backup creates a new S3 file-level backup.
func (s *Server) handleCreateS3Backup(w http.ResponseWriter, r *http.Request) {
	var req types.BackupCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request body: %v", err))
		return
	}
	if req.InstanceName == "" {
		s.writeError(w, http.StatusBadRequest, "instance_name is required")
		return
	}
	if req.BackupName == "" {
		s.writeError(w, http.StatusBadRequest, "backup_name is required")
		return
	}

	// Test mode: return a plausible mock response for any instance
	if s.testMode {
		s.writeJSON(w, http.StatusCreated, &types.BackupCreateResult{
			BackupName:                 req.BackupName,
			BackupID:                   "ssm-cmd-mock-001",
			SourceInstance:             req.InstanceName,
			BackupType:                 "full",
			StorageType:                "s3",
			StorageLocation:            fmt.Sprintf("s3://prism-backups-123456789012-us-west-2/%s/%s/", req.InstanceName, req.BackupName),
			EstimatedCompletionMinutes: 5,
			StorageCostMonthly:         0,
			CreatedAt:                  time.Now(),
			Encrypted:                  false,
			Message:                    "S3 file backup started (test mode)",
		})
		return
	}

	s.withAWSManager(w, r, func(awsManager *aws.Manager) error {
		result, err := awsManager.CreateS3FileBackup(req)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				s.writeError(w, http.StatusNotFound, err.Error())
			} else if strings.Contains(err.Error(), "must be running") {
				s.writeError(w, http.StatusBadRequest, err.Error())
			} else {
				s.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create S3 backup: %v", err))
			}
			return nil
		}
		s.writeJSON(w, http.StatusCreated, result)
		return nil
	})
}

// handleListS3Backups lists all S3 file backups.
func (s *Server) handleListS3Backups(w http.ResponseWriter, r *http.Request) {
	// Test mode: return mock S3 backup list
	if s.testMode {
		now := time.Now()
		s.writeJSON(w, http.StatusOK, types.BackupListResponse{
			Backups: []types.BackupInfo{
				{
					BackupName:         "prism-mock-s3-backup-1",
					BackupID:           "ssm-cmd-mock-001",
					SourceInstance:     "prism-mock-instance",
					SourceInstanceId:   "i-mock000001",
					BackupType:         "full",
					StorageType:        "s3",
					StorageLocation:    "s3://prism-backups-123456789012-us-west-2/prism-mock-instance/prism-mock-s3-backup-1/",
					State:              "available",
					SizeBytes:          1073741824, // 1 GB
					StorageCostMonthly: 0.023,
					CreatedAt:          now.Add(-24 * time.Hour),
				},
			},
			Count:     1,
			TotalSize: 1073741824,
			TotalCost: 0.023,
			StorageTypes: map[string]int{
				"s3": 1,
			},
		})
		return
	}

	s.withAWSManager(w, r, func(awsManager *aws.Manager) error {
		backups, err := awsManager.ListS3FileBackups()
		if err != nil {
			s.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to list S3 backups: %v", err))
			return nil
		}

		var totalSize int64
		var totalCost float64
		for _, b := range backups {
			totalSize += b.SizeBytes
			totalCost += b.StorageCostMonthly
		}

		s.writeJSON(w, http.StatusOK, types.BackupListResponse{
			Backups:   backups,
			Count:     len(backups),
			TotalSize: totalSize,
			TotalCost: totalCost,
			StorageTypes: map[string]int{
				"s3": len(backups),
			},
		})
		return nil
	})
}

// handleGetS3Backup retrieves a specific S3 backup by name.
func (s *Server) handleGetS3Backup(w http.ResponseWriter, r *http.Request, backupName string) {
	// Test mode
	if s.testMode {
		if backupName == "prism-mock-s3-backup-1" {
			now := time.Now()
			completed := now.Add(-23 * time.Hour)
			s.writeJSON(w, http.StatusOK, &types.BackupInfo{
				BackupName:         backupName,
				BackupID:           "ssm-cmd-mock-001",
				SourceInstance:     "prism-mock-instance",
				SourceInstanceId:   "i-mock000001",
				BackupType:         "full",
				StorageType:        "s3",
				StorageLocation:    "s3://prism-backups-123456789012-us-west-2/prism-mock-instance/prism-mock-s3-backup-1/",
				State:              "available",
				SizeBytes:          1073741824,
				StorageCostMonthly: 0.023,
				CreatedAt:          now.Add(-24 * time.Hour),
				CompletedAt:        &completed,
			})
			return
		}
		s.writeError(w, http.StatusNotFound, fmt.Sprintf("S3 backup '%s' not found", backupName))
		return
	}

	s.withAWSManager(w, r, func(awsManager *aws.Manager) error {
		info, err := awsManager.GetS3BackupInfo(backupName)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				s.writeError(w, http.StatusNotFound, err.Error())
			} else {
				s.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get S3 backup: %v", err))
			}
			return nil
		}
		s.writeJSON(w, http.StatusOK, info)
		return nil
	})
}

// handleDeleteS3Backup deletes an S3 backup's metadata entry.
func (s *Server) handleDeleteS3Backup(w http.ResponseWriter, r *http.Request, backupName string) {
	// Test mode
	if s.testMode {
		if strings.HasPrefix(backupName, "prism-mock-") {
			s.writeJSON(w, http.StatusOK, types.BackupDeleteResult{
				BackupName:            backupName,
				BackupID:              "ssm-cmd-mock-001",
				StorageType:           "s3",
				StorageLocation:       fmt.Sprintf("s3://prism-backups-123456789012-us-west-2/prism-mock-instance/%s/", backupName),
				DeletedSizeBytes:      1073741824,
				StorageSavingsMonthly: 0.023,
				DeletedAt:             time.Now(),
			})
			return
		}
		s.writeError(w, http.StatusNotFound, fmt.Sprintf("S3 backup '%s' not found (test mode)", backupName))
		return
	}

	s.withAWSManager(w, r, func(awsManager *aws.Manager) error {
		result, err := awsManager.DeleteS3FileBackup(backupName)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				s.writeError(w, http.StatusNotFound, err.Error())
			} else {
				s.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to delete S3 backup: %v", err))
			}
			return nil
		}
		s.writeJSON(w, http.StatusOK, result)
		return nil
	})
}
