package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/acdgbrasil/svc-analysis-bi/internal/store"
)

// mockIndicators implements IndicatorQuerier for handler tests.
type mockIndicators struct {
	result *store.IndicatorResult
	err    error
	called string
}

func (m *mockIndicators) QueryDemographics(_ context.Context, _ store.IndicatorParams) (*store.IndicatorResult, error) {
	m.called = "demographics"
	return m.result, m.err
}
func (m *mockIndicators) QueryEpidemiological(_ context.Context, _ store.IndicatorParams) (*store.IndicatorResult, error) {
	m.called = "epidemiological"
	return m.result, m.err
}
func (m *mockIndicators) QuerySocioeconomic(_ context.Context, _ store.IndicatorParams) (*store.IndicatorResult, error) {
	m.called = "socioeconomic"
	return m.result, m.err
}
func (m *mockIndicators) QueryProtection(_ context.Context, _ store.IndicatorParams) (*store.IndicatorResult, error) {
	m.called = "protection"
	return m.result, m.err
}
func (m *mockIndicators) QueryCare(_ context.Context, _ store.IndicatorParams) (*store.IndicatorResult, error) {
	m.called = "care"
	return m.result, m.err
}

// chiContext creates a request with chi URL param context.
func chiContext(r *http.Request, key, val string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, val)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func TestIndicatorsHandler_ValidAxis(t *testing.T) {
	axes := []string{"demographics", "epidemiological", "socioeconomic", "protection", "care"}

	for _, axis := range axes {
		t.Run(axis, func(t *testing.T) {
			mock := &mockIndicators{
				result: &store.IndicatorResult{
					Rows: []store.IndicatorRow{
						{Labels: map[string]string{"test": "label"}, Value: 10, Period: "2025-01"},
					},
					Suppressed: 2,
				},
			}

			handler := IndicatorsHandler(mock)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/indicators/"+axis+"?period_start=2025-01&period_end=2025-06", nil)
			req = chiContext(req, "axis", axis)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("expected 200, got %d; body: %s", w.Code, w.Body.String())
			}
			if mock.called != axis {
				t.Errorf("expected %s to be called, got %s", axis, mock.called)
			}

			var resp Response
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("decode error: %v", err)
			}
			if resp.Meta.SuppressedGroups != 2 {
				t.Errorf("expected suppressed_groups=2, got %d", resp.Meta.SuppressedGroups)
			}
			if resp.Meta.KThreshold != 5 {
				t.Errorf("expected k_threshold=5, got %d", resp.Meta.KThreshold)
			}
		})
	}
}

func TestIndicatorsHandler_UnknownAxis(t *testing.T) {
	mock := &mockIndicators{
		result: &store.IndicatorResult{},
	}
	handler := IndicatorsHandler(mock)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/indicators/unknown?period_start=2025-01&period_end=2025-06", nil)
	req = chiContext(req, "axis", "unknown")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestIndicatorsHandler_MissingPeriod(t *testing.T) {
	mock := &mockIndicators{
		result: &store.IndicatorResult{},
	}
	handler := IndicatorsHandler(mock)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/indicators/demographics", nil)
	req = chiContext(req, "axis", "demographics")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestIndicatorsHandler_InvalidPeriodFormat(t *testing.T) {
	mock := &mockIndicators{
		result: &store.IndicatorResult{},
	}
	handler := IndicatorsHandler(mock)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/indicators/demographics?period_start=invalid&period_end=2025-06", nil)
	req = chiContext(req, "axis", "demographics")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestIndicatorsHandler_StoreError(t *testing.T) {
	mock := &mockIndicators{
		err: errors.New("database error"),
	}
	handler := IndicatorsHandler(mock)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/indicators/demographics?period_start=2025-01&period_end=2025-06", nil)
	req = chiContext(req, "axis", "demographics")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestIndicatorsHandler_TopParam(t *testing.T) {
	mock := &mockIndicators{
		result: &store.IndicatorResult{Rows: []store.IndicatorRow{}},
	}
	handler := IndicatorsHandler(mock)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/indicators/epidemiological?period_start=2025-01&period_end=2025-06&top=10", nil)
	req = chiContext(req, "axis", "epidemiological")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}
}

func TestIndicatorsHandler_InvalidTop(t *testing.T) {
	mock := &mockIndicators{
		result: &store.IndicatorResult{},
	}
	handler := IndicatorsHandler(mock)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/indicators/demographics?period_start=2025-01&period_end=2025-06&top=abc", nil)
	req = chiContext(req, "axis", "demographics")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestIndicatorsHandler_GranularityParam(t *testing.T) {
	for _, g := range []string{"monthly", "quarterly", "yearly"} {
		t.Run(g, func(t *testing.T) {
			mock := &mockIndicators{
				result: &store.IndicatorResult{Rows: []store.IndicatorRow{}},
			}
			handler := IndicatorsHandler(mock)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/indicators/demographics?period_start=2025-01&period_end=2025-12&granularity="+g, nil)
			req = chiContext(req, "axis", "demographics")
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("expected 200 for granularity=%s, got %d", g, w.Code)
			}
		})
	}
}
