// Package templates provides plugin registry for discovery and management.
package templates

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// PluginRegistry manages plugin manifests and provides discovery
type PluginRegistry struct {
	plugins     map[string]*PluginManifest
	pluginPaths []string
}

// NewPluginRegistry creates a new plugin registry
func NewPluginRegistry() *PluginRegistry {
	return &PluginRegistry{
		plugins:     make(map[string]*PluginManifest),
		pluginPaths: make([]string, 0),
	}
}

// LoadPluginsFromDirectory loads all plugin manifests from a directory
func (r *PluginRegistry) LoadPluginsFromDirectory(dir string) error {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read plugin directory %s: %w", dir, err)
	}

	loadedCount := 0
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		ext := filepath.Ext(file.Name())
		if ext != ".yml" && ext != ".yaml" {
			continue
		}

		pluginPath := filepath.Join(dir, file.Name())
		if err := r.LoadPlugin(pluginPath); err != nil {
			// Log warning but continue loading other plugins
			fmt.Printf("Warning: failed to load plugin %s: %v\n", file.Name(), err)
			continue
		}
		loadedCount++
	}

	if loadedCount == 0 {
		return fmt.Errorf("no valid plugins found in directory %s", dir)
	}

	return nil
}

// LoadPlugin loads a single plugin manifest from a file
func (r *PluginRegistry) LoadPlugin(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read plugin file: %w", err)
	}

	var manifest PluginManifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return fmt.Errorf("failed to parse plugin manifest: %w", err)
	}

	// Validate manifest
	if err := r.validateManifest(&manifest); err != nil {
		return fmt.Errorf("invalid plugin manifest: %w", err)
	}

	// Store plugin
	r.plugins[manifest.Metadata.Name] = &manifest
	r.pluginPaths = append(r.pluginPaths, path)

	return nil
}

// GetPlugin retrieves a plugin by name
func (r *PluginRegistry) GetPlugin(name string) (*PluginManifest, error) {
	plugin, exists := r.plugins[name]
	if !exists {
		return nil, fmt.Errorf("plugin %s not found", name)
	}
	return plugin, nil
}

// ListPlugins returns all registered plugins
func (r *PluginRegistry) ListPlugins() []*PluginManifest {
	plugins := make([]*PluginManifest, 0, len(r.plugins))
	for _, plugin := range r.plugins {
		plugins = append(plugins, plugin)
	}
	return plugins
}

// GetPluginsByCategory returns plugins in a specific category
func (r *PluginRegistry) GetPluginsByCategory(category string) []*PluginManifest {
	plugins := make([]*PluginManifest, 0)
	for _, plugin := range r.plugins {
		if plugin.Metadata.Category == category {
			plugins = append(plugins, plugin)
		}
	}
	return plugins
}

// HasPlugin checks if a plugin exists
func (r *PluginRegistry) HasPlugin(name string) bool {
	_, exists := r.plugins[name]
	return exists
}

// validateManifest validates a plugin manifest structure
func (r *PluginRegistry) validateManifest(manifest *PluginManifest) error {
	if manifest.Metadata.Name == "" {
		return fmt.Errorf("plugin name is required")
	}
	if manifest.Metadata.Version == "" {
		return fmt.Errorf("plugin version is required")
	}
	if manifest.APIVersion == "" {
		return fmt.Errorf("apiVersion is required")
	}
	if manifest.Kind == "" {
		return fmt.Errorf("kind is required")
	}
	if len(manifest.Spec.Installers) == 0 {
		return fmt.Errorf("at least one installer is required")
	}

	// Validate installer specs
	for toolName, installer := range manifest.Spec.Installers {
		if len(installer.Packages) == 0 {
			return fmt.Errorf("installer %s must have at least one package", toolName)
		}
		for _, pkg := range installer.Packages {
			if pkg.Name == "" {
				return fmt.Errorf("package name is required in installer %s", toolName)
			}
			if pkg.InstallCommand == "" {
				return fmt.Errorf("install_command is required for package %s in installer %s", pkg.Name, toolName)
			}
		}
	}

	return nil
}

// ResolveDependencies resolves all dependencies for a plugin (recursive)
func (r *PluginRegistry) ResolveDependencies(plugin *PluginManifest) ([]*PluginManifest, error) {
	resolved := make([]*PluginManifest, 0)
	visited := make(map[string]bool)

	if err := r.resolveDependenciesRecursive(plugin, &resolved, visited); err != nil {
		return nil, err
	}

	return resolved, nil
}

