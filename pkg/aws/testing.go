package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

// MockSSMClient provides a mock implementation of SSMClientInterface for testing.
// This is exported so it can be used by other packages for testing purposes.
type MockSSMClient struct {
	SendCommandFunc                 func(ctx context.Context, params *ssm.SendCommandInput) (*ssm.SendCommandOutput, error)
	GetCommandInvocationFunc        func(ctx context.Context, params *ssm.GetCommandInvocationInput) (*ssm.GetCommandInvocationOutput, error)
	DescribeInstanceInformationFunc func(ctx context.Context, params *ssm.DescribeInstanceInformationInput) (*ssm.DescribeInstanceInformationOutput, error)
	PutParameterFunc                func(ctx context.Context, params *ssm.PutParameterInput) (*ssm.PutParameterOutput, error)
	GetParameterFunc                func(ctx context.Context, params *ssm.GetParameterInput) (*ssm.GetParameterOutput, error)
	DeleteParameterFunc             func(ctx context.Context, params *ssm.DeleteParameterInput) (*ssm.DeleteParameterOutput, error)
	GetParametersByPathFunc         func(ctx context.Context, params *ssm.GetParametersByPathInput) (*ssm.GetParametersByPathOutput, error)
}

func (m *MockSSMClient) SendCommand(ctx context.Context, params *ssm.SendCommandInput, optFns ...func(*ssm.Options)) (*ssm.SendCommandOutput, error) {
	if m.SendCommandFunc != nil {
		return m.SendCommandFunc(ctx, params)
	}
	return &ssm.SendCommandOutput{}, nil
}

func (m *MockSSMClient) GetCommandInvocation(ctx context.Context, params *ssm.GetCommandInvocationInput, optFns ...func(*ssm.Options)) (*ssm.GetCommandInvocationOutput, error) {
	if m.GetCommandInvocationFunc != nil {
		return m.GetCommandInvocationFunc(ctx, params)
	}
	return &ssm.GetCommandInvocationOutput{}, nil
}

func (m *MockSSMClient) DescribeInstanceInformation(ctx context.Context, params *ssm.DescribeInstanceInformationInput, optFns ...func(*ssm.Options)) (*ssm.DescribeInstanceInformationOutput, error) {
	if m.DescribeInstanceInformationFunc != nil {
		return m.DescribeInstanceInformationFunc(ctx, params)
	}
	return &ssm.DescribeInstanceInformationOutput{}, nil
}

func (m *MockSSMClient) PutParameter(ctx context.Context, params *ssm.PutParameterInput, optFns ...func(*ssm.Options)) (*ssm.PutParameterOutput, error) {
	if m.PutParameterFunc != nil {
		return m.PutParameterFunc(ctx, params)
	}
	return &ssm.PutParameterOutput{}, nil
}

func (m *MockSSMClient) GetParameter(ctx context.Context, params *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error) {
	if m.GetParameterFunc != nil {
		return m.GetParameterFunc(ctx, params)
	}
	return &ssm.GetParameterOutput{}, nil
}

func (m *MockSSMClient) DeleteParameter(ctx context.Context, params *ssm.DeleteParameterInput, optFns ...func(*ssm.Options)) (*ssm.DeleteParameterOutput, error) {
	if m.DeleteParameterFunc != nil {
		return m.DeleteParameterFunc(ctx, params)
	}
	return &ssm.DeleteParameterOutput{}, nil
}

func (m *MockSSMClient) GetParametersByPath(ctx context.Context, params *ssm.GetParametersByPathInput, optFns ...func(*ssm.Options)) (*ssm.GetParametersByPathOutput, error) {
	if m.GetParametersByPathFunc != nil {
		return m.GetParametersByPathFunc(ctx, params)
	}
	return &ssm.GetParametersByPathOutput{}, nil
}
