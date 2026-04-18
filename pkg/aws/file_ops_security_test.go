package aws

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateRemotePath(t *testing.T) {
	t.Run("rejects_disallowed_inputs", func(t *testing.T) {
		rejected := []struct {
			name  string
			input string
		}{
			{"empty_path", ""},
			{"path_traversal_middle", "/tmp/file/../etc/shadow"},
			{"path_traversal_relative", "../etc/passwd"},
			{"path_traversal_double_dot_only", ".."},
			{"semicolon", "/tmp/file;rm -rf /"},
			{"pipe", "/tmp/file|cat /etc/shadow"},
			{"backtick", "/tmp/file`whoami`"},
			{"command_substitution", "/tmp/file$(echo x)"},
			{"ampersand", "/tmp/file&background"},
			{"single_quote", "/tmp/file'injection"},
			{"newline", "/tmp/file\ninjection"},
			{"carriage_return", "/tmp/file\rinjection"},
		}

		for _, tt := range rejected {
			t.Run(tt.name, func(t *testing.T) {
				err := validateRemotePath(tt.input)
				assert.Error(t, err, "expected error for input %q", tt.input)
			})
		}
	})

	t.Run("accepts_safe_paths", func(t *testing.T) {
		accepted := []struct {
			name  string
			input string
		}{
			{"absolute_path_with_extension", "/home/ubuntu/file.txt"},
			{"path_with_dashes_and_dots", "/mnt/efs/data/results_2026-04-09.csv"},
			{"path_with_underscores", "/tmp/my_file-v2.tar.gz"},
			{"path_alphanumeric", "/home/user123/workspace"},
			{"path_with_spaces_in_name", "/home/ubuntu/my results.txt"},
		}

		for _, tt := range accepted {
			t.Run(tt.name, func(t *testing.T) {
				err := validateRemotePath(tt.input)
				assert.NoError(t, err, "expected no error for input %q", tt.input)
			})
		}
	})
}

// TestValidateRemotePathCalledByFileOps verifies that PushFile, PullFile, and
// ListRemoteFiles all call validateRemotePath and return errors for invalid paths.
// This ensures the validation is not bypassed at the call sites.
func TestValidateRemotePathCalledByFileOps(t *testing.T) {
	// We don't need a real Manager — we just need to confirm the path validation
	// error is returned before any AWS operations are attempted.
	// Use a minimal Manager with nil clients; any AWS call would panic, so
	// if we get the validation error we know it fired first.
	m := &Manager{} // nil AWS clients — will panic if AWS is called

	ctx := t.Context()
	injected := "/tmp/file;rm -rf /"

	t.Run("PushFile_rejects_bad_remote_path", func(t *testing.T) {
		_, err := m.PushFile(ctx, "instance", "/local/file.txt", injected)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid remote path")
	})

	t.Run("PullFile_rejects_bad_remote_path", func(t *testing.T) {
		_, err := m.PullFile(ctx, "instance", injected, "/local/file.txt")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid remote path")
	})

	t.Run("ListRemoteFiles_rejects_bad_remote_path", func(t *testing.T) {
		_, err := m.ListRemoteFiles(ctx, "instance", injected)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid remote path")
	})
}
