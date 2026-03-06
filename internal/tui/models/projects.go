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

// ProjectsModel represents the project management view
type ProjectsModel struct {
	apiClient         apiClient
	projectsTable     components.Table
	membersTable      components.Table
	statusBar         components.StatusBar
	spinner           components.Spinner
	width             int
	height            int
	loading           bool
	error             string
	projects          []api.ProjectResponse
	members           []api.MemberResponse
	membersLoading    bool
	selectedProject   int
	selectedTab       int // 0=list, 1=members, 2=instances, 3=budget
	showCreateDialog  bool
	createName        string
	createDescription string
}

// ProjectDataMsg represents project data retrieved from the API
type ProjectDataMsg struct {
	Projects []api.ProjectResponse
	Error    error
}

// ProjectMembersMsg represents member data retrieved from the API
type ProjectMembersMsg struct {
	Members   []api.MemberResponse
	ProjectID string
	Error     error
}

// NewProjectsModel creates a new projects model
func NewProjectsModel(apiClient apiClient) ProjectsModel {
	// Create projects table
	columns := []table.Column{
		{Title: "NAME", Width: 20},
		{Title: "OWNER", Width: 15},
		{Title: "STATUS", Width: 10},
		{Title: "MEMBERS", Width: 8},
		{Title: "INSTANCES", Width: 10},
		{Title: "COST", Width: 12},
		{Title: "BUDGET", Width: 12},
	}
	projectsTable := components.NewTable(columns, []table.Row{}, 80, 10, true)

	// Create members table
	memberColumns := []table.Column{
		{Title: "USER", Width: 30},
		{Title: "ROLE", Width: 12},
		{Title: "ADDED BY", Width: 20},
		{Title: "ADDED", Width: 12},
	}
	membersTable := components.NewTable(memberColumns, []table.Row{}, 80, 10, false)

	// Create status bar and spinner
	statusBar := components.NewStatusBar("Prism Project Management", "")
	spinner := components.NewSpinner("Loading projects...")

	return ProjectsModel{
		apiClient:       apiClient,
		projectsTable:   projectsTable,
		membersTable:    membersTable,
		statusBar:       statusBar,
		spinner:         spinner,
		width:           80,
		height:          24,
		loading:         true,
		selectedTab:     0,
		selectedProject: 0,
	}
}

// Init initializes the model
func (m ProjectsModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.InitialCmd(),
		m.fetchProjects,
	)
}

// fetchProjects retrieves project data from the API
func (m ProjectsModel) fetchProjects() tea.Msg {
	resp, err := m.apiClient.ListProjects(context.Background(), nil)
	if err != nil {
		return ProjectDataMsg{Error: fmt.Errorf("failed to list projects: %w", err)}
	}

	return ProjectDataMsg{
		Projects: resp.Projects,
		Error:    nil,
	}
}

// fetchProjectMembers retrieves member data for the selected project
func (m ProjectsModel) fetchProjectMembers() tea.Msg {
	if m.selectedProject >= len(m.projects) {
		return ProjectMembersMsg{}
	}
	project := m.projects[m.selectedProject]
	resp, err := m.apiClient.GetProjectMembers(context.Background(), project.ID)
	if err != nil {
		return ProjectMembersMsg{Error: fmt.Errorf("failed to get members: %w", err)}
	}
	return ProjectMembersMsg{Members: resp.Members, ProjectID: project.ID}
}

// Update handles messages and updates the model
func (m ProjectsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.projectsTable.SetSize(msg.Width-4, msg.Height-12)
		m.membersTable.SetSize(msg.Width-4, msg.Height-14)
		return m, nil

	case ProjectMembersMsg:
		m.membersLoading = false
		if msg.Error != nil {
			m.error = msg.Error.Error()
		} else {
			m.members = msg.Members
		}
		return m, nil

	case ProjectDataMsg:
		if msg.Error != nil {
			m.error = msg.Error.Error()
			m.loading = false
			return m, nil
		}
		m.projects = msg.Projects
		m.loading = false
		m.error = ""
		m.updateProjectsTable()
		return m, nil

	case tea.KeyMsg:
		return updateProjectsOnKey(m, msg)

	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

// updateProjectsOnKey handles keyboard input for the projects view.
func updateProjectsOnKey(m ProjectsModel, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.loading {
		return m, nil
	}
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "r", "f5":
		m.loading = true
		return m, m.fetchProjects
	case "tab":
		m.selectedTab = (m.selectedTab + 1) % 4
		if m.selectedTab == 1 && len(m.projects) > 0 {
			m.membersLoading = true
			m.members = nil
			return m, m.fetchProjectMembers
		}
		return m, nil
	case "n":
		if m.selectedTab == 0 {
			m.showCreateDialog = true
			return m, nil
		}
	case "esc":
		return handleProjectsDialogKey(m, "esc")
	case "enter":
		return handleProjectsDialogKey(m, "enter")
	case "up", "k":
		return handleProjectsCursorKey(m, -1)
	case "down", "j":
		return handleProjectsCursorKey(m, 1)
	}
	return m, nil
}

