package main

import (
	"context"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/soroban-fleet/soroban-fleet-registry/internal/api"
	"github.com/soroban-fleet/soroban-fleet-registry/internal/cap85"
	"github.com/soroban-fleet/soroban-fleet-registry/internal/config"
	"github.com/soroban-fleet/soroban-fleet-registry/internal/fleet"
	"github.com/soroban-fleet/soroban-fleet-registry/internal/history"
	"github.com/soroban-fleet/soroban-fleet-registry/internal/ingest"
	"github.com/soroban-fleet/soroban-fleet-registry/internal/rpc"
	"github.com/soroban-fleet/soroban-fleet-registry/internal/verification"
	"github.com/soroban-fleet/soroban-fleet-registry/migrations"
)

// Standard CLI Exit Codes
const (
	ExitSuccess         = 0
	ExitDrift           = 1
	ExitBrokenReference = 2
	ExitIncomplete      = 3
	ExitInvalidInput    = 4
	ExitInternalError   = 5
)

func main() {
	code := run(os.Args[1:], os.Stdout, os.Stderr)
	os.Exit(code)
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printUsage(stderr)
		return ExitInvalidInput
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(stderr, "Error loading configuration: %v\n", err)
		return ExitInternalError
	}

	switch args[0] {
	case "fleet":
		return runFleetCommand(args[1:], cfg, stdout, stderr)
	case "contract":
		return runContractCommand(args[1:], cfg, stdout, stderr)
	case "migrate":
		return runMigrateCommand(args[1:], cfg, stdout, stderr)
	case "api":
		return runAPICommand(cfg, stdout, stderr)
	case "ingest":
		return runIngestCommand(cfg, stdout, stderr)
	case "--help", "-h", "help":
		printUsage(stdout)
		return ExitSuccess
	default:
		fmt.Fprintf(stderr, "Unknown command: %s\n", args[0])
		printUsage(stderr)
		return ExitInvalidInput
	}
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, `Usage: sfr <command> [subcommand] [flags]

Commands:
  fleet list                               List all fleets
  fleet inspect --owner C... --tag TAG     Inspect fleet details
  fleet members --owner C... --tag TAG     List fleet members
  fleet history --owner C... --tag TAG     List verification history
  fleet releases --owner C... --tag TAG    List release history
  fleet verify --owner C... --tag TAG --expected-wasm HASH  Verify fleet integrity
  contract inspect --contract C...         Inspect contract and resolved executable
  migrate [up|down]                        Run database schema migrations
  api                                      Start REST API server
  ingest                                   Run historical ingestion pipeline

Global Flags:
  --json                                   Machine-readable JSON output`)
}

func getDB(cfg *config.Config) (*sql.DB, error) {
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("SFR_DATABASE_URL is not configured")
	}
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return db, nil
}

