package models

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/scttfrdmn/prism/internal/tui/api"
	"github.com/scttfrdmn/prism/internal/tui/components"
	"github.com/scttfrdmn/prism/internal/tui/styles"
)

// BudgetModel represents the budget management view
type BudgetModel struct {
	apiClient         apiClient
	budgetsTable      components.Table
	statusBar         components.StatusBar
	spinner           components.Spinner
	width             int
	height            int
	loading           bool
	error             string
	projects          []api.ProjectResponse
	selectedTab       int // 0=list, 1=breakdown, 2=forecast, 3=savings
	selectedBudget    int
	showCreateDialog  bool
	createProjectName string
	createAmount      string
}

// BudgetDataMsg represents budget data retrieved from the API
type BudgetDataMsg struct {
	Projects []api.ProjectResponse
	Error    error
}

// NewBudgetModel creates a new budget model
func NewBudgetModel(apiClient apiClient) BudgetModel {
	// Create budgets table
	columns := []table.Column{
		{Title: "PROJECT", Width: 20},
		{Title: "BUDGET", Width: 12},
		{Title: "SPENT", Width: 12},
		{Title: "REMAINING", Width: 12},
		{Title: "%USED", Width: 8},
		{Title: "STATUS", Width: 10},
		{Title: "ALERTS", Width: 8},
	}

	budgetsTable := components.NewTable(columns, []table.Row{}, 80, 10, true)

	// Create status bar and spinner
	statusBar := components.NewStatusBar("Prism Budget Management", "")
	spinner := components.NewSpinner("Loading budgets...")

	return BudgetModel{
		apiClient:      apiClient,
		budgetsTable:   budgetsTable,
		statusBar:      statusBar,
		spinner:        spinner,
		width:          80,
		height:         24,
		loading:        true,
		selectedTab:    0,
		selectedBudget: 0,
	}
}

// Init initializes the model
func (m BudgetModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.InitialCmd(),
		m.fetchBudgets,
	)
}

// fetchBudgets retrieves budget data from the API
func (m BudgetModel) fetchBudgets() tea.Msg {
	// Fetch all projects with budget information
	resp, err := m.apiClient.ListProjects(context.Background(), nil)
	if err != nil {
		return BudgetDataMsg{Error: fmt.Errorf("failed to list projects: %w", err)}
	}

	return BudgetDataMsg{
		Projects: resp.Projects,
		Error:    nil,
	}
}

// Update handles messages and updates the model
func (m BudgetModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.budgetsTable.SetSize(msg.Width-4, msg.Height-12)
		return m, nil

	case BudgetDataMsg:
		if msg.Error != nil {
			m.error = msg.Error.Error()
			m.loading = false
			return m, nil
		}
		m.projects = msg.Projects
		m.loading = false
		m.error = ""
		m.updateBudgetsTable()
		return m, nil

	case tea.KeyMsg:
		return updateBudgetOnKey(m, msg)

	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

// updateBudgetOnKey handles keyboard input for the budget view.
func updateBudgetOnKey(m BudgetModel, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.loading {
		return m, nil
	}
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "r", "f5":
		m.loading = true
		return m, m.fetchBudgets
	case "tab":
		m.selectedTab = (m.selectedTab + 1) % 4
		return m, nil
	case "n":
		if m.selectedTab == 0 {
			m.statusBar.SetStatus("Set budget via CLI: prism project budget set <project> <amount>", components.StatusSuccess)
			return m, nil
		}
	case "esc":
		return handleBudgetDialogKey(m, "esc")
	case "enter":
		return handleBudgetDialogKey(m, "enter")
	case "up", "k":
		return handleBudgetCursorKey(m, -1)
	case "down", "j":
		return handleBudgetCursorKey(m, 1)
	}
	return m, nil
}

