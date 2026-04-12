package middleware

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestRateLimit(t *testing.T) {
	t.Run("allows requests within burst", func(t *testing.T) {
		cfg := RateLimitConfig{RequestsPerSecond: 10, BurstSize: 5}
		now := time.Now()
		mw := RateLimitWithClock(cfg, func() time.Time { return now })

		handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		for i := 0; i < 5; i++ {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				t.Errorf("request %d: expected 200, got %d", i, w.Code)
			}
		}
	})

	t.Run("rejects when burst exceeded", func(t *testing.T) {
		cfg := RateLimitConfig{RequestsPerSecond: 10, BurstSize: 2}
		now := time.Now()
		mw := RateLimitWithClock(cfg, func() time.Time { return now })

		handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		// Exhaust burst
		for i := 0; i < 2; i++ {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
		}

		// Next request should be rejected
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusTooManyRequests {
			t.Errorf("expected 429, got %d", w.Code)
		}
	})

	t.Run("refills tokens over time", func(t *testing.T) {
		cfg := RateLimitConfig{RequestsPerSecond: 10, BurstSize: 1}
		currentTime := time.Now()
		mu := sync.Mutex{}
		mw := RateLimitWithClock(cfg, func() time.Time {
			mu.Lock()
			defer mu.Unlock()
			return currentTime
		})

		handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		// Use the single token
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("first request: expected 200, got %d", w.Code)
		}

		// Advance time by 200ms (should refill ~2 tokens at 10/s)
		mu.Lock()
		currentTime = currentTime.Add(200 * time.Millisecond)
		mu.Unlock()

		req = httptest.NewRequest(http.MethodGet, "/test", nil)
		w = httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("after refill: expected 200, got %d", w.Code)
		}
	})

	t.Run("concurrent access is safe", func(t *testing.T) {
		cfg := RateLimitConfig{RequestsPerSecond: 1000, BurstSize: 100}
		mw := RateLimit(cfg)

		handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		var wg sync.WaitGroup
		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				w := httptest.NewRecorder()
				handler.ServeHTTP(w, req)
				// Just verify no panics
			}()
		}
		wg.Wait()
	})
}
