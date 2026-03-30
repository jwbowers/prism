package templates

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestVersionSatisfiesConstraint_TildeOperator verifies patch-level (~) range semantics.
func TestVersionSatisfiesConstraint_TildeOperator(t *testing.T) {
	cases := []struct {
		version    string
		constraint string
		want       bool
	}{
		// ~1.2.3 → >=1.2.3 <1.3.0
		{"1.2.3", "~1.2.3", true},
		{"1.2.9", "~1.2.3", true},
		{"1.3.0", "~1.2.3", false},
		{"1.1.9", "~1.2.3", false},
		{"2.0.0", "~1.2.3", false},
		// ~2.0 → >=2.0 <2.1.0
		{"2.0.5", "~2.0", true},
		{"2.1.0", "~2.0", false},
	}
	for _, tc := range cases {
		got := versionSatisfiesConstraint(tc.version, tc.constraint)
		assert.Equal(t, tc.want, got, "versionSatisfiesConstraint(%q, %q)", tc.version, tc.constraint)
	}
}

// TestVersionSatisfiesConstraint_CaretOperator verifies minor-level (^) range semantics.
func TestVersionSatisfiesConstraint_CaretOperator(t *testing.T) {
	cases := []struct {
		version    string
		constraint string
		want       bool
	}{
		// ^1.2.3 → >=1.2.3 <2.0.0
		{"1.2.3", "^1.2.3", true},
		{"1.9.9", "^1.2.3", true},
		{"2.0.0", "^1.2.3", false},
		{"1.2.2", "^1.2.3", false},
		{"0.9.9", "^1.2.3", false},
		// ^3.0.0 → >=3.0.0 <4.0.0
		{"3.5.1", "^3.0.0", true},
		{"4.0.0", "^3.0.0", false},
	}
	for _, tc := range cases {
		got := versionSatisfiesConstraint(tc.version, tc.constraint)
		assert.Equal(t, tc.want, got, "versionSatisfiesConstraint(%q, %q)", tc.version, tc.constraint)
	}
}

// TestVersionSatisfiesConstraint_ExistingOperators ensures existing >=/>/<=/< still work.
func TestVersionSatisfiesConstraint_ExistingOperators(t *testing.T) {
	assert.True(t, versionSatisfiesConstraint("3.8.1", ">=3.8"))
	assert.False(t, versionSatisfiesConstraint("3.7.9", ">=3.8"))
	assert.True(t, versionSatisfiesConstraint("1.0.0", "<=2.0.0"))
	assert.False(t, versionSatisfiesConstraint("2.0.1", "<=2.0.0"))
	assert.True(t, versionSatisfiesConstraint("1.2.3", "1.2.3"))
	assert.False(t, versionSatisfiesConstraint("1.2.4", "1.2.3"))
	assert.True(t, versionSatisfiesConstraint("anything", "latest"))
	assert.True(t, versionSatisfiesConstraint("anything", ""))
}

// TestParseVersionFromOutput verifies that version strings are extracted correctly.
func TestParseVersionFromOutput(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"JupyterLab 3.6.5", "3.6.5"},
		{"R version 4.3.1 (2023-06-16)", "4.3.1"},
		{"1.87.2\n7f6fe36\nx64", "1.87.2"}, // VS Code format
		{"Python 3.11.4", "3.11.4"},
		{"no version here", ""},
		{"version: 2.1", "2.1"},
	}
	for _, tc := range cases {
		got := parseVersionFromOutput(tc.input)
		assert.Equal(t, tc.want, got, "parseVersionFromOutput(%q)", tc.input)
	}
}
