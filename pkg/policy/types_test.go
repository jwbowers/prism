package policy

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestPolicyType_Constants(t *testing.T) {
	assert.Equal(t, PolicyType("template_access"), PolicyTypeTemplateAccess)
	assert.Equal(t, PolicyType("resource_limits"), PolicyTypeResourceLimits)
	assert.Equal(t, PolicyType("research_user"), PolicyTypeResearchUser)
	assert.Equal(t, PolicyType("instance"), PolicyTypeInstance)
}

func TestPolicyEffect_Constants(t *testing.T) {
	assert.Equal(t, PolicyEffect("allow"), PolicyEffectAllow)
	assert.Equal(t, PolicyEffect("deny"), PolicyEffectDeny)
}

func TestPolicy_JSON(t *testing.T) {
	now := time.Now()
	policy := Policy{
		ID:          "policy-1",
		Name:        "Test Policy",
		Description: "A test policy",
		Type:        PolicyTypeTemplateAccess,
		Effect:      PolicyEffectAllow,
		Conditions: map[string]interface{}{
			"user_group": "researchers",
		},
		Resources: []string{"template:python-ml"},
		Actions:   []string{"read", "launch"},
		CreatedAt: now,
		UpdatedAt: now,
		Enabled:   true,
	}

	// Marshal to JSON
	data, err := json.Marshal(policy)
	require.NoError(t, err)

	// Unmarshal back
	var decoded Policy
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, policy.ID, decoded.ID)
	assert.Equal(t, policy.Name, decoded.Name)
	assert.Equal(t, policy.Description, decoded.Description)
	assert.Equal(t, policy.Type, decoded.Type)
	assert.Equal(t, policy.Effect, decoded.Effect)
	assert.Equal(t, policy.Resources, decoded.Resources)
	assert.Equal(t, policy.Actions, decoded.Actions)
	assert.Equal(t, policy.Enabled, decoded.Enabled)
}

func TestPolicy_YAML(t *testing.T) {
	now := time.Now()
	policy := Policy{
		ID:        "policy-1",
		Name:      "Test Policy",
		Type:      PolicyTypeResourceLimits,
		Effect:    PolicyEffectDeny,
		Resources: []string{"instance:*"},
		Actions:   []string{"launch"},
		CreatedAt: now,
		UpdatedAt: now,
		Enabled:   true,
	}

	// Marshal to YAML
	data, err := yaml.Marshal(policy)
	require.NoError(t, err)

	// Unmarshal back
	var decoded Policy
	err = yaml.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, policy.ID, decoded.ID)
	assert.Equal(t, policy.Name, decoded.Name)
	assert.Equal(t, policy.Type, decoded.Type)
	assert.Equal(t, policy.Effect, decoded.Effect)
}

func TestPolicySet_JSON(t *testing.T) {
	now := time.Now()
	policy1 := &Policy{
		ID:     "policy-1",
		Name:   "Allow ML Templates",
		Type:   PolicyTypeTemplateAccess,
		Effect: PolicyEffectAllow,
	}
	policy2 := &Policy{
		ID:     "policy-2",
		Name:   "Deny Expensive Instances",
		Type:   PolicyTypeResourceLimits,
		Effect: PolicyEffectDeny,
	}

	policySet := PolicySet{
		ID:          "set-1",
		Name:        "Researcher Policy Set",
		Description: "Standard policies for researchers",
		Policies:    []*Policy{policy1, policy2},
		Tags: map[string]string{
			"environment": "production",
			"team":        "research",
		},
		CreatedAt: now,
		UpdatedAt: now,
		Enabled:   true,
	}

	// Marshal to JSON
	data, err := json.Marshal(policySet)
	require.NoError(t, err)

	// Unmarshal back
	var decoded PolicySet
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, policySet.ID, decoded.ID)
	assert.Equal(t, policySet.Name, decoded.Name)
	assert.Equal(t, policySet.Description, decoded.Description)
	assert.Len(t, decoded.Policies, 2)
	assert.Equal(t, policySet.Tags, decoded.Tags)
	assert.Equal(t, policySet.Enabled, decoded.Enabled)
}

func TestPolicyRequest_JSON(t *testing.T) {
	request := PolicyRequest{
		UserID:   "user-123",
		Action:   "launch",
		Resource: "template:python-ml",
		Context: map[string]interface{}{
			"region":        "us-west-2",
			"instance_type": "t3.large",
		},
		ProfileID: "profile-1",
	}

	// Marshal to JSON
	data, err := json.Marshal(request)
	require.NoError(t, err)

	// Unmarshal back
	var decoded PolicyRequest
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, request.UserID, decoded.UserID)
	assert.Equal(t, request.Action, decoded.Action)
	assert.Equal(t, request.Resource, decoded.Resource)
	assert.Equal(t, request.ProfileID, decoded.ProfileID)
	assert.NotNil(t, decoded.Context)
}

