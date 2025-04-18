// Package anytype provides a Go client for interacting with the Anytype API.
package anytype

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/epheo/anytype-go/internal/log"
)

// LoggingOptions configures the logging middleware
type LoggingOptions struct {
	// LogLevel is the level at which to log requests and responses
	LogLevel log.Level
	// Logger is the logger to use
	Logger log.Logger
	// LogRequestHeaders determines whether request headers should be logged
	LogRequestHeaders bool
	// LogResponseHeaders determines whether response headers should be logged
	LogResponseHeaders bool
	// LogRequestBody determines whether request bodies should be logged
	LogRequestBody bool
	// LogResponseBody determines whether response bodies should be logged
	LogResponseBody bool
	// LogTiming determines whether request timing should be logged
	LogTiming bool
	// RedactedHeaders is a list of headers whose values should be redacted in logs
	RedactedHeaders []string
}

// DefaultLoggingOptions provides sensible default settings for the logging middleware
func DefaultLoggingOptions() LoggingOptions {
	return LoggingOptions{
		LogLevel:           log.LevelInfo,
		LogRequestHeaders:  true,
		LogResponseHeaders: true,
		LogRequestBody:     true,
		LogResponseBody:    true,
		LogTiming:          true,
		RedactedHeaders: []string{
			"Authorization",
			"Token",
			"Api-Key",
			"Apikey",
			"Secret",
		},
	}
}

// NewLoggingMiddleware creates a middleware that logs requests and responses
func NewLoggingMiddleware(options LoggingOptions) Middleware {
	return MiddlewareFunc(func(next http.RoundTripper) http.RoundTripper {
		return RoundTripFunc(func(req *http.Request) (*http.Response, error) {
			var requestBodyCopy []byte

			// Start timer
			start := time.Now()

			// Log request
			if options.Logger != nil {
				logWithLevel(options.Logger, options.LogLevel, ">> Request: %s %s", req.Method, req.URL.String())

				// Log request headers with sensitive data redacted
				if options.LogRequestHeaders {
					logHeaders(options.Logger, options.LogLevel, req.Header, options.RedactedHeaders)
				}

				// Log request body if present and configured
				if options.LogRequestBody && req.Body != nil {
					var err error
					requestBodyCopy, err = io.ReadAll(req.Body)
					if err != nil {
						options.Logger.Error("Failed to read request body for logging: %v", err)
					} else {
						// Log the request body, formatting JSON if possible
						logBody(options.Logger, options.LogLevel, requestBodyCopy, "Request Body")

						// Reset the request body so it can be read again
						req.Body = io.NopCloser(bytes.NewReader(requestBodyCopy))
					}
				}
			}

			// Execute the request
			resp, err := next.RoundTrip(req)

			// Calculate request duration
			duration := time.Since(start)

			// Log the response or error
			if options.Logger != nil {
				if err != nil {
					logWithLevel(options.Logger, options.LogLevel, "<< Error: %v (took %v)", err, duration)
					return nil, err
				}

				// Log response status and timing
				statusText := fmt.Sprintf("<< Response: %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
				if options.LogTiming {
					statusText += fmt.Sprintf(" (took %v)", duration)
				}
				logWithLevel(options.Logger, options.LogLevel, statusText)

				// Log response headers
				if options.LogResponseHeaders {
					logHeaders(options.Logger, options.LogLevel, resp.Header, options.RedactedHeaders)
				}

				// Log response body if configured
				if options.LogResponseBody && resp.Body != nil {
					// Read the response body for logging
					respBody, err := io.ReadAll(resp.Body)
					if err != nil {
						options.Logger.Error("Failed to read response body for logging: %v", err)
					} else {
						// Log the response body, formatting JSON if possible
						logBody(options.Logger, options.LogLevel, respBody, "Response Body")

						// Reset the response body so it can be read again
						resp.Body = io.NopCloser(bytes.NewReader(respBody))
					}
				}
			}

			return resp, err
		})
	})
}

// logWithLevel logs a message at the specified level
func logWithLevel(logger log.Logger, level log.Level, format string, args ...interface{}) {
	switch level {
	case log.LevelError:
		logger.Error(format, args...)
	case log.LevelInfo:
		logger.Info(format, args...)
	case log.LevelDebug:
		logger.Debug(format, args...)
	default:
		logger.Info(format, args...)
	}
}

// logHeaders logs HTTP headers with sensitive values redacted
func logHeaders(logger log.Logger, level log.Level, headers http.Header, redactedHeaders []string) {
	for key, values := range headers {
		value := values[0]
		if len(values) > 1 {
			value = fmt.Sprintf("%s (+%d more)", value, len(values)-1)
		}

		// Check if this header should be redacted
		for _, redacted := range redactedHeaders {
			if key == redacted {
				value = "[REDACTED]"
				break
			}
		}

		logWithLevel(logger, level, "  %s: %s", key, value)
	}
}

// logBody formats and logs the body content
func logBody(logger log.Logger, level log.Level, body []byte, prefix string) {
	if len(body) > 1000 {
		// If body is too large, truncate it for logging
		logWithLevel(logger, level, "  %s: %s... [truncated, %d bytes total]", prefix, string(body[:1000]), len(body))
		return
	}

	// Try to format as JSON for better readability
	var formattedJSON bytes.Buffer
	if json.Indent(&formattedJSON, body, "    ", "  ") == nil {
		logWithLevel(logger, level, "  %s (JSON): \n%s", prefix, formattedJSON.String())
	} else {
		// If not JSON or invalid JSON, log as plain text
		logWithLevel(logger, level, "  %s: %s", prefix, string(body))
	}
}
