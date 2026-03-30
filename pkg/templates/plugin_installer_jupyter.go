// Package templates provides JupyterLab extension plugin installer implementation.
package templates

import (
	"context"
	"fmt"
	"strings"
	"text/template"
)

// JupyterExtensionInstaller installs JupyterLab extensions on remote instances
type JupyterExtensionInstaller struct {
	executor     RemoteExecutor
	instanceName string
}

// NewJupyterExtensionInstaller creates a new JupyterLab extension installer
func NewJupyterExtensionInstaller(executor RemoteExecutor, instanceName string) *JupyterExtensionInstaller {
	return &JupyterExtensionInstaller{
		executor:     executor,
		instanceName: instanceName,
	}
}

// Name returns the installer name
func (j *JupyterExtensionInstaller) Name() string {
	return "jupyter"
}

// Install installs a JupyterLab extension with specified version and configuration
func (j *JupyterExtensionInstaller) Install(ctx context.Context, plugin *PluginManifest, version string, params map[string]interface{}) error {
	// Get JupyterLab-specific installer from plugin manifest
	jupyterInstaller, exists := plugin.Spec.Installers["jupyter"]
	if !exists {
		return fmt.Errorf("plugin %s does not support JupyterLab installer", plugin.Metadata.Name)
	}

	// Merge plugin configuration parameters with provided parameters
	mergedParams := j.mergeParameters(plugin.Spec.Configuration.Parameters, params)
	mergedParams["version"] = version

	// Install each extension
	for _, pkg := range jupyterInstaller.Packages {
		// Substitute parameters in install command
		installCmd, err := j.substituteParameters(pkg.InstallCommand, mergedParams)
		if err != nil {
			return fmt.Errorf("failed to substitute parameters for extension %s: %w", pkg.Name, err)
		}

		// Execute installation command
		result, err := j.executor.Execute(ctx, j.instanceName, installCmd)
		if err != nil {
			return fmt.Errorf("failed to execute install command for extension %s: %w", pkg.Name, err)
		}

		if result.ExitCode != 0 {
			return fmt.Errorf("install command failed for extension %s (exit code %d): %s",
				pkg.Name, result.ExitCode, result.Stderr)
		}
	}

	// Create configuration files if specified
	for _, config := range jupyterInstaller.Config {
		// Substitute parameters in config content
		content, err := j.substituteParameters(config.Content, mergedParams)
		if err != nil {
			return fmt.Errorf("failed to substitute parameters in config %s: %w", config.File, err)
		}

		// Create directory and write config file
		writeCmd := fmt.Sprintf("mkdir -p $(dirname %s) && cat > %s <<'EOF'\n%s\nEOF",
			config.Path, config.Path, content)

		result, err := j.executor.Execute(ctx, j.instanceName, writeCmd)
		if err != nil {
			return fmt.Errorf("failed to create config file %s: %w", config.File, err)
		}

		if result.ExitCode != 0 {
			return fmt.Errorf("failed to write config %s (exit code %d): %s",
				config.File, result.ExitCode, result.Stderr)
		}
	}

	// Execute post-install script if specified
	if jupyterInstaller.PostInstall != "" {
		postInstall, err := j.substituteParameters(jupyterInstaller.PostInstall, mergedParams)
		if err != nil {
			return fmt.Errorf("failed to substitute parameters in post-install script: %w", err)
		}

		result, err := j.executor.ExecuteScript(ctx, j.instanceName, postInstall)
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

// Uninstall removes a JupyterLab extension
func (j *JupyterExtensionInstaller) Uninstall(ctx context.Context, plugin *PluginManifest) error {
	jupyterInstaller, exists := plugin.Spec.Installers["jupyter"]
	if !exists {
		return fmt.Errorf("plugin %s does not support JupyterLab installer", plugin.Metadata.Name)
	}

	for _, pkg := range jupyterInstaller.Packages {
		// Use jupyter labextension uninstall
		uninstallCmd := fmt.Sprintf("jupyter labextension uninstall %s", pkg.Name)

		result, err := j.executor.Execute(ctx, j.instanceName, uninstallCmd)
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

// Validate checks if JupyterLab is available and prerequisites are met
func (j *JupyterExtensionInstaller) Validate(ctx context.Context, plugin *PluginManifest) error {
	// Check if JupyterLab is installed
	result, err := j.executor.Execute(ctx, j.instanceName, "jupyter --version")
	if err != nil {
		return fmt.Errorf("JupyterLab is not installed: %w", err)
	}

	if result.ExitCode != 0 {
		return fmt.Errorf("JupyterLab is not available (exit code %d): %s", result.ExitCode, result.Stderr)
	}

	// Check compatibility constraints if specified
	if plugin.Spec.Compatibility.Tools != nil {
		if constraint, exists := plugin.Spec.Compatibility.Tools["jupyter"]; exists {
			if installed := parseVersionFromOutput(result.Stdout); installed != "" {
				if constraint.MinVersion != "" && !versionSatisfiesConstraint(installed, ">="+constraint.MinVersion) {
					return fmt.Errorf("installed JupyterLab version %s is below minimum required %s",
						installed, constraint.MinVersion)
				}
				if constraint.MaxVersion != "" && !versionSatisfiesConstraint(installed, "<="+constraint.MaxVersion) {
					return fmt.Errorf("installed JupyterLab version %s exceeds maximum allowed %s",
						installed, constraint.MaxVersion)
				}
			}
		}
	}

	return nil
}

// GetInstalledVersion returns the installed version of an extension
func (j *JupyterExtensionInstaller) GetInstalledVersion(ctx context.Context, plugin *PluginManifest) (string, error) {
	jupyterInstaller, exists := plugin.Spec.Installers["jupyter"]
	if !exists {
		return "", fmt.Errorf("plugin %s does not support JupyterLab installer", plugin.Metadata.Name)
	}

	if len(jupyterInstaller.Packages) == 0 {
		return "", fmt.Errorf("no extensions defined for plugin %s", plugin.Metadata.Name)
	}

	// Check version of first extension
	pkg := jupyterInstaller.Packages[0]
	checkCmd := fmt.Sprintf("jupyter labextension list | grep -i '%s'", pkg.Name)

	result, err := j.executor.Execute(ctx, j.instanceName, checkCmd)
	if err != nil {
		return "", fmt.Errorf("failed to check version: %w", err)
	}

	if result.ExitCode != 0 {
		return "", nil // Not installed
	}

	// Parse version from output (format varies by extension)
	output := strings.TrimSpace(result.Stdout)
	if strings.Contains(output, "v") {
		// Try to extract version (e.g., "extension v1.2.3")
		parts := strings.Fields(output)
		for _, part := range parts {
			if strings.HasPrefix(part, "v") {
				return strings.TrimPrefix(part, "v"), nil
			}
		}
	}

	// Extension is installed but version couldn't be parsed
	return "installed", nil
}

// IsCompatible checks if the plugin is compatible with the environment
func (j *JupyterExtensionInstaller) IsCompatible(ctx context.Context, plugin *PluginManifest) (bool, error) {
	// Check if JupyterLab installer is defined
	_, exists := plugin.Spec.Installers["jupyter"]
	if !exists {
		return false, nil
	}

	// Validate JupyterLab is available
	if err := j.Validate(ctx, plugin); err != nil {
		return false, nil
	}

	return true, nil
}

// mergeParameters merges plugin configuration parameters with provided parameters
func (j *JupyterExtensionInstaller) mergeParameters(pluginParams map[string]TemplateParameter, providedParams map[string]interface{}) map[string]interface{} {
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
func (j *JupyterExtensionInstaller) substituteParameters(text string, params map[string]interface{}) (string, error) {
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
