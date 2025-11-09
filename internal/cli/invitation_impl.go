package cli

import (
	"fmt"
	"time"

	"github.com/scttfrdmn/prism/pkg/api/client"
	"github.com/scttfrdmn/prism/pkg/invitation"
	"github.com/scttfrdmn/prism/pkg/types"
)

// Invitation handles user-facing invitation commands (accept, decline, list)
func (a *App) Invitation(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: prism invitation <action> [args]")
	}

	action := args[0]
	invArgs := args[1:]

	// Check daemon is running
	if err := a.apiClient.Ping(a.ctx); err != nil {
		return fmt.Errorf("daemon not running. Start with: prism daemon start")
	}

	switch action {
	case "add":
		return a.invitationAdd(invArgs)
	case "list":
		return a.invitationList(invArgs)
	case "accept":
		return a.invitationAccept(invArgs)
	case "decline":
		return a.invitationDecline(invArgs)
	case "info":
		return a.invitationInfo(invArgs)
	case "remove":
		return a.invitationRemove(invArgs)
	default:
		return fmt.Errorf("unknown invitation action: %s", action)
	}
}

// invitationAdd adds an invitation token to local cache
func (a *App) invitationAdd(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: prism invitation add <token>")
	}

	token := args[0]

	// Initialize cache
	cache, err := invitation.NewInvitationCache()
	if err != nil {
		return fmt.Errorf("failed to initialize invitation cache: %w", err)
	}

	// Get invitation details from daemon by token
	fmt.Printf("🔍 Validating invitation token...\n")
	fmt.Printf("   Token: %s...\n", token[:min(len(token), 20)])

	// Fetch invitation from API
	resp, err := a.apiClient.GetInvitationByToken(a.ctx, token)
	if err != nil {
		return fmt.Errorf("failed to validate invitation token: %w", err)
	}

	inv := resp.Invitation
	projectName := resp.Project.Name

	// Add to cache
	if err := cache.Add(a.ctx, token, inv, projectName); err != nil {
		return fmt.Errorf("failed to add invitation to cache: %w", err)
	}

	fmt.Printf("\n📬 Invitation Added to Cache\n")
	fmt.Printf("   ├─ Project: %s\n", projectName)
	fmt.Printf("   ├─ Role: %s\n", formatInvitationRole(inv.Role))
	fmt.Printf("   ├─ From: %s\n", inv.InvitedBy)
	fmt.Printf("   └─ Expires: %s (%s)\n\n", inv.ExpiresAt.Format("2006-01-02"), formatTimeRemaining(inv.ExpiresAt))

	fmt.Println("💡 View all invitations: prism invitation list")
	fmt.Printf("💡 Accept: prism invitation accept %s\n", projectName)

	return nil
}

// invitationList lists all cached invitations
func (a *App) invitationList(args []string) error {
	// Initialize cache
	cache, err := invitation.NewInvitationCache()
	if err != nil {
		return fmt.Errorf("failed to initialize invitation cache: %w", err)
	}

	// Automatically cleanup expired invitations
	removed, err := cache.CleanupExpired()
	if err != nil {
		fmt.Printf("⚠️  Warning: Failed to cleanup expired invitations: %v\n", err)
	} else if removed > 0 {
		fmt.Printf("🧹 Automatically removed %d expired invitation(s)\n\n", removed)
	}

	// Get all cached invitations
	invitations, err := cache.List()
	if err != nil {
		return fmt.Errorf("failed to list invitations: %w", err)
	}

	// Get summary
	total, pending, accepted, declined, expired := cache.GetSummary()

	if len(invitations) == 0 {
		fmt.Println("📬 No cached invitations")
		fmt.Println("\n💡 Add an invitation: prism invitation add <token>")
		return nil
	}

	fmt.Printf("📬 Your Invitations (%d total)\n\n", total)
	fmt.Printf("   Summary: %d pending, %d accepted, %d declined, %d expired\n\n", pending, accepted, declined, expired)

	// Display each invitation
	for i, inv := range invitations {
		prefix := "├─"
		if i == len(invitations)-1 {
			prefix = "└─"
		}

		// Check if expired
		isExpired := time.Now().After(inv.ExpiresAt)
		statusIcon := "⏳"
		if isExpired {
			statusIcon = "⏰"
		} else if inv.Status == types.InvitationAccepted {
			statusIcon = "✅"
		} else if inv.Status == types.InvitationDeclined {
			statusIcon = "❌"
		}

		fmt.Printf("%s %s %s\n", prefix, statusIcon, inv.ProjectName)
		fmt.Printf("   ├─ Role: %s\n", formatInvitationRole(inv.Role))
		fmt.Printf("   ├─ From: %s\n", inv.InvitedBy)
		fmt.Printf("   └─ Expires: %s\n", formatTimeRemaining(inv.ExpiresAt))

		if inv.Status == types.InvitationPending && !isExpired {
			fmt.Printf("      💡 Accept: prism invitation accept %s\n", inv.ProjectName)
		}

		if i < len(invitations)-1 {
			fmt.Println()
		}
	}

	fmt.Println("\n💡 Cleanup expired: prism invitation cleanup")

	return nil
}

