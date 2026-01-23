package templates

import (
	"fmt"
	"time"
)

// TrustManager manages trust levels for community templates
// Simplified implementation for v0.7.2 - full trust system in v0.7.3
type TrustManager struct {
	config *TrustConfig
}

// TrustConfig configures trust management
type TrustConfig struct {
	DefaultTrustLevel string
	AutoVerify        bool
}

// NewTrustManager creates a new trust manager
func NewTrustManager(config *TrustConfig) *TrustManager {
	if config == nil {
		config = &TrustConfig{
			DefaultTrustLevel: "community",
			AutoVerify:        false,
		}
	}
	return &TrustManager{
		config: config,
	}
}

// VerificationResult contains trust verification results
type VerificationResult struct {
	TrustLevel   string // verified, community, unverified
	Verified     bool
	Passed       bool
	Issues       []string
	Warnings     []string
	VerifiedDate time.Time
}

// VerifyTemplate verifies a template's trust level
// Simplified for v0.7.2 - assigns default "community" level
func (tm *TrustManager) VerifyTemplate(template *Template, sourceURL string) (*VerificationResult, error) {
	return &VerificationResult{
		TrustLevel:   "community",
		Verified:     false,
		Passed:       true,
		Issues:       []string{},
		Warnings:     []string{"Note: Full verification available in v0.7.3"},
		VerifiedDate: time.Now(),
	}, nil
}

// GetTrustLevel returns the trust level for a template source
func (tm *TrustManager) GetTrustLevel(sourceURL string) string {
	// Simplified - always return community level
	return "community"
}

// SetTrustLevel sets the trust level for a template source
func (tm *TrustManager) SetTrustLevel(sourceURL, trustLevel string) error {
	// Simplified - accept but don't persist
	if trustLevel != "verified" && trustLevel != "community" && trustLevel != "unverified" {
		return fmt.Errorf("invalid trust level: %s", trustLevel)
	}
	return nil
}
