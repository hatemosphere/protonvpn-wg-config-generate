package constants

// Session defaults
const (
	SessionFileName      = ".protonvpn-session.json"
	SessionFileMode      = 0600    // Read/write for owner only
	SessionRefreshDays   = 7       // Refresh when less than 7 days remain
	SessionExpirySeconds = 2592000 // 30 days in seconds (from API)
)