// invitationAccept accepts an invitation (individual or shared token)
func (a *App) invitationAccept(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: prism invitation accept <project-name|token>")
	}

	input := args[0]

	// Check if this is a shared token (format: WORD-WORD-YEAR-XXXX)
	if isSharedToken(input) {
		return a.redeemSharedToken(input)
	}

	// Standard individual invitation flow
	// Initialize cache
	cache, err := invitation.NewInvitationCache()
	if err != nil {
		return fmt.Errorf("failed to initialize invitation cache: %w", err)
	}

	// Try to find invitation in cache by project name
	var token string
	var projectName string
	var cached *invitation.CachedInvitation

	// Check if input looks like a token (long string) or project name
	if len(input) > 50 {
		// Looks like a token - find in cache
		invitations, _ := cache.List()
		found := false
		for _, inv := range invitations {
			if inv.Token == input {
				token = inv.Token
				projectName = inv.ProjectName
				cached = inv
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("invitation token not found in cache. Add it first with: prism invitation add <token>")
		}
	} else {
		// Assume it's a project name
		var err error
		cached, err = cache.Get(input)
		if err != nil {
			return fmt.Errorf("no cached invitation for project '%s'. Add it first with: prism invitation add <token>", input)
		}
		token = cached.Token
		projectName = input
	}

	// Check if invitation is expired
	if time.Now().After(cached.ExpiresAt) {
		return fmt.Errorf("❌ Invitation has expired\n\n"+
			"   Project: %s\n"+
			"   Expired: %s\n"+
			"   Time ago: %s\n\n"+
			"   Please request a new invitation from the project owner.\n"+
			"   Remove expired invitation: prism invitation remove %s",
			projectName,
			cached.ExpiresAt.Format("2006-01-02 15:04"),
			formatTimeAgo(cached.ExpiresAt),
			projectName)
	}

	// Check if already accepted
	if cached.Status == types.InvitationAccepted {
		return fmt.Errorf("⚠️  Invitation already accepted\n\n"+
			"   Project: %s\n"+
			"   You already have access to this project.\n"+
			"   Remove from cache: prism invitation remove %s",
			projectName, projectName)
	}

	// Check if already declined
	if cached.Status == types.InvitationDeclined {
		return fmt.Errorf("⚠️  Invitation was declined\n\n"+
			"   Project: %s\n"+
			"   You previously declined this invitation.\n"+
			"   Remove from cache: prism invitation remove %s",
			projectName, projectName)
	}

	// Accept invitation via API
	fmt.Printf("✅ Accepting invitation to %s...\n", projectName)
	fmt.Printf("   Token: %s...\n", token[:min(len(token), 20)])

	resp, err := a.apiClient.AcceptInvitation(a.ctx, token)
	if err != nil {
		return fmt.Errorf("failed to accept invitation: %w", err)
	}

	// Update cache status
	if err := cache.UpdateStatus(projectName, types.InvitationAccepted); err != nil {
		fmt.Printf("⚠️  Warning: Failed to update cache: %v\n", err)
	}

	fmt.Printf("\n✅ %s\n", resp.Message)
	fmt.Printf("   Project: %s\n", projectName)
	fmt.Println("   You can now manage resources in this project")

	return nil
}

