package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// InitCobraCommands handles the init wizard command
type InitCobraCommands struct {
	app *App
}

// NewInitCobraCommands creates new init cobra commands
func NewInitCobraCommands(app *App) *InitCobraCommands {
	return &InitCobraCommands{app: app}
}

// CreateInitCommand creates the init command
func (ic *InitCobraCommands) CreateInitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "init",
		Short:   "Launch your first workspace (guided setup)",
		GroupID: "core",
		Long: `Interactive wizard to launch your first Prism workspace.

This command guides you through:
  1. Template selection from categorized options
  2. Workspace configuration (name and size)
  3. Review and confirmation
  4. Launch with real-time progress
  5. Connection details and next steps

Perfect for first-time users - launches your workspace in under 30 seconds!`,
		RunE: func(cmd *cobra.Command, args []string) error {
			skipAWSCheck, _ := cmd.Flags().GetBool("skip-aws-check")
			return ic.runInitWizard(skipAWSCheck)
		},
	}

	cmd.Flags().Bool("skip-aws-check", false, "Skip AWS credential validation")

	return cmd
}

// runInitWizard executes the init wizard flow
func (ic *InitCobraCommands) runInitWizard(skipAWSCheck bool) error {
	// Step 1: Welcome message
	ic.printWelcome()

	// Step 2: Check AWS credentials
	if !skipAWSCheck {
		if err := ic.checkAWSCredentials(); err != nil {
			return ic.guideAWSSetup(err)
		}
		fmt.Println("✅ AWS credentials validated")
		fmt.Println()
	}

	// Step 3: Select template
	template, err := ic.selectTemplate()
	if err != nil {
		return err
	}

	// Step 4: Configure workspace
	name, size, err := ic.configureWorkspace(template)
	if err != nil {
		return err
	}

	// Step 5: Review and confirm
	if !ic.reviewAndConfirm(template, name, size) {
		fmt.Println("\n❌ Launch cancelled")
		return nil
	}

	// Step 6: Launch with progress
	if err := ic.launchWorkspace(template, name, size); err != nil {
		return err
	}

	// Step 7: Display success and connection info
	return ic.displaySuccess(name)
}

// printWelcome displays the welcome message
func (ic *InitCobraCommands) printWelcome() {
	fmt.Println("🎉 Welcome to Prism!")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Println("This wizard will help you launch your first research workspace.")
	fmt.Println("Launch time: ~30 seconds")
	fmt.Println()
}

// checkAWSCredentials validates AWS credentials
func (ic *InitCobraCommands) checkAWSCredentials() error {
	// Ensure daemon is running
	if err := ic.app.ensureDaemonRunning(); err != nil {
		return fmt.Errorf("failed to start daemon: %w", err)
	}

	// Basic API call to validate credentials
	client := ic.app.apiClient
	ctx := context.Background()
	_, err := client.ListInstances(ctx)
	return err
}

// guideAWSSetup provides AWS setup guidance
func (ic *InitCobraCommands) guideAWSSetup(err error) error {
	fmt.Println("❌ AWS credentials not configured")
	fmt.Println()
	fmt.Println("Are you at a university or research institution?")
	fmt.Println("  Your institution may provide free AWS accounts for researchers.")
	fmt.Println("  Check if yours is registered:")
	fmt.Println()
	fmt.Println("    prism schools search \"your university name\"")
	fmt.Println()
	fmt.Println("  Or browse all registered institutions:")
	fmt.Println()
	fmt.Println("    prism schools list")
	fmt.Println()
	fmt.Println("To set up a personal AWS account:")
	fmt.Println()
	fmt.Println("  1. Install AWS CLI v2 (https://aws.amazon.com/cli/):")
	fmt.Println("       brew install awscli  # macOS")
	fmt.Println()
	fmt.Println("  2. Authenticate (browser-based, no keys needed):")
	fmt.Println("       aws login")
	fmt.Println()
	fmt.Println("     Or with long-term access keys:")
	fmt.Println("       aws configure")
	fmt.Println()
	fmt.Println("  3. Run 'prism init' again")
	fmt.Println()
	fmt.Println("For detailed setup: https://scttfrdmn.github.io/prism/user-guides/aws-setup/")
	fmt.Println()

	return fmt.Errorf("AWS credentials required: %w", err)
}

