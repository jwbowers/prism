package templates

import (
	"strings"
	"testing"
)

// TestGenerateFileProvisioningScript_Empty tests empty file list
func TestGenerateFileProvisioningScript_Empty(t *testing.T) {
	script := GenerateFileProvisioningScript([]FileConfig{}, "us-west-2")
	if script != "" {
		t.Errorf("Expected empty script for empty file list, got: %s", script)
	}
}

// TestGenerateFileProvisioningScript_SingleFile tests basic file provisioning
func TestGenerateFileProvisioningScript_SingleFile(t *testing.T) {
	files := []FileConfig{
		{
			S3Bucket:        "test-bucket",
			S3Key:           "path/to/file.txt",
			DestinationPath: "/home/ubuntu/file.txt",
			Description:     "Test file",
			Owner:           "ubuntu",
			Group:           "ubuntu",
			Permissions:     "0644",
			Checksum:        true,
			Required:        true,
		},
	}

	script := GenerateFileProvisioningScript(files, "us-east-1")

	// Verify key components are present
	requiredComponents := []string{
		"File Provisioning from S3",
		"AWS_DEFAULT_REGION='us-east-1'",
		"test-bucket",
		"path/to/file.txt",
		"/home/ubuntu/file.txt",
		"Test file",
		"chown ubuntu:ubuntu",
		"chmod 0644",
		"s3://test-bucket/path/to/file.txt",
	}

	for _, component := range requiredComponents {
		if !strings.Contains(script, component) {
			t.Errorf("Script missing required component: %s", component)
		}
	}
}

// TestGenerateFileProvisioningScript_MultipleFiles tests multiple file provisioning
func TestGenerateFileProvisioningScript_MultipleFiles(t *testing.T) {
	files := []FileConfig{
		{
			S3Bucket:        "bucket1",
			S3Key:           "file1.txt",
			DestinationPath: "/home/ubuntu/file1.txt",
			Description:     "First file",
			Checksum:        true,
			Required:        true,
		},
		{
			S3Bucket:        "bucket2",
			S3Key:           "file2.txt",
			DestinationPath: "/home/ubuntu/file2.txt",
			Description:     "Second file",
			Checksum:        false,
			Required:        false,
		},
	}

	script := GenerateFileProvisioningScript(files, "us-west-2")

	// Verify both files are referenced
	if !strings.Contains(script, "bucket1") || !strings.Contains(script, "file1.txt") {
		t.Error("Script missing first file configuration")
	}
	if !strings.Contains(script, "bucket2") || !strings.Contains(script, "file2.txt") {
		t.Error("Script missing second file configuration")
	}

	// Verify file numbering
	if !strings.Contains(script, "File 1:") || !strings.Contains(script, "File 2:") {
		t.Error("Script missing file numbering")
	}
}

// TestGenerateFileProvisioningScript_AutoCleanup tests auto-cleanup feature
func TestGenerateFileProvisioningScript_AutoCleanup(t *testing.T) {
	files := []FileConfig{
		{
			S3Bucket:        "test-bucket",
			S3Key:           "temp-file.txt",
			DestinationPath: "/tmp/temp-file.txt",
			Description:     "Temporary file",
			AutoCleanup:     true,
			Required:        false,
		},
	}

	script := GenerateFileProvisioningScript(files, "us-east-1")

	// Verify auto-cleanup command is present
	if !strings.Contains(script, "aws s3 rm") {
		t.Error("Script missing auto-cleanup command")
	}
	if !strings.Contains(script, "Removed from S3") {
		t.Error("Script missing auto-cleanup message")
	}
}

// TestGenerateFileProvisioningScript_Conditional tests conditional file provisioning
func TestGenerateFileProvisioningScript_Conditional(t *testing.T) {
	files := []FileConfig{
		{
			S3Bucket:        "test-bucket",
			S3Key:           "x86-only-file.txt",
			DestinationPath: "/opt/x86-file.txt",
			Description:     "x86-only file",
			OnlyIf:          "arch == 'x86_64'",
			Required:        false,
		},
	}

	script := GenerateFileProvisioningScript(files, "us-east-1")

	// Verify conditional logic is present
	if !strings.Contains(script, "CONDITION_MET") {
		t.Error("Script missing conditional check")
	}
	if !strings.Contains(script, "uname -m") {
		t.Error("Script missing architecture detection")
	}
	if !strings.Contains(script, "condition not met") {
		t.Error("Script missing conditional skip message")
	}
}

