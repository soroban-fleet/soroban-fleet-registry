package config

import (
	"fmt"
	"math"
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
	AllowedOrigins     []string

	// RateLimitRPS and RateLimitBurst bound per-client request rates for
	// ordinary API reads. A rate of zero disables the limit.
	RateLimitRPS   float64
	RateLimitBurst int

	// HeavyRateLimitRPS and HeavyRateLimitBurst bound per-client rates for
	// endpoints that resolve live ledger state through Stellar RPC.
	HeavyRateLimitRPS   float64
	HeavyRateLimitBurst int

	// TrustProxy keys rate limit buckets by X-Forwarded-For. Enable only when
	// the API is reachable exclusively through a trusted reverse proxy.
	TrustProxy bool
}

const (
	defaultLogLevel            = "info"
	defaultHTTPAddr            = ":8080"
	defaultIndexerBatchSize    = uint32(50)
	defaultRateLimitRPS        = 20.0
	defaultRateLimitBurst      = 60
	defaultHeavyRateLimitRPS   = 1.0
	defaultHeavyRateLimitBurst = 5
)

// Load reads configuration from the environment and applies defaults.
func Load() (*Config, error) {
	cfg := &Config{
		Network:             os.Getenv("SFR_NETWORK"),
		RPCURL:              os.Getenv("SFR_RPC_URL"),
		DatabaseURL:         os.Getenv("SFR_DATABASE_URL"),
		LogLevel:            os.Getenv("SFR_LOG_LEVEL"),
		HTTPAddr:            os.Getenv("SFR_HTTP_ADDR"),
		MetricsAddr:         os.Getenv("SFR_METRICS_ADDR"),
		IndexerBatchSize:    defaultIndexerBatchSize,
		RateLimitRPS:        defaultRateLimitRPS,
		RateLimitBurst:      defaultRateLimitBurst,
		HeavyRateLimitRPS:   defaultHeavyRateLimitRPS,
		HeavyRateLimitBurst: defaultHeavyRateLimitBurst,
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

	if originsStr := os.Getenv("SFR_CORS_ALLOWED_ORIGINS"); originsStr != "" {
		parts := strings.Split(originsStr, ",")
		for _, p := range parts {
			trimmed := strings.TrimSpace(p)
			if trimmed != "" {
				cfg.AllowedOrigins = append(cfg.AllowedOrigins, trimmed)
			}
		}
	}

	var err error

	if cfg.RateLimitRPS, err = readRate("SFR_RATE_LIMIT_RPS", defaultRateLimitRPS); err != nil {
		return nil, err
	}
	if cfg.RateLimitBurst, err = readBurst("SFR_RATE_LIMIT_BURST", defaultRateLimitBurst); err != nil {
		return nil, err
	}
	if cfg.HeavyRateLimitRPS, err = readRate("SFR_RATE_LIMIT_HEAVY_RPS", defaultHeavyRateLimitRPS); err != nil {
		return nil, err
	}
	if cfg.HeavyRateLimitBurst, err = readBurst("SFR_RATE_LIMIT_HEAVY_BURST", defaultHeavyRateLimitBurst); err != nil {
		return nil, err
	}

	if trustProxyStr := os.Getenv("SFR_TRUST_PROXY"); trustProxyStr != "" {
		val, err := strconv.ParseBool(trustProxyStr)
		if err != nil {
			return nil, fmt.Errorf("invalid SFR_TRUST_PROXY %q: %w", trustProxyStr, err)
		}
		cfg.TrustProxy = val
	}

	if err := cfg.validateRateLimits(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// readRate parses a per-second rate where 0 means unlimited.
func readRate(name string, def float64) (float64, error) {
	raw := os.Getenv(name)
	if raw == "" {
		return def, nil
	}
	val, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid %s %q: %w", name, raw, err)
	}
	if math.IsNaN(val) || math.IsInf(val, 0) || val < 0 {
		return 0, fmt.Errorf("%s must be a finite non-negative number, got %q", name, raw)
	}
	return val, nil
}

// readBurst parses a bucket capacity.
func readBurst(name string, def int) (int, error) {
	raw := os.Getenv(name)
	if raw == "" {
		return def, nil
	}
	val, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid %s %q: %w", name, raw, err)
	}
	if val < 0 {
		return 0, fmt.Errorf("%s must not be negative, got %d", name, val)
	}
	return val, nil
}

// validateRateLimits rejects buckets that could never admit a single request.
func (c *Config) validateRateLimits() error {
	if c.RateLimitRPS > 0 && c.RateLimitBurst < 1 {
		return fmt.Errorf("SFR_RATE_LIMIT_BURST must be at least 1 when SFR_RATE_LIMIT_RPS is enabled")
	}
	if c.HeavyRateLimitRPS > 0 && c.HeavyRateLimitBurst < 1 {
		return fmt.Errorf("SFR_RATE_LIMIT_HEAVY_BURST must be at least 1 when SFR_RATE_LIMIT_HEAVY_RPS is enabled")
	}
	return nil
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
