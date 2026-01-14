package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/scttfrdmn/prism/pkg/project"
)

// handleBudgetOperations routes budget-related requests (v0.5.10+)
func (s *Server) handleBudgetOperations(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleListBudgets(w, r)
	case http.MethodPost:
		s.handleCreateBudget(w, r)
	default:
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleBudgetByID routes budget-specific requests (v0.5.10+)
func (s *Server) handleBudgetByID(w http.ResponseWriter, r *http.Request) {
	// Parse budget ID from path
	path := r.URL.Path[len("/api/v1/budgets/"):]
	parts := splitPath(path)
	if len(parts) == 0 {
		s.writeError(w, http.StatusBadRequest, "Missing budget ID")
		return
	}

	budgetID := parts[0]

	if len(parts) == 1 {
		// Direct budget operations
		switch r.Method {
		case http.MethodGet:
			s.handleGetBudget(w, r, budgetID)
		case http.MethodPut:
			s.handleUpdateBudget(w, r, budgetID)
		case http.MethodDelete:
			s.handleDeleteBudget(w, r, budgetID)
		default:
			s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
		return
	}

	// Sub-operations
	operation := parts[1]
	switch operation {
	case "summary":
		s.handleBudgetSummary(w, r, budgetID)
	case "allocations":
		s.handleBudgetAllocations(w, r, budgetID)
	case "reallocations":
		s.handleBudgetReallocationHistory(w, r, budgetID)
	case "report":
		s.handleBudgetSummaryReport(w, r, budgetID)
	default:
		s.writeError(w, http.StatusNotFound, "Unknown budget operation")
	}
}

// handleListBudgets lists all budget pools
func (s *Server) handleListBudgets(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	budgets, err := s.budgetManager.ListBudgets(context.Background())
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to list budgets: %v", err))
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"budgets": budgets,
		"count":   len(budgets),
	})
}

// handleCreateBudget creates a new budget pool
func (s *Server) handleCreateBudget(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req project.CreateBudgetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	budget, err := s.budgetManager.CreateBudget(context.Background(), &req)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create budget: %v", err))
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(budget)
}

// handleGetBudget retrieves a specific budget
func (s *Server) handleGetBudget(w http.ResponseWriter, r *http.Request, budgetID string) {
	budget, err := s.budgetManager.GetBudget(context.Background(), budgetID)
	if err != nil {
		s.writeError(w, http.StatusNotFound, fmt.Sprintf("Budget not found: %v", err))
		return
	}

	_ = json.NewEncoder(w).Encode(budget)
}

// handleUpdateBudget updates an existing budget
func (s *Server) handleUpdateBudget(w http.ResponseWriter, r *http.Request, budgetID string) {
	var req project.UpdateBudgetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	budget, err := s.budgetManager.UpdateBudget(context.Background(), budgetID, &req)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to update budget: %v", err))
		return
	}

	_ = json.NewEncoder(w).Encode(budget)
}

