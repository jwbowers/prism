//go:build integration

// Package ratelimit provides integration tests for the token bucket rate limiter
//
// These tests verify rate limiting behavior under various scenarios including:
// - Normal operation within rate limits
// - Burst capacity handling
// - Concurrent request handling
// - Recovery after rate limit exhaustion
package ratelimit_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/scttfrdmn/prism/pkg/ratelimit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRateLimitIntegration verifies basic rate limiting functionality
// with the token bucket algorithm (v0.5.12 feature).
//
// Test Coverage:
// - Token refill rate (2 tokens/minute default)
// - Request acceptance within limits
// - Request rejection when limit exceeded
// - Success metrics and error tracking
func TestRateLimitIntegration(t *testing.T) {
	// Test Setup: 2 requests/minute, burst 3
	limiter := ratelimit.NewTokenBucket(2, 3)
	ctx := context.Background()

	// Phase 1: Send 3 requests immediately (should succeed - burst capacity)
	for i := 0; i < 3; i++ {
		err := limiter.Allow(ctx)
		assert.NoError(t, err, "Request %d should succeed (within burst capacity)", i+1)
	}

	// Phase 2: Send 4th request immediately (should fail - over burst)
	err := limiter.Allow(ctx)
	assert.Error(t, err, "4th request should fail (exceeded burst)")

	var rateLimitErr *ratelimit.RateLimitError
	require.ErrorAs(t, err, &rateLimitErr, "Error should be RateLimitError")
	assert.Equal(t, 2.0, rateLimitErr.Rate, "Rate should be 2/min")
	assert.Equal(t, 3, rateLimitErr.Burst, "Burst should be 3")
	assert.Greater(t, rateLimitErr.RetryAfter, time.Duration(0), "RetryAfter should be positive")

	// Phase 3: Wait 30 seconds (refill 1 token: 2 tokens/minute = 1 per 30 sec)
	time.Sleep(31 * time.Second)

	// Phase 4: Send request (should succeed)
	err = limiter.Allow(ctx)
	assert.NoError(t, err, "Request after refill should succeed")

	// Phase 5: Verify metrics
	metrics := limiter.GetMetrics()
	assert.Equal(t, int64(5), metrics.TotalRequests, "Should have 5 total requests")
	assert.Equal(t, int64(4), metrics.AllowedRequests, "Should have 4 allowed requests")
	assert.Equal(t, int64(1), metrics.RateLimited, "Should have 1 rate limited request")
	assert.InDelta(t, 0.8, metrics.SuccessRate, 0.01, "Success rate should be 80%")
}

// TestBurstCapacity validates burst handling for short-term spikes
// in request volume (v0.5.12 feature).
//
// Test Coverage:
// - Initial burst capacity available
// - Burst depletion behavior
// - Gradual refill after burst
// - Multiple burst cycles
func TestBurstCapacity(t *testing.T) {
	// Test Setup: 1 req/minute, burst 5
	limiter := ratelimit.NewTokenBucket(1, 5)
	ctx := context.Background()

	// Phase 1: Send 5 requests rapidly (all should succeed)
	for i := 0; i < 5; i++ {
		err := limiter.Allow(ctx)
		assert.NoError(t, err, "Request %d should succeed (burst capacity)", i+1)
	}

	// Phase 2: Send 6th request (should fail)
	err := limiter.Allow(ctx)
	assert.Error(t, err, "6th request should fail (burst exhausted)")

	// Phase 3: Wait 3 minutes (refill 3 tokens at 1/min)
	time.Sleep(3*time.Minute + 500*time.Millisecond)

	// Phase 4: Send 3 requests (should succeed)
	for i := 0; i < 3; i++ {
		err := limiter.Allow(ctx)
		assert.NoError(t, err, "Request %d after refill should succeed", i+1)
	}

	// Phase 5: Send 4th request (should fail)
	err = limiter.Allow(ctx)
	assert.Error(t, err, "4th request should fail (tokens exhausted)")

	// Verify metrics
	metrics := limiter.GetMetrics()
	assert.Equal(t, int64(10), metrics.TotalRequests, "Should have 10 total requests")
	assert.Equal(t, int64(8), metrics.AllowedRequests, "Should have 8 allowed requests")
	assert.Equal(t, int64(2), metrics.RateLimited, "Should have 2 rate limited requests")
}

