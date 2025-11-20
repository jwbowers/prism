// Package project provides budget and allocation management for v0.5.10+
//
// This file implements the multi-budget system enabling:
//   - 1 budget → N projects (single grant funding multiple projects)
//   - N budgets → 1 project (multi-source funding)
package project

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/scttfrdmn/prism/pkg/types"
)

// BudgetManager handles budget pools and project allocations
type BudgetManager struct {
	budgetsPath       string
	allocationsPath   string
	reallocationsPath string
	mutex             sync.RWMutex
	budgets           map[string]*types.Budget                  // budget_id → Budget
	allocations       map[string]*types.ProjectBudgetAllocation // allocation_id → Allocation
	reallocations     map[string]*ReallocationRecord            // reallocation_id → Record
	// Index for efficient lookups
	projectAllocations      map[string][]*types.ProjectBudgetAllocation // project_id → [Allocations]
	budgetAllocations       map[string][]*types.ProjectBudgetAllocation // budget_id → [Allocations]
	allocationReallocations map[string][]*ReallocationRecord            // allocation_id → [Reallocations]
}

// NewBudgetManager creates a new budget manager
func NewBudgetManager() (*BudgetManager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	stateDir := filepath.Join(homeDir, ".prism")
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create state directory: %w", err)
	}

	budgetsPath := filepath.Join(stateDir, "budgets.json")
	allocationsPath := filepath.Join(stateDir, "budget_allocations.json")
	reallocationsPath := filepath.Join(stateDir, "budget_reallocations.json")

	manager := &BudgetManager{
		budgetsPath:             budgetsPath,
		allocationsPath:         allocationsPath,
		reallocationsPath:       reallocationsPath,
		budgets:                 make(map[string]*types.Budget),
		allocations:             make(map[string]*types.ProjectBudgetAllocation),
		reallocations:           make(map[string]*ReallocationRecord),
		projectAllocations:      make(map[string][]*types.ProjectBudgetAllocation),
		budgetAllocations:       make(map[string][]*types.ProjectBudgetAllocation),
		allocationReallocations: make(map[string][]*ReallocationRecord),
	}

	// Load existing data
	if err := manager.loadBudgets(); err != nil {
		return nil, fmt.Errorf("failed to load budgets: %w", err)
	}

	if err := manager.loadAllocations(); err != nil {
		return nil, fmt.Errorf("failed to load allocations: %w", err)
	}

	if err := manager.loadReallocations(); err != nil {
		return nil, fmt.Errorf("failed to load reallocations: %w", err)
	}

	// Build indexes
	manager.rebuildIndexes()

	return manager, nil
}

// CreateBudget creates a new budget pool
func (bm *BudgetManager) CreateBudget(ctx context.Context, req *CreateBudgetRequest) (*types.Budget, error) {
	bm.mutex.Lock()
	defer bm.mutex.Unlock()

	// Validate request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid budget request: %w", err)
	}

	// Check for duplicate names
	for _, budget := range bm.budgets {
		if budget.Name == req.Name {
			return nil, fmt.Errorf("budget with name %q already exists", req.Name)
		}
	}

	// Create budget
	budget := &types.Budget{
		ID:              uuid.New().String(),
		Name:            req.Name,
		Description:     req.Description,
		TotalAmount:     req.TotalAmount,
		AllocatedAmount: 0.0,
		SpentAmount:     0.0,
		Period:          req.Period,
		StartDate:       req.StartDate,
		EndDate:         req.EndDate,
		AlertThreshold:  req.AlertThreshold,
		CreatedBy:       req.CreatedBy,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		Tags:            req.Tags,
	}

	// Store budget
	bm.budgets[budget.ID] = budget
	if err := bm.saveBudgets(); err != nil {
		delete(bm.budgets, budget.ID)
		return nil, fmt.Errorf("failed to save budget: %w", err)
	}

	return budget, nil
}

// GetBudget retrieves a budget by ID
func (bm *BudgetManager) GetBudget(ctx context.Context, budgetID string) (*types.Budget, error) {
	bm.mutex.RLock()
	defer bm.mutex.RUnlock()

	budget, exists := bm.budgets[budgetID]
	if !exists {
		return nil, fmt.Errorf("budget %q not found", budgetID)
	}

	// Return a copy to prevent external modification
	budgetCopy := *budget
	return &budgetCopy, nil
}

// GetBudgetByName retrieves a budget by name
func (bm *BudgetManager) GetBudgetByName(ctx context.Context, name string) (*types.Budget, error) {
	bm.mutex.RLock()
	defer bm.mutex.RUnlock()

	for _, budget := range bm.budgets {
		if budget.Name == name {
			budgetCopy := *budget
			return &budgetCopy, nil
		}
	}

	return nil, fmt.Errorf("budget with name %q not found", name)
}

