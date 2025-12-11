// Package constants defines constants used throughout the application.
package constants

// API endpoints
const (
	DefaultAPIURL   = "https://vpn-api.proton.me"
	AuthInfoPath    = "/core/v4/auth/info"
	AuthPath        = "/core/v4/auth"
	RefreshPath     = "/auth/refresh"
	CertificatePath = "/vpn/v1/certificate"
	LogicalsPath    = "/vpn/v1/logicals"
)

// API version headers
const (
	AppVersion = "linux-vpn@4.12.0"
	UserAgent  = "ProtonVPN/4.12.0 (Linux; Ubuntu)"
)
