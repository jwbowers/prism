package templates

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPluginValidationRule_ValidatePluginExists(t *testing.T) {
	registry := NewPluginRegistry()
	tmpDir := t.TempDir()

	// Create a test plugin
	pluginContent := `apiVersion: v1
kind: Plugin
metadata:
  name: "test-plugin"
  version: "1.0.0"
spec:
  installers:
    apt:
      packages:
        - name: "test-pkg"
          install_command: "apt-get install -y test-pkg"
  configuration:
    parameters:
      version:
        type: "choice"
        choices: ["latest", "1.0.0"]
        default: "latest"
`

	if err := os.WriteFile(filepath.Join(tmpDir, "test-plugin.yml"), []byte(pluginContent), 0644); err != nil {
		t.Fatalf("Failed to create test plugin: %v", err)
	}

	registry.LoadPluginsFromDirectory(tmpDir)

	validator := NewPluginValidationRule(registry)

	// Test with valid plugin
	tmpl := &Template{
		Name: "Test Template",
		Plugins: TemplatePlugins{
			Enabled: []EnabledPlugin{
				{
					Name:    "test-plugin",
					Version: "latest",
				},
			},
		},
	}

	results := validator.Validate(tmpl)
	if len(results) > 0 {
		t.Errorf("Expected no validation errors, got %d: %v", len(results), results)
	}

	// Test with non-existent plugin
	tmplInvalid := &Template{
		Name: "Test Template",
		Plugins: TemplatePlugins{
			Enabled: []EnabledPlugin{
				{
					Name:    "nonexistent-plugin",
					Version: "latest",
				},
			},
		},
	}

	results = validator.Validate(tmplInvalid)
	if len(results) == 0 {
		t.Error("Expected validation error for non-existent plugin")
	}

	if results[0].Level != ValidationError {
		t.Errorf("Expected ValidationError level, got %v", results[0].Level)
	}
}

func TestPluginValidationRule_ValidateVersion(t *testing.T) {
	registry := NewPluginRegistry()
	tmpDir := t.TempDir()

	pluginContent := `apiVersion: v1
kind: Plugin
metadata:
  name: "versioned-plugin"
  version: "1.0.0"
spec:
  installers:
    apt:
      packages:
        - name: "pkg"
          install_command: "apt-get install -y pkg"
  configuration:
    parameters:
      version:
        type: "choice"
        choices: ["latest", "1.0.0", "0.9.0"]
        default: "latest"
`

	if err := os.WriteFile(filepath.Join(tmpDir, "versioned-plugin.yml"), []byte(pluginContent), 0644); err != nil {
		t.Fatalf("Failed to create plugin: %v", err)
	}

	registry.LoadPluginsFromDirectory(tmpDir)
	validator := NewPluginValidationRule(registry)

	// Test with valid version
	tmpl := &Template{
		Name: "Test Template",
		Plugins: TemplatePlugins{
			Enabled: []EnabledPlugin{
				{
					Name:    "versioned-plugin",
					Version: "1.0.0",
				},
			},
		},
	}

	results := validator.Validate(tmpl)
	if len(results) > 0 {
		for _, result := range results {
			if result.Level == ValidationError {
				t.Errorf("Unexpected validation error: %s", result.Message)
			}
		}
	}

	// Test with invalid version
	tmplInvalid := &Template{
		Name: "Test Template",
		Plugins: TemplatePlugins{
			Enabled: []EnabledPlugin{
				{
					Name:    "versioned-plugin",
					Version: "999.999.999",
				},
			},
		},
	}

	results = validator.Validate(tmplInvalid)
	foundWarning := false
	for _, result := range results {
		if result.Level == ValidationWarning && result.Field == "plugins.enabled[0].version" {
			foundWarning = true
			break
		}
	}

	if !foundWarning {
		t.Error("Expected validation warning for invalid version")
	}
}

func TestPluginValidationRule_ValidateTemplateVariableVersion(t *testing.T) {
	registry := NewPluginRegistry()
	tmpDir := t.TempDir()

	pluginContent := `apiVersion: v1
kind: Plugin
metadata:
  name: "template-var-plugin"
  version: "1.0.0"
spec:
  installers:
    apt:
      packages:
        - name: "pkg"
          install_command: "apt-get install -y pkg"
  configuration:
    parameters:
      version:
        type: "choice"
        choices: ["latest", "1.0.0"]
        default: "latest"
`

	if err := os.WriteFile(filepath.Join(tmpDir, "template-var-plugin.yml"), []byte(pluginContent), 0644); err != nil {
		t.Fatalf("Failed to create plugin: %v", err)
	}

	registry.LoadPluginsFromDirectory(tmpDir)
	validator := NewPluginValidationRule(registry)

	// Test with template variable (should be valid)
	tmpl := &Template{
		Name: "Test Template",
		Plugins: TemplatePlugins{
			Enabled: []EnabledPlugin{
				{
					Name:    "template-var-plugin",
					Version: "{{.plugin_version}}",
				},
			},
		},
	}

	results := validator.Validate(tmpl)
	for _, result := range results {
		if result.Level == ValidationError {
			t.Errorf("Template variables should be valid, got error: %s", result.Message)
		}
	}
}