// TestGenerateFileProvisioningScript_RequiredVsOptional tests required vs optional files
func TestGenerateFileProvisioningScript_RequiredVsOptional(t *testing.T) {
	filesRequired := []FileConfig{
		{
			S3Bucket:        "test-bucket",
			S3Key:           "required.txt",
			DestinationPath: "/opt/required.txt",
			Description:     "Required file",
			Required:        true,
		},
	}

	filesOptional := []FileConfig{
		{
			S3Bucket:        "test-bucket",
			S3Key:           "optional.txt",
			DestinationPath: "/opt/optional.txt",
			Description:     "Optional file",
			Required:        false,
		},
	}

	scriptRequired := GenerateFileProvisioningScript(filesRequired, "us-east-1")
	scriptOptional := GenerateFileProvisioningScript(filesOptional, "us-east-1")

	// Required file should have exit 1 on failure
	if !strings.Contains(scriptRequired, "exit 1") {
		t.Error("Required file script missing exit on failure")
	}
	if !strings.Contains(scriptRequired, "ERROR: Required file download failed") {
		t.Error("Required file script missing error message")
	}

	// Optional file should have warning and continue
	if strings.Contains(scriptOptional, "exit 1") {
		t.Error("Optional file script should not exit on failure")
	}
	if !strings.Contains(scriptOptional, "WARNING: Optional file download failed") {
		t.Error("Optional file script missing warning message")
	}
}

// TestGenerateFileProvisioningScript_Checksum tests checksum verification
func TestGenerateFileProvisioningScript_Checksum(t *testing.T) {
	filesWithChecksum := []FileConfig{
		{
			S3Bucket:        "test-bucket",
			S3Key:           "verified.txt",
			DestinationPath: "/opt/verified.txt",
			Description:     "Verified file",
			Checksum:        true,
		},
	}

	filesWithoutChecksum := []FileConfig{
		{
			S3Bucket:        "test-bucket",
			S3Key:           "unverified.txt",
			DestinationPath: "/opt/unverified.txt",
			Description:     "Unverified file",
			Checksum:        false,
		},
	}

	scriptWithChecksum := GenerateFileProvisioningScript(filesWithChecksum, "us-east-1")
	scriptWithoutChecksum := GenerateFileProvisioningScript(filesWithoutChecksum, "us-east-1")

	// With checksum should NOT have --no-verify-ssl
	if strings.Contains(scriptWithChecksum, "--no-verify-ssl") {
		t.Error("Checksum-enabled script should not have --no-verify-ssl")
	}

	// Without checksum should have --no-verify-ssl
	if !strings.Contains(scriptWithoutChecksum, "--no-verify-ssl") {
		t.Error("Checksum-disabled script missing --no-verify-ssl")
	}
}

// TestValidateFileConfig_Valid tests validation of valid configurations
func TestValidateFileConfig_Valid(t *testing.T) {
	validConfigs := []FileConfig{
		{
			S3Bucket:        "test-bucket",
			S3Key:           "file.txt",
			DestinationPath: "/opt/file.txt",
		},
		{
			S3Bucket:        "test-bucket",
			S3Key:           "file.txt",
			DestinationPath: "/opt/file.txt",
			Permissions:     "0644",
		},
		{
			S3Bucket:        "test-bucket",
			S3Key:           "file.txt",
			DestinationPath: "/opt/file.txt",
			Permissions:     "0755",
		},
	}

	for i, config := range validConfigs {
		err := ValidateFileConfig(config)
		if err != nil {
			t.Errorf("Config %d should be valid, got error: %v", i, err)
		}
	}
}

// TestValidateFileConfig_Invalid tests validation of invalid configurations
func TestValidateFileConfig_Invalid(t *testing.T) {
	invalidConfigs := []struct {
		name   string
		config FileConfig
	}{
		{
			name: "Missing S3 bucket",
			config: FileConfig{
				S3Key:           "file.txt",
				DestinationPath: "/opt/file.txt",
			},
		},
		{
			name: "Missing S3 key",
			config: FileConfig{
				S3Bucket:        "test-bucket",
				DestinationPath: "/opt/file.txt",
			},
		},
		{
			name: "Missing destination path",
			config: FileConfig{
				S3Bucket: "test-bucket",
				S3Key:    "file.txt",
			},
		},
		{
			name: "Invalid permissions format",
			config: FileConfig{
				S3Bucket:        "test-bucket",
				S3Key:           "file.txt",
				DestinationPath: "/opt/file.txt",
				Permissions:     "644", // Missing leading 0
			},
		},
		{
			name: "Invalid permissions length",
			config: FileConfig{
				S3Bucket:        "test-bucket",
				S3Key:           "file.txt",
				DestinationPath: "/opt/file.txt",
				Permissions:     "0777777", // Too long
			},
		},
	}

	for _, tc := range invalidConfigs {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateFileConfig(tc.config)
			if err == nil {
				t.Errorf("Expected validation error for %s", tc.name)
			}
		})
	}
}

