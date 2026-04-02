package aws

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

// fastConfig returns a RetryConfig with near-zero delays so tests run in <10ms.
func fastConfig(maxAttempts int) *RetryConfig {
	return &RetryConfig{
		MaxAttempts:     maxAttempts,
		InitialDelay:    time.Millisecond,
		MaxDelay:        5 * time.Millisecond,
		Multiplier:      2.0,
		JitterFraction:  0,
		RetryableErrors: []string{"transient"},
	}
}

// ── ExponentialBackoff ─────────────────────────────────────────────────────

func TestExponentialBackoff(t *testing.T) {
	tests := []struct {
		attempt int
		initial time.Duration
		max     time.Duration
		mult    float64
		want    time.Duration
	}{
		{1, time.Second, 30 * time.Second, 2.0, time.Second},
		{2, time.Second, 30 * time.Second, 2.0, 2 * time.Second},
		{3, time.Second, 30 * time.Second, 2.0, 4 * time.Second},
		{5, time.Second, 30 * time.Second, 2.0, 16 * time.Second},
		// Capped at max
		{10, time.Second, 30 * time.Second, 2.0, 30 * time.Second},
		// Multiplier 1 — flat
		{3, time.Second, 30 * time.Second, 1.0, time.Second},
	}
	for _, tt := range tests {
		got := ExponentialBackoff(tt.attempt, tt.initial, tt.max, tt.mult)
		if got != tt.want {
			t.Errorf("attempt=%d got=%s want=%s", tt.attempt, got, tt.want)
		}
	}
}

// ── isRetryableError ──────────────────────────────────────────────────────

func TestIsRetryableError(t *testing.T) {
	cfg := DefaultRetryConfig()

	retryable := []string{
		"Throttling: rate exceeded",
		"ThrottlingException",
		"ServiceUnavailable",
		"InternalError",
		"InsufficientInstanceCapacity",
		"connection reset by peer",
		"timeout waiting for response",
		"temporarily unavailable, try again",
	}
	for _, msg := range retryable {
		if !isRetryableError(fmt.Errorf("%s", msg), cfg) {
			t.Errorf("expected retryable: %q", msg)
		}
	}

	nonRetryable := []string{
		"InvalidInstanceID.NotFound",
		"AuthFailure: invalid credentials",
		"ValidationError: missing required field",
		"AccessDenied",
	}
	for _, msg := range nonRetryable {
		if isRetryableError(fmt.Errorf("%s", msg), cfg) {
			t.Errorf("expected non-retryable: %q", msg)
		}
	}

	if isRetryableError(nil, cfg) {
		t.Error("nil error should not be retryable")
	}
}

func TestIsRetryableError_CustomPatterns(t *testing.T) {
	cfg := &RetryConfig{
		RetryableErrors: []string{"CUSTOM_ERR"},
	}
	if !isRetryableError(fmt.Errorf("got CUSTOM_ERR from server"), cfg) {
		t.Error("custom pattern should be retryable")
	}
	if isRetryableError(fmt.Errorf("some other error"), cfg) {
		t.Error("non-matching error should not be retryable")
	}
}

// ── isTimeoutError ────────────────────────────────────────────────────────

func TestIsTimeoutError(t *testing.T) {
	if !isTimeoutError(fmt.Errorf("operation timeout")) {
		t.Error("'timeout' string should match")
	}
	if !isTimeoutError(fmt.Errorf("context deadline exceeded")) {
		t.Error("'deadline exceeded' should match")
	}
	if isTimeoutError(fmt.Errorf("access denied")) {
		t.Error("unrelated error should not match")
	}
}

// ── WithRetry ─────────────────────────────────────────────────────────────

