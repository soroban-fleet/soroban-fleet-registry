package ingest

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// PostgresCheckpointStore stores indexer stream checkpoints in PostgreSQL.
type PostgresCheckpointStore struct {
	db *sql.DB
}

// NewPostgresCheckpointStore creates a new PostgresCheckpointStore.
func NewPostgresCheckpointStore(db *sql.DB) *PostgresCheckpointStore {
	return &PostgresCheckpointStore{db: db}
}

// GetCheckpoint retrieves the last processed ledger sequence for a stream.
// Returns 0 and nil error if no checkpoint exists yet.
func (s *PostgresCheckpointStore) GetCheckpoint(ctx context.Context, stream string) (uint32, error) {
	query := `
		SELECT ledger
		FROM indexer_checkpoints
		WHERE stream = $1
	`
	var ledger int64
	err := s.db.QueryRowContext(ctx, query, stream).Scan(&ledger)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, fmt.Errorf("get checkpoint for stream %q: %w", stream, err)
	}

	return uint32(ledger), nil
}

// SetCheckpointTx updates the checkpoint sequence within an active database transaction.
func (s *PostgresCheckpointStore) SetCheckpointTx(ctx context.Context, tx *sql.Tx, stream string, ledger uint32) error {
	now := time.Now().UTC()
	query := `
		INSERT INTO indexer_checkpoints (stream, ledger, updated_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (stream) DO UPDATE SET
			ledger = EXCLUDED.ledger,
			updated_at = EXCLUDED.updated_at
	`
	_, err := tx.ExecContext(ctx, query, stream, ledger, now)
	if err != nil {
		return fmt.Errorf("set checkpoint tx for stream %q to %d: %w", stream, ledger, err)
	}
	return nil
}