func runFleetCommand(args []string, cfg *config.Config, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "Subcommand required: list, inspect, members, history, releases, verify")
		return ExitInvalidInput
	}

	subcmd := args[0]
	fs := flag.NewFlagSet("fleet "+subcmd, flag.ContinueOnError)
	fs.SetOutput(stderr)

	owner := fs.String("owner", "", "Owner contract StrKey address")
	tag := fs.String("tag", "", "Fleet tag")
	expectedWASM := fs.String("expected-wasm", "", "Expected WASM hash (hex)")
	jsonOutput := fs.Bool("json", false, "Output format JSON")
	limit := fs.Int("limit", 20, "Page limit")
	offset := fs.Int("offset", 0, "Page offset")
	activeOnly := fs.Bool("active-only", true, "Active members only")

	if err := fs.Parse(args[1:]); err != nil {
		return ExitInvalidInput
	}

	db, err := getDB(cfg)
	if err != nil {
		fmt.Fprintf(stderr, "Database error: %v\n", err)
		return ExitInternalError
	}
	defer db.Close()

	fleetRepo := fleet.NewPostgresRepository(db)
	var rpcClient rpc.Client
	var resolver cap85.Resolver
	if cfg.RPCURL != "" {
		rpcClient = rpc.NewClient(cfg.RPCURL)
		resolver = cap85.NewResolver(rpcClient)
	}
	fleetService := fleet.NewService(fleetRepo, resolver)
	checkpoints := ingest.NewPostgresCheckpointStore(db)
	verifier := verification.NewEngine(db, fleetRepo, checkpoints, resolver, rpcClient)
	releaseRepo := history.NewPostgresRepository(db)

	ctx := context.Background()

	switch subcmd {
	case "list":
		fleets, total, err := fleetService.ListFleets(ctx, *limit, *offset)
		if err != nil {
			fmt.Fprintf(stderr, "Error listing fleets: %v\n", err)
			return ExitInternalError
		}
		if *jsonOutput {
			_ = json.NewEncoder(stdout).Encode(map[string]any{"data": fleets, "total": total})
		} else {
			fmt.Fprintf(stdout, "Total Fleets: %d\n", total)
			for _, f := range fleets {
				fmt.Fprintf(stdout, "- %s:%s | Members: %d | WASM: %x | Ledgers: %d..%d\n",
					f.ID.Owner, f.ID.Tag, f.MemberCount, f.CurrentWASMHash, f.FirstSeenLedger, f.LastSeenLedger)
			}
		}
		return ExitSuccess

	case "inspect":
		if *owner == "" || *tag == "" {
			fmt.Fprintln(stderr, "Missing required flags: --owner and --tag")
			return ExitInvalidInput
		}
		f, err := fleetService.GetFleet(ctx, fleet.FleetID{Owner: *owner, Tag: *tag})
		if err != nil {
			fmt.Fprintf(stderr, "Fleet not found: %v\n", err)
			return ExitInvalidInput
		}
		if *jsonOutput {
			_ = json.NewEncoder(stdout).Encode(f)
		} else {
			fmt.Fprintf(stdout, "Fleet: %s:%s\n", f.ID.Owner, f.ID.Tag)
			fmt.Fprintf(stdout, "Current WASM Hash: %x\n", f.CurrentWASMHash)
			fmt.Fprintf(stdout, "Member Count:      %d\n", f.MemberCount)
			fmt.Fprintf(stdout, "First Seen Ledger: %d\n", f.FirstSeenLedger)
			fmt.Fprintf(stdout, "Last Seen Ledger:  %d\n", f.LastSeenLedger)
			fmt.Fprintf(stdout, "Last Indexed:      %d\n", f.LastIndexedLedger)
		}
		return ExitSuccess

	case "members":
		if *owner == "" || *tag == "" {
			fmt.Fprintln(stderr, "Missing required flags: --owner and --tag")
			return ExitInvalidInput
		}
		members, total, err := fleetService.ListMembers(ctx, fleet.FleetID{Owner: *owner, Tag: *tag}, *activeOnly, *limit, *offset)
		if err != nil {
			fmt.Fprintf(stderr, "Error listing members: %v\n", err)
			return ExitInvalidInput
		}
		if *jsonOutput {
			_ = json.NewEncoder(stdout).Encode(map[string]any{"data": members, "total": total})
		} else {
			fmt.Fprintf(stdout, "Members (%d total):\n", total)
			for _, m := range members {
				status := "ACTIVE"
				if !m.Active {
					status = "INACTIVE"
				}
				fmt.Fprintf(stdout, "- %s [%s] | WASM: %x | First: %d\n", m.ContractID, status, m.WASMHash, m.FirstSeenLedger)
			}
		}
		return ExitSuccess

	case "releases":
		if *owner == "" || *tag == "" {
			fmt.Fprintln(stderr, "Missing required flags: --owner and --tag")
			return ExitInvalidInput
		}
		releases, total, err := releaseRepo.ListReleases(ctx, history.FleetID{Owner: *owner, Tag: *tag}, *limit, *offset)
		if err != nil {
			fmt.Fprintf(stderr, "Error listing releases: %v\n", err)
			return ExitInternalError
		}
		if *jsonOutput {
			_ = json.NewEncoder(stdout).Encode(map[string]any{"data": releases, "total": total})
		} else {
			fmt.Fprintf(stdout, "Releases (%d total):\n", total)
			for _, r := range releases {
				fmt.Fprintf(stdout, "- Ledger %d | Tx %s | Old: %x -> New: %x | %s\n",
					r.Ledger, r.TxHash, r.OldWASMHash, r.NewWASMHash, r.ObservedAt.Format("2006-01-02 15:04:05"))
			}
		}
		return ExitSuccess

	case "history":
		if *owner == "" || *tag == "" {
			fmt.Fprintln(stderr, "Missing required flags: --owner and --tag")
			return ExitInvalidInput
		}
		results, total, err := verifier.ListVerifications(ctx, verification.FleetID{Owner: *owner, Tag: *tag}, *limit, *offset)
		if err != nil {
			fmt.Fprintf(stderr, "Error listing verifications: %v\n", err)
			return ExitInternalError
		}
		if *jsonOutput {
			_ = json.NewEncoder(stdout).Encode(map[string]any{"data": results, "total": total})
		} else {
			fmt.Fprintf(stdout, "Verification History (%d total):\n", total)
			for _, v := range results {
				fmt.Fprintf(stdout, "- %s | Status: %s | Matching: %d/%d | Indexed Through: %d\n",
					v.VerifiedAt.Format("2006-01-02 15:04:05"), v.Status, v.MatchingMembers, v.TotalMembers, v.IndexedThrough)
			}
		}
		return ExitSuccess

	case "verify":
		if *owner == "" || *tag == "" {
			fmt.Fprintln(stderr, "Missing required flags: --owner and --tag")
			return ExitInvalidInput
		}
		var expHash []byte
		if *expectedWASM != "" {
			h, err := hex.DecodeString(*expectedWASM)
			if err != nil || len(h) != 32 {
				fmt.Fprintln(stderr, "Invalid --expected-wasm: must be a 64-character hex string (32 bytes)")
				return ExitInvalidInput
			}
			expHash = h
		} else {
			f, err := fleetService.GetFleet(ctx, fleet.FleetID{Owner: *owner, Tag: *tag})
			if err != nil || len(f.CurrentWASMHash) != 32 {
				fmt.Fprintln(stderr, "Could not infer expected WASM from fleet; specify --expected-wasm explicitly")
				return ExitInvalidInput
			}
			expHash = f.CurrentWASMHash
		}

		res, err := verifier.VerifyFleet(ctx, verification.FleetID{Owner: *owner, Tag: *tag}, expHash)
		if err != nil {
			fmt.Fprintf(stderr, "Verification error: %v\n", err)
			return ExitInternalError
		}

		if *jsonOutput {
			_ = json.NewEncoder(stdout).Encode(res)
		} else {
			fmt.Fprintf(stdout, "Verification Result: %s\n", res.Status)
			fmt.Fprintf(stdout, "Fleet:               %s:%s\n", res.FleetID.Owner, res.FleetID.Tag)
			fmt.Fprintf(stdout, "Expected WASM:       %x\n", res.ExpectedWASMHash)
			fmt.Fprintf(stdout, "Total Members:       %d\n", res.TotalMembers)
			fmt.Fprintf(stdout, "Matching Members:    %d\n", res.MatchingMembers)
			fmt.Fprintf(stdout, "Mismatching Members: %d\n", res.MismatchingMembers)
			fmt.Fprintf(stdout, "Missing Members:     %d\n", res.MissingMembers)
			fmt.Fprintf(stdout, "Indexed Through:     %d\n", res.IndexedThrough)
		}

		switch res.Status {
		case verification.StatusHealthy:
			return ExitSuccess
		case verification.StatusDrift:
			return ExitDrift
		case verification.StatusBrokenReference:
			return ExitBrokenReference
		case verification.StatusIncomplete:
			return ExitIncomplete
		default:
			return ExitInternalError
		}

	default:
		fmt.Fprintf(stderr, "Unknown fleet subcommand: %s\n", subcmd)
		return ExitInvalidInput
	}
}