// isSharedToken checks if a token follows the shared token format
// Format: WORD-WORD-YEAR-XXXX (e.g., WORKSHOP-NEURIPS-2025-A4F2)
func isSharedToken(token string) bool {
	// Shared tokens are relatively short (20-40 characters)
	// Individual tokens are much longer (60+ characters)
	if len(token) > 50 {
		return false
	}

	// Check for year pattern (2024-2099)
	// Shared tokens contain a 4-digit year
	for i := 0; i < len(token)-3; i++ {
		if token[i] >= '2' && token[i] <= '2' &&
			token[i+1] >= '0' && token[i+1] <= '0' &&
			token[i+2] >= '0' && token[i+2] <= '9' &&
			token[i+3] >= '0' && token[i+3] <= '9' {
			// Found year pattern 20XX - likely shared token
			return true
		}
	}

	return false
}

// redeemSharedToken redeems a shared invitation token
func (a *App) redeemSharedToken(tokenCode string) error {
	fmt.Printf("🎫 Redeeming shared token...\n")
	fmt.Printf("   Token: %s\n", tokenCode)

	// Get token info first to show details
	token, err := a.apiClient.GetSharedToken(a.ctx, tokenCode)
	if err != nil {
		return fmt.Errorf("failed to get token information: %w", err)
	}

	// Display token information
	fmt.Println()
	fmt.Printf("   Name: %s\n", token.Name)
	fmt.Printf("   Role: %s\n", token.Role)
	fmt.Printf("   Redemptions: %d / %d used\n", token.RedemptionCount, token.RedemptionLimit)
	if token.ExpiresAt != nil {
		fmt.Printf("   Expires: %s\n", token.ExpiresAt.Format("Jan 2, 2006"))
	}
	fmt.Println()

	// Check if token can be redeemed
	if !token.CanRedeem() {
		switch token.Status {
		case types.SharedTokenExpired:
			return fmt.Errorf("❌ Token has expired\n\n" +
				"   Contact the project owner for a new token or extension")
		case types.SharedTokenRevoked:
			return fmt.Errorf("❌ Token has been revoked\n\n" +
				"   Contact the project owner for more information")
		case types.SharedTokenExhausted:
			return fmt.Errorf("❌ Token redemption limit reached\n\n"+
				"   All %d slots have been claimed\n"+
				"   Contact the project owner to increase the limit",
				token.RedemptionLimit)
		default:
			return fmt.Errorf("❌ Token cannot be redeemed: %s", token.Status)
		}
	}

	// TODO: Get actual username from authentication system
	// For now, use a placeholder
	username := "cli-user"

	// Redeem token via API
	req := &types.RedeemSharedTokenRequest{
		Token:      tokenCode,
		RedeemedBy: username,
	}

	resp, err := a.apiClient.RedeemSharedToken(a.ctx, req)
	if err != nil {
		// Handle specific error cases
		if err.Error() == "Token not found" {
			return fmt.Errorf("token not found")
		}
		if err.Error() == "Token has expired" {
			return fmt.Errorf("token has expired")
		}
		if err.Error() == "Token has been revoked" {
			return fmt.Errorf("token has been revoked")
		}
		if err.Error() == "Token has reached redemption limit" {
			return fmt.Errorf("token redemption limit reached")
		}
		if err.Error() == "You have already redeemed this token" {
			return fmt.Errorf("you have already redeemed this token")
		}
		return fmt.Errorf("failed to redeem token: %w", err)
	}

	// Display success
	fmt.Println("✅ Token redeemed successfully!")
	fmt.Println()
	fmt.Printf("   You are participant #%d of %d\n", resp.RedemptionPosition, token.RedemptionLimit)
	fmt.Printf("   Role assigned: %s\n", token.Role)
	if resp.RemainingRedemptions > 0 {
		fmt.Printf("   Remaining slots: %d\n", resp.RemainingRedemptions)
	} else {
		fmt.Printf("   ⚠️  This was the last available slot!\n")
	}
	fmt.Println()
	fmt.Println("   You now have access to the project")
	fmt.Println()
	fmt.Println("💡 View your projects: prism project list")

	return nil
}

