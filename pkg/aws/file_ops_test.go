package aws

import (
	"testing"
	"time"
)

func TestParseLsOutput(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		basePath string
		want     []RemoteFileEntry
	}{
		{
			name:     "empty output",
			output:   "",
			basePath: "/home/ec2-user",
			want:     nil,
		},
		{
			name:     "total line only",
			output:   "total 0",
			basePath: "/home",
			want:     nil,
		},
		{
			name:     "skip dot entries",
			output:   "-rw-r--r-- 1 root root 0 2024-01-01T00:00:00 .\n-rw-r--r-- 1 root root 0 2024-01-01T00:00:00 ..",
			basePath: "/home",
			want:     nil,
		},
		{
			name:     "single file",
			output:   "-rw-r--r-- 1 ubuntu ubuntu 1024 2024-03-15T10:30:00 data.csv",
			basePath: "/home/ubuntu",
			want: []RemoteFileEntry{
				{
					Path:        "/home/ubuntu/data.csv",
					SizeBytes:   1024,
					IsDir:       false,
					ModifiedAt:  time.Date(2024, 3, 15, 10, 30, 0, 0, time.UTC),
					Permissions: "-rw-r--r--",
				},
			},
		},
		{
			name:     "directory entry",
			output:   "drwxr-xr-x 2 ubuntu ubuntu 0 2024-03-15T12:00:00 results",
			basePath: "/home/ubuntu",
			want: []RemoteFileEntry{
				{
					Path:        "/home/ubuntu/results",
					SizeBytes:   0,
					IsDir:       true,
					ModifiedAt:  time.Date(2024, 3, 15, 12, 0, 0, 0, time.UTC),
					Permissions: "drwxr-xr-x",
				},
			},
		},
		{
			name: "mixed entries with header",
			output: `total 8
drwxr-xr-x 3 ubuntu ubuntu    0 2024-01-01T00:00:00 .
drwxr-xr-x 5 root   root      0 2024-01-01T00:00:00 ..
-rw-r--r-- 1 ubuntu ubuntu 2048 2024-06-01T08:00:00 notes.txt
drwxr-xr-x 2 ubuntu ubuntu    0 2024-06-01T09:00:00 logs`,
			basePath: "/home/ubuntu",
			want: []RemoteFileEntry{
				{
					Path:        "/home/ubuntu/notes.txt",
					SizeBytes:   2048,
					IsDir:       false,
					ModifiedAt:  time.Date(2024, 6, 1, 8, 0, 0, 0, time.UTC),
					Permissions: "-rw-r--r--",
				},
				{
					Path:        "/home/ubuntu/logs",
					SizeBytes:   0,
					IsDir:       true,
					ModifiedAt:  time.Date(2024, 6, 1, 9, 0, 0, 0, time.UTC),
					Permissions: "drwxr-xr-x",
				},
			},
		},
		{
			name:     "ERROR line skipped",
			output:   "ERROR: some failure\n-rw-r--r-- 1 root root 100 2024-01-01T00:00:00 ok.txt",
			basePath: "/tmp",
			want: []RemoteFileEntry{
				{
					Path:        "/tmp/ok.txt",
					SizeBytes:   100,
					IsDir:       false,
					ModifiedAt:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					Permissions: "-rw-r--r--",
				},
			},
		},
		{
			name:     "short line skipped",
			output:   "-rw-r--r-- 1 root root 0 2024-01-01",
			basePath: "/tmp",
			want:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseLsOutput(tt.output, tt.basePath)
			if len(got) != len(tt.want) {
				t.Fatalf("len(got)=%d, want %d; got=%+v", len(got), len(tt.want), got)
			}
			for i, entry := range got {
				w := tt.want[i]
				if entry.Path != w.Path {
					t.Errorf("[%d] Path = %q, want %q", i, entry.Path, w.Path)
				}
				if entry.SizeBytes != w.SizeBytes {
					t.Errorf("[%d] SizeBytes = %d, want %d", i, entry.SizeBytes, w.SizeBytes)
				}
				if entry.IsDir != w.IsDir {
					t.Errorf("[%d] IsDir = %v, want %v", i, entry.IsDir, w.IsDir)
				}
				if entry.Permissions != w.Permissions {
					t.Errorf("[%d] Permissions = %q, want %q", i, entry.Permissions, w.Permissions)
				}
				if !entry.ModifiedAt.IsZero() && !entry.ModifiedAt.Equal(w.ModifiedAt) {
					t.Errorf("[%d] ModifiedAt = %v, want %v", i, entry.ModifiedAt, w.ModifiedAt)
				}
			}
		})
	}
}

func TestRemoteFileEntryJSON(t *testing.T) {
	entries := parseLsOutput("-rw-r--r-- 1 ubuntu ubuntu 512 2024-05-10T14:00:00 report.pdf", "/home/ubuntu")
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	e := entries[0]
	if e.Path != "/home/ubuntu/report.pdf" {
		t.Errorf("Path = %q, want /home/ubuntu/report.pdf", e.Path)
	}
	if e.SizeBytes != 512 {
		t.Errorf("SizeBytes = %d, want 512", e.SizeBytes)
	}
	if e.IsDir {
		t.Error("IsDir should be false for file")
	}
}
