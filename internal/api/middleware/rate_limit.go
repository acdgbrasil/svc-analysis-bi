package middleware

import (
	"net/http"
	"sync"
	"time"
)

// RateLimitConfig configures the token bucket rate limiter.
type RateLimitConfig struct {
	// RequestsPerSecond is the sustained rate of allowed requests.
	RequestsPerSecond float64

	// BurstSize is the maximum number of requests allowed in a single burst.
	BurstSize int
}

// tokenBucket implements a simple token bucket algorithm.
// It is safe for concurrent use via sync.Mutex.
type tokenBucket struct {
	mu            sync.Mutex
	tokens        float64
	maxTokens     float64
	refillRate    float64 // tokens per nanosecond
	lastRefill    time.Time
	timeNow       func() time.Time // injectable clock for testing
}

func newTokenBucket(rps float64, burst int, nowFn func() time.Time) *tokenBucket {
	if nowFn == nil {
		nowFn = time.Now
	}
	return &tokenBucket{
		tokens:     float64(burst),
		maxTokens:  float64(burst),
		refillRate: rps / float64(time.Second), // tokens per nanosecond
		lastRefill: nowFn(),
		timeNow:    nowFn,
	}
}

// allow attempts to consume one token. Returns true if the request
// is allowed, false if rate limited.
func (tb *tokenBucket) allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := tb.timeNow()
	elapsed := now.Sub(tb.lastRefill)
	tb.lastRefill = now

	// Refill tokens based on elapsed time
	tb.tokens += float64(elapsed) * tb.refillRate
	if tb.tokens > tb.maxTokens {
		tb.tokens = tb.maxTokens
	}

	if tb.tokens < 1.0 {
		return false
	}

	tb.tokens--
	return true
}

// RateLimit returns middleware that limits requests using a token bucket
// algorithm. When the limit is exceeded, it returns 429 Too Many Requests.
func RateLimit(cfg RateLimitConfig) func(http.Handler) http.Handler {
	return RateLimitWithClock(cfg, nil)
}

// RateLimitWithClock is like RateLimit but accepts an injectable clock
// function for deterministic testing.
func RateLimitWithClock(cfg RateLimitConfig, nowFn func() time.Time) func(http.Handler) http.Handler {
	bucket := newTokenBucket(cfg.RequestsPerSecond, cfg.BurstSize, nowFn)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !bucket.allow() {
				writeJSONError(w, http.StatusTooManyRequests, "rate limit exceeded")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
