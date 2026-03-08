package project

import (
	"errors"
	"fmt"
	"time"

	"github.com/scttfrdmn/prism/pkg/types"
)

// Error constants
var (
	// ErrDuplicateProjectName is returned when attempting to create a project with a name that already exists
	ErrDuplicateProjectName = errors.New("project name already exists")
)

// CreateProjectRequest represents a request to create a new project
type CreateProjectRequest struct {
	// Name is the project name (required)
	Name string `json:"name"`

	// Description provides project details
	Description string `json:"description"`

	// Owner is the project owner/principal investigator
	Owner string `json:"owner"`

	// Tags for project organization
	Tags map[string]string `json:"tags,omitempty"`

	// Budget contains optional budget configuration (DEPRECATED in v0.5.10)
	// Use Budget/Allocation system instead
	Budget *CreateProjectBudgetRequest `json:"budget,omitempty"`
}

// Validate validates the create project request
func (r *CreateProjectRequest) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("project name is required")
	}

	if len(r.Name) > 100 {
		return fmt.Errorf("project name cannot exceed 100 characters")
	}

	if len(r.Description) > 1000 {
		return fmt.Errorf("project description cannot exceed 1000 characters")
	}

	if r.Budget != nil {
		if err := r.Budget.Validate(); err != nil {
			return fmt.Errorf("invalid budget configuration: %w", err)
		}
	}

	return nil
}

// CreateProjectBudgetRequest represents a request to create a project budget (DEPRECATED in v0.5.10)
// Use CreateBudgetRequest + CreateAllocationRequest instead
type CreateProjectBudgetRequest struct {
	// TotalBudget is the total project budget in USD
	TotalBudget float64 `json:"total_budget"`

	// MonthlyLimit is the optional monthly spending limit in USD
	MonthlyLimit *float64 `json:"monthly_limit,omitempty"`

	// DailyLimit is the optional daily spending limit in USD
	DailyLimit *float64 `json:"daily_limit,omitempty"`

	// AlertThresholds define when to send budget alerts
	AlertThresholds []types.BudgetAlert `json:"alert_thresholds,omitempty"`

	// AutoActions define automatic actions when thresholds are reached
	AutoActions []types.BudgetAutoAction `json:"auto_actions,omitempty"`

	// BudgetPeriod defines the budget period
	BudgetPeriod types.BudgetPeriod `json:"budget_period"`

	// EndDate is when the budget period ends (optional)
	EndDate *time.Time `json:"end_date,omitempty"`
}

// Validate validates the create project budget request (DEPRECATED in v0.5.10)
func (r *CreateProjectBudgetRequest) Validate() error {
	if r.TotalBudget <= 0 {
		return fmt.Errorf("total budget must be greater than 0")
	}

	if r.MonthlyLimit != nil && *r.MonthlyLimit <= 0 {
		return fmt.Errorf("monthly limit must be greater than 0")
	}

	if r.DailyLimit != nil && *r.DailyLimit <= 0 {
		return fmt.Errorf("daily limit must be greater than 0")
	}

	// Validate alert thresholds
	for i, alert := range r.AlertThresholds {
		if alert.Threshold < 0 || alert.Threshold > 1 {
			return fmt.Errorf("alert threshold %d must be between 0.0 and 1.0", i)
		}
	}

	// Validate auto actions
	for i, action := range r.AutoActions {
		if action.Threshold < 0 || action.Threshold > 1 {
			return fmt.Errorf("auto action threshold %d must be between 0.0 and 1.0", i)
		}
	}

	return nil
}

// UpdateProjectRequest represents a request to update a project
type UpdateProjectRequest struct {
	// Name is the new project name (optional)
	Name *string `json:"name,omitempty"`

	// Description is the new project description (optional)
	Description *string `json:"description,omitempty"`

	// Tags are the new project tags (optional)
	Tags map[string]string `json:"tags,omitempty"`

	// Status is the new project status (optional)
	Status *types.ProjectStatus `json:"status,omitempty"`
}

