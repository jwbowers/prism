package integration

import (
	"context"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/scttfrdmn/prism/pkg/api/client"
)

// Global test suite manager
var (
	suiteManager     *TestSuiteManager
	suiteManagerOnce sync.Once
)

// TestSuiteManager ensures proper resource management across all tests
type TestSuiteManager struct {
	mu                sync.Mutex
	activeInstances   map[string]bool
	maxInstances      int
	instanceSemaphore chan struct{}
	client            client.PrismAPI
}

// GetSuiteManager returns the singleton test suite manager
func GetSuiteManager() *TestSuiteManager {
	suiteManagerOnce.Do(func() {
		suiteManager = &TestSuiteManager{
			activeInstances:   make(map[string]bool),
			maxInstances:      2, // CRITICAL: Only 2 instances at a time
			instanceSemaphore: make(chan struct{}, 2),
			client:            nil, // Will be set when first context is created
		}
	})
	return suiteManager
}

// SetClient sets the API client for cleanup operations
func (sm *TestSuiteManager) SetClient(c client.PrismAPI) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if sm.client == nil {
		sm.client = c
	}
}

// AcquireInstanceSlot blocks until a slot is available (max 2 concurrent instances)
func (sm *TestSuiteManager) AcquireInstanceSlot(t *testing.T, instanceName string) {
	t.Helper()

	// Block until we can acquire a slot
	select {
	case sm.instanceSemaphore <- struct{}{}:
		sm.mu.Lock()
		sm.activeInstances[instanceName] = true
		count := len(sm.activeInstances)
		sm.mu.Unlock()

		t.Logf("✓ Acquired instance slot for %s (active: %d/%d)", instanceName, count, sm.maxInstances)
	case <-time.After(5 * time.Minute):
		t.Fatalf("Timeout waiting for instance slot - other tests may have leaked instances")
	}
}

// ReleaseInstanceSlot frees a slot and ensures instance is terminated
func (sm *TestSuiteManager) ReleaseInstanceSlot(t *testing.T, instanceName string) {
	t.Helper()

	// Ensure instance is terminated before releasing slot
	if sm.client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := sm.client.DeleteInstance(ctx, instanceName); err != nil {
			t.Logf("Warning: Failed to delete instance %s: %v", instanceName, err)
		} else {
			t.Logf("✓ Instance %s terminated", instanceName)
		}
	}

	sm.mu.Lock()
	delete(sm.activeInstances, instanceName)
	count := len(sm.activeInstances)
	sm.mu.Unlock()

	// Release semaphore slot
	<-sm.instanceSemaphore

	t.Logf("✓ Released instance slot for %s (active: %d/%d)", instanceName, count, sm.maxInstances)
}

// CleanupAllInstances forcefully terminates all tracked instances
func (sm *TestSuiteManager) CleanupAllInstances(t *testing.T) {
	sm.mu.Lock()
	instances := make([]string, 0, len(sm.activeInstances))
	for name := range sm.activeInstances {
		instances = append(instances, name)
	}
	sm.mu.Unlock()

	if len(instances) == 0 {
		return
	}

	t.Logf("⚠️  Cleaning up %d leaked instances", len(instances))

	if sm.client == nil {
		t.Log("Warning: No client available for cleanup")
		return
	}

	for _, name := range instances {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		if err := sm.client.DeleteInstance(ctx, name); err != nil {
			t.Logf("Warning: Failed to cleanup instance %s: %v", name, err)
		} else {
			t.Logf("✓ Cleaned up leaked instance: %s", name)
		}
		cancel()
	}
}

// GetActiveInstanceCount returns the current number of active instances
func (sm *TestSuiteManager) GetActiveInstanceCount() int {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	return len(sm.activeInstances)
}

// FailIfInstancesLeaked fails the test if any instances are still active
func (sm *TestSuiteManager) FailIfInstancesLeaked(t *testing.T) {
	count := sm.GetActiveInstanceCount()
	if count > 0 {
		sm.CleanupAllInstances(t)
		t.Errorf("❌ Test leaked %d instances - cleaned up but test should be fixed", count)
	}
}

// AggressiveCleanup runs at end of test suite to ensure NO instances left
func AggressiveCleanup(m *testing.M) {
	defer func() {
		if suiteManager != nil && suiteManager.client != nil {
			// Final safety check - terminate ANY test instances
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			instancesResp, err := suiteManager.client.ListInstances(ctx)
			if err != nil {
				log.Printf("Warning: Failed to list instances for final cleanup: %v", err)
				return
			}

			leaked := 0
			for _, inst := range instancesResp.Instances {
				if isTestInstance(inst.Name) {
					log.Printf("❌ LEAKED INSTANCE FOUND: %s - terminating", inst.Name)
					if err := suiteManager.client.DeleteInstance(ctx, inst.Name); err != nil {
						log.Printf("Error terminating leaked instance %s: %v", inst.Name, err)
					} else {
						leaked++
					}
				}
			}

			if leaked > 0 {
				log.Printf("⚠️  Cleaned up %d leaked test instances", leaked)
			} else {
				log.Printf("✓ No leaked test instances found")
			}
		}
	}()
}

// isTestInstance checks if an instance name indicates it's a test instance
func isTestInstance(name string) bool {
	testPrefixes := []string{"test-", "Test-"}
	for _, prefix := range testPrefixes {
		if len(name) >= len(prefix) && name[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}
