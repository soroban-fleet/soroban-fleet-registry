# Operations: Troubleshooting

Format for each entry: **Symptom → Cause → How to inspect → Safe fix.**

## Frontend cannot reach the API

| | |
|---|---|
| **Symptom** | Pages show the network error state: `The Fleet Registry API could not be reached: …` (`NETWORK_ERROR`) |
| **Cause** | `sfr api` not running, wrong port, or wrong `NEXT_PUBLIC_API_BASE_URL` / `NEXT_PUBLIC_SFR_API_URL` |
| **Inspect** | `curl -s http://localhost:8080/health`; check `web/.env.local`; browser devtools → Network tab → request URL |
| **Safe fix** | Start the API; point the frontend at the correct origin. Remember `NEXT_PUBLIC_*` values are baked at build time — rebuild after changing them |

## API cannot reach PostgreSQL

| | |
|---|---|
| **Symptom** | HTTP `500` with `{"error":{"code":"INTERNAL_ERROR","message":"Failed to …"}}`; CLI: `ping database: …` |
| **Cause** | Database down, wrong `SFR_DATABASE_URL`, network/SSL mismatch, migrations not applied (`relation … does not exist`) |
| **Inspect** | `docker compose ps`; `pg_isready`; `psql "$SFR_DATABASE_URL" -c '\dt'`; API logs |
| **Safe fix** | Start PostgreSQL, correct the connection string, run `sfr migrate up`. Credentials and stack traces are intentionally absent from API responses |

## RPC unavailable

| | |
|---|---|
| **Symptom** | Indexer logs `fetch ledgers error, will retry`; verification fails or reports `INCOMPLETE`; contract inspection returns no live data |
| **Cause** | `SFR_RPC_URL` wrong, endpoint down, rate-limited, or network blocked |
| **Inspect** | `curl -s "$SFR_RPC_URL"`; check provider status; indexer log cadence |
| **Safe fix** | Restore or swap the endpoint, restart `sfr ingest`. Checkpoints do not advance while RPC fails, so nothing is lost. Re-run verification after catch-up |

## Indexer not advancing

| | |
|---|---|
| **Symptom** | `indexer_checkpoints.ledger` unchanged; no `ledger_processed` log lines; verification stays `INCOMPLETE` |
| **Cause** | Process exited (config `4`, runtime `5`), RPC down, database write failing, or already caught up with the network |
| **Inspect** | Process running? Logs? `SELECT ledger, updated_at FROM indexer_checkpoints WHERE stream='main';` compare with network latest ledger |
| **Safe fix** | Fix the reported error and restart. If caught up, no action needed — an idle indexer polling an empty batch is normal |

## Verification is INCOMPLETE

| | |
|---|---|
| **Symptom** | `status: INCOMPLETE`, exit code `3` |
| **Cause** | `indexed_through < latest_ledger`, missing members during live resolution, or zero indexed members |
| **Inspect** | Response fields `indexed_through`, `total_members`, `missing_members`; checkpoint row; indexer logs |
| **Safe fix** | Let the indexer catch up (or fix RPC so resolution works), then re-verify. Never override or treat as healthy |

## Fleet reference cannot resolve (BROKEN_REFERENCE)

| | |
|---|---|
| **Symptom** | `status: BROKEN_REFERENCE`, exit code `2` |
| **Cause** | Owner contract has no entry for the tag, entry value is malformed, or owner does not exist on-chain |
| **Inspect** | `sfr contract inspect --contract <OWNER>`; look at the owner's ledger entries in an explorer; `sfr fleet inspect` for the indexed view |
| **Safe fix** | Correct the reference in the owner contract's storage — an on-chain action outside this system. SFR will index the correction on the next relevant ledger |

## CORS failure

| | |
|---|---|
| **Symptom** | Browser console: `Access-Control-Allow-Origin` missing/mismatched; API works with `curl` but not the browser |
| **Cause** | `SFR_CORS_ALLOWED_ORIGINS` does not list the frontend origin (or contains typos/whitespace) |
| **Inspect** | `curl -si -H "Origin: https://fleet.example.com" http://localhost:8080/health \| grep -i access-control` |
| **Safe fix** | Add the exact origin, comma-separated, restart the API. Preflight `OPTIONS` returns `204` when allowed |

## Migration failure

| | |
|---|---|
| **Symptom** | `Migration up failed: …` or `Migration down failed: …` |
| **Cause** | Database permissions, partially applied schema, or a bad migration |
| **Inspect** | Run with a privileged role; check applied state (`schema_migrations`/migrator output); read the error text |
| **Safe fix** | `sfr migrate down` to the last known good state, fix the migration, `sfr migrate up`. Do not hand-edit schema. CI proves the down/up cycle works |

## Exit-code quick reference

| Code | Meaning | Typical area |
|---:|---|---|
| 0 | success / `HEALTHY` | — |
| 1 | `DRIFT` | fleet data |
| 2 | `BROKEN_REFERENCE` | on-chain reference |
| 3 | `INCOMPLETE` | indexer coverage |
| 4 | invalid input / missing configuration | flags, env vars |
| 5 | internal error / `UNKNOWN` | database, RPC, unclassified |

## Still stuck?

1. Capture command output with `--json` and the exact environment variable **names** (not values).
2. Check [Recovery](./recovery.md) for restart procedures.
3. Open an issue: https://github.com/soroban-fleet/soroban-fleet-registry/issues
   (security issues go to [Security](../contributing/security.md) instead).
