package migrations

import (
	"database/sql"
	_ "embed"
	"fmt"
)

//go:embed 000001_init_schema.up.sql
var initSchemaUp string

//go:embed 000001_init_schema.down.sql
var initSchemaDown string

// Up applies all pending database migrations in order.
func Up(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin migration tx: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`); err != nil {
		return fmt.Errorf("create schema_migrations table: %w", err)
	}

	var applied bool
	err = tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = 1)`).Scan(&applied)
	if err != nil {
		return fmt.Errorf("check migration version 1: %w", err)
	}

	if !applied {
		if _, err := tx.Exec(initSchemaUp); err != nil {
			return fmt.Errorf("execute initSchemaUp: %w", err)
		}
		if _, err := tx.Exec(`INSERT INTO schema_migrations (version) VALUES (1)`); err != nil {
			return fmt.Errorf("record migration version 1: %w", err)
		}
	}

	return tx.Commit()
}

// Down rolls back all applied migrations.
func Down(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin migration tx: %w", err)
	}
	defer tx.Rollback()

	var applied bool
	err = tx.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_name = 'schema_migrations'
		) AND EXISTS (
			SELECT 1 FROM schema_migrations WHERE version = 1
		)
	`).Scan(&applied)
	if err != nil {
		// table might not exist, proceed
		applied = false
	}

	if applied {
		if _, err := tx.Exec(initSchemaDown); err != nil {
			return fmt.Errorf("execute initSchemaDown: %w", err)
		}
		if _, err := tx.Exec(`DELETE FROM schema_migrations WHERE version = 1`); err != nil {
			return fmt.Errorf("delete migration version 1: %w", err)
		}
	}

	return tx.Commit()
}
