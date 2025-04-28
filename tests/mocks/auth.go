package mocks

import (
	"context"

	"github.com/epheo/anytype-go"
)

// MockAuthService implements the anytype.AuthService interface for testing
type MockAuthService struct {
	DisplayCodeFunc func(ctx context.Context, appName string) (*anytype.DisplayCodeResponse, error)
	GetTokenFunc    func(ctx context.Context, challengeID, code string) (*anytype.TokenResponse, error)
}

// NewMockAuthService creates a new instance of MockAuthService with default implementations
func NewMockAuthService() *MockAuthService {
	return &MockAuthService{
		DisplayCodeFunc: func(ctx context.Context, appName string) (*anytype.DisplayCodeResponse, error) {
			return &anytype.DisplayCodeResponse{
				ChallengeID: "mock-challenge-id",
			}, nil
		},
		GetTokenFunc: func(ctx context.Context, challengeID, code string) (*anytype.TokenResponse, error) {
			return &anytype.TokenResponse{
				AppKey:       "mock-app-key",
				SessionToken: "mock-session-token",
			}, nil
		},
	}
}

// DisplayCode calls the mock implementation
func (s *MockAuthService) DisplayCode(ctx context.Context, appName string) (*anytype.DisplayCodeResponse, error) {
	return s.DisplayCodeFunc(ctx, appName)
}

// GetToken calls the mock implementation
func (s *MockAuthService) GetToken(ctx context.Context, challengeID, code string) (*anytype.TokenResponse, error) {
	return s.GetTokenFunc(ctx, challengeID, code)
}
