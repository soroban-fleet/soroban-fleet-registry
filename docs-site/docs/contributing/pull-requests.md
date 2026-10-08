# Contributing: Pull requests

## Expectations

From [`CONTRIBUTING.md`](https://github.com/soroban-fleet/soroban-fleet-registry/blob/main/CONTRIBUTING.md):

1. **One logical change per PR.** Do not mix unrelated refactors with a feature.
2. **Tests included.** New behavior and bug fixes need unit or integration coverage.
3. **No placeholders.** No stub implementations, `TODO` markers in production paths, or mock
   fallbacks.
4. **Clean git state.** No secrets, `.env` files, or temporary artifacts.

## Checklist

```text
[ ] Branch from main, rebased if main moved
[ ] make fmt-check && make vet && make test pass
[ ] ./scripts/verify_e2e.sh passes
[ ] cd web && npm run lint && npm test && npm run build pass (web/ changes)
[ ] cd docs-site && npm run build passes (docs-site/ changes)
[ ] api/openapi.yaml updated if endpoints/schemas changed
[ ] migrations added if schema changed (with working down migration)
[ ] README/docs updated where user-visible behavior changed
[ ] Conventional commit message
```

## Commit message format

```text
<type>(<scope>): <subject>
```

Examples:

```text
feat(verification): reject HEALTHY when member set is empty
fix(ingest): preserve tag casing from external ref
docs(docs): add verification state reference
ci(docs): add documentation build checks
```

Types and scopes: [Workflow → Commit messages](./workflow.md#commit-messages).

## Review focus

Reviewers check, in order:

1. **Scope** — does it stay read-only? No signing, no on-chain writes, no invented features.
2. **Verification semantics** — status classification rules must remain deterministic and
   conservative; `HEALTHY` must stay blocked by incomplete coverage or an empty member set.
3. **Correctness with data** — transaction boundaries, idempotency, pagination, error handling.
4. **Tests** — do they fail without the change?
5. **Docs** — user-visible changes documented, `api/openapi.yaml` in sync.

## Updating the OpenAPI spec

Any change to routes, parameters, or response shapes must update
[`api/openapi.yaml`](https://github.com/soroban-fleet/soroban-fleet-registry/blob/main/api/openapi.yaml)
in the same PR. The spec is the contract the frontend types are generated from
(`openapi-typescript` in `web/package.json`).

## Updating the schema

- Add a numbered migration in `migrations/` with a working `.down.sql`.
- CI runs `sfr migrate down && sfr migrate up` as a smoke check — both directions must pass.
- Document new tables in [State model](../architecture/state-model.md).

## Documentation changes

Docs live in `docs-site/docs/`. Run `cd docs-site && npm run build` — broken links, broken
anchors, and broken Markdown links fail the build. Keep the writing rules: direct language, real
commands, no marketing claims, no invented numbers.

## After merge

- The CI workflow runs on `main`; a red build should be fixed promptly.
- Release-worthy changes are noted in `CHANGELOG.md` following Keep a Changelog.
