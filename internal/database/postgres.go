package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"

	_ "github.com/lib/pq"
)

// DB wraps sql.DB to allow attaching helper methods if needed later.
type DB struct {
	*sql.DB
}

// Connect opens a PostgreSQL connection using the provided DSN and verifies
// it with a ping. Returns a wrapped *DB or an error.
func Connect(dsn string) (*DB, error) {
	sqlDB, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to reach database: %w", err)
	}

	// Conservative pool settings suitable for a small dashboard service.
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)

	log.Println("✅ Database connected")
	return &DB{sqlDB}, nil
}

// RunMigrations executes all .sql files found in the given directory in
// lexicographic order (001_, 002_, …). Each file is run inside a transaction
// so a failing migration rolls back cleanly without leaving partial state.
func RunMigrations(db *DB, migrationsDir string) error {
	if err := ensureMigrationsTable(db); err != nil {
		return fmt.Errorf("failed to ensure schema_migrations table: %w", err)
	}

	pattern := filepath.Join(migrationsDir, "*.sql")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to glob migration files: %w", err)
	}

	sort.Strings(files) // ensures 001 → 002 → 003 order

	for _, file := range files {
		migrationName := filepath.Base(file)

		applied, err := isMigrationApplied(db, migrationName)
		if err != nil {
			return fmt.Errorf("failed to check migration %s: %w", migrationName, err)
		}
		if applied {
			log.Printf("↪️  Migration already applied, skipping: %s", migrationName)
			continue
		}

		if err := runMigrationFile(db, file, migrationName); err != nil {
			return fmt.Errorf("migration %s failed: %w", migrationName, err)
		}
		log.Printf("✅ Migration applied: %s", migrationName)
	}

	return nil
}

// ensureMigrationsTable creates the migration history table if it does not exist.
func ensureMigrationsTable(db *DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			name       VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
		)`

	if _, err := db.Exec(query); err != nil {
		return fmt.Errorf("cannot create schema_migrations table: %w", err)
	}

	return nil
}

// isMigrationApplied checks whether a migration file has already been applied.
func isMigrationApplied(db *DB, migrationName string) (bool, error) {
	const query = `SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE name = $1)`

	var exists bool
	if err := db.QueryRow(query, migrationName).Scan(&exists); err != nil {
		return false, fmt.Errorf("cannot query schema_migrations: %w", err)
	}

	return exists, nil
}

// runMigrationFile reads a single SQL file and executes it inside a transaction.
// The migration is recorded only if execution succeeds.
func runMigrationFile(db *DB, path, migrationName string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("cannot read file: %w", err)
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("cannot begin transaction: %w", err)
	}

	if _, err := tx.Exec(string(content)); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("cannot execute SQL: %w", err)
	}

	if _, err := tx.Exec(`INSERT INTO schema_migrations (name) VALUES ($1)`, migrationName); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("cannot record migration: %w", err)
	}

	return tx.Commit()
}
