// Package tui provides the terminal user interface for Prism.
//
// This package implements a full-featured TUI using the BubbleTea framework,
// providing an interactive alternative to the command-line interface.
package tui

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/scttfrdmn/prism/internal/tui/api"
	"github.com/scttfrdmn/prism/internal/tui/models"
	"github.com/scttfrdmn/prism/pkg/api/client"
	"github.com/scttfrdmn/prism/pkg/profile"
	"github.com/scttfrdmn/prism/pkg/update"
	"github.com/scttfrdmn/prism/pkg/version"
)

// App represents the TUI application
type App struct {
	apiClient *api.TUIClient
	program   *tea.Program
}

// PageID represents different pages in the TUI
type PageID int

const (
	// DashboardPage shows the main dashboard
	DashboardPage PageID = iota
	// InstancesPage shows instance management
	InstancesPage
	// TemplatesPage shows template selection
	TemplatesPage
	// StoragePage shows storage management
	StoragePage
	// ProjectsPage shows project management (Phase 4 Enterprise)
	ProjectsPage
	// BudgetPage shows budget management (Phase 4 Enterprise)
	BudgetPage
	// UsersPage shows user management (Phase 5A.2)
	UsersPage
	// PolicyPage shows policy framework management (Phase 5A+)
	PolicyPage
	// MarketplacePage shows template marketplace (Phase 5B)
	MarketplacePage
	// IdlePage shows idle detection and hibernation management (Phase 3)
	IdlePage
	// AMIPage shows AMI management
	AMIPage
	// RightsizingPage shows rightsizing recommendations
	RightsizingPage
	// LogsPage shows logs viewer
	LogsPage
	// DaemonPage shows daemon management
	DaemonPage
	// SettingsPage shows application settings
	SettingsPage
	// ProfilesPage shows profile management
	ProfilesPage
)

// AppModel represents the main application model
type AppModel struct {
	apiClient        *api.TUIClient
	currentPage      PageID
	dashboardModel   models.DashboardModel
	instancesModel   models.InstancesModel
	templatesModel   models.TemplatesModel
	storageModel     models.StorageModel
	projectsModel    models.ProjectsModel
	budgetModel      models.BudgetModel
	usersModel       models.UsersModel
	policyModel      models.PolicyModel
	marketplaceModel models.MarketplaceModel
	idleModel        models.IdleModel
	amiModel         models.AMIModel
	rightsizingModel models.RightsizingModel
	logsModel        models.LogsModel
	daemonModel      models.DaemonModel
	settingsModel    models.SettingsModel
	profilesModel    models.ProfilesModel
	width            int
	height           int
	inSubmenu        bool   // Whether we're currently in a submenu
	submenuParent    PageID // The page that opened the submenu
	updateInfo       *update.UpdateInfo
}

// UpdateCheckMsg represents the result of an update check
type UpdateCheckMsg struct {
	UpdateInfo *update.UpdateInfo
	Error      error
}

// NewApp creates a new TUI application
func NewApp() *App {
	// Get current profile for API client configuration
	profileManager, pmErr := profile.NewManagerEnhanced()
	var currentProfile *profile.Profile
	if pmErr != nil {
		// Use default profile if manager fails to initialize
		currentProfile = &profile.Profile{
			Name:       "default",
			AWSProfile: "",
			Region:     "",
		}
	} else {
		prof, err := profileManager.GetCurrentProfile()
		if err != nil {
			// Use default profile if none exists
			currentProfile = &profile.Profile{
				Name:       "default",
				AWSProfile: "",
				Region:     "",
			}
		} else {
			currentProfile = prof
		}
	}

	// Load API key from daemon state if available
	apiKey := loadAPIKeyFromState()

	// Create API client with modern Options pattern
	apiClient := client.NewClientWithOptions("http://localhost:8947", client.Options{
		AWSProfile: currentProfile.AWSProfile,
		AWSRegion:  currentProfile.Region,
		APIKey:     apiKey,
	})

	// Wrap with TUI client
	tuiClient := api.NewTUIClient(apiClient)

	return &App{
		apiClient: tuiClient,
		program:   nil,
	}
}

