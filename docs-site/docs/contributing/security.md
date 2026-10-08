# Contributing: Security

Source of truth: [`SECURITY.md`](https://github.com/soroban-fleet/soroban-fleet-registry/blob/main/SECURITY.md).

## Report privately

> Do **not** report security vulnerabilities through public GitHub issues or discussions.

1. Use **GitHub Private Vulnerability Reporting** on the repository's
   [Security Advisories](https://github.com/soroban-fleet/soroban-fleet-registry/security/advisories)
   page.
2. If that is unavailable, contact the maintainers through a private channel referencing security
   triage.

Include:

- a detailed description of the vulnerability;
- reproduction steps or a proof of concept (including input XDR where relevant);
- affected components — `internal/cap85`, `internal/ingest`, `internal/api`, `internal/verification`,
  `web/`, migrations;
- potential impact on indexing correctness or service availability.

## What is in scope

- API security, panic resistance, memory safety.
- Ledger ingestion and Stellar XDR decoding robustness.
- Data integrity, transaction boundaries, SQL injection.
- Checkpoint persistence and stream recovery logic.
- Input validation and StrKey parsing.
- Secret leaks in repositories or container images.
- Frontend dependencies and client-side headers.

## What is out of scope

- Vulnerabilities in third-party Soroban contracts or external validators.
- Smart contract logic bugs on the network outside the indexing layer.
- Upstream Stellar RPC availability or network partitions.

## System boundary

`Soroban Fleet Registry` never holds private keys, signs transactions, manages funds, or executes
on-chain upgrades. It is a read-only observer. Reports about "funds at risk in SFR" are
misdirected — see [Soroban boundary](../architecture/soroban-boundary.md).

## Supported versions

| Version | Supported |
|---|---|
| `v1.0.x` | Yes |
| `< v1.0.0` | No |

## Response process

| Stage | Target |
|---|---|
| Acknowledgement | 48 hours |
| Triage (severity/impact) | 5 business days |
| Remediation | private branch, regression fixtures, advisory release |

## Credential hygiene

- Never commit `.env` files or real credentials.
- `NEXT_PUBLIC_*` values are public by construction — keep secrets out of them.
- Report accidental secret exposure through the private channel above.
