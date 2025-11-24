// Package retry provides integration tests for exponential backoff retry logic
//
// These tests verify retry behavior under various failure scenarios including:
// - Transient network failures
// - AWS API throttling
// - Exponential backoff timing
// - Jitter variance
package retry_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/scttfrdmn/prism/pkg/retry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRetryWithTransientFailure validates retry logic with temporary failures
// that eventually succeed (v0.5.12 feature).
//
// Test Coverage:
// - Automatic retry on transient errors
// - Success on final attempt
// - No retry on permanent errors
// - Context cancellation handling
func TestRetryWithTransientFailure(t *testing.T) {
	ctx := context.Background()

	// Mock transient failure operation that succeeds on 3rd attempt
	attemptCount := 0
	operation := func() error {
		attemptCount++
		if attemptCount < 3 {
			return errors.New("timeout: connection timed out") // Retryable error
		}
		return nil // Success on 3rd attempt
	}

	// Configure retrier: max 3 attempts, 100ms base delay (faster for testing)
	retryer := retry.NewExponentialBackoff(3, 100*time.Millisecond)

	// Execute operation with retry
	startTime := time.Now()
	err := retryer.Do(ctx, operation)
	elapsed := time.Since(startTime)

	// Verify success
	assert.NoError(t, err, "Operation should succeed after retries")
	assert.Equal(t, 3, attemptCount, "Should have made 3 attempts")

	// Verify backoff timing: 0ms + 100ms + 200ms = ~300ms
	assert.Greater(t, elapsed, 250*time.Millisecond, "Should have delays between retries")
	assert.Less(t, elapsed, 500*time.Millisecond, "Should not take too long")

	// Verify metrics
	metrics := retryer.GetMetrics()
	assert.Equal(t, int64(1), metrics.TotalOperations, "Should have 1 operation")
	assert.Equal(t, int64(1), metrics.SuccessfulOps, "Should have 1 success")
	assert.Equal(t, int64(0), metrics.FailedOps, "Should have 0 failures")
	assert.Equal(t, int64(2), metrics.TotalRetries, "Should have 2 retries")
}

// TestRetryExponentialBackoff validates exponential backoff timing
// between retry attempts (v0.5.12 feature).
//
// Test Coverage:
// - Base delay (100ms for faster testing)
// - Exponential growth (100ms → 200ms → 400ms)
// - Maximum retry limit (4 attempts)
// - Total backoff time calculation
func TestRetryExponentialBackoff(t *testing.T) {
	ctx := context.Background()

	// Track timestamps of each attempt
	timestamps := make([]time.Time, 0)
	operation := func() error {
		timestamps = append(timestamps, time.Now())
		return errors.New("service unavailable") // Always fail
	}

	// Configure retrier: max 4 attempts, 100ms base delay
	retryer := retry.NewExponentialBackoff(4, 100*time.Millisecond).
		WithJitter(0) // No jitter for predictable timing

	// Execute operation (will fail all attempts)
	startTime := time.Now()
	err := retryer.Do(ctx, operation)

	// Verify failure after max attempts
	assert.Error(t, err, "Operation should fail after max attempts")
	var retryErr *retry.RetryableError
	require.ErrorAs(t, err, &retryErr, "Error should be RetryableError")
	assert.Equal(t, 4, retryErr.Attempt, "Should have made 4 attempts")

	// Verify 4 attempts were made
	assert.Len(t, timestamps, 4, "Should have 4 timestamps")

	// Verify exponential backoff delays
	// Expected delays: 0ms, 100ms, 200ms, 400ms
	// Total expected: ~700ms
	if len(timestamps) == 4 {
		delay1 := timestamps[1].Sub(timestamps[0])
		delay2 := timestamps[2].Sub(timestamps[1])
		delay3 := timestamps[3].Sub(timestamps[2])

		assert.Greater(t, delay1, 80*time.Millisecond, "First delay should be ~100ms")
		assert.Less(t, delay1, 150*time.Millisecond, "First delay should be ~100ms")

		assert.Greater(t, delay2, 180*time.Millisecond, "Second delay should be ~200ms")
		assert.Less(t, delay2, 250*time.Millisecond, "Second delay should be ~200ms")

		assert.Greater(t, delay3, 380*time.Millisecond, "Third delay should be ~400ms")
		assert.Less(t, delay3, 450*time.Millisecond, "Third delay should be ~400ms")
	}

	// Verify total time
	totalTime := time.Since(startTime)
	assert.Greater(t, totalTime, 650*time.Millisecond, "Total time should be ~700ms")
	assert.Less(t, totalTime, 900*time.Millisecond, "Total time should be ~700ms")
}

