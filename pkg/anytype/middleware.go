// Package anytype provides a Go client for interacting with the Anytype API.
package anytype

import (
	"io"
	"net/http"
)

// RoundTripFunc is a function type that implements the http.RoundTripper interface
type RoundTripFunc func(*http.Request) (*http.Response, error)

// RoundTrip implements the http.RoundTripper interface
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// Middleware defines the interface for all client middleware
type Middleware interface {
	// Execute runs the middleware and calls the next middleware in the chain
	Execute(next http.RoundTripper) http.RoundTripper
}

// MiddlewareFunc is a function type that implements the Middleware interface
type MiddlewareFunc func(http.RoundTripper) http.RoundTripper

// Execute implements the Middleware interface
func (f MiddlewareFunc) Execute(next http.RoundTripper) http.RoundTripper {
	return f(next)
}

// Chain combines multiple middlewares into a single middleware
func Chain(middlewares ...Middleware) Middleware {
	return MiddlewareFunc(func(next http.RoundTripper) http.RoundTripper {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i].Execute(next)
		}
		return next
	})
}

// RequestContext holds context information for the middleware chain
type RequestContext struct {
	Method      string
	Path        string
	Body        io.Reader
	OriginalURL string
	Attempt     int
}

// WithMiddleware applies a middleware to the client's HTTP client
func (c *Client) WithMiddleware(middleware Middleware) *Client {
	if c.httpClient == nil {
		c.httpClient = &http.Client{Timeout: httpTimeout}
	}

	// Extract the existing transport or create a new one if none exists
	var transport http.RoundTripper
	if c.httpClient.Transport == nil {
		transport = http.DefaultTransport
	} else {
		transport = c.httpClient.Transport
	}

	// Apply the middleware to the transport
	c.httpClient.Transport = middleware.Execute(transport)

	return c
}
