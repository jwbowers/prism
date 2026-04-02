//go:build substrate
// +build substrate

package aws

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/efs"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	ctypes "github.com/scttfrdmn/prism/pkg/types"
	"github.com/scttfrdmn/substrate"
)

// setupSubstrateManager starts an in-process Substrate server and returns a
// Manager wired to it. No Docker, no external processes — the server is
// automatically shut down when the test ends via t.Cleanup().
func setupSubstrateManager(t *testing.T) (*Manager, *substrate.TestServer) {
	t.Helper()

	ts := substrate.StartTestServer(t)

	cfg := aws.Config{
		Region: "us-east-1",
		EndpointResolverWithOptions: aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{
					URL:               ts.URL,
					SigningRegion:     "us-east-1",
					HostnameImmutable: true,
				}, nil
			},
		),
		Credentials: credentials.NewStaticCredentialsProvider("test", "test", ""),
	}

	manager := &Manager{
		cfg:            cfg,
		ec2:            ec2.NewFromConfig(cfg),
		efs:            efs.NewFromConfig(cfg),
		iam:            iam.NewFromConfig(cfg),
		ssm:            ssm.NewFromConfig(cfg),
		sts:            sts.NewFromConfig(cfg),
		region:         "us-east-1",
		templates:      getTemplates(),
		discountConfig: ctypes.DiscountConfig{},
	}

	return manager, ts
}

func TestSubstrateLaunchInstance(t *testing.T) {
	// LaunchInstance resolves AMI IDs via SSM GetParameter against public AWS paths
	// (/aws/service/canonical/ubuntu/..., /aws/service/ami-amazon-linux-latest/...).
	// A fresh TestServer has no SSM parameters seeded. Re-enable once
	// scttfrdmn/substrate#267 (TestServer.SeedSSMParameter helper) is implemented,
	// then call ts.SeedSSMParameter(...) for each AMI path before launching.
	t.Skip("Skipping: waiting for substrate#267 (TestServer.SeedSSMParameter helper)")
	manager, _ := setupSubstrateManager(t)

	req := ctypes.LaunchRequest{
		Template: "basic-ubuntu",
		Name:     "test-instance-1",
		Size:     "M",
		Region:   "us-east-1",
		DryRun:   false,
	}

	instance, err := manager.LaunchInstance(req)
	if err != nil {
		t.Fatalf("LaunchInstance failed: %v", err)
	}
	if instance.Name != req.Name {
		t.Errorf("Instance name = %s, want %s", instance.Name, req.Name)
	}
	if instance.ID == "" {
		t.Error("Instance ID should not be empty")
	}
	if instance.HourlyRate <= 0 {
		t.Error("Instance should have positive hourly rate")
	}

	t.Cleanup(func() {
		if err := manager.DeleteInstance(req.Name); err != nil {
			t.Logf("cleanup: DeleteInstance: %v", err)
		}
	})
}

func TestSubstrateCreateEBSVolume(t *testing.T) {
	manager, _ := setupSubstrateManager(t)

	req := ctypes.StorageCreateRequest{
		Name:       "test-ebs-volume",
		Size:       "100",
		VolumeType: "gp3",
		Region:     "us-east-1",
	}

	volume, err := manager.CreateStorage(req)
	if err != nil {
		t.Fatalf("CreateStorage failed: %v", err)
	}
	if volume.Name != req.Name {
		t.Errorf("Volume name = %s, want %s", volume.Name, req.Name)
	}
	if volume.VolumeID == "" {
		t.Error("Volume should have a volume ID")
	}
	if volume.VolumeType != req.VolumeType {
		t.Errorf("Volume type = %s, want %s", volume.VolumeType, req.VolumeType)
	}
	if volume.SizeGB == nil || *volume.SizeGB != 100 {
		t.Errorf("Volume size = %v, want 100", volume.SizeGB)
	}

	t.Cleanup(func() {
		if err := manager.DeleteStorage(req.Name); err != nil {
			t.Logf("cleanup: DeleteStorage: %v", err)
		}
	})
}

func TestSubstrateEBSAttachDetach(t *testing.T) {
	// Depends on LaunchInstance. Re-enable alongside TestSubstrateLaunchInstance
	// once substrate#267 (TestServer.SeedSSMParameter) is implemented.
	t.Skip("Skipping: depends on LaunchInstance (blocked on substrate#267)")
	manager, ts := setupSubstrateManager(t)

	// Launch an instance
	launchReq := ctypes.LaunchRequest{
		Template: "basic-ubuntu",
		Name:     "attach-test-instance",
		Size:     "M",
		Region:   "us-east-1",
	}
	instance, err := manager.LaunchInstance(launchReq)
	if err != nil {
		t.Fatalf("LaunchInstance failed: %v", err)
	}
	t.Cleanup(func() { manager.DeleteInstance(launchReq.Name) })

	// Create a volume
	volReq := ctypes.StorageCreateRequest{
		Name:       "attach-test-vol",
		Size:       "50",
		VolumeType: "gp3",
		Region:     "us-east-1",
	}
	vol, err := manager.CreateStorage(volReq)
	if err != nil {
		t.Fatalf("CreateStorage failed: %v", err)
	}
	t.Cleanup(func() { manager.DeleteStorage(volReq.Name) })

	// Attach
	if err := manager.AttachStorage(instance.Name, vol.Name); err != nil {
		t.Fatalf("AttachStorage failed: %v", err)
	}

	// Detach
	if err := manager.DetachStorage(vol.Name); err != nil {
		t.Fatalf("DetachStorage failed: %v", err)
	}

	// Prevent unused warning
	_ = ts
}