// resolveDependenciesRecursive resolves dependencies recursively with cycle detection
func (r *PluginRegistry) resolveDependenciesRecursive(plugin *PluginManifest, resolved *[]*PluginManifest, visited map[string]bool) error {
	// Check for cycles
	if visited[plugin.Metadata.Name] {
		return nil // Already processed
	}

	visited[plugin.Metadata.Name] = true

	// Resolve required dependencies first
	for _, dep := range plugin.Spec.Dependencies.Required {
		depPlugin, err := r.GetPlugin(dep.Plugin)
		if err != nil {
			return fmt.Errorf("required dependency %s not found: %w", dep.Plugin, err)
		}

		// Check version constraint
		if !r.versionSatisfies(depPlugin.Metadata.Version, dep.Version) {
			return fmt.Errorf("dependency %s version %s does not satisfy constraint %s",
				dep.Plugin, depPlugin.Metadata.Version, dep.Version)
		}

		// Recursively resolve dependencies
		if err := r.resolveDependenciesRecursive(depPlugin, resolved, visited); err != nil {
			return err
		}
	}

	// Add this plugin to resolved list
	*resolved = append(*resolved, plugin)

	return nil
}

// CheckConflicts checks for conflicting plugins in a list
func (r *PluginRegistry) CheckConflicts(plugins []*PluginManifest) error {
	installed := make(map[string]*PluginManifest)

	for _, plugin := range plugins {
		// Check if this plugin conflicts with any already installed
		for _, conflict := range plugin.Spec.Conflicts {
			if conflictingPlugin, exists := installed[conflict.Plugin]; exists {
				return fmt.Errorf("plugin %s conflicts with %s: %s",
					plugin.Metadata.Name, conflictingPlugin.Metadata.Name, conflict.Reason)
			}
		}

		// Check if any already installed plugin conflicts with this one
		for _, installedPlugin := range installed {
			for _, conflict := range installedPlugin.Spec.Conflicts {
				if conflict.Plugin == plugin.Metadata.Name {
					return fmt.Errorf("plugin %s conflicts with %s: %s",
						installedPlugin.Metadata.Name, plugin.Metadata.Name, conflict.Reason)
				}
			}
		}

		installed[plugin.Metadata.Name] = plugin
	}

	return nil
}

// versionSatisfies checks if a version satisfies a constraint
// For now, implements basic version comparison
// TODO: Implement full semantic versioning (>=, <=, ~, ^, etc.)
func (r *PluginRegistry) versionSatisfies(version, constraint string) bool {
	// Handle "latest" and empty constraints
	if constraint == "" || constraint == "latest" {
		return true
	}

	// Handle exact version match
	if !strings.ContainsAny(constraint, "><=~^") {
		return version == constraint
	}

	// Handle >= constraint (most common)
	if strings.HasPrefix(constraint, ">=") {
		requiredVersion := strings.TrimPrefix(constraint, ">=")
		return r.compareVersions(version, requiredVersion) >= 0
	}

	// Handle > constraint
	if strings.HasPrefix(constraint, ">") {
		requiredVersion := strings.TrimPrefix(constraint, ">")
		return r.compareVersions(version, requiredVersion) > 0
	}

	// Handle <= constraint
	if strings.HasPrefix(constraint, "<=") {
		requiredVersion := strings.TrimPrefix(constraint, "<=")
		return r.compareVersions(version, requiredVersion) <= 0
	}

	// Handle < constraint
	if strings.HasPrefix(constraint, "<") {
		requiredVersion := strings.TrimPrefix(constraint, "<")
		return r.compareVersions(version, requiredVersion) < 0
	}

	// Default: assume satisfied (be permissive)
	return true
}

// compareVersions compares two version strings
// Returns: -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2
// Simple implementation for now - compares version components numerically
func (r *PluginRegistry) compareVersions(v1, v2 string) int {
	// Split versions into components
	parts1 := strings.Split(strings.TrimSpace(v1), ".")
	parts2 := strings.Split(strings.TrimSpace(v2), ".")

	// Compare component by component
	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		var n1, n2 int

		if i < len(parts1) {
			fmt.Sscanf(parts1[i], "%d", &n1)
		}
		if i < len(parts2) {
			fmt.Sscanf(parts2[i], "%d", &n2)
		}

		if n1 < n2 {
			return -1
		}
		if n1 > n2 {
			return 1
		}
	}

	return 0 // Equal
}

// Clear removes all plugins from the registry (useful for testing)
func (r *PluginRegistry) Clear() {
	r.plugins = make(map[string]*PluginManifest)
	r.pluginPaths = make([]string, 0)
}
