package db

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Pool wraps the pgxpool.Pool with institutional observability and transaction helpers.
type Pool struct {
	*pgxpool.Pool
	logger *slog.Logger
}

// Config holds pool connection parameters following golang-database best practices.
type Config struct {
	ConnString      string
	MaxOpenConns    int32
	MinIdleConns    int32
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
}

// DefaultConfig retrieves the database connection string from environment variables.
func DefaultConfig() Config {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = os.Getenv("POSTGRES_URL")
	}

	return Config{
		ConnString:      connStr,
		MaxOpenConns:    20,
		MinIdleConns:    4,
		MaxConnLifetime: 30 * time.Minute,
		MaxConnIdleTime: 5 * time.Minute,
	}
}

// Connect initializes and validates a PostgreSQL connection pool.
func Connect(ctx context.Context, cfg Config, logger *slog.Logger) (*Pool, error) {
	if cfg.ConnString == "" {
		return nil, fmt.Errorf("database connection string is empty (set DATABASE_URL)")
	}

	poolCfg, err := pgxpool.ParseConfig(cfg.ConnString)
	if err != nil {
		return nil, fmt.Errorf("parsing pgxpool config: %w", err)
	}

	// Apply production pool settings
	poolCfg.MaxConns = cfg.MaxOpenConns
	poolCfg.MinConns = cfg.MinIdleConns
	poolCfg.MaxConnLifetime = cfg.MaxConnLifetime
	poolCfg.MaxConnIdleTime = cfg.MaxConnIdleTime

	// Timeout for acquiring connections from pool
	poolCfg.HealthCheckPeriod = 1 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("creating pgx pool: %w", err)
	}

	// Verify liveness immediately
	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pinging database at %s: %w", poolCfg.ConnConfig.Host, err)
	}

	if logger == nil {
		logger = slog.Default()
	}

	logger.Info("PostgreSQL connection pool initialized",
		"host", poolCfg.ConnConfig.Host,
		"database", poolCfg.ConnConfig.Database,
		"max_conns", poolCfg.MaxConns,
		"min_conns", poolCfg.MinConns,
	)

	return &Pool{
		Pool:   pool,
		logger: logger,
	}, nil
}

// WithTransaction runs fn inside a database transaction with automatic rollback on error.
func (p *Pool) WithTransaction(ctx context.Context, fn func(tx pgx.Tx) error) error {
	tx, err := p.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}

	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback(ctx)
			panic(r)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil && rbErr != pgx.ErrTxClosed {
			p.logger.Error("failed rolling back transaction", "error", rbErr, "cause", err)
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	return nil
}