// ListBudgets retrieves all budgets
func (bm *BudgetManager) ListBudgets(ctx context.Context) ([]*types.Budget, error) {
	bm.mutex.RLock()
	defer bm.mutex.RUnlock()

	results := make([]*types.Budget, 0, len(bm.budgets))
	for _, budget := range bm.budgets {
		budgetCopy := *budget
		results = append(results, &budgetCopy)
	}

	return results, nil
}

// UpdateBudget updates an existing budget
func (bm *BudgetManager) UpdateBudget(ctx context.Context, budgetID string, req *UpdateBudgetRequest) (*types.Budget, error) {
	bm.mutex.Lock()
	defer bm.mutex.Unlock()

	budget, exists := bm.budgets[budgetID]
	if !exists {
		return nil, fmt.Errorf("budget %q not found", budgetID)
	}

	// Update fields
	if req.Name != nil {
		// Check for duplicate names
		for id, b := range bm.budgets {
			if id != budgetID && b.Name == *req.Name {
				return nil, fmt.Errorf("budget with name %q already exists", *req.Name)
			}
		}
		budget.Name = *req.Name
	}

	if req.Description != nil {
		budget.Description = *req.Description
	}

	if req.TotalAmount != nil {
		budget.TotalAmount = *req.TotalAmount
	}

	if req.AlertThreshold != nil {
		budget.AlertThreshold = *req.AlertThreshold
	}

	if req.EndDate != nil {
		budget.EndDate = req.EndDate
	}

	if req.Tags != nil {
		budget.Tags = req.Tags
	}

	budget.UpdatedAt = time.Now()

	// Save changes
	if err := bm.saveBudgets(); err != nil {
		return nil, fmt.Errorf("failed to save budget updates: %w", err)
	}

	budgetCopy := *budget
	return &budgetCopy, nil
}

// DeleteBudget removes a budget and all associated allocations
func (bm *BudgetManager) DeleteBudget(ctx context.Context, budgetID string) error {
	bm.mutex.Lock()
	defer bm.mutex.Unlock()

	budget, exists := bm.budgets[budgetID]
	if !exists {
		return fmt.Errorf("budget %q not found", budgetID)
	}

	// Check if budget has allocations
	allocations := bm.budgetAllocations[budgetID]
	if len(allocations) > 0 {
		return fmt.Errorf("cannot delete budget with %d project allocations - remove allocations first", len(allocations))
	}

	// Remove budget
	delete(bm.budgets, budgetID)
	if err := bm.saveBudgets(); err != nil {
		// Restore budget on save failure
		bm.budgets[budgetID] = budget
		return fmt.Errorf("failed to save budget deletion: %w", err)
	}

	return nil
}

// CreateAllocation creates a new project budget allocation
func (bm *BudgetManager) CreateAllocation(ctx context.Context, req *CreateAllocationRequest) (*types.ProjectBudgetAllocation, error) {
	bm.mutex.Lock()
	defer bm.mutex.Unlock()

	// Validate request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid allocation request: %w", err)
	}

	// Verify budget exists
	budget, exists := bm.budgets[req.BudgetID]
	if !exists {
		return nil, fmt.Errorf("budget %q not found", req.BudgetID)
	}

	// Check if allocation would exceed budget
	totalAllocated := budget.AllocatedAmount + req.AllocatedAmount
	if totalAllocated > budget.TotalAmount {
		return nil, fmt.Errorf("allocation would exceed budget: $%.2f allocated + $%.2f requested > $%.2f total",
			budget.AllocatedAmount, req.AllocatedAmount, budget.TotalAmount)
	}

	// Check for duplicate budget-project allocation
	for _, allocation := range bm.allocations {
		if allocation.BudgetID == req.BudgetID && allocation.ProjectID == req.ProjectID {
			return nil, fmt.Errorf("budget %q is already allocated to project %q", req.BudgetID, req.ProjectID)
		}
	}

	// Create allocation
	allocation := &types.ProjectBudgetAllocation{
		ID:                 uuid.New().String(),
		BudgetID:           req.BudgetID,
		ProjectID:          req.ProjectID,
		AllocatedAmount:    req.AllocatedAmount,
		SpentAmount:        0.0,
		AlertThreshold:     req.AlertThreshold,
		BackupAllocationID: req.BackupAllocationID,
		Notes:              req.Notes,
		AllocatedAt:        time.Now(),
		AllocatedBy:        req.AllocatedBy,
		UpdatedAt:          time.Now(),
	}

	// Update budget allocated amount
	budget.AllocatedAmount += req.AllocatedAmount
	budget.UpdatedAt = time.Now()

	// Store allocation
	bm.allocations[allocation.ID] = allocation
	bm.rebuildIndexes()

	// Save changes
	if err := bm.saveAllocations(); err != nil {
		delete(bm.allocations, allocation.ID)
		budget.AllocatedAmount -= req.AllocatedAmount
		bm.rebuildIndexes()
		return nil, fmt.Errorf("failed to save allocation: %w", err)
	}

	if err := bm.saveBudgets(); err != nil {
		delete(bm.allocations, allocation.ID)
		budget.AllocatedAmount -= req.AllocatedAmount
		bm.rebuildIndexes()
		return nil, fmt.Errorf("failed to save budget: %w", err)
	}

	return allocation, nil
}

