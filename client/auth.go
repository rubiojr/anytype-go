// filepath: /home/epheo/dev/anytype-go/pkg/anytype/client/auth.go
package client

import (
	"context"

	"github.com/epheo/anytype-go"
)

// AuthClientImpl implements the AuthClient interface
type AuthClientImpl struct {
	client *ClientImpl
}

// DisplayCode initiates a secure authentication flow
func (ac *AuthClientImpl) DisplayCode(ctx context.Context, appName string) (*anytype.DisplayCodeResponse, error) {
	// Create the URL path without query parameters
	urlPath := "/auth/display_code"

	// Create the request with an empty body as per API definition
	req, err := ac.client.newRequest(ctx, "POST", urlPath, map[string]string{})

	// Add query parameter properly to the request URL
	if appName != "" && req != nil {
		q := req.URL.Query()
		q.Add("app_name", appName)
		req.URL.RawQuery = q.Encode()
	}
	if err != nil {
		return nil, err
	}

	// Make the request and parse the response
	var result anytype.DisplayCodeResponse
	if err := ac.client.doRequest(req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetToken completes the authentication flow by providing a code
func (ac *AuthClientImpl) GetToken(ctx context.Context, challengeID string, code string) (*anytype.TokenResponse, error) {
	// Create the URL path without query parameters
	urlPath := "/auth/token"

	// Create the request with an empty body as per API definition
	req, err := ac.client.newRequest(ctx, "POST", urlPath, map[string]string{})

	// Add query parameters properly to the request URL
	if req != nil {
		q := req.URL.Query()
		q.Add("challenge_id", challengeID)
		q.Add("code", code)
		req.URL.RawQuery = q.Encode()
	}

	if err != nil {
		return nil, err
	}

	// Make the request and parse the response
	var result anytype.TokenResponse
	if err := ac.client.doRequest(req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
