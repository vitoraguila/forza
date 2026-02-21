package forza

import (
	"context"
	"math/rand"
	"net/http"
	"time"
)

// retryableStatusCodes are HTTP status codes that warrant a retry.
var retryableStatusCodes = map[int]bool{
	http.StatusTooManyRequests:     true, // 429
	http.StatusInternalServerError: true, // 500
	http.StatusBadGateway:          true, // 502
	http.StatusServiceUnavailable:  true, // 503
	http.StatusGatewayTimeout:      true, // 504
}

// retryableError wraps an error with an HTTP status code to enable retry decisions.
type retryableError struct {
	err        error
	statusCode int
}

func (e *retryableError) Error() string { return e.err.Error() }
func (e *retryableError) Unwrap() error { return e.err }

// withRetry executes fn up to maxAttempts times with exponential backoff and jitter.
// Only retries if the error is a retryableError with a retryable status code.
func withRetry(ctx context.Context, maxAttempts int, fn func() error) error {
	if maxAttempts <= 0 {
		maxAttempts = 1
	}

	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		lastErr = fn()
		if lastErr == nil {
			return nil
		}

		// Only retry on retryable errors
		if re, ok := lastErr.(*retryableError); ok {
			if !retryableStatusCodes[re.statusCode] {
				return lastErr
			}
		} else {
			return lastErr
		}

		// Don't sleep after the last attempt
		if attempt < maxAttempts-1 {
			backoff := time.Duration(1<<uint(attempt)) * 500 * time.Millisecond
			jitter := time.Duration(rand.Int63n(int64(backoff / 2)))
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff + jitter):
			}
		}
	}
	return lastErr
}
