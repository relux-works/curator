# Review verdict — TASK-260908-3bxxzi, CR-TASK-260908-3bxxzi-4 revision 4

Verdict: **changes requested** (route `to-dev`). Reviewer: Claude (claude-fable-5-1), 2026-09-16, shell zsh, `set -o pipefail`.

## Candidate verified
- Worktree at main `1cb41aa`; `git diff --stat 357b3513…` against the working tree is empty, so the reviewed tree is exactly the candidate tree.
- `scripts/remote-gate.sh` appears in the CR delta only because the CR base is `3cd1304`; it is unchanged relative to main `1cb41aa` (not part of this task's work).

## Independent reruns (all in this session)
| Command | Exit |
|---|---|
| `go build ./...` | 0 |
| `make fmt-check` | 0 |
| `go vet ./...` | 0 |
| `go test ./cmd/curator-run/ ./internal/cli/` | 0 (ok, ok) |
Full `make check` / race not rerun here; that is the remote gate's job.

## What checks out
- README exit-code/diagnostic table matches SPEC §6 code families exactly.
- README "machine wins" for ax.json matches `internal/axconfig/config.go` (machine dir first, operator never inspected when machine file exists).
- `subcommand_provider_untrusted` exists in curator-spec `protocol/environments.md`; install guidance (trusted PATH, not `~/.local/bin`) is correct.
- CHANGELOG 0.1.0 covers PRs 4–15 and the agents-management v0.5.13 dependency; no tag/release created.
- `buildVersion` bump to `0.1.0-dev` is justified: README and CHANGELOG reference it.
- CI: hosted matrix retained, `gate/**` trigger from main retained, `Test (rose-air)` job gated on `vars.ROSE_AIR_RUNNER == 'true'` with setup-go + `make check`; no release/tag jobs.
- `TestRunHelpGolden` drives the production `run` entry point via `runNoResolve` for both `--help` and `-h` and pins stdout byte-for-byte; it fails if help text drifts (a gate that would fail).

## Defect requiring rework
**Help option column misaligned** (`internal/cli/cli.go:45` and `:47`). Relative to main, two previously aligned lines were changed:
- `--ax-profile <standard|yolo>` now has 7 spaces before its description (was 6), pushing it one column right of every other option.
- `--version` now has 32 spaces (was 33), pulling it one column left.
All other option rows keep the description at column 37. This is a visible regression in the headline `--help` deliverable, and it is now enshrined in `testdata/help.golden` and all ten `forbidden-*.golden` fixtures (usage is printed on every usage error). DoD item 2 ("`--help` output complete and consistent") is not met.

Fix: restore the original spacing on those two lines (`--ax-profile <standard|yolo>      ax execution …`, `--version                         print the …`), regenerate `help.golden` and the ten `forbidden-*.golden` files, rerun `go test ./cmd/curator-run/ ./internal/cli/` and `make fmt-check`, and hand off again. No other content changes are needed.
