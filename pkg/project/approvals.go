// Package project provides approval workflow management for Prism.
//
// This file implements the approval system introduced in v0.12.0, enabling
// governance workflows for expensive resources, budget overages, and
// delegated budget access.
//
// Approval Types:
//   - gpu_instance: Required before launching GPU/p*/g* instance types (#149)
//   - expensive_instance: Required for instances costing >$2/hr (#149)
//   - budget_overage: Required when spending would exceed budget (#157)
//   - emergency: Emergency budget increase request (#157)
//   - sub_budget: Sub-budget delegation request (#148)
//
// Owner/Admin users bypass approval gates automatically.
// Member/Viewer users receive HTTP 202 with a request ID when approval is needed.
package project

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ApprovalType identifies the kind of approval being requested
type ApprovalType string

const (
	// ApprovalTypeGPUInstance requires approval before launching GPU/p*/g* instances (#149)
	ApprovalTypeGPUInstance ApprovalType = "gpu_instance"

	// ApprovalTypeExpensiveInstance requires approval for instances >$2/hr (#149)
	ApprovalTypeExpensiveInstance ApprovalType = "expensive_instance"

	// ApprovalTypeBudgetOverage requires approval when spending would exceed budget
	ApprovalTypeBudgetOverage ApprovalType = "budget_overage"

	// ApprovalTypeEmergency requests an emergency budget increase (#157)
	ApprovalTypeEmergency ApprovalType = "emergency"

	// ApprovalTypeSubBudget requests sub-budget delegation from PI to member (#148)
	ApprovalTypeSubBudget ApprovalType = "sub_budget"
)

// ApprovalStatus represents the current state of an approval request
type ApprovalStatus string

const (
	// ApprovalStatusPending indicates the request is awaiting review
	ApprovalStatusPending ApprovalStatus = "pending"

	// ApprovalStatusApproved indicates the request was approved
	ApprovalStatusApproved ApprovalStatus = "approved"

	// ApprovalStatusDenied indicates the request was denied
	ApprovalStatusDenied ApprovalStatus = "denied"

	// ApprovalStatusExpired indicates the request expired before being reviewed
	ApprovalStatusExpired ApprovalStatus = "expired"
)

// ApprovalRequest is a single governance approval request
type ApprovalRequest struct {
	// ID is the unique request identifier
	ID string `json:"id"`

	// ProjectID is the project this request is associated with
	ProjectID string `json:"project_id"`

	// RequestedBy is the user who submitted the request
	RequestedBy string `json:"requested_by"`

	// Type is the kind of approval needed
	Type ApprovalType `json:"type"`

	// Status is the current state of the request
	Status ApprovalStatus `json:"status"`

	// Details contains type-specific context (instance_type, amount, etc.)
	Details map[string]interface{} `json:"details"`

	// Reason is the requester's justification
	Reason string `json:"reason"`

	// ReviewedBy is the approver's user ID (empty until reviewed)
	ReviewedBy string `json:"reviewed_by,omitempty"`

	// ReviewNote is an optional note from the approver
	ReviewNote string `json:"review_note,omitempty"`

	// CreatedAt is when the request was submitted
	CreatedAt time.Time `json:"created_at"`

	// ExpiresAt is when the pending request auto-expires (default: 7 days)
	ExpiresAt time.Time `json:"expires_at"`

	// ReviewedAt is when the request was approved or denied
	ReviewedAt *time.Time `json:"reviewed_at,omitempty"`
}

// ApprovalManager handles approval request lifecycle with file-backed persistence
type ApprovalManager struct {
	dataPath string
	mu       sync.RWMutex
	requests map[string]*ApprovalRequest // keyed by request ID
}

// NewApprovalManager creates an approval manager persisting to ~/.prism/approvals.json
func NewApprovalManager() (*ApprovalManager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	stateDir := filepath.Join(homeDir, ".prism")
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create state directory: %w", err)
	}

	am := &ApprovalManager{
		dataPath: filepath.Join(stateDir, "approvals.json"),
		requests: make(map[string]*ApprovalRequest),
	}

	if err := am.load(); err != nil {
		return nil, fmt.Errorf("failed to load approvals: %w", err)
	}

	return am, nil
}

