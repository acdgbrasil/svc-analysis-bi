package domain

import (
	"errors"
	"testing"
	"time"
)

func TestNewPeriod(t *testing.T) {
	tests := []struct {
		name      string
		year      int
		month     int
		wantYear  int
		wantMonth int
		wantErr   error
	}{
		{
			name:      "valid period 2026-03",
			year:      2026,
			month:     3,
			wantYear:  2026,
			wantMonth: 3,
			wantErr:   nil,
		},
		{
			name:      "valid period January",
			year:      2026,
			month:     1,
			wantYear:  2026,
			wantMonth: 1,
			wantErr:   nil,
		},
		{
			name:      "valid period December",
			year:      2026,
			month:     12,
			wantYear:  2026,
			wantMonth: 12,
			wantErr:   nil,
		},
		{
			name:    "month 0 returns ErrInvalidPeriod",
			year:    2026,
			month:   0,
			wantErr: ErrInvalidPeriod,
		},
		{
			name:    "month 13 returns ErrInvalidPeriod",
			year:    2026,
			month:   13,
			wantErr: ErrInvalidPeriod,
		},
		{
			name:    "negative month returns ErrInvalidPeriod",
			year:    2026,
			month:   -1,
			wantErr: ErrInvalidPeriod,
		},
		{
			name:    "year 0 returns ErrInvalidPeriod",
			year:    0,
			month:   6,
			wantErr: ErrInvalidPeriod,
		},
		{
			name:    "negative year returns ErrInvalidPeriod",
			year:    -1,
			month:   6,
			wantErr: ErrInvalidPeriod,
		},
		{
			name:      "year 1 month 1 is valid",
			year:      1,
			month:     1,
			wantYear:  1,
			wantMonth: 1,
			wantErr:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewPeriod(tt.year, tt.month)
			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("NewPeriod(%d, %d) expected error %v, got nil", tt.year, tt.month, tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("NewPeriod(%d, %d) error = %v, want %v", tt.year, tt.month, err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("NewPeriod(%d, %d) unexpected error: %v", tt.year, tt.month, err)
			}
			if got.Year != tt.wantYear {
				t.Errorf("NewPeriod(%d, %d).Year = %d, want %d", tt.year, tt.month, got.Year, tt.wantYear)
			}
			if got.Month != tt.wantMonth {
				t.Errorf("NewPeriod(%d, %d).Month = %d, want %d", tt.year, tt.month, got.Month, tt.wantMonth)
			}
		})
	}
}

func TestPeriodFromTime(t *testing.T) {
	tests := []struct {
		name      string
		input     time.Time
		wantYear  int
		wantMonth int
	}{
		{
			name:      "April 15 2026",
			input:     time.Date(2026, 4, 15, 10, 30, 0, 0, time.UTC),
			wantYear:  2026,
			wantMonth: 4,
		},
		{
			name:      "January 1 2025",
			input:     time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			wantYear:  2025,
			wantMonth: 1,
		},
		{
			name:      "December 31 2024",
			input:     time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
			wantYear:  2024,
			wantMonth: 12,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PeriodFromTime(tt.input)
			if got.Year != tt.wantYear {
				t.Errorf("PeriodFromTime(%v).Year = %d, want %d", tt.input, got.Year, tt.wantYear)
			}
			if got.Month != tt.wantMonth {
				t.Errorf("PeriodFromTime(%v).Month = %d, want %d", tt.input, got.Month, tt.wantMonth)
			}
		})
	}
}

func TestPeriod_YearMonth(t *testing.T) {
	tests := []struct {
		name  string
		year  int
		month int
		want  string
	}{
		{"March 2026", 2026, 3, "2026-03"},
		{"January 2025", 2025, 1, "2025-01"},
		{"December 2024", 2024, 12, "2024-12"},
		{"October 2026", 2026, 10, "2026-10"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Period{Year: tt.year, Month: tt.month}
			got := p.YearMonth()
			if got != tt.want {
				t.Errorf("Period{%d, %d}.YearMonth() = %q, want %q", tt.year, tt.month, got, tt.want)
			}
		})
	}
}

