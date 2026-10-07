package cap85

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/soroban-fleet/soroban-fleet-registry/internal/rpc"
	"github.com/stellar/go-stellar-sdk/xdr"
)

// ResolvedExecutable represents the resolved executable type and WASM hash of a contract.
type ResolvedExecutable struct {
	Kind     string   `json:"kind"`
	Fleet    *FleetID `json:"fleet,omitempty"`
	WASMHash []byte   `json:"wasm_hash,omitempty"`
}

// Resolver defines the interface for resolving contract executables against live ledger state.
type Resolver interface {
	ResolveContractExecutable(ctx context.Context, contractID string) (ResolvedExecutable, error)
	ResolveContractExecutablesBatch(ctx context.Context, contractIDs []string) (map[string]ResolvedExecutable, error)
}

// LiveResolver resolves contract executables using a Stellar RPC client.
type LiveResolver struct {
	rpcClient rpc.Client
}

// NewResolver creates a new LiveResolver.
func NewResolver(rpcClient rpc.Client) *LiveResolver {
	return &LiveResolver{
		rpcClient: rpcClient,
	}
}

// ResolveContractExecutable resolves current on-chain executable state for a single contract.
func (r *LiveResolver) ResolveContractExecutable(ctx context.Context, contractID string) (ResolvedExecutable, error) {
	batch, err := r.ResolveContractExecutablesBatch(ctx, []string{contractID})
	if err != nil {
		return ResolvedExecutable{}, err
	}
	res, ok := batch[contractID]
	if !ok {
		return ResolvedExecutable{}, fmt.Errorf("%w: contract %s instance not found", ErrBrokenReference, contractID)
	}
	return res, nil
}