// TestConcurrentRequests verifies thread-safety of rate limiter
// under high concurrency (v0.5.12 feature).
//
// Test Coverage:
// - Concurrent goroutine access
// - Mutex protection of token bucket
// - Correct token allocation under contention
// - No race conditions
func TestConcurrentRequests(t *testing.T) {
	// Test Setup: 10 req/minute, burst 10
	limiter := ratelimit.NewTokenBucket(10, 10)
	ctx := context.Background()

	// Launch 50 goroutines concurrently
	const numGoroutines = 50
	var wg sync.WaitGroup
	results := make([]error, numGoroutines)

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			defer wg.Done()
			results[index] = limiter.Allow(ctx)
		}(i)
	}

	wg.Wait()

	// Count successes and failures
	successes := 0
	failures := 0
	for _, err := range results {
		if err == nil {
			successes++
		} else {
			failures++
			// Verify error type
			var rateLimitErr *ratelimit.RateLimitError
			assert.ErrorAs(t, err, &rateLimitErr, "Error should be RateLimitError")
		}
	}

	// Verify exactly 10 succeed (burst capacity) and 40 fail
	assert.Equal(t, 10, successes, "Should have exactly 10 successful requests (burst capacity)")
	assert.Equal(t, 40, failures, "Should have exactly 40 rate limited requests")

	// Verify metrics
	metrics := limiter.GetMetrics()
	assert.Equal(t, int64(50), metrics.TotalRequests, "Should have 50 total requests")
	assert.Equal(t, int64(10), metrics.AllowedRequests, "Should have 10 allowed requests")
	assert.Equal(t, int64(40), metrics.RateLimited, "Should have 40 rate limited requests")
	assert.InDelta(t, 0.2, metrics.SuccessRate, 0.01, "Success rate should be 20%")
}

// TestRateLimitRecovery validates recovery after rate limit exhaustion
// and proper token refill behavior (v0.5.12 feature).
//
// Test Coverage:
// - Token refill timing accuracy
// - Gradual recovery over time
// - Sustained throughput after recovery
// - No token accumulation beyond burst limit
func TestRateLimitRecovery(t *testing.T) {
	// Test Setup: 6 req/minute (1 every 10 seconds), burst 2
	limiter := ratelimit.NewTokenBucket(6, 2)
	ctx := context.Background()

	// Phase 1: Exhaust burst (send 2 requests)
	for i := 0; i < 2; i++ {
		err := limiter.Allow(ctx)
		assert.NoError(t, err, "Initial request %d should succeed", i+1)
	}

	// Phase 2: Verify exhaustion
	err := limiter.Allow(ctx)
	assert.Error(t, err, "Request should fail (burst exhausted)")

	// Phase 3: Wait 10 seconds (should refill 1 token)
	time.Sleep(11 * time.Second)

	// Phase 4: Send 1 request (should succeed)
	err = limiter.Allow(ctx)
	assert.NoError(t, err, "Request after 10s should succeed")

	// Phase 5: Wait 10 seconds (refill 1 token)
	time.Sleep(11 * time.Second)

	// Phase 6: Send 1 request (should succeed)
	err = limiter.Allow(ctx)
	assert.NoError(t, err, "Request after another 10s should succeed")

	// Phase 7: Verify sustained throughput
	// At 6 req/minute, we should be able to send 1 request every 10 seconds
	for i := 0; i < 3; i++ {
		time.Sleep(11 * time.Second)
		err := limiter.Allow(ctx)
		assert.NoError(t, err, "Sustained request %d should succeed", i+1)
	}

	// Verify metrics
	metrics := limiter.GetMetrics()
	assert.GreaterOrEqual(t, metrics.AllowedRequests, int64(7), "Should have at least 7 allowed requests")
	assert.LessOrEqual(t, metrics.CurrentTokens, float64(2), "Tokens should not exceed burst capacity")
}

