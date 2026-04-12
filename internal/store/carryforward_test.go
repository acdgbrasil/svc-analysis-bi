package store

import (
	"testing"

	"github.com/acdgbrasil/svc-analysis-bi/internal/domain"
)

func TestPreviousPeriod(t *testing.T) {
	tests := []struct {
		name     string
		input    domain.Period
		expected domain.Period
	}{
		{
			name:     "normal month decrements",
			input:    domain.Period{Year: 2025, Month: 6},
			expected: domain.Period{Year: 2025, Month: 5},
		},
		{
			name:     "January wraps to December of previous year",
			input:    domain.Period{Year: 2025, Month: 1},
			expected: domain.Period{Year: 2024, Month: 12},
		},
		{
			name:     "March to February",
			input:    domain.Period{Year: 2026, Month: 3},
			expected: domain.Period{Year: 2026, Month: 2},
		},
		{
			name:     "February to January",
			input:    domain.Period{Year: 2026, Month: 2},
			expected: domain.Period{Year: 2026, Month: 1},
		},
		{
			name:     "December to November",
			input:    domain.Period{Year: 2025, Month: 12},
			expected: domain.Period{Year: 2025, Month: 11},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := previousPeriod(tt.input)
			if got.Year != tt.expected.Year || got.Month != tt.expected.Month {
				t.Errorf("previousPeriod(%v) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestNewCarryForwardJob(t *testing.T) {
	// NewCarryForwardJob should not panic with a nil pool (construction only).
	// The pool is only used at Run() time.
	job := NewCarryForwardJob(nil)
	if job == nil {
		t.Fatal("NewCarryForwardJob returned nil")
	}
}
