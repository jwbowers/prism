package storage

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAnalyticsManager(t *testing.T) {
	mgr := NewAnalyticsManager(aws.Config{})
	require.NotNil(t, mgr, "NewAnalyticsManager should return a non-nil manager")
}

func TestAnalyticsPeriodConstants(t *testing.T) {
	assert.Equal(t, AnalyticsPeriod("daily"), AnalyticsPeriodDaily)
	assert.Equal(t, AnalyticsPeriod("weekly"), AnalyticsPeriodWeekly)
	assert.Equal(t, AnalyticsPeriod("monthly"), AnalyticsPeriodMonthly)
}

func TestStorageTypeConstants(t *testing.T) {
	assert.Equal(t, StorageType("efs"), StorageTypeEFS)
	assert.Equal(t, StorageType("ebs"), StorageTypeEBS)
	assert.Equal(t, StorageType("s3"), StorageTypeS3)
}

func TestAnalyticsRequestZeroValue(t *testing.T) {
	// Zero-value AnalyticsRequest should be constructible without panic.
	req := AnalyticsRequest{}
	assert.Equal(t, AnalyticsPeriod(""), req.Period)
	assert.True(t, req.StartTime.IsZero())
	assert.True(t, req.EndTime.IsZero())
	assert.Nil(t, req.Resources)
}

func TestAnalyticsRequestWithResources(t *testing.T) {
	now := time.Now().UTC()
	req := AnalyticsRequest{
		Period:    AnalyticsPeriodDaily,
		StartTime: now.AddDate(0, 0, -1),
		EndTime:   now,
		Resources: []StorageResource{
			{Name: "my-efs", Type: StorageTypeEFS, ResourceID: "fs-abc123"},
			{Name: "my-ebs", Type: StorageTypeEBS, ResourceID: "vol-def456"},
		},
	}
	assert.Equal(t, 2, len(req.Resources))
	assert.Equal(t, "my-efs", req.Resources[0].Name)
	assert.Equal(t, StorageTypeEFS, req.Resources[0].Type)
	assert.Equal(t, "fs-abc123", req.Resources[0].ResourceID)
}