// GetAllocation retrieves an allocation by ID
func (bm *BudgetManager) GetAllocation(ctx context.Context, allocationID string) (*types.ProjectBudgetAllocation, error) {
	bm.mutex.RLock()
	defer bm.mutex.RUnlock()

	allocation, exists := bm.allocations[allocationID]
	if !exists {
		return nil, fmt.Errorf("allocation %q not found", allocationID)
	}

	// Return a copy to prevent external modification
	allocationCopy := *allocation
	return &allocationCopy, nil
}

// GetProjectAllocations retrieves all allocations for a project
func (bm *BudgetManager) GetProjectAllocations(ctx context.Context, projectID string) ([]*types.ProjectBudgetAllocation, error) {
	bm.mutex.RLock()
	defer bm.mutex.RUnlock()

	allocations := bm.projectAllocations[projectID]
	results := make([]*types.ProjectBudgetAllocation, len(allocations))
	for i, alloc := range allocations {
		allocCopy := *alloc
		results[i] = &allocCopy
	}

	return results, nil
}

// GetBudgetAllocations retrieves all allocations for a budget
func (bm *BudgetManager) GetBudgetAllocations(ctx context.Context, budgetID string) ([]*types.ProjectBudgetAllocation, error) {
	bm.mutex.RLock()
	defer bm.mutex.RUnlock()

	allocations := bm.budgetAllocations[budgetID]
	results := make([]*types.ProjectBudgetAllocation, len(allocations))
	for i, alloc := range allocations {
		allocCopy := *alloc
		results[i] = &allocCopy
	}

	return results, nil
}

// UpdateAllocation updates an existing allocation
func (bm *BudgetManager) UpdateAllocation(ctx context.Context, allocationID string, req *UpdateAllocationRequest) (*types.ProjectBudgetAllocation, error) {
	bm.mutex.Lock()
	defer bm.mutex.Unlock()

	allocation, exists := bm.allocations[allocationID]
	if !exists {
		return nil, fmt.Errorf("allocation %q not found", allocationID)
	}

	budget, exists := bm.budgets[allocation.BudgetID]
	if !exists {
		return nil, fmt.Errorf("budget %q not found", allocation.BudgetID)
	}

	// Handle allocated amount change
	if req.AllocatedAmount != nil {
		oldAmount := allocation.AllocatedAmount
		newAmount := *req.AllocatedAmount
		delta := newAmount - oldAmount

		// Check if new allocation would exceed budget
		if budget.AllocatedAmount+delta > budget.TotalAmount {
			return nil, fmt.Errorf("allocation change would exceed budget: $%.2f total - $%.2f allocated + $%.2f delta > $%.2f available",
				budget.TotalAmount, budget.AllocatedAmount, delta, budget.TotalAmount-budget.AllocatedAmount)
		}

		allocation.AllocatedAmount = newAmount
		budget.AllocatedAmount += delta
		budget.UpdatedAt = time.Now()
	}

	// Update other fields
	if req.AlertThreshold != nil {
		allocation.AlertThreshold = req.AlertThreshold
	}

	if req.BackupAllocationID != nil {
		allocation.BackupAllocationID = req.BackupAllocationID
	}

	if req.Notes != nil {
		allocation.Notes = *req.Notes
	}

	allocation.UpdatedAt = time.Now()

	// Save changes
	if err := bm.saveAllocations(); err != nil {
		return nil, fmt.Errorf("failed to save allocation updates: %w", err)
	}

	if err := bm.saveBudgets(); err != nil {
		return nil, fmt.Errorf("failed to save budget updates: %w", err)
	}

	allocationCopy := *allocation
	return &allocationCopy, nil
}

// DeleteAllocation removes an allocation
func (bm *BudgetManager) DeleteAllocation(ctx context.Context, allocationID string) error {
	bm.mutex.Lock()
	defer bm.mutex.Unlock()

	allocation, exists := bm.allocations[allocationID]
	if !exists {
		return fmt.Errorf("allocation %q not found", allocationID)
	}

	budget, exists := bm.budgets[allocation.BudgetID]
	if !exists {
		return fmt.Errorf("budget %q not found", allocation.BudgetID)
	}

	// Update budget allocated amount
	budget.AllocatedAmount -= allocation.AllocatedAmount
	budget.UpdatedAt = time.Now()

	// Remove allocation
	delete(bm.allocations, allocationID)
	bm.rebuildIndexes()

	// Save changes
	if err := bm.saveAllocations(); err != nil {
		bm.allocations[allocationID] = allocation
		budget.AllocatedAmount += allocation.AllocatedAmount
		bm.rebuildIndexes()
		return fmt.Errorf("failed to save allocation deletion: %w", err)
	}

	if err := bm.saveBudgets(); err != nil {
		bm.allocations[allocationID] = allocation
		budget.AllocatedAmount += allocation.AllocatedAmount
		bm.rebuildIndexes()
		return fmt.Errorf("failed to save budget: %w", err)
	}

	return nil
}

