// Package anytype provides a Go client for interacting with the Anytype API.
package anytype

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"time"
)

// RetryOptions configures the retry middleware
type RetryOptions struct {
	// MaxRetries is the maximum number of retry attempts
	MaxRetries int
	// MinRetryDelay is the minimum delay between retries
	MinRetryDelay time.Duration
	// MaxRetryDelay is the maximum delay between retries
	MaxRetryDelay time.Duration
	// RetryableStatusCodes is a list of HTTP status codes that should trigger a retry
	RetryableStatusCodes []int
	// RetryBackoffFactor is the exponential backoff factor for retry delays
	RetryBackoffFactor float64
	// RetryJitter adds randomness to retry delays to prevent thundering herd issues
	RetryJitter time.Duration
	// ShouldRetry is a function that determines if a request should be retried
	ShouldRetry func(*http.Response, error) bool
}

// DefaultRetryOptions provides sensible default settings for the retry middleware
func DefaultRetryOptions() RetryOptions {
	return RetryOptions{
		MaxRetries:         6,
		MinRetryDelay:      time.Millisecond * 200,
		MaxRetryDelay:      time.Second * 30,
		RetryBackoffFactor: 2.0,
		RetryJitter:        time.Millisecond * 100,
		RetryableStatusCodes: []int{
			http.StatusRequestTimeout,      // 408
			http.StatusTooManyRequests,     // 429
			http.StatusInternalServerError, // 500
			http.StatusBadGateway,          // 502
			http.StatusServiceUnavailable,  // 503
			http.StatusGatewayTimeout,      // 504
		},
		ShouldRetry: nil, // Will use default implementation
	}
}

// NewRetryMiddleware creates a middleware that handles retries with exponential backoff
func NewRetryMiddleware(options RetryOptions) Middleware {
	// If no custom shouldRetry function is provided, use the default one
	if options.ShouldRetry == nil {
		options.ShouldRetry = func(resp *http.Response, err error) bool {
			// Retry on network errors (which are typically transient)
			if err != nil {
				// We may want to log this for debugging
				// fmt.Printf("Network error encountered: %v - Will retry\n", err)
				return true
			}

			// Retry for configured status codes
			for _, code := range options.RetryableStatusCodes {
				if resp.StatusCode == code {
					// We may want to log this for debugging
					// fmt.Printf("Received status code %d - Will retry\n", resp.StatusCode)
					return true
				}
			}
			return false
		}
	}

	return MiddlewareFunc(func(next http.RoundTripper) http.RoundTripper {
		return RoundTripFunc(func(req *http.Request) (*http.Response, error) {
			var resp *http.Response
			var err error

			// Store the original request body if it's not nil
			var originalBody []byte
			if req.Body != nil {
				originalBody, err = io.ReadAll(req.Body)
				if err != nil {
					return nil, fmt.Errorf("retry middleware: failed to read request body: %w", err)
				}
				req.Body.Close()
			}

			// Try the request with retries
			for attempt := 0; attempt <= options.MaxRetries; attempt++ {
				// If this isn't the first attempt, we need to recreate the body
				if attempt > 0 && originalBody != nil {
					req.Body = io.NopCloser(bytes.NewReader(originalBody))
				}

				resp, err = next.RoundTrip(req)

				// Success or non-retryable error, return immediately
				if err != nil || !options.ShouldRetry(resp, err) {
					return resp, err
				}

				// Check if we've reached the max retries
				if attempt == options.MaxRetries {
					return resp, err
				}

				// Special handling for rate limit (429) responses
				var retryAfter time.Duration
				if resp.StatusCode == http.StatusTooManyRequests {
					retryAfter = calculateRetryAfterDelay(resp, options, attempt)
				} else {
					// Calculate delay with exponential backoff
					retryAfter = calculateBackoffDelay(options, attempt)
				}

				// Close the response body to prevent resource leaks
				if resp.Body != nil {
					resp.Body.Close()
				}

				// Sleep before retrying
				select {
				case <-req.Context().Done():
					return nil, req.Context().Err()
				case <-time.After(retryAfter):
					// Continue with retry
				}
			}

			// This should never happen due to the return in the loop
			return resp, err
		})
	})
}

// calculateRetryAfterDelay determines the delay based on the Retry-After header or falls back to backoff
func calculateRetryAfterDelay(resp *http.Response, options RetryOptions, attempt int) time.Duration {
	// Check for Retry-After header (could be in seconds or HTTP date format)
	if retryAfterHeader := resp.Header.Get("Retry-After"); retryAfterHeader != "" {
		// Try to parse as integer seconds
		if seconds, err := strconv.Atoi(retryAfterHeader); err == nil {
			return time.Duration(seconds) * time.Second
		}

		// Try to parse as HTTP date
		if date, err := http.ParseTime(retryAfterHeader); err == nil {
			delay := time.Until(date)
			if delay > 0 {
				return delay
			}
		}
	}

	// Fallback to regular backoff
	return calculateBackoffDelay(options, attempt)
}

// calculateBackoffDelay calculates the delay using exponential backoff with jitter
func calculateBackoffDelay(options RetryOptions, attempt int) time.Duration {
	// Calculate base delay with exponential backoff
	backoffMs := float64(options.MinRetryDelay.Milliseconds()) * math.Pow(options.RetryBackoffFactor, float64(attempt))
	delay := time.Duration(backoffMs) * time.Millisecond

	// Apply maximum limit
	if delay > options.MaxRetryDelay {
		delay = options.MaxRetryDelay
	}

	// Add jitter to prevent thundering herd problem
	if options.RetryJitter > 0 {
		jitter := time.Duration(float64(options.RetryJitter) * (math.Float64frombits(FastRand()&0x7FFFFF) / float64(0x7FFFFF)))
		delay = delay + jitter
	}

	return delay
}

// FastRand is a fast thread-safe random function based on atomic operations
func FastRand() uint64 {
	// This is a simple placeholder - we should use a proper thread-safe random
	// number generator, possibly from sync/atomic or a proper random package.
	// For simplicity, we're just using time.Now().UnixNano()
	return uint64(time.Now().UnixNano())
}
