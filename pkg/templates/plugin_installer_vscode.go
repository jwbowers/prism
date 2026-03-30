// Package templates provides VS Code extension plugin installer implementation.
package templates

import (
	"context"
	"fmt"
	"strings"
	"text/template"
)

// VSCodeExtensionInstaller installs VS Code extensions on remote instances
type VSCodeExtensionInstaller struct {
	executor     RemoteExecutor
	instanceName string
}

// NewVSCodeExtensionInstaller creates a new VS Code extension installer
func NewVSCodeExtensionInstaller(executor RemoteExecutor, instanceName string) *VSCodeExtensionInstaller {
	return &VSCodeExtensionInstaller{
		executor:     executor,
		instanceName: instanceName,
	}
}

// Name returns the installer name
func (v *VSCodeExtensionInstaller) Name() string {
	return "vscode"
}

// Install installs a VS Code extension with specified version and configuration
func (v *VSCodeExtensionInstaller) Install(ctx context.Context, plugin *PluginManifest, version string, params map[string]interface{}) error {
	// Get VS Code-specific installer from plugin manifest
	vscodeInstaller, exists := plugin.Spec.Installers["vscode"]
	if !exists {
		return fmt.Errorf("plugin %s does not support VS Code installer", plugin.Metadata.Name)
	}

	// Merge plugin configuration parameters with provided parameters
	mergedParams := v.mergeParameters(plugin.Spec.Configuration.Parameters, params)
	mergedParams["version"] = version

	// Install each extension
	for _, pkg := range vscodeInstaller.Packages {
		// Substitute parameters in install command
		installCmd, err := v.substituteParameters(pkg.InstallCommand, mergedParams)
		if err != nil {
			return fmt.Errorf("failed to substitute parameters for extension %s: %w", pkg.Name, err)
		}

		// Execute installation command
		result, err := v.executor.Execute(ctx, v.instanceName, installCmd)
		if err != nil {
			return fmt.Errorf("failed to execute install command for extension %s: %w", pkg.Name, err)
		}

		if result.ExitCode != 0 {
			return fmt.Errorf("install command failed for extension %s (exit code %d): %s",
				pkg.Name, result.ExitCode, result.Stderr)
		}
	}

	// Create configuration files if specified
	for _, config := range vscodeInstaller.Config {
		// Substitute parameters in config content
		content, err := v.substituteParameters(config.Content, mergedParams)
		if err != nil {
			return fmt.Errorf("failed to substitute parameters in config %s: %w", config.File, err)
		}

		// Create directory and write config file
		writeCmd := fmt.Sprintf("mkdir -p $(dirname %s) && cat > %s <<'EOF'\n%s\nEOF",
			config.Path, config.Path, content)

		result, err := v.executor.Execute(ctx, v.instanceName, writeCmd)
		if err != nil {
			return fmt.Errorf("failed to create config file %s: %w", config.File, err)
		}

		if result.ExitCode != 0 {
			return fmt.Errorf("failed to write config %s (exit code %d): %s",
				config.File, result.ExitCode, result.Stderr)
		}
	}

	// Execute post-install script if specified
	if vscodeInstaller.PostInstall != "" {
		postInstall, err := v.substituteParameters(vscodeInstaller.PostInstall, mergedParams)
		if err != nil {
			return fmt.Errorf("failed to substitute parameters in post-install script: %w", err)
		}

		result, err := v.executor.ExecuteScript(ctx, v.instanceName, postInstall)
		if err != nil {
			return fmt.Errorf("failed to execute post-install script: %w", err)
		}

		if result.ExitCode != 0 {
			return fmt.Errorf("post-install script failed (exit code %d): %s",
				result.ExitCode, result.Stderr)
		}
	}

	return nil
}

// Uninstall removes a VS Code extension
func (v *VSCodeExtensionInstaller) Uninstall(ctx context.Context, plugin *PluginManifest) error {
	vscodeInstaller, exists := plugin.Spec.Installers["vscode"]
	if !exists {
		return fmt.Errorf("plugin %s does not support VS Code installer", plugin.Metadata.Name)
	}

	for _, pkg := range vscodeInstaller.Packages {
		uninstallCmd := fmt.Sprintf("code --uninstall-extension %s", pkg.Name)

		result, err := v.executor.Execute(ctx, v.instanceName, uninstallCmd)
		if err != nil {
			return fmt.Errorf("failed to uninstall extension %s: %w", pkg.Name, err)
		}

		if result.ExitCode != 0 {
			// Log warning but continue - extension might not be installed
			fmt.Printf("Warning: failed to uninstall extension %s: %s\n", pkg.Name, result.Stderr)
		}
	}

	return nil
}

