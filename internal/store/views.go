package store

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// MaterializedViews is the single source of truth for view names.
// Used by both schema.go (migration) and RefreshMaterializedViews.
var MaterializedViews = []string{
	"mv_demographics",
	"mv_epidemiological",
}

// RefreshMaterializedViews concurrently refreshes all materialized views.
// Should be called after carry-forward or periodically.
func RefreshMaterializedViews(ctx context.Context, pool *pgxpool.Pool) error {
	for _, v := range MaterializedViews {
		_, err := pool.Exec(ctx, fmt.Sprintf("REFRESH MATERIALIZED VIEW CONCURRENTLY %s", v))
		if err != nil {
			return fmt.Errorf("refresh %s: %w", v, err)
		}
	}
	return nil
}
