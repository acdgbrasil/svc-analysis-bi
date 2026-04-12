package domain

import (
	"fmt"
	"time"
)

// NewPeriod creates a Period from explicit year and month values.
// Returns ErrInvalidPeriod if month is outside 1-12 or year is not positive.
func NewPeriod(year int, month int) (Period, error) {
	if year < 1 || month < 1 || month > 12 {
		return Period{}, ErrInvalidPeriod
	}
	return Period{Year: year, Month: month}, nil
}

// PeriodFromTime extracts the year and month from a time.Time value.
func PeriodFromTime(t time.Time) Period {
	return Period{Year: t.Year(), Month: int(t.Month())}
}

// YearMonth formats the period as "YYYY-MM".
func (p Period) YearMonth() string {
	return fmt.Sprintf("%04d-%02d", p.Year, p.Month)
}

// Quarter returns the calendar quarter (1-4) for the period's month.
func (p Period) Quarter() int {
	return (p.Month-1)/3 + 1
}

// QuarterLabel formats the period as "YYYY-QN".
func (p Period) QuarterLabel() string {
	return fmt.Sprintf("%04d-Q%d", p.Year, p.Quarter())
}

// YearLabel formats the period as "YYYY".
func (p Period) YearLabel() string {
	return fmt.Sprintf("%d", p.Year)
}

// Before reports whether period p is chronologically before other.
func (p Period) Before(other Period) bool {
	if p.Year != other.Year {
		return p.Year < other.Year
	}
	return p.Month < other.Month
}

// Equal reports whether period p represents the same month as other.
func (p Period) Equal(other Period) bool {
	return p.Year == other.Year && p.Month == other.Month
}
