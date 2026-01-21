package templates

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPluginRegistry_LoadPlugin(t *testing.T) {
	registry := NewPluginRegistry()

	// Create a temporary plugin file
	tmpDir := t.TempDir()
	pluginPath := filepath.Join(tmpDir, "test-plugin.yml")

	pluginContent := `apiVersion: v1
kind: Plugin
metadata:
  name: "test-plugin"
  display_name: "Test Plugin"
  description: "A test plugin"
  category: "testing"
  version: "1.0.0"
  maintainer: "Test Team"

spec:
  provides:
    - name: "test-capability"
      type: "capability"

  installers:
    apt:
      package_manager: "apt"
      packages:
        - name: "test-package"
          install_command: "apt-get install -y test-package"

  configuration:
    parameters:
      version:
        type: "choice"
        choices: ["latest", "1.0.0"]
        default: "latest"
`

	if err := os.WriteFile(pluginPath, []byte(pluginContent), 0644); err != nil {
		t.Fatalf("Failed to create test plugin file: %v", err)
	}

	// Test loading the plugin
	if err := registry.LoadPlugin(pluginPath); err != nil {
		t.Fatalf("Failed to load plugin: %v", err)
	}

	// Verify plugin was loaded
	if !registry.HasPlugin("test-plugin") {
		t.Error("Plugin was not registered")
	}

	// Retrieve and verify plugin
	plugin, err := registry.GetPlugin("test-plugin")
	if err != nil {
		t.Fatalf("Failed to get plugin: %v", err)
	}

	if plugin.Metadata.Name != "test-plugin" {
		t.Errorf("Expected plugin name 'test-plugin', got '%s'", plugin.Metadata.Name)
	}

	if plugin.Metadata.DisplayName != "Test Plugin" {
		t.Errorf("Expected display name 'Test Plugin', got '%s'", plugin.Metadata.DisplayName)
	}
}

func TestPluginRegistry_LoadPluginsFromDirectory(t *testing.T) {
	registry := NewPluginRegistry()

	// Create temporary directory with multiple plugins
	tmpDir := t.TempDir()

	plugin1Content := `apiVersion: v1
kind: Plugin
metadata:
  name: "plugin1"
  version: "1.0.0"
spec:
  installers:
    apt:
      packages:
        - name: "pkg1"
          install_command: "apt-get install -y pkg1"
`

	plugin2Content := `apiVersion: v1
kind: Plugin
metadata:
  name: "plugin2"
  version: "1.0.0"
spec:
  installers:
    apt:
      packages:
        - name: "pkg2"
          install_command: "apt-get install -y pkg2"
`

	if err := os.WriteFile(filepath.Join(tmpDir, "plugin1.yml"), []byte(plugin1Content), 0644); err != nil {
		t.Fatalf("Failed to create plugin1: %v", err)
	}

	if err := os.WriteFile(filepath.Join(tmpDir, "plugin2.yml"), []byte(plugin2Content), 0644); err != nil {
		t.Fatalf("Failed to create plugin2: %v", err)
	}

	// Load plugins from directory
	if err := registry.LoadPluginsFromDirectory(tmpDir); err != nil {
		t.Fatalf("Failed to load plugins from directory: %v", err)
	}

	// Verify both plugins were loaded
	if !registry.HasPlugin("plugin1") {
		t.Error("plugin1 was not loaded")
	}

	if !registry.HasPlugin("plugin2") {
		t.Error("plugin2 was not loaded")
	}

	// List plugins
	plugins := registry.ListPlugins()
	if len(plugins) != 2 {
		t.Errorf("Expected 2 plugins, got %d", len(plugins))
	}
}

func TestPluginRegistry_GetPluginsByCategory(t *testing.T) {
	registry := NewPluginRegistry()
	tmpDir := t.TempDir()

	// Create plugins in different categories
	devPluginContent := `apiVersion: v1
kind: Plugin
metadata:
  name: "dev-plugin"
  category: "development"
  version: "1.0.0"
spec:
  installers:
    apt:
      packages:
        - name: "dev-pkg"
          install_command: "apt-get install -y dev-pkg"
`

	infraPluginContent := `apiVersion: v1
kind: Plugin
metadata:
  name: "infra-plugin"
  category: "infrastructure"
  version: "1.0.0"
spec:
  installers:
    apt:
      packages:
        - name: "infra-pkg"
          install_command: "apt-get install -y infra-pkg"
`

	if err := os.WriteFile(filepath.Join(tmpDir, "dev.yml"), []byte(devPluginContent), 0644); err != nil {
		t.Fatalf("Failed to create dev plugin: %v", err)
	}

	if err := os.WriteFile(filepath.Join(tmpDir, "infra.yml"), []byte(infraPluginContent), 0644); err != nil {
		t.Fatalf("Failed to create infra plugin: %v", err)
	}

	registry.LoadPluginsFromDirectory(tmpDir)

	// Test category filtering
	devPlugins := registry.GetPluginsByCategory("development")
	if len(devPlugins) != 1 {
		t.Errorf("Expected 1 development plugin, got %d", len(devPlugins))
	}

	if devPlugins[0].Metadata.Name != "dev-plugin" {
		t.Errorf("Expected 'dev-plugin', got '%s'", devPlugins[0].Metadata.Name)
	}

	infraPlugins := registry.GetPluginsByCategory("infrastructure")
	if len(infraPlugins) != 1 {
		t.Errorf("Expected 1 infrastructure plugin, got %d", len(infraPlugins))
	}
}

