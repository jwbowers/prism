// Package daemon provides profile manager adapter for invitation validation
package daemon

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/scttfrdmn/prism/pkg/invitation"
)

// ProfileManagerAdapter adapts the Server to implement invitation.ProfileManager interface
type ProfileManagerAdapter struct {
	server *Server
}

// NewProfileManagerAdapter creates a new ProfileManagerAdapter
func NewProfileManagerAdapter(s *Server) *ProfileManagerAdapter {
	return &ProfileManagerAdapter{
		server: s,
	}
}

// GetProfile retrieves a profile by ID
func (p *ProfileManagerAdapter) GetProfile(ctx context.Context, profileID string) (*invitation.Profile, error) {
	// Get state to get default profile information
	if p.server.stateManager == nil {
		return nil, fmt.Errorf("state manager not initialized")
	}

	state, err := p.server.stateManager.LoadState()
	if err != nil {
		return nil, fmt.Errorf("failed to load state: %w", err)
	}

	// Get default AWS profile from config
	awsProfile := state.Config.DefaultProfile
	if awsProfile == "" {
		awsProfile = "default" // Fall back to default AWS profile
	}

	region := state.Config.DefaultRegion
	if region == "" {
		region = "us-west-2" // Fall back to default region
	}

	// Create profile from state config
	// In a full implementation, you would look up the specific profile by ID
	// For now, we'll use the default profile configuration
	profile := &invitation.Profile{
		ID:         profileID,
		Name:       awsProfile,
		AWSProfile: awsProfile,
		Region:     region,
		ExpiresAt:  nil, // TODO: Look up expiration if stored
	}

	return profile, nil
}

// GetProfileCredentials retrieves AWS credentials for a profile
func (p *ProfileManagerAdapter) GetProfileCredentials(ctx context.Context, profileID string) (aws.Credentials, error) {
	// Get profile to determine AWS profile name
	profile, err := p.GetProfile(ctx, profileID)
	if err != nil {
		return aws.Credentials{}, err
	}

	// Load AWS config using the profile's AWS profile name
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithSharedConfigProfile(profile.AWSProfile),
		config.WithRegion(profile.Region),
	)
	if err != nil {
		return aws.Credentials{}, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Retrieve credentials
	creds, err := cfg.Credentials.Retrieve(ctx)
	if err != nil {
		return aws.Credentials{}, fmt.Errorf("failed to retrieve credentials: %w", err)
	}

	return creds, nil
}