// GetBudgetSummary generates a summary view of a budget
func (bm *BudgetManager) GetBudgetSummary(ctx context.Context, budgetID string) (*types.BudgetSummary, error) {
	bm.mutex.RLock()
	defer bm.mutex.RUnlock()

	budget, exists := bm.budgets[budgetID]
	if !exists {
		return nil, fmt.Errorf("budget %q not found", budgetID)
	}

	allocations := bm.budgetAllocations[budgetID]
	allocationCopies := make([]types.ProjectBudgetAllocation, len(allocations))
	for i, alloc := range allocations {
		allocationCopies[i] = *alloc
	}

	remainingAmount := budget.TotalAmount - budget.AllocatedAmount
	var utilizationRate float64
	if budget.AllocatedAmount > 0 {
		utilizationRate = budget.SpentAmount / budget.AllocatedAmount
	}

	return &types.BudgetSummary{
		Budget:          *budget,
		Allocations:     allocationCopies,
		ProjectNames:    make(map[string]string), // TODO: Populate from project manager
		RemainingAmount: remainingAmount,
		UtilizationRate: utilizationRate,
	}, nil
}

// GetProjectFundingSummary generates a summary view of all funding for a project
func (bm *BudgetManager) GetProjectFundingSummary(ctx context.Context, projectID string) (*types.ProjectFundingSummary, error) {
	bm.mutex.RLock()
	defer bm.mutex.RUnlock()

	allocations := bm.projectAllocations[projectID]
	if len(allocations) == 0 {
		return nil, fmt.Errorf("no budget allocations found for project %q", projectID)
	}

	allocationCopies := make([]types.ProjectBudgetAllocation, len(allocations))
	budgetNames := make(map[string]string)
	var totalAllocated, totalSpent float64

	for i, alloc := range allocations {
		allocationCopies[i] = *alloc
		totalAllocated += alloc.AllocatedAmount
		totalSpent += alloc.SpentAmount

		if budget, exists := bm.budgets[alloc.BudgetID]; exists {
			budgetNames[alloc.BudgetID] = budget.Name
		}
	}

	return &types.ProjectFundingSummary{
		ProjectID:           projectID,
		ProjectName:         "", // TODO: Populate from project manager
		Allocations:         allocationCopies,
		BudgetNames:         budgetNames,
		TotalAllocated:      totalAllocated,
		TotalSpent:          totalSpent,
		DefaultAllocationID: nil, // TODO: Populate from project manager
	}, nil
}

// SpendingResult contains information about spending and potential backup activation
type SpendingResult struct {
	AllocationID        string
	AllocationExhausted bool
	BackupActivated     bool
	BackupAllocationID  string
	WarningMessage      string
}

// RecordSpending records spending against an allocation with backup funding support (v0.5.10+)
// Returns SpendingResult indicating if backup was activated
func (bm *BudgetManager) RecordSpending(ctx context.Context, allocationID string, amount float64) (*SpendingResult, error) {
	bm.mutex.Lock()
	defer bm.mutex.Unlock()

	result := &SpendingResult{
		AllocationID: allocationID,
	}

	allocation, exists := bm.allocations[allocationID]
	if !exists {
		return nil, fmt.Errorf("allocation %q not found", allocationID)
	}

	budget, exists := bm.budgets[allocation.BudgetID]
	if !exists {
		return nil, fmt.Errorf("budget %q not found", allocation.BudgetID)
	}

	// Update spending
	allocation.SpentAmount += amount
	allocation.UpdatedAt = time.Now()

	budget.SpentAmount += amount
	budget.UpdatedAt = time.Now()

	// Check if allocation is exhausted (v0.5.10+ backup funding)
	if allocation.SpentAmount >= allocation.AllocatedAmount {
		result.AllocationExhausted = true

		// Check for backup allocation (Issue #234)
		if allocation.BackupAllocationID != nil && *allocation.BackupAllocationID != "" {
			backupID := *allocation.BackupAllocationID

			// Verify backup allocation exists and is not exhausted
			backupAlloc, exists := bm.allocations[backupID]
			if exists && backupAlloc.SpentAmount < backupAlloc.AllocatedAmount {
				result.BackupActivated = true
				result.BackupAllocationID = backupID

				// Get budget names for messaging
				budgetName := budget.Name
				backupBudget := bm.budgets[backupAlloc.BudgetID]
				backupBudgetName := backupBudget.Name

				result.WarningMessage = fmt.Sprintf(
					"⚠️  Primary funding exhausted: %s ($%.2f spent / $%.2f allocated)\n"+
						"✅ Automatically switched to backup funding: %s ($%.2f available)\n"+
						"   Project will continue using backup allocation.",
					budgetName, allocation.SpentAmount, allocation.AllocatedAmount,
					backupBudgetName, backupAlloc.AllocatedAmount-backupAlloc.SpentAmount)
			} else {
				// Backup exists but is also exhausted or invalid
				result.WarningMessage = fmt.Sprintf(
					"❌ Primary funding exhausted: %s ($%.2f spent / $%.2f allocated)\n"+
						"❌ Backup funding unavailable or also exhausted\n"+
						"   No additional funds available for this project.",
					budget.Name, allocation.SpentAmount, allocation.AllocatedAmount)
			}
		} else {
			// No backup configured
			result.WarningMessage = fmt.Sprintf(
				"❌ Allocation exhausted: %s ($%.2f spent / $%.2f allocated)\n"+
					"   No backup funding configured for this allocation.",
				budget.Name, allocation.SpentAmount, allocation.AllocatedAmount)
		}
	}

	// Save changes
	if err := bm.saveAllocations(); err != nil {
		return nil, fmt.Errorf("failed to save allocation: %w", err)
	}

	if err := bm.saveBudgets(); err != nil {
		return nil, fmt.Errorf("failed to save budget: %w", err)
	}

	return result, nil
}

