package rpc

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	protocol "github.com/stellar/go-stellar-sdk/protocols/rpc"
	"github.com/stellar/go-stellar-sdk/xdr"
)

type jsonRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

type jsonRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

func TestRPCClient_GetLatestLedger(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req jsonRPCRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if req.Method != "getLatestLedger" {
			t.Fatalf("unexpected method: %s", req.Method)
		}

		res := jsonRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: protocol.GetLatestLedgerResponse{
				Hash:            "abcd1234abcd1234",
				ProtocolVersion: 28,
				Sequence:        50000,
				LedgerCloseTime: 1700000000,
			},
		}
		_ = json.NewEncoder(w).Encode(res)
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	ctx := context.Background()

	latest, err := client.GetLatestLedger(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if latest.Sequence != 50000 {
		t.Errorf("expected sequence 50000, got %d", latest.Sequence)
	}
	if latest.ProtocolVersion != 28 {
		t.Errorf("expected protocol version 28, got %d", latest.ProtocolVersion)
	}
	if latest.Hash != "abcd1234abcd1234" {
		t.Errorf("expected hash abcd1234abcd1234, got %s", latest.Hash)
	}
}

func TestRPCClient_GetLedgerEntries_Batching(t *testing.T) {
	batchCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req jsonRPCRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if req.Method != "getLedgerEntries" {
			t.Fatalf("unexpected method: %s", req.Method)
		}

		var params protocol.GetLedgerEntriesRequest
		if err := json.Unmarshal(req.Params, &params); err != nil {
			t.Fatalf("failed to unmarshal params: %v", err)
		}

		batchCount++
		if len(params.Keys) > MaxLedgerKeysPerRequest {
			t.Fatalf("batch size %d exceeded max %d", len(params.Keys), MaxLedgerKeysPerRequest)
		}

		entries := make([]protocol.LedgerEntryResult, len(params.Keys))
		for i, k := range params.Keys {
			entries[i] = protocol.LedgerEntryResult{
				KeyXDR:             k,
				LastModifiedLedger: 100,
			}
		}

		res := jsonRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: protocol.GetLedgerEntriesResponse{
				Entries:      entries,
				LatestLedger: 50000,
			},
		}
		_ = json.NewEncoder(w).Encode(res)
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	ctx := context.Background()

	// Empty keys
	emptyResp, err := client.GetLedgerEntries(ctx, nil)
	if err != nil {
		t.Fatalf("unexpected error on empty keys: %v", err)
	}
	if len(emptyResp.Entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(emptyResp.Entries))
	}

	// Create 250 keys to verify batching into 3 requests (100 + 100 + 50)
	totalKeys := 250
	keys := make([]xdr.LedgerKey, totalKeys)
	for i := 0; i < totalKeys; i++ {
		var key xdr.LedgerKey
		var contractID xdr.ContractId
		contractID[0] = byte(i)
		contractID[1] = byte(i >> 8)
		scAddr, err := xdr.NewScAddress(xdr.ScAddressTypeScAddressTypeContract, contractID)
		if err != nil {
			t.Fatalf("error creating sc address: %v", err)
		}
		scVal, err := xdr.NewScVal(xdr.ScValTypeScvLedgerKeyContractInstance, nil)
		if err != nil {
			t.Fatalf("error creating sc val: %v", err)
		}
		if err := key.SetContractData(scAddr, scVal, xdr.ContractDataDurabilityPersistent); err != nil {
			t.Fatalf("error setting contract data: %v", err)
		}
		keys[i] = key
	}

	resp, err := client.GetLedgerEntries(ctx, keys)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Entries) != totalKeys {
		t.Errorf("expected %d entries, got %d", totalKeys, len(resp.Entries))
	}
	if batchCount != 3 {
		t.Errorf("expected 3 batches, got %d", batchCount)
	}
	if resp.LatestLedger != 50000 {
		t.Errorf("expected latest ledger 50000, got %d", resp.LatestLedger)
	}
}

func TestRPCClient_GetLedgers(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req jsonRPCRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if req.Method != "getLedgers" {
			t.Fatalf("unexpected method: %s", req.Method)
		}

		res := jsonRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: protocol.GetLedgersResponse{
				Ledgers: []protocol.LedgerInfo{
					{
						Hash:            "hash123",
						Sequence:        1000,
						LedgerCloseTime: 1700000000,
					},
				},
				LatestLedger: 1005,
			},
		}
		_ = json.NewEncoder(w).Encode(res)
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	ctx := context.Background()

	resp, err := client.GetLedgers(ctx, 1000, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Ledgers) != 1 {
		t.Fatalf("expected 1 ledger, got %d", len(resp.Ledgers))
	}
	if resp.Ledgers[0].Sequence != 1000 {
		t.Errorf("expected sequence 1000, got %d", resp.Ledgers[0].Sequence)
	}
}