// invitationDecline declines an invitation
func (a *App) invitationDecline(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: prism invitation decline <project-name> [--reason <reason>]")
	}

	projectName := args[0]
	reason := ""

	// Parse flags
	for i := 1; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--reason" && i+1 < len(args):
			reason = args[i+1]
			i++
		default:
			return fmt.Errorf("unknown option: %s", arg)
		}
	}

	// Initialize cache
	cache, err := invitation.NewInvitationCache()
	if err != nil {
		return fmt.Errorf("failed to initialize invitation cache: %w", err)
	}

	// Get invitation from cache
	cached, err := cache.Get(projectName)
	if err != nil {
		return fmt.Errorf("no cached invitation for project '%s'", projectName)
	}

	// Check if invitation is expired
	if time.Now().After(cached.ExpiresAt) {
		return fmt.Errorf("❌ Invitation has expired\n\n"+
			"   Project: %s\n"+
			"   Expired: %s\n"+
			"   Time ago: %s\n\n"+
			"   No need to decline - invitation is already invalid.\n"+
			"   Remove expired invitation: prism invitation remove %s",
			projectName,
			cached.ExpiresAt.Format("2006-01-02 15:04"),
			formatTimeAgo(cached.ExpiresAt),
			projectName)
	}

	// Check if already accepted
	if cached.Status == types.InvitationAccepted {
		return fmt.Errorf("⚠️  Invitation already accepted\n\n"+
			"   Project: %s\n"+
			"   You already have access to this project.\n"+
			"   To leave the project, contact the project owner",
			projectName)
	}

	// Check if already declined
	if cached.Status == types.InvitationDeclined {
		return fmt.Errorf("⚠️  Invitation already declined\n\n"+
			"   Project: %s\n"+
			"   You previously declined this invitation.\n"+
			"   Remove from cache: prism invitation remove %s",
			projectName, projectName)
	}

	// Decline invitation via API
	fmt.Printf("❌ Declining invitation to %s...\n", projectName)
	if reason != "" {
		fmt.Printf("   Reason: %s\n", reason)
	}

	resp, err := a.apiClient.DeclineInvitation(a.ctx, cached.Token, reason)
	if err != nil {
		return fmt.Errorf("failed to decline invitation: %w", err)
	}

	// Update cache status
	if err := cache.UpdateStatus(projectName, types.InvitationDeclined); err != nil {
		fmt.Printf("⚠️  Warning: Failed to update cache: %v\n", err)
	}

	fmt.Printf("\n✅ %s\n", resp.Message)

	return nil
}

// invitationInfo shows invitation details
func (a *App) invitationInfo(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: prism invitation info <project-name>")
	}

	projectName := args[0]

	// Initialize cache
	cache, err := invitation.NewInvitationCache()
	if err != nil {
		return fmt.Errorf("failed to initialize invitation cache: %w", err)
	}

	// Get invitation from cache
	cached, err := cache.Get(projectName)
	if err != nil {
		return fmt.Errorf("no cached invitation for project '%s'", projectName)
	}

	// Display invitation details
	isExpired := time.Now().After(cached.ExpiresAt)

	fmt.Printf("\n📧 Invitation Details\n")
	fmt.Printf("   ├─ Project: %s\n", cached.ProjectName)
	fmt.Printf("   ├─ Status: %s", formatInvitationStatus(cached.Status))
	if isExpired {
		fmt.Printf(" ⏰ (expired)")
	}
	fmt.Println()
	fmt.Printf("   ├─ Role: %s\n", formatInvitationRole(cached.Role))
	fmt.Printf("   ├─ Email: %s\n", cached.Email)
	fmt.Printf("   ├─ Invited by: %s\n", cached.InvitedBy)
	fmt.Printf("   ├─ Invited at: %s\n", cached.InvitedAt.Format("2006-01-02 15:04"))
	fmt.Printf("   ├─ Expires: %s (%s)\n", cached.ExpiresAt.Format("2006-01-02 15:04"), formatTimeRemaining(cached.ExpiresAt))
	fmt.Printf("   ├─ Added to cache: %s\n", cached.AddedAt.Format("2006-01-02 15:04"))

	if cached.Message != "" {
		fmt.Printf("   └─ Message: %s\n", cached.Message)
	} else {
		fmt.Printf("   └─ Token: %s...\n", cached.Token[:min(len(cached.Token), 30)])
	}

	if cached.Status == types.InvitationPending && !isExpired {
		fmt.Printf("\n💡 Accept: prism invitation accept %s\n", projectName)
		fmt.Printf("💡 Decline: prism invitation decline %s\n", projectName)
	}

	return nil
}

