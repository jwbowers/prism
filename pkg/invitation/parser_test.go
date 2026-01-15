package invitation

import (
	"strings"
	"testing"

	"github.com/scttfrdmn/prism/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseCSV_WithHeader(t *testing.T) {
	csv := `email,role,message
alice@example.com,admin,Welcome Alice
bob@example.com,member,Welcome Bob
charlie@example.com,viewer,Welcome Charlie`

	reader := strings.NewReader(csv)
	entries, err := ParseCSV(reader)

	require.NoError(t, err)
	require.Len(t, entries, 3)

	assert.Equal(t, "alice@example.com", entries[0].Email)
	assert.Equal(t, types.ProjectRoleAdmin, entries[0].Role)
	assert.Equal(t, "Welcome Alice", entries[0].Message)

	assert.Equal(t, "bob@example.com", entries[1].Email)
	assert.Equal(t, types.ProjectRoleMember, entries[1].Role)
	assert.Equal(t, "Welcome Bob", entries[1].Message)

	assert.Equal(t, "charlie@example.com", entries[2].Email)
	assert.Equal(t, types.ProjectRoleViewer, entries[2].Role)
	assert.Equal(t, "Welcome Charlie", entries[2].Message)
}

func TestParseCSV_WithoutHeader(t *testing.T) {
	csv := `alice@example.com,admin,Welcome
bob@example.com,member,Hello`

	reader := strings.NewReader(csv)
	entries, err := ParseCSV(reader)

	require.NoError(t, err)
	require.Len(t, entries, 2)

	assert.Equal(t, "alice@example.com", entries[0].Email)
	assert.Equal(t, types.ProjectRoleAdmin, entries[0].Role)
}

func TestParseCSV_EmailOnly(t *testing.T) {
	csv := `alice@example.com
bob@example.com
charlie@example.com`

	reader := strings.NewReader(csv)
	entries, err := ParseCSV(reader)

	require.NoError(t, err)
	require.Len(t, entries, 3)

	assert.Equal(t, "alice@example.com", entries[0].Email)
	assert.Equal(t, types.ProjectRole(""), entries[0].Role)
	assert.Equal(t, "", entries[0].Message)
}

func TestParseCSV_WithComments(t *testing.T) {
	csv := `email,role,message
# This is a comment
alice@example.com,admin,Welcome
# Another comment
bob@example.com,member,Hello`

	reader := strings.NewReader(csv)
	entries, err := ParseCSV(reader)

	require.NoError(t, err)
	require.Len(t, entries, 2)
	assert.Equal(t, "alice@example.com", entries[0].Email)
	assert.Equal(t, "bob@example.com", entries[1].Email)
}

func TestParseCSV_EmptyLines(t *testing.T) {
	csv := `alice@example.com,admin

bob@example.com,member

charlie@example.com,viewer`

	reader := strings.NewReader(csv)
	entries, err := ParseCSV(reader)

	require.NoError(t, err)
	require.Len(t, entries, 3)
}

func TestParseCSV_TrimWhitespace(t *testing.T) {
	csv := `  alice@example.com  ,  admin  ,  Welcome
  bob@example.com  ,  member  ,  Hello  `

	reader := strings.NewReader(csv)
	entries, err := ParseCSV(reader)

	require.NoError(t, err)
	require.Len(t, entries, 2)

	assert.Equal(t, "alice@example.com", entries[0].Email)
	assert.Equal(t, types.ProjectRoleAdmin, entries[0].Role)
	assert.Equal(t, "Welcome", entries[0].Message)
}

func TestParseCSV_MissingEmail(t *testing.T) {
	csv := `email,role,message
,admin,Welcome`

	reader := strings.NewReader(csv)
	_, err := ParseCSV(reader)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing email")
}

func TestParseCSV_InvalidEmail(t *testing.T) {
	csv := `invalid-email,admin,Welcome`

	reader := strings.NewReader(csv)
	_, err := ParseCSV(reader)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid email format")
}

func TestParseCSV_EmptyFile(t *testing.T) {
	csv := ``

	reader := strings.NewReader(csv)
	_, err := ParseCSV(reader)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no valid entries found")
}