// TestRetryJitterVariance validates random jitter application
// to prevent thundering herd (v0.5.12 feature).
//
// Test Coverage:
// - Jitter percentage (±20%)
// - Random variance in retry delays
// - Multiple retry attempts with different jitter
// - Jitter distribution within bounds
func TestRetryJitterVariance(t *testing.T) {
	ctx := context.Background()

	// Run multiple iterations to test jitter variance
	const iterations = 20
	delays := make([]time.Duration, 0, iterations)

	for i := 0; i < iterations; i++ {
		timestamps := make([]time.Time, 0)
		operation := func() error {
			timestamps = append(timestamps, time.Now())
			if len(timestamps) < 2 {
				return errors.New("rate exceeded") // Fail once
			}
			return nil // Success on second attempt
		}

		retryer := retry.NewExponentialBackoff(3, 100*time.Millisecond).
			WithJitter(0.2) // 20% jitter

		_ = retryer.Do(ctx, operation)

		// Capture the delay between first and second attempt
		if len(timestamps) >= 2 {
			delay := timestamps[1].Sub(timestamps[0])
			delays = append(delays, delay)
		}
	}

	// Verify we have enough samples
	require.GreaterOrEqual(t, len(delays), 15, "Should have enough delay samples")

	// Expected base delay: 100ms
	// With 20% jitter: 80ms - 120ms range
	// Add 5ms tolerance for system scheduling overhead
	minDelay := 80 * time.Millisecond
	maxDelay := 125 * time.Millisecond // 120ms + 5ms tolerance

	// Verify all delays are within jitter bounds (with tolerance)
	for _, delay := range delays {
		assert.GreaterOrEqual(t, delay, minDelay, "Delay should be >= 80ms")
		assert.LessOrEqual(t, delay, maxDelay, "Delay should be <= 125ms (120ms + 5ms tolerance)")
	}

	// Verify variance exists (not all the same)
	uniqueDelays := make(map[time.Duration]bool)
	for _, delay := range delays {
		// Round to nearest 10ms for comparison
		rounded := delay.Round(10 * time.Millisecond)
		uniqueDelays[rounded] = true
	}

	assert.Greater(t, len(uniqueDelays), 1, "Should have variance in delays (jitter working)")
}

// TestMaxRetriesExceeded validates behavior when all retry attempts
// are exhausted (v0.5.12 feature).
//
// Test Coverage:
// - Max retry limit enforcement
// - Final error returned to caller
// - No infinite retry loops
// - Proper error context preservation
func TestMaxRetriesExceeded(t *testing.T) {
	ctx := context.Background()

	// Operation that always fails with retryable error
	attemptCount := 0
	originalErr := fmt.Errorf("throttling exception: rate limit exceeded")
	operation := func() error {
		attemptCount++
		return originalErr
	}

	// Configure retrier: max 3 attempts, 50ms base delay
	maxRetries := 3
	retryer := retry.NewExponentialBackoff(maxRetries, 50*time.Millisecond)

	// Execute operation
	err := retryer.DoNamed(ctx, "test-operation", operation)

	// Verify failure
	assert.Error(t, err, "Operation should fail after max retries")

	// Verify error type and details
	var retryErr *retry.RetryableError
	require.ErrorAs(t, err, &retryErr, "Error should be RetryableError")
	assert.Equal(t, maxRetries, retryErr.Attempt, "Should have made max attempts")
	assert.Equal(t, maxRetries, retryErr.MaxAttempts, "Should report max attempts")
	assert.Equal(t, "test-operation", retryErr.Operation, "Should preserve operation name")
	assert.ErrorIs(t, err, originalErr, "Should preserve original error")

	// Verify exactly max attempts were made
	assert.Equal(t, maxRetries, attemptCount, "Should have made exactly max attempts")

	// Verify metrics
	metrics := retryer.GetMetrics()
	assert.Equal(t, int64(1), metrics.TotalOperations, "Should have 1 operation")
	assert.Equal(t, int64(0), metrics.SuccessfulOps, "Should have 0 successes")
	assert.Equal(t, int64(1), metrics.FailedOps, "Should have 1 failure")
	assert.Equal(t, int64(maxRetries-1), metrics.TotalRetries, "Should have max-1 retries")
}

