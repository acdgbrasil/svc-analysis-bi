package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// stubValidator is a test double for JWTValidator.
type stubValidator struct {
	claims *Claims
	err    error
}

func (s *stubValidator) Validate(_ context.Context, _ string) (*Claims, error) {
	return s.claims, s.err
}

func TestJWTAuth(t *testing.T) {
	validClaims := &Claims{Subject: "user-123", Roles: []string{"analyst"}}

	t.Run("skips authentication for exempt paths", func(t *testing.T) {
		skip := map[string]bool{"/health": true}
		handler := JWTAuth(&stubValidator{err: errors.New("should not be called")}, skip)(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}),
		)

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})

	t.Run("returns 401 for missing authorization header", func(t *testing.T) {
		handler := JWTAuth(&stubValidator{claims: validClaims}, nil)(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				t.Error("handler should not be called")
			}),
		)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/data", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", w.Code)
		}
	})

	t.Run("returns 401 for non-Bearer scheme", func(t *testing.T) {
		handler := JWTAuth(&stubValidator{claims: validClaims}, nil)(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				t.Error("handler should not be called")
			}),
		)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/data", nil)
		req.Header.Set("Authorization", "Basic dXNlcjpwYXNz")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", w.Code)
		}
	})

	t.Run("returns 401 for invalid token", func(t *testing.T) {
		handler := JWTAuth(&stubValidator{err: errors.New("invalid")}, nil)(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				t.Error("handler should not be called")
			}),
		)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/data", nil)
		req.Header.Set("Authorization", "Bearer bad-token")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", w.Code)
		}
	})

	t.Run("sets claims in context for valid token", func(t *testing.T) {
		var gotClaims *Claims
		handler := JWTAuth(&stubValidator{claims: validClaims}, nil)(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotClaims = ClaimsFromContext(r.Context())
				w.WriteHeader(http.StatusOK)
			}),
		)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/data", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
		if gotClaims == nil {
			t.Fatal("expected claims in context, got nil")
		}
		if gotClaims.Subject != "user-123" {
			t.Errorf("expected subject 'user-123', got %q", gotClaims.Subject)
		}
	})
}

func TestClaimsFromContext_NoClaims(t *testing.T) {
	claims := ClaimsFromContext(context.Background())
	if claims != nil {
		t.Error("expected nil claims for empty context")
	}
}