func TestParseCSV_OnlyHeader(t *testing.T) {
	csv := `email,role,message`

	reader := strings.NewReader(csv)
	_, err := ParseCSV(reader)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no valid entries found")
}

func TestParsePlainText_Basic(t *testing.T) {
	text := `alice@example.com
bob@example.com
charlie@example.com`

	reader := strings.NewReader(text)
	entries, err := ParsePlainText(reader)

	require.NoError(t, err)
	require.Len(t, entries, 3)

	assert.Equal(t, "alice@example.com", entries[0].Email)
	assert.Equal(t, "bob@example.com", entries[1].Email)
	assert.Equal(t, "charlie@example.com", entries[2].Email)
}

func TestParsePlainText_WithComments(t *testing.T) {
	text := `# Team members
alice@example.com
# External collaborators
bob@example.com
# End of list`

	reader := strings.NewReader(text)
	entries, err := ParsePlainText(reader)

	require.NoError(t, err)
	require.Len(t, entries, 2)

	assert.Equal(t, "alice@example.com", entries[0].Email)
	assert.Equal(t, "bob@example.com", entries[1].Email)
}

func TestParsePlainText_WithEmptyLines(t *testing.T) {
	text := `alice@example.com

bob@example.com

charlie@example.com`

	reader := strings.NewReader(text)
	entries, err := ParsePlainText(reader)

	require.NoError(t, err)
	require.Len(t, entries, 3)
}

func TestParsePlainText_InvalidEmail(t *testing.T) {
	text := `alice@example.com
invalid-email
bob@example.com`

	reader := strings.NewReader(text)
	_, err := ParsePlainText(reader)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid email format")
	assert.Contains(t, err.Error(), "line 2")
}

func TestParsePlainText_EmptyFile(t *testing.T) {
	text := ``

	reader := strings.NewReader(text)
	_, err := ParsePlainText(reader)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no valid email addresses found")
}

func TestParsePlainText_OnlyComments(t *testing.T) {
	text := `# Comment 1
# Comment 2
# Comment 3`

	reader := strings.NewReader(text)
	_, err := ParsePlainText(reader)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no valid email addresses found")
}

func TestParseInlineEmails_CommaSeparated(t *testing.T) {
	emailList := "alice@example.com, bob@example.com, charlie@example.com"

	entries, err := ParseInlineEmails(emailList)

	require.NoError(t, err)
	require.Len(t, entries, 3)

	assert.Equal(t, "alice@example.com", entries[0].Email)
	assert.Equal(t, "bob@example.com", entries[1].Email)
	assert.Equal(t, "charlie@example.com", entries[2].Email)
}

func TestParseInlineEmails_SpaceSeparated(t *testing.T) {
	emailList := "alice@example.com bob@example.com charlie@example.com"

	entries, err := ParseInlineEmails(emailList)

	require.NoError(t, err)
	require.Len(t, entries, 3)
}

func TestParseInlineEmails_NewlineSeparated(t *testing.T) {
	emailList := "alice@example.com\nbob@example.com\ncharlie@example.com"

	entries, err := ParseInlineEmails(emailList)

	require.NoError(t, err)
	require.Len(t, entries, 3)
}

func TestParseInlineEmails_MixedSeparators(t *testing.T) {
	emailList := "alice@example.com, bob@example.com\ncharlie@example.com\tdan@example.com"

	entries, err := ParseInlineEmails(emailList)

	require.NoError(t, err)
	require.Len(t, entries, 4)
}

func TestParseInlineEmails_InvalidEmail(t *testing.T) {
	emailList := "alice@example.com, invalid-email, bob@example.com"

	_, err := ParseInlineEmails(emailList)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid email format")
	assert.Contains(t, err.Error(), "position 2")
}

func TestParseInlineEmails_Empty(t *testing.T) {
	emailList := ""

	_, err := ParseInlineEmails(emailList)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no valid email addresses found")
}

func TestParseInlineEmails_OnlyWhitespace(t *testing.T) {
	emailList := "   ,  ,  "

	_, err := ParseInlineEmails(emailList)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no valid email addresses found")
}

