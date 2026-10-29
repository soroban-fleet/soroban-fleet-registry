# Drips Wave Appeal Dossier — Soroban Fleet Registry

## 1. Appeal Overview

| Parameter | Value |
|---|---|
| **Project** | **Soroban Fleet Registry (`sfr`)** |
| **Organization** | [`soroban-fleet`](https://github.com/soroban-fleet) |
| **Repository** | [`soroban-fleet/soroban-fleet-registry`](https://github.com/soroban-fleet/soroban-fleet-registry) |
| **Live Documentation** | [https://soroban-fleet.github.io/soroban-fleet-registry/](https://soroban-fleet.github.io/soroban-fleet-registry/) |
| **API Specification** | [`api/openapi.yaml`](./api/openapi.yaml) (OpenAPI 3.1) |
| **Ecosystem Relevance** | Protocol 28 & 29 CAP-85 externally managed Soroban contract executable fleet tracking, verification, and drift prevention. |
| **Current CI Status** | **100% Green** across all workflows (Go test with race detector, Next.js tests, ESLint, Docusaurus build) |

---

## 2. Rejection Analysis & Remediation Strategy

Drips Wave programs evaluate open-source repositories on ecosystem impact, code quality, active maintainer velocity, and immediate operational utility. Initial submissions often suffer from:
1. **Lack of Operational Visibility**: Absence of stream sync endpoints to verify indexer coverage vs. network head.
2. **Missing Compliance / Audit Tooling**: Inability to export membership directories for governance votes and auditors.
3. **Absence of DevSecOps Integration**: Inability to integrate with automated CI pipelines and GitHub Code Scanning for contract drift alerts.
4. **Limited Fleet Discovery Filters**: Inability to filter fleets by operational health status (e.g. `DRIFT` vs `HEALTHY`).

---

## 3. Verifiable Substantive Improvements (Commit Evidence)

Across recent development sprints, the maintainer team has implemented and pushed substantive enhancements directly addressing all operational requirements:

### Milestone 1: Indexer Stream Health & Sync API (`GET /v1/indexer/status`)
- **Commit**: [`d051373`](https://github.com/soroban-fleet/soroban-fleet-registry/commit/d051373) (Resolves Issue #3)
- **Delivered**:
  - Exposes `GET /v1/indexer/status` and `/status` reporting stream checkpoint sequence, network latest ledger height, real-time ledger gap, and sync health classification (`HEALTHY`, `LAGGING`, `SYNCING`, `UNKNOWN`).
  - Documented in OpenAPI 3.1 specification.
  - Full unit test coverage in `internal/api/handlers_test.go`.

### Milestone 2: Fleet Status & Tag Prefix Filtering (`GET /v1/fleets`)
- **Commit**: [`7d13cf1`](https://github.com/soroban-fleet/soroban-fleet-registry/commit/7d13cf1) (Resolves Issue #6)
- **Delivered**:
  - Implemented `FleetFilter` in the repository and service layers supporting `status` (`HEALTHY`, `DRIFT`, `BROKEN_REFERENCE`, `INCOMPLETE`) and `tag_prefix` pattern matching.
  - High-performance PostgreSQL indexed subquery for latest verification status lookup.
  - Updated Next.js frontend TypeScript API client and test suites (`47/47 passing`).

### Milestone 3: Verifiable Membership Export (`sfr fleet export`)
- **Commit**: [`e231097`](https://github.com/soroban-fleet/soroban-fleet-registry/commit/e231097) (Resolves Issue #9)
- **Delivered**:
  - Implemented `sfr fleet export --owner ... --tag ... --format [csv|json]` streaming complete membership directories with contract IDs, owner addresses, tags, WASM bytecode hashes, and ledger observation sequences.
  - Enables governance delegates and security auditors to capture verifiable snapshots of fleet membership.

### Milestone 4: CI/CD Security Integration via OASIS SARIF v2.1.0 (`sfr fleet verify`)
- **Commit**: [`e231097`](https://github.com/soroban-fleet/soroban-fleet-registry/commit/e231097) (Resolves Issue #1)
- **Delivered**:
  - Added `--format sarif` flag to `sfr fleet verify` to generate standard OASIS SARIF v2.1.0 vulnerability reports.
  - Ingestible by GitHub Code Scanning and DevSecOps pipelines to highlight fleet executable drift (`SFR001`), broken references (`SFR002`), and indexer coverage gaps (`SFR003`).
  - Complete unit test coverage validating SARIF schema output.

---

## 4. Test Matrix & Quality Verification

| Component | Test Suite | Pass Count | CI Status |
|---|---|---|---|
| **Go Core & CLI** | `go test -v -race -p 1 ./...` | All packages passing | ✅ **PASS** |
| **Go Code Quality** | `gofmt -l .` & `go vet ./...` | 0 formatting issues, 0 vet warnings | ✅ **PASS** |
| **Next.js Web UI** | Vitest (`npm run test`) | **47 / 47 passing** | ✅ **PASS** |
| **Frontend Lint** | ESLint (`next lint`) | 0 warnings, 0 errors | ✅ **PASS** |
| **Web Build** | Next.js production build (`npm run build`) | Optimized production build verified | ✅ **PASS** |
| **Documentation Site** | Docusaurus 3 (`docusaurus build`) | Static build clean | ✅ **PASS** |
| **Live Docs** | GitHub Pages Deployment | Active & Deployed | ✅ **LIVE** |

---

## 5. Instructions for Drips Wave Appeal Submission

Project maintainers should submit the appeal through the official Drips Wave application:

1. **Log in**: Open [https://drips.network/wave](https://drips.network/wave) with GitHub account `oluwa1to9-web`.
2. **Navigate**: Go to **Maintainers** → **Orgs and Repos**.
3. **Locate**: Find `soroban-fleet/soroban-fleet-registry` in the repository list.
4. **Click Appeal**: Click the **Appeal** button next to the repository.
5. **Paste Justification**:
   ```text
   Appeal for soroban-fleet/soroban-fleet-registry:
   Following our initial review, the Soroban Fleet Registry codebase has undergone substantive, production-grade engineering improvements to fulfill Stellar ecosystem requirements for CAP-85 externally managed contract fleets:
   1. Live Operations & Ingestion Health: Implemented GET /v1/indexer/status and /status with real-time ledger lag detection and stream checkpointing (Issue #3, commit d051373).
   2. Status & Tag Filtering: Enhanced GET /v1/fleets with query filters for operational status (HEALTHY, DRIFT, BROKEN_REFERENCE) and tag prefixes (Issue #6, commit 7d13cf1).
   3. Auditor Snapshot Export: Implemented CLI streaming export supporting CSV and JSON formats (Issue #9, commit e231097).
   4. DevSecOps SARIF Integration: Added OASIS SARIF v2.1.0 reporting to 'sfr fleet verify' for native GitHub Code Scanning drift detection (Issue #1, commit e231097).
   5. 100% Green CI & Live Docs: Full test suite passing with race detector, 47/47 frontend vitest tests passing, and live documentation published at https://soroban-fleet.github.io/soroban-fleet-registry/.
   Full appeal dossier available at https://github.com/soroban-fleet/soroban-fleet-registry/blob/main/APPEAL.md.
   ```
6. **Submit**: Confirm the submission.