// Validate checks if VS Code is available and prerequisites are met
func (v *VSCodeExtensionInstaller) Validate(ctx context.Context, plugin *PluginManifest) error {
	// Check if code command is available
	result, err := v.executor.Execute(ctx, v.instanceName, "code --version")
	if err != nil {
		return fmt.Errorf("VS Code is not installed: %w", err)
	}

	if result.ExitCode != 0 {
		return fmt.Errorf("VS Code is not available (exit code %d): %s", result.ExitCode, result.Stderr)
	}

	// Check compatibility constraints if specified
	if plugin.Spec.Compatibility.Tools != nil {
		if constraint, exists := plugin.Spec.Compatibility.Tools["vscode"]; exists {
			if installed := parseVersionFromOutput(result.Stdout); installed != "" {
				if constraint.MinVersion != "" && !versionSatisfiesConstraint(installed, ">="+constraint.MinVersion) {
					return fmt.Errorf("installed VS Code version %s is below minimum required %s",
						installed, constraint.MinVersion)
				}
				if constraint.MaxVersion != "" && !versionSatisfiesConstraint(installed, "<="+constraint.MaxVersion) {
					return fmt.Errorf("installed VS Code version %s exceeds maximum allowed %s",
						installed, constraint.MaxVersion)
				}
			}
		}
	}

	return nil
}

// GetInstalledVersion returns the installed version of an extension
func (v *VSCodeExtensionInstaller) GetInstalledVersion(ctx context.Context, plugin *PluginManifest) (string, error) {
	vscodeInstaller, exists := plugin.Spec.Installers["vscode"]
	if !exists {
		return "", fmt.Errorf("plugin %s does not support VS Code installer", plugin.Metadata.Name)
	}

	if len(vscodeInstaller.Packages) == 0 {
		return "", fmt.Errorf("no extensions defined for plugin %s", plugin.Metadata.Name)
	}

	// Check version of first extension
	pkg := vscodeInstaller.Packages[0]
	checkCmd := fmt.Sprintf("code --list-extensions --show-versions | grep -i '%s'", pkg.Name)

	result, err := v.executor.Execute(ctx, v.instanceName, checkCmd)
	if err != nil {
		return "", fmt.Errorf("failed to check version: %w", err)
	}

	if result.ExitCode != 0 {
		return "", nil // Not installed
	}

	// Parse version from output (format: "extension@version")
	output := strings.TrimSpace(result.Stdout)
	if strings.Contains(output, "@") {
		parts := strings.Split(output, "@")
		if len(parts) >= 2 {
			return parts[1], nil
		}
	}

	return "", nil
}

// IsCompatible checks if the plugin is compatible with the environment
func (v *VSCodeExtensionInstaller) IsCompatible(ctx context.Context, plugin *PluginManifest) (bool, error) {
	// Check if VS Code installer is defined
	_, exists := plugin.Spec.Installers["vscode"]
	if !exists {
		return false, nil
	}

	// Validate VS Code is available
	if err := v.Validate(ctx, plugin); err != nil {
		return false, nil
	}

	return true, nil
}

// mergeParameters merges plugin configuration parameters with provided parameters
func (v *VSCodeExtensionInstaller) mergeParameters(pluginParams map[string]TemplateParameter, providedParams map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})

	// Start with plugin defaults
	for key, param := range pluginParams {
		merged[key] = param.Default
	}

	// Override with provided parameters
	for key, value := range providedParams {
		merged[key] = value
	}

	return merged
}

// substituteParameters performs template variable substitution
func (v *VSCodeExtensionInstaller) substituteParameters(text string, params map[string]interface{}) (string, error) {
	tmpl, err := template.New("install").Parse(text)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var result strings.Builder
	if err := tmpl.Execute(&result, params); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return result.String(), nil
}
