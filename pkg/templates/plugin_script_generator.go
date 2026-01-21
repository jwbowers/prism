// Package templates provides plugin script generation for template application.
package templates

import (
	"context"
	"fmt"
	"strings"
	"text/template"
)

// PluginScriptGenerator generates installation scripts for plugins
type PluginScriptGenerator struct {
	registry          *PluginRegistry
	installerRegistry *PluginInstallerRegistry
}

// NewPluginScriptGenerator creates a new plugin script generator
func NewPluginScriptGenerator(registry *PluginRegistry, installerRegistry *PluginInstallerRegistry) *PluginScriptGenerator {
	return &PluginScriptGenerator{
		registry:          registry,
		installerRegistry: installerRegistry,
	}
}

// GeneratePluginInstallScript generates a plugin installation script for a template and tool
func (g *PluginScriptGenerator) GeneratePluginInstallScript(ctx context.Context, tmpl *Template, tool string, params TemplateParameterValues) (string, error) {
	// Skip if no plugins defined
	if len(tmpl.Plugins.Enabled) == 0 {
		return "", nil
	}

	var script strings.Builder
	script.WriteString("# Plugin Installation\n")
	script.WriteString("echo '🔌 Installing plugins...'\n\n")

	// Track installed plugins to avoid duplicates
	installed := make(map[string]bool)

	// Process each enabled plugin
	for _, enabledPlugin := range tmpl.Plugins.Enabled {
		// Check conditional
		if enabledPlugin.Conditional != "" {
			shouldInstall, err := g.evaluateConditional(enabledPlugin.Conditional, params)
			if err != nil {
				return "", fmt.Errorf("failed to evaluate conditional for plugin %s: %w", enabledPlugin.Name, err)
			}
			if !shouldInstall {
				continue
			}
		}

		// Check if already installed (dependencies might have added it)
		if installed[enabledPlugin.Name] {
			continue
		}

		// Load plugin manifest
		manifest, err := g.registry.GetPlugin(enabledPlugin.Name)
		if err != nil {
			return "", fmt.Errorf("plugin %s not found: %w", enabledPlugin.Name, err)
		}

		// Check tool support
		installer, toolExists := manifest.Spec.Installers[tool]
		if !toolExists {
			// Skip silently - plugin doesn't support this tool
			continue
		}

		// Resolve version
		version := g.resolveVersion(enabledPlugin.Version, params)

		// Merge configuration parameters
		mergedParams := g.mergeParameters(manifest.Spec.Configuration.Parameters, enabledPlugin.Configuration, params)
		mergedParams["version"] = version

		// Generate installation commands
		script.WriteString(fmt.Sprintf("# Installing: %s (version: %s)\n", manifest.Metadata.DisplayName, version))
		script.WriteString(fmt.Sprintf("echo 'Installing %s...'\n", manifest.Metadata.DisplayName))

		// Generate package installation commands
		for _, pkg := range installer.Packages {
			installCmd, err := g.substituteParameters(pkg.InstallCommand, mergedParams)
			if err != nil {
				return "", fmt.Errorf("failed to substitute parameters for package %s: %w", pkg.Name, err)
			}
			script.WriteString(installCmd)
			script.WriteString("\n")
		}

		// Generate configuration file creation
		for _, config := range installer.Config {
			content, err := g.substituteParameters(config.Content, mergedParams)
			if err != nil {
				return "", fmt.Errorf("failed to substitute parameters in config %s: %w", config.File, err)
			}

			script.WriteString(fmt.Sprintf("# Create config: %s\n", config.File))
			script.WriteString(fmt.Sprintf("mkdir -p $(dirname %s)\n", config.Path))
			script.WriteString(fmt.Sprintf("cat > %s <<'EOF'\n", config.Path))
			script.WriteString(content)
			script.WriteString("\nEOF\n")
		}

		// Generate post-install script
		if installer.PostInstall != "" {
			postInstall, err := g.substituteParameters(installer.PostInstall, mergedParams)
			if err != nil {
				return "", fmt.Errorf("failed to substitute parameters in post-install script: %w", err)
			}
			script.WriteString(fmt.Sprintf("# Post-install: %s\n", manifest.Metadata.DisplayName))
			script.WriteString(postInstall)
			script.WriteString("\n")
		}

		script.WriteString("\n")
		installed[enabledPlugin.Name] = true
	}

	return script.String(), nil
}