// ProjectFilter defines filtering options for listing projects
type ProjectFilter struct {
	// Owner filters by project owner
	Owner string `json:"owner,omitempty"`

	// Status filters by project status
	Status *types.ProjectStatus `json:"status,omitempty"`

	// Tags filters by project tags (all specified tags must match)
	Tags map[string]string `json:"tags,omitempty"`

	// CreatedAfter filters projects created after this date
	CreatedAfter *time.Time `json:"created_after,omitempty"`

	// CreatedBefore filters projects created before this date
	CreatedBefore *time.Time `json:"created_before,omitempty"`

	// HasBudget filters projects with/without budgets
	HasBudget *bool `json:"has_budget,omitempty"`
}

// Matches checks if a project matches the filter criteria
func (f *ProjectFilter) Matches(project *types.Project) bool {
	// Check basic project attributes
	if !f.matchesBasicAttributes(project) {
		return false
	}

	// Check date-based filters
	if !f.matchesDateFilters(project) {
		return false
	}

	// Check tag-based filters
	if !f.matchesTagFilters(project) {
		return false
	}

	return true
}

// matchesBasicAttributes checks owner, status, and budget filters
func (f *ProjectFilter) matchesBasicAttributes(project *types.Project) bool {
	// Check owner filter
	if !f.matchesOwnerFilter(project) {
		return false
	}

	// Check status filter
	if !f.matchesStatusFilter(project) {
		return false
	}

	// Check budget filter
	if !f.matchesBudgetFilter(project) {
		return false
	}

	return true
}

// matchesOwnerFilter checks if project matches the owner filter
func (f *ProjectFilter) matchesOwnerFilter(project *types.Project) bool {
	return f.Owner == "" || project.Owner == f.Owner
}

// matchesStatusFilter checks if project matches the status filter
func (f *ProjectFilter) matchesStatusFilter(project *types.Project) bool {
	return f.Status == nil || project.Status == *f.Status
}

// matchesBudgetFilter checks if project matches the budget filter
func (f *ProjectFilter) matchesBudgetFilter(project *types.Project) bool {
	if f.HasBudget == nil {
		return true
	}

	hasBudget := project.Budget != nil
	return hasBudget == *f.HasBudget
}

// matchesDateFilters checks creation date filters
func (f *ProjectFilter) matchesDateFilters(project *types.Project) bool {
	// Check created after filter
	if f.CreatedAfter != nil && project.CreatedAt.Before(*f.CreatedAfter) {
		return false
	}

	// Check created before filter
	if f.CreatedBefore != nil && project.CreatedAt.After(*f.CreatedBefore) {
		return false
	}

	return true
}

// matchesTagFilters checks if all specified tags match
func (f *ProjectFilter) matchesTagFilters(project *types.Project) bool {
	if len(f.Tags) == 0 {
		return true
	}

	if project.Tags == nil {
		return false
	}

	// All specified tags must match
	for key, value := range f.Tags {
		if projectValue, exists := project.Tags[key]; !exists || projectValue != value {
			return false
		}
	}

	return true
}

// BudgetStatus represents the current budget status of a project
type BudgetStatus struct {
	// ProjectID is the project identifier
	ProjectID string `json:"project_id"`

	// BudgetEnabled indicates if budget tracking is enabled
	BudgetEnabled bool `json:"budget_enabled"`

	// TotalBudget is the total project budget
	TotalBudget float64 `json:"total_budget"`

	// SpentAmount is the current amount spent
	SpentAmount float64 `json:"spent_amount"`

	// RemainingBudget is the remaining budget
	RemainingBudget float64 `json:"remaining_budget"`

	// SpentPercentage is the percentage of budget spent (0.0-1.0)
	SpentPercentage float64 `json:"spent_percentage"`

	// ProjectedMonthlySpend is the projected monthly spending based on current usage
	ProjectedMonthlySpend float64 `json:"projected_monthly_spend"`

	// DaysUntilBudgetExhausted estimates when budget will be exhausted at current rate
	DaysUntilBudgetExhausted *int `json:"days_until_exhausted,omitempty"`

	// BurnRate contains period-aware burn rate analysis (nil when history is insufficient).
	BurnRate *BurnRateInfo `json:"burn_rate,omitempty"`

	// Surplus contains banking and carry-over information (nil for project-lifetime budgets).
	Surplus *SurplusInfo `json:"surplus,omitempty"`

	// ActiveAlerts are currently active budget alerts
	ActiveAlerts []string `json:"active_alerts"`

	// TriggeredActions are actions that have been triggered
	TriggeredActions []string `json:"triggered_actions"`

	// LastUpdated is when this status was calculated
	LastUpdated time.Time `json:"last_updated"`
}

