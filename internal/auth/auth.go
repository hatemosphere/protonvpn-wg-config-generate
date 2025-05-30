// Package auth handles ProtonVPN authentication using the SRP protocol.
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
	"protonvpn-wg-config-generate/internal/constants"
	"protonvpn-wg-config-generate/pkg/timeutil"
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
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: false,
					MinVersion:         tls.VersionTLS12,
				},
			},
		},
	}
}

// handleSessionRefresh attempts to refresh a session and save it if successful
func (c *Client) handleSessionRefresh(savedSession *api.Session, reason string) (*api.Session, error) {
	fmt.Println(reason)
	refreshedSession, err := RefreshSession(c.httpClient, c.config.APIURL, savedSession)
	if err != nil {
		fmt.Printf("Token refresh failed: %v\n", err)
		fmt.Println("Re-authenticating with password...")
		fmt.Println("(Your trusted device status for MFA will be preserved)")
		_ = c.sessionStore.Delete()
		return nil, err
	}

	fmt.Println("Session refreshed successfully!")
	// Check if refresh token was rotated
	if savedSession.RefreshToken != refreshedSession.RefreshToken {
		fmt.Println("Refresh token was rotated")
	}

	// Save the refreshed session
	if !c.config.NoSession {
		sessionDuration, _ := timeutil.ParseSessionDuration(c.config.SessionDuration)
		if err := c.sessionStore.Save(refreshedSession, c.config.Username, sessionDuration); err != nil {
			fmt.Printf("Warning: Failed to save refreshed session: %v\n", err)
		}
	}

	return refreshedSession, nil
}

// tryExistingSession attempts to use an existing saved session
func (c *Client) tryExistingSession() (*api.Session, error) {
	savedSession, timeUntilExpiry, err := c.sessionStore.Load(c.config.Username)
	if err != nil {
		fmt.Printf("Warning: Failed to load saved session: %v\n", err)
		return nil, err
	}

	if savedSession == nil {
		return nil, nil
	}

	// Determine what to do with the saved session
	switch {
	case c.config.ForceRefresh:
		reason := fmt.Sprintf("Forcing session refresh (current session expires in %s)", timeutil.HumanizeDuration(timeUntilExpiry))
		return c.handleSessionRefresh(savedSession, reason)

	case timeUntilExpiry < time.Duration(constants.SessionRefreshDays)*24*time.Hour && timeUntilExpiry > 0:
		reason := fmt.Sprintf("Session expires soon (in %s), attempting refresh...", timeutil.HumanizeDuration(timeUntilExpiry))
		return c.handleSessionRefresh(savedSession, reason)

	case VerifySession(c.httpClient, c.config.APIURL, savedSession):
		fmt.Printf("Using saved session (expires in %s)\n", timeutil.HumanizeDuration(timeUntilExpiry))
		return savedSession, nil

	default:
		fmt.Println("Saved session invalid, re-authenticating...")
		_ = c.sessionStore.Delete()
		return nil, nil
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
		// Try to use existing session
		if session, err := c.tryExistingSession(); err == nil && session != nil {
			return session, nil
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
		sessionDuration, err := timeutil.ParseSessionDuration(c.config.SessionDuration)
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
		passwordBytes, err := term.ReadPassword(syscall.Stdin)
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
	defer func() { _ = resp.Body.Close() }()

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
	defer func() { _ = resp.Body.Close() }()

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
		return nil, NewError(session.Code)
	}

	return &session, nil
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-pm-appversion", constants.AppVersion)
	req.Header.Set("User-Agent", constants.UserAgent)
}
