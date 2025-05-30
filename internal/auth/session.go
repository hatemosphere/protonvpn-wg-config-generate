package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"protonvpn-wg-config-generate/internal/api"
)

const (
	sessionFileName = ".protonvpn-session.json"
	sessionFileMode = 0600 // Read/write for owner only
	
	// Maximum reasonable session duration (10 years for "no expiration")
	maxSessionDuration = 10 * 365 * 24 * time.Hour
)

// SessionStore handles persistent session storage
type SessionStore struct {
	filePath string
}

// NewSessionStore creates a new session store
func NewSessionStore() *SessionStore {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory
		homeDir = "."
	}
	
	return &SessionStore{
		filePath: filepath.Join(homeDir, sessionFileName),
	}
}

// SavedSession represents a session with metadata
type SavedSession struct {
	Session   *api.Session `json:"session"`
	Username  string       `json:"username"`
	SavedAt   time.Time    `json:"saved_at"`
	ExpiresAt time.Time    `json:"expires_at"`
}


// Save stores the session to disk
func (s *SessionStore) Save(session *api.Session, username string, duration time.Duration) error {
	savedSession := &SavedSession{
		Session:   session,
		Username:  username,
		SavedAt:   time.Now(),
	}
	
	// Calculate expiration based on API response
	apiExpiration := time.Now().Add(time.Duration(session.ExpiresIn) * time.Second)
	
	if duration == 0 {
		// Use the API's expiration
		savedSession.ExpiresAt = apiExpiration
	} else {
		// Use the user-specified duration, but cap it at API expiration
		userExpiration := time.Now().Add(duration)
		if userExpiration.After(apiExpiration) {
			savedSession.ExpiresAt = apiExpiration
		} else {
			savedSession.ExpiresAt = userExpiration
		}
	}
	
	data, err := json.MarshalIndent(savedSession, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}
	
	err = os.WriteFile(s.filePath, data, sessionFileMode)
	if err != nil {
		return fmt.Errorf("failed to write session file: %w", err)
	}
	
	return nil
}

// Load retrieves a saved session from disk
func (s *SessionStore) Load(username string) (*api.Session, time.Duration, error) {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, 0, nil // No saved session
		}
		return nil, 0, fmt.Errorf("failed to read session file: %w", err)
	}
	
	var savedSession SavedSession
	err = json.Unmarshal(data, &savedSession)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to unmarshal session: %w", err)
	}
	
	// Check if session is for the same user
	if savedSession.Username != username {
		return nil, 0, nil
	}
	
	// Check if session has expired
	now := time.Now()
	if now.After(savedSession.ExpiresAt) {
		// Delete expired session
		_ = s.Delete()
		return nil, 0, nil
	}
	
	// Calculate time until expiration
	timeUntilExpiry := savedSession.ExpiresAt.Sub(now)
	
	return savedSession.Session, timeUntilExpiry, nil
}

// Delete removes the saved session
func (s *SessionStore) Delete() error {
	err := os.Remove(s.filePath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete session file: %w", err)
	}
	return nil
}

// GetPath returns the session file path
func (s *SessionStore) GetPath() string {
	return s.filePath
}