// handleBudgetDialogKey handles "esc" and "enter" for the create-budget dialog.
func handleBudgetDialogKey(m BudgetModel, key string) (BudgetModel, tea.Cmd) {
	switch key {
	case "esc":
		if m.showCreateDialog {
			m.showCreateDialog = false
			m.createProjectName = ""
			m.createAmount = ""
		}
	case "enter":
		if m.showCreateDialog {
			return m, m.createBudget
		}
	}
	return m, nil
}

// handleBudgetCursorKey moves the budget selection up (dir<0) or down (dir>0).
func handleBudgetCursorKey(m BudgetModel, dir int) (BudgetModel, tea.Cmd) {
	if dir < 0 && m.selectedBudget > 0 {
		m.selectedBudget--
	}
	if dir > 0 && m.selectedBudget < len(m.projects)-1 {
		m.selectedBudget++
	}
	return m, nil
}

// View renders the model
func (m BudgetModel) View() string {
	if m.loading {
		return m.spinner.View()
	}

	var b strings.Builder
	theme := styles.CurrentTheme

	// Header
	header := theme.Title.Render("💰 Budget Management")
	b.WriteString(header)
	b.WriteString("\n\n")

	// Tab bar
	tabs := []string{"Overview", "Breakdown", "Forecast", "Savings"}
	tabBar := renderTabBar(tabs, m.selectedTab, theme)
	b.WriteString(tabBar)
	b.WriteString("\n\n")

	// Content based on selected tab
	switch m.selectedTab {
	case 0: // Overview
		b.WriteString(m.renderOverview())
	case 1: // Breakdown
		b.WriteString(m.renderBreakdown())
	case 2: // Forecast
		b.WriteString(m.renderForecast())
	case 3: // Savings
		b.WriteString(m.renderSavings())
	}

	// Show create dialog if active
	if m.showCreateDialog {
		dialog := m.renderCreateDialog()
		b.WriteString("\n\n")
		b.WriteString(dialog)
	}

	// Error display
	if m.error != "" {
		b.WriteString("\n\n")
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
		b.WriteString(errorStyle.Render("Error: " + m.error))
	}

	// Help text
	b.WriteString("\n\n")
	helpText := m.renderHelp()
	b.WriteString(helpText)

	return b.String()
}

// renderOverview displays the budget overview list
func (m BudgetModel) renderOverview() string {
	if len(m.projects) == 0 {
		return "No projects with budgets found.\n\nPress 'n' to create a budget for a project."
	}

	var b strings.Builder
	theme := styles.CurrentTheme

	// Summary statistics
	totalBudget := 0.0
	totalSpent := 0.0
	budgetCount := 0

	for _, proj := range m.projects {
		if proj.BudgetStatus != nil && proj.BudgetStatus.TotalBudget > 0 {
			budgetCount++
			totalBudget += proj.BudgetStatus.TotalBudget
			totalSpent += proj.BudgetStatus.SpentAmount
		}
	}

	if budgetCount > 0 {
		spentPercent := (totalSpent / totalBudget) * 100
		summary := fmt.Sprintf("Active Budgets: %d | Total Budget: $%.2f | Total Spent: $%.2f (%.1f%%) | Remaining: $%.2f",
			budgetCount, totalBudget, totalSpent, spentPercent, totalBudget-totalSpent)
		b.WriteString(theme.SubTitle.Render(summary))
		b.WriteString("\n\n")
	}

	// Budget table
	rows := []table.Row{}
	for i, proj := range m.projects {
		if proj.BudgetStatus == nil || proj.BudgetStatus.TotalBudget <= 0 {
			// No budget configured
			row := table.Row{
				proj.Name,
				"-",
				fmt.Sprintf("$%.2f", proj.TotalCost),
				"-",
				"-",
				"No Budget",
				"-",
			}
			rows = append(rows, row)
			continue
		}

		budget := proj.BudgetStatus
		remaining := budget.TotalBudget - budget.SpentAmount
		if remaining < 0 {
			remaining = 0
		}
		usedPercent := (budget.SpentAmount / budget.TotalBudget) * 100

		// Status indicator
		status := "OK"
		if usedPercent >= 95 {
			status = "CRITICAL"
		} else if usedPercent >= 80 {
			status = "WARNING"
		}

		// Alert count
		alertStatus := "-"
		if len(budget.ActiveAlerts) > 0 {
			alertStatus = fmt.Sprintf("%d", len(budget.ActiveAlerts))
		}

		// Selection indicator
		projectName := proj.Name
		if i == m.selectedBudget {
			projectName = "> " + projectName
		}

		row := table.Row{
			projectName,
			fmt.Sprintf("$%.2f", budget.TotalBudget),
			fmt.Sprintf("$%.2f", budget.SpentAmount),
			fmt.Sprintf("$%.2f", remaining),
			fmt.Sprintf("%.1f%%", usedPercent),
			status,
			alertStatus,
		}
		rows = append(rows, row)
	}

	// Update table rows
	m.budgetsTable.SetRows(rows)
	b.WriteString(m.budgetsTable.View())

	return b.String()
}

