//go:build windows

// Package security provides Windows-specific device identifier for key derivation.
package security

import "golang.org/x/sys/windows/registry"

// getWindowsMachineGUID reads the machine GUID from the Windows registry.
// HKLM\SOFTWARE\Microsoft\Cryptography\MachineGuid is set during Windows setup
// and is unique per installation.
func getWindowsMachineGUID() string {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE,
		`SOFTWARE\Microsoft\Cryptography`, registry.QUERY_VALUE)
	if err != nil {
		return "windows-machine-guid-unavailable"
	}
	defer k.Close()

	guid, _, err := k.GetStringValue("MachineGuid")
	if err != nil {
		return "windows-machine-guid-unavailable"
	}
	return guid
}
