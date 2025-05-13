package anytype

import (
	"context"
)

// AuthClient provides operations for authentication
type AuthClient interface {
	// DisplayCode initiates a secure authentication flow
	DisplayCode(ctx context.Context, appName string) (*DisplayCodeResponse, error)

	// GetToken completes the authentication flow by providing a code
	GetToken(ctx context.Context, challengeID string, code string) (*TokenResponse, error)
}

// DisplayCodeResponse represents the response from the display_code endpoint
type DisplayCodeResponse struct {
	ChallengeID string `json:"challenge_id"`
}

// TokenResponse represents the response from the token endpoint
type TokenResponse struct {
	AppKey string `json:"app_key"`
}