// ProjectSummary provides a condensed view of project information
type ProjectSummary struct {
	// ID is the project identifier
	ID string `json:"id"`

	// Name is the project name
	Name string `json:"name"`

	// Owner is the project owner
	Owner string `json:"owner"`

	// Status is the project status
	Status types.ProjectStatus `json:"status"`

	// MemberCount is the number of project members
	MemberCount int `json:"member_count"`

	// ActiveInstances is the number of currently active instances
	ActiveInstances int `json:"active_instances"`

	// TotalCost is the total project cost to date
	TotalCost float64 `json:"total_cost"`

	// BudgetStatus provides budget information if budget is enabled
	BudgetStatus *BudgetStatusSummary `json:"budget_status,omitempty"`

	// CreatedAt is when the project was created
	CreatedAt time.Time `json:"created_at"`

	// LastActivity is when the project had its last activity
	LastActivity time.Time `json:"last_activity"`
}

// BudgetStatusSummary provides a condensed view of budget status
type BudgetStatusSummary struct {
	// TotalBudget is the total project budget
	TotalBudget float64 `json:"total_budget"`

	// SpentAmount is the current amount spent
	SpentAmount float64 `json:"spent_amount"`

	// SpentPercentage is the percentage of budget spent (0.0-1.0)
	SpentPercentage float64 `json:"spent_percentage"`

	// AlertCount is the number of active alerts
	AlertCount int `json:"alert_count"`
}

// ProjectListResponse represents the response for listing projects
type ProjectListResponse struct {
	// Projects are the matching projects
	Projects []ProjectSummary `json:"projects"`

	// TotalCount is the total number of projects (before pagination)
	TotalCount int `json:"total_count"`

	// FilteredCount is the number of projects matching the filter
	FilteredCount int `json:"filtered_count"`
}

// AddMemberRequest represents a request to add a member to a project
type AddMemberRequest struct {
	// UserID is the user identifier
	UserID string `json:"user_id"`

	// Role is the project role for the member
	Role types.ProjectRole `json:"role"`

	// AddedBy is who is adding the member
	AddedBy string `json:"added_by"`
}

// Validate validates the add member request
func (r *AddMemberRequest) Validate() error {
	if r.UserID == "" {
		return fmt.Errorf("user ID is required")
	}

	if r.Role == "" {
		return fmt.Errorf("role is required")
	}

	// Validate role
	switch r.Role {
	case types.ProjectRoleOwner, types.ProjectRoleAdmin, types.ProjectRoleMember, types.ProjectRoleViewer:
		// Valid roles
	default:
		return fmt.Errorf("invalid role: %s", r.Role)
	}

	return nil
}

// UpdateMemberRequest represents a request to update a project member
type UpdateMemberRequest struct {
	// Role is the new role for the member
	Role types.ProjectRole `json:"role"`
}

// Validate validates the update member request
func (r *UpdateMemberRequest) Validate() error {
	if r.Role == "" {
		return fmt.Errorf("role is required")
	}

	// Validate role
	switch r.Role {
	case types.ProjectRoleOwner, types.ProjectRoleAdmin, types.ProjectRoleMember, types.ProjectRoleViewer:
		// Valid roles
	default:
		return fmt.Errorf("invalid role: %s", r.Role)
	}

	return nil
}

// ============================================================================
// v0.5.10: Multi-Budget System Request Types
// ============================================================================

// CreateBudgetRequest represents a request to create a budget pool (v0.5.10+)
type CreateBudgetRequest struct {
	// Name is the budget name (e.g., "NSF Grant CISE-2024-12345")
	Name string `json:"name"`

	// Description provides budget details
	Description string `json:"description"`

	// TotalAmount is the total budget pool in USD
	TotalAmount float64 `json:"total_amount"`

	// Period defines the budget period
	Period types.BudgetPeriod `json:"period"`

	// StartDate is when the budget period began
	StartDate time.Time `json:"start_date"`

	// EndDate is when the budget period ends (optional for ongoing budgets)
	EndDate *time.Time `json:"end_date,omitempty"`

	// AlertThreshold is the global alert percentage (0.0-1.0)
	AlertThreshold float64 `json:"alert_threshold"`

	// CreatedBy is the user who created the budget
	CreatedBy string `json:"created_by"`

	// Tags for budget organization
	Tags map[string]string `json:"tags,omitempty"`
}

