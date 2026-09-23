package database

import (
	"fmt"
	"log"
	"time"

	"poltekkes-cat-backend/internal/config"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func NewPostgresConnection(cfg *config.Config) *sqlx.DB {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPass, cfg.DBName, cfg.DBSSLMode,
	)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatalf("❌ Failed to connect to PostgreSQL (database: %s): %v", cfg.DBName, err)
	}

	// Enterprise Connection Pooling for High Concurrency
	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(50)
	db.SetConnMaxLifetime(15 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		log.Fatalf("❌ PostgreSQL Ping failed: %v", err)
	}

	log.Printf(" PostgreSQL Database Connection Pool (%s@%s:%s/%s) initialized successfully", cfg.DBUser, cfg.DBHost, cfg.DBPort, cfg.DBName)

	// Automatic Schema Migrations Execution (only on CBT Database)
	if err := RunMigrations(db); err != nil {
		log.Fatalf("❌ Database migration failed: %v", err)
	}

	return db
}

// NewSiakadConnection initializes a dedicated connection pool to the SIAKAD database (gate, ref, pendaftaran)
// If dual database is not active, it returns primaryDB directly for seamless single-database operation.
func NewSiakadConnection(cfg *config.Config, primaryDB *sqlx.DB) *sqlx.DB {
	if !cfg.IsDualDB() {
		log.Printf("ℹ️ Single Database Mode: SIAKAD operations will share primary CBT connection pool (%s)", cfg.DBName)
		return primaryDB
	}

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBSiakadHost, cfg.DBSiakadPort, cfg.DBSiakadUser, cfg.DBSiakadPass, cfg.DBSiakadName, cfg.DBSiakadSSLMode,
	)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatalf("❌ Failed to connect to SIAKAD PostgreSQL (database: %s): %v", cfg.DBSiakadName, err)
	}

	// Enterprise Connection Pooling for SIAKAD
	db.SetMaxOpenConns(50)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(15 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		log.Fatalf("❌ SIAKAD PostgreSQL Ping failed: %v", err)
	}

	log.Printf("✅ PostgreSQL SIAKAD Database Connection Pool (%s@%s:%s/%s) initialized successfully", cfg.DBSiakadUser, cfg.DBSiakadHost, cfg.DBSiakadPort, cfg.DBSiakadName)
	return db
}

