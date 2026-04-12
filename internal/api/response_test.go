package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteJSON(t *testing.T) {
	t.Run("writes correct status and content type", func(t *testing.T) {
		w := httptest.NewRecorder()
		resp := Response{
			Data: map[string]string{"key": "value"},
			Meta: NewResponseMeta(),
		}

		WriteJSON(w, http.StatusOK, resp)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}
		ct := w.Header().Get("Content-Type")
		if ct != "application/json; charset=utf-8" {
			t.Errorf("unexpected content type: %s", ct)
		}

		var decoded Response
		if err := json.NewDecoder(w.Body).Decode(&decoded); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
	})

	t.Run("writes 201 status", func(t *testing.T) {
		w := httptest.NewRecorder()
		resp := Response{
			Data: nil,
			Meta: NewResponseMeta(),
		}
		WriteJSON(w, http.StatusCreated, resp)
		if w.Code != http.StatusCreated {
			t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
		}
	})
}

func TestWriteError(t *testing.T) {
	t.Run("writes error with correct status", func(t *testing.T) {
		w := httptest.NewRecorder()
		WriteError(w, http.StatusBadRequest, "invalid input")

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
		}

		var decoded Response
		if err := json.NewDecoder(w.Body).Decode(&decoded); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		// Check that the data contains the error info
		dataMap, ok := decoded.Data.(map[string]any)
		if !ok {
			t.Fatal("expected data to be a map")
		}
		if msg, ok := dataMap["message"].(string); !ok || msg != "invalid input" {
			t.Errorf("expected message 'invalid input', got %v", dataMap["message"])
		}
		if status, ok := dataMap["status"].(float64); !ok || int(status) != http.StatusBadRequest {
			t.Errorf("expected status %d, got %v", http.StatusBadRequest, dataMap["status"])
		}
	})
}

func TestNewResponseMeta(t *testing.T) {
	meta := NewResponseMeta()
	if meta.Timestamp == "" {
		t.Error("expected non-empty timestamp")
	}
	if meta.KThreshold != 5 {
		t.Errorf("expected k_threshold=5, got %d", meta.KThreshold)
	}
}