// Run starts the TUI application
func (a *App) Run() error {
	// Create initial model
	model := AppModel{
		apiClient:        a.apiClient,
		currentPage:      DashboardPage,
		dashboardModel:   models.NewDashboardModel(a.apiClient),
		instancesModel:   models.NewInstancesModel(a.apiClient),
		templatesModel:   models.NewTemplatesModel(a.apiClient),
		storageModel:     models.NewStorageModel(a.apiClient),
		projectsModel:    models.NewProjectsModel(a.apiClient),
		budgetModel:      models.NewBudgetModel(a.apiClient),
		usersModel:       models.NewUsersModel(a.apiClient),
		policyModel:      models.NewPolicyModel(a.apiClient),
		marketplaceModel: models.NewMarketplaceModel(a.apiClient),
		idleModel:        models.NewIdleModel(a.apiClient),
		amiModel:         models.NewAMIModel(a.apiClient),
		rightsizingModel: models.NewRightsizingModel(a.apiClient),
		logsModel:        models.NewLogsModel(a.apiClient),
		daemonModel:      models.NewDaemonModel(a.apiClient),
		settingsModel:    models.NewSettingsModel(a.apiClient),
		profilesModel:    models.NewProfilesModel(a.apiClient),
	}

	// Create program with explicit input/output streams for maximum compatibility
	program := tea.NewProgram(
		model,
		tea.WithInput(os.Stdin),
		tea.WithOutput(os.Stderr), // Use stderr to avoid conflicts with stdout
	)

	// Store program reference
	a.program = program

	// Run the application
	_, err := program.Run()
	return err
}

// Init initializes the application model
func (m AppModel) Init() tea.Cmd {
	// Start background update check
	updateCheckCmd := checkForUpdatesBackground()

	// Initialize current page
	var pageInitCmd tea.Cmd
	switch m.currentPage {
	case DashboardPage:
		pageInitCmd = m.dashboardModel.Init()
	case InstancesPage:
		pageInitCmd = m.instancesModel.Init()
	case TemplatesPage:
		pageInitCmd = m.templatesModel.Init()
	case StoragePage:
		pageInitCmd = m.storageModel.Init()
	case ProjectsPage:
		pageInitCmd = m.projectsModel.Init()
	case BudgetPage:
		pageInitCmd = m.budgetModel.Init()
	case UsersPage:
		pageInitCmd = m.usersModel.Init()
	case PolicyPage:
		pageInitCmd = m.policyModel.Init()
	case MarketplacePage:
		pageInitCmd = m.marketplaceModel.Init()
	case IdlePage:
		pageInitCmd = m.idleModel.Init()
	case AMIPage:
		pageInitCmd = m.amiModel.Init()
	case RightsizingPage:
		pageInitCmd = m.rightsizingModel.Init()
	case LogsPage:
		pageInitCmd = m.logsModel.Init()
	case DaemonPage:
		pageInitCmd = m.daemonModel.Init()
	case SettingsPage:
		pageInitCmd = m.settingsModel.Init()
	case ProfilesPage:
		pageInitCmd = m.profilesModel.Init()
	default:
		pageInitCmd = m.dashboardModel.Init()
	}

	// Batch both commands
	return tea.Batch(updateCheckCmd, pageInitCmd)
}