func TestSubstrateIAMInstanceProfile(t *testing.T) {
	_, ts := setupSubstrateManager(t)

	ctx := context.Background()
	cfg := aws.Config{
		Region: "us-east-1",
		EndpointResolverWithOptions: aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{URL: ts.URL, HostnameImmutable: true, SigningRegion: "us-east-1"}, nil
			},
		),
		Credentials: credentials.NewStaticCredentialsProvider("test", "test", ""),
	}
	iamClient := iam.NewFromConfig(cfg)

	// Create role
	_, err := iamClient.CreateRole(ctx, &iam.CreateRoleInput{
		RoleName:                 aws.String("test-ec2-role"),
		AssumeRolePolicyDocument: aws.String(`{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"Service":"ec2.amazonaws.com"},"Action":"sts:AssumeRole"}]}`),
	})
	if err != nil {
		t.Fatalf("CreateRole failed: %v", err)
	}

	// Create instance profile
	_, err = iamClient.CreateInstanceProfile(ctx, &iam.CreateInstanceProfileInput{
		InstanceProfileName: aws.String("test-ec2-profile"),
	})
	if err != nil {
		t.Fatalf("CreateInstanceProfile failed: %v", err)
	}

	// Associate role with profile
	_, err = iamClient.AddRoleToInstanceProfile(ctx, &iam.AddRoleToInstanceProfileInput{
		InstanceProfileName: aws.String("test-ec2-profile"),
		RoleName:            aws.String("test-ec2-role"),
	})
	if err != nil {
		t.Fatalf("AddRoleToInstanceProfile failed: %v", err)
	}

	// Verify via GetInstanceProfile
	resp, err := iamClient.GetInstanceProfile(ctx, &iam.GetInstanceProfileInput{
		InstanceProfileName: aws.String("test-ec2-profile"),
	})
	if err != nil {
		t.Fatalf("GetInstanceProfile failed: %v", err)
	}
	if len(resp.InstanceProfile.Roles) != 1 {
		t.Errorf("Expected 1 role in profile, got %d", len(resp.InstanceProfile.Roles))
	}
	if *resp.InstanceProfile.Roles[0].RoleName != "test-ec2-role" {
		t.Errorf("Role name = %s, want test-ec2-role", *resp.InstanceProfile.Roles[0].RoleName)
	}
}

func TestSubstrateSSMRunCommand(t *testing.T) {
	_, ts := setupSubstrateManager(t)

	ctx := context.Background()
	cfg := aws.Config{
		Region: "us-east-1",
		EndpointResolverWithOptions: aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{URL: ts.URL, HostnameImmutable: true, SigningRegion: "us-east-1"}, nil
			},
		),
		Credentials: credentials.NewStaticCredentialsProvider("test", "test", ""),
	}
	ssmClient := ssm.NewFromConfig(cfg)

	// Send a command
	sendResp, err := ssmClient.SendCommand(ctx, &ssm.SendCommandInput{
		DocumentName: aws.String("AWS-RunShellScript"),
		InstanceIds:  []string{"i-test12345678"},
		Parameters: map[string][]string{
			"commands": {"echo hello"},
		},
	})
	if err != nil {
		t.Fatalf("SendCommand failed: %v", err)
	}
	if sendResp.Command == nil || sendResp.Command.CommandId == nil {
		t.Fatal("SendCommand returned nil CommandId")
	}
	commandID := *sendResp.Command.CommandId

	// Poll for completion
	getResp, err := ssmClient.GetCommandInvocation(ctx, &ssm.GetCommandInvocationInput{
		CommandId:  aws.String(commandID),
		InstanceId: aws.String("i-test12345678"),
	})
	if err != nil {
		t.Fatalf("GetCommandInvocation failed: %v", err)
	}
	if string(getResp.Status) != "Success" {
		t.Errorf("Command status = %s, want Success", getResp.Status)
	}
}

func TestSubstrateErrorHandling(t *testing.T) {
	manager, _ := setupSubstrateManager(t)

	// LaunchInstance with bad template should fail early (before any AWS call)
	_, err := manager.LaunchInstance(ctypes.LaunchRequest{
		Template: "nonexistent-template",
		Name:     "test-error-instance",
		Size:     "M",
		Region:   "us-east-1",
	})
	if err == nil {
		t.Error("LaunchInstance should fail with nonexistent template")
	}

	// DeleteStorage on nonexistent volume should fail
	if err := manager.DeleteStorage("nonexistent-volume"); err == nil {
		t.Error("DeleteStorage should fail for nonexistent volume")
	}
}