// CheckAllocationStatus checks if an allocation is exhausted or nearing exhaustion
func (bm *BudgetManager) CheckAllocationStatus(ctx context.Context, allocationID string) (exhausted bool, remaining float64, err error) {
	bm.mutex.RLock()
	defer bm.mutex.RUnlock()

	allocation, exists := bm.allocations[allocationID]
	if !exists {
		return false, 0, fmt.Errorf("allocation %q not found", allocationID)
	}

	remaining = allocation.AllocatedAmount - allocation.SpentAmount
	exhausted = remaining <= 0

	return exhausted, remaining, nil
}

// ActivateBackupFunding switches a project to its backup funding allocation
// This is called when the primary allocation is exhausted and backup exists
func (bm *BudgetManager) ActivateBackupFunding(ctx context.Context, projectID string, primaryAllocationID string, backupAllocationID string) error {
	bm.mutex.Lock()
	defer bm.mutex.Unlock()

	// Verify allocations exist
	primaryAlloc, exists := bm.allocations[primaryAllocationID]
	if !exists {
		return fmt.Errorf("primary allocation %q not found", primaryAllocationID)
	}

	backupAlloc, exists := bm.allocations[backupAllocationID]
	if !exists {
		return fmt.Errorf("backup allocation %q not found", backupAllocationID)
	}

	// Verify both belong to the same project
	if primaryAlloc.ProjectID != projectID || backupAlloc.ProjectID != projectID {
		return fmt.Errorf("allocations do not belong to project %q", projectID)
	}

	// Verify backup has available funds
	if backupAlloc.SpentAmount >= backupAlloc.AllocatedAmount {
		return fmt.Errorf("backup allocation %q is exhausted", backupAllocationID)
	}

	// Note: Updating project's default allocation is handled by the project manager
	// This method just validates the backup activation is possible

	return nil
}

// loadBudgets loads budgets from disk
func (bm *BudgetManager) loadBudgets() error {
	// Check if budgets file exists
	if _, err := os.Stat(bm.budgetsPath); os.IsNotExist(err) {
		// No budgets file exists yet, start with empty map
		return nil
	}

	data, err := os.ReadFile(bm.budgetsPath)
	if err != nil {
		return fmt.Errorf("failed to read budgets file: %w", err)
	}

	var budgets map[string]*types.Budget
	if err := json.Unmarshal(data, &budgets); err != nil {
		return fmt.Errorf("failed to parse budgets file: %w", err)
	}

	bm.budgets = budgets
	return nil
}

// saveBudgets saves budgets to disk
func (bm *BudgetManager) saveBudgets() error {
	data, err := json.MarshalIndent(bm.budgets, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal budgets: %w", err)
	}

	// Write to temporary file first, then rename for atomicity
	tempPath := bm.budgetsPath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temporary budgets file: %w", err)
	}

	if err := os.Rename(tempPath, bm.budgetsPath); err != nil {
		return fmt.Errorf("failed to rename budgets file: %w", err)
	}

	return nil
}

// loadAllocations loads allocations from disk
func (bm *BudgetManager) loadAllocations() error {
	// Check if allocations file exists
	if _, err := os.Stat(bm.allocationsPath); os.IsNotExist(err) {
		// No allocations file exists yet, start with empty map
		return nil
	}

	data, err := os.ReadFile(bm.allocationsPath)
	if err != nil {
		return fmt.Errorf("failed to read allocations file: %w", err)
	}

	var allocations map[string]*types.ProjectBudgetAllocation
	if err := json.Unmarshal(data, &allocations); err != nil {
		return fmt.Errorf("failed to parse allocations file: %w", err)
	}

	bm.allocations = allocations
	return nil
}

