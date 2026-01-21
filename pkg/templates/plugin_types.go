// Package templates provides plugin types for the universal plugin system.
package templates

import "time"

// PluginManifest defines a universal plugin specification
type PluginManifest struct {
	APIVersion string         `yaml:"apiVersion" json:"apiVersion"`
	Kind       string         `yaml:"kind" json:"kind"`
	Metadata   PluginMetadata `yaml:"metadata" json:"metadata"`
	Spec       PluginSpec     `yaml:"spec" json:"spec"`
}

// PluginMetadata contains plugin identification and classification
type PluginMetadata struct {
	Name        string `yaml:"name" json:"name"`
	DisplayName string `yaml:"display_name" json:"display_name"`
	Description string `yaml:"description" json:"description"`
	Category    string `yaml:"category" json:"category"` // development, infrastructure, data, etc.
	Version     string `yaml:"version" json:"version"`
	Maintainer  string `yaml:"maintainer" json:"maintainer"`
}

// PluginSpec defines the plugin specification and behavior
type PluginSpec struct {
	Provides      []PluginCapability           `yaml:"provides,omitempty" json:"provides,omitempty"`
	Dependencies  PluginDependencies           `yaml:"dependencies,omitempty" json:"dependencies,omitempty"`
	Conflicts     []PluginConflict             `yaml:"conflicts,omitempty" json:"conflicts,omitempty"`
	Alternatives  []PluginAlternative          `yaml:"alternatives,omitempty" json:"alternatives,omitempty"`
	Installers    map[string]PluginInstallSpec `yaml:"installers" json:"installers"`
	Configuration PluginConfiguration          `yaml:"configuration,omitempty" json:"configuration,omitempty"`
	Compatibility PluginCompatibility          `yaml:"compatibility,omitempty" json:"compatibility,omitempty"`
	Security      PluginSecurityMetadata       `yaml:"security,omitempty" json:"security,omitempty"`
}

// PluginCapability defines what a plugin provides
type PluginCapability struct {
	Name string `yaml:"name" json:"name"`
	Type string `yaml:"type" json:"type"` // capability, feature, service
}

// PluginDependencies defines required and optional plugin dependencies
type PluginDependencies struct {
	Required []PluginDependency `yaml:"required,omitempty" json:"required,omitempty"`
	Optional []PluginDependency `yaml:"optional,omitempty" json:"optional,omitempty"`
}

// PluginDependency defines a single plugin dependency with version constraint
type PluginDependency struct {
	Plugin  string `yaml:"plugin" json:"plugin"`
	Version string `yaml:"version" json:"version"` // Semantic version constraint (e.g., ">=2.0.0")
}

// PluginConflict defines a plugin that conflicts with this one
type PluginConflict struct {
	Plugin string `yaml:"plugin" json:"plugin"`
	Reason string `yaml:"reason" json:"reason"`
}

// PluginAlternative defines an alternative plugin that provides similar functionality
type PluginAlternative struct {
	Plugin string `yaml:"plugin" json:"plugin"`
	Reason string `yaml:"reason" json:"reason"`
}

// PluginInstallSpec defines tool-specific installation instructions
type PluginInstallSpec struct {
	PackageManager string             `yaml:"package_manager" json:"package_manager"`
	Packages       []PluginPackage    `yaml:"packages" json:"packages"`
	Config         []PluginConfigFile `yaml:"config,omitempty" json:"config,omitempty"`
	PostInstall    string             `yaml:"post_install,omitempty" json:"post_install,omitempty"`
}

// PluginPackage defines a package to install
type PluginPackage struct {
	Name           string `yaml:"name" json:"name"`
	VersionParam   string `yaml:"version_param,omitempty" json:"version_param,omitempty"`
	InstallCommand string `yaml:"install_command" json:"install_command"`
}

// PluginConfigFile defines a configuration file to create
type PluginConfigFile struct {
	File    string `yaml:"file" json:"file"`
	Path    string `yaml:"path" json:"path"`
	Content string `yaml:"content" json:"content"`
}

// PluginConfiguration defines plugin configuration parameters
type PluginConfiguration struct {
	Parameters map[string]TemplateParameter `yaml:"parameters,omitempty" json:"parameters,omitempty"`
}

// PluginCompatibility defines tool version compatibility requirements
type PluginCompatibility struct {
	Tools map[string]ToolVersionConstraint `yaml:"tools,omitempty" json:"tools,omitempty"`
}

// ToolVersionConstraint defines minimum and maximum version requirements for a tool
type ToolVersionConstraint struct {
	MinVersion string `yaml:"min_version,omitempty" json:"min_version,omitempty"`
	MaxVersion string `yaml:"max_version,omitempty" json:"max_version,omitempty"`
}

// PluginSecurityMetadata contains security scanning and verification information
type PluginSecurityMetadata struct {
	Trusted           bool      `yaml:"trusted" json:"trusted"`
	VerifiedPublisher bool      `yaml:"verified_publisher" json:"verified_publisher"`
	CVEScanned        bool      `yaml:"cve_scanned" json:"cve_scanned"`
	LastSecurityScan  time.Time `yaml:"last_security_scan" json:"last_security_scan"`
}

// TemplatePlugins defines plugins enabled in a template
type TemplatePlugins struct {
	Enabled []EnabledPlugin `yaml:"enabled,omitempty" json:"enabled,omitempty"`
}

// EnabledPlugin defines a plugin enabled in a template with configuration
type EnabledPlugin struct {
	Name          string                 `yaml:"name" json:"name"`
	Version       string                 `yaml:"version,omitempty" json:"version,omitempty"`
	Conditional   string                 `yaml:"conditional,omitempty" json:"conditional,omitempty"` // Template condition (e.g., "{{.install_plugin}}")
	Configuration map[string]interface{} `yaml:"configuration,omitempty" json:"configuration,omitempty"`
}
