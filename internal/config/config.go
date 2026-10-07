package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config represents the application runtime configuration loaded from environment variables.
type Config struct {
	Network            string
	RPCURL             string
	DatabaseURL        string
	LogLevel           string
	HTTPAddr           string
	IndexerStartLedger uint32
	IndexerBatchSize   uint32
	MetricsAddr        string
	ReadOnly           bool
}

const (
	defaultLogLevel         = "info"
	defaultHTTPAddr         = ":8080"
	defaultIndexerBatchSize = uint32(50)
)

// Load reads configuration from the environment and applies defaults.
func Load() (*Config, error) {
	cfg := &Config{
		Network:          os.Getenv("SFR_NETWORK"),
		RPCURL:           os.Getenv("SFR_RPC_URL"),
		DatabaseURL:      os.Getenv("SFR_DATABASE_URL"),
		LogLevel:         os.Getenv("SFR_LOG_LEVEL"),
		HTTPAddr:         os.Getenv("SFR_HTTP_ADDR"),
		MetricsAddr:      os.Getenv("SFR_METRICS_ADDR"),
		IndexerBatchSize: defaultIndexerBatchSize,
	}

	if cfg.LogLevel == "" {
		cfg.LogLevel = defaultLogLevel
	}
	cfg.LogLevel = strings.ToLower(cfg.LogLevel)

	if cfg.HTTPAddr == "" {
		cfg.HTTPAddr = defaultHTTPAddr
	}

	if startLedgerStr := os.Getenv("SFR_INDEXER_START_LEDGER"); startLedgerStr != "" {
		val, err := strconv.ParseUint(startLedgerStr, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid SFR_INDEXER_START_LEDGER %q: %w", startLedgerStr, err)
		}
		cfg.IndexerStartLedger = uint32(val)
	}

	if batchSizeStr := os.Getenv("SFR_INDEXER_BATCH_SIZE"); batchSizeStr != "" {
		val, err := strconv.ParseUint(batchSizeStr, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid SFR_INDEXER_BATCH_SIZE %q: %w", batchSizeStr, err)
		}
		if val == 0 {
			return nil, fmt.Errorf("SFR_INDEXER_BATCH_SIZE must be greater than 0")
		}
		cfg.IndexerBatchSize = uint32(val)
	}

	if readOnlyStr := os.Getenv("SFR_READ_ONLY"); readOnlyStr != "" {
		val, err := strconv.ParseBool(readOnlyStr)
		if err != nil {
			return nil, fmt.Errorf("invalid SFR_READ_ONLY %q: %w", readOnlyStr, err)
		}
		cfg.ReadOnly = val
	}

	return cfg, nil
}

// ValidateForIndexer checks that required configuration is present for ingestion mode.
func (c *Config) ValidateForIndexer() error {
	if c.DatabaseURL == "" {
		return fmt.Errorf("SFR_DATABASE_URL is required for indexer")
	}
	if c.RPCURL == "" {
		return fmt.Errorf("SFR_RPC_URL is required for indexer")
	}
	return nil
}

// ValidateForAPI checks that required configuration is present for API mode.
func (c *Config) ValidateForAPI() error {
	if c.DatabaseURL == "" {
		return fmt.Errorf("SFR_DATABASE_URL is required for api")
	}
	return nil
}