// truncate shortens s to at most n runes, appending "…" if truncated.
func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	if n <= 1 {
		return "…"
	}
	return string(runes[:n-1]) + "…"
}

// renderASCIIBar renders a filled/empty bar of given width for the given ratio (0.0–1.0).
func renderASCIIBar(ratio float64, width int) string {
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}
	filled := int(ratio * float64(width))
	return strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
}

// renderBreakdown displays cost breakdown with ASCII bar charts (issue #39)
func (m BudgetModel) renderBreakdown() string {
	if len(m.projects) == 0 {
		return "No projects with budgets found."
	}

	var b strings.Builder
	theme := styles.CurrentTheme
	barWidth := 30

	b.WriteString(theme.SubTitle.Render("Cost Breakdown — Budget Utilization by Project"))
	b.WriteString("\n\n")

	// Per-project budget utilization bar chart
	maxBudget := 0.0
	for _, proj := range m.projects {
		if proj.BudgetStatus != nil && proj.BudgetStatus.TotalBudget > maxBudget {
			maxBudget = proj.BudgetStatus.TotalBudget
		}
	}
	if maxBudget <= 0 {
		maxBudget = 1
	}

	for i, proj := range m.projects {
		selector := "  "
		if i == m.selectedBudget {
			selector = "> "
		}
		if proj.BudgetStatus == nil || proj.BudgetStatus.TotalBudget <= 0 {
			b.WriteString(fmt.Sprintf("%s%-20s  [no budget]\n", selector, truncate(proj.Name, 20)))
			continue
		}

		spent := proj.BudgetStatus.SpentAmount
		budget := proj.BudgetStatus.TotalBudget
		ratio := spent / budget

		// Color-code the bar label based on usage
		statusMark := "  "
		if ratio >= 0.95 {
			statusMark = "!!"
		} else if ratio >= 0.80 {
			statusMark = " !"
		}

		bar := renderASCIIBar(ratio, barWidth)
		b.WriteString(fmt.Sprintf("%s%-20s  [%s]%s  $%.0f/$%.0f (%.0f%%)\n",
			selector, truncate(proj.Name, 20), bar, statusMark,
			spent, budget, ratio*100))
	}

	// Selected project detail
	if m.selectedBudget < len(m.projects) {
		proj := m.projects[m.selectedBudget]
		if proj.BudgetStatus != nil && proj.BudgetStatus.TotalBudget > 0 {
			b.WriteString("\n")
			b.WriteString(theme.SubTitle.Render(fmt.Sprintf("Detail — %s", proj.Name)))
			b.WriteString("\n")

			budget := proj.BudgetStatus
			spent := budget.SpentAmount
			remaining := budget.TotalBudget - spent
			if remaining < 0 {
				remaining = 0
			}

			// Spent vs. remaining bars
			b.WriteString(fmt.Sprintf("  Spent     [%s]  $%.2f\n",
				renderASCIIBar(spent/budget.TotalBudget, barWidth), spent))
			b.WriteString(fmt.Sprintf("  Remaining [%s]  $%.2f\n",
				renderASCIIBar(remaining/budget.TotalBudget, barWidth), remaining))

			if budget.ProjectedMonthlySpend > 0 {
				b.WriteString(fmt.Sprintf("\n  Projected Monthly: $%.2f", budget.ProjectedMonthlySpend))
				if budget.DaysUntilBudgetExhausted != nil && *budget.DaysUntilBudgetExhausted > 0 {
					b.WriteString(fmt.Sprintf("  |  Budget lasts: %d days", *budget.DaysUntilBudgetExhausted))
				}
				b.WriteString("\n")
			}
		}
	}

	b.WriteString("\n  !! = critical (≥95%)  ! = warning (≥80%)")
	b.WriteString("\n  ↑/↓ select project  |  prism budget breakdown <project> (service-level detail)\n")
	return b.String()
}

