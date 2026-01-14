package update

import "time"

// UpdateInfo contains information about an available update
type UpdateInfo struct {
	CurrentVersion    string        `json:"current_version"`
	LatestVersion     string        `json:"latest_version"`
	IsUpdateAvailable bool          `json:"is_update_available"`
	ReleaseURL        string        `json:"release_url"`
	ReleaseNotes      string        `json:"release_notes"`
	PublishedAt       time.Time     `json:"published_at"`
	InstallMethod     InstallMethod `json:"install_method"`
	UpdateCommand     string        `json:"update_command"`
	LastChecked       time.Time     `json:"last_checked"`
}

// InstallMethod represents how Prism was installed
type InstallMethod string

const (
	InstallMethodHomebrew InstallMethod = "homebrew"
	InstallMethodScoop    InstallMethod = "scoop"
	InstallMethodBinary   InstallMethod = "binary"
	InstallMethodSource   InstallMethod = "source"
	InstallMethodUnknown  InstallMethod = "unknown"
)

// String returns the string representation of the install method
func (im InstallMethod) String() string {
	return string(im)
}

// UpdateChannel represents release channel (stable/beta/dev)
type UpdateChannel string

const (
	ChannelStable UpdateChannel = "stable"
	ChannelBeta   UpdateChannel = "beta"
	ChannelDev    UpdateChannel = "dev"
)

// CachedUpdateCheck stores cached update check results
type CachedUpdateCheck struct {
	UpdateInfo *UpdateInfo   `json:"update_info"`
	CheckedAt  time.Time     `json:"checked_at"`
	CacheTTL   time.Duration `json:"cache_ttl"`
}

// IsExpired checks if the cached update check has expired
func (c *CachedUpdateCheck) IsExpired() bool {
	return time.Since(c.CheckedAt) > c.CacheTTL
}
