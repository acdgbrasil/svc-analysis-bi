package store

import (
	"errors"
	"testing"

	"github.com/acdgbrasil/svc-analysis-bi/internal/domain"
)

func TestValidateIndicatorParams(t *testing.T) {
	tests := []struct {
		name    string
		params  IndicatorParams
		wantErr bool
	}{
		{
			name: "valid monthly params",
			params: IndicatorParams{
				PeriodStart: domain.Period{Year: 2025, Month: 1},
				PeriodEnd:   domain.Period{Year: 2025, Month: 6},
				Granularity: domain.GranularityMonthly,
			},
			wantErr: false,
		},
		{
			name: "valid single month",
			params: IndicatorParams{
				PeriodStart: domain.Period{Year: 2025, Month: 3},
				PeriodEnd:   domain.Period{Year: 2025, Month: 3},
				Granularity: domain.GranularityMonthly,
			},
			wantErr: false,
		},
		{
			name: "valid with mesoregion filter",
			params: IndicatorParams{
				PeriodStart: domain.Period{Year: 2025, Month: 1},
				PeriodEnd:   domain.Period{Year: 2025, Month: 12},
				Mesoregion:  "Metropolitana de Sao Paulo",
				Granularity: domain.GranularityQuarterly,
			},
			wantErr: false,
		},
		{
			name: "valid with top limit",
			params: IndicatorParams{
				PeriodStart: domain.Period{Year: 2025, Month: 1},
				PeriodEnd:   domain.Period{Year: 2025, Month: 6},
				Top:         10,
				Granularity: domain.GranularityMonthly,
			},
			wantErr: false,
		},
		{
			name: "invalid PeriodStart month zero",
			params: IndicatorParams{
				PeriodStart: domain.Period{Year: 2025, Month: 0},
				PeriodEnd:   domain.Period{Year: 2025, Month: 6},
			},
			wantErr: true,
		},
		{
			name: "invalid PeriodStart month 13",
			params: IndicatorParams{
				PeriodStart: domain.Period{Year: 2025, Month: 13},
				PeriodEnd:   domain.Period{Year: 2025, Month: 6},
			},
			wantErr: true,
		},
		{
			name: "invalid PeriodEnd month zero",
			params: IndicatorParams{
				PeriodStart: domain.Period{Year: 2025, Month: 1},
				PeriodEnd:   domain.Period{Year: 2025, Month: 0},
			},
			wantErr: true,
		},
		{
			name: "invalid PeriodStart year zero",
			params: IndicatorParams{
				PeriodStart: domain.Period{Year: 0, Month: 1},
				PeriodEnd:   domain.Period{Year: 2025, Month: 6},
			},
			wantErr: true,
		},
		{
			name: "invalid PeriodEnd before PeriodStart",
			params: IndicatorParams{
				PeriodStart: domain.Period{Year: 2025, Month: 6},
				PeriodEnd:   domain.Period{Year: 2025, Month: 1},
			},
			wantErr: true,
		},
		{
			name: "invalid PeriodEnd year before PeriodStart year",
			params: IndicatorParams{
				PeriodStart: domain.Period{Year: 2025, Month: 1},
				PeriodEnd:   domain.Period{Year: 2024, Month: 12},
			},
			wantErr: true,
		},
		{
			name: "invalid negative Top",
			params: IndicatorParams{
				PeriodStart: domain.Period{Year: 2025, Month: 1},
				PeriodEnd:   domain.Period{Year: 2025, Month: 6},
				Top:         -1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateIndicatorParams(tt.params)
			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}
			if tt.wantErr && err != nil {
				if !errors.Is(err, ErrInvalidIndicatorParams) {
					t.Fatalf("expected ErrInvalidIndicatorParams, got: %v", err)
				}
			}
		})
	}
}