func TestPolicyResponse_JSON(t *testing.T) {
	response := PolicyResponse{
		Allowed:         true,
		Reason:          "User has required permissions",
		MatchedPolicies: []string{"policy-1", "policy-2"},
		Suggestions:     []string{"Consider using spot instances"},
	}

	// Marshal to JSON
	data, err := json.Marshal(response)
	require.NoError(t, err)

	// Unmarshal back
	var decoded PolicyResponse
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, response.Allowed, decoded.Allowed)
	assert.Equal(t, response.Reason, decoded.Reason)
	assert.Equal(t, response.MatchedPolicies, decoded.MatchedPolicies)
	assert.Equal(t, response.Suggestions, decoded.Suggestions)
}

func TestTemplateAccessPolicy_JSON(t *testing.T) {
	policy := TemplateAccessPolicy{
		AllowedTemplates: []string{"python-ml", "r-research"},
		DeniedTemplates:  []string{"bitcoin-miner"},
		RequiredDomain:   "university.edu",
		MaxComplexity:    "high",
	}

	// Marshal to JSON
	data, err := json.Marshal(policy)
	require.NoError(t, err)

	// Unmarshal back
	var decoded TemplateAccessPolicy
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, policy.AllowedTemplates, decoded.AllowedTemplates)
	assert.Equal(t, policy.DeniedTemplates, decoded.DeniedTemplates)
	assert.Equal(t, policy.RequiredDomain, decoded.RequiredDomain)
	assert.Equal(t, policy.MaxComplexity, decoded.MaxComplexity)
}

func TestResourceLimitsPolicy_JSON(t *testing.T) {
	policy := ResourceLimitsPolicy{
		MaxInstances:     5,
		MaxInstanceTypes: []string{"t3.medium", "t3.large"},
		MaxCostPerHour:   2.50,
		MaxVolumes:       3,
		AllowedRegions:   []string{"us-west-2", "us-east-1"},
		RequireSpot:      true,
		Tags: map[string]string{
			"project": "research",
			"owner":   "lab-admin",
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(policy)
	require.NoError(t, err)

	// Unmarshal back
	var decoded ResourceLimitsPolicy
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, policy.MaxInstances, decoded.MaxInstances)
	assert.Equal(t, policy.MaxInstanceTypes, decoded.MaxInstanceTypes)
	assert.Equal(t, policy.MaxCostPerHour, decoded.MaxCostPerHour)
	assert.Equal(t, policy.MaxVolumes, decoded.MaxVolumes)
	assert.Equal(t, policy.AllowedRegions, decoded.AllowedRegions)
	assert.Equal(t, policy.RequireSpot, decoded.RequireSpot)
	assert.Equal(t, policy.Tags, decoded.Tags)
}

func TestResearchUserPolicy_JSON(t *testing.T) {
	policy := ResearchUserPolicy{
		AllowCreation:     true,
		AllowDeletion:     false,
		RequireApproval:   true,
		MaxUsers:          10,
		AllowedShells:     []string{"/bin/bash", "/bin/zsh"},
		RequiredGroups:    []string{"researchers", "students"},
		AllowSSHKeys:      true,
		AllowSudoAccess:   false,
		AllowDockerAccess: true,
	}

	// Marshal to JSON
	data, err := json.Marshal(policy)
	require.NoError(t, err)

	// Unmarshal back
	var decoded ResearchUserPolicy
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, policy.AllowCreation, decoded.AllowCreation)
	assert.Equal(t, policy.AllowDeletion, decoded.AllowDeletion)
	assert.Equal(t, policy.RequireApproval, decoded.RequireApproval)
	assert.Equal(t, policy.MaxUsers, decoded.MaxUsers)
	assert.Equal(t, policy.AllowedShells, decoded.AllowedShells)
	assert.Equal(t, policy.RequiredGroups, decoded.RequiredGroups)
	assert.Equal(t, policy.AllowSSHKeys, decoded.AllowSSHKeys)
	assert.Equal(t, policy.AllowSudoAccess, decoded.AllowSudoAccess)
	assert.Equal(t, policy.AllowDockerAccess, decoded.AllowDockerAccess)
}

func TestPolicy_Enabled(t *testing.T) {
	// Test enabled policy
	enabledPolicy := Policy{
		ID:      "policy-1",
		Name:    "Enabled Policy",
		Type:    PolicyTypeTemplateAccess,
		Effect:  PolicyEffectAllow,
		Enabled: true,
	}
	assert.True(t, enabledPolicy.Enabled)

	// Test disabled policy
	disabledPolicy := Policy{
		ID:      "policy-2",
		Name:    "Disabled Policy",
		Type:    PolicyTypeTemplateAccess,
		Effect:  PolicyEffectDeny,
		Enabled: false,
	}
	assert.False(t, disabledPolicy.Enabled)
}

func TestPolicySet_Enabled(t *testing.T) {
	policySet := PolicySet{
		ID:       "set-1",
		Name:     "Test Set",
		Policies: []*Policy{},
		Enabled:  true,
	}
	assert.True(t, policySet.Enabled)

	policySet.Enabled = false
	assert.False(t, policySet.Enabled)
}

func TestPolicyResponse_Allowed(t *testing.T) {
	// Test allowed response
	allowedResponse := PolicyResponse{
		Allowed: true,
		Reason:  "Policy allows this action",
	}
	assert.True(t, allowedResponse.Allowed)

	// Test denied response
	deniedResponse := PolicyResponse{
		Allowed: false,
		Reason:  "Policy denies this action",
	}
	assert.False(t, deniedResponse.Allowed)
}
