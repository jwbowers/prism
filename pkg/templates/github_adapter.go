package templates

import (
	"context"
	"net/http"
	"time"
)

// GitHubAdapter fetches templates from GitHub repositories
type GitHubAdapter struct {
	client *http.Client
}

// NewGitHubAdapter creates a new GitHub adapter
func NewGitHubAdapter() *GitHubAdapter {
	return &GitHubAdapter{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// FetchTemplates fetches all templates from a GitHub repository
// Simplified implementation for v0.7.2 - full features in v0.7.3
func (ga *GitHubAdapter) FetchTemplates(ctx context.Context, repoURL, branch string) ([]*Template, error) {
	// TODO v0.7.3: Implement full GitHub API integration
	// For now, return empty list
	return []*Template{}, nil
}

// GetRepositoryInfo fetches metadata about a GitHub repository
func (ga *GitHubAdapter) GetRepositoryInfo(ctx context.Context, owner, repoName string) (*RepositoryInfo, error) {
	// Simplified implementation
	return &RepositoryInfo{
		Owner:       owner,
		Name:        repoName,
		Description: "Community template repository",
		Stars:       0,
		LastUpdated: time.Now(),
	}, nil
}

// RepositoryInfo contains GitHub repository metadata
type RepositoryInfo struct {
	Owner       string
	Name        string
	Description string
	Stars       int
	LastUpdated time.Time
}
