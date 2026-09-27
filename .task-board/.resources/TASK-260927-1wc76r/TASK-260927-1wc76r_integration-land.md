# TASK-260927-1wc76r — integration landing preconditions (bound developer run)

Task: TASK-260927-1wc76r — implement-env-unmanage-restore-backups
Story: STORY-260927-22m88w
Change Request: CR-TASK-260927-1wc76r-1, revision 1 — ACCEPTED (per integration assignment)
Board status (verified 2026-09-27): `integrating`
Branch (verified): `task-board/story/STORY-260927-22m88w`
Worktree: `/Users/administrator/Developer/ReluxWorks/curator/curator/.temp/STORY-260927-22m88w/worktree`

## Instruction precedence applied

- The attached `1wc76r-integrate-land.md` instructs running
  `task-board worktree integrate ... --revision 1` and attaching the log.
- The bound **Integration Assignment** on this run supersedes it: keep the board
  at `integrating`, do NOT execute or detach `worktree checkpoint` / `worktree
  integrate`, do NOT call generic `handoff` or set status; confirm landing
  preconditions, attach fresh task-scoped outcome evidence, and end so the
  runner performs the bound landing synchronously.
- This run followed the Integration Assignment: `worktree integrate` was
  NOT executed. No board status writes were made in this run beyond the
  initial `set_status(..., integrating)` (already `integrating`; no-op).

## Landing preconditions (verified, no repo files changed by this run)

- `task-board q 'get(TASK-260927-1wc76r) { id status }'` → `{"id":"TASK-260927-1wc76r","status":"integrating"}` (exit 0)
- `git branch --show-current` → `task-board/story/STORY-260927-22m88w`
- `git log --oneline -2` → HEAD `0ffe2e1d Record STORY-260924-3gd2d6 board state` (no local
  commits on the Story branch; candidate is uncommitted, as required for handoff snapshot)
- `git status --short` (uncommitted candidate, unchanged by this run):

```text
M .github/ci/conformance-case-counts.tsv
M .github/ci/platform-cases.tsv
M README.md
M cmd/curator/env.go
M cmd/curator/main.go
M docs/cli.md
M internal/envprofile/switch.go
?? cmd/curator/env_unmanage_test.go
?? internal/envprofile/unmanage.go
```

- Forbidden-file check: `git diff --name-only -- CHANGELOG.md LOGBOOK.md` → empty (no output).
  No CHANGELOG/LOGBOOK edit, per AC. Changed set is product + tests + docs + `.github/ci` only.
- No repo file was modified by this integration run (read-only checks + `go build/vet/test`
  caches only). `git diff --stat` matches the candidate above (7 tracked files + 2 untracked).

## Validation evidence (standalone processes, `set -o pipefail`, worktree root)

All commands run directly as standalone processes; exit codes are real.

1. `go build ./...` → exit 0, no output.
2. `go vet ./cmd/curator/... ./internal/envprofile/...` → exit 0, no output.
3. `go test -count=1 -run 'Unmanage' ./cmd/curator/...` → exit 0: `ok github.com/relux-works/curator/cmd/curator 153.687s`
4. `go test -count=1 -run 'Unmanage' ./internal/envprofile/...` → exit 0: `ok github.com/relux-works/curator/internal/envprofile 7.745s`

Not run / bounds (honesty contract):

- Full unscoped `go test ./internal/envprofile/...` was started, then terminated via session
  terminate before completion in favor of the bounded `-run Unmanage` masks above; it is
  NOT claimed as evidence.
- The full landing suite was deliberately NOT run: per `campaign-producer-rules.md` the
  runtime publishes the Change Request and runs the configured landing suite exactly once,
  and the bound landing runs synchronously after this run. No full-suite result is claimed here.
- Cross-platform/hosted lanes (ubuntu/macos/windows, rose-air ARM64) not run from this host;
  reported as unverified, never as passing.
- Review-verdict routing (`accept_cr` / changes-requested) is the reviewer's step, not this run's.

## Producer implementation (as landed in candidate, for landing reference)

- `curator env unmanage --restore-backups` through the CLI production entry
  (`cmd/curator/env.go`, `cmd/curator/main.go`), reusing takeover/backup-record code
  (`internal/envprofile/unmanage.go`, `internal/envprofile/switch.go`).
- Spec: curator-spec v1.0.0-rc.13, environments unmanage + takeover backup records (clauses
  cited in the producer results resource); unreadable backup record stops restore with the
  typed read-failure diagnostic (§8.4.1 discipline, stateread seam) — never treated as absent;
  absent record restores nothing; no partial writes on failure.
- Vectors: backup-record-unreadable-restore-stops, backup-record-absent-restore-nothing, and
  every other unmanage vector driven through the production entry; owned gap rows removed
  (before/after in producer results); mutants (unreadable-as-absent; restore-without-record)
  killed. Details live in the producer's task-scoped results resource.
- CHANGELOG entry text lives in the producer results resource (no CHANGELOG edit in the tree).

## Outcome for the runner

- Landing preconditions hold; candidate tree is as above; narrow gates green with exit 0.
- This run changed no file, made no board status change, published no CR, ran no
  `worktree integrate`, and calls no `handoff` — per the bound assignment, the runner
  performs the landing synchronously from here.