// invitationRemove removes an invitation from cache
func (a *App) invitationRemove(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: prism invitation remove <project-name>")
	}

	projectName := args[0]

	// Initialize cache
	cache, err := invitation.NewInvitationCache()
	if err != nil {
		return fmt.Errorf("failed to initialize invitation cache: %w", err)
	}

	// Remove from cache
	if err := cache.Remove(projectName); err != nil {
		return fmt.Errorf("failed to remove invitation: %w", err)
	}

	fmt.Printf("🗑️  Removed invitation for %s from cache\n", projectName)

	return nil
}

// projectInvite sends an invitation to collaborate on a project
func (a *App) projectInvite(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: prism project invite <project> <email> [--role <role>] [--message <message>] [--expires-in <duration>]")
	}

	projectName := args[0]
	email := args[1]
	role := "member"  // TODO: Use in API call when implemented
	message := ""     // TODO: Use in API call when implemented
	expiresIn := "7d" // Default: 7 days
	expiresOn := ""   // Absolute date

	// Parse flags
	for i := 2; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--role" && i+1 < len(args):
			role = args[i+1]
			i++
		case arg == "--message" && i+1 < len(args):
			message = args[i+1]
			i++
		case arg == "--expires-in" && i+1 < len(args):
			expiresIn = args[i+1]
			i++
		case arg == "--expires-on" && i+1 < len(args):
			expiresOn = args[i+1]
			i++
		default:
			return fmt.Errorf("unknown option: %s", arg)
		}
	}

	// Validate expiration flags are mutually exclusive
	if expiresIn != "7d" && expiresOn != "" {
		return fmt.Errorf("cannot use both --expires-in and --expires-on flags")
	}

	// Parse expiration
	var expiresAt time.Time
	var expirationDisplay string
	var err error

	if expiresOn != "" {
		// Parse absolute date
		expiresAt, err = parseAbsoluteDate(expiresOn)
		if err != nil {
			return fmt.Errorf("invalid date: %w\n\n"+
				"Valid formats:\n"+
				"  - Date: 2025-12-25\n"+
				"  - Date and time: 2025-12-25 15:00\n"+
				"  - ISO 8601: 2025-12-25T15:00:00Z\n"+
				"Examples:\n"+
				"  --expires-on 2025-12-25\n"+
				"  --expires-on \"2025-12-25 15:00\"", err)
		}
		expirationDisplay = expiresOn
	} else {
		// Parse duration
		duration, err := parseDuration(expiresIn)
		if err != nil {
			return fmt.Errorf("invalid duration: %w\n\n"+
				"Valid formats:\n"+
				"  - Days: 7d, 14d, 30d\n"+
				"  - Hours: 24h, 48h\n"+
				"  - Combinations: 7d12h\n"+
				"Examples:\n"+
				"  --expires-in 7d\n"+
				"  --expires-in 14d\n"+
				"  --expires-in 48h", err)
		}
		expiresAt = time.Now().Add(duration)
		expirationDisplay = expiresIn
	}

	// TODO: Use role, message, and expiration in API call when implemented
	_, _, _ = role, message, expiresAt // Suppress unused variable warnings

	// Get project ID by name
	projectResponse, err := a.apiClient.ListProjects(a.ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to list projects: %w", err)
	}

	var projectID string
	for _, proj := range projectResponse.Projects {
		if proj.Name == projectName {
			projectID = proj.ID
			break
		}
	}

	if projectID == "" {
		return fmt.Errorf("project not found: %s", projectName)
	}

	// Parse role from string
	projectRole := types.ProjectRole(role)

	// Create invitation via API
	req := client.SendInvitationRequest{
		Email:     email,
		Role:      projectRole,
		Message:   message,
		ExpiresAt: &expiresAt,
	}

	resp, err := a.apiClient.SendInvitation(a.ctx, projectID, req)
	if err != nil {
		return fmt.Errorf("failed to send invitation: %w", err)
	}

	inv := resp.Invitation

	fmt.Printf("📧 Invitation Created for %s\n\n", projectName)
	fmt.Printf("   ├─ Recipient: %s\n", email)
	fmt.Printf("   ├─ Role: %s\n", role)
	fmt.Printf("   ├─ Expires: %s (%s)\n", inv.ExpiresAt.Format("2006-01-02 15:04"), formatTimeRemaining(inv.ExpiresAt))
	if expiresOn != "" {
		fmt.Printf("   └─ Expiration: on %s\n\n", expirationDisplay)
	} else {
		fmt.Printf("   └─ Expiration: in %s\n\n", expirationDisplay)
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("Share this token with %s:\n\n", email)
	fmt.Printf("%s\n\n", inv.Token)
	fmt.Println("They can add it with:")
	fmt.Printf("  prism invitation add %s\n", inv.Token)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("✨ Token copied to clipboard!")
	fmt.Println("\n💡 Send via email, Slack, Teams, or paste directly")

	fmt.Printf("\n✅ %s\n", resp.Message)

	return nil
}