// renderForecast displays spending forecast
func (m BudgetModel) renderForecast() string {
	if m.selectedBudget >= len(m.projects) {
		return "Select a project to view spending forecast."
	}

	project := m.projects[m.selectedBudget]
	var b strings.Builder

	b.WriteString(fmt.Sprintf("Spending Forecast for '%s'\n\n", project.Name))

	if project.BudgetStatus == nil || project.BudgetStatus.TotalBudget <= 0 {
		b.WriteString("No budget configured for this project.\n")
		return b.String()
	}

	budget := project.BudgetStatus

	b.WriteString(fmt.Sprintf("Current Spending: $%.2f (%.1f%%)\n", budget.SpentAmount, (budget.SpentAmount/budget.TotalBudget)*100))

	if budget.ProjectedMonthlySpend > 0 {
		b.WriteString(fmt.Sprintf("Projected Monthly: $%.2f\n", budget.ProjectedMonthlySpend))

		if budget.DaysUntilBudgetExhausted != nil {
			days := *budget.DaysUntilBudgetExhausted
			if days > 0 {
				b.WriteString(fmt.Sprintf("Budget Exhaustion: %d days\n", days))
			}
		}
	}

	b.WriteString("\n💡 Detailed forecasting available via: prism budget forecast " + project.Name + "\n")

	return b.String()
}

