package models

// PMSetting represents the private message settings.
type PMSetting struct {
	DisableIdleTime bool // Disable idle time in private messages.
	AllowAnon       bool // Allow anonymous messages.
	EmailOfflineMsg bool // Email offline messages.
}