// projectInvitations manages project invitations (list, resend, revoke)
func (a *App) projectInvitations(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: prism project invitations <action> [args]")
	}

	action := args[0]
	invArgs := args[1:]

	switch action {
	case "list":
		return a.projectInvitationsList(invArgs)
	case "resend":
		return a.projectInvitationsResend(invArgs)
	case "revoke":
		return a.projectInvitationsRevoke(invArgs)
	case "bulk":
		return a.projectInvitationsBulk(invArgs)
	default:
		return fmt.Errorf("unknown invitations action: %s", action)
	}
}

// projectInvitationsList lists invitations for a project
func (a *App) projectInvitationsList(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: prism project invitations list <project> [--status <status>]")
	}

	projectName := args[0]
	status := ""

	// Parse flags
	for i := 1; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--status" && i+1 < len(args):
			status = args[i+1]
			i++
		default:
			return fmt.Errorf("unknown option: %s", arg)
		}
	}

	// Get project ID by name
	projectResponse, err := a.apiClient.ListProjects(a.ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to list projects: %w", err)
	}

	var projectID string
	for _, proj := range projectResponse.Projects {
		if proj.Name == projectName {
			projectID = proj.ID
			break
		}
	}

	if projectID == "" {
		return fmt.Errorf("project not found: %s", projectName)
	}

	fmt.Printf("📋 Invitations for Project '%s'\n\n", projectName)
	if status != "" {
		fmt.Printf("   Status filter: %s\n", status)
	}
	fmt.Println("\n⚠️  API integration pending")
	fmt.Printf("   Invitations would be fetched via: GET /api/v1/projects/%s/invitations\n", projectID)

	return nil
}

// projectInvitationsResend resends a pending invitation
func (a *App) projectInvitationsResend(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: prism project invitations resend <invitation-id>")
	}

	invitationID := args[0]

	fmt.Printf("🔄 Resending invitation...\n")
	fmt.Printf("   Invitation ID: %s\n", invitationID)
	fmt.Println("\n⚠️  API integration pending")
	fmt.Printf("   Invitation would be resent via: POST /api/v1/invitations/%s/resend\n", invitationID)

	return nil
}

// projectInvitationsRevoke revokes a pending invitation
func (a *App) projectInvitationsRevoke(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: prism project invitations revoke <invitation-id>")
	}

	invitationID := args[0]

	fmt.Printf("🚫 Revoking invitation...\n")
	fmt.Printf("   Invitation ID: %s\n", invitationID)
	fmt.Println("\n⚠️  API integration pending")
	fmt.Printf("   Invitation would be revoked via: DELETE /api/v1/invitations/%s\n", invitationID)

	return nil
}

