#!/usr/bin/env bash
set -euo pipefail

# scripts/create_issues.sh
# Deterministic issue generator for soroban-fleet-registry

REPO="soroban-fleet/soroban-fleet-registry"
DRY_RUN=false

for arg in "$@"; do
  case $arg in
    --dry-run)
      DRY_RUN=true
      shift
      ;;
  esac
done

TOKEN="${GITHUB_TOKEN:-${GH_TOKEN:-""}}"
if [[ -z "$TOKEN" ]]; then
  echo "Error: GITHUB_TOKEN or GH_TOKEN must be set to run this script." >&2
  exit 1
fi

echo "Target repository: $REPO"
if [[ "$DRY_RUN" == "true" ]]; then
  echo "Mode: DRY-RUN (no issues will be created)"
else
  echo "Mode: LIVE (creating issues on GitHub)"
fi

# Ensure required labels exist
create_label_if_missing() {
  local name="$1"
  local color="$2"
  local desc="$3"

  if [[ "$DRY_RUN" == "true" ]]; then
    return 0
  fi

  curl -s -X POST -H "Authorization: token $TOKEN" \
    -H "Accept: application/vnd.github.v3+json" \
    "https://api.github.com/repos/$REPO/labels" \
    -d "{\"name\":\"$name\",\"color\":\"$color\",\"description\":\"$desc\"}" > /dev/null || true
}

create_label_if_missing "cap85" "5319e7" "CAP-85 protocol mechanics and decoding"
create_label_if_missing "indexer" "0e8a16" "Ingestion pipeline and checkpointing"
create_label_if_missing "verification" "1d76db" "Verification engine and consistency checks"
create_label_if_missing "api" "fbca04" "REST API and HTTP server"
create_label_if_missing "cli" "0052cc" "Command-line interface (sfr)"
create_label_if_missing "web" "b60205" "Next.js web application"
create_label_if_missing "ops" "d93f0b" "Deployment, Docker, and service topology"

# Fetch existing open and closed issues to prevent duplicates
echo "Fetching existing issues..."
EXISTING_TITLES=$(curl -s -H "Authorization: token $TOKEN" \
  -H "Accept: application/vnd.github.v3+json" \
  "https://api.github.com/repos/$REPO/issues?state=all&per_page=100" | \
  grep -o '"title": "[^"]*"' | sed 's/"title": "//;s/"$//' || true)

create_issue() {
  local title="$1"
  local labels="$2"
  local body="$3"

  if echo "$EXISTING_TITLES" | grep -F -x "$title" > /dev/null; then
    echo "[SKIPPED] Issue already exists: \"$title\""
    return 0
  fi

  if [[ "$DRY_RUN" == "true" ]]; then
    echo "=========================================================="
    echo "TITLE: $title"
    echo "LABELS: $labels"
    echo "BODY:"
    echo "$body"
    echo ""
    return 0
  fi

  echo "Creating issue: \"$title\"..."
  
  # Format labels as JSON array
  local labels_json="[]"
  if [[ -n "$labels" ]]; then
    labels_json=$(echo "$labels" | jq -R 'split(",") | map(gsub("^ +| +$";""))')
  fi

  PAYLOAD=$(jq -n \
    --arg title "$title" \
    --arg body "$body" \
    --argjson labels "$labels_json" \
    '{title: $title, body: $body, labels: $labels}')

  URL=$(curl -s -X POST -H "Authorization: token $TOKEN" \
    -H "Accept: application/vnd.github.v3+json" \
    -H "Content-Type: application/json" \
    "https://api.github.com/repos/$REPO/issues" \
    -d "$PAYLOAD" | grep -o '"html_url": "[^"]*"' | head -n 1 | cut -d'"' -f4)

  if [[ -n "$URL" ]]; then
    echo "  -> Created: $URL"
  else
    echo "  -> Failed to create issue \"$title\"" >&2
  fi
}