// Validate validates the create budget request
func (r *CreateBudgetRequest) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("budget name is required")
	}

	if len(r.Name) > 100 {
		return fmt.Errorf("budget name cannot exceed 100 characters")
	}

	if r.TotalAmount <= 0 {
		return fmt.Errorf("total amount must be greater than 0")
	}

	if r.AlertThreshold < 0 || r.AlertThreshold > 1 {
		return fmt.Errorf("alert threshold must be between 0.0 and 1.0")
	}

	if r.Period == "" {
		return fmt.Errorf("budget period is required")
	}

	if r.CreatedBy == "" {
		return fmt.Errorf("created_by is required")
	}

	return nil
}

// UpdateBudgetRequest represents a request to update a budget
type UpdateBudgetRequest struct {
	// Name is the new budget name (optional)
	Name *string `json:"name,omitempty"`

	// Description is the new budget description (optional)
	Description *string `json:"description,omitempty"`

	// TotalAmount is the new total amount (optional)
	TotalAmount *float64 `json:"total_amount,omitempty"`

	// AlertThreshold is the new alert threshold (optional)
	AlertThreshold *float64 `json:"alert_threshold,omitempty"`

	// EndDate is the new end date (optional)
	EndDate *time.Time `json:"end_date,omitempty"`

	// Tags are the new budget tags (optional)
	Tags map[string]string `json:"tags,omitempty"`
}

// CreateAllocationRequest represents a request to allocate budget to a project
type CreateAllocationRequest struct {
	// BudgetID is the parent budget pool
	BudgetID string `json:"budget_id"`

	// ProjectID is the project receiving the allocation
	ProjectID string `json:"project_id"`

	// AllocatedAmount is how much of the budget to allocate to this project
	AllocatedAmount float64 `json:"allocated_amount"`

	// AlertThreshold is an optional project-specific alert threshold (overrides budget default)
	AlertThreshold *float64 `json:"alert_threshold,omitempty"`

	// BackupAllocationID is an optional backup funding source for exhaustion (#234)
	BackupAllocationID *string `json:"backup_allocation_id,omitempty"`

	// Notes provide context for this allocation
	Notes string `json:"notes,omitempty"`

	// AllocatedBy is the user who created the allocation
	AllocatedBy string `json:"allocated_by"`
}

// Validate validates the create allocation request
func (r *CreateAllocationRequest) Validate() error {
	if r.BudgetID == "" {
		return fmt.Errorf("budget_id is required")
	}

	if r.ProjectID == "" {
		return fmt.Errorf("project_id is required")
	}

	if r.AllocatedAmount <= 0 {
		return fmt.Errorf("allocated amount must be greater than 0")
	}

	if r.AlertThreshold != nil && (*r.AlertThreshold < 0 || *r.AlertThreshold > 1) {
		return fmt.Errorf("alert threshold must be between 0.0 and 1.0")
	}

	if r.AllocatedBy == "" {
		return fmt.Errorf("allocated_by is required")
	}

	return nil
}

// UpdateAllocationRequest represents a request to update an allocation
type UpdateAllocationRequest struct {
	// AllocatedAmount is the new allocated amount (optional)
	AllocatedAmount *float64 `json:"allocated_amount,omitempty"`

	// AlertThreshold is the new alert threshold (optional)
	AlertThreshold *float64 `json:"alert_threshold,omitempty"`

	// BackupAllocationID is the new backup allocation (optional)
	BackupAllocationID *string `json:"backup_allocation_id,omitempty"`

	// Notes are the new notes (optional)
	Notes *string `json:"notes,omitempty"`
}

// ReallocateFundsRequest represents a request to reallocate funds between allocations (v0.5.10+ Issue #99)
type ReallocateFundsRequest struct {
	// SourceAllocationID is where funds are moved from
	SourceAllocationID string `json:"source_allocation_id"`

	// DestinationAllocationID is where funds are moved to
	DestinationAllocationID string `json:"destination_allocation_id"`

	// Amount is how much to reallocate (USD)
	Amount float64 `json:"amount"`

	// Reason explains why the reallocation is being made (required for audit)
	Reason string `json:"reason"`

	// PerformedBy is the user making the reallocation
	PerformedBy string `json:"performed_by"`
}