// projectInvitationsBulk sends bulk invitations from file or inline emails
func (a *App) projectInvitationsBulk(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: prism project invitations bulk <project> [--file <path> | --emails <list>] [options]")
	}

	projectName := args[0]

	// Parse flags
	var file, emails, role, message, expiresIn, expiresOn string
	role = "member"  // default
	expiresIn = "7d" // default

	for i := 1; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--file" && i+1 < len(args):
			file = args[i+1]
			i++
		case arg == "--emails" && i+1 < len(args):
			emails = args[i+1]
			i++
		case arg == "--role" && i+1 < len(args):
			role = args[i+1]
			i++
		case arg == "--message" && i+1 < len(args):
			message = args[i+1]
			i++
		case arg == "--expires-in" && i+1 < len(args):
			expiresIn = args[i+1]
			i++
		case arg == "--expires-on" && i+1 < len(args):
			expiresOn = args[i+1]
			i++
		default:
			return fmt.Errorf("unknown option: %s", arg)
		}
	}

	// Validate file or emails provided
	if file == "" && emails == "" {
		return fmt.Errorf("must provide either --file or --emails")
	}
	if file != "" && emails != "" {
		return fmt.Errorf("cannot use both --file and --emails")
	}

	// Get project ID by name
	projectResponse, err := a.apiClient.ListProjects(a.ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to list projects: %w", err)
	}

	var projectID string
	for _, proj := range projectResponse.Projects {
		if proj.Name == projectName {
			projectID = proj.ID
			break
		}
	}

	if projectID == "" {
		return fmt.Errorf("project not found: %s", projectName)
	}

	// Parse invitation entries
	var entries []types.BulkInvitationEntry
	if file != "" {
		fmt.Printf("📄 Parsing invitation file: %s\n", file)
		entries, err = invitation.ParseInvitationFile(file)
		if err != nil {
			return fmt.Errorf("failed to parse file: %w", err)
		}
	} else {
		fmt.Printf("📧 Parsing inline email list\n")
		entries, err = invitation.ParseInlineEmails(emails)
		if err != nil {
			return fmt.Errorf("failed to parse emails: %w", err)
		}
	}

	fmt.Printf("   Found %d email(s)\n\n", len(entries))

	// Validate roles
	if err := invitation.ValidateRoles(entries); err != nil {
		return fmt.Errorf("invalid roles: %w", err)
	}

	// Build bulk invitation request
	bulkReq := &types.BulkInvitationRequest{
		Invitations:    entries,
		DefaultRole:    types.ProjectRole(role),
		DefaultMessage: message,
		ExpiresIn:      expiresIn,
	}

	// Parse expires-on if provided
	if expiresOn != "" {
		expiresAt, err := time.Parse("2006-01-02", expiresOn)
		if err != nil {
			return fmt.Errorf("invalid expires-on date (use YYYY-MM-DD): %w", err)
		}
		bulkReq.ExpiresAt = &expiresAt
		bulkReq.ExpiresIn = "" // Clear ExpiresIn if ExpiresAt is set
	}

	fmt.Printf("📬 Sending bulk invitations to project '%s'...\n", projectName)
	fmt.Printf("   Default role: %s\n", role)
	if message != "" {
		fmt.Printf("   Default message: %s\n", message)
	}
	if bulkReq.ExpiresAt != nil {
		fmt.Printf("   Expires on: %s\n", bulkReq.ExpiresAt.Format("2006-01-02"))
	} else {
		fmt.Printf("   Expires in: %s\n", expiresIn)
	}
	fmt.Println()

	// Send bulk invitation request
	resp, err := a.apiClient.SendBulkInvitation(a.ctx, projectID, bulkReq)
	if err != nil {
		return fmt.Errorf("failed to send bulk invitations: %w", err)
	}

	// Display formatted results
	summary := invitation.FormatSummary(resp)
	fmt.Println(summary)

	// Return error if all failed
	if resp.Summary.Failed == resp.Summary.Total {
		return fmt.Errorf("all invitations failed")
	}

	return nil
}

