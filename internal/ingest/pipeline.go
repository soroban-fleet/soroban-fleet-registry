package ingest

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"log/slog"
	"time"

	"github.com/soroban-fleet/soroban-fleet-registry/internal/rpc"
	"github.com/stellar/go-stellar-sdk/xdr"
)

const (
	// DefaultStreamName is the identifier for the main ledger indexer checkpoint.
	DefaultStreamName = "main"
)

// CheckpointStore persists and retrieves ledger ingestion checkpoints.
type CheckpointStore interface {
	GetCheckpoint(ctx context.Context, stream string) (uint32, error)
	SetCheckpointTx(ctx context.Context, tx *sql.Tx, stream string, ledger uint32) error
}

// PipelineOptions configures pipeline execution.
type PipelineOptions struct {
	StreamName  string
	StartLedger uint32
	BatchSize   uint32
	MaxLedgers  uint32 // 0 for unbounded
	PollDelay   time.Duration
}

// Pipeline orchestrates reading ledgers, applying changes, and advancing checkpoints.
type Pipeline struct {
	db          *sql.DB
	rpcClient   rpc.Client
	processor   *LedgerProcessor
	checkpoints CheckpointStore
	logger      *slog.Logger
}

// NewPipeline creates a new ingestion Pipeline.
func NewPipeline(
	db *sql.DB,
	rpcClient rpc.Client,
	processor *LedgerProcessor,
	checkpoints CheckpointStore,
	logger *slog.Logger,
) *Pipeline {
	if logger == nil {
		logger = slog.Default()
	}
	return &Pipeline{
		db:          db,
		rpcClient:   rpcClient,
		processor:   processor,
		checkpoints: checkpoints,
		logger:      logger,
	}
}

// ProcessSingleLedger decodes and transactionally processes a single ledger.
func (p *Pipeline) ProcessSingleLedger(
	ctx context.Context,
	meta xdr.LedgerCloseMeta,
	stream string,
) (LedgerProcessingStats, error) {
	changes, err := ExtractChangesFromLedgerCloseMeta(meta)
	if err != nil {
		return LedgerProcessingStats{}, fmt.Errorf("extract changes from ledger: %w", err)
	}

	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return LedgerProcessingStats{}, fmt.Errorf("begin ledger tx: %w", err)
	}
	defer tx.Rollback()

	stats, err := p.processor.ProcessLedgerTx(ctx, tx, changes)
	if err != nil {
		return stats, fmt.Errorf("process ledger %d: %w", changes.Sequence, err)
	}

	if p.checkpoints != nil && stream != "" {
		if err := p.checkpoints.SetCheckpointTx(ctx, tx, stream, changes.Sequence); err != nil {
			return stats, fmt.Errorf("set checkpoint %d for %s: %w", changes.Sequence, stream, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return stats, fmt.Errorf("commit ledger %d: %w", changes.Sequence, err)
	}

	p.logger.Info("ledger_processed",
		"ledger", stats.Ledger,
		"duration_ms", stats.Duration.Milliseconds(),
		"changes", stats.ChangesProcessed,
		"fleets_updated", stats.FleetsUpdated,
		"members_updated", stats.MembersUpdated,
		"releases", stats.ReleasesDetected,
	)

	return stats, nil
}

// Run executes the ingestion loop until ctx is cancelled or maxLedgers is reached.
func (p *Pipeline) Run(ctx context.Context, opts PipelineOptions) error {
	stream := opts.StreamName
	if stream == "" {
		stream = DefaultStreamName
	}

	batchSize := opts.BatchSize
	if batchSize == 0 {
		batchSize = 50
	}

	pollDelay := opts.PollDelay
	if pollDelay == 0 {
		pollDelay = time.Second
	}

	startLedger := opts.StartLedger
	if p.checkpoints != nil {
		cp, err := p.checkpoints.GetCheckpoint(ctx, stream)
		if err == nil && cp > 0 {
			startLedger = cp + 1
		}
	}

	var processedCount uint32

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if opts.MaxLedgers > 0 && processedCount >= opts.MaxLedgers {
			return nil
		}

		// Fetch ledgers from RPC
		resp, err := p.rpcClient.GetLedgers(ctx, startLedger, batchSize)
		if err != nil {
			p.logger.Warn("fetch ledgers error, will retry", "startLedger", startLedger, "error", err)
			time.Sleep(pollDelay)
			continue
		}

		if len(resp.Ledgers) == 0 {
			time.Sleep(pollDelay)
			continue
		}

		for _, info := range resp.Ledgers {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			if opts.MaxLedgers > 0 && processedCount >= opts.MaxLedgers {
				return nil
			}

			metaRaw, err := base64.StdEncoding.DecodeString(info.LedgerMetadata)
			if err != nil {
				return fmt.Errorf("decode ledger metadata for %d: %w", info.Sequence, err)
			}

			var meta xdr.LedgerCloseMeta
			if err := meta.UnmarshalBinary(metaRaw); err != nil {
				return fmt.Errorf("unmarshal ledger metadata for %d: %w", info.Sequence, err)
			}

			_, err = p.ProcessSingleLedger(ctx, meta, stream)
			if err != nil {
				return fmt.Errorf("process ledger %d: %w", info.Sequence, err)
			}

			startLedger = info.Sequence + 1
			processedCount++
		}
	}
}