# 1. Verification Mismatch Details
create_issue \
  "feat(verification): expose per-member mismatch breakdown" \
  "enhancement,verification,api" \
  "### Summary
Extend the verification response structure so that when status is \`DRIFT\`, the API and CLI explicitly enumerate the member contracts that resolved to a different WASM hash alongside the mismatched hashes.

### Why It Matters
When a large fleet drifts during a partial upgrade or malfunction, operators need to immediately identify which contract instances failed to upgrade without inspecting hundreds of members manually.

### Acceptance Criteria
- [ ] Verification response payload includes optional \`mismatches\` array listing member contract ID and resolved WASM hash
- [ ] CLI renders a distinct summary table for mismatched members
- [ ] Engine tests verify that healthy fleets return empty mismatch lists while drifted fleets return exact diverging contracts

### Tech Stack
Go, PostgreSQL, OpenAPI 3.1"

# 2. Fleet Discovery Performance & Indexing
create_issue \
  "perf(indexer): optimize ledger change filtering for CAP-85 instances" \
  "enhancement,indexer,cap85" \
  "### Summary
Improve ledger processing throughput by optimizing XDR contract instance unpacking and change filtering during high-traffic ledger ingestion.

### Why It Matters
On busy networks, ledgers contain hundreds of transaction entries. Filtering non-contract ledger entries before deep XDR traversal significantly improves ingestion speed.

### Acceptance Criteria
- [ ] Add fast entry discriminator filter before full instance data decoding
- [ ] Benchmark ledger processor throughput with 1,000 synthetic entries
- [ ] Ingestion pipeline benchmarks show >=25% speedup in ledger parsing

### Tech Stack
Go, Stellar SDK XDR, Benchmark tests"

# 3. Indexing Lag Visibility & Alerts
create_issue \
  "feat(api): expose indexer stream health and ledger lag endpoint" \
  "enhancement,indexer,api" \
  "### Summary
Add a dedicated \`GET /v1/indexer/status\` endpoint reporting current stream checkpoint, network latest ledger, lag duration, and ingestion health.

### Why It Matters
Operators and load balancers need visibility into whether the background indexer stream is caught up or experiencing RPC stalls without requesting full fleet verification.

### Acceptance Criteria
- [ ] \`GET /v1/indexer/status\` returns current checkpoint sequence and latest network ledger
- [ ] Distinguishes between caught-up stream and lagging stream
- [ ] Documented in OpenAPI specification

### Tech Stack
Go, PostgreSQL, OpenAPI 3.1"

# 4. Historical Replay Tooling
create_issue \
  "feat(cli): add sfr ingest replay subcommand for bounded ledger ranges" \
  "enhancement,cli,indexer" \
  "### Summary
Provide a command-line utility \`sfr ingest replay --start <seq> --end <seq>\` to re-ingest a bounded range of historical ledgers without altering stream checkpoints.

### Why It Matters
Allows operators to test protocol migrations, verify state consistency, or diagnose past upgrades over historical ledger windows.

### Acceptance Criteria
- [ ] CLI subcommand accepts \`--start\` and \`--end\` flags
- [ ] Executes changes within temporary transactional isolation or idempotent upserts
- [ ] Verifies no duplicate rows are created

### Tech Stack
Go, PostgreSQL, CLI"

# 5. Resolver Cache Behavior
create_issue \
  "feat(cap85): implement TTL cache for immutable WASM hashes in resolver" \
  "enhancement,cap85,verification" \
  "### Summary
Introduce an in-memory LRU cache for immutable contract bytecode hashes and owner tag resolutions during verification runs.

### Why It Matters
During batch verification of large fleets (1,000+ members), repeated RPC lookups for identical bytecode entries introduce unnecessary RPC latency.

### Acceptance Criteria
- [ ] Implement thread-safe LRU cache with configurable size and TTL
- [ ] Invalidate tag resolution cache when owner ledger changes are indexed
- [ ] Unit tests verify cache hits avoid duplicate RPC calls

### Tech Stack
Go, In-memory cache"

# 6. API Pagination & Filter Improvements
create_issue \
  "feat(api): add status and tag filtering to fleet listings" \
  "enhancement,api,fleet" \
  "### Summary
Enhance \`GET /v1/fleets\` with query parameters to filter by latest verification status (e.g. \`status=DRIFT\`) and tag prefix matching.

### Why It Matters
Enables operators monitoring large numbers of fleets to instantly filter down to fleets needing operational attention.

### Acceptance Criteria
- [ ] API accepts \`?status=DRIFT\` and \`?tag_prefix=...\` query parameters
- [ ] Efficient PostgreSQL indexed queries supporting the filters
- [ ] Unit tests in \`handlers_test.go\` validating filter combinations

### Tech Stack
Go, PostgreSQL, OpenAPI"

# 7. CLI JSON Schema Validation
create_issue \
  "docs(cli): publish and validate JSON output schemas for sfr commands" \
  "documentation,cli" \
  "### Summary
Document and provide formal JSON schemas for all \`sfr --json\` outputs (fleet verify, contract inspect, members, releases).

### Why It Matters
Automated CI/CD scripts and monitoring agents rely on stable machine-readable CLI outputs.

### Acceptance Criteria
- [ ] JSON schemas documented in \`docs/cli-schemas.md\`
- [ ] Automated CLI integration tests validate stdout against schemas

### Tech Stack
Go, JSON Schema, Markdown"

# 8. Web Verification Drill-Down
create_issue \
  "feat(web): add member drift drill-down modal on verification view" \
  "enhancement,web,verification" \
  "### Summary
Enhance the Next.js verification view (\`/fleets/[owner]/[tag]/verify\`) to allow clicking directly on drifting members to view their individual resolution history.

### Why It Matters
Streamlines debugging during incident triage by linking directly from verification alerts to contract inspector views.

### Acceptance Criteria
- [ ] Divergent members list links directly to contract inspection routes
- [ ] Accessible modal or expander showing resolved bytecode comparison
- [ ] Acceptance tests in \`web/__tests__/acceptance.test.tsx\` updated

### Tech Stack
Next.js 14, React 18, TypeScript"

# 9. Fleet Export Utility
create_issue \
  "feat(cli): add sfr fleet export command supporting JSON and CSV" \
  "enhancement,cli,fleet" \
  "### Summary
Add \`sfr fleet export --owner ... --tag ... --format [json|csv]\` to export the complete membership directory and upgrade history of a fleet.

### Why It Matters
Protocol auditors and governance bodies require verifiable snapshots of fleet membership for compliance and release validation.

### Acceptance Criteria
- [ ] Support JSON array and CSV format outputs
- [ ] Stream results to handle large fleets (10,000+ members) without excessive memory usage
- [ ] End-to-end test verifying output completeness

### Tech Stack
Go, CSV/JSON streaming"

# 10. Audit Report Generation
create_issue \
  "feat(verification): generate markdown compliance audit report" \
  "enhancement,verification,cli" \
  "### Summary
Add a flag \`--report-markdown\` to \`sfr fleet verify\` that produces a formatted audit report including ledger hashes, timestamp, and verification breakdown.

### Why It Matters
Provides human-readable, cryptographically verifiable release artifacts that can be attached to GitHub releases or governance proposals.

### Acceptance Criteria
- [ ] Generates clean Markdown table summary with verification status, ledger range, and hashes
- [ ] Includes indexer coverage and verification timestamp
- [ ] Tests verify deterministic formatting

### Tech Stack
Go, Markdown templates"

# 11. API Rate Limiting & Protection
create_issue \
  "feat(api): add configurable IP rate limiting middleware" \
  "enhancement,api,ops" \
  "### Summary
Introduce token-bucket rate limiting middleware for the REST API to protect public instances against denial-of-service and RPC exhaustion.

### Why It Matters
Public verification endpoints trigger live RPC queries; protecting the API prevents third-party quota exhaustion on upstream Stellar nodes.

### Acceptance Criteria
- [ ] Configurable rate limit via \`SFR_RATE_LIMIT_RPS\` and burst capacity
- [ ] Returns HTTP 429 Too Many Requests with \`Retry-After\` header
- [ ] Configurable exclusion for internal private subnets

### Tech Stack
Go, HTTP middleware"

# 12. Prometheus Metrics Endpoint
create_issue \
  "feat(ops): add Prometheus metrics exporter for indexer and verification" \
  "enhancement,ops,indexer" \
  "### Summary
Add an optional Prometheus metrics server on \`SFR_METRICS_ADDR\` exposing ingestion latency, processed ledgers, verification counts, and status rates.

### Why It Matters
Enables production monitoring, Grafana dashboards, and alerting on lagging indexer checkpoints.

### Acceptance Criteria
- [ ] Expose standard counters: \`sfr_ledgers_processed_total\`, \`sfr_verification_runs_total{status=...}\`
- [ ] Histogram for ledger processing duration
- [ ] Tested with mock Prometheus scraper

### Tech Stack
Go, Prometheus client"

# 13. Operational Alerting Runbook
create_issue \
  "docs(ops): create operational alerting and runbook guide" \
  "documentation,ops" \
  "### Summary
Create \`docs/runbook.md\` documenting critical alerts, failure modes, database recovery steps, and indexer resynchronization procedures.

### Why It Matters
Provides site-reliability engineers and protocol maintainers clear procedures during network halts, database outages, or RPC migrations.

### Acceptance Criteria
- [ ] Runbook covers database restores, stream checkpoint reset, and RPC fallback configuration
- [ ] Step-by-step incident response playbook

### Tech Stack
Markdown"

# 14. Deterministic Fixture Generator Tooling
create_issue \
  "chore(fixtures): add CLI tool to record and serialize live testnet scenarios" \
  "help wanted,fixtures,testing" \
  "### Summary
Build an internal developer utility to capture live ledger entries from testnet and serialize them into deterministic offline test fixtures in \`fixtures/\`.

### Why It Matters
Allows contributors to reproduce complex real-world fleet upgrade bugs locally without requiring live network access.

### Acceptance Criteria
- [ ] Tool takes ledger range and contract address and writes fixture JSON
- [ ] Integrates with existing fixture test suite in \`fixtures/fixtures_test.go\`

### Tech Stack
Go, Stellar SDK XDR"

# 15. Automated Protocol 29 Compatibility Tests
create_issue \
  "test(cap85): add automated protocol migration test harness" \
  "testing,cap85" \
  "### Summary
Develop an automated test harness verifying that CAP-85 decoding and resolution maintain compatibility across Protocol 28, Protocol 29, and future protocol upgrades.

### Why It Matters
Ensures that ongoing Stellar protocol upgrades do not introduce regressions into external reference decoding or storage entry structures.

### Acceptance Criteria
- [ ] Matrix tests covering Protocol 28 and Protocol 29 ledger close metadata fixtures
- [ ] Automated CI execution verifying backward compatibility

### Tech Stack
Go, Stellar SDK"

# 16. Historical Archive Recovery Tooling
create_issue \
  "feat(indexer): support historical backfill from Stellar ledger history archives" \
  "enhancement,indexer" \
  "### Summary
Implement direct ledger ingestion from Stellar Ledger History Archives (S3/HTTP archives) in addition to live RPC \`getLedgers\`.

### Why It Matters
Allows fast backfilling of millions of historical ledgers upon initial deployment without overwhelming live RPC nodes.

### Acceptance Criteria
- [ ] Supports reading checkpointed \`history-archive\` buckets
- [ ] Seamlessly transitions from archive backfill to live RPC streaming
- [ ] Benchmark backfill speed comparison

### Tech Stack
Go, Stellar Archive Ingestion"

echo "Issue generation script complete."