// Validate validates the reallocate funds request
func (r *ReallocateFundsRequest) Validate() error {
	if r.SourceAllocationID == "" {
		return fmt.Errorf("source_allocation_id is required")
	}

	if r.DestinationAllocationID == "" {
		return fmt.Errorf("destination_allocation_id is required")
	}

	if r.SourceAllocationID == r.DestinationAllocationID {
		return fmt.Errorf("source and destination allocations cannot be the same")
	}

	if r.Amount <= 0 {
		return fmt.Errorf("reallocation amount must be greater than 0")
	}

	if r.Reason == "" {
		return fmt.Errorf("reason is required for audit trail")
	}

	if len(r.Reason) > 500 {
		return fmt.Errorf("reason cannot exceed 500 characters")
	}

	if r.PerformedBy == "" {
		return fmt.Errorf("performed_by is required")
	}

	return nil
}

// ReallocationRecord tracks a funds reallocation for audit purposes (v0.5.10+ Issue #99)
type ReallocationRecord struct {
	// ID is the unique reallocation identifier
	ID string `json:"id"`

	// SourceAllocationID is where funds were moved from
	SourceAllocationID string `json:"source_allocation_id"`

	// DestinationAllocationID is where funds were moved to
	DestinationAllocationID string `json:"destination_allocation_id"`

	// SourceBudgetID is the source budget (for cross-budget tracking)
	SourceBudgetID string `json:"source_budget_id"`

	// DestinationBudgetID is the destination budget (for cross-budget tracking)
	DestinationBudgetID string `json:"destination_budget_id"`

	// Amount is how much was reallocated (USD)
	Amount float64 `json:"amount"`

	// Reason explains why the reallocation was made
	Reason string `json:"reason"`

	// PerformedBy is the user who made the reallocation
	PerformedBy string `json:"performed_by"`

	// Timestamp is when the reallocation occurred
	Timestamp time.Time `json:"timestamp"`
}

// ============================================================================
// v0.5.10: Multi-Project Cost Rollup and Reporting Types (Issue #100)
// ============================================================================

// BudgetRollupReport provides aggregated cost and budget information (v0.5.10+ Issue #100)
type BudgetRollupReport struct {
	// ReportID is a unique identifier for this report
	ReportID string `json:"report_id"`

	// GeneratedAt is when this report was generated
	GeneratedAt time.Time `json:"generated_at"`

	// BudgetSummaries provides per-budget rollup information
	BudgetSummaries []BudgetSummaryReport `json:"budget_summaries"`

	// TotalBudgets is the total number of budgets in the report
	TotalBudgets int `json:"total_budgets"`

	// TotalAllocated is the sum of all budget allocations
	TotalAllocated float64 `json:"total_allocated"`

	// TotalSpent is the sum of all spending across all budgets
	TotalSpent float64 `json:"total_spent"`

	// TotalRemaining is the total remaining across all budgets
	TotalRemaining float64 `json:"total_remaining"`

	// OverallUtilization is the percentage of total budget utilized (0.0-1.0)
	OverallUtilization float64 `json:"overall_utilization"`

	// ProjectCount is the total number of projects funded
	ProjectCount int `json:"project_count"`
}

// BudgetSummaryReport provides detailed budget information (v0.5.10+ Issue #100)
type BudgetSummaryReport struct {
	// BudgetID is the budget identifier
	BudgetID string `json:"budget_id"`

	// BudgetName is the budget name
	BudgetName string `json:"budget_name"`

	// Period is the budget period
	Period string `json:"period"`

	// TotalAmount is the total budget amount
	TotalAmount float64 `json:"total_amount"`

	// AllocatedAmount is how much has been allocated to projects
	AllocatedAmount float64 `json:"allocated_amount"`

	// SpentAmount is how much has been spent
	SpentAmount float64 `json:"spent_amount"`

	// RemainingAmount is the remaining unspent budget
	RemainingAmount float64 `json:"remaining_amount"`

	// Utilization is the percentage of budget utilized (0.0-1.0)
	Utilization float64 `json:"utilization"`

	// ProjectCount is the number of projects funded by this budget
	ProjectCount int `json:"project_count"`

	// AllocationCount is the number of allocations from this budget
	AllocationCount int `json:"allocation_count"`

	// Projects provides per-project information
	Projects []ProjectCostSummary `json:"projects"`
}