// TestRateLimitMetrics validates metrics collection and reporting
// for rate limiting operations (v0.5.12 feature).
//
// Test Coverage:
// - Success rate tracking
// - Rate limit error counting
// - Burst utilization metrics
// - Time-windowed statistics
func TestRateLimitMetrics(t *testing.T) {
	// Test Setup: 60 req/minute (1 per second), burst 10
	limiter := ratelimit.NewTokenBucket(60, 10)
	ctx := context.Background()

	startTime := time.Now()

	// Phase 1: Send 10 requests rapidly (burst)
	for i := 0; i < 10; i++ {
		err := limiter.Allow(ctx)
		assert.NoError(t, err, "Burst request %d should succeed", i+1)
	}

	// Phase 2: Send 5 more requests (should fail)
	for i := 0; i < 5; i++ {
		err := limiter.Allow(ctx)
		assert.Error(t, err, "Request %d should fail (burst exhausted)", i+11)
	}

	// Phase 3: Wait for refill and send sustained requests
	for i := 0; i < 5; i++ {
		time.Sleep(1100 * time.Millisecond) // 1 token per second
		err := limiter.Allow(ctx)
		assert.NoError(t, err, "Sustained request %d should succeed", i+1)
	}

	// Phase 4: Verify metrics
	metrics := limiter.GetMetrics()

	// Verify totals
	assert.Equal(t, int64(20), metrics.TotalRequests, "Should have 20 total requests")
	assert.Equal(t, int64(15), metrics.AllowedRequests, "Should have 15 allowed requests")
	assert.Equal(t, int64(5), metrics.RateLimited, "Should have 5 rate limited requests")

	// Verify success rate
	expectedSuccessRate := 15.0 / 20.0
	assert.InDelta(t, expectedSuccessRate, metrics.SuccessRate, 0.01, "Success rate should be 75%")

	// Verify configuration
	assert.Equal(t, 60.0, metrics.Rate, "Rate should be 60/min")
	assert.Equal(t, 10, metrics.Burst, "Burst should be 10")

	// Verify uptime
	elapsedTime := time.Since(startTime)
	assert.GreaterOrEqual(t, metrics.Uptime, elapsedTime-time.Second, "Uptime should be approximately elapsed time")

	// Verify current tokens (should be recovering)
	assert.GreaterOrEqual(t, metrics.CurrentTokens, 0.0, "Current tokens should be non-negative")
	assert.LessOrEqual(t, metrics.CurrentTokens, float64(metrics.Burst), "Current tokens should not exceed burst")

	// Test metrics reset
	limiter.ResetMetrics()
	resetMetrics := limiter.GetMetrics()
	assert.Equal(t, int64(0), resetMetrics.TotalRequests, "Total requests should reset to 0")
	assert.Equal(t, int64(0), resetMetrics.AllowedRequests, "Allowed requests should reset to 0")
	assert.Equal(t, int64(0), resetMetrics.RateLimited, "Rate limited should reset to 0")
	assert.Equal(t, 0.0, resetMetrics.SuccessRate, "Success rate should reset to 0")

	// Verify String() method produces readable output
	metricsStr := metrics.String()
	assert.Contains(t, metricsStr, "Rate Limiter Metrics", "Metrics string should contain header")
	assert.Contains(t, metricsStr, "60 req/min", "Metrics string should contain rate")
	assert.Contains(t, metricsStr, "burst 10", "Metrics string should contain burst")
}
