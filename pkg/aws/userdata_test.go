package aws

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io"
	"strings"
	"testing"
)

func TestGzipBytes(t *testing.T) {
	input := []byte("#!/bin/bash\necho hello world\n")
	compressed, err := gzipBytes(input)
	if err != nil {
		t.Fatalf("gzipBytes failed: %v", err)
	}

	// Verify gzip magic bytes (1f 8b)
	if len(compressed) < 2 || compressed[0] != 0x1f || compressed[1] != 0x8b {
		t.Fatal("compressed output does not have gzip magic bytes")
	}

	// Verify round-trip decompression
	reader, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		t.Fatalf("gzip.NewReader failed: %v", err)
	}
	decompressed, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("reading decompressed data failed: %v", err)
	}
	reader.Close()

	if !bytes.Equal(decompressed, input) {
		t.Fatalf("round-trip mismatch:\n  got:  %q\n  want: %q", decompressed, input)
	}
}

func TestGzipBytesCompressesLargeScript(t *testing.T) {
	// Simulate a large cloud-init script (~20 KB of repetitive shell commands)
	var sb strings.Builder
	sb.WriteString("#!/bin/bash\nset -e\n")
	for i := 0; i < 500; i++ {
		sb.WriteString("apt-get install -y some-package-name-that-is-reasonably-long\n")
	}
	input := []byte(sb.String())

	compressed, err := gzipBytes(input)
	if err != nil {
		t.Fatalf("gzipBytes failed: %v", err)
	}

	// Shell scripts with repetitive text should compress well (at least 3x)
	ratio := float64(len(input)) / float64(len(compressed))
	if ratio < 3.0 {
		t.Errorf("compression ratio %.1fx is worse than expected (want >= 3x); raw=%d compressed=%d",
			ratio, len(input), len(compressed))
	}

	// After base64 encoding, result should fit in EC2's 25600 byte limit
	encoded := base64.StdEncoding.EncodeToString(compressed)
	if len(input) > 19200 && len(encoded) < 25600 {
		t.Logf("Success: %d byte script -> %d bytes compressed -> %d bytes base64 (fits EC2 limit)",
			len(input), len(compressed), len(encoded))
	}
}
