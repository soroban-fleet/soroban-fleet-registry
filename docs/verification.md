# Verification Engine

The verification engine provides deterministic, reproducible assertions about the operational state of a fleet.

## Algorithm

1. **Load Active Set**: Fetch all members of `(owner, tag)` where `active = TRUE`.
2. **Evaluate Indexer Completeness**: Check the indexer checkpoint sequence against the latest ledger closed on the network.
3. **Resolve Live State**: Use Stellar RPC to query the current on-chain executable state of every active member.
4. **Compare WASM Hashes**: Check each resolved hash against the expected WASM hash.
5. **Classify Status**: Compute exact counts (`total`, `matching`, `mismatching`, `missing`) and determine status.
6. **Persist Audit Trail**: Store the result in `fleet_verifications`.

## Status Classifications

### HEALTHY
Returned **only** when:
- All active member contracts resolve successfully on-chain;
- Every resolved WASM hash matches the expected hash;
- Total active members is greater than zero;
- Indexer coverage is complete (`checkpoint >= latest_ledger`).

> **Safety Rule**: Never return `HEALTHY` merely because no mismatch was detected.

### DRIFT
Returned when:
- One or more active member contracts resolve on-chain to a WASM hash different from the expected hash.

### BROKEN_REFERENCE
Returned when:
- The fleet's external reference cannot be resolved on-chain (e.g. the owner contract does not exist, the owner has no entry keyed by the executable tag, or the entry is malformed).

### INCOMPLETE
Returned when:
- The indexer has not reached the ledger height required to make an assertion (`checkpoint < latest_ledger`);
- Any active member cannot be resolved due to indexer lag or missing RPC entries;
- The fleet has zero active members indexed.

### UNKNOWN
Returned when:
- Insufficient information exists to safely categorize the fleet into the states above.

## CLI Exit Codes

| Status | Exit Code | Description |
|---|---|---|
| `HEALTHY` | `0` | Fleet is healthy and all active members match |
| `DRIFT` | `1` | One or more active members have drifted |
| `BROKEN_REFERENCE` | `2` | Fleet reference is unresolvable or broken |
| `INCOMPLETE` | `3` | Indexer is lagging or information is incomplete |
| Invalid Input | `4` | Invalid CLI arguments, bad hex, or missing parameters |
| Internal Error | `5` | Database failure, RPC failure, or unexpected error |
