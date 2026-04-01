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
	pattern := filepath.Join(migrationsDir, "*.sql")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to glob migration files: %w", err)
	}

	sort.Strings(files) // ensures 001 → 002 → 003 order

	for _, file := range files {
		if err := runMigrationFile(db, file); err != nil {
			return fmt.Errorf("migration %s failed: %w", filepath.Base(file), err)
		}
		log.Printf("✅ Migration applied: %s", filepath.Base(file))
	}

	return nil
}

// runMigrationFile reads a single SQL file and executes it inside a transaction.
func runMigrationFile(db *DB, path string) error {
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

	return tx.Commit()
}
