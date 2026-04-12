package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// stubDB is a test double for HealthChecker.
type stubDB struct {
	err error
}

func (s *stubDB) Ping(_ context.Context) error { return s.err }

// stubNATS is a test double for NATSChecker.
type stubNATS struct {
	connected bool
}

func (s *stubNATS) IsConnected() bool { return s.connected }

func TestHealthHandler(t *testing.T) {
	handler := HealthHandler()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp Response
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}

	data, ok := resp.Data.(map[string]any)
	if !ok {
		t.Fatal("expected data to be a map")
	}
	if data["status"] != "ok" {
		t.Errorf("expected status 'ok', got %v", data["status"])
	}
}

func TestReadyHandler(t *testing.T) {
	t.Run("returns 200 when all healthy", func(t *testing.T) {
		handler := ReadyHandler(&stubDB{}, &stubNATS{connected: true})
		req := httptest.NewRequest(http.MethodGet, "/ready", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}

		var resp Response
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode: %v", err)
		}
		data := resp.Data.(map[string]any)
		if data["status"] != "ready" {
			t.Errorf("expected status 'ready', got %v", data["status"])
		}
	})

	t.Run("returns 503 when DB unhealthy", func(t *testing.T) {
		handler := ReadyHandler(&stubDB{err: errors.New("db down")}, &stubNATS{connected: true})
		req := httptest.NewRequest(http.MethodGet, "/ready", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusServiceUnavailable {
			t.Errorf("expected 503, got %d", w.Code)
		}
	})

	t.Run("returns 503 when NATS unhealthy", func(t *testing.T) {
		handler := ReadyHandler(&stubDB{}, &stubNATS{connected: false})
		req := httptest.NewRequest(http.MethodGet, "/ready", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusServiceUnavailable {
			t.Errorf("expected 503, got %d", w.Code)
		}
	})

	t.Run("returns 200 with nil dependencies", func(t *testing.T) {
		handler := ReadyHandler(nil, nil)
		req := httptest.NewRequest(http.MethodGet, "/ready", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
}
