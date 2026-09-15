# TASK-260908-1bpra2 producer handoff revalidation

Preserved uncommitted candidate inspected; no product edits needed in this run.
Source pin: agents-infra 459742ea67e3c6b84169520b92d74fe7f73e3002.
Fresh origin HEAD advertisement, fetched main and selected worktree HEAD all equal
3125e491275c23c0e8b7f4da5ebfd3fe900220a0.
Curator parser checkout: f750344f5e6fe3ef539485981c12e24240cca1ed.

| Package | Fields | Release gate | Parser oracle |
| --- | --- | --- | --- |
| figma | schema 1, version 1.0.0, http, https://mcp.figma.com/mcp, env_names=[] | 0 | 0 ACCEPT |
| safari | schema 1, version 1.0.0, stdio, safaridriver-mcp, args=[--mcp], env_names=[] | 0 | 0 ACCEPT |

Commands rerun directly as standalone processes in zsh, without pipelines:

| Command | Actual exit | Result |
| --- | --- | --- |
| scripts/validate.sh | 0 | 2/2 manifests |
| python3 -m unittest discover -s tests -v | 0 | 6 tests, 3/3 narrowing mutants detected |
| sh -n bin/safaridriver-mcp scripts/validate.sh | 0 | shell syntax |
| git diff --check | 0 | whitespace |
| go run -mod=mod . ../../packages/figma/agent-mcp.json ../../packages/safari/agent-mcp.json | 0 | both ACCEPT |
| go run -mod=mod . absolute.json | 1 | expected bare-command refusal |
| go run -mod=mod . token.json | 0 | structural parser accepts literal argument; release gate refuses |
| go run -mod=mod . unknown.json | 1 | expected unknown-field refusal |

Go commands ran in .temp/oracle using internal/contextpkg.ParseMCP, the real
MCP declaration parser (internal/mcp checks configured servers).

```text
../../packages/figma/agent-mcp.json: ACCEPT
../../packages/safari/agent-mcp.json: ACCEPT
absolute.json: REJECT mcp_declaration_invalid: server.command must be a bare executable name without a path separator
exit status 1
token.json: ACCEPT
unknown.json: REJECT mcp_declaration_invalid: unknown field server.token
exit status 1
```

| Narrowing mutant | Original gate exit | Narrowed gate exit |
| --- | --- | --- |
| command check narrowed to ./ and exact-name guard subsumed | 1 | 0 |
| argument check narrowed to first argument only | 1 | 0 |
| unknown-field check narrowed to required-key subset | 1 | 0 |

Tests drive scripts/validate.sh in isolated copies. Wrapper tests substitute only
the fixed driver path with sandbox fixtures; real Safari execution is unverified.
The gate covers this exact 2-package inventory, not arbitrary secret detection.
No full landing suite manually run; handoff owns that invocation.
No commits, tags, installs, home writes or LOGBOOK.md edits by this producer.

Checklist interpretation follows current explicit campaign instructions: findings
are recorded in board notes instead of LOGBOOK.md. Item 9 assigns independent
acceptance to the reviewer run; checking the producer obligation does NOT attest
that acceptance has occurred. Independent review and signed scoped PR integration
remain pending and owned by reviewer/orchestrator respectively. Prior attempts
misread this role assignment as requiring acceptance before producer handoff.
