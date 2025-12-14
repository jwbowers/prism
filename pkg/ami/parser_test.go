package ami

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createValidTemplate creates a valid template for testing
func createValidTemplate() *Template {
	return &Template{
		Name:        "test-template",
		Base:        "ubuntu-22.04-server-lts",
		Description: "A test template",
		BuildSteps: []BuildStep{
			{
				Name:   "Install packages",
				Script: "apt-get update && apt-get install -y curl",
			},
		},
	}
}

func TestParser_ParseTemplate_Valid(t *testing.T) {
	parser := &Parser{}

	yamlContent := `
name: test-template
base: ubuntu-22.04-server-lts
description: A test template
build_steps:
  - name: Install packages
    script: apt-get update && apt-get install -y curl
`

	template, err := parser.ParseTemplate(yamlContent)
	require.NoError(t, err)
	assert.Equal(t, "test-template", template.Name)
	assert.Equal(t, "ubuntu-22.04-server-lts", template.Base)
	assert.Equal(t, "A test template", template.Description)
	assert.Len(t, template.BuildSteps, 1)
	assert.Equal(t, "Install packages", template.BuildSteps[0].Name)
}

func TestParser_ParseTemplate_MissingName(t *testing.T) {
	parser := &Parser{}

	yamlContent := `
base: ubuntu-22.04-server-lts
description: A test template
build_steps:
  - name: Install packages
    script: apt-get update
`

	_, err := parser.ParseTemplate(yamlContent)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name")
}

func TestParser_ParseTemplate_MissingBase(t *testing.T) {
	parser := &Parser{}

	yamlContent := `
name: test-template
description: A test template
build_steps:
  - name: Install packages
    script: apt-get update
`

	_, err := parser.ParseTemplate(yamlContent)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "base")
}

func TestParser_ParseTemplate_NoBuildSteps(t *testing.T) {
	parser := &Parser{}

	yamlContent := `
name: test-template
base: ubuntu-22.04-server-lts
description: A test template
build_steps: []
`

	_, err := parser.ParseTemplate(yamlContent)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "build step")
}

func TestParser_ParseTemplate_WithValidation(t *testing.T) {
	parser := &Parser{}

	yamlContent := `
name: test-template
base: ubuntu-22.04-server-lts
description: A test template
build_steps:
  - name: Install curl
    script: apt-get install -y curl
validation:
  - name: Check curl installed
    command: which curl
    success: true
`

	template, err := parser.ParseTemplate(yamlContent)
	require.NoError(t, err)
	assert.Len(t, template.Validation, 1)
	assert.Equal(t, "Check curl installed", template.Validation[0].Name)
	assert.Equal(t, "which curl", template.Validation[0].Command)
	assert.True(t, template.Validation[0].Success)
}

func TestParser_ParseTemplate_WithTags(t *testing.T) {
	parser := &Parser{}

	yamlContent := `
name: test-template
base: ubuntu-22.04-server-lts
description: A test template
build_steps:
  - name: Install packages
    script: apt-get update
tags:
  Environment: production
  Team: platform
`

	template, err := parser.ParseTemplate(yamlContent)
	require.NoError(t, err)
	assert.Len(t, template.Tags, 2)
	assert.Equal(t, "production", template.Tags["Environment"])
	assert.Equal(t, "platform", template.Tags["Team"])
}

func TestParser_ParseTemplate_WithArchitecture(t *testing.T) {
	parser := &Parser{}

	yamlContent := `
name: test-template
base: ubuntu-22.04-server-lts
description: A test template
architecture: arm64
build_steps:
  - name: Install packages
    script: apt-get update
`

	template, err := parser.ParseTemplate(yamlContent)
	require.NoError(t, err)
	assert.Equal(t, "arm64", template.Architecture)
}

func TestParser_ParseTemplate_InvalidYAML(t *testing.T) {
	parser := &Parser{}

	yamlContent := `
name: test-template
base: ubuntu-22.04-server-lts
description: A test template
  - invalid indentation
build_steps:
  - name: Install packages
`

	_, err := parser.ParseTemplate(yamlContent)
	assert.Error(t, err)
}

func TestParser_WriteTemplate(t *testing.T) {
	parser := &Parser{}
	template := createValidTemplate()

	var buf bytes.Buffer
	err := parser.WriteTemplate(template, &buf)
	require.NoError(t, err)

	// Verify written YAML can be parsed back
	parsedTemplate, err := parser.ParseTemplate(buf.String())
	require.NoError(t, err)
	assert.Equal(t, template.Name, parsedTemplate.Name)
	assert.Equal(t, template.Base, parsedTemplate.Base)
	assert.Equal(t, template.Description, parsedTemplate.Description)
}

func TestParser_ValidateTemplate_Valid(t *testing.T) {
	parser := &Parser{}
	template := createValidTemplate()

	err := parser.ValidateTemplate(template)
	assert.NoError(t, err)
}

func TestParser_ValidateTemplate_EmptyName(t *testing.T) {
	parser := &Parser{}
	template := createValidTemplate()
	template.Name = ""

	err := parser.ValidateTemplate(template)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name")
}

func TestParser_ValidateTemplate_EmptyBase(t *testing.T) {
	parser := &Parser{}
	template := createValidTemplate()
	template.Base = ""

	err := parser.ValidateTemplate(template)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "base")
}

func TestParser_ValidateTemplate_NoBuildSteps(t *testing.T) {
	parser := &Parser{}
	template := createValidTemplate()
	template.BuildSteps = []BuildStep{}

	err := parser.ValidateTemplate(template)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "build step")
}

func TestParser_ValidateTemplate_BuildStepMissingName(t *testing.T) {
	parser := &Parser{}
	template := createValidTemplate()
	template.BuildSteps[0].Name = ""

	err := parser.ValidateTemplate(template)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name")
}

func TestParser_ValidateTemplate_BuildStepMissingScript(t *testing.T) {
	parser := &Parser{}
	template := createValidTemplate()
	template.BuildSteps[0].Script = ""

	err := parser.ValidateTemplate(template)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "script")
}

func TestParser_ValidateTemplate_ValidationMissingName(t *testing.T) {
	parser := &Parser{}
	template := createValidTemplate()
	template.Validation = []Validation{
		{
			Command: "which curl",
			Success: true,
		},
	}

	err := parser.ValidateTemplate(template)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name")
}

func TestParser_ValidateTemplate_ValidationMissingCommand(t *testing.T) {
	parser := &Parser{}
	template := createValidTemplate()
	template.Validation = []Validation{
		{
			Name:    "Check curl",
			Success: true,
		},
	}

	err := parser.ValidateTemplate(template)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "command")
}

func TestBuildStep_TimeoutDefault(t *testing.T) {
	step := BuildStep{
		Name:   "Test step",
		Script: "echo test",
	}

	// Default timeout should be 0 (will be set to 600 in actual usage)
	assert.Equal(t, 0, step.TimeoutSeconds)
}

func TestValidation_Types(t *testing.T) {
	// Test success validation
	successValidation := Validation{
		Name:    "Check success",
		Command: "exit 0",
		Success: true,
	}
	assert.Equal(t, "Check success", successValidation.Name)
	assert.True(t, successValidation.Success)

	// Test contains validation
	containsValidation := Validation{
		Name:     "Check output",
		Command:  "echo hello",
		Contains: "hello",
	}
	assert.Equal(t, "hello", containsValidation.Contains)

	// Test equals validation
	equalsValidation := Validation{
		Name:    "Check exact output",
		Command: "echo hello",
		Equals:  "hello",
	}
	assert.Equal(t, "hello", equalsValidation.Equals)
}
