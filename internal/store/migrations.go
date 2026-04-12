package store

import (
	"context"
	"fmt"
	"sort"
)

// Migration represents a single forward-only database migration.
type Migration struct {
	Version int
	Name    string
	SQL     string
}

// validateMigrations checks for duplicate versions and non-ascending order.
func validateMigrations(migrations []Migration) error {
	if len(migrations) == 0 {
		return nil
	}

	seen := make(map[int]bool, len(migrations))
	for _, m := range migrations {
		if seen[m.Version] {
			return ErrMigrationDuplicate
		}
		seen[m.Version] = true
	}

	for i := 1; i < len(migrations); i++ {
		if migrations[i].Version <= migrations[i-1].Version {
			return fmt.Errorf("%w: versions must be strictly ascending", ErrMigrationFailed)
		}
	}

	return nil
}

// RunMigrations applies all pending migrations in version order.
// Creates the schema_migrations tracking table if it does not exist.
func RunMigrations(ctx context.Context, db *DB, migrations []Migration) error {
	if err := validateMigrations(migrations); err != nil {
		return err
	}

	if len(migrations) == 0 {
		return nil
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	pool := db.Pool()

	// Create tracking table
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INT PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("%w: failed to create schema_migrations table: %v", ErrMigrationFailed, err)
	}

	for _, m := range migrations {
		var exists bool
		err := pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)", m.Version).Scan(&exists)
		if err != nil {
			return fmt.Errorf("%w: version check failed for v%d: %v", ErrMigrationFailed, m.Version, err)
		}
		if exists {
			continue
		}

		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("%w: failed to begin tx for v%d: %v", ErrMigrationFailed, m.Version, err)
		}

		// Use simple protocol for multi-statement SQL (pgx extended protocol
		// does not support multiple statements in a single Exec call).
		if _, err := tx.Conn().Exec(ctx, m.SQL); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("%w: v%d (%s) failed: %v", ErrMigrationFailed, m.Version, m.Name, err)
		}

		if _, err := tx.Exec(ctx, "INSERT INTO schema_migrations (version, name) VALUES ($1, $2)", m.Version, m.Name); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("%w: failed to record v%d: %v", ErrMigrationFailed, m.Version, err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("%w: commit failed for v%d: %v", ErrMigrationFailed, m.Version, err)
		}
	}

	return nil
}
