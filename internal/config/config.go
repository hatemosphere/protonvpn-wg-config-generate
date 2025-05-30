package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all configuration options
type Config struct {
	Username          string
	Password          string
	Countries         []string
	OutputFile        string
	ClientPrivateKey  string
	DNSServers        []string
	AllowedIPs        []string
	EnableAccelerator bool
	APIURL            string
	PlusServersOnly   bool
	P2PServersOnly    bool
	DeviceName        string
	Duration          string
	ClearSession      bool
	NoSession         bool
	SessionDuration   string
}

// Parse parses command-line flags and returns a Config
func Parse() (*Config, error) {
	cfg := &Config{}

	var countriesFlag string
	var dnsServersFlag string
	var allowedIPsFlag string

	flag.StringVar(&cfg.Username, "username", "", "ProtonVPN username")
	flag.StringVar(&cfg.Password, "password", "", "ProtonVPN password (will prompt if not provided)")
	flag.StringVar(&countriesFlag, "countries", "", "Comma-separated list of country codes (e.g., US,NL,CH)")
	flag.StringVar(&cfg.OutputFile, "output", "protonvpn.conf", "Output WireGuard configuration file")
	flag.StringVar(&dnsServersFlag, "dns", "10.2.0.1", "Comma-separated list of DNS servers")
	flag.StringVar(&allowedIPsFlag, "allowed-ips", "0.0.0.0/0,::/0", "Comma-separated list of allowed IPs")
	flag.BoolVar(&cfg.EnableAccelerator, "accelerator", true, "Enable VPN accelerator")
	flag.StringVar(&cfg.APIURL, "api-url", "https://vpn-api.proton.me", "ProtonVPN API URL")
	flag.BoolVar(&cfg.PlusServersOnly, "plus-only", true, "Use only Plus servers (Tier 2)")
	flag.BoolVar(&cfg.P2PServersOnly, "p2p-only", true, "Use only P2P-enabled servers")
	flag.StringVar(&cfg.DeviceName, "device-name", "", "Device name for WireGuard config (auto-generated if empty)")
	flag.StringVar(&cfg.Duration, "duration", "365d", "Certificate duration (e.g., 30m, 24h, 7d, 1h30m). Max: 365d")
	flag.BoolVar(&cfg.ClearSession, "clear-session", false, "Clear saved session and force re-authentication")
	flag.BoolVar(&cfg.NoSession, "no-session", false, "Don't save or use session persistence")
	flag.StringVar(&cfg.SessionDuration, "session-duration", "0", "Session cache duration (e.g., 12h, 24h, 7d). 0 = no expiration")
	flag.Parse()

	if countriesFlag == "" {
		return nil, fmt.Errorf("countries flag is required")
	}

	cfg.Countries = strings.Split(strings.ToUpper(countriesFlag), ",")
	cfg.DNSServers = strings.Split(dnsServersFlag, ",")
	cfg.AllowedIPs = strings.Split(allowedIPsFlag, ",")

	// Clean up username
	cfg.Username = cleanUsername(cfg.Username)

	return cfg, nil
}

// cleanUsername removes email domain suffixes from username
func cleanUsername(username string) string {
	username = strings.TrimSpace(username)
	username = strings.TrimSuffix(username, "@protonmail.com")
	username = strings.TrimSuffix(username, "@proton.me")
	return username
}

// ValidateCredentials checks if we have the required credentials
func (c *Config) ValidateCredentials() error {
	if c.Username == "" {
		return fmt.Errorf("username is required")
	}
	return nil
}

// PrintUsage prints usage information
func PrintUsage() {
	fmt.Fprintf(os.Stderr, "Usage: %s -username <username> -countries <country-codes> [options]\n\n", os.Args[0])
	flag.PrintDefaults()
}

// ParseDuration parses a duration string and converts it to minutes for the API
func ParseDuration(durationStr string) (string, error) {
	// Handle simple formats like "7d", "30d", "365d"
	if strings.HasSuffix(durationStr, "d") {
		daysStr := strings.TrimSuffix(durationStr, "d")
		days, err := strconv.Atoi(daysStr)
		if err != nil {
			return "", fmt.Errorf("invalid duration format: %s", durationStr)
		}
		if days < 1 {
			return "", fmt.Errorf("duration must be at least 1 day")
		}
		if days > 365 {
			return "", fmt.Errorf("duration cannot exceed 365 days")
		}
		minutes := days * 24 * 60
		return fmt.Sprintf("%d min", minutes), nil
	}

	// Try parsing as a Go duration
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return "", fmt.Errorf("invalid duration format: %s", durationStr)
	}

	// Convert to minutes
	minutes := int(duration.Minutes())
	if minutes < 1 {
		return "", fmt.Errorf("duration must be at least 1 minute")
	}

	// Max 365 days = 525600 minutes
	maxMinutes := 365 * 24 * 60
	if minutes > maxMinutes {
		return "", fmt.Errorf("duration cannot exceed 365 days")
	}

	return fmt.Sprintf("%d min", minutes), nil
}

// ParseSessionDuration parses a session duration string
func ParseSessionDuration(durationStr string) (time.Duration, error) {
	// Handle "0" as use API default
	if durationStr == "0" {
		return 0, nil
	}

	var duration time.Duration
	var err error

	// Handle simple formats like "7d", "30d"
	if strings.HasSuffix(durationStr, "d") {
		daysStr := strings.TrimSuffix(durationStr, "d")
		days, parseErr := strconv.Atoi(daysStr)
		if parseErr != nil {
			return 0, fmt.Errorf("invalid session duration format: %s", durationStr)
		}
		duration = time.Duration(days) * 24 * time.Hour
	} else {
		// Try parsing as a Go duration
		duration, err = time.ParseDuration(durationStr)
		if err != nil {
			return 0, fmt.Errorf("invalid session duration format: %s", durationStr)
		}
	}

	// Validate max duration (30 days)
	maxDuration := 30 * 24 * time.Hour
	if duration > maxDuration {
		return 0, fmt.Errorf("session duration cannot exceed 30 days (got: %s)", duration)
	}

	return duration, nil
}
