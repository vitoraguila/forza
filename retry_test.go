package forza

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
)

func TestWithRetry_SuccessOnFirstAttempt(t *testing.T) {
	attempts := 0
	err := withRetry(context.Background(), 3, func() error {
		attempts++
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if attempts != 1 {
		t.Errorf("expected 1 attempt, got %d", attempts)
	}
}

func TestWithRetry_SuccessAfterRetries(t *testing.T) {
	attempts := 0
	err := withRetry(context.Background(), 3, func() error {
		attempts++
		if attempts < 3 {
			return &retryableError{
				err:        fmt.Errorf("%w: server error", ErrCompletionFailed),
				statusCode: http.StatusServiceUnavailable,
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestWithRetry_NonRetryableError(t *testing.T) {
	attempts := 0
	err := withRetry(context.Background(), 3, func() error {
		attempts++
		return &retryableError{
			err:        fmt.Errorf("%w: bad request", ErrCompletionFailed),
			statusCode: http.StatusBadRequest,
		}
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if attempts != 1 {
		t.Errorf("expected 1 attempt for non-retryable error, got %d", attempts)
	}
}

func TestWithRetry_NonRetryableErrorType(t *testing.T) {
	attempts := 0
	err := withRetry(context.Background(), 3, func() error {
		attempts++
		return errors.New("plain error, not retryable")
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if attempts != 1 {
		t.Errorf("expected 1 attempt for non-retryable error type, got %d", attempts)
	}
}

func TestWithRetry_ExhaustsMaxAttempts(t *testing.T) {
	attempts := 0
	err := withRetry(context.Background(), 3, func() error {
		attempts++
		return &retryableError{
			err:        fmt.Errorf("%w: rate limited", ErrCompletionFailed),
			statusCode: http.StatusTooManyRequests,
		}
	})
	if err == nil {
		t.Fatal("expected error after exhausting retries")
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
	if !errors.Is(err, ErrCompletionFailed) {
		t.Errorf("expected ErrCompletionFailed, got %v", err)
	}
}

func TestWithRetry_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	attempts := 0
	err := withRetry(ctx, 3, func() error {
		attempts++
		return &retryableError{
			err:        fmt.Errorf("server error"),
			statusCode: http.StatusServiceUnavailable,
		}
	})
	if err == nil {
		t.Fatal("expected error")
	}
	// Should return after first failed attempt when context is cancelled
	if attempts > 2 {
		t.Errorf("expected at most 2 attempts with cancelled context, got %d", attempts)
	}
}

func TestWithRetry_ZeroMaxAttempts(t *testing.T) {
	attempts := 0
	err := withRetry(context.Background(), 0, func() error {
		attempts++
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if attempts != 1 {
		t.Errorf("expected 1 attempt (minimum), got %d", attempts)
	}
}

func TestWithRetry_RetryableStatusCodes(t *testing.T) {
	retryableCodes := []int{429, 500, 502, 503, 504}
	for _, code := range retryableCodes {
		attempts := 0
		_ = withRetry(context.Background(), 2, func() error {
			attempts++
			if attempts == 1 {
				return &retryableError{
					err:        fmt.Errorf("error"),
					statusCode: code,
				}
			}
			return nil
		})
		if attempts != 2 {
			t.Errorf("status %d: expected 2 attempts, got %d", code, attempts)
		}
	}
}
