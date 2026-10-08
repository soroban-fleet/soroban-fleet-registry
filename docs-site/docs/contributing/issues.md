# Contributing: Issues

Report bugs, propose features, or flag documentation gaps through GitHub Issues:
https://github.com/soroban-fleet/soroban-fleet-registry/issues

## Templates

The repository ships three issue templates:

| Template | Title prefix | Label | Use for |
|---|---|---|---|
| **Bug Report** | `fix:` | `bug` | Incorrect behavior with reproduction steps |
| **Feature Request** | `feat:` | `enhancement` | New capability or improvement |
| **Documentation Request** | `docs:` | `documentation` | Unclear, outdated, or missing docs |

Security vulnerabilities do **not** go through public issues — see
[Security](./security.md).

## Good bug reports include

1. **What happened** vs. **what you expected**.
2. **Reproduction steps** — exact commands, flags, and endpoints.
3. **Environment** — OS, Go/Node versions, network (`testnet`/`pubnet`), local or remote RPC.
4. **Evidence** — command output (prefer `--json`), API response bodies, relevant log lines.
5. **Version** — release tag or commit SHA.

```text
$ ./bin/sfr fleet verify --owner <OWNER> --tag <TAG> --json
{ ... }
exit=3
```

Do **not** paste secrets: database passwords, RPC API keys, or any private key. Redact connection
strings down to host and database name.

## Good feature requests state

- The operational problem, not the implementation you have in mind.
- Why existing commands/endpoints do not solve it.
- Whether the proposal stays within the read-only scope. Wallet connections, signing, and on-chain
  mutations are explicitly out of scope (the Feature Request template says so).

## Documentation gaps

Use the Documentation Request template and point at the page — the site URL or the file path under
`docs-site/docs/`. Documentation PRs are welcome and do not require code changes.

## Labels and triage

Bugs carry the `bug` label; enhancements `enhancement`; doc requests `documentation`. Maintainers
triage by severity and scope. Duplicate reports are linked rather than deleted.

## Before opening

- Search existing issues for the same symptom.
- Check whether the behavior is documented (it may be intended) — for verification semantics see
  [Statuses](../reference/statuses.md).
- For operations questions, [Troubleshooting](../operations/troubleshooting.md) covers the common
  cases.
