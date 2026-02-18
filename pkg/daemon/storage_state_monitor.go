package daemon

import (
	"log"
	"sync"
	"time"

	"github.com/scttfrdmn/prism/pkg/aws"
	"github.com/scttfrdmn/prism/pkg/state"
	"github.com/scttfrdmn/prism/pkg/types"
)

// StorageStateMonitor monitors storage volume state changes in the background
// It polls AWS for volumes in transitional states and updates local state
type StorageStateMonitor struct {
	awsManager   *aws.Manager
	stateManager *state.Manager
	ticker       *time.Ticker
	stopCh       chan struct{}
	wg           sync.WaitGroup
	mu           sync.Mutex
	running      bool
}

// NewStorageStateMonitor creates a new storage state monitor
func NewStorageStateMonitor(awsManager *aws.Manager, stateManager *state.Manager) *StorageStateMonitor {
	return &StorageStateMonitor{
		awsManager:   awsManager,
		stateManager: stateManager,
		stopCh:       make(chan struct{}),
	}
}

// Start begins background storage state monitoring
func (ssm *StorageStateMonitor) Start() error {
	ssm.mu.Lock()
	defer ssm.mu.Unlock()

	if ssm.running {
		return nil // Already running
	}

	ssm.ticker = time.NewTicker(10 * time.Second)
	ssm.running = true

	ssm.wg.Add(1)
	go ssm.monitorLoop()

	log.Printf("✅ Storage state monitor started (10s polling interval)")
	return nil
}

// Stop gracefully stops the storage state monitor
func (ssm *StorageStateMonitor) Stop() {
	ssm.mu.Lock()
	defer ssm.mu.Unlock()

	if !ssm.running {
		return // Not running
	}

	close(ssm.stopCh)
	ssm.ticker.Stop()
	ssm.running = false

	// Wait for monitor loop to finish
	ssm.wg.Wait()

	log.Printf("✅ Storage state monitor stopped")
}

// monitorLoop runs the background monitoring loop
func (ssm *StorageStateMonitor) monitorLoop() {
	defer ssm.wg.Done()

	for {
		select {
		case <-ssm.ticker.C:
			ssm.checkTransitionalVolumes()
		case <-ssm.stopCh:
			return
		}
	}
}

// checkTransitionalVolumes checks AWS for volumes in transitional states
func (ssm *StorageStateMonitor) checkTransitionalVolumes() {
	// Load current state
	state, err := ssm.stateManager.LoadState()
	if err != nil {
		log.Printf("Warning: Storage state monitor failed to load state: %v", err)
		return
	}

	// Find volumes in transitional states
	var transitionalVolumes []types.StorageVolume
	for _, vol := range state.StorageVolumes {
		if isTransitionalVolumeState(vol.State) {
			transitionalVolumes = append(transitionalVolumes, vol)
		}
	}

	if len(transitionalVolumes) == 0 {
		return // No volumes to monitor
	}

	log.Printf("🔍 Storage state monitor checking %d volume(s) in transitional states", len(transitionalVolumes))

	// Check each transitional volume
	for _, vol := range transitionalVolumes {
		ssm.checkVolume(vol)
	}
}

// checkVolume checks a single volume's state from AWS
func (ssm *StorageStateMonitor) checkVolume(vol types.StorageVolume) {
	var currentState string
	var err error

	// Query AWS for current state
	if vol.IsShared() {
		currentState, err = ssm.awsManager.GetEFSVolumeState(vol.FileSystemID)
	} else if vol.IsWorkspace() {
		currentState, err = ssm.awsManager.GetEBSVolumeState(vol.VolumeID)
	} else {
		return // Unknown volume type
	}

	if err != nil {
		// Volume might be deleted and gone from AWS
		if vol.State == "deleting" {
			ssm.handleDeletedVolume(vol)
		}
		return
	}

	// Check if state changed
	if currentState != vol.State {
		log.Printf("✅ Storage state changed: %s (%s → %s)", vol.Name, vol.State, currentState)

		// Update state
		vol.State = currentState
		if err := ssm.stateManager.SaveStorageVolume(vol); err != nil {
			log.Printf("Warning: Failed to update volume state: %v", err)
		}

		// Handle deleted volumes
		if currentState == "deleted" {
			ssm.handleDeletedVolume(vol)
		}
	}
}

// handleDeletedVolume removes a deleted volume after AWS confirms it's gone
func (ssm *StorageStateMonitor) handleDeletedVolume(vol types.StorageVolume) {
	var err error

	// Wait for volume to disappear from AWS (eventual consistency)
	// Poll for up to 5 minutes with 10-second intervals
	maxAttempts := 30 // 5 minutes / 10 seconds
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if vol.IsShared() {
			_, err = ssm.awsManager.GetEFSVolumeState(vol.FileSystemID)
		} else if vol.IsWorkspace() {
			_, err = ssm.awsManager.GetEBSVolumeState(vol.VolumeID)
		}

		if err != nil {
			// Volume not found - it's gone from AWS
			log.Printf("✅ Deleted volume %s confirmed gone from AWS, removing from state", vol.Name)

			// Remove from local state
			if err := ssm.stateManager.RemoveStorageVolume(vol.Name); err != nil {
				log.Printf("Warning: Failed to remove deleted volume: %v", err)
			}
			return
		}

		// Still visible in AWS, wait before retrying
		if attempt < maxAttempts {
			time.Sleep(10 * time.Second)
		}
	}

	log.Printf("Warning: Deleted volume %s still visible in AWS after 5 minutes", vol.Name)
}

// isTransitionalVolumeState returns true if the volume state is transitional
func isTransitionalVolumeState(state string) bool {
	switch state {
	case "creating", "deleting":
		return true
	default:
		return false
	}
}