// handleProjectsDialogKey handles "esc" and "enter" for the create-project dialog.
func handleProjectsDialogKey(m ProjectsModel, key string) (ProjectsModel, tea.Cmd) {
	switch key {
	case "esc":
		if m.showCreateDialog {
			m.showCreateDialog = false
			m.createName = ""
			m.createDescription = ""
		}
	case "enter":
		if m.showCreateDialog {
			return m, m.createProject
		}
	}
	return m, nil
}

// handleProjectsCursorKey moves the project selection up (dir<0) or down (dir>0).
func handleProjectsCursorKey(m ProjectsModel, dir int) (ProjectsModel, tea.Cmd) {
	if dir < 0 && m.selectedProject > 0 {
		m.selectedProject--
	}
	if dir > 0 && m.selectedProject < len(m.projects)-1 {
		m.selectedProject++
	}
	return m, nil
}

// View renders the model
func (m ProjectsModel) View() string {
	if m.loading {
		return m.spinner.View()
	}

	var b strings.Builder
	theme := styles.CurrentTheme

	// Header
	header := theme.Title.Render("📁 Project Management")
	b.WriteString(header)
	b.WriteString("\n\n")

	// Tab bar
	tabs := []string{"Overview", "Members", "Instances", "Budget"}
	tabBar := renderProjectTabBar(tabs, m.selectedTab, theme)
	b.WriteString(tabBar)
	b.WriteString("\n\n")

	// Content based on selected tab
	switch m.selectedTab {
	case 0: // Overview
		b.WriteString(m.renderOverview())
	case 1: // Members
		b.WriteString(m.renderMembers())
	case 2: // Instances
		b.WriteString(m.renderInstances())
	case 3: // Budget
		b.WriteString(m.renderBudget())
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

// renderOverview displays the project overview list
func (m ProjectsModel) renderOverview() string {
	if len(m.projects) == 0 {
		return "No projects found.\n\nPress 'n' to create a new project."
	}

	var b strings.Builder
	theme := styles.CurrentTheme

	// Summary statistics
	activeCount := 0
	totalCost := 0.0
	totalBudget := 0.0

	for _, proj := range m.projects {
		if proj.Status == "active" {
			activeCount++
		}
		totalCost += proj.TotalCost
		if proj.BudgetStatus != nil {
			totalBudget += proj.BudgetStatus.TotalBudget
		}
	}

	summary := fmt.Sprintf("Total Projects: %d | Active: %d | Total Cost: $%.2f | Total Budget: $%.2f",
		len(m.projects), activeCount, totalCost, totalBudget)
	b.WriteString(theme.SubTitle.Render(summary))
	b.WriteString("\n\n")

	// Projects table
	rows := []table.Row{}
	for i, proj := range m.projects {
		// Budget status
		budgetStr := "-"
		if proj.BudgetStatus != nil {
			budgetStr = fmt.Sprintf("$%.2f", proj.BudgetStatus.TotalBudget)
		}

		// Selection indicator
		projectName := proj.Name
		if i == m.selectedProject {
			projectName = "> " + projectName
		}

		row := table.Row{
			projectName,
			proj.Owner,
			proj.Status,
			fmt.Sprintf("%d", proj.MemberCount),
			fmt.Sprintf("%d", proj.ActiveInstances),
			fmt.Sprintf("$%.2f", proj.TotalCost),
			budgetStr,
		}
		rows = append(rows, row)
	}

	// Update table rows
	m.projectsTable.SetRows(rows)
	b.WriteString(m.projectsTable.View())

	return b.String()
}

// renderMembers displays project members in a paginated table
func (m ProjectsModel) renderMembers() string {
	if len(m.projects) == 0 {
		return "No projects found. Select a project in the Overview tab first."
	}
	if m.selectedProject >= len(m.projects) {
		return "Select a project to view members."
	}

	project := m.projects[m.selectedProject]
	var b strings.Builder
	theme := styles.CurrentTheme

	b.WriteString(theme.SubTitle.Render(fmt.Sprintf("Members — %s", project.Name)))
	b.WriteString(fmt.Sprintf("\nOwner: %s  |  Total Members: %d\n\n", project.Owner, project.MemberCount))

	if m.membersLoading {
		return b.String() + "  Loading members..."
	}

	if len(m.members) == 0 {
		b.WriteString("  No members found.\n\n")
		b.WriteString("  Add members:    prism project members add --project " + project.Name + " <email> <role>\n")
		b.WriteString("  Remove members: prism project members remove --project " + project.Name + " <email>\n")
		return b.String()
	}

	rows := []table.Row{}
	for _, member := range m.members {
		addedAt := "-"
		if !member.AddedAt.IsZero() {
			addedAt = member.AddedAt.Format("2006-01-02")
		}
		addedBy := member.AddedBy
		if addedBy == "" {
			addedBy = "-"
		}
		rows = append(rows, table.Row{member.UserID, member.Role, addedBy, addedAt})
	}
	m.membersTable.SetRows(rows)
	b.WriteString(m.membersTable.View())
	b.WriteString("\n\n  r: refresh  |  prism project members add --project " + project.Name + " <email> <role>")

	return b.String()
}

// renderInstances displays project instances
func (m ProjectsModel) renderInstances() string {
	if m.selectedProject >= len(m.projects) {
		return "Select a project to view instances."
	}

	project := m.projects[m.selectedProject]
	var b strings.Builder

	b.WriteString(fmt.Sprintf("Instances for project '%s'\n\n", project.Name))
	b.WriteString(fmt.Sprintf("Active Instances: %d\n", project.ActiveInstances))
	b.WriteString(fmt.Sprintf("Total Cost: $%.2f\n\n", project.TotalCost))

	// Design Decision: TUI shows instance summary; detailed instance list requires CLI/Instance view
	// Rationale: Instance details are available in main Instances view (tab 3)
	// Project-filtered instance list would duplicate existing TUI functionality
	b.WriteString("💡 View project instances: prism project instances " + project.Name + "\n")

	return b.String()
}

// renderBudget displays project budget information
func (m ProjectsModel) renderBudget() string {
	if m.selectedProject >= len(m.projects) {
		return "Select a project to view budget."
	}

	project := m.projects[m.selectedProject]
	var b strings.Builder

	b.WriteString(fmt.Sprintf("Budget for project '%s'\n\n", project.Name))

	if project.BudgetStatus == nil || project.BudgetStatus.TotalBudget <= 0 {
		b.WriteString("No budget configured for this project.\n\n")
		b.WriteString("💡 Set budget: prism project budget set " + project.Name + " <amount>\n")
		return b.String()
	}

	budget := project.BudgetStatus
	remaining := budget.TotalBudget - budget.SpentAmount
	if remaining < 0 {
		remaining = 0
	}

	b.WriteString(fmt.Sprintf("Total Budget: $%.2f\n", budget.TotalBudget))
	b.WriteString(fmt.Sprintf("Spent: $%.2f (%.1f%%)\n", budget.SpentAmount, budget.SpentPercentage))
	b.WriteString(fmt.Sprintf("Remaining: $%.2f\n\n", remaining))

	// Alert status
	if len(budget.ActiveAlerts) > 0 {
		b.WriteString(fmt.Sprintf("⚠️  Active Alerts: %d\n", len(budget.ActiveAlerts)))
		for _, alert := range budget.ActiveAlerts {
			b.WriteString(fmt.Sprintf("  • %s\n", alert))
		}
		b.WriteString("\n")
	}

	// Projected spending
	if budget.ProjectedMonthlySpend > 0 {
		b.WriteString(fmt.Sprintf("Projected Monthly: $%.2f\n", budget.ProjectedMonthlySpend))
	}

	b.WriteString("\n💡 Budget management: prism project budget status " + project.Name + "\n")

	return b.String()
}

// renderCreateDialog displays the project creation dialog
func (m ProjectsModel) renderCreateDialog() string {
	theme := styles.CurrentTheme

	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2).
		Width(60)

	var content strings.Builder
	content.WriteString(theme.SubTitle.Render("Create Project") + "\n\n")
	content.WriteString("Project Name: " + m.createName + "\n")
	content.WriteString("Description: " + m.createDescription + "\n\n")
	content.WriteString("Press Enter to create, Esc to cancel\n")

	return dialogStyle.Render(content.String())
}

// renderHelp displays help text
func (m ProjectsModel) renderHelp() string {
	theme := styles.CurrentTheme

	var helps []string
	if m.showCreateDialog {
		helps = []string{"esc: cancel", "enter: create"}
	} else {
		helps = []string{
			"tab: switch tabs",
			"↑/↓: select",
			"n: new project",
			"r: refresh",
			"q: quit",
		}
	}

	return theme.Help.Render(strings.Join(helps, " • "))
}

// createProject creates a new project via the API
func (m ProjectsModel) createProject() tea.Msg {
	// Design Decision: Project creation requires CLI for proper input validation
	// Rationale: TUI input forms are complex; CLI provides better error handling and validation
	// Future Enhancement: Add TUI form dialog if demand exists
	// Use CLI command: prism project create <name> --owner <email> [--description "..."]
	return ProjectDataMsg{Error: fmt.Errorf("project creation via TUI not implemented - use CLI: prism project create <name> --owner <email>")}
}

// renderProjectTabBar renders a tab bar for navigation
func renderProjectTabBar(tabs []string, selected int, theme styles.Theme) string {
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

// updateProjectsTable updates the projects table with current data
func (m *ProjectsModel) updateProjectsTable() {
	// This method updates the table rows with current project data
	// The actual update happens in renderOverview()
}
