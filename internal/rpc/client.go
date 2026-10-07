package rpc

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/stellar/go-stellar-sdk/clients/rpcclient"
	protocol "github.com/stellar/go-stellar-sdk/protocols/rpc"
	"github.com/stellar/go-stellar-sdk/xdr"
)

const (
	// MaxLedgerKeysPerRequest is the safe maximum keys requested per getLedgerEntries batch.
	MaxLedgerKeysPerRequest = 100
)

// LatestLedger contains metadata about the most recent ledger closed on the network.
type LatestLedger struct {
	Hash            string
	ProtocolVersion uint32
	Sequence        uint32
	CloseTime       time.Time
}

// LedgerEntriesResponse wraps the parsed ledger entries returned by the RPC.
type LedgerEntriesResponse struct {
	Entries      []protocol.LedgerEntryResult
	LatestLedger uint32
}

// Client defines the interface for Stellar RPC operations needed by soroban-fleet-registry.
type Client interface {
	GetLedgerEntries(ctx context.Context, keys []xdr.LedgerKey) (LedgerEntriesResponse, error)
	GetLatestLedger(ctx context.Context) (LatestLedger, error)
	GetLedgers(ctx context.Context, startLedger uint32, limit uint32) (protocol.GetLedgersResponse, error)
}

// SDKClient wraps the official Stellar SDK RPC client.
type SDKClient struct {
	sdk *rpcclient.Client
}

// NewClient creates a new RPC client for the specified endpoint.
func NewClient(endpointURL string) *SDKClient {
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}
	return &SDKClient{
		sdk: rpcclient.NewClient(endpointURL, httpClient),
	}
}

// NewClientWithHTTP creates a new RPC client with a custom HTTP client.
func NewClientWithHTTP(endpointURL string, httpClient *http.Client) *SDKClient {
	return &SDKClient{
		sdk: rpcclient.NewClient(endpointURL, httpClient),
	}
}

// GetLatestLedger returns the latest ledger sequence, hash, and close time.
func (c *SDKClient) GetLatestLedger(ctx context.Context) (LatestLedger, error) {
	resp, err := c.sdk.GetLatestLedger(ctx)
	if err != nil {
		return LatestLedger{}, fmt.Errorf("rpc getLatestLedger: %w", err)
	}

	return LatestLedger{
		Hash:            resp.Hash,
		ProtocolVersion: resp.ProtocolVersion,
		Sequence:        resp.Sequence,
		CloseTime:       time.Unix(resp.LedgerCloseTime, 0).UTC(),
	}, nil
}

// GetLedgerEntries retrieves ledger entries for the specified keys, batching into chunks if needed.
func (c *SDKClient) GetLedgerEntries(ctx context.Context, keys []xdr.LedgerKey) (LedgerEntriesResponse, error) {
	if len(keys) == 0 {
		return LedgerEntriesResponse{}, nil
	}

	var allEntries []protocol.LedgerEntryResult
	var latestLedger uint32

	// Encode all keys to base64 XDR
	encodedKeys := make([]string, len(keys))
	for i, k := range keys {
		raw, err := k.MarshalBinary()
		if err != nil {
			return LedgerEntriesResponse{}, fmt.Errorf("marshal ledger key %d: %w", i, err)
		}
		encodedKeys[i] = base64.StdEncoding.EncodeToString(raw)
	}

	// Chunk keys to respect RPC batch limits
	for i := 0; i < len(encodedKeys); i += MaxLedgerKeysPerRequest {
		end := i + MaxLedgerKeysPerRequest
		if end > len(encodedKeys) {
			end = len(encodedKeys)
		}
		chunk := encodedKeys[i:end]

		req := protocol.GetLedgerEntriesRequest{
			Keys: chunk,
		}

		resp, err := c.sdk.GetLedgerEntries(ctx, req)
		if err != nil {
			return LedgerEntriesResponse{}, fmt.Errorf("rpc getLedgerEntries batch [%d:%d]: %w", i, end, err)
		}

		if resp.LatestLedger > latestLedger {
			latestLedger = resp.LatestLedger
		}
		allEntries = append(allEntries, resp.Entries...)
	}

	return LedgerEntriesResponse{
		Entries:      allEntries,
		LatestLedger: latestLedger,
	}, nil
}

// GetLedgers fetches a batch of ledgers starting at startLedger up to limit.
func (c *SDKClient) GetLedgers(ctx context.Context, startLedger uint32, limit uint32) (protocol.GetLedgersResponse, error) {
	req := protocol.GetLedgersRequest{
		StartLedger: startLedger,
		Pagination: &protocol.LedgerPaginationOptions{
			Limit: uint(limit),
		},
	}
	resp, err := c.sdk.GetLedgers(ctx, req)
	if err != nil {
		return protocol.GetLedgersResponse{}, fmt.Errorf("rpc getLedgers (start=%d limit=%d): %w", startLedger, limit, err)
	}
	return resp, nil
}
