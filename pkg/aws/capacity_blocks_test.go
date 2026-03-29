package aws

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockCapacityEC2 is a minimal mock that implements the capacity reservation
// subset of EC2ClientInterface used by capacity_blocks.go.
type mockCapacityEC2 struct {
	MockEC2Client
	createOut   *ec2.CreateCapacityReservationOutput
	createErr   error
	listOut     *ec2.DescribeCapacityReservationsOutput
	listErr     error
	cancelOut   *ec2.CancelCapacityReservationOutput
	cancelErr   error
	describeOut *ec2.DescribeCapacityReservationsOutput
	describeErr error
}

func (m *mockCapacityEC2) CreateCapacityReservation(_ context.Context, input *ec2.CreateCapacityReservationInput, _ ...func(*ec2.Options)) (*ec2.CreateCapacityReservationOutput, error) {
	return m.createOut, m.createErr
}

func (m *mockCapacityEC2) DescribeCapacityReservations(_ context.Context, input *ec2.DescribeCapacityReservationsInput, _ ...func(*ec2.Options)) (*ec2.DescribeCapacityReservationsOutput, error) {
	if m.describeOut != nil {
		return m.describeOut, m.describeErr
	}
	return m.listOut, m.listErr
}

func (m *mockCapacityEC2) CancelCapacityReservation(_ context.Context, input *ec2.CancelCapacityReservationInput, _ ...func(*ec2.Options)) (*ec2.CancelCapacityReservationOutput, error) {
	return m.cancelOut, m.cancelErr
}

func newCapacityBlockManager(mock EC2ClientInterface) *Manager {
	return &Manager{ec2: mock}
}

func TestReserveCapacityBlock(t *testing.T) {
	start := time.Now().Add(24 * time.Hour).UTC().Truncate(time.Second)
	end := start.Add(8 * time.Hour)
	mock := &mockCapacityEC2{
		createOut: &ec2.CreateCapacityReservationOutput{
			CapacityReservation: &ec2types.CapacityReservation{
				CapacityReservationId: aws.String("cr-test001"),
				InstanceType:          aws.String("p3.8xlarge"),
				TotalInstanceCount:    aws.Int32(2),
				AvailabilityZone:      aws.String("us-west-2a"),
				StartDate:             aws.Time(start),
				EndDate:               aws.Time(end),
				State:                 ec2types.CapacityReservationStateActive,
			},
		},
	}

	mgr := newCapacityBlockManager(mock)
	req := CapacityBlockRequest{
		InstanceType:     "p3.8xlarge",
		InstanceCount:    2,
		AvailabilityZone: "us-west-2a",
		StartTime:        start,
		DurationHours:    8,
	}

	block, err := mgr.ReserveCapacityBlock(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, "cr-test001", block.ID)
	assert.Equal(t, "p3.8xlarge", block.InstanceType)
	assert.Equal(t, 2, block.InstanceCount)
	assert.Equal(t, "active", block.State)
}

func TestListCapacityBlocks(t *testing.T) {
	now := time.Now().UTC()
	mock := &mockCapacityEC2{
		listOut: &ec2.DescribeCapacityReservationsOutput{
			CapacityReservations: []ec2types.CapacityReservation{
				{
					CapacityReservationId: aws.String("cr-aaa"),
					InstanceType:          aws.String("g5.xlarge"),
					TotalInstanceCount:    aws.Int32(1),
					AvailabilityZone:      aws.String("us-west-2b"),
					StartDate:             aws.Time(now),
					EndDate:               aws.Time(now.Add(4 * time.Hour)),
					State:                 ec2types.CapacityReservationStateActive,
				},
				{
					CapacityReservationId: aws.String("cr-bbb"),
					InstanceType:          aws.String("p4d.24xlarge"),
					TotalInstanceCount:    aws.Int32(4),
					AvailabilityZone:      aws.String("us-west-2c"),
					StartDate:             aws.Time(now),
					EndDate:               aws.Time(now.Add(24 * time.Hour)),
					State:                 ec2types.CapacityReservationStateActive,
				},
			},
		},
	}

	mgr := newCapacityBlockManager(mock)
	blocks, err := mgr.ListCapacityBlocks(context.Background())
	require.NoError(t, err)
	require.Len(t, blocks, 2)
	assert.Equal(t, "cr-aaa", blocks[0].ID)
	assert.Equal(t, "g5.xlarge", blocks[0].InstanceType)
	assert.Equal(t, "cr-bbb", blocks[1].ID)
}

func TestDescribeCapacityBlock(t *testing.T) {
	now := time.Now().UTC()
	mock := &mockCapacityEC2{
		describeOut: &ec2.DescribeCapacityReservationsOutput{
			CapacityReservations: []ec2types.CapacityReservation{
				{
					CapacityReservationId: aws.String("cr-specific"),
					InstanceType:          aws.String("p3.2xlarge"),
					TotalInstanceCount:    aws.Int32(1),
					AvailabilityZone:      aws.String("us-east-1a"),
					StartDate:             aws.Time(now),
					EndDate:               aws.Time(now.Add(2 * time.Hour)),
					State:                 ec2types.CapacityReservationStateActive,
				},
			},
		},
	}

	mgr := newCapacityBlockManager(mock)
	block, err := mgr.DescribeCapacityBlock(context.Background(), "cr-specific")
	require.NoError(t, err)
	assert.Equal(t, "cr-specific", block.ID)
	assert.Equal(t, "p3.2xlarge", block.InstanceType)
}

func TestDescribeCapacityBlock_NotFound(t *testing.T) {
	mock := &mockCapacityEC2{
		describeOut: &ec2.DescribeCapacityReservationsOutput{
			CapacityReservations: []ec2types.CapacityReservation{},
		},
	}

	mgr := newCapacityBlockManager(mock)
	_, err := mgr.DescribeCapacityBlock(context.Background(), "cr-missing")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cr-missing")
}

func TestCancelCapacityBlock(t *testing.T) {
	mock := &mockCapacityEC2{
		cancelOut: &ec2.CancelCapacityReservationOutput{
			Return: aws.Bool(true),
		},
	}

	mgr := newCapacityBlockManager(mock)
	err := mgr.CancelCapacityBlock(context.Background(), "cr-to-cancel")
	assert.NoError(t, err)
}
