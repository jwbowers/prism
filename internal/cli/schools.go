package cli

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const schoolRegistryURL = "https://raw.githubusercontent.com/scttfrdmn/prism-school-registry/main/schools.yaml"

// schoolRegistry is the parsed top-level structure of schools.yaml
type schoolRegistry struct {
	Version string   `yaml:"version"`
	Schools []school `yaml:"schools"`
}

// school represents one institution in the registry
type school struct {
	ID           string   `yaml:"id"`
	Name         string   `yaml:"name"`
	ShortName    string   `yaml:"short_name"`
	Country      string   `yaml:"country"`
	State        string   `yaml:"state"`
	EmailDomains []string `yaml:"email_domains"`
	AWSPortalURL string   `yaml:"aws_portal_url"`
	AWSProgram   string   `yaml:"aws_program"`
	Notes        string   `yaml:"notes"`
	Verified     bool     `yaml:"verified"`
	LastVerified string   `yaml:"last_verified"`
}

// Schools processes schools-related commands
func (a *App) Schools(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing schools command (list, search, info)")
	}

	subcommand := args[0]
	subargs := args[1:]

	switch subcommand {
	case "list":
		return a.handleSchoolsList(subargs)
	case "search":
		return a.handleSchoolsSearch(subargs)
	case "info":
		return a.handleSchoolsInfo(subargs)
	default:
		return fmt.Errorf("unknown schools command: %s", subcommand)
	}
}

// fetchSchoolRegistry downloads and parses the schools registry
func fetchSchoolRegistry() (*schoolRegistry, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(schoolRegistryURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch school registry: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("school registry returned HTTP %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read school registry: %w", err)
	}

	var registry schoolRegistry
	if err := yaml.Unmarshal(data, &registry); err != nil {
		return nil, fmt.Errorf("failed to parse school registry: %w", err)
	}

	return &registry, nil
}

// handleSchoolsList lists all registered institutions
func (a *App) handleSchoolsList(_ []string) error {
	registry, err := fetchSchoolRegistry()
	if err != nil {
		return err
	}

	if len(registry.Schools) == 0 {
		fmt.Println("No institutions registered yet.")
		fmt.Println("Add yours: https://github.com/scttfrdmn/prism-school-registry/issues/new/choose")
		return nil
	}

	fmt.Printf("🏛️  Registered Institutions (%d)\n\n", len(registry.Schools))

	for _, s := range registry.Schools {
		printSchoolSummary(s)
	}

	fmt.Printf("💡 Find yours: prism schools search \"<institution name>\"\n")
	fmt.Printf("   Register yours: https://github.com/scttfrdmn/prism-school-registry/issues/new/choose\n")
	return nil
}

// handleSchoolsSearch searches institutions by name, domain, state, or country
func (a *App) handleSchoolsSearch(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: prism schools search <query>")
	}

	query := strings.ToLower(strings.Join(args, " "))

	registry, err := fetchSchoolRegistry()
	if err != nil {
		return err
	}

	var matches []school
	for _, s := range registry.Schools {
		if schoolMatchesQuery(s, query) {
			matches = append(matches, s)
		}
	}

	if len(matches) == 0 {
		fmt.Printf("🔍 No institutions found matching: %q\n\n", query)
		fmt.Printf("💡 Try a shorter query, or register your institution:\n")
		fmt.Printf("   https://github.com/scttfrdmn/prism-school-registry/issues/new/choose\n")
		return nil
	}

	fmt.Printf("🔍 Institutions matching %q (%d found)\n\n", query, len(matches))
	for _, s := range matches {
		printSchoolSummary(s)
	}

	return nil
}

// handleSchoolsInfo shows detailed information about a specific institution
func (a *App) handleSchoolsInfo(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: prism schools info <id-or-name>")
	}

	query := strings.ToLower(strings.Join(args, " "))

	registry, err := fetchSchoolRegistry()
	if err != nil {
		return err
	}

	// Try exact ID match first, then partial name match
	var found *school
	for i := range registry.Schools {
		s := &registry.Schools[i]
		if strings.EqualFold(s.ID, query) {
			found = s
			break
		}
		if strings.Contains(strings.ToLower(s.Name), query) ||
			strings.Contains(strings.ToLower(s.ShortName), query) {
			found = s
			// Don't break — if there's an exact ID match later, prefer that
		}
	}

	if found == nil {
		return fmt.Errorf("no institution found for %q\n\nTry: prism schools search %q", query, query)
	}

	printSchoolDetail(*found)
	return nil
}

// schoolMatchesQuery returns true if the school matches the search query
func schoolMatchesQuery(s school, query string) bool {
	if strings.Contains(strings.ToLower(s.Name), query) {
		return true
	}
	if strings.Contains(strings.ToLower(s.ShortName), query) {
		return true
	}
	if strings.Contains(strings.ToLower(s.State), query) {
		return true
	}
	if strings.Contains(strings.ToLower(s.Country), query) {
		return true
	}
	for _, domain := range s.EmailDomains {
		if strings.Contains(strings.ToLower(domain), query) {
			return true
		}
	}
	return false
}

// printSchoolSummary prints a one-line summary of an institution
func printSchoolSummary(s school) {
	verified := ""
	if s.Verified {
		verified = " ✅"
	}
	location := s.State
	if s.Country != "US" && s.Country != "" {
		location = s.Country
	}
	fmt.Printf("🏛️  %s (%s)%s — %s\n", s.Name, location, verified, s.AWSPortalURL)
	fmt.Printf("   💡 prism schools info %s\n\n", s.ID)
}

// printSchoolDetail prints full details for an institution
func printSchoolDetail(s school) {
	fmt.Printf("🏛️  %s\n\n", s.Name)

	if s.ShortName != "" {
		fmt.Printf("   Short name:    %s\n", s.ShortName)
	}
	if s.State != "" {
		fmt.Printf("   Location:      %s, %s\n", s.State, s.Country)
	} else if s.Country != "" {
		fmt.Printf("   Country:       %s\n", s.Country)
	}
	if len(s.EmailDomains) > 0 {
		fmt.Printf("   Email domains: %s\n", strings.Join(s.EmailDomains, ", "))
	}
	fmt.Printf("   AWS program:   %s\n", s.AWSProgram)
	fmt.Printf("\n")
	fmt.Printf("   🌐 AWS Portal: %s\n", s.AWSPortalURL)
	fmt.Printf("\n")
	if s.Notes != "" {
		fmt.Printf("   📝 Notes: %s\n\n", s.Notes)
	}
	if s.Verified {
		fmt.Printf("   ✅ Verified (%s)\n\n", s.LastVerified)
	}

	fmt.Printf("Next steps:\n")
	fmt.Printf("  1. Visit the portal above and request an AWS account\n")
	fmt.Printf("  2. Install the AWS CLI:  brew install awscli\n")
	fmt.Printf("  3. Configure credentials: aws configure\n")
	fmt.Printf("  4. Launch your workspace: prism init\n")
}
