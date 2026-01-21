// Package templates provides plugin installer interfaces and registry.
package templates

import "context"

// PluginInstaller defines the interface for installing plugins on remote instances
type PluginInstaller interface {
	// Name returns the installer name (e.g., "r", "vscode", "vim")
	Name() string

	// Install installs a plugin with specified version and configuration
	Install(ctx context.Context, plugin *PluginManifest, version string, params map[string]interface{}) error

	// Uninstall removes a plugin
	Uninstall(ctx context.Context, plugin *PluginManifest) error

	// Validate checks if plugin can be installed (prerequisites met)
	Validate(ctx context.Context, plugin *PluginManifest) error

	// GetInstalledVersion returns the installed version (empty string if not installed)
	GetInstalledVersion(ctx context.Context, plugin *PluginManifest) (string, error)

	// IsCompatible checks if plugin is compatible with the current environment
	IsCompatible(ctx context.Context, plugin *PluginManifest) (bool, error)
}

// PluginInstallerRegistry manages registered plugin installers
type PluginInstallerRegistry struct {
	installers map[string]PluginInstaller
}

// NewPluginInstallerRegistry creates a new plugin installer registry
func NewPluginInstallerRegistry() *PluginInstallerRegistry {
	return &PluginInstallerRegistry{
		installers: make(map[string]PluginInstaller),
	}
}

// Register registers a plugin installer
func (r *PluginInstallerRegistry) Register(installer PluginInstaller) {
	r.installers[installer.Name()] = installer
}

// Get retrieves an installer by name
func (r *PluginInstallerRegistry) Get(name string) (PluginInstaller, bool) {
	installer, exists := r.installers[name]
	return installer, exists
}

// ListInstallers returns all registered installer names
func (r *PluginInstallerRegistry) ListInstallers() []string {
	names := make([]string, 0, len(r.installers))
	for name := range r.installers {
		names = append(names, name)
	}
	return names
}

// HasInstaller checks if an installer is registered
func (r *PluginInstallerRegistry) HasInstaller(name string) bool {
	_, exists := r.installers[name]
	return exists
}

// UnregisterAll removes all registered installers (useful for testing)
func (r *PluginInstallerRegistry) UnregisterAll() {
	r.installers = make(map[string]PluginInstaller)
}