// Submit creates a new pending approval request and returns it
func (am *ApprovalManager) Submit(projectID, requestedBy string, typ ApprovalType, details map[string]interface{}, reason string) (*ApprovalRequest, error) {
	am.mu.Lock()
	defer am.mu.Unlock()

	req := &ApprovalRequest{
		ID:          uuid.New().String(),
		ProjectID:   projectID,
		RequestedBy: requestedBy,
		Type:        typ,
		Status:      ApprovalStatusPending,
		Details:     details,
		Reason:      reason,
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(7 * 24 * time.Hour),
	}

	am.requests[req.ID] = req
	return req, am.save()
}

// Get retrieves an approval request by ID
func (am *ApprovalManager) Get(id string) (*ApprovalRequest, error) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	req, ok := am.requests[id]
	if !ok {
		return nil, fmt.Errorf("approval request %q not found", id)
	}

	copy := *req
	return &copy, nil
}

// List returns all approval requests, optionally filtered by project and status
func (am *ApprovalManager) List(projectID string, status ApprovalStatus) ([]*ApprovalRequest, error) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	var results []*ApprovalRequest
	for _, req := range am.requests {
		if projectID != "" && req.ProjectID != projectID {
			continue
		}
		if status != "" && req.Status != status {
			continue
		}
		copy := *req
		results = append(results, &copy)
	}

	return results, nil
}

// Approve marks a request as approved
func (am *ApprovalManager) Approve(id, reviewerID, note string) (*ApprovalRequest, error) {
	am.mu.Lock()
	defer am.mu.Unlock()

	req, ok := am.requests[id]
	if !ok {
		return nil, fmt.Errorf("approval request %q not found", id)
	}

	if req.Status != ApprovalStatusPending {
		return nil, fmt.Errorf("cannot approve request in status %q", req.Status)
	}

	now := time.Now()
	req.Status = ApprovalStatusApproved
	req.ReviewedBy = reviewerID
	req.ReviewNote = note
	req.ReviewedAt = &now

	copy := *req
	return &copy, am.save()
}

// Deny marks a request as denied
func (am *ApprovalManager) Deny(id, reviewerID, note string) (*ApprovalRequest, error) {
	am.mu.Lock()
	defer am.mu.Unlock()

	req, ok := am.requests[id]
	if !ok {
		return nil, fmt.Errorf("approval request %q not found", id)
	}

	if req.Status != ApprovalStatusPending {
		return nil, fmt.Errorf("cannot deny request in status %q", req.Status)
	}

	now := time.Now()
	req.Status = ApprovalStatusDenied
	req.ReviewedBy = reviewerID
	req.ReviewNote = note
	req.ReviewedAt = &now

	copy := *req
	return &copy, am.save()
}

// PruneExpired marks all pending requests that have passed their ExpiresAt as expired
func (am *ApprovalManager) PruneExpired() (int, error) {
	am.mu.Lock()
	defer am.mu.Unlock()

	count := 0
	now := time.Now()
	for _, req := range am.requests {
		if req.Status == ApprovalStatusPending && now.After(req.ExpiresAt) {
			req.Status = ApprovalStatusExpired
			count++
		}
	}

	if count == 0 {
		return 0, nil
	}

	return count, am.save()
}

// load reads approval data from disk (must be called with lock or during init)
func (am *ApprovalManager) load() error {
	if _, err := os.Stat(am.dataPath); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(am.dataPath)
	if err != nil {
		return fmt.Errorf("failed to read approvals file: %w", err)
	}

	var requests map[string]*ApprovalRequest
	if err := json.Unmarshal(data, &requests); err != nil {
		return fmt.Errorf("failed to parse approvals file: %w", err)
	}

	am.requests = requests
	return nil
}

// save writes approval data to disk atomically (must be called with lock held)
func (am *ApprovalManager) save() error {
	data, err := json.MarshalIndent(am.requests, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal approvals: %w", err)
	}

	tmpPath := am.dataPath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write approvals file: %w", err)
	}

	return os.Rename(tmpPath, am.dataPath)
}
