// Package templates provides validation for plugin usage in templates.
package templates

import (
	"fmt"
	"strings"
)

// PluginValidationRule validates plugin usage in templates
type PluginValidationRule struct {
	registry *PluginRegistry
}

// NewPluginValidationRule creates a new plugin validation rule
func NewPluginValidationRule(registry *PluginRegistry) *PluginValidationRule {
	return &PluginValidationRule{
		registry: registry,
	}
}

// Name returns the rule name
func (r *PluginValidationRule) Name() string {
	return "PluginValidation"
}

// Validate validates template plugin configuration
func (r *PluginValidationRule) Validate(template *Template) []ValidationResult {
	if len(template.Plugins.Enabled) == 0 {
		return nil
	}
	results := r.validateEnabledPluginConfigs(template)
	allPlugins, depResults := r.resolvePluginDependencies(template)
	results = append(results, depResults...)
	if err := r.registry.CheckConflicts(allPlugins); err != nil {
		results = append(results, ValidationResult{
			Level: ValidationError, Field: "plugins.enabled",
			Message: fmt.Sprintf("Plugin conflict detected: %v", err),
		})
	}
	results = append(results, r.checkRequiredDeps(template)...)
	results = append(results, r.suggestOptionalDeps(template)...)
	return results
}

// validateEnabledPluginConfigs validates each enabled plugin's existence and configuration.
func (r *PluginValidationRule) validateEnabledPluginConfigs(template *Template) []ValidationResult {
	var results []ValidationResult
	for i, enabledPlugin := range template.Plugins.Enabled {
		manifest, err := r.registry.GetPlugin(enabledPlugin.Name)
		if err != nil {
			results = append(results, ValidationResult{
				Level:   ValidationError,
				Field:   fmt.Sprintf("plugins.enabled[%d].name", i),
				Message: fmt.Sprintf("Plugin not found: %s", enabledPlugin.Name),
			})
			continue
		}
		if enabledPlugin.Version != "" && !r.isValidVersion(enabledPlugin.Version, manifest) {
			results = append(results, ValidationResult{
				Level:   ValidationWarning,
				Field:   fmt.Sprintf("plugins.enabled[%d].version", i),
				Message: fmt.Sprintf("Unknown version: %s (plugin: %s)", enabledPlugin.Version, enabledPlugin.Name),
			})
		}
		for key := range enabledPlugin.Configuration {
			if _, exists := manifest.Spec.Configuration.Parameters[key]; !exists {
				results = append(results, ValidationResult{
					Level:   ValidationWarning,
					Field:   fmt.Sprintf("plugins.enabled[%d].configuration.%s", i, key),
					Message: fmt.Sprintf("Unknown configuration parameter: %s (plugin: %s)", key, enabledPlugin.Name),
				})
			}
		}
	}
	return results
}

// resolvePluginDependencies collects all transitive plugin deps and any resolution errors.
func (r *PluginValidationRule) resolvePluginDependencies(template *Template) ([]*PluginManifest, []ValidationResult) {
	var allPlugins []*PluginManifest
	var results []ValidationResult
	for _, enabledPlugin := range template.Plugins.Enabled {
		manifest, err := r.registry.GetPlugin(enabledPlugin.Name)
		if err != nil {
			continue // already reported in validateEnabledPluginConfigs
		}
		deps, err := r.registry.ResolveDependencies(manifest)
		if err != nil {
			results = append(results, ValidationResult{
				Level:   ValidationError,
				Field:   fmt.Sprintf("plugins.enabled[%s]", enabledPlugin.Name),
				Message: fmt.Sprintf("Dependency resolution failed: %v", err),
			})
			continue
		}
		allPlugins = append(allPlugins, deps...)
	}
	return allPlugins, results
}

// checkRequiredDeps reports missing required plugin dependencies.
func (r *PluginValidationRule) checkRequiredDeps(template *Template) []ValidationResult {
	var results []ValidationResult
	for _, enabledPlugin := range template.Plugins.Enabled {
		manifest, err := r.registry.GetPlugin(enabledPlugin.Name)
		if err != nil {
			continue
		}
		for _, dep := range manifest.Spec.Dependencies.Required {
			if !r.isPluginEnabled(template, dep.Plugin) {
				results = append(results, ValidationResult{
					Level:   ValidationError,
					Field:   fmt.Sprintf("plugins.enabled[%s]", enabledPlugin.Name),
					Message: fmt.Sprintf("Missing required dependency: %s (version: %s)", dep.Plugin, dep.Version),
				})
			}
		}
	}
	return results
}

// suggestOptionalDeps reports optional plugin dependencies that are not enabled.
func (r *PluginValidationRule) suggestOptionalDeps(template *Template) []ValidationResult {
	var results []ValidationResult
	for _, enabledPlugin := range template.Plugins.Enabled {
		manifest, err := r.registry.GetPlugin(enabledPlugin.Name)
		if err != nil {
			continue
		}
		for _, dep := range manifest.Spec.Dependencies.Optional {
			if !r.isPluginEnabled(template, dep.Plugin) {
				results = append(results, ValidationResult{
					Level:   ValidationInfo,
					Field:   fmt.Sprintf("plugins.enabled[%s]", enabledPlugin.Name),
					Message: fmt.Sprintf("Optional dependency available: %s (version: %s)", dep.Plugin, dep.Version),
				})
			}
		}
	}
	return results
}

// isValidVersion checks if a version is valid for a plugin
func (r *PluginValidationRule) isValidVersion(version string, manifest *PluginManifest) bool {
	// "latest" is always valid
	if version == "latest" {
		return true
	}

	// Check if version contains template variables ({{.param}})
	if strings.Contains(version, "{{") && strings.Contains(version, "}}") {
		// Template variables are valid - they'll be resolved at launch time
		return true
	}

	// Check against plugin's supported versions
	versionParam, exists := manifest.Spec.Configuration.Parameters["version"]
	if !exists {
		// No version parameter defined - any version is valid
		return true
	}

	// Check if version is in choices
	for _, choice := range versionParam.Choices {
		if fmt.Sprintf("%v", choice) == version {
			return true
		}
	}

	return false
}

// isPluginEnabled checks if a plugin is enabled in the template
func (r *PluginValidationRule) isPluginEnabled(template *Template, pluginName string) bool {
	for _, enabled := range template.Plugins.Enabled {
		if enabled.Name == pluginName {
			return true
		}
	}
	return false
}
