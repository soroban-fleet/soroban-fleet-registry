# Soroban boundary

## The statement

```text
SFR V1 does not deploy or own a custom Soroban contract.

Its Soroban-specific role is reading and interpreting CAP-85
externally managed executable state.
```

## What that means concretely

| SFR does | SFR does not |
|---|---|
| Read ledger entries over Stellar RPC | Call any contract function |
| Decode `ContractExecutableExternalRef` | Deploy a contract |
| Resolve owner tag entries to a WASM hash | Own or upgrade an executable |
| Index fleets and memberships | Sign or submit transactions |
| Compare live state against expectations | Hold keys or connect wallets |
| Store and serve what it observed | Act as a bridge, oracle, or escrow |

There is no SFR contract on the Stellar network. No contract ID, no deployment address, and no
contract API exists for this project in V1 — because no such contract exists.

## Why the boundary matters to reviewers

**Independence.** The service observes fleets it does not control. Its report about a fleet is
not produced by the fleet's own operator.

**Auditability.** Every claim reduces to ledger state a reviewer can check directly: read the
member's executable, read the owner's tag entry, compare the hashes.

**Failure isolation.** A bug in SFR cannot corrupt on-chain state, because SFR never writes to
the network. The worst case is wrong reporting, which the verification history and the
`indexed_through` field make detectable.

**Security surface.** No keys, no signing path, no upgrade authority. The threat model covers the
API, the decoder, and the database — not custody of assets.

## Protocol boundary

CAP-85 is implemented by the Stellar protocol itself (host functions such as
`create_external_ref_contract` and `update_current_contract_executable_ref`, plus the
`SCV_EXECUTABLE_TAG` storage rules). SFR sits entirely on the consumer side of that boundary: it
reads the state those rules produce.

Primary source:
[CAP-85 — Externally managed contract executables](https://github.com/stellar/stellar-protocol/blob/master/core/cap-0085.md).

## Out of scope for V1 by design

Also absent from this project, deliberately:

- a project token, staking, fees, or any financial contract;
- governance or admin keys;
- generic blockchain exploration (SFR indexes CAP-85 fleets, not the whole ledger);
- transaction relay or mempool services.

If a future version adds capabilities, they will appear in the changelog and in a tagged release.
This site describes **v1.0.0** only.
