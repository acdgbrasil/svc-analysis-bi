package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/acdgbrasil/svc-analysis-bi/internal/api/middleware"
	"github.com/acdgbrasil/svc-analysis-bi/internal/store"
)

// stubDB implements HealthChecker for tests.
type stubDB struct {
	err error
}

func (s *stubDB) Ping(_ context.Context) error { return s.err }

// stubNATS implements NATSChecker for tests.
type stubNATS struct {
	connected bool
}

func (s *stubNATS) IsConnected() bool { return s.connected }

// stubJWT implements middleware.JWTValidator for tests.
type stubJWT struct {
	claims *middleware.Claims
	err    error
}

func (s *stubJWT) Validate(_ context.Context, _ string) (*middleware.Claims, error) {
	return s.claims, s.err
}

// stubIndicators implements handlers.IndicatorQuerier for tests.
type stubIndicators struct {
	result *store.IndicatorResult
	err    error
}

func (s *stubIndicators) QueryDemographics(_ context.Context, _ store.IndicatorParams) (*store.IndicatorResult, error) {
	return s.result, s.err
}
func (s *stubIndicators) QueryEpidemiological(_ context.Context, _ store.IndicatorParams) (*store.IndicatorResult, error) {
	return s.result, s.err
}
func (s *stubIndicators) QuerySocioeconomic(_ context.Context, _ store.IndicatorParams) (*store.IndicatorResult, error) {
	return s.result, s.err
}
func (s *stubIndicators) QueryProtection(_ context.Context, _ store.IndicatorParams) (*store.IndicatorResult, error) {
	return s.result, s.err
}
func (s *stubIndicators) QueryCare(_ context.Context, _ store.IndicatorParams) (*store.IndicatorResult, error) {
	return s.result, s.err
}

func newTestRouter(t *testing.T) http.Handler {
	t.Helper()
	return NewRouter(RouterDeps{
		DB:   &stubDB{},
		NATS: &stubNATS{connected: true},
		JWTValidator: &stubJWT{
			claims: &middleware.Claims{Subject: "test-user", Roles: []string{"analyst"}},
		},
		Indicators: &stubIndicators{
			result: &store.IndicatorResult{Rows: []store.IndicatorRow{}, Suppressed: 0},
		},
	})
}

func newTestRouterNoIndicators(t *testing.T) http.Handler {
	t.Helper()
	return NewRouter(RouterDeps{
		DB:   &stubDB{},
		NATS: &stubNATS{connected: true},
		JWTValidator: &stubJWT{
			claims: &middleware.Claims{Subject: "test-user", Roles: []string{"analyst"}},
		},
	})
}

func TestRouter_Health(t *testing.T) {
	router := newTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Verify security headers are set
	if w.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Error("expected X-Content-Type-Options: nosniff")
	}
}

func TestRouter_Ready(t *testing.T) {
	t.Run("healthy", func(t *testing.T) {
		router := newTestRouter(t)
		req := httptest.NewRequest(http.MethodGet, "/ready", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})

	t.Run("unhealthy DB", func(t *testing.T) {
		router := NewRouter(RouterDeps{
			DB:   &stubDB{err: errors.New("db down")},
			NATS: &stubNATS{connected: true},
		})
		req := httptest.NewRequest(http.MethodGet, "/ready", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusServiceUnavailable {
			t.Errorf("expected 503, got %d", w.Code)
		}
	})
}

func TestRouter_Indicators_WithStore(t *testing.T) {
	router := newTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/indicators/demographics?period_start=2025-01&period_end=2025-06", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}

	var resp Response
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
}

func TestRouter_Indicators_Placeholder(t *testing.T) {
	router := newTestRouterNoIndicators(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/indicators/demographics", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotImplemented {
		t.Errorf("expected 501, got %d", w.Code)
	}

	var resp Response
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
}

func TestRouter_Export_Placeholder(t *testing.T) {
	router := newTestRouterNoIndicators(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/export/csv", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotImplemented {
		t.Errorf("expected 501, got %d", w.Code)
	}
}

func TestRouter_Metadata_Datasets(t *testing.T) {
	router := newTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/metadata/datasets", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}
}

func TestRouter_JWT_RequiredForAPI(t *testing.T) {
	router := NewRouter(RouterDeps{
		DB:   &stubDB{},
		NATS: &stubNATS{connected: true},
		JWTValidator: &stubJWT{
			err: errors.New("invalid"),
		},
	})

	// API endpoint without auth should return 401
	req := httptest.NewRequest(http.MethodGet, "/api/v1/indicators/demographics", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestRouter_SecurityHeaders(t *testing.T) {
	router := newTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	headers := []struct {
		name string
		want string
	}{
		{"Strict-Transport-Security", "max-age=31536000; includeSubDomains"},
		{"X-Content-Type-Options", "nosniff"},
		{"X-Frame-Options", "DENY"},
		{"Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'"},
		{"Cache-Control", "no-store"},
	}

	for _, h := range headers {
		t.Run(h.name, func(t *testing.T) {
			got := w.Header().Get(h.name)
			if got != h.want {
				t.Errorf("%s = %q, want %q", h.name, got, h.want)
			}
		})
	}
}

func TestRouter_MethodNotAllowed(t *testing.T) {
	router := newTestRouter(t)

	// POST to /health should not match GET /health
	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// chi returns 405 for wrong method on a registered route
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}