// saveAllocations saves allocations to disk
func (bm *BudgetManager) saveAllocations() error {
	data, err := json.MarshalIndent(bm.allocations, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal allocations: %w", err)
	}

	// Write to temporary file first, then rename for atomicity
	tempPath := bm.allocationsPath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temporary allocations file: %w", err)
	}

	if err := os.Rename(tempPath, bm.allocationsPath); err != nil {
		return fmt.Errorf("failed to rename allocations file: %w", err)
	}

	return nil
}

// rebuildIndexes rebuilds the lookup indexes for efficient queries
func (bm *BudgetManager) rebuildIndexes() {
	bm.projectAllocations = make(map[string][]*types.ProjectBudgetAllocation)
	bm.budgetAllocations = make(map[string][]*types.ProjectBudgetAllocation)
	bm.allocationReallocations = make(map[string][]*ReallocationRecord)

	for _, allocation := range bm.allocations {
		bm.projectAllocations[allocation.ProjectID] = append(bm.projectAllocations[allocation.ProjectID], allocation)
		bm.budgetAllocations[allocation.BudgetID] = append(bm.budgetAllocations[allocation.BudgetID], allocation)
	}

	for _, reallocation := range bm.reallocations {
		bm.allocationReallocations[reallocation.SourceAllocationID] = append(bm.allocationReallocations[reallocation.SourceAllocationID], reallocation)
		bm.allocationReallocations[reallocation.DestinationAllocationID] = append(bm.allocationReallocations[reallocation.DestinationAllocationID], reallocation)
	}
}

// loadReallocations loads reallocation history from disk
func (bm *BudgetManager) loadReallocations() error {
	// Check if reallocations file exists
	if _, err := os.Stat(bm.reallocationsPath); os.IsNotExist(err) {
		// No reallocations file exists yet, start with empty map
		return nil
	}

	data, err := os.ReadFile(bm.reallocationsPath)
	if err != nil {
		return fmt.Errorf("failed to read reallocations file: %w", err)
	}

	var reallocations map[string]*ReallocationRecord
	if err := json.Unmarshal(data, &reallocations); err != nil {
		return fmt.Errorf("failed to parse reallocations file: %w", err)
	}

	bm.reallocations = reallocations
	return nil
}

// saveReallocations saves reallocation history to disk
func (bm *BudgetManager) saveReallocations() error {
	data, err := json.MarshalIndent(bm.reallocations, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal reallocations: %w", err)
	}

	// Write to temporary file first, then rename for atomicity
	tempPath := bm.reallocationsPath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temporary reallocations file: %w", err)
	}

	if err := os.Rename(tempPath, bm.reallocationsPath); err != nil {
		return fmt.Errorf("failed to rename reallocations file: %w", err)
	}

	return nil
}