// checkForUpdatesBackground performs a background update check
func checkForUpdatesBackground() tea.Cmd {
	return func() tea.Msg {
		checker, err := update.NewChecker()
		if err != nil {
			return UpdateCheckMsg{UpdateInfo: nil, Error: err}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		updateInfo, err := checker.CheckForUpdates(ctx)
		if err != nil {
			// Silently fail - don't disrupt user experience
			return UpdateCheckMsg{UpdateInfo: nil, Error: err}
		}

		return UpdateCheckMsg{UpdateInfo: updateInfo, Error: nil}
	}
}

// AppMessageHandler interface for handling different app messages (Command Pattern - SOLID)
type AppMessageHandler interface {
	CanHandle(msg tea.Msg) bool
	Handle(m AppModel, msg tea.Msg) (AppModel, []tea.Cmd)
}

// WindowSizeHandler handles window size messages
type WindowSizeHandler struct{}

func (h *WindowSizeHandler) CanHandle(msg tea.Msg) bool {
	_, ok := msg.(tea.WindowSizeMsg)
	return ok
}

func (h *WindowSizeHandler) Handle(m AppModel, msg tea.Msg) (AppModel, []tea.Cmd) {
	windowMsg := msg.(tea.WindowSizeMsg)
	m.width = windowMsg.Width
	m.height = windowMsg.Height
	return m, nil
}

// QuitKeyHandler handles quit key messages
type QuitKeyHandler struct{}

func (h *QuitKeyHandler) CanHandle(msg tea.Msg) bool {
	keyMsg, ok := msg.(tea.KeyMsg)
	return ok && (keyMsg.String() == "ctrl+c" || keyMsg.String() == "q")
}

func (h *QuitKeyHandler) Handle(m AppModel, msg tea.Msg) (AppModel, []tea.Cmd) {
	return m, []tea.Cmd{tea.Quit}
}

// UpdateCheckHandler handles update check messages
type UpdateCheckHandler struct{}

func (h *UpdateCheckHandler) CanHandle(msg tea.Msg) bool {
	_, ok := msg.(UpdateCheckMsg)
	return ok
}

func (h *UpdateCheckHandler) Handle(m AppModel, msg tea.Msg) (AppModel, []tea.Cmd) {
	updateMsg := msg.(UpdateCheckMsg)
	if updateMsg.Error == nil && updateMsg.UpdateInfo != nil {
		m.updateInfo = updateMsg.UpdateInfo
		// Update all status bars with update information
		if updateMsg.UpdateInfo.IsUpdateAvailable {
			m.settingsModel.SetUpdateInfo(updateMsg.UpdateInfo)
		}
	}
	return m, nil
}

// PageNavigationHandler handles page navigation keys
type PageNavigationHandler struct{}

// pageNavKeys is the set of keys handled by PageNavigationHandler.
var pageNavKeys = map[string]bool{
	"1": true, "2": true, "3": true, "4": true, "5": true,
	"6": true, "7": true, "8": true, "9": true, "0": true,
	"m": true, "i": true, "a": true, "r": true, "l": true,
	"d": true, "esc": true,
}

func (h *PageNavigationHandler) CanHandle(msg tea.Msg) bool {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return false
	}
	return pageNavKeys[keyMsg.String()]
}

func (h *PageNavigationHandler) Handle(m AppModel, msg tea.Msg) (AppModel, []tea.Cmd) {
	keyMsg := msg.(tea.KeyMsg)
	var cmds []tea.Cmd
	key := keyMsg.String()

	// Handle ESC - exit submenu if we're in one
	if key == "esc" && m.inSubmenu {
		m.inSubmenu = false
		m.currentPage = m.submenuParent
		return m, cmds
	}

	// If we're in the settings advanced submenu, handle submenu navigation
	if m.inSubmenu && m.submenuParent == SettingsPage {
		return handleSettingsSubmenuNav(m, key)
	}

	// Handle 'a' key on Settings page - enter advanced submenu
	if key == "a" && m.currentPage == SettingsPage && !m.inSubmenu {
		m.inSubmenu = true
		m.submenuParent = SettingsPage
		return m, cmds
	}

	// Normal page navigation (not in submenu)
	return handleMainPageNav(m, key)
}

// handleSettingsSubmenuNav handles key presses when inside the Settings advanced submenu.
func handleSettingsSubmenuNav(m AppModel, key string) (AppModel, []tea.Cmd) {
	var cmds []tea.Cmd
	switch key {
	case "1":
		m.currentPage = AMIPage
		m.inSubmenu = false
		cmds = append(cmds, m.amiModel.Init())
	case "2":
		m.currentPage = RightsizingPage
		m.inSubmenu = false
		cmds = append(cmds, m.rightsizingModel.Init())
	case "3":
		m.currentPage = IdlePage
		m.inSubmenu = false
		cmds = append(cmds, m.idleModel.Init())
	case "4":
		m.currentPage = PolicyPage
		m.inSubmenu = false
		cmds = append(cmds, m.policyModel.Init())
	case "5":
		m.currentPage = MarketplacePage
		m.inSubmenu = false
		cmds = append(cmds, m.marketplaceModel.Init())
	case "6":
		m.currentPage = LogsPage
		m.inSubmenu = false
		cmds = append(cmds, m.logsModel.Init())
	}
	return m, cmds
}

