# Security Policy

## 1. Scope & System Boundary

`soroban-fleet-registry` is an **indexer, observer, and deterministic verification service** for Soroban CAP-85 contract executables.

### In Scope
- API service security, panic resistance, and memory safety.
- Ledger change ingestion and Stellar XDR decoding robustness.
- Data integrity, database transaction boundaries, and SQL injection prevention.
- Checkpoint persistence and stream recovery logic.
- Input validation and StrKey contract address parsing.
- Secret leaks or credential exposure in repositories and container images.
- Next.js frontend dependencies and client-side security headers.

### Out of Scope
- Security vulnerabilities in third-party Soroban smart contracts or external Stellar validators.
- Smart contract logic bugs residing on the Stellar network outside the indexing layer.
- Upstream Stellar RPC availability or network partition events.

> **Disclaimer**: `soroban-fleet-registry` never holds private keys, signs transactions, manages funds, or executes blockchain upgrades. It is a read-only observer.

---

## 2. Supported Versions

| Version | Supported |
|---|---|
| `v1.0.x` | :white_check_mark: Active security support |
| `< v1.0.0` | :x: Not supported |

---

## 3. Reporting a Vulnerability

If you discover a security vulnerability or potential data-integrity exploit in `soroban-fleet-registry`:

1. **Do NOT report security vulnerabilities through public GitHub issues or discussions.**
2. Please use **GitHub Private Vulnerability Reporting** directly via the repository's [Security Advisories](https://github.com/soroban-fleet/soroban-fleet-registry/security/advisories) page.
3. If GitHub Private Vulnerability Reporting is unavailable, contact the repository maintainers through a private issue request referencing security triage.

### What to Include in Your Report
- A detailed description of the vulnerability.
- Steps to reproduce or proof-of-concept code/input XDR.
- Affected components (e.g. `internal/cap85`, `internal/ingest`, `internal/api`, `web/`).
- Potential impact on fleet indexing correctness or service availability.

---

## 4. Response & Remediation Process

- **Acknowledgement**: The maintainers will acknowledge receipt within 48 hours.
- **Triage**: An assessment of severity and impact will follow within 5 business days.
- **Remediation**: A fix will be developed in a private branch, verified against regression fixtures, and released with an advisory.
