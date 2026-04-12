package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/acdgbrasil/svc-analysis-bi/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Sentinel errors for indicator queries.
var (
	ErrIndicatorQueryFailed   = errors.New("store: indicator query failed")
	ErrInvalidIndicatorParams = errors.New("store: invalid indicator params")
)

// IndicatorParams configures a query against indicator fact tables.
type IndicatorParams struct {
	// PeriodStart is the inclusive start of the time range (YYYY-MM).
	PeriodStart domain.Period

	// PeriodEnd is the inclusive end of the time range (YYYY-MM).
	PeriodEnd domain.Period

	// Mesoregion optionally filters by a specific mesoregion name.
	// An empty string means no filter.
	Mesoregion string

	// Granularity controls how time-series data is aggregated.
	Granularity domain.TimeGranularity

	// Top limits the number of rows returned for top-N queries (e.g.
	// epidemiological top diagnoses). Zero means no limit.
	Top int
}

// IndicatorRow is a single row in an indicator query result.
type IndicatorRow struct {
	// Labels holds the dimension values for this row (e.g. "age_band", "sex").
	Labels map[string]string

	// Value is the aggregate count or sum for this group.
	Value int

	// Period is the formatted time label for this row (e.g. "2025-03",
	// "2025-Q1", "2025") depending on the requested granularity.
	Period string
}

// IndicatorResult wraps the query results with suppression metadata.
type IndicatorResult struct {
	Rows       []IndicatorRow
	Suppressed int
}

// ValidateIndicatorParams checks that the parameters are well-formed.
func ValidateIndicatorParams(p IndicatorParams) error {
	if p.PeriodStart.Year < 1 || p.PeriodStart.Month < 1 || p.PeriodStart.Month > 12 {
		return fmt.Errorf("%w: invalid PeriodStart", ErrInvalidIndicatorParams)
	}
	if p.PeriodEnd.Year < 1 || p.PeriodEnd.Month < 1 || p.PeriodEnd.Month > 12 {
		return fmt.Errorf("%w: invalid PeriodEnd", ErrInvalidIndicatorParams)
	}
	if p.PeriodEnd.Before(p.PeriodStart) {
		return fmt.Errorf("%w: PeriodEnd is before PeriodStart", ErrInvalidIndicatorParams)
	}
	if p.Top < 0 {
		return fmt.Errorf("%w: Top must be non-negative", ErrInvalidIndicatorParams)
	}
	return nil
}

// IndicatorStore executes indicator queries against the star schema with
// K-anonymity filtering applied at query time.
type IndicatorStore struct {
	pool *pgxpool.Pool
}

// NewIndicatorStore creates an IndicatorStore backed by the given pool.
func NewIndicatorStore(pool *pgxpool.Pool) *IndicatorStore {
	return &IndicatorStore{pool: pool}
}

// periodLabel formats a year_month string according to the requested granularity.
// It returns the original year_month for monthly, or derives quarter/year labels.
func periodLabel(yearMonth string, g domain.TimeGranularity) string {
	// yearMonth is "YYYY-MM" from the database.
	if len(yearMonth) < 7 {
		return yearMonth
	}
	switch g {
	case domain.GranularityQuarterly:
		// Parse month to compute quarter.
		var year, month int
		if _, err := fmt.Sscanf(yearMonth, "%d-%d", &year, &month); err == nil {
			q := (month-1)/3 + 1
			return fmt.Sprintf("%d-Q%d", year, q)
		}
		return yearMonth
	case domain.GranularityYearly:
		return yearMonth[:4]
	default:
		return yearMonth
	}
}

// periodGroupExpr returns the SQL expression used for grouping time
// at the requested granularity. The expression references the dim_period
// table aliased as "p".
func periodGroupExpr(g domain.TimeGranularity) string {
	switch g {
	case domain.GranularityQuarterly:
		return "p.year || '-Q' || p.quarter"
	case domain.GranularityYearly:
		return "CAST(p.year AS TEXT)"
	default:
		return "p.year_month"
	}
}

// combinedPeriodGroupExpr returns the SQL expression used for grouping time
// in queries over a subquery/CTE where dim_period columns are exposed as
// bare column names (year_month, year, quarter) without a table alias.
func combinedPeriodGroupExpr(g domain.TimeGranularity) string {
	switch g {
	case domain.GranularityQuarterly:
		return "year || '-Q' || quarter"
	case domain.GranularityYearly:
		return "CAST(year AS TEXT)"
	default:
		return "year_month"
	}
}

