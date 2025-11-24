//go:build linux
// +build linux

package sleepwake

import (
	"log"
)

// linuxMonitor implements sleep/wake monitoring for Linux (stub for now)
type linuxMonitor struct {
	eventCh chan<- Event
}

// newLinuxMonitor creates a new Linux sleep/wake monitor
func newLinuxMonitor() (platformMonitor, error) {
	log.Println("Warning: Linux sleep/wake monitoring not yet implemented")
	return &linuxMonitor{}, ErrPlatformNotSupported
}

// newPlatformMonitor creates a platform-specific monitor (Linux implementation)
func newPlatformMonitor() (platformMonitor, error) {
	return newLinuxMonitor()
}

// Start begins monitoring (stub)
func (m *linuxMonitor) Start(eventCh chan<- Event) error {
	return ErrPlatformNotSupported
}

// Stop stops monitoring (stub)
func (m *linuxMonitor) Stop() error {
	return nil
}

// Platform returns the platform name
func (m *linuxMonitor) Platform() string {
	return "linux"
}