func TestPeriodLabel(t *testing.T) {
	tests := []struct {
		yearMonth   string
		granularity domain.TimeGranularity
		want        string
	}{
		{"2025-03", domain.GranularityMonthly, "2025-03"},
		{"2025-01", domain.GranularityQuarterly, "2025-Q1"},
		{"2025-04", domain.GranularityQuarterly, "2025-Q2"},
		{"2025-07", domain.GranularityQuarterly, "2025-Q3"},
		{"2025-10", domain.GranularityQuarterly, "2025-Q4"},
		{"2025-06", domain.GranularityYearly, "2025"},
		{"2024-12", domain.GranularityYearly, "2024"},
		// Edge case: short input.
		{"2025", domain.GranularityMonthly, "2025"},
		{"", domain.GranularityMonthly, ""},
	}

	for _, tt := range tests {
		t.Run(tt.yearMonth+"_"+string(tt.granularity), func(t *testing.T) {
			got := periodLabel(tt.yearMonth, tt.granularity)
			if got != tt.want {
				t.Fatalf("periodLabel(%q, %q) = %q, want %q", tt.yearMonth, tt.granularity, got, tt.want)
			}
		})
	}
}

func TestPeriodGroupExpr(t *testing.T) {
	tests := []struct {
		granularity domain.TimeGranularity
		wantExpr    string
	}{
		{domain.GranularityMonthly, "p.year_month"},
		{domain.GranularityQuarterly, "p.year || '-Q' || p.quarter"},
		{domain.GranularityYearly, "CAST(p.year AS TEXT)"},
		// Unknown granularity defaults to monthly.
		{"unknown", "p.year_month"},
	}

	for _, tt := range tests {
		t.Run(string(tt.granularity), func(t *testing.T) {
			got := periodGroupExpr(tt.granularity)
			if got != tt.wantExpr {
				t.Fatalf("periodGroupExpr(%q) = %q, want %q", tt.granularity, got, tt.wantExpr)
			}
		})
	}
}

func TestNewIndicatorStoreCreation(t *testing.T) {
	store := NewIndicatorStore(nil)
	if store == nil {
		t.Fatal("expected non-nil IndicatorStore")
	}
}

func TestIndicatorStoreQueryWithInvalidParams(t *testing.T) {
	store := NewIndicatorStore(nil)
	invalidParams := IndicatorParams{
		PeriodStart: domain.Period{Year: 2025, Month: 0},
		PeriodEnd:   domain.Period{Year: 2025, Month: 6},
	}

	queries := []struct {
		name string
		fn   func() error
	}{
		{"QueryDemographics", func() error { _, err := store.QueryDemographics(t.Context(), invalidParams); return err }},
		{"QueryEpidemiological", func() error { _, err := store.QueryEpidemiological(t.Context(), invalidParams); return err }},
		{"QuerySocioeconomic", func() error { _, err := store.QuerySocioeconomic(t.Context(), invalidParams); return err }},
		{"QueryProtection", func() error { _, err := store.QueryProtection(t.Context(), invalidParams); return err }},
		{"QueryCare", func() error { _, err := store.QueryCare(t.Context(), invalidParams); return err }},
	}

	for _, q := range queries {
		t.Run(q.name, func(t *testing.T) {
			err := q.fn()
			if err == nil {
				t.Fatal("expected validation error for invalid params")
			}
			if !errors.Is(err, ErrInvalidIndicatorParams) {
				t.Fatalf("expected ErrInvalidIndicatorParams, got: %v", err)
			}
		})
	}
}

func TestIndicatorSentinelErrors(t *testing.T) {
	sentinels := []error{
		ErrIndicatorQueryFailed,
		ErrInvalidIndicatorParams,
	}

	seen := make(map[string]bool, len(sentinels))
	for _, err := range sentinels {
		msg := err.Error()
		if seen[msg] {
			t.Fatalf("duplicate sentinel error message: %s", msg)
		}
		seen[msg] = true

		if msg == "" {
			t.Fatal("sentinel error has empty message")
		}
	}
}

func TestIndicatorResultType(t *testing.T) {
	// Verify IndicatorResult struct works as expected.
	result := IndicatorResult{
		Rows: []IndicatorRow{
			{Labels: map[string]string{"age_band": "0-4"}, Value: 10, Period: "2025-03"},
			{Labels: map[string]string{"age_band": "5-9"}, Value: 3, Period: "2025-03"},
		},
		Suppressed: 1,
	}

	if len(result.Rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(result.Rows))
	}
	if result.Suppressed != 1 {
		t.Fatalf("expected 1 suppressed, got %d", result.Suppressed)
	}
	if result.Rows[0].Labels["age_band"] != "0-4" {
		t.Fatalf("expected age_band '0-4', got %q", result.Rows[0].Labels["age_band"])
	}
	if result.Rows[1].Value != 3 {
		t.Fatalf("expected value 3, got %d", result.Rows[1].Value)
	}
}
