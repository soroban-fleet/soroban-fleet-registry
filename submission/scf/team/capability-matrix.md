# Team Capability Matrix

This matrix maps required engineering capabilities for building and maintaining Soroban Fleet Registry against verified repository evidence.

---

## Capability Verification Table

| Requirement | Evaluation Criteria | Verifiable Repository Evidence | Primary Owner | Status |
|---|---|---|---|---|
| **Go Backend Systems** | High-concurrency systems, race-free architecture, idiomatic Go. | `internal/ingest`, `internal/api`, passing `go test -v -race -p 1 ./...` | `oluwa1to9-web` | ✅ **VERIFIED** |
| **Soroban & CAP-85 Protocol** | Binary XDR decoding, SCVal parsing, StrKey address conversion. | `internal/cap85`, CAP-85 decoding unit tests | `oluwa1to9-web` | ✅ **VERIFIED** |
| **Ledger Ingestion & Checkpoints** | Reliable catchup, RPC batching, atomic checkpoint updates, idempotent resume. | `internal/ingest/pipeline.go`, `TestPipeline_IdempotentReplayAndResume` | `oluwa1to9-web` | ✅ **VERIFIED** |
| **PostgreSQL Database** | Normalized relational schema, two-way migrations, connection safety. | `migrations/000001_initial_schema.up.sql`, `TestMigrations_Up_Idempotency_Down` | `oluwa1to9-web` | ✅ **VERIFIED** |
| **Frontend & UI/UX** | Server components, responsive design, accessibility, comprehensive test coverage. | `web/`, Vitest suite (11 files, 47 tests passed) | `oluwa1to9-web` | ✅ **VERIFIED** |
| **CLI Development** | Ergonomic POSIX commands, structured exit codes, JSON support. | `cmd/sfr/main.go`, `TestCLI_ExitCodesAndCommands` | `oluwa1to9-web` | ✅ **VERIFIED** |
| **CI/CD & Release Automation** | Automated multi-runtime workflows, containerized testing, static analysis. | `.github/workflows/ci.yml`, `.github/workflows/docs-pages.yml` | `oluwa1to9-web` | ✅ **VERIFIED** |
| **Documentation Engineering** | Clear, navigable documentation site, zero broken links, accurate schemas. | `docs-site/`, Docusaurus 3 build passing | `oluwa1to9-web` | ✅ **VERIFIED** |
| **Production DevOps / Hosting** | Cloud infrastructure provisioning, SSL/TLS, reverse proxy, DDoS mitigation. | Proposed topology in `docs/deployment.md`; pending mainnet cluster setup | *Input required* | ⚠️ **PARTIAL** (Local & containerized verified; cloud cluster pending) |
| **Ecosystem Outreach & Partnerships** | Conducting builder interviews, onboarding protocols, integrating feedback. | Framework ready in `submission/scf/validation/` | *Input required* | ❌ **OUTSTANDING** (Interviews pending) |
