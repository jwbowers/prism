package types

import "time"

// ProgressStage represents a stage in the instance setup process
type ProgressStage struct {
	Name        string    `json:"name"`         // Internal name: "system-packages"
	DisplayName string    `json:"display_name"` // User-facing: "Installing system packages"
	Status      string    `json:"status"`       // "pending", "running", "complete", "error"
	StartTime   time.Time `json:"start_time"`   // When stage started
	EndTime     time.Time `json:"end_time"`     // When stage ended
	Output      string    `json:"output"`       // Last output line from this stage
}

// ProgressResponse represents progress information for an instance launch
type ProgressResponse struct {
	InstanceName      string          `json:"instance_name"`
	OverallProgress   float64         `json:"overall_progress"`    // 0-100
	CurrentStage      string          `json:"current_stage"`       // Human-readable current stage
	CurrentStageIndex int             `json:"current_stage_index"` // Index of current stage
	Stages            []ProgressStage `json:"stages"`
	IsComplete        bool            `json:"is_complete"`
	EstimatedTimeLeft string          `json:"estimated_time_left"` // e.g., "~5 minutes"
}