// renderSavings displays hibernation savings analysis with visualization (issue #40)
func (m BudgetModel) renderSavings() string {
	var b strings.Builder
	theme := styles.CurrentTheme
	barWidth := 24

	b.WriteString(theme.SubTitle.Render("Hibernation Savings Analysis"))
	b.WriteString("\n\n")
	b.WriteString("  Estimate: hibernating workspaces ~12h/day saves ~50% of compute costs.\n\n")

	type savingsRow struct {
		name            string
		currentMonthly  float64
		savingsEstimate float64
	}

	var rows []savingsRow
	totalCurrent := 0.0
	totalSavings := 0.0
	maxMonthly := 0.0

	for _, proj := range m.projects {
		var monthly float64
		if proj.BudgetStatus != nil && proj.BudgetStatus.ProjectedMonthlySpend > 0 {
			monthly = proj.BudgetStatus.ProjectedMonthlySpend
		} else {
			monthly = proj.TotalCost
		}
		// Estimate: ~80% of spend is compute; hibernating half the time saves ~40% total
		savingsEst := monthly * 0.40
		rows = append(rows, savingsRow{proj.Name, monthly, savingsEst})
		totalCurrent += monthly
		totalSavings += savingsEst
		if monthly > maxMonthly {
			maxMonthly = monthly
		}
	}

	if maxMonthly <= 0 {
		maxMonthly = 1
	}

	if len(rows) == 0 {
		b.WriteString("  No projects found. Launch a workspace to see savings potential.\n")
		return b.String()
	}

	// Header
	b.WriteString(fmt.Sprintf("  %-20s  %-26s  %s\n", "PROJECT", "SAVINGS POTENTIAL", "EST. $/MO SAVED"))
	b.WriteString(fmt.Sprintf("  %-20s  %-26s  %s\n",
		strings.Repeat("─", 20), strings.Repeat("─", 26), strings.Repeat("─", 14)))

	for _, row := range rows {
		ratio := 0.0
		if row.currentMonthly > 0 {
			ratio = row.savingsEstimate / row.currentMonthly
		}
		bar := renderASCIIBar(ratio, barWidth)
		b.WriteString(fmt.Sprintf("  %-20s  [%s]  $%.2f\n",
			truncate(row.name, 20), bar, row.savingsEstimate))
	}

	// Totals
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  Current Monthly Spend (est):  $%.2f\n", totalCurrent))
	b.WriteString(fmt.Sprintf("  Potential Savings (est):       $%.2f/mo\n", totalSavings))
	if totalCurrent > 0 {
		b.WriteString(fmt.Sprintf("  Savings Percentage:            %.0f%%\n", (totalSavings/totalCurrent)*100))
	}

	// Savings bar showing proportion saved
	if totalCurrent > 0 {
		b.WriteString("\n")
		savingsRatio := totalSavings / totalCurrent
		b.WriteString(fmt.Sprintf("  Savings vs. current  [%s]  %.0f%%\n",
			renderASCIIBar(savingsRatio, barWidth), savingsRatio*100))
	}

	b.WriteString("\n  Based on 12h/day hibernation schedule (compute-heavy workloads).")
	b.WriteString("\n  prism workspace hibernate <name>  to hibernate a workspace\n")
	return b.String()
}

// renderCreateDialog displays the budget creation dialog
func (m BudgetModel) renderCreateDialog() string {
	theme := styles.CurrentTheme

	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2).
		Width(60)

	var content strings.Builder
	content.WriteString(theme.SubTitle.Render("Create Budget") + "\n\n")
	content.WriteString("Project Name: " + m.createProjectName + "\n")
	content.WriteString("Budget Amount: $" + m.createAmount + "\n\n")
	content.WriteString("Press Enter to create, Esc to cancel\n")

	return dialogStyle.Render(content.String())
}

// renderHelp displays help text
func (m BudgetModel) renderHelp() string {
	theme := styles.CurrentTheme

	var helps []string
	if m.showCreateDialog {
		helps = []string{"esc: cancel", "enter: create"}
	} else {
		helps = []string{
			"tab: switch tabs",
			"↑/↓: select",
			"n: new budget",
			"r: refresh",
			"q: quit",
		}
	}

	return theme.Help.Render(strings.Join(helps, " • "))
}

// createBudget creates a new budget via the API
func (m BudgetModel) createBudget() tea.Msg {
	// Design Decision: Budget creation requires CLI for comprehensive configuration
	// Rationale: Budgets have many optional parameters (alerts, actions, limits, etc.)
	// TUI form input would be complex; CLI provides better UX for advanced configuration
	// Use CLI command: prism budget create <project> <amount> [--alert ...] [--action ...]
	return BudgetDataMsg{Error: fmt.Errorf("budget creation via TUI not implemented - use CLI: prism budget create <project> <amount>")}
}

// renderTabBar renders a tab bar for navigation
func renderTabBar(tabs []string, selected int, theme styles.Theme) string {
	var b strings.Builder

	for i, tab := range tabs {
		if i == selected {
			b.WriteString(theme.Tab.Active.Render("[" + tab + "]"))
		} else {
			b.WriteString(theme.Tab.Inactive.Render(" " + tab + " "))
		}
		if i < len(tabs)-1 {
			b.WriteString(" ")
		}
	}

	return b.String()
}

// updateBudgetsTable updates the budgets table with current data
func (m *BudgetModel) updateBudgetsTable() {
	// This method updates the table rows with current project/budget data
	// The actual update happens in renderOverview()
}
