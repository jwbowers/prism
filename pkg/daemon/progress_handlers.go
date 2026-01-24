package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/scttfrdmn/prism/pkg/types"
)

// ProgressTracker manages progress monitoring for multiple instances
type ProgressTracker struct {
	monitors map[string]*LaunchProgressMonitor
	mu       sync.RWMutex
}

// NewProgressTracker creates a new progress tracker
func NewProgressTracker() *ProgressTracker {
	return &ProgressTracker{
		monitors: make(map[string]*LaunchProgressMonitor),
	}
}

// StartMonitoring starts progress monitoring for an instance
func (pt *ProgressTracker) StartMonitoring(instance *types.Instance, sshKeyPath string) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	// Create monitor
	monitor := NewLaunchProgressMonitor(instance, sshKeyPath, instance.Username)
	pt.monitors[instance.Name] = monitor

	// Start monitoring in background
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()
		defer pt.StopMonitoring(instance.Name)

		if err := monitor.Start(ctx); err != nil {
			// Log error but don't fail - progress monitoring is best-effort
			fmt.Printf("Progress monitoring failed for %s: %v\n", instance.Name, err)
		}
	}()
}

// StopMonitoring stops progress monitoring for an instance
func (pt *ProgressTracker) StopMonitoring(instanceName string) {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	delete(pt.monitors, instanceName)
}

// GetProgress returns the current progress for an instance
func (pt *ProgressTracker) GetProgress(instanceName string) *LaunchProgressMonitor {
	pt.mu.RLock()
	defer pt.mu.RUnlock()
	return pt.monitors[instanceName]
}

// handleGetProgress returns the current progress for an instance
func (s *Server) handleGetProgress(w http.ResponseWriter, r *http.Request, instanceName string) {
	if instanceName == "" {
		s.writeError(w, http.StatusBadRequest, "Instance name required")
		return
	}

	// Get progress monitor
	monitor := s.progressTracker.GetProgress(instanceName)
	if monitor == nil {
		// No active monitoring - instance is either very new or already complete
		response := types.ProgressResponse{
			InstanceName:    instanceName,
			OverallProgress: 0,
			CurrentStage:    "Initializing",
			Stages:          []types.ProgressStage{},
			IsComplete:      false,
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			s.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to encode response: %v", err))
		}
		return
	}

	// Get stages and calculate progress
	stages := monitor.GetStages()
	progress := monitor.GetProgress()

	// Find current stage
	currentStage := "Initializing"
	currentStageIndex := -1
	for i, stage := range stages {
		if stage.Status == "running" {
			currentStage = stage.DisplayName
			currentStageIndex = i
			break
		}
	}

	// Check if complete
	isComplete := progress >= 100

	// Estimate time remaining (rough heuristic based on elapsed time and progress)
	estimatedTimeLeft := ""
	if !isComplete && progress > 0 {
		// Calculate based on average time per percentage point
		elapsed := time.Since(stages[0].StartTime)
		if elapsed > 0 {
			totalEstimated := time.Duration(float64(elapsed) / progress * 100)
			remaining := totalEstimated - elapsed
			if remaining > 0 {
				minutes := int(remaining.Minutes())
				if minutes < 1 {
					estimatedTimeLeft = "< 1 minute"
				} else if minutes == 1 {
					estimatedTimeLeft = "~1 minute"
				} else {
					estimatedTimeLeft = fmt.Sprintf("~%d minutes", minutes)
				}
			}
		}
	}

	response := types.ProgressResponse{
		InstanceName:      instanceName,
		OverallProgress:   progress,
		CurrentStage:      currentStage,
		CurrentStageIndex: currentStageIndex,
		Stages:            stages,
		IsComplete:        isComplete,
		EstimatedTimeLeft: estimatedTimeLeft,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to encode response: %v", err))
		return
	}
}
