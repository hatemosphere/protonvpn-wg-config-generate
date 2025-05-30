package auth

import "fmt"

// Error codes from ProtonVPN API
const (
	CodeSuccess              = 1000
	CodeWrongPassword        = 8002
	CodeWrongPasswordFormat  = 8004 // Different error code for password format
	CodeCaptchaRequired      = 9001
	Code2FARequired          = 10002
	CodeInvalid2FA           = 10003
	CodeMailboxPasswordError = 10013
)

// Error represents an authentication error with ProtonVPN-specific error code
type Error struct {
	Code    int
	Message string
}

// Error implements the error interface
func (e Error) Error() string {
	return e.Message
}

// NewError creates a new authentication error from an API response code
func NewError(code int) error {
	message := getErrorMessage(code)
	return Error{
		Code:    code,
		Message: message,
	}
}

// getErrorMessage returns a human-readable error message for a given error code
func getErrorMessage(code int) string {
	switch code {
	case CodeWrongPassword:
		return "incorrect username or password"
	case CodeWrongPasswordFormat:
		return "password format is incorrect"
	case CodeCaptchaRequired:
		return "CAPTCHA verification required"
	case Code2FARequired:
		return "2FA code is required"
	case CodeInvalid2FA:
		return "invalid 2FA code"
	case CodeMailboxPasswordError:
		return "unexpected mailbox password request - account might still be in 2-password mode"
	default:
		return fmt.Sprintf("authentication failed with code: %d", code)
	}
}

// Is2FAError checks if the error is a 2FA-related error
func Is2FAError(err error) bool {
	authErr, ok := err.(Error)
	if !ok {
		return false
	}
	return authErr.Code == Code2FARequired || authErr.Code == CodeInvalid2FA
}

// IsCaptchaError checks if the error requires CAPTCHA verification
func IsCaptchaError(err error) bool {
	authErr, ok := err.(Error)
	return ok && authErr.Code == CodeCaptchaRequired
}