// ReallocateFunds moves funds between allocations atomically (v0.5.10+ Issue #99)
func (bm *BudgetManager) ReallocateFunds(ctx context.Context, req *ReallocateFundsRequest) (*ReallocationRecord, error) {
	bm.mutex.Lock()
	defer bm.mutex.Unlock()

	// Validate request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid reallocation request: %w", err)
	}

	// Get source and destination allocations
	sourceAlloc, exists := bm.allocations[req.SourceAllocationID]
	if !exists {
		return nil, fmt.Errorf("source allocation %q not found", req.SourceAllocationID)
	}

	destAlloc, exists := bm.allocations[req.DestinationAllocationID]
	if !exists {
		return nil, fmt.Errorf("destination allocation %q not found", req.DestinationAllocationID)
	}

	// Get source and destination budgets
	sourceBudget, exists := bm.budgets[sourceAlloc.BudgetID]
	if !exists {
		return nil, fmt.Errorf("source budget %q not found", sourceAlloc.BudgetID)
	}

	destBudget, exists := bm.budgets[destAlloc.BudgetID]
	if !exists {
		return nil, fmt.Errorf("destination budget %q not found", destAlloc.BudgetID)
	}

	// Validate reallocation amount
	availableInSource := sourceAlloc.AllocatedAmount - sourceAlloc.SpentAmount
	if req.Amount > availableInSource {
		return nil, fmt.Errorf("insufficient unspent funds in source allocation: $%.2f available, $%.2f requested",
			availableInSource, req.Amount)
	}

	// For cross-budget reallocations, check destination budget has capacity
	if sourceAlloc.BudgetID != destAlloc.BudgetID {
		availableInDestBudget := destBudget.TotalAmount - destBudget.AllocatedAmount
		if req.Amount > availableInDestBudget {
			return nil, fmt.Errorf("insufficient capacity in destination budget: $%.2f available, $%.2f requested",
				availableInDestBudget, req.Amount)
		}
	}

	// Create reallocation record
	record := &ReallocationRecord{
		ID:                      uuid.New().String(),
		SourceAllocationID:      req.SourceAllocationID,
		DestinationAllocationID: req.DestinationAllocationID,
		SourceBudgetID:          sourceAlloc.BudgetID,
		DestinationBudgetID:     destAlloc.BudgetID,
		Amount:                  req.Amount,
		Reason:                  req.Reason,
		PerformedBy:             req.PerformedBy,
		Timestamp:               time.Now(),
	}

	// Perform the reallocation atomically
	sourceAlloc.AllocatedAmount -= req.Amount
	destAlloc.AllocatedAmount += req.Amount

	// Update budget allocated amounts if cross-budget
	if sourceAlloc.BudgetID != destAlloc.BudgetID {
		sourceBudget.AllocatedAmount -= req.Amount
		destBudget.AllocatedAmount += req.Amount
		sourceBudget.UpdatedAt = time.Now()
		destBudget.UpdatedAt = time.Now()
	}

	sourceAlloc.UpdatedAt = time.Now()
	destAlloc.UpdatedAt = time.Now()

	// Store reallocation record
	bm.reallocations[record.ID] = record
	bm.rebuildIndexes()

	// Save all changes
	if err := bm.saveAllocations(); err != nil {
		// Rollback changes
		sourceAlloc.AllocatedAmount += req.Amount
		destAlloc.AllocatedAmount -= req.Amount
		if sourceAlloc.BudgetID != destAlloc.BudgetID {
			sourceBudget.AllocatedAmount += req.Amount
			destBudget.AllocatedAmount -= req.Amount
		}
		delete(bm.reallocations, record.ID)
		bm.rebuildIndexes()
		return nil, fmt.Errorf("failed to save allocations: %w", err)
	}

	if err := bm.saveBudgets(); err != nil {
		// Rollback changes
		sourceAlloc.AllocatedAmount += req.Amount
		destAlloc.AllocatedAmount -= req.Amount
		if sourceAlloc.BudgetID != destAlloc.BudgetID {
			sourceBudget.AllocatedAmount += req.Amount
			destBudget.AllocatedAmount -= req.Amount
		}
		delete(bm.reallocations, record.ID)
		bm.rebuildIndexes()
		return nil, fmt.Errorf("failed to save budgets: %w", err)
	}

	if err := bm.saveReallocations(); err != nil {
		// Rollback changes
		sourceAlloc.AllocatedAmount += req.Amount
		destAlloc.AllocatedAmount -= req.Amount
		if sourceAlloc.BudgetID != destAlloc.BudgetID {
			sourceBudget.AllocatedAmount += req.Amount
			destBudget.AllocatedAmount -= req.Amount
		}
		delete(bm.reallocations, record.ID)
		bm.rebuildIndexes()
		return nil, fmt.Errorf("failed to save reallocation record: %w", err)
	}

	recordCopy := *record
	return &recordCopy, nil
}

// GetReallocationHistory retrieves reallocation history for a specific allocation (v0.5.10+ Issue #99)
func (bm *BudgetManager) GetReallocationHistory(ctx context.Context, allocationID string) ([]*ReallocationRecord, error) {
	bm.mutex.RLock()
	defer bm.mutex.RUnlock()

	// Verify allocation exists
	if _, exists := bm.allocations[allocationID]; !exists {
		return nil, fmt.Errorf("allocation %q not found", allocationID)
	}

	records := bm.allocationReallocations[allocationID]
	results := make([]*ReallocationRecord, len(records))
	for i, record := range records {
		recordCopy := *record
		results[i] = &recordCopy
	}

	return results, nil
}

// GetBudgetReallocationHistory retrieves all reallocations for a budget (v0.5.10+ Issue #99)
func (bm *BudgetManager) GetBudgetReallocationHistory(ctx context.Context, budgetID string) ([]*ReallocationRecord, error) {
	bm.mutex.RLock()
	defer bm.mutex.RUnlock()

	// Verify budget exists
	if _, exists := bm.budgets[budgetID]; !exists {
		return nil, fmt.Errorf("budget %q not found", budgetID)
	}

	// Collect all reallocations involving this budget
	var results []*ReallocationRecord
	for _, record := range bm.reallocations {
		if record.SourceBudgetID == budgetID || record.DestinationBudgetID == budgetID {
			recordCopy := *record
			results = append(results, &recordCopy)
		}
	}

	return results, nil
}

// ============================================================================
// Multi-Project Cost Rollup and Reporting (v0.5.10+ Issue #100)
// ============================================================================

// GenerateBudgetRollupReport generates a comprehensive rollup report across all budgets (v0.5.10+ Issue #100)
func (bm *BudgetManager) GenerateBudgetRollupReport(ctx context.Context) (*BudgetRollupReport, error) {
	bm.mutex.RLock()
	defer bm.mutex.RUnlock()

	report := &BudgetRollupReport{
		ReportID:        uuid.New().String(),
		GeneratedAt:     time.Now(),
		BudgetSummaries: make([]BudgetSummaryReport, 0),
	}

	projectsMap := make(map[string]bool)

	// Generate summary for each budget
	for _, budget := range bm.budgets {
		summary := bm.generateBudgetSummary(budget)
		report.BudgetSummaries = append(report.BudgetSummaries, summary)

		// Track unique projects
		for _, proj := range summary.Projects {
			projectsMap[proj.ProjectID] = true
		}

		// Aggregate totals
		report.TotalAllocated += budget.AllocatedAmount
		report.TotalSpent += budget.SpentAmount
	}

	report.TotalBudgets = len(bm.budgets)
	report.TotalRemaining = report.TotalAllocated - report.TotalSpent
	if report.TotalAllocated > 0 {
		report.OverallUtilization = report.TotalSpent / report.TotalAllocated
	}
	report.ProjectCount = len(projectsMap)

	return report, nil
}

