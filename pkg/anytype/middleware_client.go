// Package anytype provides a Go client for interacting with the Anytype API.
package anytype

// WithRetry adds retry capability to the client with default settings.
//
// This middleware will automatically retry requests that fail with certain
// HTTP status codes, including 429 (rate limit) errors. It uses exponential
// backoff with jitter to prevent thundering herd problems.
//
// Example:
//
//	client := anytype.NewClient(
//	    anytype.WithAppKey("your-app-key"),
//	    anytype.WithToken("your-token"),
//	)
//	client.WithRetry() // Add retry with default settings
func (c *Client) WithRetry() *Client {
	options := DefaultRetryOptions()
	return c.WithRetryOptions(options)
}

// WithRetryOptions adds retry capability to the client with custom settings.
//
// This allows for fine-grained control over retry behavior, including
// max retries, delay between retries, and which status codes trigger retries.
//
// Example:
//
//	options := anytype.DefaultRetryOptions()
//	options.MaxRetries = 5
//	options.RetryableStatusCodes = []int{429, 500, 502, 503, 504}
//
//	client := anytype.NewClient(
//	    anytype.WithAppKey("your-app-key"),
//	    anytype.WithToken("your-token"),
//	)
//	client.WithRetryOptions(options)
func (c *Client) WithRetryOptions(options RetryOptions) *Client {
	middleware := NewRetryMiddleware(options)
	return c.WithMiddleware(middleware)
}

// WithLogging adds logging capability to the client with default settings.
//
// This middleware will log details about requests and responses,
// including method, URL, status code, and timing information.
//
// Example:
//
//	client := anytype.NewClient(
//	    anytype.WithAppKey("your-app-key"),
//	    anytype.WithToken("your-token"),
//	)
//	client.WithLogging() // Add logging with default settings
func (c *Client) WithLogging() *Client {
	options := DefaultLoggingOptions()
	options.Logger = c.logger
	return c.WithLoggingOptions(options)
}

// WithLoggingOptions adds logging capability to the client with custom settings.
//
// This allows for fine-grained control over logging behavior, including
// which parts of requests and responses are logged and at what level.
//
// Example:
//
//	options := anytype.DefaultLoggingOptions()
//	options.LogRequestBody = false
//	options.LogResponseBody = false
//	options.LogLevel = log.LevelDebug
//
//	client := anytype.NewClient(
//	    anytype.WithAppKey("your-app-key"),
//	    anytype.WithToken("your-token"),
//	)
//	client.WithLoggingOptions(options)
func (c *Client) WithLoggingOptions(options LoggingOptions) *Client {
	// If no logger is specified in options but the client has one, use that
	if options.Logger == nil && c.logger != nil {
		options.Logger = c.logger
	}

	middleware := NewLoggingMiddleware(options)
	return c.WithMiddleware(middleware)
}
