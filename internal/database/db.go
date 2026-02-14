package database

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

// InitDB initializes the SQLite database connection.
func InitDB(filepath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// RunMigrations applies the database schema.
func RunMigrations(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS vouchers (
		id TEXT PRIMARY KEY,
		code_hash TEXT UNIQUE NOT NULL,
		value INTEGER NOT NULL,
		status TEXT NOT NULL check(status in ('unused', 'used', 'blocked')),
		batch_id TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		redeemed_at DATETIME,
		redeemed_by TEXT
	);
	
	CREATE INDEX IF NOT EXISTS idx_vouchers_status ON vouchers(status);
	`

	_, err := db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}
