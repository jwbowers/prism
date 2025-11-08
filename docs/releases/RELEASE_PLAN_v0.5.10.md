# Prism v0.5.10 Release Plan: Multi-Project Budgets

**Release Date**: Target February 14, 2026
**Focus**: Budget system redesign enabling multi-project allocation

## 🎯 Release Goals

### Primary Objective
Redesign the budget system to support many-to-many relationships between budgets and projects, enabling realistic research funding workflows.

**Current State**: 1 budget : 1 project (rigid, doesn't match research reality)
**New State**: Many-to-many (flexible, matches real funding models)

**Supported Scenarios**:
- ✅ **1 budget → N projects**: Single NSF grant funding 3-5 related projects
- ✅ **N budgets → 1 project**: Multi-source funding (NSF + department + AWS credits)

### Success Metrics
- Grant-funded research: Single NSF grant → 3-5 related projects
- Lab budgets: Department budget → 10+ research group projects
- Multi-source projects: 2-3 funding sources per major project
- Budget reallocation: Move funds between projects within 60 seconds

---

## 📦 Features & Implementation

### 1. Shared Budget Pools
**Priority**: P0 (Core requirement)
**Effort**: Large (4-5 days)
**Impact**: Critical (Enables multi-project allocation)

**Current Budget Model**:
```go
type Budget struct {
    ID          string
    ProjectID   string  // 1:1 relationship
    Amount      float64
    Period      string
    AlertThreshold float64
}
```

**New Budget Model**:
```go
type Budget struct {
    ID              string
    Name            string        // "NSF Grant #12345", "Department Q1 Budget"
    Description     string
    TotalAmount     float64       // Total budget pool
    Period          string        // "monthly", "quarterly", "grant-period"
    StartDate       time.Time
    EndDate         *time.Time    // Optional for ongoing budgets
    AlertThreshold  float64       // Percentage for global alert
    CreatedBy       string        // User ID
    CreatedAt       time.Time
}

type ProjectBudgetAllocation struct {
    ID              string
    BudgetID        string        // Parent budget pool
    ProjectID       string        // Allocated project
    AllocatedAmount float64       // Amount allocated to this project
    SpentAmount     float64       // Current spending (cached)
    AlertThreshold  *float64      // Optional project-specific threshold
    Notes           string
    AllocatedAt     time.Time
    AllocatedBy     string        // User ID
}
```

**Database Schema**:
```sql
-- budgets table (existing, modified)
ALTER TABLE budgets ADD COLUMN name TEXT NOT NULL;
ALTER TABLE budgets ADD COLUMN description TEXT;
ALTER TABLE budgets DROP COLUMN project_id;  -- Remove 1:1 constraint
ALTER TABLE budgets ADD COLUMN start_date TIMESTAMP NOT NULL DEFAULT NOW();
ALTER TABLE budgets ADD COLUMN end_date TIMESTAMP;
ALTER TABLE budgets ADD COLUMN created_by TEXT;
ALTER TABLE budgets ADD COLUMN created_at TIMESTAMP NOT NULL DEFAULT NOW();

-- project_budget_allocations table (new)
CREATE TABLE project_budget_allocations (
    id TEXT PRIMARY KEY,
    budget_id TEXT NOT NULL REFERENCES budgets(id) ON DELETE CASCADE,
    project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    allocated_amount REAL NOT NULL CHECK (allocated_amount >= 0),
    spent_amount REAL NOT NULL DEFAULT 0,
    alert_threshold REAL,
    notes TEXT,
    allocated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    allocated_by TEXT NOT NULL,
    UNIQUE(budget_id, project_id)  -- One allocation per budget-project pair
);

CREATE INDEX idx_allocations_budget ON project_budget_allocations(budget_id);
CREATE INDEX idx_allocations_project ON project_budget_allocations(project_id);
```

**API Endpoints**:
```
POST   /api/v1/budgets                       # Create shared budget pool
GET    /api/v1/budgets                       # List all budgets
GET    /api/v1/budgets/{id}                  # Get budget with allocations
PUT    /api/v1/budgets/{id}                  # Update budget
DELETE /api/v1/budgets/{id}                  # Delete budget (cascades allocations)

POST   /api/v1/budgets/{id}/allocations      # Allocate budget to project
GET    /api/v1/budgets/{id}/allocations      # List allocations for budget
PUT    /api/v1/allocations/{id}              # Update allocation amount
DELETE /api/v1/allocations/{id}              # Remove project from budget

GET    /api/v1/projects/{id}/budget          # Get project's budget allocation
```

**Implementation Tasks**:
- [ ] Update `pkg/types/project.go` with new budget structures
- [ ] Create database migration script
- [ ] Update `pkg/project/budget_tracker.go` for multi-project tracking
- [ ] Implement allocation validation (total allocations ≤ budget amount)
- [ ] Add allocation API endpoints in `pkg/daemon/project_handlers.go`
- [ ] Update cost calculation to aggregate by allocation
- [ ] Add budget pool summary (allocated, spent, remaining)

---

### 1.5. Multi-Budget Projects & Spending Attribution
**Priority**: P0 (Core requirement)
**Effort**: Medium (2-3 days)
**Impact**: Critical (Enables multi-source funding)

The many-to-many design naturally supports **multiple budgets funding a single project**:

**Example**:
```
Project: "Climate Model Development"
├─ NSF Grant CISE-2024-12345: $20,000 allocated
├─ CS Department Q1 2026: $5,000 allocated
└─ AWS Research Credits: $2,000 allocated
   Total funding: $27,000
```

**Common Use Cases**:
- Multi-source funding: NSF grant + department matching funds
- Grant transitions: Old grant winding down, new grant starting
- Shared infrastructure: Multiple PIs funding shared compute cluster
- Institutional matches: Grant requires university matching (tracked separately)
- Mixed funding: Grant money + research credits + donations

#### Spending Attribution (v0.5.10)

**Design Principle**: PIs/project owners control funding, not students/researchers.

**Two-Tier Approach**:

**1. Default Project Funding (Primary Workflow)**
- Project owners set **default funding allocation** when creating/configuring project
- Students/researchers launch workspaces under the project → automatic funding attribution
- No funding decisions required from end users
- Example: "All workspaces in 'ML Research' project use NSF Grant by default"

**2. Manual Override (Power User)**
- Project owners/admins can override funding at launch with `--funding` flag
- Useful for multi-source projects where different resources use different budgets
- Example: `prism launch gpu-ml training --project ml-research --funding "AWS Credits"`

**Launch Behavior**:
```
# Student/researcher (simple)
prism launch python-ml my-workspace --project ml-research
# ↳ Automatically uses project's default funding allocation

# Project owner (override)
prism launch python-ml my-workspace --project ml-research --funding "Department Matching"
# ↳ Uses specified allocation instead of default
```

**Implementation**:
- Add `default_allocation_id` to projects table
- Project launch automatically uses default allocation if not specified
- Add `funding_allocation_id` to workspace/volume launch API
- Track spending per allocation ID
- CLI: `--funding` flag optional (defaults to project's default)
- GUI: Funding selector hidden for regular users, shown for project owners

**Future Roadmap** (v0.6.0+):
- [ ] **Automatic proportional split** across allocations (policy-driven)
- [ ] **Waterfall spending** (exhaust Budget A, then Budget B) (policy-driven)
- [ ] **Policy-based attribution rules** (e.g., "GPU workloads use NSF, CPU uses department")

#### Alert Strategy

**Design Principle**: Prism is for research compute management, not PI grant management.

**Alerts trigger based on resource consumption, not budget-wide status**:
- ✅ Alert when a **workspace** nears its allocation threshold
- ✅ Alert when a **project's resources** collectively near allocation
- ❌ Don't alert PIs about overall grant spending (that's their grants office)

**Alert Levels**:
1. **Per-Allocation Alerts**: "Project X using NSF Grant is at 80% ($16k spent / $20k allocated)"
2. **Per-Resource Alerts**: "Workspace 'ml-training' has consumed $800 of its $1,000 budget"
3. **Cross-Budget Alerts**: "Project X total spending at 85% across all funding sources"

**Budget Visibility**: Users only see budgets they have access to based on project roles.

#### Budget Exhaustion & Emergency Funding

**Budget Cushion System** (v0.5.10):

When a primary allocation is exhausted:

1. **Backup Funding Source** (if configured):
   ```
   Primary: NSF Grant ($20k allocated, $20k spent) ❌ EXHAUSTED
   Backup: Department Emergency Fund ($2k available) ✅ ACTIVE
   ```
   - Automatically switch to backup funding
   - Notify user of switch
   - Log exhaustion event for audit

2. **No Backup** (default behavior):
   - Stop new launches from this allocation
   - Hibernate running workspaces using this allocation
   - Restrict access until budget resolved
   - Email project owner and admins
   - Display clear "Budget Exhausted" message

**Implementation**:
- Add `backup_allocation_id` to ProjectBudgetAllocation
- Pre-launch validation: Check funding source has available balance
- Real-time exhaustion detection during cost tracking
- Automated hibernation for resources using exhausted allocation
- Audit trail for budget switches and exhaustion events

**Future Enhancements** (v0.6.0+):
- [ ] Per-user budget cushions (faculty vs students)
- [ ] Temporary over-budget allowance with approval workflow
- [ ] Automatic budget extension requests

**Implementation Tasks**:
- [ ] Add `funding_allocation_id` to instance/volume tables
- [ ] Update launch API to accept allocation selection
- [ ] Implement pre-launch balance validation
- [ ] Add allocation exhaustion detection
- [ ] Implement backup funding switch logic
- [ ] Create resource hibernation on exhaustion
- [ ] Add backup_allocation_id to allocation schema
- [ ] Build funding source selector component
- [ ] Add exhaustion notification system
- [ ] Create audit trail for budget events

---

### 2. Project Budget Allocation Interface
**Priority**: P0 (Core UX)
**Effort**: Medium (3-4 days)
**Impact**: Critical (Primary user workflow)

**GUI Components**:

#### Budget Creation Dialog
```typescript
// cmd/prism-gui/frontend/src/components/BudgetCreateDialog.tsx
interface BudgetCreateForm {
  name: string;              // "NSF Grant CISE-2024-12345"
  description: string;       // Grant details, purpose
  totalAmount: number;       // $50,000
  period: 'monthly' | 'quarterly' | 'annual' | 'grant-period' | 'custom';
  startDate: Date;
  endDate?: Date;            // Optional for ongoing budgets
  alertThreshold: number;    // 80% (global)
}

Features:
- Budget name and description
- Total amount with currency formatting
- Period selection (preset + custom date range)
- Global alert threshold
- Initial project allocations (optional)
```

#### Budget Management Page
```typescript
// cmd/prism-gui/frontend/src/pages/Budgets.tsx
Features:
- Budget pool cards showing:
  - Name and description
  - Total amount vs spent
  - Number of projects allocated
  - Remaining unallocated funds
  - Alert status
  - Period and dates
- Create new budget button
- Filter by period, status, date range
- Search by name/description
```

#### Budget Allocation Dialog
```typescript
// cmd/prism-gui/frontend/src/components/BudgetAllocationDialog.tsx
interface AllocationForm {
  budgetId: string;          // Selected budget pool
  projects: Array<{
    projectId: string;
    name: string;
    allocatedAmount: number;
    currentSpending: number;
    notes: string;
  }>;
  validateTotal: boolean;    // Ensure allocations ≤ budget
}

Features:
- Multi-project allocation table
- Amount input per project
- Real-time validation (total ≤ budget amount)
- Project spending preview
- Bulk allocation (equal split, percentage-based)
- Notes per allocation
```

#### Project Budget Tab
```typescript
// cmd/prism-gui/frontend/src/pages/ProjectDetail.tsx (Budget Tab)
Features:
- Budget source information (parent budget pool)
- Allocated amount
- Current spending (with progress bar)
- Alert threshold (inherited or custom)
- Spending breakdown by service (EC2, EBS, EFS, etc.)
- Spending timeline chart
- Budget reallocation request (if not owner)
```

**Cloudscape Components**:
- `Table` with editable cells for allocation amounts
- `ProgressBar` for budget utilization
- `Alert` for budget threshold warnings
- `Modal` for create/edit dialogs
- `SpaceBetween` for form layout
- `FormField` with validation

**Implementation Tasks**:
- [ ] Create BudgetCreateDialog component
- [ ] Create BudgetManagementPage component
- [ ] Create BudgetAllocationDialog component
- [ ] Update ProjectDetail budget tab
- [ ] Add budget pool summary cards
- [ ] Implement allocation validation UI
- [ ] Add spending visualization charts

---

### 3. Budget Reallocation
**Priority**: P1 (Flexibility)
**Effort**: Small (2 days)
**Impact**: High (Real-world workflow)

**Use Cases**:
1. **Grant Budget Adjustment**: Project A is over budget, move $5k from Project B
2. **Research Pivot**: Project paused, reallocate funds to active projects
3. **End of Period**: Consolidate unused allocations

**Reallocation Workflow**:
```typescript
interface ReallocationRequest {
  sourceAllocationId: string;   // Project losing funds
  targetAllocationId: string;   // Project gaining funds
  amount: number;               // Amount to transfer
  reason: string;               // Audit trail
}

// API:
POST /api/v1/allocations/reallocate
{
  "sourceAllocationId": "alloc-1",
  "targetAllocationId": "alloc-2",
  "amount": 5000.00,
  "reason": "Project A exceeded compute budget for ML training"
}
```

**GUI Features**:
- Budget Reallocation Dialog (drag-and-drop between projects)
- Reallocation history table
- Audit trail with reason and timestamp
- Validation: source allocation has sufficient remaining funds

**Implementation Tasks**:
- [ ] Add reallocation API endpoint
- [ ] Implement atomic reallocation transaction
- [ ] Create ReallocationDialog component
- [ ] Add reallocation history view
- [ ] Add audit trail logging

---

### 4. Multi-Project Cost Rollup & Reporting
**Priority**: P1 (Visibility)
**Effort**: Medium (2-3 days)
**Impact**: High (Decision-making)

**Budget Dashboard Features**:
```typescript
interface BudgetDashboard {
  budget: Budget;
  totalAllocated: number;       // Sum of all allocations
  totalSpent: number;           // Sum of all project spending
  totalRemaining: number;       // Budget - spent
  unallocatedFunds: number;     // Budget - allocated
  allocations: Array<{
    project: Project;
    allocated: number;
    spent: number;
    remaining: number;
    percentUsed: number;
    alertStatus: 'ok' | 'warning' | 'critical';
  }>;
  spendingTimeline: Array<{
    date: Date;
    cumulative: number;
  }>;
  topSpenders: Project[];       // Top 5 projects by spending
}
```

**Visualizations**:
1. **Budget Utilization Chart**: Allocated vs Spent vs Remaining
2. **Project Spending Breakdown**: Pie chart of spending by project
3. **Spending Timeline**: Cumulative spending over time vs budget burn rate
4. **Service Cost Breakdown**: EC2, EBS, EFS costs across all projects
5. **Alert Status**: Projects approaching or exceeding allocations

**Export Features**:
- CSV export for accounting systems
- PDF report for grant reporting
- JSON API for custom integrations

**Implementation Tasks**:
- [ ] Create BudgetDashboard component
- [ ] Implement cost aggregation queries
- [ ] Add spending timeline calculation
- [ ] Create visualization components (charts)
- [ ] Add CSV/PDF export functionality
- [ ] Optimize performance for large project counts

---

### 5. Migration & Compatibility
**Priority**: P0 (Required for release)
**Effort**: Small (1-2 days)
**Impact**: Critical (Existing users)

**Migration Strategy**:
Since user feedback specifies "no need for backwards compatibility", we can simplify:

**Direct Migration**:
```sql
-- Migrate existing budgets
-- Each existing budget becomes a dedicated budget pool with single allocation
BEGIN TRANSACTION;

-- Add new columns to budgets table
ALTER TABLE budgets ADD COLUMN name TEXT;
ALTER TABLE budgets ADD COLUMN description TEXT;
UPDATE budgets SET name = 'Budget for ' || project_id WHERE name IS NULL;

-- Create project_budget_allocations table
CREATE TABLE project_budget_allocations (...);

-- Migrate existing budgets to allocations
INSERT INTO project_budget_allocations (
    id, budget_id, project_id, allocated_amount, allocated_by
)
SELECT
    'alloc-' || budget_id,
    budget_id,
    project_id,
    amount,
    'system-migration'
FROM budgets;

-- Remove project_id from budgets
ALTER TABLE budgets DROP COLUMN project_id;

COMMIT;
```

**API Migration**:
- Old endpoint: `GET /api/v1/projects/{id}/budget` → Returns allocation data
- New endpoints: Budget management APIs
- Remove deprecated endpoints in v0.6.0

**Implementation Tasks**:
- [ ] Create migration script
- [ ] Test migration with sample data
- [ ] Update API handlers for new schema
- [ ] Remove deprecated code paths
- [ ] Document breaking changes in release notes

---

## 📅 Implementation Schedule

### Week 1 (Feb 1-7): Backend & Data Model
**Days 1-2**: Database schema and migration
- Design new budget/allocation schema
- Write migration script
- Test with sample data

**Days 3-5**: API Implementation
- Update types and data structures
- Implement budget pool CRUD endpoints
- Implement allocation endpoints
- Add validation logic
- Write unit tests

### Week 2 (Feb 8-14): Frontend & Integration
**Days 1-2**: Budget Management UI
- BudgetCreateDialog component
- BudgetManagementPage component
- Budget pool summary cards

**Days 3-4**: Allocation UI
- BudgetAllocationDialog component
- Project budget tab updates
- Spending visualizations

**Day 5**: Testing & Polish
- Extended persona walkthroughs
- Performance testing (100+ projects per budget)
- Bug fixes
- Documentation

---

## 🧪 Testing Strategy

### Backend Testing
- [ ] Budget pool CRUD operations
- [ ] Allocation validation (total ≤ budget)
- [ ] Reallocation atomic transactions
- [ ] Cost aggregation accuracy
- [ ] Migration script (existing budgets → allocations)
- [ ] Performance (100+ projects, 10+ budgets)

### Frontend Testing
- [ ] Budget creation workflow
- [ ] Multi-project allocation
- [ ] Budget reallocation
- [ ] Spending visualization
- [ ] Alert threshold triggers
- [ ] CSV/PDF export

### Persona Walkthroughs

#### Grant-Funded Research (Multi-Project)
**Scenario**: PI receives $50k NSF grant for 3 related projects

1. Create budget pool "NSF CISE-2024-12345" ($50,000)
2. Allocate $20k to "Project A: Algorithm Development"
3. Allocate $15k to "Project B: User Study"
4. Allocate $10k to "Project C: Paper Experiments"
5. Launch workspaces under each project
6. Monitor spending across all 3 projects
7. Reallocate $5k from Project C to Project A (algorithm needs more compute)
8. Generate grant report showing spending by project

#### Lab Budget (Department Allocation)
**Scenario**: Lab manager receives $100k department budget for research group

1. Create budget pool "CS Lab Q1 2026" ($100,000)
2. Allocate $40k to "Genomics Project" (Prof. Smith)
3. Allocate $30k to "Climate Modeling" (Prof. Jones)
4. Allocate $20k to "ML Research" (Prof. Lee)
5. Leave $10k unallocated for emergency use
6. Each PI manages their project independently
7. Lab manager monitors total spending
8. Reallocate unused funds at end of quarter

#### Class Budget (Student Projects)
**Scenario**: Professor receives $5k for CS499 class with 25 student projects

1. Create budget pool "CS499 Spring 2026" ($5,000)
2. Allocate $150 per student project (25 × $150 = $3,750)
3. Leave $1,250 for instructor demos and reserves
4. Students launch workspaces under their projects
5. Monitor spending to prevent overruns
6. Hibernate/stop student workspaces approaching limits
7. Generate per-student spending report for grading

#### Multi-Source Funding (N Budgets → 1 Project)
**Scenario**: Major research project funded by multiple sources

1. Create project "Climate Simulation Infrastructure"
2. Create budget "NSF Grant CISE-2024-12345" ($50,000)
   - Allocate $50k to Climate Simulation project
3. Create budget "CS Department Matching" ($10,000)
   - Allocate $10k to Climate Simulation project
4. Create budget "AWS Research Credits" ($5,000)
   - Allocate $5k to Climate Simulation project
5. **Launch workspace with funding selection**:
   - Template: "GPU Compute Cluster"
   - Size: p3.8xlarge (GPU intensive)
   - Select funding: **NSF Grant** (GPU work funded by grant)
6. **Launch storage with different funding**:
   - Size: 500GB EFS
   - Select funding: **Department Matching** (storage from matching funds)
7. Monitor spending per funding source:
   - NSF Grant: $32k spent (64% - GPU workspaces)
   - Department: $8k spent (80% - storage and support)
   - AWS Credits: $2k spent (40% - development/testing)
8. **Exhaustion scenario**: NSF Grant reaches $50k
   - Backup funding: Department Emergency (if configured)
   - Otherwise: Hibernate GPU workspaces, notify PI
9. Generate multi-source report for grant closeout

**Key Features Tested**:
- Multiple budget allocations to single project
- Funding source selection at launch
- Per-allocation spending tracking
- Budget exhaustion with backup funding
- Cross-budget project reporting

---

## 📚 Documentation Updates

### New Documentation
- [ ] Multi-project budget guide
- [ ] Multi-source funding guide (N budgets → 1 project)
- [ ] Budget allocation best practices
- [ ] Funding source selection at launch
- [ ] Budget cushion and backup funding setup
- [ ] Grant reporting workflow
- [ ] Budget reallocation tutorial

### Updated Documentation
- [ ] Budget management section
- [ ] Project budget tab documentation
- [ ] API reference (new endpoints)
- [ ] Migration guide (v0.5.9 → v0.5.10)

### Release Notes
- [ ] Breaking changes (budget schema)
- [ ] New features (multi-project budgets)
- [ ] Migration instructions
- [ ] API changes

---

## 🚀 Release Criteria

### Must Have (Blocking)
- ✅ Shared budget pools implemented
- ✅ Project allocation system working
- ✅ Budget reallocation functional
- ✅ Cost aggregation accurate
- ✅ Migration script tested
- ✅ All persona tests pass
- ✅ Documentation complete

### Nice to Have (Non-Blocking)
- Budget templates (common allocations)
- Budget forecasting (burn rate projection)
- Advanced reporting (custom date ranges)
- Budget approval workflow (for institutions)

---

## 📊 Success Metrics (Post-Release)

Track for 2 weeks after release:

1. **Multi-Project Adoption**
   - Measure: % of budgets with 2+ projects
   - Target: >60% of new budgets

2. **Reallocation Usage**
   - Measure: Number of reallocations per budget
   - Target: Average 2-3 reallocations per grant-period budget

3. **Budget Utilization**
   - Measure: % of allocated funds actually spent
   - Target: >85% utilization (reduced waste)

4. **Time to Allocate**
   - Measure: Time from budget creation to full allocation
   - Target: <5 minutes for 10 projects

5. **Support Tickets**
   - Measure: Budget-related confusion tickets
   - Target: <5% of total support volume

---

## 🔗 Related Documents

- ROADMAP.md - Overall project roadmap
- RELEASE_PLAN_v0.5.9.md - Navigation Restructure (prerequisite)
- RELEASE_PLAN_v0.5.11.md - User Invitation & Roles (follows this)
- User Guide: Budget Management (to be updated)

---

**Last Updated**: October 27, 2025
**Status**: 📋 Planned
**Dependencies**: v0.5.9 (Navigation Restructure)
