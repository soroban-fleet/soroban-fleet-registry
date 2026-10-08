# Guide: Use the web app

The `web/` directory contains a read-only Next.js interface over the REST API.

## Run it

```bash
cd web
npm install
npm run dev
```

Open `http://localhost:3000`. The app calls the API at `http://localhost:8080` by default
(`NEXT_PUBLIC_API_BASE_URL`, or `NEXT_PUBLIC_SFR_API_URL` — the first one set wins).

Make sure `sfr api` is running, otherwise every page shows a connection error state.

## Routes

| Route | Purpose |
|---|---|
| `/` | Dashboard: service health and quick links |
| `/fleets` | Fleet directory — searchable, paginated list with health badges |
| `/fleets/{owner}/{tag}` | Fleet overview: identity, current WASM, member counts, stats |
| `/fleets/{owner}/{tag}/members` | Paginated member table with drift highlighting |
| `/fleets/{owner}/{tag}/history` | Timeline of releases and past verifications |
| `/fleets/{owner}/{tag}/verify` | On-demand verification with matched/drifted breakdown |
| `/contracts/{contractId}` | Contract inspector: executable kind and fleet membership |

`{owner}` and `{tag}` are URL-encoded path segments. Example:

```text
/fleets/CA3D5KRYMCMCZVACQY3CGINPVPUM2OAL6YHT3ZVM5C2MUMAA252AAAAB/vault-v1/members
```

Unknown routes render the `not-found` page.

## Page by page

### Fleet directory — `/fleets`

Lists indexed fleets with member counts, current WASM, and a status badge. Supports `limit` and
`offset` query parameters for pagination.

### Fleet overview — `/fleets/{owner}/{tag}`

The hub for one fleet: identity, member count, first/last seen ledgers, current WASM hash, and
links to the members, history, and verify subviews.

### Members — `/fleets/{owner}/{tag}/members`

Table of member contracts with WASM hash, first-seen ledger, and active flag. Drifted contracts
are highlighted. Query parameters:

```text
?active_only=true|false   (default true)
?limit=20&offset=0
```

### History — `/fleets/{owner}/{tag}/history`

Chronological view of releases (old hash → new hash with ledger and transaction) and recorded
verification results.

### Verify — `/fleets/{owner}/{tag}/verify`

Triggers a verification through `GET /v1/fleets/{owner}/{tag}/verify` and renders the counts
(total, matching, mismatched, missing), `indexed_through`, and the status. To check against a
specific hash, pass it in the query string:

```text
/fleets/{owner}/{tag}/verify?expected_wasm=<64-hex-chars>
```

### Contract inspector {#contract-inspector}

Route: `/contracts/{contractId}`.

Shows whether the contract is a fleet member, its fleet identity, and the live resolved
executable kind and hash.

## What the UI never does

- no wallet connection, no signing, no transactions;
- no direct database or RPC access — only the REST API;
- no re-implementation of verification logic — statuses from the API are displayed verbatim,
  including `INCOMPLETE`, which is never rendered as healthy.

## Tests

```bash
cd web
npm test          # vitest run — component and acceptance suites
npm run lint      # next lint
npm run build     # production build
```

See [Development → Testing](../development/testing.md).

## Production build

```bash
cd web
npm ci
NEXT_PUBLIC_API_BASE_URL=https://api.example.com npm run build
npm run start
```

`NEXT_PUBLIC_*` values are embedded at build time — set them before `npm run build`.
