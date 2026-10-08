# Fleet

A **fleet** is the set of Soroban contract instances that share one externally managed executable
reference.

## Identity

```text
Fleet identity = (owner_address, tag)
```

| Component | Meaning |
|---|---|
| `owner_address` | StrKey contract address of the contract that stores the active WASM hash |
| `tag` | Exact, case-sensitive string key used inside the owner's storage |

The pair is the primary key of the `fleets` table:

```sql
PRIMARY KEY (owner_address, tag)
```

Two different owners using the same tag are two different fleets. The same owner using two tags
publishes two fleets.

## Terms

- **Owner**: the contract holding the canonical hash under a tag.
- **Tag**: the exact string key, for example `vault-v1`.
- **Member**: a contract instance whose executable is an external reference to `(owner, tag)`.
- **Current WASM**: the 32-byte hash the owner publishes under the tag right now.
- **Release**: an observed change of that hash from one value to another.
- **Verification**: a recorded comparison of member executables against an expected hash.

## Example

```text
owner = C...
tag   = vault-v1

members:
  Contract A
  Contract B
  Contract C
```

At ledger `N`, the owner publishes WASM A:

```text
vault-v1  ──▶  WASM A
```

Later the owner publishes WASM B:

```text
vault-v1  ──▶  WASM B
```

What changed and what did not:

| | Before | After |
|---|---|---|
| Fleet identity | `(C..., vault-v1)` | `(C..., vault-v1)` |
| Active WASM | A | B |
| Members | A, B, C | A, B, C |
| Release record | — | WASM A → WASM B at ledger `M` |

The fleet identity stays the same across releases. Membership and releases are what change.

## Membership lifecycle

The indexer never deletes a member row. A member is:

- **Active** when its executable references `(owner, tag)` as of the last indexed ledger;
- **Inactive** when it later points at a different fleet, a direct WASM, or no longer exists.

`first_seen_ledger` is preserved across re joins; `last_seen_ledger` records the most recent
ledger in which the membership was observed. `member_count` counts active members only.

## Data

| Table | Contents |
|---|---|
| `fleets` | identity, `current_wasm_hash`, member count, first/last seen ledger |
| `fleet_members` | per-contract membership, WASM hash, active flag, ledger bounds |
| `fleet_releases` | old/new hash, ledger, transaction hash, observation time |

Schema: [`migrations/000001_init_schema.up.sql`](https://github.com/soroban-fleet/soroban-fleet-registry/blob/main/migrations/000001_init_schema.up.sql).

Next: [External executable](./external-executable.md) ·
[Verification states](./verification-states.md)