// QueryDemographics returns a population pyramid: count grouped by age band,
// sex, mesoregion, and period. Groups below K=5 are suppressed by a HAVING
// clause.
func (s *IndicatorStore) QueryDemographics(ctx context.Context, params IndicatorParams) (*IndicatorResult, error) {
	if err := ValidateIndicatorParams(params); err != nil {
		return nil, err
	}

	periodExpr := periodGroupExpr(params.Granularity)
	startYM := params.PeriodStart.YearMonth()
	endYM := params.PeriodEnd.YearMonth()

	// Build base query. The HAVING clause enforces K-anonymity.
	query := fmt.Sprintf(`
		SELECT ab.band_label, s.label AS sex, g.mesoregion_name,
		       %s AS period_label, COUNT(*) AS cnt
		FROM fact_patient_snapshot fps
		JOIN dim_age_band ab ON fps.age_band_id = ab.id
		JOIN dim_sex s ON fps.sex_id = s.id
		JOIN dim_geography g ON fps.geography_id = g.id
		JOIN dim_period p ON fps.period_id = p.id
		WHERE p.year_month BETWEEN $1 AND $2
	`, periodExpr)

	args := []any{startYM, endYM}
	argIdx := 3

	if params.Mesoregion != "" {
		query += fmt.Sprintf(" AND g.mesoregion_name = $%d", argIdx)
		args = append(args, params.Mesoregion)
		argIdx++
	}

	query += fmt.Sprintf(`
		GROUP BY ab.band_label, s.label, g.mesoregion_name, %s
		HAVING COUNT(*) >= %d
		ORDER BY period_label, ab.band_label, s.label
	`, periodExpr, domain.KThreshold)

	_ = argIdx // suppress unused lint

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%w: demographics: %v", ErrIndicatorQueryFailed, err)
	}
	defer rows.Close()

	var result []IndicatorRow
	for rows.Next() {
		var bandLabel, sex, mesoregion, periodLbl string
		var cnt int
		if err := rows.Scan(&bandLabel, &sex, &mesoregion, &periodLbl, &cnt); err != nil {
			return nil, fmt.Errorf("%w: demographics scan: %v", ErrIndicatorQueryFailed, err)
		}
		result = append(result, IndicatorRow{
			Labels: map[string]string{
				"age_band":   bandLabel,
				"sex":        sex,
				"mesoregion": mesoregion,
			},
			Value:  cnt,
			Period: periodLbl,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: demographics rows: %v", ErrIndicatorQueryFailed, err)
	}

	// Count suppressed: query total groups and subtract returned.
	suppressed, err := s.countSuppressedDemographics(ctx, startYM, endYM, params)
	if err != nil {
		// Non-fatal: return results without suppression count.
		suppressed = 0
	}

	return &IndicatorResult{Rows: result, Suppressed: suppressed}, nil
}

// countSuppressedDemographics counts how many groups fall below the K threshold.
func (s *IndicatorStore) countSuppressedDemographics(ctx context.Context, startYM, endYM string, params IndicatorParams) (int, error) {
	periodExpr := periodGroupExpr(params.Granularity)

	query := `
		SELECT COUNT(*) FROM (
			SELECT 1
			FROM fact_patient_snapshot fps
			JOIN dim_age_band ab ON fps.age_band_id = ab.id
			JOIN dim_sex s ON fps.sex_id = s.id
			JOIN dim_geography g ON fps.geography_id = g.id
			JOIN dim_period p ON fps.period_id = p.id
			WHERE p.year_month BETWEEN $1 AND $2
	`

	args := []any{startYM, endYM}
	argIdx := 3

	if params.Mesoregion != "" {
		query += fmt.Sprintf(" AND g.mesoregion_name = $%d", argIdx)
		args = append(args, params.Mesoregion)
		argIdx++
	}

	query += fmt.Sprintf(`
			GROUP BY ab.band_label, s.label, g.mesoregion_name, %s
			HAVING COUNT(*) < %d
		) suppressed
	`, periodExpr, domain.KThreshold)

	_ = argIdx // suppress unused lint

	var count int
	if err := s.pool.QueryRow(ctx, query, args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("%w: suppressed count: %v", ErrIndicatorQueryFailed, err)
	}
	return count, nil
}

// QueryEpidemiological returns the top-N diagnoses by total cases for the
// given period range. Groups below K=5 are suppressed.
func (s *IndicatorStore) QueryEpidemiological(ctx context.Context, params IndicatorParams) (*IndicatorResult, error) {
	if err := ValidateIndicatorParams(params); err != nil {
		return nil, err
	}

	periodExpr := periodGroupExpr(params.Granularity)
	startYM := params.PeriodStart.YearMonth()
	endYM := params.PeriodEnd.YearMonth()

	query := fmt.Sprintf(`
		SELECT d.icd_code, d.icd_label, %s AS period_label,
		       SUM(fd.total_cases) AS total
		FROM fact_diagnosis fd
		JOIN dim_diagnosis d ON fd.diagnosis_id = d.id
		JOIN dim_period p ON fd.period_id = p.id
		WHERE p.year_month BETWEEN $1 AND $2
		GROUP BY d.icd_code, d.icd_label, %s
		HAVING SUM(fd.total_cases) >= %d
		ORDER BY total DESC
	`, periodExpr, periodExpr, domain.KThreshold)

	args := []any{startYM, endYM}
	argIdx := 3

	if params.Top > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIdx)
		args = append(args, params.Top)
		argIdx++
	}

	_ = argIdx // suppress unused lint

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%w: epidemiological: %v", ErrIndicatorQueryFailed, err)
	}
	defer rows.Close()

	var result []IndicatorRow
	for rows.Next() {
		var icdCode, icdLabel, periodLbl string
		var total int
		if err := rows.Scan(&icdCode, &icdLabel, &periodLbl, &total); err != nil {
			return nil, fmt.Errorf("%w: epidemiological scan: %v", ErrIndicatorQueryFailed, err)
		}
		result = append(result, IndicatorRow{
			Labels: map[string]string{
				"icd_code":  icdCode,
				"icd_label": icdLabel,
			},
			Value:  total,
			Period: periodLbl,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: epidemiological rows: %v", ErrIndicatorQueryFailed, err)
	}

	// Count suppressed diagnosis groups.
	suppressed, err := s.countSuppressedEpidemiological(ctx, startYM, endYM, params)
	if err != nil {
		suppressed = 0
	}

	return &IndicatorResult{Rows: result, Suppressed: suppressed}, nil
}

func (s *IndicatorStore) countSuppressedEpidemiological(ctx context.Context, startYM, endYM string, params IndicatorParams) (int, error) {
	periodExpr := periodGroupExpr(params.Granularity)

	query := fmt.Sprintf(`
		SELECT COUNT(*) FROM (
			SELECT 1
			FROM fact_diagnosis fd
			JOIN dim_diagnosis d ON fd.diagnosis_id = d.id
			JOIN dim_period p ON fd.period_id = p.id
			WHERE p.year_month BETWEEN $1 AND $2
			GROUP BY d.icd_code, d.icd_label, %s
			HAVING SUM(fd.total_cases) < %d
		) suppressed
	`, periodExpr, domain.KThreshold)

	var count int
	if err := s.pool.QueryRow(ctx, query, startYM, endYM).Scan(&count); err != nil {
		return 0, fmt.Errorf("%w: suppressed count: %v", ErrIndicatorQueryFailed, err)
	}
	return count, nil
}

// QuerySocioeconomic returns socioeconomic indicators from patient snapshots:
// income band distribution, benefit recipients, food insecurity, and
// overcrowding rates. Groups below K=5 are suppressed.
func (s *IndicatorStore) QuerySocioeconomic(ctx context.Context, params IndicatorParams) (*IndicatorResult, error) {
	if err := ValidateIndicatorParams(params); err != nil {
		return nil, err
	}

	periodExpr := periodGroupExpr(params.Granularity)
	startYM := params.PeriodStart.YearMonth()
	endYM := params.PeriodEnd.YearMonth()

	query := fmt.Sprintf(`
		SELECT fps.income_band, g.mesoregion_name,
		       %s AS period_label, COUNT(*) AS cnt
		FROM fact_patient_snapshot fps
		JOIN dim_geography g ON fps.geography_id = g.id
		JOIN dim_period p ON fps.period_id = p.id
		WHERE p.year_month BETWEEN $1 AND $2
		  AND fps.income_band IS NOT NULL AND fps.income_band != ''
		GROUP BY fps.income_band, g.mesoregion_name, %s
		HAVING COUNT(*) >= %d
		ORDER BY period_label, cnt DESC
	`, periodExpr, periodExpr, domain.KThreshold)

	args := []any{startYM, endYM}

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%w: socioeconomic: %v", ErrIndicatorQueryFailed, err)
	}
	defer rows.Close()

	var result []IndicatorRow
	for rows.Next() {
		var incomeBand, mesoregion, periodLbl string
		var cnt int
		if err := rows.Scan(&incomeBand, &mesoregion, &periodLbl, &cnt); err != nil {
			return nil, fmt.Errorf("%w: socioeconomic scan: %v", ErrIndicatorQueryFailed, err)
		}
		result = append(result, IndicatorRow{
			Labels: map[string]string{
				"income_band": incomeBand,
				"mesoregion":  mesoregion,
			},
			Value:  cnt,
			Period: periodLbl,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: socioeconomic rows: %v", ErrIndicatorQueryFailed, err)
	}

	suppressed, err := s.countSuppressedSocioeconomic(ctx, startYM, endYM, params)
	if err != nil {
		suppressed = 0
	}

	return &IndicatorResult{Rows: result, Suppressed: suppressed}, nil
}

func (s *IndicatorStore) countSuppressedSocioeconomic(ctx context.Context, startYM, endYM string, params IndicatorParams) (int, error) {
	periodExpr := periodGroupExpr(params.Granularity)

	query := fmt.Sprintf(`
		SELECT COUNT(*) FROM (
			SELECT 1
			FROM fact_patient_snapshot fps
			JOIN dim_geography g ON fps.geography_id = g.id
			JOIN dim_period p ON fps.period_id = p.id
			WHERE p.year_month BETWEEN $1 AND $2
			  AND fps.income_band IS NOT NULL AND fps.income_band != ''
			GROUP BY fps.income_band, g.mesoregion_name, %s
			HAVING COUNT(*) < %d
		) suppressed
	`, periodExpr, domain.KThreshold)

	var count int
	if err := s.pool.QueryRow(ctx, query, startYM, endYM).Scan(&count); err != nil {
		return 0, fmt.Errorf("%w: suppressed count: %v", ErrIndicatorQueryFailed, err)
	}
	return count, nil
}

// QueryProtection returns protection indicators: referrals by destination
// and violations by type. Groups below K=5 are suppressed.
func (s *IndicatorStore) QueryProtection(ctx context.Context, params IndicatorParams) (*IndicatorResult, error) {
	if err := ValidateIndicatorParams(params); err != nil {
		return nil, err
	}

	combinedExpr := combinedPeriodGroupExpr(params.Granularity)
	startYM := params.PeriodStart.YearMonth()
	endYM := params.PeriodEnd.YearMonth()

	// Combine referrals and violations via UNION ALL. The subquery exposes
	// dim_period columns as bare names so the outer query must use
	// combinedPeriodGroupExpr (without "p." prefix).
	query := fmt.Sprintf(`
		SELECT category, label, %s AS period_label, SUM(cnt) AS total FROM (
			SELECT 'referral' AS category, rd.label, p.year_month, p.year, p.quarter,
			       fr.count AS cnt
			FROM fact_referral fr
			JOIN dim_referral_destination rd ON fr.destination_id = rd.id
			JOIN dim_period p ON fr.period_id = p.id
			WHERE p.year_month BETWEEN $1 AND $2
			UNION ALL
			SELECT 'violation' AS category, vt.label, p.year_month, p.year, p.quarter,
			       fv.count AS cnt
			FROM fact_violation fv
			JOIN dim_violation_type vt ON fv.violation_type_id = vt.id
			JOIN dim_period p ON fv.period_id = p.id
			WHERE p.year_month BETWEEN $1 AND $2
		) combined
		GROUP BY category, label, %s
		HAVING SUM(cnt) >= %d
		ORDER BY period_label, total DESC
	`, combinedExpr, combinedExpr, domain.KThreshold)

	args := []any{startYM, endYM}

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%w: protection: %v", ErrIndicatorQueryFailed, err)
	}
	defer rows.Close()

	var result []IndicatorRow
	for rows.Next() {
		var category, label, periodLbl string
		var total int
		if err := rows.Scan(&category, &label, &periodLbl, &total); err != nil {
			return nil, fmt.Errorf("%w: protection scan: %v", ErrIndicatorQueryFailed, err)
		}
		result = append(result, IndicatorRow{
			Labels: map[string]string{
				"category": category,
				"label":    label,
			},
			Value:  total,
			Period: periodLbl,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: protection rows: %v", ErrIndicatorQueryFailed, err)
	}

	// Protection suppressed count is harder with UNION; use 0 as fallback.
	return &IndicatorResult{Rows: result, Suppressed: 0}, nil
}

// QueryCare returns care indicators: appointment counts by type. Groups
// below K=5 are suppressed.
func (s *IndicatorStore) QueryCare(ctx context.Context, params IndicatorParams) (*IndicatorResult, error) {
	if err := ValidateIndicatorParams(params); err != nil {
		return nil, err
	}

	periodExpr := periodGroupExpr(params.Granularity)
	startYM := params.PeriodStart.YearMonth()
	endYM := params.PeriodEnd.YearMonth()

	query := fmt.Sprintf(`
		SELECT fa.appointment_type, g.mesoregion_name,
		       %s AS period_label, SUM(fa.count) AS total
		FROM fact_appointment fa
		JOIN dim_geography g ON fa.geography_id = g.id
		JOIN dim_period p ON fa.period_id = p.id
		WHERE p.year_month BETWEEN $1 AND $2
		GROUP BY fa.appointment_type, g.mesoregion_name, %s
		HAVING SUM(fa.count) >= %d
		ORDER BY period_label, total DESC
	`, periodExpr, periodExpr, domain.KThreshold)

	args := []any{startYM, endYM}

	if params.Mesoregion != "" {
		// Rebuild with mesoregion filter.
		query = fmt.Sprintf(`
			SELECT fa.appointment_type, g.mesoregion_name,
			       %s AS period_label, SUM(fa.count) AS total
			FROM fact_appointment fa
			JOIN dim_geography g ON fa.geography_id = g.id
			JOIN dim_period p ON fa.period_id = p.id
			WHERE p.year_month BETWEEN $1 AND $2
			  AND g.mesoregion_name = $3
			GROUP BY fa.appointment_type, g.mesoregion_name, %s
			HAVING SUM(fa.count) >= %d
			ORDER BY period_label, total DESC
		`, periodExpr, periodExpr, domain.KThreshold)
		args = append(args, params.Mesoregion)
	}

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%w: care: %v", ErrIndicatorQueryFailed, err)
	}
	defer rows.Close()

	var result []IndicatorRow
	for rows.Next() {
		var appointmentType, mesoregion, periodLbl string
		var total int
		if err := rows.Scan(&appointmentType, &mesoregion, &periodLbl, &total); err != nil {
			return nil, fmt.Errorf("%w: care scan: %v", ErrIndicatorQueryFailed, err)
		}
		result = append(result, IndicatorRow{
			Labels: map[string]string{
				"appointment_type": appointmentType,
				"mesoregion":       mesoregion,
			},
			Value:  total,
			Period: periodLbl,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: care rows: %v", ErrIndicatorQueryFailed, err)
	}

	suppressed, err := s.countSuppressedCare(ctx, startYM, endYM, params)
	if err != nil {
		suppressed = 0
	}

	return &IndicatorResult{Rows: result, Suppressed: suppressed}, nil
}

func (s *IndicatorStore) countSuppressedCare(ctx context.Context, startYM, endYM string, params IndicatorParams) (int, error) {
	periodExpr := periodGroupExpr(params.Granularity)

	query := `
		SELECT COUNT(*) FROM (
			SELECT 1
			FROM fact_appointment fa
			JOIN dim_geography g ON fa.geography_id = g.id
			JOIN dim_period p ON fa.period_id = p.id
			WHERE p.year_month BETWEEN $1 AND $2
	`

	args := []any{startYM, endYM}
	argIdx := 3

	if params.Mesoregion != "" {
		query += fmt.Sprintf(" AND g.mesoregion_name = $%d", argIdx)
		args = append(args, params.Mesoregion)
		argIdx++
	}

	query += fmt.Sprintf(`
			GROUP BY fa.appointment_type, g.mesoregion_name, %s
			HAVING SUM(fa.count) < %d
		) suppressed
	`, periodExpr, domain.KThreshold)

	_ = argIdx // suppress unused lint

	var count int
	if err := s.pool.QueryRow(ctx, query, args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("%w: suppressed count: %v", ErrIndicatorQueryFailed, err)
	}
	return count, nil
}
