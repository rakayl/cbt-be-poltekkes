package database

import (
	"context"
	"embed"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

// RunMigrations automatically checks and executes pending SQL migrations in order
func RunMigrations(db *sqlx.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	log.Println("🔄 Checking and executing database schema migrations...")

	// 1. Ensure schema and migration tracking table exists
	initQuery := `
		CREATE SCHEMA IF NOT EXISTS cat;
		CREATE TABLE IF NOT EXISTS cat.schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
	`
	if _, err := db.ExecContext(ctx, initQuery); err != nil {
		return fmt.Errorf("failed to initialize schema_migrations table: %w", err)
	}

	// 2. Read embedded migration files
	entries, err := migrationFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			files = append(files, entry.Name())
		}
	}
	sort.Strings(files)

	// 3. Process each migration in transaction
	for _, file := range files {
		var exists int
		err := db.GetContext(ctx, &exists, "SELECT COUNT(*) FROM cat.schema_migrations WHERE version = $1", file)
		if err != nil {
			return fmt.Errorf("failed to check migration status for %s: %w", file, err)
		}

		if exists > 0 {
			// Already applied
			continue
		}

		content, err := migrationFS.ReadFile("migrations/" + file)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file, err)
		}

		log.Printf("⏳ Applying migration: %s ...", file)

		tx, err := db.BeginTxx(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to begin transaction for %s: %w", file, err)
		}

		if _, err := tx.ExecContext(ctx, string(content)); err != nil {
			tx.Rollback()
			return fmt.Errorf("migration %s failed to execute: %w", file, err)
		}

		if _, err := tx.ExecContext(ctx, "INSERT INTO cat.schema_migrations (version) VALUES ($1)", file); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %s in tracking table: %w", file, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", file, err)
		}

		log.Printf("✅ Migration %s applied successfully", file)
	}

	log.Println(" Database schema is up-to-date")
	return nil
}