// handleDeleteBudget removes a budget
func (s *Server) handleDeleteBudget(w http.ResponseWriter, r *http.Request, budgetID string) {
	if err := s.budgetManager.DeleteBudget(context.Background(), budgetID); err != nil {
		s.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to delete budget: %v", err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleBudgetSummary generates a summary view of a budget
func (s *Server) handleBudgetSummary(w http.ResponseWriter, r *http.Request, budgetID string) {
	if r.Method != http.MethodGet {
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	summary, err := s.budgetManager.GetBudgetSummary(context.Background(), budgetID)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get budget summary: %v", err))
		return
	}

	// Populate project names from project manager
	for _, allocation := range summary.Allocations {
		if proj, err := s.projectManager.GetProject(context.Background(), allocation.ProjectID); err == nil {
			summary.ProjectNames[allocation.ProjectID] = proj.Name
		}
	}

	_ = json.NewEncoder(w).Encode(summary)
}

// handleBudgetAllocations lists allocations for a budget
func (s *Server) handleBudgetAllocations(w http.ResponseWriter, r *http.Request, budgetID string) {
	if r.Method != http.MethodGet {
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	allocations, err := s.budgetManager.GetBudgetAllocations(context.Background(), budgetID)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get budget allocations: %v", err))
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"allocations": allocations,
		"count":       len(allocations),
	})
}

// ============================================================================
// Allocation Endpoints
// ============================================================================

// handleAllocationOperations routes allocation-related requests (v0.5.10+)
func (s *Server) handleAllocationOperations(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		s.handleCreateAllocation(w, r)
	default:
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleAllocationByID routes allocation-specific requests (v0.5.10+)
func (s *Server) handleAllocationByID(w http.ResponseWriter, r *http.Request) {
	// Parse allocation ID from path
	path := r.URL.Path[len("/api/v1/allocations/"):]
	parts := splitPath(path)
	if len(parts) == 0 {
		s.writeError(w, http.StatusBadRequest, "Missing allocation ID")
		return
	}

	allocationID := parts[0]

	if len(parts) == 1 {
		// Direct allocation operations
		switch r.Method {
		case http.MethodGet:
			s.handleGetAllocation(w, r, allocationID)
		case http.MethodPut:
			s.handleUpdateAllocation(w, r, allocationID)
		case http.MethodDelete:
			s.handleDeleteAllocation(w, r, allocationID)
		default:
			s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
		return
	}

	// Sub-operations
	operation := parts[1]
	switch operation {
	case "status":
		s.handleAllocationStatus(w, r, allocationID)
	case "reallocations":
		s.handleAllocationReallocationHistory(w, r, allocationID)
	case "spending":
		s.handleRecordSpending(w, r, allocationID)
	default:
		s.writeError(w, http.StatusNotFound, "Unknown allocation operation")
	}
}

// handleCreateAllocation creates a new project budget allocation
func (s *Server) handleCreateAllocation(w http.ResponseWriter, r *http.Request) {
	var req project.CreateAllocationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	allocation, err := s.budgetManager.CreateAllocation(context.Background(), &req)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create allocation: %v", err))
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(allocation)
}

// handleGetAllocation retrieves a specific allocation
func (s *Server) handleGetAllocation(w http.ResponseWriter, r *http.Request, allocationID string) {
	allocation, err := s.budgetManager.GetAllocation(context.Background(), allocationID)
	if err != nil {
		s.writeError(w, http.StatusNotFound, fmt.Sprintf("Allocation not found: %v", err))
		return
	}

	_ = json.NewEncoder(w).Encode(allocation)
}

// handleUpdateAllocation updates an existing allocation
func (s *Server) handleUpdateAllocation(w http.ResponseWriter, r *http.Request, allocationID string) {
	var req project.UpdateAllocationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	allocation, err := s.budgetManager.UpdateAllocation(context.Background(), allocationID, &req)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to update allocation: %v", err))
		return
	}

	_ = json.NewEncoder(w).Encode(allocation)
}

// handleDeleteAllocation removes an allocation
func (s *Server) handleDeleteAllocation(w http.ResponseWriter, r *http.Request, allocationID string) {
	if err := s.budgetManager.DeleteAllocation(context.Background(), allocationID); err != nil {
		s.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to delete allocation: %v", err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleRecordSpending records spending against an allocation (v0.5.10+)
func (s *Server) handleRecordSpending(w http.ResponseWriter, r *http.Request, allocationID string) {
	if r.Method != http.MethodPost {
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		Amount float64 `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	result, err := s.budgetManager.RecordSpending(context.Background(), allocationID, req.Amount)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to record spending: %v", err))
		return
	}

	_ = json.NewEncoder(w).Encode(result)
}

// ============================================================================
// Project Funding Endpoints
// ============================================================================

// handleProjectFunding gets all funding sources for a project (v0.5.10+)
func (s *Server) handleProjectFunding(w http.ResponseWriter, r *http.Request, projectID string) {
	if r.Method != http.MethodGet {
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Get project to populate name
	project, err := s.projectManager.GetProject(context.Background(), projectID)
	if err != nil {
		s.writeError(w, http.StatusNotFound, fmt.Sprintf("Project not found: %v", err))
		return
	}

	summary, err := s.budgetManager.GetProjectFundingSummary(context.Background(), projectID)
	if err != nil {
		// If no allocations found, return empty summary
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"project_id":            projectID,
			"project_name":          project.Name,
			"allocations":           []interface{}{},
			"budget_names":          map[string]string{},
			"total_allocated":       0.0,
			"total_spent":           0.0,
			"default_allocation_id": project.DefaultAllocationID,
		})
		return
	}

	// Populate project name and default allocation
	summary.ProjectName = project.Name
	summary.DefaultAllocationID = project.DefaultAllocationID

	_ = json.NewEncoder(w).Encode(summary)
}

// handleSetDefaultAllocation sets the default funding source for a project
func (s *Server) handleSetDefaultAllocation(w http.ResponseWriter, r *http.Request, projectID string) {
	if r.Method != http.MethodPut {
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		AllocationID *string `json:"allocation_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	ctx := context.Background()

	// If clearing the default allocation
	if req.AllocationID == nil {
		if err := s.projectManager.ClearDefaultAllocation(ctx, projectID); err != nil {
			s.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to clear default allocation: %v", err))
			return
		}

		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"project_id":    projectID,
			"allocation_id": nil,
			"message":       "Default allocation cleared successfully",
		})
		return
	}

	// Verify allocation exists and belongs to this project
	allocationID := *req.AllocationID
	allocation, err := s.budgetManager.GetAllocation(ctx, allocationID)
	if err != nil {
		s.writeError(w, http.StatusNotFound, fmt.Sprintf("Allocation not found: %v", err))
		return
	}
	if allocation.ProjectID != projectID {
		s.writeError(w, http.StatusBadRequest, "Allocation does not belong to this project")
		return
	}

	// Set the default allocation
	if err := s.projectManager.SetDefaultAllocation(ctx, projectID, allocationID); err != nil {
		s.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to set default allocation: %v", err))
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"project_id":    projectID,
		"allocation_id": allocationID,
		"message":       "Default allocation set successfully",
	})
}

// handleAllocationStatus checks the status of an allocation (v0.5.10+ Issue #234)
func (s *Server) handleAllocationStatus(w http.ResponseWriter, r *http.Request, allocationID string) {
	if r.Method != http.MethodGet {
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	ctx := context.Background()

	exhausted, remaining, err := s.budgetManager.CheckAllocationStatus(ctx, allocationID)
	if err != nil {
		s.writeError(w, http.StatusNotFound, fmt.Sprintf("Allocation not found: %v", err))
		return
	}

	allocation, err := s.budgetManager.GetAllocation(ctx, allocationID)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get allocation: %v", err))
		return
	}

	// Get backup allocation info if configured
	var backupInfo map[string]interface{}
	if allocation.BackupAllocationID != nil && *allocation.BackupAllocationID != "" {
		backupAlloc, err := s.budgetManager.GetAllocation(ctx, *allocation.BackupAllocationID)
		if err == nil {
			backupExhausted, backupRemaining, _ := s.budgetManager.CheckAllocationStatus(ctx, *allocation.BackupAllocationID)
			backupInfo = map[string]interface{}{
				"backup_allocation_id": *allocation.BackupAllocationID,
				"backup_exhausted":     backupExhausted,
				"backup_remaining":     backupRemaining,
				"backup_allocated":     backupAlloc.AllocatedAmount,
				"backup_spent":         backupAlloc.SpentAmount,
			}
		}
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"allocation_id": allocationID,
		"exhausted":     exhausted,
		"remaining":     remaining,
		"allocated":     allocation.AllocatedAmount,
		"spent":         allocation.SpentAmount,
		"backup":        backupInfo,
	})
}

// ============================================================================
// Reallocation Endpoints (v0.5.10+ Issue #99)
// ============================================================================

// handleReallocationOperations routes reallocation-related requests (v0.5.10+ Issue #99)
func (s *Server) handleReallocationOperations(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		s.handleCreateReallocation(w, r)
	default:
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleCreateReallocation creates a new funds reallocation (v0.5.10+ Issue #99)
func (s *Server) handleCreateReallocation(w http.ResponseWriter, r *http.Request) {
	var req project.ReallocateFundsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	record, err := s.budgetManager.ReallocateFunds(context.Background(), &req)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to reallocate funds: %v", err))
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(record)
}

// handleBudgetReallocationHistory retrieves reallocation history for a budget (v0.5.10+ Issue #99)
func (s *Server) handleBudgetReallocationHistory(w http.ResponseWriter, r *http.Request, budgetID string) {
	if r.Method != http.MethodGet {
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	records, err := s.budgetManager.GetBudgetReallocationHistory(context.Background(), budgetID)
	if err != nil {
		s.writeError(w, http.StatusNotFound, fmt.Sprintf("Failed to get budget reallocation history: %v", err))
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"budget_id":     budgetID,
		"reallocations": records,
		"count":         len(records),
	})
}

// handleAllocationReallocationHistory retrieves reallocation history for an allocation (v0.5.10+ Issue #99)
func (s *Server) handleAllocationReallocationHistory(w http.ResponseWriter, r *http.Request, allocationID string) {
	if r.Method != http.MethodGet {
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	records, err := s.budgetManager.GetReallocationHistory(context.Background(), allocationID)
	if err != nil {
		s.writeError(w, http.StatusNotFound, fmt.Sprintf("Failed to get allocation reallocation history: %v", err))
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"allocation_id": allocationID,
		"reallocations": records,
		"count":         len(records),
	})
}

// ============================================================================
// Cost Rollup and Reporting Endpoints (v0.5.10+ Issue #100)
// ============================================================================

// handleBudgetRollupReport generates a comprehensive rollup report (v0.5.10+ Issue #100)
func (s *Server) handleBudgetRollupReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	report, err := s.budgetManager.GenerateBudgetRollupReport(context.Background())
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to generate rollup report: %v", err))
		return
	}

	// Populate project names if project manager is available
	if s.projectManager != nil {
		for i := range report.BudgetSummaries {
			for j := range report.BudgetSummaries[i].Projects {
				proj := &report.BudgetSummaries[i].Projects[j]
				if project, err := s.projectManager.GetProject(context.Background(), proj.ProjectID); err == nil {
					proj.ProjectName = project.Name
				}
			}
		}
	}

	_ = json.NewEncoder(w).Encode(report)
}

// handleBudgetSummaryReport generates a detailed summary for a specific budget (v0.5.10+ Issue #100)
func (s *Server) handleBudgetSummaryReport(w http.ResponseWriter, r *http.Request, budgetID string) {
	if r.Method != http.MethodGet {
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	summary, err := s.budgetManager.GetBudgetSummaryReport(context.Background(), budgetID)
	if err != nil {
		s.writeError(w, http.StatusNotFound, fmt.Sprintf("Failed to get budget summary: %v", err))
		return
	}

	// Populate project names if project manager is available
	if s.projectManager != nil {
		for i := range summary.Projects {
			proj := &summary.Projects[i]
			if project, err := s.projectManager.GetProject(context.Background(), proj.ProjectID); err == nil {
				proj.ProjectName = project.Name
			}
		}
	}

	_ = json.NewEncoder(w).Encode(summary)
}

// handleProjectCostRollup generates cost rollup for specific projects (v0.5.10+ Issue #100)
func (s *Server) handleProjectCostRollup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		ProjectIDs []string `json:"project_ids"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	if len(req.ProjectIDs) == 0 {
		s.writeError(w, http.StatusBadRequest, "project_ids is required")
		return
	}

	summaries, err := s.budgetManager.GetProjectCostRollup(context.Background(), req.ProjectIDs)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to generate project cost rollup: %v", err))
		return
	}

	// Populate project names if project manager is available
	if s.projectManager != nil {
		for i := range summaries {
			proj := &summaries[i]
			if project, err := s.projectManager.GetProject(context.Background(), proj.ProjectID); err == nil {
				proj.ProjectName = project.Name
			}
		}
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"projects": summaries,
		"count":    len(summaries),
	})
}