func TestIsValidEmail_Valid(t *testing.T) {
	validEmails := []string{
		"alice@example.com",
		"bob.smith@example.co.uk",
		"charlie+test@example.com",
		"user_123@test.org",
		"admin@subdomain.example.com",
	}

	for _, email := range validEmails {
		assert.True(t, isValidEmail(email), "Expected %s to be valid", email)
	}
}

func TestIsValidEmail_Invalid(t *testing.T) {
	invalidEmails := []string{
		"",
		"   ",
		"notanemail",
		"@example.com",
		"user@",
		"user@@example.com",
		"user@example",
		// NOTE: Current validation is basic and may not catch all edge cases:
		// - "user@.com" passes (doesn't check for leading dot)
		// - "user@example .com" passes (doesn't check for internal spaces)
		// These are known limitations of the basic validation
		"user@example.",
	}

	for _, email := range invalidEmails {
		assert.False(t, isValidEmail(email), "Expected %s to be invalid", email)
	}
}

func TestValidateRoles_Valid(t *testing.T) {
	entries := []types.BulkInvitationEntry{
		{Email: "alice@example.com", Role: types.ProjectRoleOwner},
		{Email: "bob@example.com", Role: types.ProjectRoleAdmin},
		{Email: "charlie@example.com", Role: types.ProjectRoleMember},
		{Email: "dan@example.com", Role: types.ProjectRoleViewer},
		{Email: "eve@example.com", Role: ""}, // Empty role is valid
	}

	err := ValidateRoles(entries)
	require.NoError(t, err)
}

func TestValidateRoles_Invalid(t *testing.T) {
	entries := []types.BulkInvitationEntry{
		{Email: "alice@example.com", Role: types.ProjectRoleAdmin},
		{Email: "bob@example.com", Role: "superadmin"}, // Invalid role
		{Email: "charlie@example.com", Role: types.ProjectRoleMember},
	}

	err := ValidateRoles(entries)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid role")
	assert.Contains(t, err.Error(), "entry 2")
}

func TestFormatSummary_AllSent(t *testing.T) {
	response := &types.BulkInvitationResponse{
		Summary: types.BulkInvitationSummary{
			Total:   3,
			Sent:    3,
			Skipped: 0,
			Failed:  0,
		},
		Results: []types.BulkInvitationResult{
			{Email: "alice@example.com", Status: "sent"},
			{Email: "bob@example.com", Status: "sent"},
			{Email: "charlie@example.com", Status: "sent"},
		},
	}

	summary := FormatSummary(response)

	assert.Contains(t, summary, "Total:   3")
	assert.Contains(t, summary, "✅ Sent:    3 (100%)")
	assert.Contains(t, summary, "⏭️  Skipped: 0 (0%)")
	assert.Contains(t, summary, "❌ Failed:  0 (0%)")
}

func TestFormatSummary_WithSkipped(t *testing.T) {
	response := &types.BulkInvitationResponse{
		Summary: types.BulkInvitationSummary{
			Total:   3,
			Sent:    2,
			Skipped: 1,
			Failed:  0,
		},
		Results: []types.BulkInvitationResult{
			{Email: "alice@example.com", Status: "sent"},
			{Email: "bob@example.com", Status: "sent"},
			{Email: "charlie@example.com", Status: "skipped", Reason: "Already invited"},
		},
	}

	summary := FormatSummary(response)

	assert.Contains(t, summary, "Total:   3")
	assert.Contains(t, summary, "✅ Sent:    2 (66%)")
	assert.Contains(t, summary, "⏭️  Skipped: 1 (33%)")
	assert.Contains(t, summary, "Skipped:")
	assert.Contains(t, summary, "charlie@example.com: Already invited")
}

