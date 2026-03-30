package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/stretchr/testify/assert"
)

// TestExtractTags_PrismProjectIDKey verifies that the correct tag key
// "prism:project-id" is used to populate the project field.
// Regression guard: the old key "Project" was silently ignored, making
// instance.ProjectID always empty when reconstructed from AWS.
func TestExtractTags_PrismProjectIDKey(t *testing.T) {
	extractor := &InstanceTagExtractor{}

	ec2Instance := ec2types.Instance{
		Tags: []ec2types.Tag{
			{Key: aws.String("Name"), Value: aws.String("my-instance")},
			{Key: aws.String("Template"), Value: aws.String("python-ml")},
			{Key: aws.String("prism:project-id"), Value: aws.String("proj-abc123")},
		},
	}

	name, template, project := extractor.ExtractTags(ec2Instance)

	assert.Equal(t, "my-instance", name)
	assert.Equal(t, "python-ml", template)
	assert.Equal(t, "proj-abc123", project, "prism:project-id tag must populate project field")
}

// TestExtractTags_OldProjectKeyIgnored verifies that the legacy tag key
// "Project" (capital P) is NOT mapped to the project field — confirming the
// regression does not re-appear.
func TestExtractTags_OldProjectKeyIgnored(t *testing.T) {
	extractor := &InstanceTagExtractor{}

	ec2Instance := ec2types.Instance{
		Tags: []ec2types.Tag{
			{Key: aws.String("Name"), Value: aws.String("old-instance")},
			{Key: aws.String("Project"), Value: aws.String("should-be-ignored")},
		},
	}

	_, _, project := extractor.ExtractTags(ec2Instance)

	assert.Equal(t, "", project, "legacy 'Project' key must not populate project field")
}

// TestExtractTags_NoTags handles an instance with no tags.
func TestExtractTags_NoTags(t *testing.T) {
	extractor := &InstanceTagExtractor{}

	ec2Instance := ec2types.Instance{
		Tags: []ec2types.Tag{},
	}

	name, template, project := extractor.ExtractTags(ec2Instance)

	assert.Equal(t, "", name)
	assert.Equal(t, "", template)
	assert.Equal(t, "", project)
}

// TestExtractTags_NilKeyOrValue handles tags with nil Key or Value (defensive).
func TestExtractTags_NilKeyOrValue(t *testing.T) {
	extractor := &InstanceTagExtractor{}

	ec2Instance := ec2types.Instance{
		Tags: []ec2types.Tag{
			{Key: nil, Value: aws.String("value-no-key")},
			{Key: aws.String("prism:project-id"), Value: nil},
			{Key: aws.String("Name"), Value: aws.String("safe-instance")},
		},
	}

	name, template, project := extractor.ExtractTags(ec2Instance)

	assert.Equal(t, "safe-instance", name)
	assert.Equal(t, "", template)
	assert.Equal(t, "", project, "nil Value for prism:project-id must not panic or populate field")
}
