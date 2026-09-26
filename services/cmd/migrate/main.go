package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/jackc/pgx/v5"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	connStr := os.Getenv("DATABASE_URL_UNPOOLED")
	if connStr == "" {
		connStr = os.Getenv("DATABASE_URL")
	}
	if connStr == "" {
		logger.Error("Neither DATABASE_URL_UNPOOLED nor DATABASE_URL is set")
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		logger.Error("Failed to connect to PostgreSQL", "error", err)
		os.Exit(1)
	}
	defer conn.Close(ctx)

	logger.Info("Connected to PostgreSQL successfully", "host", conn.Config().Host, "db", conn.Config().Database)

	// Create schema_migrations tracker table if not exists
	_, err = conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`)
	if err != nil {
		logger.Error("Failed to initialize schema_migrations table", "error", err)
		os.Exit(1)
	}

	// If participants already exists, backfill 000001 and 000002 into schema_migrations
	var participantsExists bool
	_ = conn.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'participants');`).Scan(&participantsExists)
	if participantsExists {
		_, _ = conn.Exec(ctx, `
			INSERT INTO schema_migrations (version) VALUES 
			('migrations/000001_init_schema.up.sql'),
			('migrations/000002_seed_initial_grade.up.sql')
			ON CONFLICT (version) DO NOTHING;
		`)
	}

	// Migration files in sequence
	migrationFiles := []string{
		"migrations/000001_init_schema.up.sql",
		"migrations/000002_seed_initial_grade.up.sql",
		"migrations/000003_index_and_surveillance.up.sql",
	}

	// Try relative or standard root path
	for _, relPath := range migrationFiles {
		var alreadyApplied bool
		_ = conn.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = $1)", relPath).Scan(&alreadyApplied)
		if alreadyApplied {
			logger.Info("Migration already applied; skipping", "version", relPath)
			continue
		}

		fullPath := relPath
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			fullPath = filepath.Join("services", relPath)
		}
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			logger.Error("Migration file not found", "path", relPath)
			os.Exit(1)
		}

		sqlBytes, err := os.ReadFile(fullPath)
		if err != nil {
			logger.Error("Failed to read migration file", "file", fullPath, "error", err)
			os.Exit(1)
		}

		logger.Info("Executing migration", "file", fullPath, "bytes", len(sqlBytes))
		if _, err := conn.Exec(ctx, string(sqlBytes)); err != nil {
			logger.Error("Migration execution failed", "file", fullPath, "error", err)
			os.Exit(1)
		}

		_, err = conn.Exec(ctx, "INSERT INTO schema_migrations (version) VALUES ($1)", relPath)
		if err != nil {
			logger.Error("Failed to record migration in schema_migrations", "file", relPath, "error", err)
			os.Exit(1)
		}
		logger.Info("Migration completed successfully", "file", fullPath)
	}

	// Verify all tables created
	rows, err := conn.Query(ctx, `
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		ORDER BY table_name;
	`)
	if err != nil {
		logger.Error("Failed to inspect schema tables", "error", err)
		os.Exit(1)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tbl string
		if err := rows.Scan(&tbl); err != nil {
			logger.Error("Failed scanning table name", "error", err)
			os.Exit(1)
		}
		tables = append(tables, tbl)
	}

	logger.Info("Schema verification complete", "public_tables_count", len(tables), "tables", tables)

	// Verify benchmark grade exists
	var count int
	err = conn.QueryRow(ctx, "SELECT COUNT(*) FROM grades WHERE id = 'H100-SXM-8XNV'").Scan(&count)
	if err != nil {
		logger.Error("Failed querying grades table", "error", err)
		os.Exit(1)
	}

	if count > 0 {
		logger.Info("Benchmark grade seed verified", "grade_id", "H100-SXM-8XNV", "count", count)
		fmt.Println("SUCCESS: Database migrations and seed verification complete!")
	} else {
		logger.Error("Benchmark grade missing from database")
		os.Exit(1)
	}
}
