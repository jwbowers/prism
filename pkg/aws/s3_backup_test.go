package aws

import (
	"math"
	"strings"
	"testing"
	"time"

	ctypes "github.com/scttfrdmn/prism/pkg/types"
)

const floatEps = 1e-9

func approxEqual(a, b float64) bool {
	return math.Abs(a-b) < floatEps
}

// ── s3BackupParameterName ─────────────────────────────────────────────────

func TestS3BackupParameterName(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"my-backup", "/prism/backups/my-backup"},
		{"research-2024-06-01", "/prism/backups/research-2024-06-01"},
		{"", "/prism/backups/"},
	}
	for _, tt := range tests {
		got := s3BackupParameterName(tt.name)
		if got != tt.want {
			t.Errorf("s3BackupParameterName(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}

// ── estimateS3MonthlyCost ─────────────────────────────────────────────────

func TestEstimateS3MonthlyCost(t *testing.T) {
	const rate = 0.023 // $/GB/month

	gb := int64(1024 * 1024 * 1024)
	tests := []struct {
		bytes int64
		want  float64
	}{
		{0, 0},
		{gb, rate},           // 1 GB
		{10 * gb, 10 * rate}, // 10 GB
		// 512 MB = exactly 0.5 GB in binary units
		{gb / 2, rate / 2},
	}
	for _, tt := range tests {
		got := estimateS3MonthlyCost(tt.bytes)
		if !approxEqual(got, tt.want) {
			t.Errorf("estimateS3MonthlyCost(%d) = %.9f, want %.9f", tt.bytes, got, tt.want)
		}
	}
}

// ── s3BackupMetadataToInfo ────────────────────────────────────────────────

func TestS3BackupMetadataToInfo(t *testing.T) {
	now := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	done := time.Date(2024, 6, 1, 12, 30, 0, 0, time.UTC)

	meta := &S3BackupMetadata{
		BackupName:   "my-backup",
		InstanceName: "research-1",
		InstanceID:   "i-abc123",
		SSMCommandID: "cmd-xyz",
		S3Bucket:     "prism-backups-123456789-us-east-1",
		S3Prefix:     "my-backup/2024-06-01",
		IncludePaths: []string{"/home", "/data"},
		ExcludePaths: []string{"/home/ubuntu/.cache"},
		CreatedAt:    now,
		CompletedAt:  &done,
		State:        "available",
		SizeBytes:    2 * 1024 * 1024 * 1024,
		FileCount:    1500,
	}

	info := s3BackupMetadataToInfo(meta)

	if info.BackupName != "my-backup" {
		t.Errorf("BackupName = %q, want my-backup", info.BackupName)
	}
	if info.BackupID != "cmd-xyz" {
		t.Errorf("BackupID = %q, want cmd-xyz", info.BackupID)
	}
	if info.SourceInstance != "research-1" {
		t.Errorf("SourceInstance = %q, want research-1", info.SourceInstance)
	}
	if info.StorageType != "s3" {
		t.Errorf("StorageType = %q, want s3", info.StorageType)
	}
	if info.BackupType != "full" {
		t.Errorf("BackupType = %q, want full", info.BackupType)
	}
	want := "s3://prism-backups-123456789-us-east-1/my-backup/2024-06-01/"
	if info.StorageLocation != want {
		t.Errorf("StorageLocation = %q, want %q", info.StorageLocation, want)
	}
	if info.State != "available" {
		t.Errorf("State = %q, want available", info.State)
	}
	if info.SizeBytes != meta.SizeBytes {
		t.Errorf("SizeBytes = %d, want %d", info.SizeBytes, meta.SizeBytes)
	}
	if info.FileCount != 1500 {
		t.Errorf("FileCount = %d, want 1500", info.FileCount)
	}
	if len(info.IncludedPaths) != 2 {
		t.Errorf("IncludedPaths len = %d, want 2", len(info.IncludedPaths))
	}
	if len(info.ExcludedPaths) != 1 {
		t.Errorf("ExcludedPaths len = %d, want 1", len(info.ExcludedPaths))
	}
	if info.Metadata != nil {
		t.Errorf("Metadata should be nil when no error, got %v", info.Metadata)
	}
	// cost: 2 GB * $0.023
	wantCost := 2 * 0.023
	if info.StorageCostMonthly != wantCost {
		t.Errorf("StorageCostMonthly = %.4f, want %.4f", info.StorageCostMonthly, wantCost)
	}
}

func TestS3BackupMetadataToInfo_WithError(t *testing.T) {
	meta := &S3BackupMetadata{
		BackupName:   "failed-backup",
		State:        "error",
		ErrorMessage: "instance unreachable",
	}
	info := s3BackupMetadataToInfo(meta)
	if info.Metadata == nil {
		t.Fatal("Metadata should be non-nil when ErrorMessage is set")
	}
	if info.Metadata["error"] != "instance unreachable" {
		t.Errorf("Metadata[error] = %q, want 'instance unreachable'", info.Metadata["error"])
	}
}

// Ensure s3BackupMetadataToInfo returns *BackupInfo (compile-time type check)
var _ *ctypes.BackupInfo = s3BackupMetadataToInfo(&S3BackupMetadata{})

// ── buildS3BackupScript ───────────────────────────────────────────────────

func TestBuildS3BackupScript_DefaultPaths(t *testing.T) {
	script := buildS3BackupScript("my-bucket", "backup-prefix", nil, nil)

	if !strings.HasPrefix(script, "#!/bin/bash") {
		t.Error("script should start with #!/bin/bash")
	}
	if !strings.Contains(script, `BUCKET="my-bucket"`) {
		t.Error("script should set BUCKET variable")
	}
	if !strings.Contains(script, `PREFIX="backup-prefix"`) {
		t.Error("script should set PREFIX variable")
	}
	// Default paths: /home and /data
	if !strings.Contains(script, `"/home"`) {
		t.Error("script should include /home by default")
	}
	if !strings.Contains(script, `"/data"`) {
		t.Error("script should include /data by default")
	}
	if !strings.Contains(script, "prism-backup] Backup complete") {
		t.Error("script should have completion message")
	}
}

func TestBuildS3BackupScript_CustomPaths(t *testing.T) {
	script := buildS3BackupScript("bucket", "pfx", []string{"/research", "/scratch"}, nil)

	if !strings.Contains(script, `"/research"`) {
		t.Error("script should include /research")
	}
	if !strings.Contains(script, `"/scratch"`) {
		t.Error("script should include /scratch")
	}
	// Default paths should NOT appear when custom paths given
	if strings.Contains(script, `"/data"`) {
		t.Error("script should not include /data when custom paths given")
	}
}

func TestBuildS3BackupScript_ExcludePaths(t *testing.T) {
	script := buildS3BackupScript("bucket", "pfx", nil, []string{"/home/ubuntu/.cache", "/tmp"})

	if !strings.Contains(script, "--exclude") {
		t.Error("script should contain --exclude flags")
	}
	if !strings.Contains(script, ".cache") {
		t.Error("script should reference .cache in excludes")
	}
}

func TestBuildS3BackupScript_S3DestFormat(t *testing.T) {
	script := buildS3BackupScript("my-bucket", "my-prefix", []string{"/home"}, nil)

	// S3 destination should use the bucket/prefix vars and append the relative path
	if !strings.Contains(script, `s3://"$BUCKET"/"$PREFIX"/home/`) {
		t.Errorf("unexpected S3 destination format in script:\n%s", script)
	}
}
