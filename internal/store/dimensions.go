package store

import (
	"context"
	"fmt"

	"github.com/acdgbrasil/svc-analysis-bi/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AllowedGenericDimensions returns the dimension table names that
// GenericDimensionStore.GetOrCreate accepts. Returns a fresh slice
// to prevent mutation of the allowlist.
func AllowedGenericDimensions() []string {
	return []string{
		"dim_housing_type",
		"dim_education_level",
		"dim_benefit_type",
		"dim_referral_destination",
		"dim_violation_type",
	}
}

// DimensionStores aggregates all dimension store interfaces.
type DimensionStores interface {
	GeographyStore
	AgeBandStore
	DiagnosisStore
	SexStore
	PeriodStore
	GenericDimensionStore
}

type GeographyStore interface {
	UpsertGeography(ctx context.Context, geo domain.Geography) (int, error)
}

type AgeBandStore interface {
	GetOrCreateAgeBand(ctx context.Context, band domain.AgeBand) (int, error)
}

type DiagnosisStore interface {
	GetOrCreateDiagnosis(ctx context.Context, icdCode, icdLabel, chapter, block string) (int, error)
}

type SexStore interface {
	GetOrCreateSex(ctx context.Context, label string) (int, error)
}

type PeriodStore interface {
	GetOrCreatePeriod(ctx context.Context, period domain.Period) (int, error)
}

type GenericDimensionStore interface {
	GetOrCreate(ctx context.Context, table string, label string) (int, error)
}

// PgDimensionStore implements all dimension store interfaces using pgx.
type PgDimensionStore struct {
	pool *pgxpool.Pool
}

// NewDimensionStore creates a PgDimensionStore backed by the given pool.
func NewDimensionStore(pool *pgxpool.Pool) *PgDimensionStore {
	return &PgDimensionStore{pool: pool}
}

func (s *PgDimensionStore) UpsertGeography(ctx context.Context, geo domain.Geography) (int, error) {
	var id int
	err := s.pool.QueryRow(ctx, `
		INSERT INTO dim_geography (mesoregion_code, mesoregion_name, microregion_code, microregion_name, state_code, region)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (mesoregion_code, microregion_code) DO UPDATE SET mesoregion_name = EXCLUDED.mesoregion_name
		RETURNING id
	`, string(geo.MesoregionCode), geo.MesoregionName, geo.MicroregionCode, geo.MicroregionName, geo.StateCode, geo.Region).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("%w: geography: %v", ErrDimensionInsertFailed, err)
	}
	return id, nil
}

func (s *PgDimensionStore) GetOrCreateAgeBand(ctx context.Context, band domain.AgeBand) (int, error) {
	var id int
	err := s.pool.QueryRow(ctx, `
		INSERT INTO dim_age_band (band_label, min_age, max_age)
		VALUES ($1, $2, $3)
		ON CONFLICT (band_label) DO UPDATE SET band_label = EXCLUDED.band_label
		RETURNING id
	`, band.Label, band.MinAge, band.MaxAge).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("%w: age_band: %v", ErrDimensionInsertFailed, err)
	}
	return id, nil
}

func (s *PgDimensionStore) GetOrCreateDiagnosis(ctx context.Context, icdCode, icdLabel, chapter, block string) (int, error) {
	var id int
	err := s.pool.QueryRow(ctx, `
		INSERT INTO dim_diagnosis (icd_code, icd_label, chapter, block)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (icd_code) DO UPDATE SET icd_label = EXCLUDED.icd_label
		RETURNING id
	`, icdCode, icdLabel, chapter, block).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("%w: diagnosis: %v", ErrDimensionInsertFailed, err)
	}
	return id, nil
}

func (s *PgDimensionStore) GetOrCreateSex(ctx context.Context, label string) (int, error) {
	var id int
	err := s.pool.QueryRow(ctx, `
		INSERT INTO dim_sex (label) VALUES ($1)
		ON CONFLICT (label) DO UPDATE SET label = EXCLUDED.label
		RETURNING id
	`, label).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("%w: sex: %v", ErrDimensionInsertFailed, err)
	}
	return id, nil
}

func (s *PgDimensionStore) GetOrCreatePeriod(ctx context.Context, period domain.Period) (int, error) {
	var id int
	err := s.pool.QueryRow(ctx, `
		INSERT INTO dim_period (year, month, year_month, quarter)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (year_month) DO UPDATE SET year_month = EXCLUDED.year_month
		RETURNING id
	`, period.Year, period.Month, period.YearMonth(), period.Quarter()).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("%w: period: %v", ErrDimensionInsertFailed, err)
	}
	return id, nil
}

func (s *PgDimensionStore) GetOrCreate(ctx context.Context, table string, label string) (int, error) {
	if !isAllowedTable(table) {
		return 0, fmt.Errorf("%w: table %q is not in the allowlist", ErrDimensionInsertFailed, table)
	}

	var id int
	query := fmt.Sprintf(`
		INSERT INTO %s (label) VALUES ($1)
		ON CONFLICT (label) DO UPDATE SET label = EXCLUDED.label
		RETURNING id
	`, table)
	err := s.pool.QueryRow(ctx, query, label).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("%w: %s: %v", ErrDimensionInsertFailed, table, err)
	}
	return id, nil
}

func isAllowedTable(table string) bool {
	for _, t := range AllowedGenericDimensions() {
		if t == table {
			return true
		}
	}
	return false
}