// TestValidateTemplateFiles tests template-level file validation
func TestValidateTemplateFiles(t *testing.T) {
	validTemplate := &Template{
		Files: []FileConfig{
			{
				S3Bucket:        "test-bucket",
				S3Key:           "file1.txt",
				DestinationPath: "/opt/file1.txt",
			},
			{
				S3Bucket:        "test-bucket",
				S3Key:           "file2.txt",
				DestinationPath: "/opt/file2.txt",
			},
		},
	}

	invalidTemplate := &Template{
		Files: []FileConfig{
			{
				S3Bucket:        "test-bucket",
				S3Key:           "file1.txt",
				DestinationPath: "/opt/file1.txt",
			},
			{
				// Missing S3Key
				S3Bucket:        "test-bucket",
				DestinationPath: "/opt/file2.txt",
			},
		},
	}

	err := ValidateTemplateFiles(validTemplate)
	if err != nil {
		t.Errorf("Valid template should pass validation, got error: %v", err)
	}

	err = ValidateTemplateFiles(invalidTemplate)
	if err == nil {
		t.Error("Invalid template should fail validation")
	}
}

// TestEstimateFileProvisioningTime tests time estimation
func TestEstimateFileProvisioningTime(t *testing.T) {
	tests := []struct {
		name         string
		fileCount    int
		expectedTime int
		checkGreater bool
	}{
		{
			name:         "No files",
			fileCount:    0,
			expectedTime: 0,
		},
		{
			name:         "Single file",
			fileCount:    1,
			checkGreater: true,
		},
		{
			name:         "Multiple files",
			fileCount:    3,
			checkGreater: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files := make([]FileConfig, tt.fileCount)
			for i := range files {
				files[i] = FileConfig{
					S3Bucket:        "test-bucket",
					S3Key:           "file.txt",
					DestinationPath: "/opt/file.txt",
				}
			}

			estimate := EstimateFileProvisioningTime(files)

			if tt.checkGreater {
				if estimate <= 0 {
					t.Errorf("Expected positive time estimate, got: %d", estimate)
				}
				// Should include AWS CLI install time (1 min) plus file time
				minExpected := 1 + tt.fileCount
				if estimate < minExpected {
					t.Errorf("Expected at least %d minutes, got: %d", minExpected, estimate)
				}
			} else {
				if estimate != tt.expectedTime {
					t.Errorf("Expected %d minutes, got: %d", tt.expectedTime, estimate)
				}
			}
		})
	}
}

// TestGenerateConditionalCheck_x86 tests x86_64 architecture check
func TestGenerateConditionalCheck_x86(t *testing.T) {
	check := generateConditionalCheck("arch == 'x86_64'")

	if !strings.Contains(check, "CONDITION_MET=false") {
		t.Error("Conditional check missing initialization")
	}
	if !strings.Contains(check, "uname -m") {
		t.Error("Conditional check missing architecture detection")
	}
	if !strings.Contains(check, "x86_64") {
		t.Error("Conditional check missing x86_64 comparison")
	}
}

// TestGenerateConditionalCheck_ARM tests ARM architecture check
func TestGenerateConditionalCheck_ARM(t *testing.T) {
	check := generateConditionalCheck("arch == 'arm64'")

	if !strings.Contains(check, "aarch64") && !strings.Contains(check, "arm64") {
		t.Error("Conditional check missing ARM architecture comparison")
	}
}

// TestGenerateFileProvisioningScript_AWSCLIInstall tests AWS CLI installation
func TestGenerateFileProvisioningScript_AWSCLIInstall(t *testing.T) {
	files := []FileConfig{
		{
			S3Bucket:        "test-bucket",
			S3Key:           "file.txt",
			DestinationPath: "/opt/file.txt",
			Description:     "Test file",
		},
	}

	script := GenerateFileProvisioningScript(files, "us-east-1")

	// Verify AWS CLI installation logic
	requiredComponents := []string{
		"command -v aws",
		"Installing AWS CLI",
		"awscli.amazonaws.com/awscli-exe-linux-x86_64.zip",
		"/tmp/aws/install",
	}

	for _, component := range requiredComponents {
		if !strings.Contains(script, component) {
			t.Errorf("Script missing AWS CLI install component: %s", component)
		}
	}
}

// TestGenerateFileProvisioningScript_ProgressMessages tests progress feedback
func TestGenerateFileProvisioningScript_ProgressMessages(t *testing.T) {
	files := []FileConfig{
		{
			S3Bucket:        "test-bucket",
			S3Key:           "file.txt",
			DestinationPath: "/opt/file.txt",
			Description:     "Test file",
			Required:        true,
		},
	}

	script := GenerateFileProvisioningScript(files, "us-east-1")

	// Verify progress messages
	progressMessages := []string{
		"Starting file provisioning from S3",
		"Downloading Test file",
		"✓ Downloaded: Test file",
		"✗ Failed to download: Test file",
		"File provisioning complete",
	}

	for _, message := range progressMessages {
		if !strings.Contains(script, message) {
			t.Errorf("Script missing progress message: %s", message)
		}
	}
}