// templateInfo holds template information for selection
type templateInfo struct {
	Name            string
	Slug            string
	Description     string
	RecommendedSize string
	EstimatedCost   float64
}

// selectTemplate guides the user through template selection
func (ic *InitCobraCommands) selectTemplate() (*templateInfo, error) {
	fmt.Println("📦 Step 1: Select a Template")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	// Fetch templates
	templates, err := ic.fetchTemplates()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch templates: %w", err)
	}

	if len(templates) == 0 {
		return nil, fmt.Errorf("no templates available")
	}

	// Categorize templates
	categories := ic.categorizeTemplates(templates)

	// Display categories
	fmt.Println("Choose a category:")
	fmt.Println()
	categoryNames := []string{"ML/AI", "Data Science", "Bioinformatics", "Web Development", "All Templates"}
	for i, cat := range categoryNames {
		count := len(categories[cat])
		fmt.Printf("  %d) %s (%d templates)\n", i+1, cat, count)
	}
	fmt.Println()

	// Get category selection
	catIdx := ic.promptChoice("Select category", 1, len(categoryNames))
	selectedCategory := categoryNames[catIdx-1]

	// Display templates in category
	categoryTemplates := categories[selectedCategory]
	if len(categoryTemplates) == 0 {
		return nil, fmt.Errorf("no templates in category: %s", selectedCategory)
	}

	fmt.Println()
	fmt.Printf("📋 %s Templates:\n\n", selectedCategory)

	for i, tmpl := range categoryTemplates {
		fmt.Printf("  %d) %s\n", i+1, tmpl.Name)
		if tmpl.Description != "" {
			fmt.Printf("     %s\n", tmpl.Description)
		}
		if tmpl.RecommendedSize != "" {
			fmt.Printf("     Recommended: %s (~$%.2f/hour)\n", tmpl.RecommendedSize, tmpl.EstimatedCost)
		}
		fmt.Println()
	}

	// Get template selection
	tmplIdx := ic.promptChoice("Select template", 1, len(categoryTemplates))
	selectedTemplate := categoryTemplates[tmplIdx-1]

	return selectedTemplate, nil
}

// fetchTemplates retrieves available templates from the API
func (ic *InitCobraCommands) fetchTemplates() ([]*templateInfo, error) {
	client := ic.app.apiClient
	ctx := context.Background()
	templatesMap, err := client.ListTemplates(ctx)
	if err != nil {
		return nil, err
	}

	templates := make([]*templateInfo, 0, len(templatesMap))
	for slug, tmpl := range templatesMap {
		// Extract description from template
		desc := ""
		if tmpl.Description != "" {
			desc = tmpl.Description
		}

		// Default to Medium size for all templates
		recommendedSize := "M"

		info := &templateInfo{
			Name:            tmpl.Name,
			Slug:            slug,
			Description:     desc,
			RecommendedSize: recommendedSize,
			EstimatedCost:   ic.estimateCost(recommendedSize),
		}
		templates = append(templates, info)
	}

	return templates, nil
}

func isMlAiTemplate(name, desc string) bool {
	return strings.Contains(name, "ml") || strings.Contains(name, "machine learning") ||
		strings.Contains(name, "ai") || strings.Contains(desc, "tensorflow") ||
		strings.Contains(desc, "pytorch")
}

func isDataScienceTemplate(name, _ string) bool {
	return strings.Contains(name, "python") || strings.Contains(name, "jupyter") ||
		strings.Contains(name, "data") || strings.Contains(name, "r ") ||
		strings.Contains(name, "rstudio")
}

func isBioinformaticsTemplate(name, _ string) bool {
	return strings.Contains(name, "bio") || strings.Contains(name, "genomics") ||
		strings.Contains(name, "blast")
}

func isWebDevTemplate(name, _ string) bool {
	return strings.Contains(name, "web") || strings.Contains(name, "node") ||
		strings.Contains(name, "nginx")
}