func TestWithRetry_SuccessFirstAttempt(t *testing.T) {
	calls := 0
	err := WithRetry(context.Background(), fastConfig(3), func(ctx context.Context) error {
		calls++
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 1 {
		t.Errorf("calls=%d, want 1", calls)
	}
}

func TestWithRetry_SucceedsAfterRetries(t *testing.T) {
	calls := 0
	err := WithRetry(context.Background(), fastConfig(5), func(ctx context.Context) error {
		calls++
		if calls < 3 {
			return fmt.Errorf("transient failure")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 3 {
		t.Errorf("calls=%d, want 3", calls)
	}
}

func TestWithRetry_ExhaustsRetries(t *testing.T) {
	cfg := fastConfig(3)
	calls := 0
	err := WithRetry(context.Background(), cfg, func(ctx context.Context) error {
		calls++
		return fmt.Errorf("transient failure")
	})
	if err == nil {
		t.Fatal("expected error after exhausting retries")
	}
	if calls != cfg.MaxAttempts {
		t.Errorf("calls=%d, want %d", calls, cfg.MaxAttempts)
	}
}

func TestWithRetry_NonRetryableFailsImmediately(t *testing.T) {
	calls := 0
	err := WithRetry(context.Background(), fastConfig(5), func(ctx context.Context) error {
		calls++
		return fmt.Errorf("AccessDenied: not authorized")
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if calls != 1 {
		t.Errorf("calls=%d, want 1 (non-retryable should not retry)", calls)
	}
}

func TestWithRetry_NilConfigUsesDefaults(t *testing.T) {
	// nil config should not panic and should use DefaultRetryConfig
	calls := 0
	err := WithRetry(context.Background(), nil, func(ctx context.Context) error {
		calls++
		return nil
	})
	if err != nil || calls != 1 {
		t.Errorf("nil config: err=%v calls=%d", err, calls)
	}
}

func TestWithRetry_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	calls := 0
	err := WithRetry(ctx, fastConfig(10), func(ctx context.Context) error {
		calls++
		if calls == 2 {
			cancel()
		}
		return fmt.Errorf("transient failure")
	})

	if err == nil {
		t.Fatal("expected error on context cancel")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("err should wrap context.Canceled, got: %v", err)
	}
}

// ── RetryTracker ──────────────────────────────────────────────────────────

func TestRetryTracker_Metrics(t *testing.T) {
	tracker := NewRetryTracker()

	tracker.RecordSuccess(1) // no retry
	tracker.RecordSuccess(3) // 2 retries before success
	tracker.RecordFailure(5) // all retries exhausted

	m := tracker.GetMetrics()
	if m.TotalOperations != 3 {
		t.Errorf("TotalOperations=%d, want 3", m.TotalOperations)
	}
	if m.SuccessfulRetries != 1 {
		t.Errorf("SuccessfulRetries=%d, want 1 (only attempts>1 count)", m.SuccessfulRetries)
	}
	if m.FailedRetries != 1 {
		t.Errorf("FailedRetries=%d, want 1", m.FailedRetries)
	}
	// AverageAttempts = (1+3+5)/3 = 3.0
	if m.AverageAttempts != 3.0 {
		t.Errorf("AverageAttempts=%.2f, want 3.0", m.AverageAttempts)
	}
}

// ── WithRetryAndTracking ──────────────────────────────────────────────────

func TestWithRetryAndTracking_Success(t *testing.T) {
	tracker := NewRetryTracker()
	calls := 0
	err := WithRetryAndTracking(context.Background(), fastConfig(3), tracker, func(ctx context.Context) error {
		calls++
		if calls < 2 {
			return fmt.Errorf("transient failure")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := tracker.GetMetrics()
	if m.TotalOperations != 1 {
		t.Errorf("TotalOperations=%d, want 1", m.TotalOperations)
	}
	if m.SuccessfulRetries != 1 {
		t.Errorf("SuccessfulRetries=%d, want 1", m.SuccessfulRetries)
	}
}

func TestWithRetryAndTracking_Failure(t *testing.T) {
	tracker := NewRetryTracker()
	err := WithRetryAndTracking(context.Background(), fastConfig(2), tracker, func(ctx context.Context) error {
		return fmt.Errorf("transient failure")
	})
	if err == nil {
		t.Fatal("expected error")
	}
	m := tracker.GetMetrics()
	if m.FailedRetries != 1 {
		t.Errorf("FailedRetries=%d, want 1", m.FailedRetries)
	}
}

func TestWithRetryAndTracking_NilTracker(t *testing.T) {
	// nil tracker should not panic
	err := WithRetryAndTracking(context.Background(), fastConfig(3), nil, func(ctx context.Context) error {
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