func TestPeriod_Quarter(t *testing.T) {
	tests := []struct {
		name  string
		month int
		want  int
	}{
		{"January is Q1", 1, 1},
		{"February is Q1", 2, 1},
		{"March is Q1", 3, 1},
		{"April is Q2", 4, 2},
		{"May is Q2", 5, 2},
		{"June is Q2", 6, 2},
		{"July is Q3", 7, 3},
		{"August is Q3", 8, 3},
		{"September is Q3", 9, 3},
		{"October is Q4", 10, 4},
		{"November is Q4", 11, 4},
		{"December is Q4", 12, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Period{Year: 2026, Month: tt.month}
			got := p.Quarter()
			if got != tt.want {
				t.Errorf("Period{2026, %d}.Quarter() = %d, want %d", tt.month, got, tt.want)
			}
		})
	}
}

func TestPeriod_QuarterLabel(t *testing.T) {
	tests := []struct {
		name  string
		year  int
		month int
		want  string
	}{
		{"January 2026 is 2026-Q1", 2026, 1, "2026-Q1"},
		{"April 2026 is 2026-Q2", 2026, 4, "2026-Q2"},
		{"July 2026 is 2026-Q3", 2026, 7, "2026-Q3"},
		{"October 2026 is 2026-Q4", 2026, 10, "2026-Q4"},
		{"December 2025 is 2025-Q4", 2025, 12, "2025-Q4"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Period{Year: tt.year, Month: tt.month}
			got := p.QuarterLabel()
			if got != tt.want {
				t.Errorf("Period{%d, %d}.QuarterLabel() = %q, want %q", tt.year, tt.month, got, tt.want)
			}
		})
	}
}

func TestPeriod_YearLabel(t *testing.T) {
	tests := []struct {
		name string
		year int
		want string
	}{
		{"2026", 2026, "2026"},
		{"2025", 2025, "2025"},
		{"2000", 2000, "2000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Period{Year: tt.year, Month: 3}
			got := p.YearLabel()
			if got != tt.want {
				t.Errorf("Period{%d, 3}.YearLabel() = %q, want %q", tt.year, got, tt.want)
			}
		})
	}
}

func TestPeriod_Before(t *testing.T) {
	tests := []struct {
		name string
		a    Period
		b    Period
		want bool
	}{
		{
			name: "earlier month same year is before",
			a:    Period{Year: 2026, Month: 1},
			b:    Period{Year: 2026, Month: 2},
			want: true,
		},
		{
			name: "later month same year is not before",
			a:    Period{Year: 2026, Month: 2},
			b:    Period{Year: 2026, Month: 1},
			want: false,
		},
		{
			name: "same period is not before",
			a:    Period{Year: 2026, Month: 3},
			b:    Period{Year: 2026, Month: 3},
			want: false,
		},
		{
			name: "earlier year is before",
			a:    Period{Year: 2025, Month: 12},
			b:    Period{Year: 2026, Month: 1},
			want: true,
		},
		{
			name: "later year is not before",
			a:    Period{Year: 2027, Month: 1},
			b:    Period{Year: 2026, Month: 12},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.a.Before(tt.b)
			if got != tt.want {
				t.Errorf("Period{%d,%d}.Before(Period{%d,%d}) = %v, want %v",
					tt.a.Year, tt.a.Month, tt.b.Year, tt.b.Month, got, tt.want)
			}
		})
	}
}

func TestPeriod_Equal(t *testing.T) {
	tests := []struct {
		name string
		a    Period
		b    Period
		want bool
	}{
		{
			name: "same year and month are equal",
			a:    Period{Year: 2026, Month: 4},
			b:    Period{Year: 2026, Month: 4},
			want: true,
		},
		{
			name: "different month is not equal",
			a:    Period{Year: 2026, Month: 4},
			b:    Period{Year: 2026, Month: 5},
			want: false,
		},
		{
			name: "different year is not equal",
			a:    Period{Year: 2025, Month: 4},
			b:    Period{Year: 2026, Month: 4},
			want: false,
		},
		{
			name: "both different is not equal",
			a:    Period{Year: 2025, Month: 3},
			b:    Period{Year: 2026, Month: 4},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.a.Equal(tt.b)
			if got != tt.want {
				t.Errorf("Period{%d,%d}.Equal(Period{%d,%d}) = %v, want %v",
					tt.a.Year, tt.a.Month, tt.b.Year, tt.b.Month, got, tt.want)
			}
		})
	}
}
