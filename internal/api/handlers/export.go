package handlers

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/acdgbrasil/svc-analysis-bi/internal/domain"
	"github.com/acdgbrasil/svc-analysis-bi/internal/export"
	"github.com/acdgbrasil/svc-analysis-bi/internal/store"
)

// ExportHandler returns an http.HandlerFunc that exports indicator data in the
// requested format. The {format} path parameter selects the encoder from the
// registry.
//
// Query parameters:
//   - dataset: one of demographics, epidemiological, socioeconomic, protection, care
//   - period_start, period_end: YYYY-MM range
//   - mesoregion: optional mesoregion filter
func ExportHandler(indicators IndicatorQuerier, encoders map[string]export.Encoder, logger ...*slog.Logger) http.HandlerFunc {
	log := slog.Default()
	if len(logger) > 0 && logger[0] != nil {
		log = logger[0]
	}

	return func(w http.ResponseWriter, r *http.Request) {
		format := chi.URLParam(r, "format")

		enc, ok := encoders[format]
		if !ok {
			WriteError(w, http.StatusBadRequest, "unsupported export format: "+format)
			return
		}

		dataset := r.URL.Query().Get("dataset")
		if dataset == "" {
			dataset = "demographics"
		}
		if !validAxes[dataset] {
			WriteError(w, http.StatusBadRequest, "unknown dataset: "+dataset)
			return
		}

		params, err := parseIndicatorParams(r)
		if err != nil {
			WriteError(w, http.StatusBadRequest, err.Error())
			return
		}

		result, err := queryByAxis(r.Context(), indicators, dataset, params)
		if err != nil {
			if isValidationError(err) {
				WriteError(w, http.StatusBadRequest, err.Error())
				return
			}
			WriteError(w, http.StatusInternalServerError, "internal error")
			return
		}

		exportData := toExportData(dataset, params, result)

		period := params.PeriodStart.YearMonth() + "_" + params.PeriodEnd.YearMonth()
		w.Header().Set("Content-Type", enc.ContentType())
		w.Header().Set("Content-Disposition", export.ContentDisposition(dataset, period, enc.FileExtension()))

		if err := enc.Encode(w, exportData); err != nil {
			// Headers already sent; cannot change status code.
			log.Warn("export encode failed after headers sent",
				"format", format, "dataset", dataset, "error", err)
			return
		}
	}
}

// toExportData converts an IndicatorResult to export.ExportData.
func toExportData(dataset string, params store.IndicatorParams, result *store.IndicatorResult) export.ExportData {
	rows := make([]export.ExportRow, len(result.Rows))
	for i, row := range result.Rows {
		labels := make(map[string]string, len(row.Labels))
		for k, v := range row.Labels {
			labels[k] = v
		}
		rows[i] = export.ExportRow{
			Labels: labels,
			Values: map[string]any{
				"value":  row.Value,
				"period": row.Period,
			},
		}
	}

	return export.ExportData{
		Dataset: dataset,
		Rows:    rows,
		Metadata: export.ExportMetadata{
			Period:       params.PeriodStart.YearMonth() + "/" + params.PeriodEnd.YearMonth(),
			KThreshold:   domain.KThreshold,
			Suppressed:   result.Suppressed,
			TotalRecords: len(result.Rows),
			GeneratedAt:  time.Now().UTC(),
		},
	}
}
