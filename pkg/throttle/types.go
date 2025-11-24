// Package throttle provides launch throttling for cost control and resource management
//
// This package implements configurable launch rate limiting to prevent:
// - Accidental mass launches from script loops
// - Resource exhaustion and cost overruns
// - Unfair resource usage in multi-user environments
//
// Key Features:
// - Global, per-user, and per-project throttling
// - Token bucket algorithm with configurable rates
// - Budget-aware throttling (more aggressive near limits)
// - Integration with policy and budget systems
package throttle

import (
	"time"
)

// Config defines throttling configuration
type Config struct {
	// Global throttling
	Enabled     bool   `json:"enabled"`
	MaxLaunches int    `json:"max_launches"` // Max launches in time window
	TimeWindow  string `json:"time_window"`  // e.g., "1h", "24h"
	BurstSize   int    `json:"burst_size"`   // Burst capacity (0 = same as max)

	// Granular throttling
	PerUser    bool `json:"per_user"`    // Separate buckets per user
	PerProject bool `json:"per_project"` // Separate buckets per project

	// Budget integration
	BudgetAware      bool    `json:"budget_aware"`      // Throttle more near budget limits
	BudgetThreshold  float64 `json:"budget_threshold"`  // Throttle threshold (0.0-1.0, e.g., 0.8 = 80%)
	BudgetMultiplier float64 `json:"budget_multiplier"` // Rate reduction factor (e.g., 0.5 = half rate)

	// Behavior
	QueueMode string `json:"queue_mode"` // "reject", "queue", "warn"
}

// DefaultConfig returns sensible default throttling configuration
func DefaultConfig() Config {
	return Config{
		Enabled:          false, // Opt-in for v0.6.0
		MaxLaunches:      10,
		TimeWindow:       "1h",
		BurstSize:        0, // Same as max by default
		PerUser:          true,
		PerProject:       true,
		BudgetAware:      true,
		BudgetThreshold:  0.80, // Throttle at 80% budget
		BudgetMultiplier: 0.5,  // Half rate near limit
		QueueMode:        "reject",
	}
}

// Status represents current throttling status for a scope
type Status struct {
	Scope            string        `json:"scope"` // "global", "user:username", "project:projectid"
	Enabled          bool          `json:"enabled"`
	MaxLaunches      int           `json:"max_launches"`
	TimeWindow       string        `json:"time_window"`
	CurrentTokens    float64       `json:"current_tokens"`     // Available tokens
	LaunchesInWindow int           `json:"launches_in_window"` // Launches in current window
	NextTokenRefill  time.Time     `json:"next_token_refill"`  // When next token available
	TimeUntilRefill  time.Duration `json:"time_until_refill"`  // Human-readable wait time
	BudgetAdjusted   bool          `json:"budget_adjusted"`    // Whether rate adjusted for budget
	EffectiveRate    float64       `json:"effective_rate"`     // Current rate (may be reduced)
	ConfiguredRate   float64       `json:"configured_rate"`    // Original configured rate
	LastLaunchTime   time.Time     `json:"last_launch_time"`   // Most recent launch
	TotalThrottled   int64         `json:"total_throttled"`    // Total launches rejected
	AllowedLaunches  int64         `json:"allowed_launches"`   // Total launches allowed
	SuccessRate      float64       `json:"success_rate"`       // Percentage allowed (0.0-1.0)
}

// Override represents a project-specific throttling override
type Override struct {
	ProjectID   string    `json:"project_id"`
	MaxLaunches int       `json:"max_launches"`
	TimeWindow  string    `json:"time_window"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	Reason      string    `json:"reason"` // Why override was created
}

// ThrottleError indicates a launch was rejected due to throttling
type ThrottleError struct {
	Scope         string        `json:"scope"`
	MaxLaunches   int           `json:"max_launches"`
	LaunchCount   int           `json:"launch_count"`
	TimeWindow    string        `json:"time_window"`
	RetryAfter    time.Duration `json:"retry_after"`
	CurrentTokens float64       `json:"current_tokens"`
	Message       string        `json:"message"`
}

func (e *ThrottleError) Error() string {
	return e.Message
}

// LaunchRequest contains information about a launch request for throttling check
type LaunchRequest struct {
	UserID       string // Research user or system user
	ProjectID    string // Project ID if applicable
	TemplateName string // Template being launched
	InstanceType string // EC2 instance type
	IsGPU        bool   // Whether GPU instance
}
