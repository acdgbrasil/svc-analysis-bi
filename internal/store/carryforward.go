package store

import (
	"context"
	"fmt"

	"github.com/acdgbrasil/svc-analysis-bi/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CarryForwardJob copies active patient snapshots from the previous period
// to a new period, ensuring time-series continuity. Without this, patients
// "disappear" from indicators in months where no new event is received.
type CarryForwardJob struct {
	pool *pgxpool.Pool
	dims *PgDimensionStore
}

// NewCarryForwardJob creates a CarryForwardJob backed by the given pool.
func NewCarryForwardJob(pool *pgxpool.Pool) *CarryForwardJob {
	return &CarryForwardJob{
		pool: pool,
		dims: NewDimensionStore(pool),
	}
}

// Run carries forward all patient snapshots from prevPeriod to newPeriod.
// It skips patients that already have a snapshot in the new period (idempotent).
// Returns the number of rows carried forward.
func (j *CarryForwardJob) Run(ctx context.Context, newPeriod domain.Period) (int, error) {
	// 1. Ensure new period dimension exists
	newPeriodID, err := j.dims.GetOrCreatePeriod(ctx, newPeriod)
	if err != nil {
		return 0, fmt.Errorf("carryforward: failed to create period: %w", err)
	}

	// 2. Get previous period
	prevPeriod := previousPeriod(newPeriod)
	// Previous period must already exist — if not, nothing to carry forward

	// 3. Copy snapshots that don't already exist in the new period
	tag, err := j.pool.Exec(ctx, `
		INSERT INTO fact_patient_snapshot (
			period_id, age_band_id, sex_id, geography_id,
			housing_type_id, education_level_id, income_band,
			receives_benefit, has_deficiency, food_insecurity,
			is_overcrowded, family_size, assessment_completeness,
			patient_hash
		)
		SELECT
			$1, age_band_id, sex_id, geography_id,
			housing_type_id, education_level_id, income_band,
			receives_benefit, has_deficiency, food_insecurity,
			is_overcrowded, family_size, assessment_completeness,
			patient_hash
		FROM fact_patient_snapshot
		WHERE period_id = (
			SELECT id FROM dim_period WHERE year = $2 AND month = $3
		)
		ON CONFLICT (period_id, patient_hash) DO NOTHING
	`, newPeriodID, prevPeriod.Year, prevPeriod.Month)
	if err != nil {
		return 0, fmt.Errorf("carryforward: %w", err)
	}

	return int(tag.RowsAffected()), nil
}

// previousPeriod calculates the period immediately before the given period.
// January wraps to December of the previous year.
func previousPeriod(p domain.Period) domain.Period {
	if p.Month == 1 {
		return domain.Period{Year: p.Year - 1, Month: 12}
	}
	return domain.Period{Year: p.Year, Month: p.Month - 1}
}