// handleMainPageNav handles top-level page navigation key presses.
func handleMainPageNav(m AppModel, key string) (AppModel, []tea.Cmd) {
	var cmds []tea.Cmd
	switch key {
	case "1":
		m.currentPage = DashboardPage
	case "2":
		m.currentPage = InstancesPage
		cmds = append(cmds, m.instancesModel.Init())
	case "3":
		m.currentPage = TemplatesPage
		cmds = append(cmds, m.templatesModel.Init())
	case "4":
		m.currentPage = StoragePage
		cmds = append(cmds, m.storageModel.Init())
	case "5":
		m.currentPage = ProjectsPage
		cmds = append(cmds, m.projectsModel.Init())
	case "6":
		m.currentPage = BudgetPage
		cmds = append(cmds, m.budgetModel.Init())
	case "7":
		m.currentPage = UsersPage
		cmds = append(cmds, m.usersModel.Init())
	case "8":
		m.currentPage = SettingsPage
		cmds = append(cmds, m.settingsModel.Init())
	case "9":
		m.currentPage = ProfilesPage
		m.profilesModel.SetSize(m.width, m.height)
		cmds = append(cmds, func() tea.Msg { return models.ProfileInitMsg{} })
	}
	return m, cmds
}

// PageModelUpdater handles updating the current page model
type PageModelUpdater struct{}

func (u *PageModelUpdater) UpdateCurrentPage(m AppModel, msg tea.Msg) (AppModel, tea.Cmd) {
	switch m.currentPage {
	case DashboardPage:
		newModel, newCmd := m.dashboardModel.Update(msg)
		m.dashboardModel = newModel.(models.DashboardModel)
		return m, newCmd
	case InstancesPage:
		newModel, newCmd := m.instancesModel.Update(msg)
		m.instancesModel = newModel.(models.InstancesModel)
		return m, newCmd
	case TemplatesPage:
		newModel, newCmd := m.templatesModel.Update(msg)
		m.templatesModel = newModel.(models.TemplatesModel)
		return m, newCmd
	case StoragePage:
		newModel, newCmd := m.storageModel.Update(msg)
		m.storageModel = newModel.(models.StorageModel)
		return m, newCmd
	case ProjectsPage:
		newModel, newCmd := m.projectsModel.Update(msg)
		m.projectsModel = newModel.(models.ProjectsModel)
		return m, newCmd
	case BudgetPage:
		newModel, newCmd := m.budgetModel.Update(msg)
		m.budgetModel = newModel.(models.BudgetModel)
		return m, newCmd
	case UsersPage:
		newModel, newCmd := m.usersModel.Update(msg)
		m.usersModel = newModel.(models.UsersModel)
		return m, newCmd
	case PolicyPage:
		newModel, newCmd := m.policyModel.Update(msg)
		m.policyModel = newModel.(models.PolicyModel)
		return m, newCmd
	case MarketplacePage:
		newModel, newCmd := m.marketplaceModel.Update(msg)
		m.marketplaceModel = newModel.(models.MarketplaceModel)
		return m, newCmd
	case IdlePage:
		newModel, newCmd := m.idleModel.Update(msg)
		m.idleModel = newModel.(models.IdleModel)
		return m, newCmd
	case AMIPage:
		newModel, newCmd := m.amiModel.Update(msg)
		m.amiModel = newModel.(models.AMIModel)
		return m, newCmd
	case RightsizingPage:
		newModel, newCmd := m.rightsizingModel.Update(msg)
		m.rightsizingModel = newModel.(models.RightsizingModel)
		return m, newCmd
	case LogsPage:
		newModel, newCmd := m.logsModel.Update(msg)
		m.logsModel = newModel.(models.LogsModel)
		return m, newCmd
	case DaemonPage:
		newModel, newCmd := m.daemonModel.Update(msg)
		m.daemonModel = newModel.(models.DaemonModel)
		return m, newCmd
	case SettingsPage:
		newModel, newCmd := m.settingsModel.Update(msg)
		m.settingsModel = newModel.(models.SettingsModel)
		return m, newCmd
	case ProfilesPage:
		newModel, newCmd := m.profilesModel.Update(msg)
		m.profilesModel = newModel.(models.ProfilesModel)
		return m, newCmd
	}
	return m, nil
}

// AppMessageDispatcher manages app message handlers (Command Pattern - SOLID)
type AppMessageDispatcher struct {
	handlers []AppMessageHandler
	updater  *PageModelUpdater
}

