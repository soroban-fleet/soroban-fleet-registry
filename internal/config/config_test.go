package config

import (
	"os"
	"testing"
)

func TestConfigDefaults(t *testing.T) {
	// Clear env
	os.Unsetenv("SFR_NETWORK")
	os.Unsetenv("SFR_RPC_URL")
	os.Unsetenv("SFR_DATABASE_URL")
	os.Unsetenv("SFR_LOG_LEVEL")
	os.Unsetenv("SFR_HTTP_ADDR")
	os.Unsetenv("SFR_INDEXER_START_LEDGER")
	os.Unsetenv("SFR_INDEXER_BATCH_SIZE")
	os.Unsetenv("SFR_READ_ONLY")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error loading defaults: %v", err)
	}

	if cfg.LogLevel != "info" {
		t.Errorf("expected default log level 'info', got %s", cfg.LogLevel)
	}
	if cfg.HTTPAddr != ":8080" {
		t.Errorf("expected default http addr ':8080', got %s", cfg.HTTPAddr)
	}
	if cfg.IndexerBatchSize != 50 {
		t.Errorf("expected default batch size 50, got %d", cfg.IndexerBatchSize)
	}
	if cfg.ReadOnly {
		t.Errorf("expected default read only to be false")
	}
}

func TestConfigCustomEnv(t *testing.T) {
	t.Setenv("SFR_NETWORK", "testnet")
	t.Setenv("SFR_RPC_URL", "https://rpc.testnet.stellar.org")
	t.Setenv("SFR_DATABASE_URL", "postgres://localhost/test")
	t.Setenv("SFR_LOG_LEVEL", "DEBUG")
	t.Setenv("SFR_HTTP_ADDR", ":9090")
	t.Setenv("SFR_INDEXER_START_LEDGER", "12345")
	t.Setenv("SFR_INDEXER_BATCH_SIZE", "100")
	t.Setenv("SFR_READ_ONLY", "true")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Network != "testnet" {
		t.Errorf("expected network 'testnet', got %s", cfg.Network)
	}
	if cfg.RPCURL != "https://rpc.testnet.stellar.org" {
		t.Errorf("expected rpc url, got %s", cfg.RPCURL)
	}
	if cfg.DatabaseURL != "postgres://localhost/test" {
		t.Errorf("expected database url, got %s", cfg.DatabaseURL)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("expected log level 'debug', got %s", cfg.LogLevel)
	}
	if cfg.HTTPAddr != ":9090" {
		t.Errorf("expected http addr ':9090', got %s", cfg.HTTPAddr)
	}
	if cfg.IndexerStartLedger != 12345 {
		t.Errorf("expected start ledger 12345, got %d", cfg.IndexerStartLedger)
	}
	if cfg.IndexerBatchSize != 100 {
		t.Errorf("expected batch size 100, got %d", cfg.IndexerBatchSize)
	}
	if !cfg.ReadOnly {
		t.Errorf("expected read only to be true")
	}

	if err := cfg.ValidateForIndexer(); err != nil {
		t.Errorf("expected indexer validation to pass, got: %v", err)
	}
	if err := cfg.ValidateForAPI(); err != nil {
		t.Errorf("expected api validation to pass, got: %v", err)
	}
}

func TestConfigValidationErrors(t *testing.T) {
	t.Setenv("SFR_DATABASE_URL", "")
	t.Setenv("SFR_RPC_URL", "")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected load error: %v", err)
	}
	if err := cfg.ValidateForIndexer(); err == nil {
		t.Errorf("expected error validating empty config for indexer")
	}
	if err := cfg.ValidateForAPI(); err == nil {
		t.Errorf("expected error validating empty config for api")
	}

	t.Setenv("SFR_INDEXER_START_LEDGER", "invalid")
	if _, err := Load(); err == nil {
		t.Errorf("expected error for invalid start ledger")
	}

	t.Setenv("SFR_INDEXER_START_LEDGER", "10")
	t.Setenv("SFR_INDEXER_BATCH_SIZE", "0")
	if _, err := Load(); err == nil {
		t.Errorf("expected error for 0 batch size")
	}

	t.Setenv("SFR_INDEXER_BATCH_SIZE", "50")
	t.Setenv("SFR_READ_ONLY", "notabool")
	if _, err := Load(); err == nil {
		t.Errorf("expected error for invalid read only boolean")
	}
}
