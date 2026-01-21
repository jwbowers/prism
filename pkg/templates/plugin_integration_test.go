package templates

import (
	"strings"
	"testing"
)

func TestPluginScriptGeneration(t *testing.T) {
	// Create plugin registry and load Git plugin
	pluginRegistry := NewPluginRegistry()
	if err := pluginRegistry.LoadPluginsFromDirectory("../../plugins"); err != nil {
		t.Fatalf("Failed to load plugins: %v", err)
	}

	// Verify Git plugin loaded
	if !pluginRegistry.HasPlugin("git") {
		t.Fatal("Git plugin not loaded")
	}

	// Create test template with Git plugin enabled
	tmpl := &Template{
		Name:           "Test Template with Plugins",
		Base:           "ubuntu-22.04",
		PackageManager: "apt",
		Packages: PackageDefinitions{
			System: []string{"curl", "wget"},
		},
		Plugins: TemplatePlugins{
			Enabled: []EnabledPlugin{
				{
					Name:    "git",
					Version: "latest",
					Configuration: map[string]interface{}{
						"enable_lfs": true,
					},
				},
			},
		},
	}

	// Create script generator with plugin support
	scriptGenerator := NewScriptGeneratorWithPlugins(pluginRegistry)

	// Generate installation script
	script, err := scriptGenerator.GenerateScript(tmpl, PackageManagerApt)
	if err != nil {
		t.Fatalf("Failed to generate script: %v", err)
	}

	// Verify script contains plugin installation
	if !strings.Contains(script, "🔌 Installing plugins...") {
		t.Error("Script does not contain plugin installation section")
	}

	// Verify script contains Git installation command
	if !strings.Contains(script, "apt-get install -y git") {
		t.Error("Script does not contain Git installation command")
	}

	// Verify script contains Git LFS installation (since enable_lfs is true)
	if !strings.Contains(script, "git-lfs") {
		t.Error("Script does not contain Git LFS installation")
	}

	t.Logf("Generated script with %d lines", len(strings.Split(script, "\n")))
}

func TestPluginScriptGeneration_NoPlugins(t *testing.T) {
	// Create plugin registry
	pluginRegistry := NewPluginRegistry()
	if err := pluginRegistry.LoadPluginsFromDirectory("../../plugins"); err != nil {
		t.Fatalf("Failed to load plugins: %v", err)
	}

	// Create test template WITHOUT plugins
	tmpl := &Template{
		Name:           "Test Template without Plugins",
		Base:           "ubuntu-22.04",
		PackageManager: "apt",
		Packages: PackageDefinitions{
			System: []string{"curl", "wget"},
		},
		Plugins: TemplatePlugins{
			Enabled: []EnabledPlugin{}, // No plugins
		},
	}

	// Create script generator with plugin support
	scriptGenerator := NewScriptGeneratorWithPlugins(pluginRegistry)

	// Generate installation script
	script, err := scriptGenerator.GenerateScript(tmpl, PackageManagerApt)
	if err != nil {
		t.Fatalf("Failed to generate script: %v", err)
	}

	// Verify script does NOT contain plugin installation section
	if strings.Contains(script, "🔌 Installing plugins...") {
		t.Error("Script should not contain plugin installation section when no plugins enabled")
	}
}

func TestPluginScriptGeneration_WithoutPluginRegistry(t *testing.T) {
	// Create test template with plugins enabled
	tmpl := &Template{
		Name:           "Test Template",
		Base:           "ubuntu-22.04",
		PackageManager: "apt",
		Packages: PackageDefinitions{
			System: []string{"curl"},
		},
		Plugins: TemplatePlugins{
			Enabled: []EnabledPlugin{
				{
					Name:    "git",
					Version: "latest",
				},
			},
		},
	}

	// Create script generator WITHOUT plugin support (backward compatibility)
	scriptGenerator := NewScriptGenerator()

	// Generate installation script
	script, err := scriptGenerator.GenerateScript(tmpl, PackageManagerApt)
	if err != nil {
		t.Fatalf("Failed to generate script: %v", err)
	}

	// Verify script generates successfully but does not contain plugin section
	if strings.Contains(script, "🔌 Installing plugins...") {
		t.Error("Script should not contain plugin section when plugin registry not provided")
	}

	// Verify basic script content still present
	if !strings.Contains(script, "curl") {
		t.Error("Script does not contain system packages")
	}
}
