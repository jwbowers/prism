package templates

// Tests for VimPluginInstaller — verifies that Validate() enforces MinVersion/MaxVersion
// constraints and that GetInstalledVersion() extracts the version number from vim --version
// output rather than returning the opaque "installed" string.

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// makeVimPlugin builds a minimal PluginManifest for the vim installer with an optional
// tool constraint so tests don't need to repeat the full struct literal.
func makeVimPlugin(minVersion, maxVersion string) *PluginManifest {
	var tools map[string]ToolVersionConstraint
	if minVersion != "" || maxVersion != "" {
		tools = map[string]ToolVersionConstraint{
			"vim": {MinVersion: minVersion, MaxVersion: maxVersion},
		}
	}
	return &PluginManifest{
		Metadata: PluginMetadata{Name: "test-vim-plugin"},
		Spec: PluginSpec{
			Installers: map[string]PluginInstallSpec{
				"vim": {
					Packages: []PluginPackage{{Name: "ale", InstallCommand: "echo install ale"}},
				},
			},
			Compatibility: PluginCompatibility{Tools: tools},
		},
	}
}

// TestVimValidate_VersionConstraintEnforced verifies that Validate() returns an error when
// the installed Vim version does not satisfy the MinVersion or MaxVersion constraint.
func TestVimValidate_VersionConstraintEnforced(t *testing.T) {
	vimVersionOutput := "VIM - Vi IMproved 9.0 (2022 Jun 28, compiled ...)"

	t.Run("satisfies MinVersion", func(t *testing.T) {
		mock := NewMockRemoteExecutor()
		mock.SetResult("vim --version", &ExecutionResult{ExitCode: 0, Stdout: vimVersionOutput})
		mock.SetResult("test -f ~/.vim/autoload/plug.vim", &ExecutionResult{ExitCode: 0})
		installer := NewVimPluginInstaller(mock, "test-instance")

		err := installer.Validate(context.Background(), makeVimPlugin("8.0", ""))
		assert.NoError(t, err, "version 9.0 should satisfy MinVersion 8.0")
	})

	t.Run("violates MinVersion", func(t *testing.T) {
		mock := NewMockRemoteExecutor()
		mock.SetResult("vim --version", &ExecutionResult{ExitCode: 0, Stdout: vimVersionOutput})
		mock.SetResult("test -f ~/.vim/autoload/plug.vim", &ExecutionResult{ExitCode: 0})
		installer := NewVimPluginInstaller(mock, "test-instance")

		err := installer.Validate(context.Background(), makeVimPlugin("10.0", ""))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "below minimum required")
	})

	t.Run("satisfies MaxVersion", func(t *testing.T) {
		mock := NewMockRemoteExecutor()
		mock.SetResult("vim --version", &ExecutionResult{ExitCode: 0, Stdout: vimVersionOutput})
		mock.SetResult("test -f ~/.vim/autoload/plug.vim", &ExecutionResult{ExitCode: 0})
		installer := NewVimPluginInstaller(mock, "test-instance")

		err := installer.Validate(context.Background(), makeVimPlugin("", "9.1"))
		assert.NoError(t, err, "version 9.0 should satisfy MaxVersion 9.1")
	})

	t.Run("violates MaxVersion", func(t *testing.T) {
		mock := NewMockRemoteExecutor()
		mock.SetResult("vim --version", &ExecutionResult{ExitCode: 0, Stdout: vimVersionOutput})
		mock.SetResult("test -f ~/.vim/autoload/plug.vim", &ExecutionResult{ExitCode: 0})
		installer := NewVimPluginInstaller(mock, "test-instance")

		err := installer.Validate(context.Background(), makeVimPlugin("", "8.2"))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "exceeds maximum allowed")
	})
}

// TestVimValidate_NoConstraint verifies that Validate() succeeds when no version
// constraints are defined (empty Tools map).
func TestVimValidate_NoConstraint(t *testing.T) {
	mock := NewMockRemoteExecutor()
	mock.SetResult("vim --version", &ExecutionResult{
		ExitCode: 0,
		Stdout:   "VIM - Vi IMproved 9.0 (2022 Jun 28, compiled ...)",
	})
	mock.SetResult("test -f ~/.vim/autoload/plug.vim", &ExecutionResult{ExitCode: 0})
	installer := NewVimPluginInstaller(mock, "test-instance")

	err := installer.Validate(context.Background(), makeVimPlugin("", ""))
	assert.NoError(t, err)
}

// TestVimGetInstalledVersion_ParsesFromStdout verifies that GetInstalledVersion() extracts
// the numeric version from vim --version output instead of returning "installed".
func TestVimGetInstalledVersion_ParsesFromStdout(t *testing.T) {
	dirCheck := "test -d ~/.vim/plugged/ale && echo 'installed' || echo 'not installed'"
	mock := NewMockRemoteExecutor()
	mock.SetResult(dirCheck, &ExecutionResult{ExitCode: 0, Stdout: "installed\n"})
	mock.SetResult("vim --version", &ExecutionResult{
		ExitCode: 0,
		Stdout:   "VIM - Vi IMproved 9.1 (2024 Apr 03, compiled ...)\n",
	})
	installer := NewVimPluginInstaller(mock, "test-instance")

	version, err := installer.GetInstalledVersion(context.Background(), makeVimPlugin("", ""))
	require.NoError(t, err)
	assert.Equal(t, "9.1", version)
}