// categorizeTemplates groups templates by category
func (ic *InitCobraCommands) categorizeTemplates(templates []*templateInfo) map[string][]*templateInfo {
	categories := map[string][]*templateInfo{
		"ML/AI":           {},
		"Data Science":    {},
		"Bioinformatics":  {},
		"Web Development": {},
		"All Templates":   templates,
	}

	for _, tmpl := range templates {
		name := strings.ToLower(tmpl.Name)
		desc := strings.ToLower(tmpl.Description)
		if isMlAiTemplate(name, desc) {
			categories["ML/AI"] = append(categories["ML/AI"], tmpl)
		}
		if isDataScienceTemplate(name, desc) {
			categories["Data Science"] = append(categories["Data Science"], tmpl)
		}
		if isBioinformaticsTemplate(name, desc) {
			categories["Bioinformatics"] = append(categories["Bioinformatics"], tmpl)
		}
		if isWebDevTemplate(name, desc) {
			categories["Web Development"] = append(categories["Web Development"], tmpl)
		}
	}

	return categories
}

// configureWorkspace prompts for workspace configuration
func (ic *InitCobraCommands) configureWorkspace(template *templateInfo) (string, string, error) {
	fmt.Println()
	fmt.Println("⚙️  Step 2: Configure Workspace")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	// Prompt for workspace name
	name := ic.promptWorkspaceName()

	// Prompt for size
	size := ic.promptSize(template.RecommendedSize)

	return name, size, nil
}