// generateBudgetSummary creates a detailed summary for a single budget
func (bm *BudgetManager) generateBudgetSummary(budget *types.Budget) BudgetSummaryReport {
	summary := BudgetSummaryReport{
		BudgetID:        budget.ID,
		BudgetName:      budget.Name,
		Period:          string(budget.Period),
		TotalAmount:     budget.TotalAmount,
		AllocatedAmount: budget.AllocatedAmount,
		SpentAmount:     budget.SpentAmount,
		RemainingAmount: budget.TotalAmount - budget.AllocatedAmount,
		Projects:        make([]ProjectCostSummary, 0),
	}

	if budget.AllocatedAmount > 0 {
		summary.Utilization = budget.SpentAmount / budget.AllocatedAmount
	}

	// Get allocations for this budget
	allocations := bm.budgetAllocations[budget.ID]
	summary.AllocationCount = len(allocations)

	// Group allocations by project
	projectAllocations := make(map[string][]*types.ProjectBudgetAllocation)
	for _, alloc := range allocations {
		projectAllocations[alloc.ProjectID] = append(projectAllocations[alloc.ProjectID], alloc)
	}

	summary.ProjectCount = len(projectAllocations)

	// Generate project summaries
	for projectID, projAllocs := range projectAllocations {
		projectSummary := bm.generateProjectCostSummary(projectID, projAllocs)
		summary.Projects = append(summary.Projects, projectSummary)
	}

	return summary
}

// generateProjectCostSummary creates a cost summary for a project
func (bm *BudgetManager) generateProjectCostSummary(projectID string, allocations []*types.ProjectBudgetAllocation) ProjectCostSummary {
	summary := ProjectCostSummary{
		ProjectID:      projectID,
		ProjectName:    projectID, // Will be populated by caller if project manager available
		FundingSources: make([]AllocationSummary, 0),
	}

	for _, alloc := range allocations {
		allocSummary := AllocationSummary{
			AllocationID:    alloc.ID,
			BudgetID:        alloc.BudgetID,
			BudgetName:      bm.budgets[alloc.BudgetID].Name,
			AllocatedAmount: alloc.AllocatedAmount,
			SpentAmount:     alloc.SpentAmount,
			RemainingAmount: alloc.AllocatedAmount - alloc.SpentAmount,
			HasBackup:       alloc.BackupAllocationID != nil && *alloc.BackupAllocationID != "",
		}

		if alloc.AllocatedAmount > 0 {
			allocSummary.Utilization = alloc.SpentAmount / alloc.AllocatedAmount
		}

		summary.FundingSources = append(summary.FundingSources, allocSummary)
		summary.TotalAllocated += alloc.AllocatedAmount
		summary.TotalSpent += alloc.SpentAmount
	}

	summary.TotalRemaining = summary.TotalAllocated - summary.TotalSpent
	if summary.TotalAllocated > 0 {
		summary.Utilization = summary.TotalSpent / summary.TotalAllocated
	}

	return summary
}

// GetBudgetSummaryReport generates a detailed summary for a specific budget (v0.5.10+ Issue #100)
func (bm *BudgetManager) GetBudgetSummaryReport(ctx context.Context, budgetID string) (*BudgetSummaryReport, error) {
	bm.mutex.RLock()
	defer bm.mutex.RUnlock()

	budget, exists := bm.budgets[budgetID]
	if !exists {
		return nil, fmt.Errorf("budget %q not found", budgetID)
	}

	summary := bm.generateBudgetSummary(budget)
	return &summary, nil
}

// GetProjectCostRollup generates cost rollup for specific projects (v0.5.10+ Issue #100)
func (bm *BudgetManager) GetProjectCostRollup(ctx context.Context, projectIDs []string) ([]ProjectCostSummary, error) {
	bm.mutex.RLock()
	defer bm.mutex.RUnlock()

	summaries := make([]ProjectCostSummary, 0, len(projectIDs))

	for _, projectID := range projectIDs {
		allocations := bm.projectAllocations[projectID]
		if len(allocations) == 0 {
			continue // Skip projects with no allocations
		}

		summary := bm.generateProjectCostSummary(projectID, allocations)
		summaries = append(summaries, summary)
	}

	return summaries, nil
}

// Close cleanly shuts down the budget manager
func (bm *BudgetManager) Close() error {
	bm.mutex.Lock()
	defer bm.mutex.Unlock()

	// Nothing to clean up currently, but method exists for future needs
	return nil
}
