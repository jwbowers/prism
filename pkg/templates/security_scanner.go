package templates

// SecurityScanner validates templates for security issues
// Simplified implementation for v0.7.2 - full scanning in v0.7.3
type SecurityScanner struct {
	strictMode bool
}

// NewSecurityScanner creates a new security scanner
func NewSecurityScanner(strictMode bool) *SecurityScanner {
	return &SecurityScanner{
		strictMode: strictMode,
	}
}

// SecurityScanResult represents the result of a security scan
type SecurityScanResult struct {
	TemplateName string
	Passed       bool
	Issues       []SecurityIssue
	Warnings     []SecurityWarning
	Score        int // 0-100, higher is better
}

// SecurityIssue represents a security problem
type SecurityIssue struct {
	Severity    string // critical, high, medium, low
	Category    string
	Description string
	Location    string
	Remediation string
}

// SecurityWarning represents a potential concern
type SecurityWarning struct {
	Category    string
	Description string
	Location    string
	Suggestion  string
}

// ScanTemplate performs security scan on a template
// Simplified for v0.7.2 - always passes with default score
func (ss *SecurityScanner) ScanTemplate(template *Template) *SecurityScanResult {
	return &SecurityScanResult{
		TemplateName: template.Name,
		Passed:       true,
		Issues:       []SecurityIssue{},
		Warnings:     []SecurityWarning{},
		Score:        75, // Default neutral score
	}
}
