package research

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// newTestResearchUserService creates a ResearchUserService suitable for migration tests.
func newTestResearchUserService() *ResearchUserService {
	profileMgr := &MockProfileManager{currentProfile: "test-profile"}
	return NewResearchUserService(&ResearchUserServiceConfig{
		ConfigDir:  "/tmp/test-research-migration",
		ProfileMgr: profileMgr,
	})
}

// TestMigrateExistingUser_ConnectionFailure verifies that a connection failure
// returns an error that includes the instance IP.
func TestMigrateExistingUser_ConnectionFailure(t *testing.T) {
	svc := newTestResearchUserService()

	// Use an unroutable address to guarantee connection failure.
	err := svc.MigrateExistingUser("192.0.2.1", "olduser", "researcher", "/nonexistent/key.pem")

	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "failed to connect"),
		"expected 'failed to connect' in error, got: %s", err.Error())
}

// TestMigrateExistingUser_BadKeyPath verifies that an unreadable SSH key path
// causes a connection error before any remote commands are executed.
func TestMigrateExistingUser_BadKeyPath(t *testing.T) {
	svc := newTestResearchUserService()

	err := svc.MigrateExistingUser("127.0.0.1", "olduser", "researcher", "/does/not/exist.pem")

	assert.Error(t, err)
	// Should fail during connect, not during a migration step.
	assert.False(t, strings.Contains(err.Error(), "migration step"),
		"expected connect error, not step failure: %s", err.Error())
}

// TestMigrateExistingUser_IntegrationSkipped documents the full migration flow
// and is skipped unless a real instance is available.
func TestMigrateExistingUser_IntegrationSkipped(t *testing.T) {
	t.Skip("Integration test: requires a real SSH-accessible instance")

	svc := newTestResearchUserService()

	// Would verify:
	// 1. Backup of existingUser's home dir created at /tmp/<user>-migration-backup.tar.gz
	// 2. newResearchUser created on the instance
	// 3. Home directory contents copied (excluding .ssh)
	// 4. SSH authorized_keys merged and deduplicated
	// 5. Ownership of new home directory set correctly
	err := svc.MigrateExistingUser("instance-ip", "ubuntu", "researcher", "/path/to/key.pem")
	assert.NoError(t, err)
}
