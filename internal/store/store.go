package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Option configures the database connection pool.
type Option func(*pgxpool.Config)

// WithIdleTimeout sets the maximum idle time for connections in the pool.
func WithIdleTimeout(d time.Duration) Option {
	return func(cfg *pgxpool.Config) {
		cfg.MaxConnIdleTime = d
	}
}

var (
	ErrConnectionFailed      = errors.New("store: failed to connect to database")
	ErrPingFailed            = errors.New("store: database ping failed")
	ErrMigrationFailed       = errors.New("store: migration failed")
	ErrMigrationDuplicate    = errors.New("store: duplicate migration version")
	ErrDimensionInsertFailed = errors.New("store: dimension insert failed")
	ErrDimensionNotFound     = errors.New("store: dimension not found")
	ErrEventAlreadyProcessed = errors.New("store: event already processed")
	ErrDLQInsertFailed       = errors.New("store: failed to insert into dead-letter queue")
	ErrEventCheckFailed      = errors.New("store: failed to check event processing status")
	ErrEventMarkFailed       = errors.New("store: failed to mark event as processed")
)

// DB wraps a pgx connection pool.
type DB struct {
	pool *pgxpool.Pool
}

// New creates a new DB by establishing a pgx connection pool.
func New(ctx context.Context, dsn string, maxConns int, opts ...Option) (*DB, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}
	cfg.MaxConns = int32(maxConns)
	for _, opt := range opts {
		opt(cfg)
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("%w: %v", ErrPingFailed, err)
	}

	return &DB{pool: pool}, nil
}

// Close releases all connections in the pool.
func (db *DB) Close() {
	if db.pool != nil {
		db.pool.Close()
	}
}

// Pool returns the underlying *pgxpool.Pool.
func (db *DB) Pool() *pgxpool.Pool {
	return db.pool
}

// Ping verifies database connectivity.
func (db *DB) Ping(ctx context.Context) error {
	if err := db.pool.Ping(ctx); err != nil {
		return fmt.Errorf("%w: %v", ErrPingFailed, err)
	}
	return nil
}
