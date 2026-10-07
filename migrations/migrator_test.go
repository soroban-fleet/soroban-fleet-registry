package migrations

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

func getTestDB(t *testing.T) *sql.DB {
	dbURL := os.Getenv("SFR_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres@localhost:5433/sfr_test?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Skipf("skipping db test, could not open connection: %v", err)
	}

	if err := db.Ping(); err != nil {
		t.Skipf("skipping db test, database not reachable: %v", err)
	}

	return db
}

func TestMigrations_Up_Idempotency_Down(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	// Initial rollback if exists
	_ = Down(db)

	// Apply migrations
	if err := Up(db); err != nil {
		t.Fatalf("failed to apply migrations Up: %v", err)
	}

	// Verify all tables exist
	tables := []string{
		"fleets",
		"fleet_members",
		"fleet_releases",
		"fleet_verifications",
		"indexer_checkpoints",
		"schema_migrations",
	}

	for _, tbl := range tables {
		var exists bool
		err := db.QueryRow(`
			SELECT EXISTS (
				SELECT 1 FROM information_schema.tables 
				WHERE table_name = $1
			)
		`, tbl).Scan(&exists)
		if err != nil {
			t.Fatalf("failed to query table existence for %s: %v", tbl, err)
		}
		if !exists {
			t.Fatalf("expected table %s to exist after migration", tbl)
		}
	}

	// Idempotency: Running Up again must succeed without errors
	if err := Up(db); err != nil {
		t.Fatalf("migration Up is not idempotent: %v", err)
	}

	// Test Down
	if err := Down(db); err != nil {
		t.Fatalf("failed to apply migrations Down: %v", err)
	}

	// Verify tables are dropped
	for _, tbl := range []string{"fleets", "fleet_members", "fleet_releases", "fleet_verifications", "indexer_checkpoints"} {
		var exists bool
		err := db.QueryRow(`
			SELECT EXISTS (
				SELECT 1 FROM information_schema.tables 
				WHERE table_name = $1
			)
		`, tbl).Scan(&exists)
		if err != nil {
			t.Fatalf("failed to check table %s after Down: %v", tbl, err)
		}
		if exists {
			t.Fatalf("expected table %s to be dropped after Down", tbl)
		}
	}

	// Re-apply Up so DB is left in valid state
	if err := Up(db); err != nil {
		t.Fatalf("failed to re-apply migrations: %v", err)
	}
}
