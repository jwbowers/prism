package daemon

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/scttfrdmn/prism/pkg/aws"
	"github.com/scttfrdmn/prism/pkg/types"
)

// handleSnapshotOperations routes snapshot operations based on URL path
func (s *Server) handleSnapshotOperations(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/snapshots/")

	// Split path to handle nested operations
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		s.writeError(w, http.StatusBadRequest, "Missing snapshot identifier")
		return
	}

	snapshotName := parts[0]

	// Handle restore operation
	if len(parts) > 1 && parts[1] == "restore" {
		s.handleRestoreInstanceFromSnapshot(w, r, snapshotName)
		return
	}

	// Handle individual snapshot operations
	s.handleSnapshot(w, r, snapshotName)
}

// handleSnapshots handles instance snapshot collection operations
func (s *Server) handleSnapshots(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleListInstanceSnapshots(w, r)
	case http.MethodPost:
		s.handleCreateInstanceSnapshot(w, r)
	default:
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleSnapshot handles individual snapshot operations
func (s *Server) handleSnapshot(w http.ResponseWriter, r *http.Request, snapshotName string) {
	switch r.Method {
	case http.MethodGet:
		s.handleGetInstanceSnapshot(w, r, snapshotName)
	case http.MethodDelete:
		s.handleDeleteInstanceSnapshot(w, r, snapshotName)
	default:
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleCreateInstanceSnapshot creates a snapshot from an instance
func (s *Server) handleCreateInstanceSnapshot(w http.ResponseWriter, r *http.Request) {
	var req types.InstanceSnapshotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request body: %v", err))
		return
	}

	// Validate required fields
	if req.InstanceName == "" {
		s.writeError(w, http.StatusBadRequest, "instance_name is required")
		return
	}
	if req.SnapshotName == "" {
		s.writeError(w, http.StatusBadRequest, "snapshot_name is required")
		return
	}

	// In test mode, accept mock instance names; reject others
	if s.testMode {
		if strings.HasPrefix(req.InstanceName, "prism-mock-") {
			s.writeJSON(w, http.StatusCreated, &types.InstanceSnapshotInfo{
				SnapshotID:         "ami-mock-new-001",
				SnapshotName:       req.SnapshotName,
				SourceInstance:     req.InstanceName,
				SourceInstanceId:   "i-mock000001",
				SourceTemplate:     "Python ML",
				Description:        req.Description,
				State:              "pending",
				Architecture:       "x86_64",
				StorageCostMonthly: 1.25,
				CreatedAt:          time.Now(),
			})
		} else {
			s.writeError(w, http.StatusInternalServerError, "Snapshot creation not available in test mode")
		}
		return
	}

	s.withAWSManager(w, r, func(awsManager *aws.Manager) error {
		// Create the snapshot
		result, err := awsManager.CreateInstanceAMISnapshot(
			req.InstanceName,
			req.SnapshotName,
			req.Description,
			req.NoReboot,
		)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				s.writeError(w, http.StatusNotFound, err.Error())
			} else if strings.Contains(err.Error(), "must be running") {
				s.writeError(w, http.StatusBadRequest, err.Error())
			} else {
				s.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create snapshot: %v", err))
			}
			return nil // Error already written to response
		}

		// If wait is requested, monitor snapshot creation
		if req.Wait {
			s.writeJSON(w, http.StatusAccepted, map[string]interface{}{
				"snapshot": result,
				"message":  "Snapshot creation initiated. Monitoring progress...",
			})
		} else {
			s.writeJSON(w, http.StatusCreated, result)
		}
		return nil
	})
}

// handleListInstanceSnapshots lists all instance snapshots
func (s *Server) handleListInstanceSnapshots(w http.ResponseWriter, r *http.Request) {
	// In test mode, return mock snapshots without making AWS calls
	if s.testMode {
		now := time.Now()
		s.writeJSON(w, http.StatusOK, types.InstanceSnapshotListResponse{
			Snapshots: []types.InstanceSnapshotInfo{
				{
					SnapshotID:         "ami-mock000001",
					SnapshotName:       "prism-mock-backup-1",
					SourceInstance:     "prism-mock-instance",
					SourceInstanceId:   "i-mock000001",
					SourceTemplate:     "Python ML",
					Description:        "Test backup for E2E testing",
					State:              "available",
					Architecture:       "x86_64",
					StorageCostMonthly: 1.25,
					CreatedAt:          now.Add(-24 * time.Hour),
				},
				{
					SnapshotID:         "ami-mock000002",
					SnapshotName:       "prism-mock-backup-2",
					SourceInstance:     "prism-mock-instance",
					SourceInstanceId:   "i-mock000002",
					SourceTemplate:     "R Research",
					Description:        "Second test backup for E2E testing",
					State:              "available",
					Architecture:       "x86_64",
					StorageCostMonthly: 2.50,
					CreatedAt:          now.Add(-48 * time.Hour),
				},
			},
			Count: 2,
		})
		return
	}

	s.withAWSManager(w, r, func(awsManager *aws.Manager) error {
		// List snapshots
		snapshots, err := awsManager.ListInstanceSnapshots()
		if err != nil {
			s.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to list snapshots: %v", err))
			return nil // Error already written to response
		}

		response := types.InstanceSnapshotListResponse{
			Snapshots: snapshots,
			Count:     len(snapshots),
		}

		s.writeJSON(w, http.StatusOK, response)
		return nil
	})
}

