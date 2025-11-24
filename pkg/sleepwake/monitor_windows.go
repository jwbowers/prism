//go:build windows
// +build windows

package sleepwake

import (
	"log"
)

// windowsMonitor implements sleep/wake monitoring for Windows (stub for now)
type windowsMonitor struct {
	eventCh chan<- Event
}

// newWindowsMonitor creates a new Windows sleep/wake monitor
func newWindowsMonitor() (platformMonitor, error) {
	log.Println("Warning: Windows sleep/wake monitoring not yet implemented")
	return &windowsMonitor{}, ErrPlatformNotSupported
}

// newPlatformMonitor creates a platform-specific monitor (Windows implementation)
func newPlatformMonitor() (platformMonitor, error) {
	return newWindowsMonitor()
}

// Start begins monitoring (stub)
func (m *windowsMonitor) Start(eventCh chan<- Event) error {
	return ErrPlatformNotSupported
}

// Stop stops monitoring (stub)
func (m *windowsMonitor) Stop() error {
	return nil
}

// Platform returns the platform name
func (m *windowsMonitor) Platform() string {
	return "windows"
}