// NewAppMessageDispatcher creates app message dispatcher
func NewAppMessageDispatcher() *AppMessageDispatcher {
	return &AppMessageDispatcher{
		handlers: []AppMessageHandler{
			&WindowSizeHandler{},
			&UpdateCheckHandler{},
			&QuitKeyHandler{},
			&PageNavigationHandler{},
		},
		updater: &PageModelUpdater{},
	}
}

// Dispatch processes message using appropriate handler
func (d *AppMessageDispatcher) Dispatch(m AppModel, msg tea.Msg) (AppModel, tea.Cmd) {
	var allCmds []tea.Cmd

	// Try global handlers first
	for _, handler := range d.handlers {
		if handler.CanHandle(msg) {
			newModel, cmds := handler.Handle(m, msg)
			if cmds != nil {
				allCmds = append(allCmds, cmds...)
			}
			m = newModel
			break
		}
	}

	// Update current page model
	newModel, pageCmd := d.updater.UpdateCurrentPage(m, msg)
	if pageCmd != nil {
		allCmds = append(allCmds, pageCmd)
	}

	return newModel, tea.Batch(allCmds...)
}

// Update handles messages using Command Pattern (SOLID: Single Responsibility)
func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	dispatcher := NewAppMessageDispatcher()
	newModel, cmd := dispatcher.Dispatch(m, msg)
	return newModel, cmd
}

// View renders the application
func (m AppModel) View() string {
	// Get base view
	var baseView string
	switch m.currentPage {
	case DashboardPage:
		baseView = m.dashboardModel.View()
	case InstancesPage:
		baseView = m.instancesModel.View()
	case TemplatesPage:
		baseView = m.templatesModel.View()
	case StoragePage:
		baseView = m.storageModel.View()
	case ProjectsPage:
		baseView = m.projectsModel.View()
	case BudgetPage:
		baseView = m.budgetModel.View()
	case UsersPage:
		baseView = m.usersModel.View()
	case PolicyPage:
		baseView = m.policyModel.View()
	case MarketplacePage:
		baseView = m.marketplaceModel.View()
	case IdlePage:
		baseView = m.idleModel.View()
	case AMIPage:
		baseView = m.amiModel.View()
	case RightsizingPage:
		baseView = m.rightsizingModel.View()
	case LogsPage:
		baseView = m.logsModel.View()
	case DaemonPage:
		baseView = m.daemonModel.View()
	case SettingsPage:
		baseView = m.settingsModel.View()
	case ProfilesPage:
		baseView = m.profilesModel.View()
	default:
		baseView = fmt.Sprintf("Prism v%s\n\nUnknown page", version.GetVersion())
	}

	// If we're in the advanced submenu, overlay the submenu
	if m.inSubmenu && m.submenuParent == SettingsPage {
		submenuOverlay := "\n\n" +
			"╔═══════════════════════════════════════════════════════════╗\n" +
			"║                    ADVANCED SETTINGS                      ║\n" +
			"╠═══════════════════════════════════════════════════════════╣\n" +
			"║  [1] AMI Management         - Custom image management     ║\n" +
			"║  [2] Rightsizing            - Instance optimization       ║\n" +
			"║  [3] Idle Detection         - Hibernation policies        ║\n" +
			"║  [4] Policy Framework       - Access control & governance ║\n" +
			"║  [5] Template Marketplace   - Community templates         ║\n" +
			"║  [6] Logs Viewer            - System logs and diagnostics ║\n" +
			"╠═══════════════════════════════════════════════════════════╣\n" +
			"║  Press number to navigate • ESC to return to Settings     ║\n" +
			"╚═══════════════════════════════════════════════════════════╝\n"
		return baseView + submenuOverlay
	}

	return baseView
}

// loadAPIKeyFromState attempts to load the API key from daemon state
func loadAPIKeyFromState() string {
	// Try to load daemon state to get API key
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "" // No API key available
	}

	stateFile := filepath.Join(homeDir, ".prism", "state.json")
	data, err := os.ReadFile(stateFile)
	if err != nil {
		return "" // No state file or can't read it
	}

	// Parse state to extract API key
	var state struct {
		Config struct {
			APIKey string `json:"api_key"`
		} `json:"config"`
	}

	if err := json.Unmarshal(data, &state); err != nil {
		return "" // Invalid state format
	}

	return state.Config.APIKey
}
