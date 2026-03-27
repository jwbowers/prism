// Package course — tests for the academic integrity audit log.
package course

import (
	"fmt"
	"os"
	"sync"
	"testing"
	"time"
)

func TestAuditLogAppendAndQuery(t *testing.T) {
	dir := t.TempDir()
	log := NewAuditLog("cs101", dir)

	// Append a few entries
	entries := []AuditEntry{
		{CourseID: "cs101", Actor: "prof1", Action: AuditActionEnroll, Target: "alice"},
		{CourseID: "cs101", Actor: "alice", Action: AuditActionInstanceLaunch, Target: "alice-ws"},
		{CourseID: "cs101", Actor: "prof1", Action: AuditActionBudgetDistribute, Detail: map[string]interface{}{"amount": 50.0}},
		{CourseID: "cs101", Actor: "ta1", Action: AuditActionTADebug, Target: "bob"},
	}
	for _, e := range entries {
		if err := log.Append(e); err != nil {
			t.Fatalf("Append failed: %v", err)
		}
	}

	// Query all
	results, err := log.Query("", time.Time{}, 0)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	if len(results) != 4 {
		t.Errorf("expected 4 entries, got %d", len(results))
	}

	// Query by studentID
	results, err = log.Query("alice", time.Time{}, 0)
	if err != nil {
		t.Fatalf("Query(alice) failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 entries for alice, got %d", len(results))
	}

	// Query with limit
	results, err = log.Query("", time.Time{}, 2)
	if err != nil {
		t.Fatalf("Query(limit=2) failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 entries with limit, got %d", len(results))
	}

	// Query since future — should return nothing
	results, err = log.Query("", time.Now().Add(time.Hour), 0)
	if err != nil {
		t.Fatalf("Query(future) failed: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 entries for future since, got %d", len(results))
	}
}

func TestAuditLogConcurrency(t *testing.T) {
	dir := t.TempDir()
	log := NewAuditLog("concurrent-course", dir)

	const goroutines = 20
	const entriesEach = 10

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(n int) {
			defer wg.Done()
			for j := 0; j < entriesEach; j++ {
				entry := AuditEntry{
					CourseID: "concurrent-course",
					Actor:    fmt.Sprintf("user%d", n),
					Action:   AuditActionBudgetSpend,
					Target:   fmt.Sprintf("user%d", n),
				}
				if err := log.Append(entry); err != nil {
					t.Errorf("concurrent Append failed: %v", err)
				}
			}
		}(i)
	}
	wg.Wait()

	results, err := log.Query("", time.Time{}, 0)
	if err != nil {
		t.Fatalf("Query after concurrent writes failed: %v", err)
	}
	if len(results) != goroutines*entriesEach {
		t.Errorf("expected %d entries, got %d", goroutines*entriesEach, len(results))
	}
}

func TestAuditLogNoFile(t *testing.T) {
	dir := t.TempDir()
	log := NewAuditLog("nonexistent-course", dir)

	// Query on a log that has no file yet should return nil, nil
	results, err := log.Query("", time.Time{}, 0)
	if err != nil {
		t.Fatalf("Query on missing file should not error: %v", err)
	}
	if results != nil {
		t.Errorf("expected nil results for missing file, got %v", results)
	}
}

func TestAuditLogTimestampAutoSet(t *testing.T) {
	dir := t.TempDir()
	log := NewAuditLog("ts-course", dir)

	before := time.Now()
	entry := AuditEntry{
		CourseID: "ts-course",
		Actor:    "system",
		Action:   AuditActionCourseClose,
		// Timestamp is intentionally zero — should be auto-set
	}
	if err := log.Append(entry); err != nil {
		t.Fatalf("Append failed: %v", err)
	}
	after := time.Now()

	results, err := log.Query("", time.Time{}, 0)
	if err != nil || len(results) != 1 {
		t.Fatalf("unexpected query result: err=%v len=%d", err, len(results))
	}
	ts := results[0].Timestamp
	if ts.Before(before) || ts.After(after) {
		t.Errorf("timestamp %v not in expected range [%v, %v]", ts, before, after)
	}
}

// TestAuditLogFileNotWritable verifies that Append errors are surfaced properly.
func TestAuditLogFileNotWritable(t *testing.T) {
	dir := t.TempDir()
	// Make the directory read-only so the file cannot be created
	if err := os.Chmod(dir, 0555); err != nil {
		t.Skip("cannot set directory permissions:", err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0755) })

	log := NewAuditLog("unwritable", dir)
	err := log.Append(AuditEntry{CourseID: "unwritable", Actor: "x", Action: AuditActionEnroll})
	if err == nil {
		t.Error("expected error writing to read-only directory, got nil")
	}
}
