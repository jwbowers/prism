//go:build !windows

// Package security provides a stub Windows machine GUID for non-Windows platforms.
package security

func getWindowsMachineGUID() string {
	return "windows-machine-guid-unavailable"
}
