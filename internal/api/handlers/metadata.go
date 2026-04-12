package handlers

import (
	"net/http"
	"sort"

	"github.com/go-chi/chi/v5"

	"github.com/acdgbrasil/svc-analysis-bi/internal/export"
)

// datasetInfo describes an available indicator dataset.
type datasetInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// formatInfo describes an available export format.
type formatInfo struct {
	Name        string `json:"name"`
	ContentType string `json:"content_type"`
	Extension   string `json:"extension"`
}

// availableDatasets is the static list of indicator datasets.
var availableDatasets = []datasetInfo{
	{Name: "demographics", Description: "Population pyramid by age band, sex, and mesoregion"},
	{Name: "epidemiological", Description: "Top diagnoses by ICD code and total cases"},
	{Name: "socioeconomic", Description: "Income band distribution by mesoregion"},
	{Name: "protection", Description: "Referrals and rights violation reports"},
	{Name: "care", Description: "Appointment counts by type and mesoregion"},
}

// MetadataHandler returns an http.HandlerFunc that serves metadata about
// available datasets, export formats, and regions.
func MetadataHandler(encoders map[string]export.Encoder) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resource := chi.URLParam(r, "resource")

		switch resource {
		case "datasets":
			meta := NewResponseMeta()
			meta.TotalRecords = len(availableDatasets)
			WriteJSON(w, http.StatusOK, Response{
				Data: availableDatasets,
				Meta: meta,
			})

		case "formats":
			formats := buildFormatList(encoders)
			meta := NewResponseMeta()
			meta.TotalRecords = len(formats)
			WriteJSON(w, http.StatusOK, Response{
				Data: formats,
				Meta: meta,
			})

		case "regions":
			// Placeholder: in a full implementation this would query dim_geography.
			meta := NewResponseMeta()
			WriteJSON(w, http.StatusOK, Response{
				Data: []any{},
				Meta: meta,
			})

		default:
			WriteError(w, http.StatusBadRequest, "unknown metadata resource: "+resource)
		}
	}
}

// buildFormatList converts the encoder registry into a sorted list of formatInfo.
func buildFormatList(encoders map[string]export.Encoder) []formatInfo {
	formats := make([]formatInfo, 0, len(encoders))
	for name, enc := range encoders {
		formats = append(formats, formatInfo{
			Name:        name,
			ContentType: enc.ContentType(),
			Extension:   enc.FileExtension(),
		})
	}
	sort.Slice(formats, func(i, j int) bool {
		return formats[i].Name < formats[j].Name
	})
	return formats
}
