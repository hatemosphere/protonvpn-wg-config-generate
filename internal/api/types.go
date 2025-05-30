package api

// AuthInfoResponse represents the response from the auth info endpoint
type AuthInfoResponse struct {
	Code            int    `json:"Code"`
	Version         int    `json:"Version"`
	Modulus         string `json:"Modulus"`
	ServerEphemeral string `json:"ServerEphemeral"`
	Salt            string `json:"Salt"`
	SRPSession      string `json:"SRPSession"`
	TwoFA           struct {
		Enabled int `json:"Enabled"`
		TOTP    int `json:"TOTP"`
	} `json:"2FA"`
}

// AuthRequest represents the authentication request payload
type AuthRequest struct {
	Username        string `json:"Username"`
	ClientEphemeral string `json:"ClientEphemeral"`
	ClientProof     string `json:"ClientProof"`
	SRPSession      string `json:"SRPSession"`
	TwoFactorCode   string `json:"TwoFactorCode,omitempty"`
}

// Session represents a ProtonVPN session
type Session struct {
	Code               int      `json:"Code"`
	AccessToken        string   `json:"AccessToken"`
	RefreshToken       string   `json:"RefreshToken"`
	TokenType          string   `json:"TokenType"`
	Scopes             []string `json:"Scopes"`
	UID                string   `json:"UID"`
	UserID             string   `json:"UserID"`
	EventID            string   `json:"EventID"`
	ServerProof        string   `json:"ServerProof"`
	PasswordMode       int      `json:"PasswordMode"`
	ExpiresIn          int      `json:"ExpiresIn"` // Session expiration in seconds
	TwoFA              struct {
		Enabled int `json:"Enabled"`
		TOTP    int `json:"TOTP"`
	} `json:"2FA"`
}

// VPNInfo represents VPN certificate information
type VPNInfo struct {
	Code                  int    `json:"Code"`
	SerialNumber         string `json:"SerialNumber"`
	ClientKeyFingerprint string `json:"ClientKeyFingerprint"`
	ClientKey            string `json:"ClientKey"`
	Certificate          string `json:"Certificate"`
	ExpirationTime       int64  `json:"ExpirationTime"`
	RefreshTime          int64  `json:"RefreshTime"`
	Mode                 string `json:"Mode"`
	DeviceName           string `json:"DeviceName"`
	ServerPublicKeyMode  string `json:"ServerPublicKeyMode"`
	ServerPublicKey      string `json:"ServerPublicKey"`
	Features             struct {
		Bouncing         bool `json:"bouncing"`
		ModerateNAT      bool `json:"moderate-nat"`
		NetshieldLevel   int  `json:"netshield-level"`
		PortForwarding   bool `json:"port-forwarding"`
		VPNAccelerator   bool `json:"vpn-accelerator"`
	} `json:"Features"`
}

// LogicalServer represents a ProtonVPN logical server
type LogicalServer struct {
	ID           string `json:"ID"`
	Name         string `json:"Name"`
	EntryCountry string `json:"EntryCountry"`
	ExitCountry  string `json:"ExitCountry"`
	Domain       string `json:"Domain"`
	Tier         int    `json:"Tier"`
	Features     int    `json:"Features"`
	Region       string `json:"Region"`
	City         string `json:"City"`
	Score        float64 `json:"Score"`
	Load         int     `json:"Load"`
	Status       int     `json:"Status"`
	Servers      []PhysicalServer `json:"Servers"`
	HostCountry  string `json:"HostCountry"`
	Location     struct {
		Lat  float64 `json:"Lat"`
		Long float64 `json:"Long"`
	} `json:"Location"`
}

// PhysicalServer represents a physical VPN server
type PhysicalServer struct {
	ID                 string `json:"ID"`
	EntryIP            string `json:"EntryIP"`
	ExitIP             string `json:"ExitIP"`
	Domain             string `json:"Domain"`
	Status             int    `json:"Status"`
	Label              string `json:"Label"`
	X25519PublicKey    string `json:"X25519PublicKey"`
	Generation         int    `json:"Generation"`
	ServicesDownReason string `json:"ServicesDownReason"`
}

// LogicalsResponse represents the response from the logicals endpoint
type LogicalsResponse struct {
	Code           int             `json:"Code"`
	LogicalServers []LogicalServer `json:"LogicalServers"`
}

// Server feature constants
const (
	FeatureSecureCore = 1
	FeatureTor        = 2
	FeatureP2P        = 4
	FeatureStreaming  = 8
	FeatureIPv6       = 16
)

// Server tier constants
const (
	TierFree = 0
	TierPlus = 2
	TierPM   = 3
)

// Password mode constants
const (
	PasswordModeSingle = 1
	PasswordModeTwo    = 2
)