// evaluateConditional evaluates a conditional expression using template parameters
func (g *PluginScriptGenerator) evaluateConditional(conditional string, params TemplateParameterValues) (bool, error) {
	// Parse as Go template
	tmpl, err := template.New("conditional").Parse(conditional)
	if err != nil {
		return false, fmt.Errorf("failed to parse conditional: %w", err)
	}

	// Execute template to evaluate condition
	var result strings.Builder
	if err := tmpl.Execute(&result, params); err != nil {
		return false, fmt.Errorf("failed to execute conditional: %w", err)
	}

	// Check if result is truthy
	output := strings.TrimSpace(result.String())
	return output == "true" || output == "1" || output == "yes", nil
}

// resolveVersion resolves a version string, handling template variables
func (g *PluginScriptGenerator) resolveVersion(version string, params TemplateParameterValues) string {
	// If version is empty, default to "latest"
	if version == "" {
		return "latest"
	}

	// If version contains template variables, substitute them
	if strings.Contains(version, "{{") && strings.Contains(version, "}}") {
		resolved, err := g.substituteParameters(version, params)
		if err != nil {
			// Fallback to original version if substitution fails
			return version
		}
		return resolved
	}

	return version
}

// mergeParameters merges plugin defaults, enabled plugin config, and template parameters
func (g *PluginScriptGenerator) mergeParameters(pluginParams map[string]TemplateParameter, pluginConfig map[string]interface{}, templateParams TemplateParameterValues) map[string]interface{} {
	merged := make(map[string]interface{})

	// Start with plugin defaults
	for key, param := range pluginParams {
		merged[key] = param.Default
	}

	// Override with plugin configuration from template
	for key, value := range pluginConfig {
		merged[key] = value
	}

	// Override with runtime template parameters
	for key, value := range templateParams {
		merged[key] = value
	}

	return merged
}

// substituteParameters performs template variable substitution
func (g *PluginScriptGenerator) substituteParameters(text string, params map[string]interface{}) (string, error) {
	tmpl, err := template.New("substitute").Parse(text)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var result strings.Builder
	if err := tmpl.Execute(&result, params); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return result.String(), nil
}

// GeneratePluginInstallScriptForTool generates plugin installation script for a specific tool
// This is a convenience method that automatically determines the tool from the package manager
func (g *PluginScriptGenerator) GeneratePluginInstallScriptForTool(ctx context.Context, tmpl *Template, packageManager PackageManagerType, params TemplateParameterValues) (string, error) {
	// Map package manager to tool identifier
	tool := g.packageManagerToTool(packageManager)
	return g.GeneratePluginInstallScript(ctx, tmpl, tool, params)
}

// packageManagerToTool maps a package manager to a tool identifier for plugins
func (g *PluginScriptGenerator) packageManagerToTool(pm PackageManagerType) string {
	// For now, we use generic tool identifiers
	// In the future, this could be more sophisticated
	switch pm {
	case PackageManagerConda:
		return "conda"
	case PackageManagerApt:
		return "apt"
	case PackageManagerDnf:
		return "dnf"
	case PackageManagerSpack:
		return "spack"
	default:
		return "system"
	}
}

// ValidatePlugins validates all plugins in a template
func (g *PluginScriptGenerator) ValidatePlugins(tmpl *Template) []ValidationResult {
	validator := NewPluginValidationRule(g.registry)
	return validator.Validate(tmpl)
}
