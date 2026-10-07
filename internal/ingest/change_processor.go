package ingest

import (
	"encoding/hex"

	"github.com/soroban-fleet/soroban-fleet-registry/internal/cap85"
	"github.com/stellar/go-stellar-sdk/xdr"
)

// ContractInstanceChange represents an observed modification to a contract instance executable.
type ContractInstanceChange struct {
	ContractID string
	OldExec    *xdr.ContractExecutable
	NewExec    *xdr.ContractExecutable
	TxHash     string
}

// TagExecutableChange represents an observed modification to an owner contract's executable tag entry.
type TagExecutableChange struct {
	OwnerAddress string
	Tag          string
	OldWASMHash  []byte
	NewWASMHash  []byte
	TxHash       string
}

// LedgerChanges holds all parsed CAP-85 relevant changes for a single ledger.
type LedgerChanges struct {
	Sequence        uint32
	CloseTimeUnix   int64
	ContractChanges []ContractInstanceChange
	TagChanges      []TagExecutableChange
}

// ExtractChangesFromLedgerCloseMeta parses an XDR LedgerCloseMeta into domain-specific changes.
func ExtractChangesFromLedgerCloseMeta(meta xdr.LedgerCloseMeta) (LedgerChanges, error) {
	seq := meta.LedgerSequence()
	closeTime := meta.LedgerCloseTime()

	var contractChanges []ContractInstanceChange
	var tagChanges []TagExecutableChange

	numTxs := meta.CountTransactions()
	for i := 0; i < numTxs; i++ {
		txHashBytes := meta.TransactionHash(i)
		txHash := hex.EncodeToString(txHashBytes[:])

		txMeta := meta.TxApplyProcessing(i)
		changes := extractChangesFromTxMeta(txMeta)

		// Iterate through changes to detect [STATE, UPDATED] pairs or standalone [CREATED], [REMOVED]
		var prevState *xdr.LedgerEntry

		for _, ch := range changes {
			switch ch.Type {
			case xdr.LedgerEntryChangeTypeLedgerEntryState:
				entry := ch.MustState()
				prevState = &entry

			case xdr.LedgerEntryChangeTypeLedgerEntryCreated:
				entry := ch.MustCreated()
				processEntryChange(nil, &entry, txHash, &contractChanges, &tagChanges)
				prevState = nil

			case xdr.LedgerEntryChangeTypeLedgerEntryUpdated:
				entry := ch.MustUpdated()
				processEntryChange(prevState, &entry, txHash, &contractChanges, &tagChanges)
				prevState = nil

			case xdr.LedgerEntryChangeTypeLedgerEntryRemoved:
				if prevState != nil {
					processEntryRemoval(prevState, txHash, &contractChanges)
				}
				prevState = nil

			case xdr.LedgerEntryChangeTypeLedgerEntryRestored:
				entry := ch.MustRestored()
				processEntryChange(prevState, &entry, txHash, &contractChanges, &tagChanges)
				prevState = nil
			}
		}
	}

	return LedgerChanges{
		Sequence:        seq,
		CloseTimeUnix:   closeTime,
		ContractChanges: contractChanges,
		TagChanges:      tagChanges,
	}, nil
}

func extractChangesFromTxMeta(txMeta xdr.TransactionMeta) []xdr.LedgerEntryChange {
	var changes []xdr.LedgerEntryChange

	switch txMeta.V {
	case 3:
		v3 := txMeta.MustV3()
		changes = append(changes, v3.TxChangesBefore...)
		for _, op := range v3.Operations {
			changes = append(changes, op.Changes...)
		}
		changes = append(changes, v3.TxChangesAfter...)
	case 4:
		v4 := txMeta.MustV4()
		changes = append(changes, v4.TxChangesBefore...)
		for _, op := range v4.Operations {
			changes = append(changes, op.Changes...)
		}
		changes = append(changes, v4.TxChangesAfter...)
	}

	return changes
}

func processEntryChange(
	oldEntry *xdr.LedgerEntry,
	newEntry *xdr.LedgerEntry,
	txHash string,
	contractChanges *[]ContractInstanceChange,
	tagChanges *[]TagExecutableChange,
) {
	if newEntry == nil || newEntry.Data.Type != xdr.LedgerEntryTypeContractData {
		return
	}

	cData := newEntry.Data.ContractData
	if cData == nil {
		return
	}

	// 1. Check for contract instance change
	if cData.Key.Type == xdr.ScValTypeScvLedgerKeyContractInstance {
		contractID, err := cap85.DecodeScAddress(cData.Contract)
		if err != nil {
			return
		}

		newInst, ok := cData.Val.GetInstance()
		if !ok {
			return
		}

		var oldExec *xdr.ContractExecutable
		if oldEntry != nil && oldEntry.Data.Type == xdr.LedgerEntryTypeContractData && oldEntry.Data.ContractData != nil {
			if oldInst, ok := oldEntry.Data.ContractData.Val.GetInstance(); ok {
				execCopy := oldInst.Executable
				oldExec = &execCopy
			}
		}

		execCopy := newInst.Executable
		*contractChanges = append(*contractChanges, ContractInstanceChange{
			ContractID: contractID,
			OldExec:    oldExec,
			NewExec:    &execCopy,
			TxHash:     txHash,
		})
		return
	}

	// 2. Check for owner executable tag change
	if cData.Key.Type == xdr.ScValTypeScvExecutableTag {
		ownerAddr, err := cap85.DecodeScAddress(cData.Contract)
		if err != nil {
			return
		}

		tagStr, err := cap85.ValidateAndDecodeTag(*cData.Key.ExecutableTag)
		if err != nil {
			return
		}

		newBytes, ok := cData.Val.GetBytes()
		if !ok || len(newBytes) != 32 {
			return
		}

		var oldBytes []byte
		if oldEntry != nil && oldEntry.Data.Type == xdr.LedgerEntryTypeContractData && oldEntry.Data.ContractData != nil {
			if b, ok := oldEntry.Data.ContractData.Val.GetBytes(); ok && len(b) == 32 {
				oldBytes = append([]byte(nil), b...)
			}
		}

		*tagChanges = append(*tagChanges, TagExecutableChange{
			OwnerAddress: ownerAddr,
			Tag:          tagStr,
			OldWASMHash:  oldBytes,
			NewWASMHash:  append([]byte(nil), newBytes...),
			TxHash:       txHash,
		})
	}
}

func processEntryRemoval(
	oldEntry *xdr.LedgerEntry,
	txHash string,
	contractChanges *[]ContractInstanceChange,
) {
	if oldEntry == nil || oldEntry.Data.Type != xdr.LedgerEntryTypeContractData {
		return
	}
	cData := oldEntry.Data.ContractData
	if cData == nil {
		return
	}

	if cData.Key.Type == xdr.ScValTypeScvLedgerKeyContractInstance {
		contractID, err := cap85.DecodeScAddress(cData.Contract)
		if err != nil {
			return
		}
		if oldInst, ok := cData.Val.GetInstance(); ok {
			execCopy := oldInst.Executable
			*contractChanges = append(*contractChanges, ContractInstanceChange{
				ContractID: contractID,
				OldExec:    &execCopy,
				NewExec:    nil, // Removed
				TxHash:     txHash,
			})
		}
	}
}