// handleGetInstanceSnapshot gets information about a specific snapshot
func (s *Server) handleGetInstanceSnapshot(w http.ResponseWriter, r *http.Request, snapshotName string) {
	// In test mode, return mock data for known backups; 404 for unknown names
	if s.testMode {
		switch snapshotName {
		case "prism-mock-backup-1":
			s.writeJSON(w, http.StatusOK, &types.InstanceSnapshotInfo{
				SnapshotID:         "ami-mock000001",
				SnapshotName:       "prism-mock-backup-1",
				SourceInstance:     "prism-mock-instance",
				SourceInstanceId:   "i-mock000001",
				SourceTemplate:     "Python ML",
				Description:        "Test backup for E2E testing",
				State:              "available",
				Architecture:       "x86_64",
				StorageCostMonthly: 1.25,
				CreatedAt:          time.Now().Add(-24 * time.Hour),
			})
		case "prism-mock-backup-2":
			s.writeJSON(w, http.StatusOK, &types.InstanceSnapshotInfo{
				SnapshotID:         "ami-mock000002",
				SnapshotName:       "prism-mock-backup-2",
				SourceInstance:     "prism-mock-instance",
				SourceInstanceId:   "i-mock000002",
				SourceTemplate:     "R Research",
				Description:        "Second test backup for E2E testing",
				State:              "available",
				Architecture:       "x86_64",
				StorageCostMonthly: 2.50,
				CreatedAt:          time.Now().Add(-48 * time.Hour),
			})
		default:
			s.writeError(w, http.StatusNotFound, fmt.Sprintf("Snapshot '%s' not found", snapshotName))
		}
		return
	}

	s.withAWSManager(w, r, func(awsManager *aws.Manager) error {
		// Get snapshot info
		snapshot, err := awsManager.GetInstanceSnapshotInfo(snapshotName)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				s.writeError(w, http.StatusNotFound, err.Error())
			} else {
				s.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get snapshot info: %v", err))
			}
			return nil // Error already written to response
		}

		s.writeJSON(w, http.StatusOK, snapshot)
		return nil
	})
}

// handleDeleteInstanceSnapshot deletes a snapshot
func (s *Server) handleDeleteInstanceSnapshot(w http.ResponseWriter, r *http.Request, snapshotName string) {
	// In test mode, accept mock snapshot names; reject others
	if s.testMode {
		if strings.HasPrefix(snapshotName, "prism-mock-") {
			savings := 1.25
			if snapshotName == "prism-mock-backup-2" {
				savings = 2.50
			}
			s.writeJSON(w, http.StatusOK, types.InstanceSnapshotDeleteResult{
				SnapshotName:          snapshotName,
				SnapshotID:            fmt.Sprintf("ami-mock-%s", snapshotName),
				StorageSavingsMonthly: savings,
				DeletedAt:             time.Now(),
			})
		} else {
			s.writeError(w, http.StatusNotFound, fmt.Sprintf("Snapshot '%s' not found (test mode)", snapshotName))
		}
		return
	}

	s.withAWSManager(w, r, func(awsManager *aws.Manager) error {
		// Delete the snapshot
		result, err := awsManager.DeleteInstanceSnapshot(snapshotName)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				s.writeError(w, http.StatusNotFound, err.Error())
			} else {
				s.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to delete snapshot: %v", err))
			}
			return nil // Error already written to response
		}

		s.writeJSON(w, http.StatusOK, result)
		return nil
	})
}

// handleRestoreInstanceFromSnapshot restores a new instance from a snapshot
func (s *Server) handleRestoreInstanceFromSnapshot(w http.ResponseWriter, r *http.Request, snapshotName string) {
	var req types.InstanceRestoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request body: %v", err))
		return
	}

	// Validate required fields
	if req.NewInstanceName == "" {
		s.writeError(w, http.StatusBadRequest, "new_instance_name is required")
		return
	}

	// Override snapshot name from URL path
	req.SnapshotName = snapshotName

	// In test mode, accept mock snapshot names; reject others
	if s.testMode {
		if strings.HasPrefix(snapshotName, "prism-mock-") {
			s.writeJSON(w, http.StatusCreated, map[string]interface{}{
				"message":       fmt.Sprintf("Mock restore from '%s' to '%s' initiated", snapshotName, req.NewInstanceName),
				"instance_name": req.NewInstanceName,
				"snapshot_name": snapshotName,
			})
		} else {
			s.writeError(w, http.StatusNotFound, fmt.Sprintf("Snapshot '%s' not found (test mode)", snapshotName))
		}
		return
	}

	s.withAWSManager(w, r, func(awsManager *aws.Manager) error {
		// Restore the instance
		result, err := awsManager.RestoreInstanceFromSnapshot(req.SnapshotName, req.NewInstanceName)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				s.writeError(w, http.StatusNotFound, err.Error())
			} else {
				s.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to restore instance: %v", err))
			}
			return nil // Error already written to response
		}

		if req.Wait {
			s.writeJSON(w, http.StatusAccepted, map[string]interface{}{
				"restore": result,
				"message": "Instance restore initiated. Monitoring progress...",
			})
		} else {
			s.writeJSON(w, http.StatusCreated, result)
		}
		return nil
	})
}