// Helper functions for displaying invitation information
func formatInvitationStatus(status types.InvitationStatus) string {
	switch status {
	case types.InvitationPending:
		return "⏳ Pending"
	case types.InvitationAccepted:
		return "✅ Accepted"
	case types.InvitationDeclined:
		return "❌ Declined"
	case types.InvitationExpired:
		return "⏰ Expired"
	case types.InvitationRevoked:
		return "🚫 Revoked"
	default:
		return string(status)
	}
}

func formatInvitationRole(role types.ProjectRole) string {
	roleNames := map[types.ProjectRole]string{
		types.ProjectRoleOwner:  "👑 Owner",
		types.ProjectRoleAdmin:  "⚙️  Admin",
		types.ProjectRoleMember: "👤 Member",
		types.ProjectRoleViewer: "👁️  Viewer",
	}

	if name, ok := roleNames[role]; ok {
		return name
	}
	return string(role)
}

func formatTimeRemaining(expiresAt time.Time) string {
	remaining := time.Until(expiresAt)

	if remaining < 0 {
		return "Expired"
	}

	days := int(remaining.Hours() / 24)
	hours := int(remaining.Hours()) % 24

	if days > 0 {
		return fmt.Sprintf("%d days, %d hours", days, hours)
	}
	if hours > 0 {
		return fmt.Sprintf("%d hours", hours)
	}
	return "Less than 1 hour"
}

func formatTimeAgo(t time.Time) string {
	elapsed := time.Since(t)

	if elapsed < 0 {
		return "In the future"
	}

	days := int(elapsed.Hours() / 24)
	hours := int(elapsed.Hours()) % 24

	if days > 0 {
		return fmt.Sprintf("%d days ago", days)
	}
	if hours > 0 {
		return fmt.Sprintf("%d hours ago", hours)
	}
	return "Less than 1 hour ago"
}

// parseDuration parses duration strings like "7d", "48h", "7d12h"
func parseDuration(s string) (time.Duration, error) {
	// Handle simple cases
	if s == "" {
		return 7 * 24 * time.Hour, nil // Default 7 days
	}

	var total time.Duration
	current := ""

	for i := 0; i < len(s); i++ {
		ch := s[i]
		if ch >= '0' && ch <= '9' {
			current += string(ch)
		} else if ch == 'd' || ch == 'h' || ch == 'm' {
			if current == "" {
				return 0, fmt.Errorf("missing number before '%c'", ch)
			}
			num := 0
			fmt.Sscanf(current, "%d", &num)

			switch ch {
			case 'd':
				total += time.Duration(num) * 24 * time.Hour
			case 'h':
				total += time.Duration(num) * time.Hour
			case 'm':
				total += time.Duration(num) * time.Minute
			}
			current = ""
		} else {
			return 0, fmt.Errorf("invalid character '%c' in duration", ch)
		}
	}

	if current != "" {
		return 0, fmt.Errorf("number without unit at end of duration")
	}

	if total == 0 {
		return 0, fmt.Errorf("duration cannot be zero")
	}

	return total, nil
}

// parseAbsoluteDate parses absolute date strings like "2025-12-25", "2025-12-25 15:00"
func parseAbsoluteDate(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, fmt.Errorf("empty date string")
	}

	// Try common date formats
	formats := []string{
		"2006-01-02",          // Date only: 2025-12-25
		"2006-01-02 15:04",    // Date and time: 2025-12-25 15:00
		"2006-01-02 15:04:05", // Date and time with seconds: 2025-12-25 15:00:00
		time.RFC3339,          // ISO 8601: 2025-12-25T15:00:00Z
		time.RFC3339Nano,      // ISO 8601 with nanoseconds
	}

	var lastErr error
	for _, format := range formats {
		t, err := time.Parse(format, s)
		if err == nil {
			// Verify date is in the future
			if t.Before(time.Now()) {
				return time.Time{}, fmt.Errorf("date must be in the future (got %s)", t.Format("2006-01-02 15:04"))
			}
			return t, nil
		}
		lastErr = err
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %w", lastErr)
}
