# BUG-260825-2hrpp5 review verdict

## Verdict

ACCEPTED. The delivered fix resolves exactly the defect this bug names, at the
lifecycle layer, with the strongest possible evidence: native Windows execution
plus a mutant run proving the new assertion catches the old defect. One
out-of-scope pre-existing failure remains on the branch and needs its own task
(see "Remaining red on PR 43" below); it is not caused by, or fixable within,
this bug.

## What was reviewed

Commit `e8a16e2` ("Copy the broker wrapper on Windows instead of hard-linking
it"), head of `task/TASK-260825-1d0eo5-https-credentials-composite` (PR 43),
committed and pushed by the orchestrator after producer run RUN-260825-b68d9e.
Diff scope: `internal/buildrepo/httpsbroker.go` (+6/-2),
`internal/buildrepo/httpsbroker_test.go` (+10/-1). The story worktree's
on-disk copies of both files plus `admission.go` are blob-identical to
`e8a16e2` (verified via `git hash-object`).

## Root cause validation

The bug description guessed "the broker copy the test materializes and runs is
still held". The implementer's diagnosis is sharper and correct: neither
failing test ever executes the wrapper as a child process
(`RunHTTPSCredentialBroker` is called in-process). The wrapper was created by
the `os.Link` fast path in `copyBrokerExecutable`, so on Windows it was another
name for the *running test executable's* locked file identity — Windows opens a
running image without `FILE_SHARE_DELETE`, so unlinking any hard link to it
fails with `Access is denied`. Reproduced natively (see Evidence, mutant run):
the pre-fix code fails with byte-for-byte the same error shape as CI run
32803150486.

This was also a real production defect, not just test flake:
`internal/buildrepo/admission.go:335` materializes the wrapper from
`request.Tool.AskPass` — in production the running curator manager binary — so
fetch-workspace cleanup on Windows would have hit the same locked identity.
The fix lands at the right layer (product code), not in the tests.

## Fix correctness

- `copyBrokerExecutable` skips the hard-link fast path only on Windows
  (`runtime.GOOS != "windows"`); Unix keeps `os.Link`, where unlink is
  indifferent to open handles. The comment states the platform constraint.
- The byte-copy path closes both handles before returning: input via deferred
  `errors.Join(err, input.Close())`, output via explicit `Close` on both the
  success and the copy-error path. Nothing holds the wrapper after
  materialization returns.
- No cleanup-error suppression anywhere in the diff; the tests do the
  opposite — `assertBrokerExecutableReleased` removes the wrapper *before*
  `t.TempDir` cleanup and fails the test if removal fails. The formerly
  deferred cleanup failure is now a direct in-test lifecycle assertion.
- Only production `os.Link` on an executable in the codebase is this fixed
  call site; the SSH-story flow materializes no wrapper (checked
  `internal/gitcred`, admission paths). No sibling instance of the pattern.

## Evidence

Reviewed at `e8a16e2` in a detached scratch worktree
(`.temp/BUG-260825-2hrpp5/review-e8a16e2`), independent of the producer's runs.

| Check | Where | Result |
| --- | --- | --- |
| `go build ./...`, `gofmt -l internal/buildrepo/` | macOS, e8a16e2 | clean |
| Both named tests, `-v -count=1` | macOS, e8a16e2 | PASS |
| Full `internal/buildrepo` package | macOS, e8a16e2 | PASS (8.95s) |
| Both named tests, `-v -count=1` | native windows/amd64, go1.25.5, host DESKTOP-3PBO632 | PASS incl. TempDir cleanup |
| Full `internal/buildrepo` package | native windows/amd64 | PASS (8.03s) |
| MUTANT: pre-fix `httpsbroker.go` (e8a16e2^) + new test file | native windows/amd64 | Both tests FAIL at `httpsbroker_test.go:73`/`:106`: "remove materialized HTTPS broker before TempDir cleanup: ... Access is denied." — the exact original failure shape, now surfaced by the assertion |
| Fixed file restored after mutant | native windows/amd64 | PASS again |
| CI run 32806627556 at `e8a16e2` | Lint, Test+Race ubuntu, Test+Race macOS, all gate self-tests, interop, naming | pass |
| CI run 32806627556, Test (windows-latest), go test stage | `test-evidence-windows-latest` artifact, `test/go-test.json` | `go test overall exit=0`; both named tests **pass** (all 8 subtests pass); **zero** fail-action records in the whole windows stream; no TempDir/unlinkat/"Access is denied" anywhere in the job log |
| CI run 32806627556, Test (windows-latest), job conclusion | job log | red **only** from the platform-case gate (two unregistered skip reasons — see below), same two lines as at the parent commit |

Before/after on the same branch: parent run 32803150486 at `6f1040f` had
`go test overall exit=1` (the TempDir cleanup failures) plus the gate failure;
run 32806627556 at `e8a16e2` has `go test overall exit=0` with only the gate
failure left. The delta is exactly this fix.

The mutant run is the negative evidence: with the product fix reverted, the new
release assertion rejects the broken lifecycle on Windows. The assertion gates;
it does not decorate. On the fixed code the same assertion passes and TempDir
cleanup completes (the native runs exit 0, which includes cleanup).

## AC mapping

- "Both tests pass on windows-latest in CI" — confirmed in CI run 32806627556's
  windows evidence stream (both tests `pass`, all subtests `pass`) and
  independently on native windows/amd64.
- "the fix releases the executable rather than ignoring the cleanup failure" —
  yes: independent byte copy with both handles closed, plus an explicit
  pre-cleanup removal assertion. No error suppression.
- "macOS and Linux stay green" — Test+Race green on both in CI at `e8a16e2`;
  full package re-verified locally on macOS.

## Remaining red on PR 43 (out of scope, needs its own task)

The windows Test job also runs the platform-case gate, which fails for a cause
that predates and is untouched by this fix: two HTTPS-story tests skip on
Windows with reasons not registered in `.github/ci/skip-classes.tsv`:

- `internal/buildrepo :: TestPrivateHTTPSBrokerAuthenticatesRealGitRepository`
  — "the local TLS fixture relies on the platform Git shipped on Unix CI"
- `internal/buildrepo :: TestSelectedHTTPSFetchEnvironmentIsScopedAndOverridesBothAskPassSurfaces`
  — "the fake HTTPS transport wrapper is POSIX-only"

The existing ledger pattern `test transport wrapper is POSIX-only` (line 56)
does not match the new reason strings. This failed identically at the parent
commit (run 32803150486, where this bug's evidence came from) but was never
filed as its own task — BUG-260825-2hrpp5 is the story's only open child.
Fixing it inside this bug would smuggle a platform-coverage policy change
(registering skip classes, or making the fixtures Windows-capable) under a
lifecycle fix; it needs its own bug with its own review. Until it lands, the
windows Test job — and PR 43 — stays red for that separate cause. The
checklist item "pull request 43 turns green" is therefore met only for this
bug's contribution: the TempDir cleanup failures no longer exist in the
windows go-test stream.

## Recommendation to the orchestrator

File a follow-up bug under STORY-260825-32bopo: register the two skip reasons
in `.github/ci/skip-classes.tsv` with an appropriate class and rationale (or
decide to make the fixtures Windows-capable). Acceptance: windows Test job of
PR 43 green end-to-end.