func runContractCommand(args []string, cfg *config.Config, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "Subcommand required: inspect")
		return ExitInvalidInput
	}
	subcmd := args[0]
	fs := flag.NewFlagSet("contract "+subcmd, flag.ContinueOnError)
	fs.SetOutput(stderr)

	contract := fs.String("contract", "", "Contract StrKey address")
	jsonOutput := fs.Bool("json", false, "Output format JSON")

	if err := fs.Parse(args[1:]); err != nil {
		return ExitInvalidInput
	}

	if *contract == "" {
		fmt.Fprintln(stderr, "Missing required flag: --contract")
		return ExitInvalidInput
	}

	db, err := getDB(cfg)
	if err != nil {
		fmt.Fprintf(stderr, "Database error: %v\n", err)
		return ExitInternalError
	}
	defer db.Close()

	fleetRepo := fleet.NewPostgresRepository(db)
	var resolver cap85.Resolver
	if cfg.RPCURL != "" {
		resolver = cap85.NewResolver(rpc.NewClient(cfg.RPCURL))
	}
	fleetService := fleet.NewService(fleetRepo, resolver)

	inspection, err := fleetService.InspectContract(context.Background(), *contract)
	if err != nil {
		fmt.Fprintf(stderr, "Contract inspection error: %v\n", err)
		return ExitInvalidInput
	}

	if *jsonOutput {
		_ = json.NewEncoder(stdout).Encode(inspection)
	} else {
		fmt.Fprintf(stdout, "Contract: %s\n", inspection.ContractID)
		fmt.Fprintf(stdout, "Is Fleet Member: %v\n", inspection.IsMember)
		if inspection.Member != nil {
			fmt.Fprintf(stdout, "Fleet: %s:%s\n", inspection.Member.FleetID.Owner, inspection.Member.FleetID.Tag)
			fmt.Fprintf(stdout, "Member WASM: %x\n", inspection.Member.WASMHash)
		}
		fmt.Fprintf(stdout, "Live Executable Kind: %s\n", inspection.Resolved.Kind)
		if len(inspection.Resolved.WASMHash) > 0 {
			fmt.Fprintf(stdout, "Live WASM Hash:       %x\n", inspection.Resolved.WASMHash)
		}
	}
	return ExitSuccess
}