// ResolveContractExecutablesBatch resolves executables for multiple contracts using batched RPC calls.
func (r *LiveResolver) ResolveContractExecutablesBatch(ctx context.Context, contractIDs []string) (map[string]ResolvedExecutable, error) {
	if len(contractIDs) == 0 {
		return make(map[string]ResolvedExecutable), nil
	}

	results := make(map[string]ResolvedExecutable, len(contractIDs))
	keys := make([]xdr.LedgerKey, 0, len(contractIDs))
	contractMap := make(map[string]string) // base64 key string -> contractID

	for _, cid := range contractIDs {
		scAddr, err := AddressToScAddress(cid)
		if err != nil {
			return nil, fmt.Errorf("parse contract ID %s: %w", cid, err)
		}

		scValKey, err := xdr.NewScVal(xdr.ScValTypeScvLedgerKeyContractInstance, nil)
		if err != nil {
			return nil, fmt.Errorf("create instance key ScVal: %w", err)
		}

		var key xdr.LedgerKey
		if err := key.SetContractData(scAddr, scValKey, xdr.ContractDataDurabilityPersistent); err != nil {
			return nil, fmt.Errorf("set contract data key for %s: %w", cid, err)
		}

		keyBytes, err := key.MarshalBinary()
		if err != nil {
			return nil, fmt.Errorf("marshal instance key for %s: %w", cid, err)
		}
		keyB64 := base64.StdEncoding.EncodeToString(keyBytes)
		contractMap[keyB64] = cid
		keys = append(keys, key)
	}

	// Step 1: Batch fetch contract instances
	resp, err := r.rpcClient.GetLedgerEntries(ctx, keys)
	if err != nil {
		return nil, fmt.Errorf("batch getLedgerEntries for contract instances: %w", err)
	}

	type pendingExternal struct {
		contractID string
		fleetID    FleetID
	}
	pendingOwners := make([]pendingExternal, 0)
	ownerKeys := make([]xdr.LedgerKey, 0)
	ownerKeyToFleet := make(map[string]pendingExternal)

	for _, entryResult := range resp.Entries {
		cid, ok := contractMap[entryResult.KeyXDR]
		if !ok {
			continue
		}

		entryData, err := parseLedgerEntryData(entryResult.DataXDR)
		if err != nil {
			return nil, fmt.Errorf("parse contract data entry for %s: %w", cid, err)
		}

		contractData := entryData.ContractData
		if contractData == nil {
			return nil, fmt.Errorf("%w: ledger entry for %s is not a contract data entry", ErrBrokenReference, cid)
		}

		instance, ok := contractData.Val.GetInstance()
		if !ok {
			return nil, fmt.Errorf("%w: contract data value for %s is not an instance", ErrBrokenReference, cid)
		}

		exec := instance.Executable
		switch exec.Type {
		case xdr.ContractExecutableTypeContractExecutableWasm:
			wasmHash, ok := exec.GetWasmHash()
			if !ok {
				return nil, fmt.Errorf("%w: missing wasm hash for %s", ErrBrokenReference, cid)
			}
			results[cid] = ResolvedExecutable{
				Kind:     "WASM",
				WASMHash: append([]byte(nil), wasmHash[:]...),
			}

		case xdr.ContractExecutableTypeContractExecutableStellarAsset:
			results[cid] = ResolvedExecutable{
				Kind: "STELLAR_ASSET",
			}

		case xdr.ContractExecutableTypeContractExecutableExternalRef:
			extRef, isExt, err := DecodeExternalExecutableRef(exec)
			if err != nil {
				return nil, fmt.Errorf("%w: failed to decode external ref for %s: %v", ErrBrokenReference, cid, err)
			}
			if !isExt {
				return nil, fmt.Errorf("%w: decoded isExt is false for external ref %s", ErrBrokenReference, cid)
			}

			fleetID := extRef.FleetID()
			ownerScAddr, err := AddressToScAddress(extRef.Owner)
			if err != nil {
				return nil, fmt.Errorf("%w: invalid owner address %s: %v", ErrBrokenReference, extRef.Owner, err)
			}

			tagScVal, err := BuildTagScVal(extRef.Tag)
			if err != nil {
				return nil, fmt.Errorf("%w: invalid tag %s: %v", ErrBrokenReference, extRef.Tag, err)
			}

			var oKey xdr.LedgerKey
			if err := oKey.SetContractData(ownerScAddr, tagScVal, xdr.ContractDataDurabilityPersistent); err != nil {
				return nil, fmt.Errorf("%w: build owner data key for %s tag %s: %v", ErrBrokenReference, extRef.Owner, extRef.Tag, err)
			}

			oBytes, err := oKey.MarshalBinary()
			if err != nil {
				return nil, fmt.Errorf("marshal owner data key: %w", err)
			}
			oB64 := base64.StdEncoding.EncodeToString(oBytes)

			p := pendingExternal{
				contractID: cid,
				fleetID:    fleetID,
			}
			pendingOwners = append(pendingOwners, p)
			ownerKeys = append(ownerKeys, oKey)
			ownerKeyToFleet[oB64] = p

		default:
			return nil, fmt.Errorf("%w: unsupported contract executable type %v for %s", ErrBrokenReference, exec.Type, cid)
		}
	}

	// Step 2: Batch fetch owner tag entries for external references
	if len(ownerKeys) > 0 {
		ownerResp, err := r.rpcClient.GetLedgerEntries(ctx, ownerKeys)
		if err != nil {
			return nil, fmt.Errorf("batch getLedgerEntries for owner tag data: %w", err)
		}

		ownerResolved := make(map[string][]byte) // oB64 -> wasmHash
		for _, oResult := range ownerResp.Entries {
			entryData, err := parseLedgerEntryData(oResult.DataXDR)
			if err != nil {
				return nil, fmt.Errorf("parse owner entry: %w", err)
			}
			if entryData.ContractData == nil {
				continue
			}

			val := entryData.ContractData.Val
			// In CAP-85 the value of the tag entry is the WASM hash bytes (32 bytes)
			bytesVal, ok := val.GetBytes()
			if !ok {
				return nil, fmt.Errorf("%w: owner tag value is not bytes", ErrBrokenReference)
			}
			if len(bytesVal) != 32 {
				return nil, fmt.Errorf("%w: owner tag wasm hash length is %d, expected 32", ErrBrokenReference, len(bytesVal))
			}
			ownerResolved[oResult.KeyXDR] = append([]byte(nil), bytesVal...)
		}

		for oB64, p := range ownerKeyToFleet {
			hash, found := ownerResolved[oB64]
			if !found {
				return nil, fmt.Errorf("%w: owner %s has no contract data entry for tag %q (member %s)",
					ErrBrokenReference, p.fleetID.Owner, p.fleetID.Tag, p.contractID)
			}
			fleetCopy := p.fleetID
			results[p.contractID] = ResolvedExecutable{
				Kind:     "EXTERNAL_REF",
				Fleet:    &fleetCopy,
				WASMHash: hash,
			}
		}
	}

	return results, nil
}

// parseLedgerEntryData unmarshals either LedgerEntryData or LedgerEntry from base64 XDR.
func parseLedgerEntryData(dataXDR string) (xdr.LedgerEntryData, error) {
	raw, err := base64.StdEncoding.DecodeString(dataXDR)
	if err != nil {
		return xdr.LedgerEntryData{}, fmt.Errorf("base64 decode: %w", err)
	}

	var entryData xdr.LedgerEntryData
	if err := entryData.UnmarshalBinary(raw); err == nil {
		return entryData, nil
	}

	var fullEntry xdr.LedgerEntry
	if err := fullEntry.UnmarshalBinary(raw); err == nil {
		return fullEntry.Data, nil
	}

	return xdr.LedgerEntryData{}, fmt.Errorf("failed to unmarshal as LedgerEntryData or LedgerEntry")
}
