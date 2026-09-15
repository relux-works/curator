# Review verdict — TASK-260908-1bpra2 CR-TASK-260908-1bpra2-1 revision 1

Verdict: ACCEPTED. Reviewer: independent reviewer run (claude-fable-5-1), shell zsh, `set -o pipefail`.

## Candidate identity
- Base 3125e491275c23c0e8b7f4da5ebfd3fe900220a0, candidate tree f4129d308ffb264fdf98ba7dcb5631199fd8d1b8.
- Every one of the 8 changed paths in the worktree is byte-identical to the candidate tree blob (per-file `git show <tree>:<path> | diff -q`). Modes: bin/safaridriver-mcp 100755, scripts/validate.sh 100755, scripts/validate.py 100644.

## Independent reruns (my own, not producer evidence)
| Check | Command | Exit |
| --- | --- | --- |
| Release gate | `./scripts/validate.sh` | 0 (both manifests valid) |
| Unit tests | `python3 -m unittest discover -s tests -v` | 0, 6/6 ok, 3 narrowing mutants killed (normal=1, narrowed=0) |
| Curator parser oracle | throwaway `go run` under curator/.temp importing `internal/contextpkg.ParseMCP` at curator main f750344 | 0: figma ACCEPT (http, https://mcp.figma.com/mcp, env_names []); safari ACCEPT (stdio, safaridriver-mcp, args [--mcp], env_names []) |
| Oracle mutants | absolute command; `server.token` literal + env_names; top-level unknown field; URL with `?token=` query | 1, all four REJECT with mcp_declaration_invalid |

Oracle scratch removed after the run; curator control root left with no code changes.

## AC / brief conformance
- packages/figma: https URL, no literal secret, `env_names: []` (upstream pin declares no token; README explains how to add one via env_names only).
- packages/safari: bare command `safaridriver-mcp`, args `["--mcp"]`; bin/safaridriver-mcp is POSIX sh, execs the STP safaridriver with "$@", exit 127 with clear stderr when absent (tests drive both branches with sandboxed copies).
- lldb: no package; README documents the upstream drop. Verified: agents-infra main HEAD is 459742ea67e3c6b84169520b92d74fe7f73e3002, `.configs/codex-mcp-servers.toml` contains only figma+safari, README §"There is no shared LLDB definition" present.
- README: source pin recorded; allowlist identity `git@github.com:relux-works/relux-mcp.git` / canonical `github.com/relux-works/relux-mcp` documented as allowlist-over-packages (matches environments.md §7.8 wording); `path_prepend` correctly described as reserved and unused; machine-current shim described as the PATH route. Versions 1.0.0, no tags, no home writes, no commits on the branch.
- scripts/validate.sh: JSON validity (incl. duplicate keys), strict field sets, bare-command rule, no absolute paths, no literal token in args/URL/env_names; tests reach the real shell entry point.

## Bounds (documented, not blocking)
- The release gate is an inventory gate pinned to exactly figma+safari at 1.0.0, not a general protocol parser; README states this. Adding a package requires editing validate.py. Curator's ParseMCP remains the authoritative parser (oracle above).
- ParseMCP structurally accepts arbitrary arg strings; the "literal token in args" rejection is enforced by the release gate only. README/producer notes state this bound.
- Wrapper resolution on a real machine (shim on PATH, STP 247+ present) is not verified here; manager reports `mcp_command_unresolved` if absent, per spec.

## Notes
- Findings kept on the board; LOGBOOK.md untouched per campaign rules.
- No commit_ack supplied; integration is the producer/orchestrator path.