func runMigrateCommand(args []string, cfg *config.Config, stdout, stderr io.Writer) int {
	action := "up"
	if len(args) > 0 {
		action = args[0]
	}

	db, err := getDB(cfg)
	if err != nil {
		fmt.Fprintf(stderr, "Database error: %v\n", err)
		return ExitInternalError
	}
	defer db.Close()

	switch action {
	case "up":
		if err := migrations.Up(db); err != nil {
			fmt.Fprintf(stderr, "Migration up failed: %v\n", err)
			return ExitInternalError
		}
		fmt.Fprintln(stdout, "Migrations applied successfully.")
	case "down":
		if err := migrations.Down(db); err != nil {
			fmt.Fprintf(stderr, "Migration down failed: %v\n", err)
			return ExitInternalError
		}
		fmt.Fprintln(stdout, "Migrations rolled back successfully.")
	default:
		fmt.Fprintf(stderr, "Unknown migrate action: %s (expected up or down)\n", action)
		return ExitInvalidInput
	}
	return ExitSuccess
}

func runAPICommand(cfg *config.Config, stdout, stderr io.Writer) int {
	if err := cfg.ValidateForAPI(); err != nil {
		fmt.Fprintf(stderr, "API configuration error: %v\n", err)
		return ExitInvalidInput
	}

	db, err := getDB(cfg)
	if err != nil {
		fmt.Fprintf(stderr, "Database error: %v\n", err)
		return ExitInternalError
	}
	defer db.Close()

	fleetRepo := fleet.NewPostgresRepository(db)
	var rpcClient rpc.Client
	var resolver cap85.Resolver
	if cfg.RPCURL != "" {
		rpcClient = rpc.NewClient(cfg.RPCURL)
		resolver = cap85.NewResolver(rpcClient)
	}
	fleetService := fleet.NewService(fleetRepo, resolver)
	checkpoints := ingest.NewPostgresCheckpointStore(db)
	verifier := verification.NewEngine(db, fleetRepo, checkpoints, resolver, rpcClient)
	releaseRepo := history.NewPostgresRepository(db)

	logger := slog.New(slog.NewTextHandler(stdout, nil))
	server := api.NewServer(cfg.HTTPAddr, fleetService, verifier, releaseRepo, logger)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		errChan <- server.Start()
	}()

	select {
	case <-ctx.Done():
		logger.Info("shutting down api server...")
		shutdownCtx, sCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer sCancel()
		_ = server.Shutdown(shutdownCtx)
		return ExitSuccess
	case err := <-errChan:
		if err != nil {
			fmt.Fprintf(stderr, "API server failed: %v\n", err)
			return ExitInternalError
		}
		return ExitSuccess
	}
}

func runIngestCommand(cfg *config.Config, stdout, stderr io.Writer) int {
	if err := cfg.ValidateForIndexer(); err != nil {
		fmt.Fprintf(stderr, "Indexer configuration error: %v\n", err)
		return ExitInvalidInput
	}

	db, err := getDB(cfg)
	if err != nil {
		fmt.Fprintf(stderr, "Database error: %v\n", err)
		return ExitInternalError
	}
	defer db.Close()

	fleetRepo := fleet.NewPostgresRepository(db)
	membership := fleet.NewMembershipManager(fleetRepo)
	rpcClient := rpc.NewClient(cfg.RPCURL)
	resolver := cap85.NewResolver(rpcClient)
	releaseRepo := history.NewPostgresRepository(db)
	checkpoints := ingest.NewPostgresCheckpointStore(db)

	logger := slog.New(slog.NewTextHandler(stdout, nil))
	processor := ingest.NewLedgerProcessor(fleetRepo, membership, resolver, releaseRepo)
	pipeline := ingest.NewPipeline(db, rpcClient, processor, checkpoints, logger)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	opts := ingest.PipelineOptions{
		StreamName:  ingest.DefaultStreamName,
		StartLedger: cfg.IndexerStartLedger,
		BatchSize:   cfg.IndexerBatchSize,
	}

	logger.Info("starting ingestion pipeline", "startLedger", opts.StartLedger, "batchSize", opts.BatchSize)
	if err := pipeline.Run(ctx, opts); err != nil && err != context.Canceled {
		fmt.Fprintf(stderr, "Ingestion pipeline error: %v\n", err)
		return ExitInternalError
	}

	return ExitSuccess
}
