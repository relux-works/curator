# Brief — TASK-260908-1bpra2 b3-mcp-packages-delivery (relux-mcp)

Repository: relux-works/relux-mcp (private, created 2026-09-16, main = bootstrap README + .gitignore). Work in your managed Story worktree of this repository.

Source of truth for what agents-infra installs TODAY: /Users/administrator/Developer/ReluxWorks/relux-agents-infra-main at 459742ea67e3c6b84169520b92d74fe7f73e3002, `.configs/codex-mcp-servers.toml`: `figma` (url https://mcp.figma.com/mcp) and `safari` (absolute command "/Applications/Safari Technology Preview.app/Contents/MacOS/safaridriver" with args ["--mcp"]). `lldb-mcp` no longer exists upstream (README §"There is no shared LLDB definition"), so NO lldb package: document the drop in the repo README.

Deliver, per Decision 0012 D6 and environments.md 1.1 §7.8 (curator-spec main /Users/administrator/Developer/ReluxWorks/curator/curator-spec):
1. `packages/figma/agent-mcp.json` — https url declaration; if Figma needs a bearer token, declare it through `env_names` (never a literal secret).
2. `packages/safari/agent-mcp.json` — a BARE command `safaridriver-mcp` with args `["--mcp"]` (absolute commands are rejected by the protocol), plus `bin/safaridriver-mcp`, a POSIX sh wrapper that execs "/Applications/Safari Technology Preview.app/Contents/MacOS/safaridriver" "$@" and fails with a clear message when the app is absent. Document how the wrapper reaches PATH (machine shim of the machine-current profile; `path_prepend` is reserved in revision 1 — do not invent it).
3. Validate every manifest with the installed manager's parser: use curator's `internal/mcp` (from /Users/administrator/Developer/ReluxWorks/curator/curator) via a throwaway `go run` oracle under `.temp/`, exactly like the B1 producer did for context packages; attach the oracle output. Add `scripts/validate.sh` (JSON validity, required fields, bare-command rule, no absolute paths, no secrets) and a narrowing mutant table (absolute command, literal token, unknown field).
4. Machine allowlist: document (README) the source identities the operator must allow (`git@github.com:relux-works/relux-mcp.git` packages) per §7.8 allowlist-over-source-identities; do not write to any home.
5. Versions `1.0.0`; tags are cut by the orchestrator after review. Record the source pin (agents-infra commit) in the README.
Evidence: `TASK-260908-1bpra2_results.md` (table per package: fields, validation exits, mutants). Hand off with `task-board handoff TASK-260908-1bpra2 --role developer`.
