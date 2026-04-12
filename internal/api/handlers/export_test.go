package handlers

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/acdgbrasil/svc-analysis-bi/internal/export"
	"github.com/acdgbrasil/svc-analysis-bi/internal/store"
)

// stubEncoder implements export.Encoder for tests.
type stubEncoder struct {
	contentType   string
	fileExtension string
	encodeErr     error
	written       bool
}

func (e *stubEncoder) Encode(w io.Writer, _ export.ExportData) error {
	e.written = true
	if e.encodeErr != nil {
		return e.encodeErr
	}
	_, err := w.Write([]byte("test-output"))
	return err
}

func (e *stubEncoder) ContentType() string   { return e.contentType }
func (e *stubEncoder) FileExtension() string { return e.fileExtension }

func TestExportHandler_ValidFormat(t *testing.T) {
	mock := &mockIndicators{
		result: &store.IndicatorResult{
			Rows: []store.IndicatorRow{
				{Labels: map[string]string{"age_band": "0-4"}, Value: 10, Period: "2025-01"},
			},
		},
	}
	enc := &stubEncoder{contentType: "text/csv", fileExtension: "csv"}
	encoders := map[string]export.Encoder{"csv": enc}

	handler := ExportHandler(mock, encoders)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/export/csv?period_start=2025-01&period_end=2025-06&dataset=demographics", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("format", "csv")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}
	if ct := w.Header().Get("Content-Type"); ct != "text/csv" {
		t.Errorf("expected Content-Type text/csv, got %s", ct)
	}
	if cd := w.Header().Get("Content-Disposition"); !strings.Contains(cd, "attachment") {
		t.Errorf("expected Content-Disposition with attachment, got %s", cd)
	}
	if !enc.written {
		t.Error("expected encoder.Encode to be called")
	}
}

func TestExportHandler_UnsupportedFormat(t *testing.T) {
	mock := &mockIndicators{
		result: &store.IndicatorResult{},
	}
	encoders := map[string]export.Encoder{}

	handler := ExportHandler(mock, encoders)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/export/unknown?period_start=2025-01&period_end=2025-06", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("format", "unknown")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestExportHandler_InvalidDataset(t *testing.T) {
	mock := &mockIndicators{
		result: &store.IndicatorResult{},
	}
	enc := &stubEncoder{contentType: "text/csv", fileExtension: "csv"}
	encoders := map[string]export.Encoder{"csv": enc}

	handler := ExportHandler(mock, encoders)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/export/csv?period_start=2025-01&period_end=2025-06&dataset=invalid", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("format", "csv")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestExportHandler_MissingPeriod(t *testing.T) {
	mock := &mockIndicators{
		result: &store.IndicatorResult{},
	}
	enc := &stubEncoder{contentType: "text/csv", fileExtension: "csv"}
	encoders := map[string]export.Encoder{"csv": enc}

	handler := ExportHandler(mock, encoders)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/export/csv", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("format", "csv")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestExportHandler_DefaultDataset(t *testing.T) {
	mock := &mockIndicators{
		result: &store.IndicatorResult{Rows: []store.IndicatorRow{}},
	}
	enc := &stubEncoder{contentType: "application/json", fileExtension: "json"}
	encoders := map[string]export.Encoder{"json": enc}

	handler := ExportHandler(mock, encoders)

	// No dataset param -- should default to "demographics"
	req := httptest.NewRequest(http.MethodGet, "/api/v1/export/json?period_start=2025-01&period_end=2025-06", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("format", "json")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}
	if mock.called != "demographics" {
		t.Errorf("expected demographics query, got %s", mock.called)
	}
}

func TestExportHandler_ContentDisposition(t *testing.T) {
	mock := &mockIndicators{
		result: &store.IndicatorResult{Rows: []store.IndicatorRow{}},
	}
	enc := &stubEncoder{contentType: "text/csv", fileExtension: "csv"}
	encoders := map[string]export.Encoder{"csv": enc}

	handler := ExportHandler(mock, encoders)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/export/csv?period_start=2025-01&period_end=2025-06&dataset=care", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("format", "csv")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	cd := w.Header().Get("Content-Disposition")
	if !strings.Contains(cd, "care") {
		t.Errorf("Content-Disposition should contain dataset name, got: %s", cd)
	}
	if !strings.Contains(cd, ".csv") {
		t.Errorf("Content-Disposition should contain .csv extension, got: %s", cd)
	}
}