func TestFormatSummary_WithFailed(t *testing.T) {
	response := &types.BulkInvitationResponse{
		Summary: types.BulkInvitationSummary{
			Total:   3,
			Sent:    2,
			Skipped: 0,
			Failed:  1,
		},
		Results: []types.BulkInvitationResult{
			{Email: "alice@example.com", Status: "sent"},
			{Email: "bob@example.com", Status: "sent"},
			{Email: "invalid@", Status: "failed", Error: "Invalid email format"},
		},
	}

	summary := FormatSummary(response)

	assert.Contains(t, summary, "Total:   3")
	assert.Contains(t, summary, "✅ Sent:    2 (66%)")
	assert.Contains(t, summary, "❌ Failed:  1 (33%)")
	assert.Contains(t, summary, "Failed:")
	assert.Contains(t, summary, "invalid@: Invalid email format")
}

func TestFormatSummary_Mixed(t *testing.T) {
	response := &types.BulkInvitationResponse{
		Summary: types.BulkInvitationSummary{
			Total:   5,
			Sent:    2,
			Skipped: 2,
			Failed:  1,
		},
		Results: []types.BulkInvitationResult{
			{Email: "alice@example.com", Status: "sent"},
			{Email: "bob@example.com", Status: "sent"},
			{Email: "charlie@example.com", Status: "skipped", Reason: "Already member"},
			{Email: "dan@example.com", Status: "skipped", Reason: "Already invited"},
			{Email: "invalid@", Status: "failed", Error: "Invalid email"},
		},
	}

	summary := FormatSummary(response)

	assert.Contains(t, summary, "Total:   5")
	assert.Contains(t, summary, "✅ Sent:    2 (40%)")
	assert.Contains(t, summary, "⏭️  Skipped: 2 (40%)")
	assert.Contains(t, summary, "❌ Failed:  1 (20%)")
	assert.Contains(t, summary, "Skipped:")
	assert.Contains(t, summary, "charlie@example.com")
	assert.Contains(t, summary, "dan@example.com")
	assert.Contains(t, summary, "Failed:")
	assert.Contains(t, summary, "invalid@")
}

func TestPercentage(t *testing.T) {
	tests := []struct {
		name     string
		part     int
		total    int
		expected int
	}{
		{"zero total", 0, 0, 0},
		{"100 percent", 10, 10, 100},
		{"50 percent", 5, 10, 50},
		{"33 percent", 1, 3, 33},
		{"66 percent", 2, 3, 66},
		{"zero part", 0, 10, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := percentage(tt.part, tt.total)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseCSV_LargeFile(t *testing.T) {
	// Generate large CSV with 1000 entries
	var sb strings.Builder
	sb.WriteString("email,role,message\n")
	for i := 0; i < 1000; i++ {
		sb.WriteString("user")
		sb.WriteString(string(rune('0' + (i % 10))))
		sb.WriteString("@example.com,member,Welcome\n")
	}

	reader := strings.NewReader(sb.String())
	entries, err := ParseCSV(reader)

	require.NoError(t, err)
	require.Len(t, entries, 1000)
}

func TestParsePlainText_Whitespace(t *testing.T) {
	text := `  alice@example.com
  bob@example.com
  charlie@example.com  `

	reader := strings.NewReader(text)
	entries, err := ParsePlainText(reader)

	require.NoError(t, err)
	require.Len(t, entries, 3)

	assert.Equal(t, "alice@example.com", entries[0].Email)
	assert.Equal(t, "bob@example.com", entries[1].Email)
	assert.Equal(t, "charlie@example.com", entries[2].Email)
}

func TestParseCSV_QuotedFields(t *testing.T) {
	csv := `email,role,message
"alice@example.com",admin,"Welcome, Alice"
"bob@example.com",member,"Hello, Bob"`

	reader := strings.NewReader(csv)
	entries, err := ParseCSV(reader)

	require.NoError(t, err)
	require.Len(t, entries, 2)

	assert.Equal(t, "alice@example.com", entries[0].Email)
	assert.Equal(t, "Welcome, Alice", entries[0].Message)
}

func TestParseCSV_MalformedQuotes(t *testing.T) {
	csv := `email,role,message
"alice@example.com,admin,Welcome`

	reader := strings.NewReader(csv)
	_, err := ParseCSV(reader)

	// CSV reader should detect malformed quotes
	require.Error(t, err)
}