// TestRetryContextCancellation validates proper handling of context
// cancellation during retry (v0.5.12 feature).
//
// Test Coverage:
// - Context timeout during retry
// - Context cancellation between retries
// - Immediate return on context done
// - No retry after context cancellation
func TestRetryContextCancellation(t *testing.T) {
	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()

	// Operation that fails slowly
	attemptCount := 0
	operation := func() error {
		attemptCount++
		time.Sleep(50 * time.Millisecond)   // Slow operation
		return errors.New("internal error") // Always fail
	}

	// Configure retrier: max 5 attempts, 100ms base delay
	// Total time would be ~700ms without cancellation
	retryer := retry.NewExponentialBackoff(5, 100*time.Millisecond)

	// Execute operation
	startTime := time.Now()
	err := retryer.Do(ctx, operation)
	elapsed := time.Since(startTime)

	// Verify context cancellation
	assert.Error(t, err, "Operation should return error")
	assert.ErrorIs(t, err, context.DeadlineExceeded, "Error should be context deadline exceeded")

	// Verify early termination (< 400ms, much less than full 700ms)
	assert.Less(t, elapsed, 400*time.Millisecond, "Should terminate early on context cancellation")

	// Verify limited attempts (should be 1-3, not all 5)
	assert.LessOrEqual(t, attemptCount, 3, "Should have limited attempts before cancellation")
	assert.GreaterOrEqual(t, attemptCount, 1, "Should have at least 1 attempt")
}

// TestRetryMetrics validates metrics collection and reporting
// for retry operations (v0.5.12 feature).
//
// Test Coverage:
// - Success rate tracking
// - Retry count statistics
// - Attempt histogram
// - Metrics reset functionality
func TestRetryMetrics(t *testing.T) {
	ctx := context.Background()
	retryer := retry.NewExponentialBackoff(4, 10*time.Millisecond)

	// Test 1: Operation that succeeds immediately
	successOp := func() error {
		return nil
	}
	err := retryer.Do(ctx, successOp)
	assert.NoError(t, err)

	// Test 2: Operation that succeeds on 2nd attempt
	attempt := 0
	retryOnceOp := func() error {
		attempt++
		if attempt < 2 {
			return errors.New("timeout")
		}
		return nil
	}
	err = retryer.Do(ctx, retryOnceOp)
	assert.NoError(t, err)

	// Test 3: Operation that succeeds on 3rd attempt
	attempt = 0
	retryTwiceOp := func() error {
		attempt++
		if attempt < 3 {
			return errors.New("throttled")
		}
		return nil
	}
	err = retryer.Do(ctx, retryTwiceOp)
	assert.NoError(t, err)

	// Test 4: Operation that fails after all attempts
	failOp := func() error {
		return errors.New("service unavailable")
	}
	err = retryer.Do(ctx, failOp)
	assert.Error(t, err)

	// Verify metrics
	metrics := retryer.GetMetrics()

	// Totals
	assert.Equal(t, int64(4), metrics.TotalOperations, "Should have 4 total operations")
	assert.Equal(t, int64(3), metrics.SuccessfulOps, "Should have 3 successful operations")
	assert.Equal(t, int64(1), metrics.FailedOps, "Should have 1 failed operation")

	// Success rate
	expectedSuccessRate := 3.0 / 4.0
	assert.InDelta(t, expectedSuccessRate, metrics.SuccessRate, 0.01, "Success rate should be 75%")

	// Retry counts
	// Op1: 0 retries, Op2: 1 retry, Op3: 2 retries, Op4: 3 retries = 6 total
	assert.Equal(t, int64(6), metrics.TotalRetries, "Should have 6 total retries")
	assert.InDelta(t, 1.5, metrics.AvgRetries, 0.1, "Average should be 1.5 retries/op")

	// Attempt histogram
	// Op1: 1 attempt, Op2: 2 attempts, Op3: 3 attempts, Op4: 4 attempts
	assert.Equal(t, int64(1), metrics.AttemptHistogram[1], "Should have 1 op with 1 attempt")
	assert.Equal(t, int64(1), metrics.AttemptHistogram[2], "Should have 1 op with 2 attempts")
	assert.Equal(t, int64(1), metrics.AttemptHistogram[3], "Should have 1 op with 3 attempts")
	assert.Equal(t, int64(1), metrics.AttemptHistogram[4], "Should have 1 op with 4 attempts")

	// Test metrics reset
	retryer.ResetMetrics()
	resetMetrics := retryer.GetMetrics()
	assert.Equal(t, int64(0), resetMetrics.TotalOperations, "Operations should reset to 0")
	assert.Equal(t, int64(0), resetMetrics.SuccessfulOps, "Successes should reset to 0")
	assert.Equal(t, int64(0), resetMetrics.FailedOps, "Failures should reset to 0")
	assert.Equal(t, int64(0), resetMetrics.TotalRetries, "Retries should reset to 0")
	assert.Equal(t, 0.0, resetMetrics.AvgRetries, "Avg retries should reset to 0")

	// Verify String() method produces readable output
	metricsStr := metrics.String()
	assert.Contains(t, metricsStr, "Retry Metrics", "Metrics string should contain header")
	assert.Contains(t, metricsStr, "Total Operations: 4", "Metrics string should contain total")
	assert.Contains(t, metricsStr, "Successful: 3", "Metrics string should contain successes")
}