func TestPluginValidationRule_ValidateConfiguration(t *testing.T) {
	registry := NewPluginRegistry()
	tmpDir := t.TempDir()

	pluginContent := `apiVersion: v1
kind: Plugin
metadata:
  name: "config-plugin"
  version: "1.0.0"
spec:
  installers:
    apt:
      packages:
        - name: "pkg"
          install_command: "apt-get install -y pkg"
  configuration:
    parameters:
      param1:
        type: "string"
        default: "value1"
      param2:
        type: "integer"
        default: 10
`

	if err := os.WriteFile(filepath.Join(tmpDir, "config-plugin.yml"), []byte(pluginContent), 0644); err != nil {
		t.Fatalf("Failed to create plugin: %v", err)
	}

	registry.LoadPluginsFromDirectory(tmpDir)
	validator := NewPluginValidationRule(registry)

	// Test with valid configuration
	tmpl := &Template{
		Name: "Test Template",
		Plugins: TemplatePlugins{
			Enabled: []EnabledPlugin{
				{
					Name:    "config-plugin",
					Version: "latest",
					Configuration: map[string]interface{}{
						"param1": "custom_value",
					},
				},
			},
		},
	}

	results := validator.Validate(tmpl)
	if len(results) > 0 {
		for _, result := range results {
			if result.Level == ValidationError {
				t.Errorf("Unexpected validation error: %s", result.Message)
			}
		}
	}

	// Test with unknown configuration parameter
	tmplInvalid := &Template{
		Name: "Test Template",
		Plugins: TemplatePlugins{
			Enabled: []EnabledPlugin{
				{
					Name:    "config-plugin",
					Version: "latest",
					Configuration: map[string]interface{}{
						"unknown_param": "value",
					},
				},
			},
		},
	}

	results = validator.Validate(tmplInvalid)
	foundWarning := false
	for _, result := range results {
		if result.Level == ValidationWarning && result.Field == "plugins.enabled[0].configuration.unknown_param" {
			foundWarning = true
			break
		}
	}

	if !foundWarning {
		t.Error("Expected validation warning for unknown configuration parameter")
	}
}

func TestPluginValidationRule_ValidateDependencies(t *testing.T) {
	registry := NewPluginRegistry()
	tmpDir := t.TempDir()

	// Create base plugin
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

	// Create dependent plugin
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
	validator := NewPluginValidationRule(registry)

	// Test with missing dependency
	tmplMissingDep := &Template{
		Name: "Test Template",
		Plugins: TemplatePlugins{
			Enabled: []EnabledPlugin{
				{
					Name:    "dependent-plugin",
					Version: "latest",
				},
			},
		},
	}

	results := validator.Validate(tmplMissingDep)
	foundError := false
	for _, result := range results {
		if result.Level == ValidationError && result.Field == "plugins.enabled[dependent-plugin]" {
			foundError = true
			break
		}
	}

	if !foundError {
		t.Error("Expected validation error for missing dependency")
	}

	// Test with satisfied dependency
	tmplValid := &Template{
		Name: "Test Template",
		Plugins: TemplatePlugins{
			Enabled: []EnabledPlugin{
				{
					Name:    "base-plugin",
					Version: "latest",
				},
				{
					Name:    "dependent-plugin",
					Version: "latest",
				},
			},
		},
	}

	results = validator.Validate(tmplValid)
	for _, result := range results {
		if result.Level == ValidationError {
			t.Errorf("Unexpected validation error with satisfied dependencies: %s", result.Message)
		}
	}
}

func TestPluginValidationRule_ValidateConflicts(t *testing.T) {
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
      reason: "Cannot coexist"
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
	validator := NewPluginValidationRule(registry)

	// Test with conflicting plugins
	tmpl := &Template{
		Name: "Test Template",
		Plugins: TemplatePlugins{
			Enabled: []EnabledPlugin{
				{
					Name:    "plugin1",
					Version: "latest",
				},
				{
					Name:    "plugin2",
					Version: "latest",
				},
			},
		},
	}

	results := validator.Validate(tmpl)
	foundError := false
	for _, result := range results {
		if result.Level == ValidationError && result.Field == "plugins.enabled" {
			foundError = true
			break
		}
	}

	if !foundError {
		t.Error("Expected validation error for conflicting plugins")
	}
}