func TestPluginRegistry_ResolveDependencies(t *testing.T) {
	registry := NewPluginRegistry()
	tmpDir := t.TempDir()

	// Create plugin with dependency
	basePluginContent := `apiVersion: v1
kind: Plugin
metadata:
  name: "base-plugin"
  version: "1.0.0"
spec:
  installers:
    apt:
      packages:
        - name: "base-pkg"
          install_command: "apt-get install -y base-pkg"
`

	dependentPluginContent := `apiVersion: v1
kind: Plugin
metadata:
  name: "dependent-plugin"
  version: "1.0.0"
spec:
  dependencies:
    required:
      - plugin: "base-plugin"
        version: ">=1.0.0"
  installers:
    apt:
      packages:
        - name: "dependent-pkg"
          install_command: "apt-get install -y dependent-pkg"
`

	if err := os.WriteFile(filepath.Join(tmpDir, "base.yml"), []byte(basePluginContent), 0644); err != nil {
		t.Fatalf("Failed to create base plugin: %v", err)
	}

	if err := os.WriteFile(filepath.Join(tmpDir, "dependent.yml"), []byte(dependentPluginContent), 0644); err != nil {
		t.Fatalf("Failed to create dependent plugin: %v", err)
	}

	registry.LoadPluginsFromDirectory(tmpDir)

	// Resolve dependencies
	dependent, _ := registry.GetPlugin("dependent-plugin")
	resolved, err := registry.ResolveDependencies(dependent)
	if err != nil {
		t.Fatalf("Failed to resolve dependencies: %v", err)
	}

	// Should have both base and dependent
	if len(resolved) != 2 {
		t.Errorf("Expected 2 plugins (base + dependent), got %d", len(resolved))
	}

	// Base should be first (topological order)
	if resolved[0].Metadata.Name != "base-plugin" {
		t.Errorf("Expected base-plugin first, got '%s'", resolved[0].Metadata.Name)
	}

	if resolved[1].Metadata.Name != "dependent-plugin" {
		t.Errorf("Expected dependent-plugin second, got '%s'", resolved[1].Metadata.Name)
	}
}

func TestPluginRegistry_CheckConflicts(t *testing.T) {
	registry := NewPluginRegistry()
	tmpDir := t.TempDir()

	// Create conflicting plugins
	plugin1Content := `apiVersion: v1
kind: Plugin
metadata:
  name: "plugin1"
  version: "1.0.0"
spec:
  conflicts:
    - plugin: "plugin2"
      reason: "Cannot coexist with plugin2"
  installers:
    apt:
      packages:
        - name: "pkg1"
          install_command: "apt-get install -y pkg1"
`

	plugin2Content := `apiVersion: v1
kind: Plugin
metadata:
  name: "plugin2"
  version: "1.0.0"
spec:
  installers:
    apt:
      packages:
        - name: "pkg2"
          install_command: "apt-get install -y pkg2"
`

	if err := os.WriteFile(filepath.Join(tmpDir, "plugin1.yml"), []byte(plugin1Content), 0644); err != nil {
		t.Fatalf("Failed to create plugin1: %v", err)
	}

	if err := os.WriteFile(filepath.Join(tmpDir, "plugin2.yml"), []byte(plugin2Content), 0644); err != nil {
		t.Fatalf("Failed to create plugin2: %v", err)
	}

	registry.LoadPluginsFromDirectory(tmpDir)

	// Test conflict detection
	plugin1, _ := registry.GetPlugin("plugin1")
	plugin2, _ := registry.GetPlugin("plugin2")

	plugins := []*PluginManifest{plugin1, plugin2}
	err := registry.CheckConflicts(plugins)

	if err == nil {
		t.Error("Expected conflict error, got nil")
	}
}

func TestPluginRegistry_VersionSatisfies(t *testing.T) {
	registry := NewPluginRegistry()

	tests := []struct {
		version    string
		constraint string
		expected   bool
	}{
		{"1.0.0", ">=1.0.0", true},
		{"1.0.0", ">1.0.0", false},
		{"1.1.0", ">1.0.0", true},
		{"1.0.0", "<=1.0.0", true},
		{"1.0.0", "<1.0.0", false},
		{"0.9.0", "<1.0.0", true},
		{"1.0.0", "1.0.0", true},
		{"1.0.0", "1.0.1", false},
		{"1.0.0", "", true},
		{"1.0.0", "latest", true},
	}

	for _, tt := range tests {
		result := registry.versionSatisfies(tt.version, tt.constraint)
		if result != tt.expected {
			t.Errorf("versionSatisfies(%s, %s) = %v, expected %v",
				tt.version, tt.constraint, result, tt.expected)
		}
	}
}

func TestPluginRegistry_CompareVersions(t *testing.T) {
	registry := NewPluginRegistry()

	tests := []struct {
		v1       string
		v2       string
		expected int
	}{
		{"1.0.0", "1.0.0", 0},
		{"1.0.0", "1.0.1", -1},
		{"1.0.1", "1.0.0", 1},
		{"1.0.0", "2.0.0", -1},
		{"2.0.0", "1.0.0", 1},
		{"1.0.0", "1.1.0", -1},
		{"1.1.0", "1.0.0", 1},
	}

	for _, tt := range tests {
		result := registry.compareVersions(tt.v1, tt.v2)
		if result != tt.expected {
			t.Errorf("compareVersions(%s, %s) = %d, expected %d",
				tt.v1, tt.v2, result, tt.expected)
		}
	}
}
