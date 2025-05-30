package auth

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/ProtonMail/go-srp"
	"golang.org/x/term"
	"protonvpn-wg-config-generate/internal/api"
	"protonvpn-wg-config-generate/internal/config"
)

// Client handles ProtonVPN authentication
type Client struct {
	config       *config.Config
	httpClient   *http.Client
	sessionStore *SessionStore
}

// NewClient creates a new authentication client
func NewClient(cfg *config.Config) *Client {
	return &Client{
		config:       cfg,
		sessionStore: NewSessionStore(),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
			},
		},
	}
}

// Authenticate performs the full authentication flow
func (c *Client) Authenticate() (*api.Session, error) {
	// Get username if not provided
	if err := c.ensureUsername(); err != nil {
		return nil, err
	}

	// Clear session if requested
	if c.config.ClearSession {
		fmt.Println("Clearing saved session...")
		_ = c.sessionStore.Delete()
	} else if !c.config.NoSession {
		// Try to load saved session
		savedSession, timeUntilExpiry, err := c.sessionStore.Load(c.config.Username)
		if err != nil {
			fmt.Printf("Warning: Failed to load saved session: %v\n", err)
		} else if savedSession != nil {
			// Verify the session is still valid by making a test request
			if c.verifySession(savedSession) {
				fmt.Printf("Using saved session (expires in %s)\n", humanizeDuration(timeUntilExpiry))
				return savedSession, nil
			}
			fmt.Println("Saved session invalid, re-authenticating...")
			_ = c.sessionStore.Delete()
		}
	}

	// Get password if not provided
	if err := c.ensurePassword(); err != nil {
		return nil, err
	}

	// Get auth info
	authInfo, err := c.getAuthInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get auth info: %w", err)
	}

	// Perform SRP authentication
	auth, err := srp.NewAuth(
		authInfo.Version,
		c.config.Username,
		[]byte(c.config.Password),
		authInfo.Salt,
		authInfo.Modulus,
		authInfo.ServerEphemeral,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create SRP auth: %w", err)
	}

	clientProofs, err := auth.GenerateProofs(2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate SRP proofs: %w", err)
	}

	// Build auth request
	authReq := map[string]interface{}{
		"Username":          c.config.Username,
		"ClientEphemeral":   base64.StdEncoding.EncodeToString(clientProofs.ClientEphemeral),
		"ClientProof":       base64.StdEncoding.EncodeToString(clientProofs.ClientProof),
		"SRPSession":        authInfo.SRPSession,
		"PersistentCookies": 0,
	}

	// Handle 2FA if needed
	if authInfo.TwoFA.Enabled == 1 && authInfo.TwoFA.TOTP == 1 {
		code, err := c.get2FACode()
		if err != nil {
			return nil, err
		}
		authReq["TwoFactorCode"] = code
	}

	// Send auth request
	session, err := c.sendAuthRequest(authReq)
	if err != nil {
		return nil, err
	}

	// Verify server proof
	if session.ServerProof != base64.StdEncoding.EncodeToString(clientProofs.ExpectedServerProof) {
		return nil, fmt.Errorf("server proof verification failed")
	}

	// Save the session for future use (unless disabled)
	if !c.config.NoSession {
		// Parse session duration
		sessionDuration, err := config.ParseSessionDuration(c.config.SessionDuration)
		if err != nil {
			fmt.Printf("Warning: Invalid session duration, using default: %v\n", err)
			sessionDuration = 0 // Default to no expiration
		}
		
		if err := c.sessionStore.Save(session, c.config.Username, sessionDuration); err != nil {
			fmt.Printf("Warning: Failed to save session: %v\n", err)
		}
	}

	return session, nil
}

func (c *Client) ensureUsername() error {
	if c.config.Username == "" {
		fmt.Print("Username (without @protonmail.com): ")
		reader := bufio.NewReader(os.Stdin)
		username, _ := reader.ReadString('\n')
		c.config.Username = strings.TrimSpace(username)
	}
	return nil
}

func (c *Client) ensurePassword() error {
	if c.config.Password == "" {
		fmt.Print("Password: ")
		passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Println()
		if err != nil {
			return fmt.Errorf("error reading password: %w", err)
		}
		c.config.Password = string(passwordBytes)
	}
	return nil
}

func (c *Client) get2FACode() (string, error) {
	fmt.Print("2FA Code: ")
	reader := bufio.NewReader(os.Stdin)
	code, _ := reader.ReadString('\n')
	return strings.TrimSpace(code), nil
}

