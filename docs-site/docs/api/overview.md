# API overview

SFR exposes a read-only REST API described by OpenAPI 3.1.

Specification: [`api/openapi.yaml`](https://github.com/soroban-fleet/soroban-fleet-registry/blob/main/api/openapi.yaml)

## Properties

- **OpenAPI 3.1** — the YAML file is the contract; response examples on this site follow its
  schemas.
- **JSON only** — every response is `application/json; charset=utf-8`.
- **Read-only** — only `GET` and `OPTIONS` are routed. There are no POST/PUT/PATCH/DELETE
  handlers and no authentication, because no state-changing operation exists.
- **Pagination** — collection endpoints accept `limit` and `offset`.
- **Structured errors** — failures return `{"error":{"code","message"}}`.

## Base URL

| Environment | Base |
|---|---|
| Local | `http://localhost:8080` |
| Production | your HTTPS origin (TLS terminated at the load balancer) |

All endpoints are prefixed with `/health` (unversioned alias) or `/v1/…`.

## Endpoints

| Method | Path | Purpose |
|---|---|---|
| GET | `/health`, `/v1/health` | Service liveness |
| GET | `/v1/fleets` | List fleets |
| GET | `/v1/fleets/{owner}/{tag}` | Fleet details |
| GET | `/v1/fleets/{owner}/{tag}/members` | Member list |
| GET | `/v1/fleets/{owner}/{tag}/history` | Verification history |
| GET | `/v1/fleets/{owner}/{tag}/releases` | Release history |
| GET | `/v1/fleets/{owner}/{tag}/verify` | Run verification |
| GET | `/v1/contracts/{contract_id}` | Inspect a contract |

Detail pages: [Fleets](./fleets.md) · [Members](./members.md) · [History](./history.md) ·
[Verification](./verification.md) · [Contracts](./contracts.md)

## Pagination

| Parameter | Default | Max | Notes |
|---|---:|---:|---|
| `limit` | 20 | 100 | Values above 100 are clamped; non-numeric values are ignored |
| `offset` | 0 | — | Negative values are ignored |

Collection response envelope:

```json
{
  "data": [ ... ],
  "meta": { "total": 137, "limit": 20, "offset": 0 }
}
```

Single-resource response envelope:

```json
{ "data": { ... } }
```

## Errors

```json
{
  "error": {
    "code": "FLEET_NOT_FOUND",
    "message": "Fleet was not found"
  }
}
```

| Code | HTTP | When |
|---|---:|---|
| `INVALID_ARGUMENT` | 400 | Bad address, invalid tag, malformed `expected_wasm` hex |
| `FLEET_NOT_FOUND` | 404 | Fleet does not exist in the index |
| `INTERNAL_ERROR` | 500 | Database failure, unexpected handler error |

Database messages are generic; connection strings and stack traces are never returned.

## CORS

Controlled by `SFR_CORS_ALLOWED_ORIGINS` (comma-separated). Allowed methods are `GET, OPTIONS`;
allowed headers are `Content-Type, Accept`; preflight max age is 86400 seconds. With the variable
unset, any origin is accepted — suitable for local development only.

```bash
SFR_CORS_ALLOWED_ORIGINS=https://fleet.example.com
```

## Health

```bash
curl -s http://localhost:8080/health
```

```json
{
  "status": "UP",
  "service": "soroban-fleet-registry",
  "timestamp": "2026-10-07T12:00:00Z"
}
```

`status: UP` means the HTTP process is serving. It says nothing about indexer coverage or fleet
state — see [Operations → Health](../operations/health.md).

## Byte encoding

Fields described as `format: byte` (`current_wasm_hash`, `wasm_hash`, `expected_wasm_hash`,
`old_wasm_hash`, `new_wasm_hash`) are **base64**-encoded 32-byte values in JSON. Hexadecimal
forms appear only in CLI text output and in query parameters such as `expected_wasm`.
