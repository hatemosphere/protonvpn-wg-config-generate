package config

import "fmt"

// Config holds all configuration options
type Config struct {
	// Authentication
	Username string
	Password string

	// Server selection
	Countries       []string
	PlusServersOnly bool
	P2PServersOnly  bool

	// Output configuration
	OutputFile       string
	ClientPrivateKey string
	DeviceName       string

	// Network configuration
	DNSServers        []string
	AllowedIPs        []string
	EnableAccelerator bool
	EnableIPv6        bool

	// Certificate configuration
	Duration string

	// Session management
	ClearSession    bool
	NoSession       bool
	ForceRefresh    bool
	SessionDuration string

	// Advanced configuration
	APIURL string
}

// ValidateCredentials checks if we have the required credentials
func (c *Config) ValidateCredentials() error {
	if c.Username == "" {
		return fmt.Errorf("username is required")
	}
	return nil
}
