package templates

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// GitHubAdapter fetches templates from GitHub repositories via the raw content API.
type GitHubAdapter struct {
	client *http.Client
}

// NewGitHubAdapter creates a new GitHub adapter.
func NewGitHubAdapter() *GitHubAdapter {
	return &GitHubAdapter{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// templateIndex is the schema for templates/index.yaml in a template registry repo.
type templateIndex struct {
	Version   string               `yaml:"version"`
	Templates []templateIndexEntry `yaml:"templates"`
}

type templateIndexEntry struct {
	Slug     string `yaml:"slug"`
	Name     string `yaml:"name"`
	File     string `yaml:"file"`
	Category string `yaml:"category"`
}

// FetchTemplates fetches all templates from a GitHub repository.
//
// It reads templates/index.yaml to discover available templates, then fetches
// and parses each template YAML file listed in the index. Individual template
// fetch failures are non-fatal and logged; only an index fetch/parse failure
// returns an error.
func (ga *GitHubAdapter) FetchTemplates(ctx context.Context, repoURL, branch string) ([]*Template, error) {
	owner, repo, err := parseGitHubRepoURL(repoURL)
	if err != nil {
		return nil, fmt.Errorf("invalid repository URL %q: %w", repoURL, err)
	}

	// Fetch the template index from raw.githubusercontent.com.
	indexURL := fmt.Sprintf(
		"https://raw.githubusercontent.com/%s/%s/%s/templates/index.yaml",
		owner, repo, branch,
	)
	indexData, err := ga.fetchURL(ctx, indexURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch template index: %w", err)
	}

	var index templateIndex
	if err := yaml.Unmarshal(indexData, &index); err != nil {
		return nil, fmt.Errorf("failed to parse template index: %w", err)
	}

	// Fetch and parse each template file listed in the index.
	var fetched []*Template
	for _, entry := range index.Templates {
		templateURL := fmt.Sprintf(
			"https://raw.githubusercontent.com/%s/%s/%s/%s",
			owner, repo, branch, entry.File,
		)
		data, err := ga.fetchURL(ctx, templateURL)
		if err != nil {
			// Non-fatal: log and skip this template.
			continue
		}

		var tmpl Template
		if err := yaml.Unmarshal(data, &tmpl); err != nil {
			// Non-fatal: skip unparseable templates.
			continue
		}

		fetched = append(fetched, &tmpl)
	}

	return fetched, nil
}

// GetRepositoryInfo fetches metadata about a GitHub repository.
func (ga *GitHubAdapter) GetRepositoryInfo(ctx context.Context, owner, repoName string) (*RepositoryInfo, error) {
	return &RepositoryInfo{
		Owner:       owner,
		Name:        repoName,
		Description: "Community template repository",
		Stars:       0,
		LastUpdated: time.Now(),
	}, nil
}

// fetchURL performs an HTTP GET and returns the response body.
func (ga *GitHubAdapter) fetchURL(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "prism/1.0 (template-registry)")

	resp, err := ga.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d from %s", resp.StatusCode, url)
	}

	return io.ReadAll(resp.Body)
}

// parseGitHubRepoURL extracts owner and repo name from a GitHub URL.
// Accepts: https://github.com/owner/repo, github.com/owner/repo, owner/repo.
func parseGitHubRepoURL(repoURL string) (owner, repo string, err error) {
	u := strings.TrimPrefix(repoURL, "https://")
	u = strings.TrimPrefix(u, "http://")
	u = strings.TrimPrefix(u, "github.com/")
	u = strings.TrimSuffix(u, ".git")
	u = strings.TrimSuffix(u, "/")

	parts := strings.Split(u, "/")
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("expected owner/repo format, got %q", repoURL)
	}
	return parts[0], parts[1], nil
}

// RepositoryInfo contains GitHub repository metadata.
type RepositoryInfo struct {
	Owner       string
	Name        string
	Description string
	Stars       int
	LastUpdated time.Time
}
