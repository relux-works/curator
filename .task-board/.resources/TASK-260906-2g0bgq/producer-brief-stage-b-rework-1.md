# Producer brief: stage (b) — rework 1 (F1–F6 plus the recorded minors)

Worktree `/Users/iv/Developer/ReluxWorks/.worktrees/curator-stage-b`, branch
`feat/agent-environments-stage-b`, head `73174cc3`. Findings:
`TASK-260906-2g0bgq_review-findings-stage-b-1.md` — three blocking, three major, four minors, each
with the reviewer's driving command. New signed commits on top of `73174cc3`, no rewrite. Note: the
stage (a) work has landed on curator main as `981b1eeb`; `git fetch origin && git rebase -S origin/main`
first and prove identity with `git range-diff`.

## Author decisions
- **F1 (blocking)** — gate the `shared` refusal on `atOrAbovePinned`: below the pinned release fall
  through to the `shared` default and keep only the `isolated` refusal. As written, `claude_code` on a
  macOS below the pin refuses every isolation value, so no managed home can be provisioned at all. Add
  both missing matrix rows to `TestIsolationMatrix` and a narrowing mutant (refuse `shared` on darwin
  regardless of the pin) that the new rows kill.
- **F2 (blocking)** — byte drift of a linked manager-authored surface must not be invisible: verify the
  recorded hash of the link target for every surface whose target is a rendered document (everything
  the plan publishes through `p.docs`), keeping the link-identity fast path only for targets under
  `contextstore.Root` (skills trees and referenced module files, which are immutable store entries).
  Add the write-through-link case to `TestResolveDriftRepair` for every adapter and every managed
  surface, plus a narrowing mutant that keeps the link check and drops only the byte check.
- **F3 (blocking)** — the lint gate is red on the head under review while the drafting report attests
  it green. Fix the finding, re-run `golangci-lint run ./...`, and correct the report line with the
  re-run's own output. Treat this as the rule for every gate line in every future report: paste what
  the command printed, never what it should print.
- **F4 (major)** — add a `CheckBoundary` case asserting that a value below the environments root's
  *parent* is refused, and re-run the root-parent mutant to see it fail.
- **F5 (major)** — assert `os.Lstat(<home>/CLAUDE.md).Mode()&os.ModeSymlink == 0` and the recorded copy
  reason under **both** forms, so the §8.1 always-copied rule is tested under `referenced` too.
- **F6 (major)** — the referenced-form staleness check tests the project entry's presence, not the
  approval key: check the key itself. Extend `claudeProjects` to report the flag and add the negative
  case to `TestResolveClaudeProjectEntry`.
- **Minors (F7–F10)** — apply all four: the §7.6 second secondary target missing from the registry;
  an unreadable native `config.toml` treated as the `file` credential store (absence vs unreadable —
  make it fail, not default); the three missing §12 `env status` rows; and `--format env|shell` emitting
  variables the fragment's `env` object does not carry. If any one of them is genuinely a spec question
  rather than a defect, say so with the sentence you read and leave it — but the default is to fix.

## Gates and delivery
`go build ./...`, `go vet ./...`, `gofmt -l`, `golangci-lint run ./...` (must be 0 issues — F3),
`go test -count=1 -race` on the touched packages, the vector families with
`CURATOR_CONFORMANCE_ROOT=/Users/iv/Developer/ReluxWorks/curator-spec/conformance/v1`,
`bash .github/ci/gate-selftest.sh`, the platform-case gate for the three GOOS values, and
`go test -count=1 -timeout 30m ./cmd/curator`. Check every new fixture for the Windows class stage (a)
shipped: never interpolate a native path into a git config value; use one shared helper and the
`file:///C:/...` form. Every fix carries a **narrowing** mutant that kills a named test. Signed commits;
do not push, tag, or open a PR. Attach `TASK-260906-2g0bgq_rework-report-1.md`;
`task-board handoff TASK-260906-2g0bgq --role developer`. Never write LOGBOOK.md or anything into the
control root.
