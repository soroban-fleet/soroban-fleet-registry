# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [1.0.0] - 2026-10-07

### Added
- **CAP-85 Reference Resolution**: Comprehensive XDR decoder and resolver for `CONTRACT_EXECUTABLE_EXTERNAL_REF` with exact tag preservation.
- **Historical Ingestion Pipeline**: Streaming ingestion from `LedgerCloseMeta` extracting contract executable and instance changes.
- **Atomic Checkpoint Management**: Transactional stream checkpointing with safe rollback and resume guarantees.
- **Fleet & Membership Engine**: Persistent tracking of fleet identities `(owner, tag)`, active/inactive member lifecycle, and chronological release detection.
- **Deterministic Verification Engine**: Multi-tier consistency evaluation reporting `HEALTHY`, `DRIFT`, `BROKEN_REFERENCE`, `INCOMPLETE`, and `UNKNOWN`.
- **CLI Utility (`sfr`)**: Full suite of subcommands (`fleet`, `contract`, `migrate`, `ingest`, `api`) with standard UNIX exit codes.
- **REST API**: OpenAPI 3.1 compliant HTTP server with pagination, service health check (`/health`), and configurable CORS.
- **Next.js Web Application**: Responsive read-only explorer featuring fleet directories, member tables, release history timelines, drift inspection, and contract lookup.
- **Containerization & Deployment**: Multi-stage unprivileged Docker build and hardened Docker Compose configuration.

### Security
- Explicit read-only architectural boundary: no private key handling, wallet connections, or transaction signing.
- Sanitized environment configuration templates and credential auditing.
