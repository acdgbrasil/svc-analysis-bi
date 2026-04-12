package store

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// RefreshMaterializedViews concurrently refreshes all materialized views.
// Should be called after carry-forward or periodically.
func RefreshMaterializedViews(ctx context.Context, pool *pgxpool.Pool) error {
	views := []string{"mv_demographics", "mv_epidemiological"}
	for _, v := range views {
		_, err := pool.Exec(ctx, fmt.Sprintf("REFRESH MATERIALIZED VIEW CONCURRENTLY %s", v))
		if err != nil {
			return fmt.Errorf("refresh %s: %w", v, err)
		}
	}
	return nil
}