func (c *Client) getAuthInfo() (*api.AuthInfoResponse, error) {
	reqBody := map[string]interface{}{
		"Username": c.config.Username,
		"Intent":   "Proton",
	}
	
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.config.APIURL+"/core/v4/auth/info", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(respBody))
	}

	var authInfo api.AuthInfoResponse
	if err := json.Unmarshal(respBody, &authInfo); err != nil {
		return nil, fmt.Errorf("failed to parse auth info: %w", err)
	}

	if authInfo.Code != 1000 {
		return nil, fmt.Errorf("failed to get auth info, code: %d", authInfo.Code)
	}

	// Validate required fields
	if authInfo.Modulus == "" {
		return nil, fmt.Errorf("received empty modulus from auth info")
	}
	if authInfo.ServerEphemeral == "" {
		return nil, fmt.Errorf("received empty server ephemeral from auth info")
	}

	return &authInfo, nil
}

func (c *Client) sendAuthRequest(authReq map[string]interface{}) (*api.Session, error) {
	body, err := json.Marshal(authReq)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.config.APIURL+"/core/v4/auth", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("authentication HTTP error %d: %s", resp.StatusCode, string(respBody))
	}


	var session api.Session
	if err := json.Unmarshal(respBody, &session); err != nil {
		return nil, err
	}

	// Handle mailbox password (should not be needed with single password mode)
	if session.Code == 10013 {
		return nil, fmt.Errorf("unexpected mailbox password request - account might still be in 2-password mode")
	}

	if session.Code != 1000 {
		errMsg := c.getErrorMessage(session.Code)
		return nil, fmt.Errorf(errMsg)
	}

	return &session, nil
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-pm-appversion", "linux-vpn@4.2.0")
	req.Header.Set("User-Agent", "ProtonVPN/4.2.0 (Linux; Ubuntu)")
}

func (c *Client) getErrorMessage(code int) string {
	switch code {
	case 8004:
		return "incorrect username or password"
	case 8002:
		return "password format is incorrect"
	case 9001:
		return "CAPTCHA verification required. This typically happens when ProtonVPN detects automated access. Try: 1) Login via web browser first, 2) Use a different IP, or 3) Wait some time before retrying"
	case 10002:
		return "2FA code is required"
	case 10003:
		return "invalid 2FA code"
	default:
		return fmt.Sprintf("authentication failed with code: %d", code)
	}
}

// verifySession checks if a saved session is still valid
func (c *Client) verifySession(session *api.Session) bool {
	// Make a simple request to verify the session
	req, err := http.NewRequest("GET", c.config.APIURL+"/vpn/v1/logicals", nil)
	if err != nil {
		return false
	}
	
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", session.AccessToken))
	req.Header.Set("x-pm-uid", session.UID)
	req.Header.Set("x-pm-appversion", "linux-vpn@4.2.0")
	req.Header.Set("User-Agent", "ProtonVPN/4.2.0 (Linux; Ubuntu)")
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	
	// If we get a 401, the session is invalid
	if resp.StatusCode == http.StatusUnauthorized {
		return false
	}
	
	// Any 2xx response means the session is valid
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

// humanizeDuration converts a duration to a human-readable format
func humanizeDuration(d time.Duration) string {
	if d < 0 {
		return "expired"
	}
	
	// Less than a minute
	if d < time.Minute {
		return "less than a minute"
	}
	
	// Less than an hour
	if d < time.Hour {
		minutes := int(d.Minutes())
		if minutes == 1 {
			return "1 minute"
		}
		return fmt.Sprintf("%d minutes", minutes)
	}
	
	// Less than a day
	if d < 24*time.Hour {
		hours := int(d.Hours())
		minutes := int(d.Minutes()) % 60
		if hours == 1 && minutes == 0 {
			return "1 hour"
		}
		if minutes == 0 {
			return fmt.Sprintf("%d hours", hours)
		}
		return fmt.Sprintf("%d hours %d minutes", hours, minutes)
	}
	
	// Days
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	
	// Less than a week
	if days < 7 {
		if days == 1 && hours == 0 {
			return "1 day"
		}
		if hours == 0 {
			return fmt.Sprintf("%d days", days)
		}
		return fmt.Sprintf("%d days %d hours", days, hours)
	}
	
	// Weeks to months
	if days < 30 {
		weeks := days / 7
		remainingDays := days % 7
		if weeks == 1 && remainingDays == 0 {
			return "1 week"
		}
		if remainingDays == 0 {
			return fmt.Sprintf("%d weeks", weeks)
		}
		return fmt.Sprintf("%d weeks %d days", weeks, remainingDays)
	}
	
	// Months to years
	if days < 365 {
		months := days / 30
		remainingDays := days % 30
		if months == 1 && remainingDays == 0 {
			return "1 month"
		}
		if remainingDays == 0 {
			return fmt.Sprintf("%d months", months)
		}
		return fmt.Sprintf("%d months %d days", months, remainingDays)
	}
	
	// Years
	years := days / 365
	remainingDays := days % 365
	if years == 1 && remainingDays == 0 {
		return "1 year"
	}
	if remainingDays == 0 {
		return fmt.Sprintf("%d years", years)
	}
	if remainingDays < 30 {
		return fmt.Sprintf("%d years %d days", years, remainingDays)
	}
	months := remainingDays / 30
	return fmt.Sprintf("%d years %d months", years, months)
}