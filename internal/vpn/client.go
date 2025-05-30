package vpn

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ProtonVPN/go-vpn-lib/ed25519"
	"protonvpn-wg-config-generate/internal/api"
	"protonvpn-wg-config-generate/internal/config"
)

// Client handles VPN operations
type Client struct {
	config     *config.Config
	session    *api.Session
	httpClient *http.Client
}

// NewClient creates a new VPN client
func NewClient(cfg *config.Config, session *api.Session) *Client {
	return &Client{
		config:     cfg,
		session:    session,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// GetCertificate generates a VPN certificate
func (c *Client) GetCertificate(keyPair *ed25519.KeyPair) (*api.VPNInfo, error) {
	publicKeyPEM, err := keyPair.PublicKeyPKIXPem()
	if err != nil {
		return nil, fmt.Errorf("failed to get public key PEM: %w", err)
	}

	// Use provided device name or generate one
	deviceName := c.config.DeviceName
	if deviceName == "" {
		deviceName = fmt.Sprintf("WireGuard-%s-%d", c.config.Username, time.Now().Unix())
	}

	// Parse duration
	duration, err := config.ParseDuration(c.config.Duration)
	if err != nil {
		return nil, fmt.Errorf("failed to parse duration: %w", err)
	}

	certReq := map[string]interface{}{
		"ClientPublicKey":     publicKeyPEM,
		"ClientPublicKeyMode": "EC",
		"Mode":                "persistent", // Create persistent configuration
		"DeviceName":          deviceName,
		"Duration":            duration,
		"Features": map[string]interface{}{
			"netshield-level":  0,
			"moderate-nat":     false,
			"port-forwarding":  false,
			"vpn-accelerator":  c.config.EnableAccelerator,
			"bouncing":         c.config.EnableAccelerator,
		},
	}

	certJSON, err := json.Marshal(certReq)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.config.APIURL+"/vpn/v1/certificate", bytes.NewBuffer(certJSON))
	if err != nil {
		return nil, err
	}

	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var vpnInfo api.VPNInfo
	if err := json.Unmarshal(body, &vpnInfo); err != nil {
		return nil, err
	}

	if vpnInfo.Code != 1000 {
		return nil, fmt.Errorf("failed to get VPN certificate, code: %d", vpnInfo.Code)
	}

	return &vpnInfo, nil
}

// GetServers fetches the list of VPN servers
func (c *Client) GetServers() ([]api.LogicalServer, error) {
	req, err := http.NewRequest("GET", c.config.APIURL+"/vpn/v1/logicals", nil)
	if err != nil {
		return nil, err
	}

	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response api.LogicalsResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	if response.Code != 1000 {
		return nil, fmt.Errorf("API returned error code: %d", response.Code)
	}

	return response.LogicalServers, nil
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.session.AccessToken))
	req.Header.Set("x-pm-uid", c.session.UID)
	req.Header.Set("x-pm-appversion", "linux-vpn@4.2.0")
	req.Header.Set("User-Agent", "ProtonVPN/4.2.0 (Linux; Ubuntu)")
}