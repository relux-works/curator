# BUG-260920-3vfwch brief (orchestrator, binding)

Story STORY-260915-3w11un (workstation/checkout readiness; leaves land through separate
PRs — publish a Change Request as usual). Hosted `macos-latest` gate flake:
`fork/exec /opt/homebrew/bin/git: permission denied`.

## Evidence already gathered (verify, do not redo from scratch)

- Run 35510984798 (head 5f99188e), job `Test (macos-latest)` 106078737528, artifact
  `test-evidence-macos-latest` (id 10605552532): `internal/install ::
  TestGlobalDryRunPlansBuildsWithoutSessionOrPersistentState` failed at
  2026-09-20T12:48:35Z, 0.04 s after start, in `stage_test.go:723` — the fixture helper
  (`e.buildSkill` → `git [add .]`), i.e. a test-helper spawn, not product code. The
  package started at 12:40:03Z; at the failure ~100 `t.Parallel` tests of
  `internal/install` were in flight (heavy spawn pressure). Earlier occurrences: runs
  35340496757 (`TestDryRunEffectBindingsSeeWhatARealOperationWrites`) and 35481906193
  (`TestEndToEndInstall`) — confirm from their artifacts that both are the same shape
  (helper git spawn in `internal/install`, same error text) and record the timing/
  concurrency for each.
- Go reports every child-startup failure (chdir into `cmd.Dir`, dup, exec) as
  `fork/exec <path>: <errno>`; do not assume the git binary's mode is wrong. Candidate
  causes to discriminate with evidence: (a) transient `EACCES` from `posix_spawn` under
  load on macOS (search the Go issue tracker: darwin `os/exec` "permission denied"
  flake); (b) `chdir(cmd.Dir)` refused because a concurrent test or product routine
  changed permissions on a shared ancestor (TMPDIR root) — check every `Chmod`/
  `chflags`/protection helper reachable from `internal/install` tests for writes
  outside the test's own `t.TempDir()`; (c) the Homebrew git link/target really being
  non-executable at that instant (image-level).

## Ruling

R1 Product code (`cmd/curator`, `internal/**` non-test) is out of scope: no change to
how curator resolves or spawns git.
R2 Deliver two things regardless of which cause you prove: (1) a diagnostic captured
at the moment of the failure — in the shared test git helper(s) of `internal/install`
(and the same helper shape elsewhere if identical), on a spawn error wrap the error with
`command -v`/`ls -lL`/`stat`-equivalent facts gathered in-process (`os.Stat` of the
resolved binary and of `cmd.Dir`, `os.Getwd`, `syscall.Getrlimit(RLIMIT_NPROC/NOFILE)`)
so the next occurrence proves the cause from the go-test stream; (2) a bounded
mitigation that keeps test semantics intact: one retry of the fixture git spawn only
when the error is `EACCES`/`EAGAIN` at spawn (not a non-zero exit, not any git
diagnostic), after a short backoff, with the retry recorded in the test log. If you
prove (b), fix the offending test/helper instead (that is the real bug) and keep the
diagnostic. If you prove (c), add the gate-level diagnostic step in
`.github/ci/test-gate.sh` (record `ls -lL "$(command -v git)"`, `git --version`, `stat`
on failure) and bind the hosted darwin lane's tests to a stable git through the
existing lane-local prefix pattern in `ci.yml` (see the pnpm/rustup prefix comments),
never `/usr/bin/git` silently (document the git version consequence).
R3 Evidence: a unit row that drives the retry path with an injected spawn `EACCES`
(fake command runner or a non-executable fixture binary on POSIX; declared Windows skip
only if a sibling row already skips), a row proving a non-zero git exit is NOT retried,
mutants (remove retry → row fails; retry on any error → negative row fails). Windows
and Linux lanes unchanged in behaviour.
R4 results.md: cause classification with the evidence per run, what was proven vs
bounded, the 10-consecutive-gate criterion stated as an observation window that the
orchestrator tracks after landing (you cannot wait for it), CHANGELOG not required for
test-only changes (say so). Publish the Change Request only when the configured gate is
green; on a recurrence of this very flake during your gate, the diagnostic must show in
the go-test stream — attach it.
