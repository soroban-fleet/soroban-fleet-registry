# Fleet lifecycle

Every fleet passes through the same five stages. Each stage answers one question.

```text
DISCOVER   →  INDEX      →  RESOLVE    →  VERIFY     →  HISTORY
find refs     persist        read owner     compare       keep the
owner+tag     membership     entry on RPC   live state    record
```

## Stage by stage

### 1. DISCOVER — find the executable reference

The indexer reads ledger changes from Stellar RPC (`getLedgers` → `LedgerCloseMeta`). When a
contract instance is created or updated with a `CONTRACT_EXECUTABLE_EXTERNAL_REF`, the reference
is detected.

### 2. INDEX — decode owner and tag, identify the fleet

The decoder extracts `(executable_owner, tag)` exactly as encoded. The pair is the fleet identity.
The membership manager upserts the fleet row and the member row, and refreshes `member_count`.

### 3. RESOLVE — find the active WASM

The resolver reads the owner contract's persistent entry keyed by `SCV_EXECUTABLE_TAG` and returns
the 32-byte hash. If the hash differs from the previously recorded one, a release row is written
with the old hash, new hash, ledger, and transaction hash.

### 4. VERIFY — check consistency

The verification engine loads active members, resolves each one's live executable over RPC,
compares hashes, checks indexer coverage against the latest network ledger, classifies the result,
and persists it to `fleet_verifications`.

### 5. HISTORY — keep the record

Releases and verification results are stored with ledger and timestamp, so the sequence of changes
and audits can be replayed at any time.

## The seven-step lifecycle

```text
1. discover executable reference
2. decode owner and tag
3. identify fleet
4. resolve active WASM
5. persist membership
6. observe changes
7. verify consistency
```

Steps 1–6 run continuously inside the ingestion pipeline. Step 7 runs on demand (`GET
/v1/fleets/{owner}/{tag}/verify`, `sfr fleet verify`) and is recorded each time.

## Observed ≠ verified

```text
release observed   ≠   release verified
```

- **Release observed**: the indexer saw the owner's hash change in a ledger. It proves the change
  happened. It says nothing about whether members currently execute that hash.
- **Release verified**: the verification engine resolved every active member live and compared it
  against an expected hash with complete indexing coverage.

A fleet can have a clean release history and still be in `DRIFT`, `BROKEN_REFERENCE`, or
`INCOMPLETE`. Read the release timeline and the verification panel separately.

## Failure at any stage

If any stage cannot complete, the system does not guess:

| Failure | Result |
|---|---|
| owner entry missing or malformed | `BROKEN_REFERENCE` |
| indexer behind network ledger | `INCOMPLETE` |
| member resolves to a different hash | `DRIFT` |
| RPC unavailable during resolution | member counted missing → `INCOMPLETE` |

See [Verification states](./verification-states.md) and
[Failure and recovery](../architecture/failure-recovery.md).