// promptWorkspaceName prompts for and validates workspace name
func (ic *InitCobraCommands) promptWorkspaceName() string {
	reader := bufio.NewReader(os.Stdin)

	// Suggest a default name
	defaultName := fmt.Sprintf("my-workspace-%s", time.Now().Format("0102"))

	// Pre-compile regexes for performance
	namePattern := regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]*[a-zA-Z0-9]$`)
	singleCharPattern := regexp.MustCompile(`^[a-zA-Z0-9]$`)

	for {
		fmt.Printf("Workspace name (default: %s): ", defaultName)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "" {
			return defaultName
		}

		// Validate name (alphanumeric and hyphens, must start and end with alphanumeric)
		if namePattern.MatchString(input) {
			return input
		}

		// Handle single character names
		if len(input) == 1 && singleCharPattern.MatchString(input) {
			return input
		}

		fmt.Println("❌ Name must contain only letters, numbers, and hyphens")
		fmt.Println("   Must start and end with a letter or number")
		fmt.Println()
	}
}

// promptSize prompts for workspace size selection
func (ic *InitCobraCommands) promptSize(recommendedSize string) string {
	fmt.Println()
	fmt.Println("Choose workspace size:")
	fmt.Println()

	sizes := []struct {
		name string
		spec string
		cost string
	}{
		{"S", "2 vCPU, 4GB RAM", "~$0.08/hour"},
		{"M", "4 vCPU, 8GB RAM", "~$0.16/hour"},
		{"L", "8 vCPU, 16GB RAM", "~$0.32/hour"},
		{"XL", "16 vCPU, 32GB RAM", "~$0.64/hour"},
	}

	for i, size := range sizes {
		marker := "  "
		if size.name == recommendedSize {
			marker = "→ "
		}
		fmt.Printf("%s%d) %s - %s (%s)\n", marker, i+1, size.name, size.spec, size.cost)
	}

	fmt.Println()
	if recommendedSize != "" {
		fmt.Printf("💡 Tip: Size '%s' is recommended for this template\n\n", recommendedSize)
	}

	choice := ic.promptChoice("Select size", 1, len(sizes))
	return sizes[choice-1].name
}

// reviewAndConfirm displays configuration summary and asks for confirmation
func (ic *InitCobraCommands) reviewAndConfirm(template *templateInfo, name, size string) bool {
	fmt.Println()
	fmt.Println("📋 Step 3: Review Configuration")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	// Display summary
	fmt.Println("Your workspace configuration:")
	fmt.Println()
	fmt.Printf("  Template:  %s\n", template.Name)
	fmt.Printf("  Name:      %s\n", name)
	fmt.Printf("  Size:      %s\n", size)
	fmt.Println()

	// Show cost estimate
	costPerHour := ic.estimateCost(size)
	costPerMonth := costPerHour * 730 // Average hours per month
	fmt.Printf("  Estimated cost: $%.2f/hour (~$%.2f/month if running 24/7)\n", costPerHour, costPerMonth)
	fmt.Println()
	fmt.Println("💡 Tip: Use 'prism workspace stop' when not in use to save costs")
	fmt.Println()

	// Confirm
	return ic.promptConfirm("Launch this workspace?")
}

// launchWorkspace initiates the workspace launch
func (ic *InitCobraCommands) launchWorkspace(template *templateInfo, name, size string) error {
	fmt.Println()
	fmt.Println("🚀 Step 4: Launching Workspace")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	// Build launch request
	launchArgs := []string{
		"launch",
		template.Slug,
		name,
		"--size", size,
		"--wait", // Wait for launch to complete
	}

	// Launch via existing Launch method
	fmt.Println("⏳ Launching workspace... This may take 1-2 minutes")
	fmt.Println()

	err := ic.app.Launch(launchArgs)
	if err != nil {
		fmt.Println()
		fmt.Printf("❌ Failed to launch workspace: %v\n", err)
		return err
	}

	return nil
}

// displaySuccess shows success message and connection information
func (ic *InitCobraCommands) displaySuccess(name string) error {
	fmt.Println()
	fmt.Println("✅ Success! Your workspace is ready")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	// Get workspace details
	client := ic.app.apiClient
	ctx := context.Background()
	instance, err := client.GetInstance(ctx, name)
	if err != nil {
		// Still show basic success even if we can't get details
		fmt.Println("📚 Next Steps:")
		fmt.Println("  • Connect:  prism workspace connect", name)
		fmt.Println("  • Monitor:  prism workspace list")
		fmt.Println("  • Stop:     prism workspace stop", name)
		fmt.Println()
		return nil
	}

	// Display connection info
	fmt.Println("📡 Connection Information:")
	fmt.Println()
	fmt.Printf("  Name:      %s\n", instance.Name)
	fmt.Printf("  Status:    %s\n", instance.State)
	if instance.PublicIP != "" {
		fmt.Printf("  Public IP: %s\n", instance.PublicIP)
	}
	fmt.Println()

	// SSH command
	if instance.PublicIP != "" {
		fmt.Println("🔗 Connect via SSH:")
		fmt.Printf("  ssh ubuntu@%s\n", instance.PublicIP)
		fmt.Println()
	}

	// Next steps
	fmt.Println("📚 Next Steps:")
	fmt.Println("  • Connect:  prism workspace connect", name)
	fmt.Println("  • Monitor:  prism workspace list")
	fmt.Println("  • Stop:     prism workspace stop", name)
	fmt.Println("  • Delete:   prism workspace delete", name)
	fmt.Println()
	fmt.Println("💡 Run 'prism --help' to see all available commands")
	fmt.Println()

	writeInitialized()
	return nil
}

// Helper functions

// promptChoice prompts for a numeric choice within a range
func (ic *InitCobraCommands) promptChoice(prompt string, min, max int) int {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("%s [%d-%d]: ", prompt, min, max)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		choice, err := strconv.Atoi(input)
		if err == nil && choice >= min && choice <= max {
			return choice
		}

		fmt.Printf("❌ Please enter a number between %d and %d\n\n", min, max)
	}
}

// promptConfirm prompts for yes/no confirmation
func (ic *InitCobraCommands) promptConfirm(prompt string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s [y/N]: ", prompt)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))
	return input == "y" || input == "yes"
}

// IsInitialized reports whether the user has previously completed prism init.
func IsInitialized() bool {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	_, err = os.Stat(filepath.Join(homeDir, ".prism", ".initialized"))
	return err == nil
}

// writeInitialized records that the user has completed init.
func writeInitialized() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}
	dir := filepath.Join(homeDir, ".prism")
	_ = os.MkdirAll(dir, 0700)
	content := fmt.Sprintf("initialized: true\ndate: %s\n", time.Now().Format("2006-01-02"))
	_ = os.WriteFile(filepath.Join(dir, ".initialized"), []byte(content), 0600)
}

// estimateCost returns estimated hourly cost for a size
func (ic *InitCobraCommands) estimateCost(size string) float64 {
	costs := map[string]float64{
		"S":  0.08,
		"M":  0.16,
		"L":  0.32,
		"XL": 0.64,
	}
	if cost, ok := costs[size]; ok {
		return cost
	}
	return 0.16 // default to M
}
