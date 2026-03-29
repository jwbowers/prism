package invitation

import (
	"context"
	"fmt"
	"testing"

	"github.com/scttfrdmn/prism/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLogEmailSender_SendInvitation verifies that LogEmailSender returns nil and
// records the call without panicking.
func TestLogEmailSender_SendInvitation(t *testing.T) {
	sender := &LogEmailSender{}
	inv := &types.Invitation{
		ID:        "inv-1",
		Email:     "alice@example.com",
		ProjectID: "proj-1",
		Role:      types.ProjectRoleMember,
		Token:     "tok-abc",
	}
	project := &types.Project{ID: "proj-1", Name: "Test Project"}

	err := sender.SendInvitation(context.Background(), inv, project, "bob")
	assert.NoError(t, err)
}

// TestLogEmailSender_SendInvitation_NilProject verifies the nil-project path
// does not panic.
func TestLogEmailSender_SendInvitation_NilProject(t *testing.T) {
	sender := &LogEmailSender{}
	inv := &types.Invitation{
		Email:     "alice@example.com",
		ProjectID: "proj-x",
		Token:     "tok-xyz",
	}

	err := sender.SendInvitation(context.Background(), inv, nil, "system")
	assert.NoError(t, err)
}

// TestLogEmailSender_SendAcceptanceConfirmation verifies that
// SendAcceptanceConfirmation returns nil without panicking.
func TestLogEmailSender_SendAcceptanceConfirmation(t *testing.T) {
	sender := &LogEmailSender{}
	inv := &types.Invitation{
		Email:     "alice@example.com",
		ProjectID: "proj-1",
	}
	project := &types.Project{ID: "proj-1", Name: "Test Project"}

	err := sender.SendAcceptanceConfirmation(context.Background(), inv, project)
	assert.NoError(t, err)
}

// TestManager_SendInvitationEmail_FireAndForget verifies that a send error from
// the underlying sender is swallowed by the Manager wrapper (fire-and-forget).
func TestManager_SendInvitationEmail_FireAndForget(t *testing.T) {
	tmpDir := t.TempDir()
	invitationsPath := tmpDir + "/invitations.json"

	errSender := &errEmailSender{}
	mgr := &Manager{
		invitationsPath: invitationsPath,
		invitations:     make(map[string]*types.Invitation),
		tokenIndex:      make(map[string]*types.Invitation),
		projectIndex:    make(map[string][]*types.Invitation),
		emailIndex:      make(map[string][]*types.Invitation),
		emailSender:     errSender,
	}

	inv := &types.Invitation{Email: "alice@example.com", ProjectID: "p1", Token: "t1"}
	project := &types.Project{ID: "p1", Name: "Test"}

	// Must not panic or return an error (fire-and-forget).
	require.NotPanics(t, func() {
		mgr.SendInvitationEmail(context.Background(), inv, project, "bob")
	})
}

// TestManager_SendAcceptanceEmail_FireAndForget verifies fire-and-forget on
// acceptance email errors.
func TestManager_SendAcceptanceEmail_FireAndForget(t *testing.T) {
	tmpDir := t.TempDir()
	invitationsPath := tmpDir + "/invitations.json"

	errSender := &errEmailSender{}
	mgr := &Manager{
		invitationsPath: invitationsPath,
		invitations:     make(map[string]*types.Invitation),
		tokenIndex:      make(map[string]*types.Invitation),
		projectIndex:    make(map[string][]*types.Invitation),
		emailIndex:      make(map[string][]*types.Invitation),
		emailSender:     errSender,
	}

	inv := &types.Invitation{Email: "alice@example.com", ProjectID: "p1"}
	project := &types.Project{ID: "p1", Name: "Test"}

	require.NotPanics(t, func() {
		mgr.SendAcceptanceEmail(context.Background(), inv, project)
	})
}

// errEmailSender always returns an error — used to verify fire-and-forget behaviour.
type errEmailSender struct{}

func (e *errEmailSender) SendInvitation(_ context.Context, _ *types.Invitation, _ *types.Project, _ string) error {
	return fmt.Errorf("smtp timeout")
}

func (e *errEmailSender) SendAcceptanceConfirmation(_ context.Context, _ *types.Invitation, _ *types.Project) error {
	return fmt.Errorf("smtp timeout")
}
