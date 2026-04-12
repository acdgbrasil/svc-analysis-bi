package handlers

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/acdgbrasil/svc-analysis-bi/internal/domain"
	"github.com/acdgbrasil/svc-analysis-bi/internal/store"
)

// IndicatorQuerier defines the query methods used by the indicator handler.
// Implemented by store.IndicatorStore.
type IndicatorQuerier interface {
	QueryDemographics(ctx context.Context, params store.IndicatorParams) (*store.IndicatorResult, error)
	QueryEpidemiological(ctx context.Context, params store.IndicatorParams) (*store.IndicatorResult, error)
	QuerySocioeconomic(ctx context.Context, params store.IndicatorParams) (*store.IndicatorResult, error)
	QueryProtection(ctx context.Context, params store.IndicatorParams) (*store.IndicatorResult, error)
	QueryCare(ctx context.Context, params store.IndicatorParams) (*store.IndicatorResult, error)
}

// validAxes is the set of indicator axes that this handler supports.
var validAxes = map[string]bool{
	"demographics":    true,
	"epidemiological": true,
	"socioeconomic":   true,
	"protection":      true,
	"care":            true,
}

// IndicatorsHandler returns an http.HandlerFunc that dispatches indicator
// queries based on the {axis} URL parameter.
func IndicatorsHandler(indicators IndicatorQuerier) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		axis := chi.URLParam(r, "axis")
		if !validAxes[axis] {
			WriteError(w, http.StatusBadRequest, "unknown indicator axis: "+axis)
			return
		}

		params, err := parseIndicatorParams(r)
		if err != nil {
			WriteError(w, http.StatusBadRequest, err.Error())
			return
		}

		result, err := queryByAxis(r.Context(), indicators, axis, params)
		if err != nil {
			if isValidationError(err) {
				WriteError(w, http.StatusBadRequest, err.Error())
				return
			}
			WriteError(w, http.StatusInternalServerError, "internal error")
			return
		}

		meta := NewResponseMeta()
		meta.SuppressedGroups = result.Suppressed
		meta.TotalRecords = len(result.Rows)
		meta.Period = params.PeriodStart.YearMonth() + "/" + params.PeriodEnd.YearMonth()

		WriteJSON(w, http.StatusOK, Response{
			Data: result.Rows,
			Meta: meta,
		})
	}
}

// queryByAxis dispatches to the correct IndicatorStore method based on axis.
func queryByAxis(ctx context.Context, q IndicatorQuerier, axis string, params store.IndicatorParams) (*store.IndicatorResult, error) {
	switch axis {
	case "demographics":
		return q.QueryDemographics(ctx, params)
	case "epidemiological":
		return q.QueryEpidemiological(ctx, params)
	case "socioeconomic":
		return q.QuerySocioeconomic(ctx, params)
	case "protection":
		return q.QueryProtection(ctx, params)
	case "care":
		return q.QueryCare(ctx, params)
	default:
		return nil, store.ErrInvalidIndicatorParams
	}
}

// parseIndicatorParams extracts IndicatorParams from query string values.
// Required: period_start, period_end (format YYYY-MM).
// Optional: mesoregion, granularity (monthly|quarterly|yearly), top (int).
func parseIndicatorParams(r *http.Request) (store.IndicatorParams, error) {
	q := r.URL.Query()

	periodStart, err := parsePeriod(q.Get("period_start"))
	if err != nil {
		return store.IndicatorParams{}, err
	}

	periodEnd, err := parsePeriod(q.Get("period_end"))
	if err != nil {
		return store.IndicatorParams{}, err
	}

	granularity := parseGranularity(q.Get("granularity"))

	top := 0
	if topStr := q.Get("top"); topStr != "" {
		top, err = strconv.Atoi(topStr)
		if err != nil || top < 0 {
			return store.IndicatorParams{}, store.ErrInvalidIndicatorParams
		}
	}

	return store.IndicatorParams{
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		Mesoregion:  q.Get("mesoregion"),
		Granularity: granularity,
		Top:         top,
	}, nil
}

// parsePeriod parses a "YYYY-MM" string into a domain.Period.
func parsePeriod(s string) (domain.Period, error) {
	if s == "" {
		return domain.Period{}, store.ErrInvalidIndicatorParams
	}
	parts := strings.SplitN(s, "-", 2)
	if len(parts) != 2 {
		return domain.Period{}, store.ErrInvalidIndicatorParams
	}
	year, err := strconv.Atoi(parts[0])
	if err != nil {
		return domain.Period{}, store.ErrInvalidIndicatorParams
	}
	month, err := strconv.Atoi(parts[1])
	if err != nil {
		return domain.Period{}, store.ErrInvalidIndicatorParams
	}
	return domain.NewPeriod(year, month)
}

// parseGranularity converts a string to a TimeGranularity with "monthly" as default.
func parseGranularity(s string) domain.TimeGranularity {
	switch s {
	case "quarterly":
		return domain.GranularityQuarterly
	case "yearly":
		return domain.GranularityYearly
	default:
		return domain.GranularityMonthly
	}
}

// isValidationError checks whether the error is a parameter validation error.
func isValidationError(err error) bool {
	return strings.Contains(err.Error(), "invalid indicator params")
}
