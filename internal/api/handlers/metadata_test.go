package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/acdgbrasil/svc-analysis-bi/internal/export"
)

// metaEncoder implements export.Encoder for metadata tests.
type metaEncoder struct {
	ct  string
	ext string
}

func (e *metaEncoder) Encode(_ io.Writer, _ export.ExportData) error { return nil }
func (e *metaEncoder) ContentType() string                          { return e.ct }
func (e *metaEncoder) FileExtension() string                       { return e.ext }

func TestMetadataHandler_Datasets(t *testing.T) {
	encoders := map[string]export.Encoder{}
	handler := MetadataHandler(encoders)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/metadata/datasets", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("resource", "datasets")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp Response
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}

	datasets, ok := resp.Data.([]any)
	if !ok {
		t.Fatal("expected data to be an array")
	}
	if len(datasets) != 5 {
		t.Errorf("expected 5 datasets, got %d", len(datasets))
	}
	if resp.Meta.TotalRecords != 5 {
		t.Errorf("expected total_records=5, got %d", resp.Meta.TotalRecords)
	}
}

func TestMetadataHandler_Formats(t *testing.T) {
	encoders := map[string]export.Encoder{
		"csv":  &metaEncoder{ct: "text/csv", ext: "csv"},
		"json": &metaEncoder{ct: "application/json", ext: "json"},
	}
	handler := MetadataHandler(encoders)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/metadata/formats", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("resource", "formats")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp Response
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}

	formats, ok := resp.Data.([]any)
	if !ok {
		t.Fatal("expected data to be an array")
	}
	if len(formats) != 2 {
		t.Errorf("expected 2 formats, got %d", len(formats))
	}
	if resp.Meta.TotalRecords != 2 {
		t.Errorf("expected total_records=2, got %d", resp.Meta.TotalRecords)
	}

	// Verify sorted order: csv before json
	first, ok := formats[0].(map[string]any)
	if !ok {
		t.Fatal("expected format to be a map")
	}
	if first["name"] != "csv" {
		t.Errorf("expected first format to be csv, got %v", first["name"])
	}
}

func TestMetadataHandler_Regions(t *testing.T) {
	encoders := map[string]export.Encoder{}
	handler := MetadataHandler(encoders)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/metadata/regions", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("resource", "regions")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestMetadataHandler_UnknownResource(t *testing.T) {
	encoders := map[string]export.Encoder{}
	handler := MetadataHandler(encoders)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/metadata/unknown", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("resource", "unknown")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}
