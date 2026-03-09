package project

import (
	"fmt"
	"sync"
	"time"
)

// GDEWStatus represents the current GDEW credit status for a project.
type GDEWStatus struct {
	ProjectID        string    `json:"project_id"`
	MonthStart       time.Time `json:"month_start"`
	MonthEnd         time.Time `json:"month_end"`
	TotalSpendMTD    float64   `json:"total_spend_mtd"`
	EgressChargesMTD float64   `json:"egress_charges_mtd"`
	AvailableCredit  float64   `json:"available_credit"` // 15% of total spend
	UsedCredit       float64   `json:"used_credit"`      // min(egress, available)
	RemainingCredit  float64   `json:"remaining_credit"`
	NetEgressCost    float64   `json:"net_egress_cost"`
	CoveragePercent  float64   `json:"coverage_percent"` // used/egress * 100
	FullyCovered     bool      `json:"fully_covered"`
	NearingLimit     bool      `json:"nearing_limit"` // remaining < $20
	StatusMessage    string    `json:"status_message"`
	LastCalculated   time.Time `json:"last_calculated"`
}

// GDEWTracker computes AWS Global Data Egress Waiver credits.
// GDEW provides credits up to 15% of total monthly AWS spend against egress charges.
type GDEWTracker struct {
	mu   sync.RWMutex
	data map[string]*GDEWStatus // projectID → status
}

// NewGDEWTracker creates a new GDEWTracker.
func NewGDEWTracker() *GDEWTracker { return &GDEWTracker{data: make(map[string]*GDEWStatus)} }

// Update recalculates GDEW status for a project given current month spend and egress.
func (g *GDEWTracker) Update(projectID string, totalSpendMTD, egressChargesMTD float64) *GDEWStatus {
	g.mu.Lock()
	defer g.mu.Unlock()

	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	monthEnd := monthStart.AddDate(0, 1, 0).Add(-time.Nanosecond)

	availableCredit := totalSpendMTD * 0.15
	usedCredit := egressChargesMTD
	if usedCredit > availableCredit {
		usedCredit = availableCredit
	}
	remainingCredit := availableCredit - usedCredit
	netEgressCost := egressChargesMTD - usedCredit

	coveragePct := 0.0
	if egressChargesMTD > 0 {
		coveragePct = usedCredit / egressChargesMTD * 100
	}

	fullyCovered := netEgressCost <= 0
	nearingLimit := remainingCredit < 20.0 && !fullyCovered

	msg := fmt.Sprintf("✅ FULLY COVERED — $%.2f GDEW credit available, $%.2f remaining", availableCredit, remainingCredit)
	if !fullyCovered {
		msg = fmt.Sprintf("⚠️  PARTIAL — $%.2f net egress cost (%.0f%% covered by GDEW)", netEgressCost, coveragePct)
	}
	if nearingLimit {
		msg = fmt.Sprintf("🚨 NEARING LIMIT — only $%.2f GDEW credit remaining this month", remainingCredit)
	}

	status := &GDEWStatus{
		ProjectID: projectID, MonthStart: monthStart, MonthEnd: monthEnd,
		TotalSpendMTD: totalSpendMTD, EgressChargesMTD: egressChargesMTD,
		AvailableCredit: availableCredit, UsedCredit: usedCredit,
		RemainingCredit: remainingCredit, NetEgressCost: netEgressCost,
		CoveragePercent: coveragePct, FullyCovered: fullyCovered,
		NearingLimit: nearingLimit, StatusMessage: msg, LastCalculated: now,
	}
	g.data[projectID] = status
	return status
}

// Get returns the current GDEW status for a project, or nil if not tracked.
func (g *GDEWTracker) Get(projectID string) *GDEWStatus {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.data[projectID]
}
