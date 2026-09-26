# TASK-260907-187z6x — integration evidence (CR rev 4, role developer)

Scope: git same-source reinstall drops `--use`/`--takeover` (trunk defect).
Integration run: confirm landing preconditions, attach fresh task-scoped
outcome evidence, end without status change, handoff call, checkpoint, or integrate.

## Preconditions confirmed

- Board status at check: `integrating` (via `task-board q`); no spawn
  directives recorded for RUN-260923-f7e76b.
- Worktree branch `task-board/story/STORY-260906-1a2i5a`; work left
  UNCOMMITTED, no commit past checkpoint:
  - `M CHANGELOG.md`, `M cmd/curator/profile.go`,
    `M internal/envprofile/envprofile.go`,
    `?? cmd/curator/profile_git_reinstall_test.go`
  - `git stash list` empty after the mutant check (fix restored).
- No `set_status`, no `handoff`, no `worktree checkpoint` / `worktree
  integrate` executed in this run (per the integration assignment, which
  supersedes the checkpoint instruction attached to the task body).

## Implementation (unchanged from accepted CR rev 4)

- `internal/envprofile/envprofile.go` (`installLocked`): a same-source
  reinstall of a git root still re-resolves via `updateLocked`, then runs
  the install row's activation — `reinstallActivation(home, name,
  options.Use)`, and `activateReinstall(...)` when it returns true — exactly
  as a path reinstall does. `--takeover` alone still activates nothing.
  `reinstallActivation` doc updated to cover both root kinds. Path-root
  branch (`reinstallPathLocked`) untouched.
- `cmd/curator/profile.go` (`cmdProfileInstall`): a partial switch on a
  same-source reinstall is reported as `updated profile` (lock/source WAS
  updated, only the switch did not complete) instead of `installed profile`.
- `CHANGELOG.md`: Fixed entry for the git-reinstall `--use`/`--takeover`
  handling, with path-root behaviour noted unchanged.
- `cmd/curator/profile_git_reinstall_test.go` (new):
  `TestProfileGitReinstallHonoursUseAndTakeover` — eight `run()`-driven
  rows (four flag combinations × current/not-current git root) through the
  `insteadOf` fixture pattern (fake canonical `https://example.com/groot`
  rewritten to a local repo via `GIT_CONFIG_GLOBAL`; no `file://`
  operand), asserting activation, backup, and the `updated profile`
  operator line.

## Validation run foreground in this session (real exit codes, no pipes)

- `go test ./cmd/curator/ -run TestProfileGitReinstallHonoursUseAndTakeover -count=1 -v`
  → exit 0 — PASS, 8/8 subtests (~55s).
- `go test ./cmd/curator/ -run TestProfileInstallReinstallHonoursUseAndTakeover -count=1 -v`
  → exit 0 — PASS (path-root parity, no path behaviour change, ~40s).
- `go test ./internal/envprofile/ -run 'TestReinstallSameSourceIsAnUpdate|TestInstallUseSwitchesAndAgrees|TestInstallUsePartialLeavesCurrent|TestPathReinstallBareAndNonCurrentMovesPin' -count=1 -v`
  → exit 0 — PASS, 4/4 (~32s).
- `go test ./internal/envprofile/ -count=1` → exit 0 —
  `ok github.com/relux-works/curator/internal/envprofile 448.631s` (full suite green).
- `go vet ./cmd/curator/` → exit 0. `go build ./...` → exit 0.
  `gofmt -l` on the three touched files → clean (no output).
- Narrowing mutant (stash-revert of the three tracked files, new test kept):
  `go test ./cmd/curator/ -run 'TestProfileGitReinstallHonoursUseAndTakeover/not-current/use_with_takeover_takes_over_and_activates' -count=1 -v`
  → exit 1 — FAIL with the trunk-defect signature: exit 0 saying
  `updated profile groot (lock sha256:…)` with no `switched` line. Named test
  kills the mutant; fix restored via `git stash pop`.

## Explicitly not run

- Full `./cmd/curator/` package suite: only the two `-run` subsets above ran
  foreground in this session. Full `./internal/envprofile/` suite: ran green
  (exit 0) as cited above.