// ProjectCostSummary provides project-level cost information (v0.5.10+ Issue #100)
type ProjectCostSummary struct {
	// ProjectID is the project identifier
	ProjectID string `json:"project_id"`

	// ProjectName is the project name
	ProjectName string `json:"project_name"`

	// FundingSources lists all funding allocations for this project
	FundingSources []AllocationSummary `json:"funding_sources"`

	// TotalAllocated is the total allocated to this project across all sources
	TotalAllocated float64 `json:"total_allocated"`

	// TotalSpent is the total spent by this project
	TotalSpent float64 `json:"total_spent"`

	// TotalRemaining is the remaining budget for this project
	TotalRemaining float64 `json:"total_remaining"`

	// Utilization is the percentage of allocated budget spent (0.0-1.0)
	Utilization float64 `json:"utilization"`
}

// AllocationSummary provides allocation-level information (v0.5.10+ Issue #100)
type AllocationSummary struct {
	// AllocationID is the allocation identifier
	AllocationID string `json:"allocation_id"`

	// BudgetID is the parent budget identifier
	BudgetID string `json:"budget_id"`

	// BudgetName is the parent budget name
	BudgetName string `json:"budget_name"`

	// AllocatedAmount is how much was allocated
	AllocatedAmount float64 `json:"allocated_amount"`

	// SpentAmount is how much has been spent
	SpentAmount float64 `json:"spent_amount"`

	// RemainingAmount is the remaining allocation
	RemainingAmount float64 `json:"remaining_amount"`

	// Utilization is the percentage of allocation spent (0.0-1.0)
	Utilization float64 `json:"utilization"`

	// HasBackup indicates if backup funding is configured
	HasBackup bool `json:"has_backup"`
}

// ============================================================================
// Project Transfer and Forecast Types (Issue #326)
// ============================================================================

// TransferProjectRequest represents a request to transfer project ownership
type TransferProjectRequest struct {
	// NewOwnerID is the user ID of the new project owner
	NewOwnerID string `json:"new_owner_id"`

	// TransferredBy is the user making the transfer (must be current owner)
	TransferredBy string `json:"transferred_by"`

	// Reason provides context for the transfer (optional)
	Reason string `json:"reason,omitempty"`
}

// Validate validates the transfer project request
func (r *TransferProjectRequest) Validate() error {
	if r.NewOwnerID == "" {
		return fmt.Errorf("new_owner_id is required")
	}

	if r.TransferredBy == "" {
		return fmt.Errorf("transferred_by is required")
	}

	if len(r.Reason) > 500 {
		return fmt.Errorf("reason cannot exceed 500 characters")
	}

	return nil
}

// ProjectForecastRequest represents a request for cost forecast data
type ProjectForecastRequest struct {
	// Months is the number of months to forecast (default: 6)
	Months int `json:"months,omitempty"`

	// IncludeHistorical includes historical spending data in response
	IncludeHistorical bool `json:"include_historical,omitempty"`
}

// ProjectForecastResponse provides cost forecast information
type ProjectForecastResponse struct {
	// ProjectID is the project identifier
	ProjectID string `json:"project_id"`

	// GeneratedAt is when this forecast was generated
	GeneratedAt time.Time `json:"generated_at"`

	// CurrentMonthlyRate is the current monthly spending rate
	CurrentMonthlyRate float64 `json:"current_monthly_rate"`

	// ForecastData contains month-by-month projections
	ForecastData []ForecastDataPoint `json:"forecast_data"`

	// HistoricalData contains past spending (if requested)
	HistoricalData []ForecastDataPoint `json:"historical_data,omitempty"`

	// ProjectedExhaustion estimates when budget will be exhausted (if applicable)
	ProjectedExhaustion *time.Time `json:"projected_exhaustion,omitempty"`

	// Confidence indicates forecast confidence level (0.0-1.0)
	Confidence float64 `json:"confidence"`
}

// ForecastDataPoint represents a single month's forecast
type ForecastDataPoint struct {
	// Month is the month for this data point (YYYY-MM format)
	Month string `json:"month"`

	// ProjectedCost is the projected spending for this month
	ProjectedCost float64 `json:"projected_cost"`

	// CumulativeCost is the cumulative cost to date
	CumulativeCost float64 `json:"cumulative_cost"`

	// ActualCost is the actual cost (for historical data only)
	ActualCost *float64 `json:"actual_cost,omitempty"`
}
