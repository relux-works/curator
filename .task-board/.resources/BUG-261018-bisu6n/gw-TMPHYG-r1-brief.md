# BUG-261018-bisu6n: the pnpm conformance test leaves a read-only tmp tree that breaks the self-hosted Linux runner (developer, P0 for the wave)

## Problem

Every curator CI job that lands on the self-hosted runner `relux-rise1-curator-4` fails. The repository steps pass; the failure is in "Complete runner": the host's job-completed hook removes the runner tmp and stops on `rm: Permission denied` for a stale tree left by an earlier job:
`<runner tmp>/curator-4/TestRealPinnedPNPMPrivateStoreAndOfflineMaterialization<N>/010/store/v10/index/..` and `.../store/v10/file+..+capture+tarballs+000N.tgz/tarball-integrity`.
pnpm writes its content-addressed store read-only. `TestRealPinnedPNPMPrivateStoreAndOfflineMaterialization` (`internal/pnpmsource/conformance_test.go`, around line 1127) creates that store outside `t.TempDir()` cleanup's reach, or in a form `os.RemoveAll` cannot remove, and never restores write permission. The same jobs pass on curator-2 and curator-3. All rc.7 producers are currently parked on this failure.

## Required

1. **Test fix.** The test (and every helper it uses) restores owner write permission on everything it created, recursively, before removal. Register the restore with `t.Cleanup` **before** the tree is created so failure and early-exit paths also restore. The tree must be fully removable when the test ends, pass or fail.
2. **Sweep.** Find every other test or helper in the repository that creates read-only trees (pnpm or npm stores, Go module caches via `GOMODCACHE`/`GOFLAGS=-modcacherw` absence, `chmod 0o444/0o555`, tar extraction modes, snapshot fixtures). Fix those that share the failure class; list the rest with file and line in the gaps file.
3. **CI self-heal.** In `.github/workflows/ci.yml`, every job that can run on the self-hosted Linux runners ends with a final step that runs with `if: always()` and only when `runner.environment == 'self-hosted' && runner.os == 'Linux'`. It makes the job's own temporary root writable for its owner (`chmod -R u+w`) so that the host cleanup hook can remove it, including trees left by earlier jobs. Determine from the runner setup doc and the job environment which variable names that root (the stale path above is the runner's tmp root, for example `TMPDIR`); refuse to act (log and exit 0) if the variable is unset, empty, `/`, or not a directory. The step only changes permissions: it never deletes, moves, or touches anything outside that root, and never fails the job. Keep it as one shared shell script under `scripts/` called from each job rather than copied inline, if the workflow layout allows.
4. **Tests** (Go, one-line claim comment per test):
   - Regression witness: after the pnpm conformance test body runs (pass and forced failure), no directory under its tree lacks owner write permission and `os.RemoveAll` succeeds.
   - Neighbouring variants from the sweep.
   - The self-heal script: on a fixture tree with read-only directories and files it makes them removable; it refuses unset, empty, `/` and non-directory roots without changing anything; it never deletes. If the repository has workflow structure tests, extend them so a Linux self-hosted job without the final step fails.
   - Plant control: removing the `t.Cleanup` restore makes the regression witness fail for that reason.
5. **Validation.** Full local suite and the full hosted gate on the exact signed candidate. If a job of the hosted gate lands on curator-4, its own final step must heal the stale tree and the job must pass; record which runners the jobs ran on. A failure only in "Complete runner" after the step ran is evidence the heal does not work: report it, never a pass.

## Scope

- In scope: `internal/pnpmsource` tests and test helpers, other test helpers from the sweep, `.github/workflows/ci.yml`, one script under `scripts/`, `docs/self-hosted-runner-setup.md` (a short note on the self-heal step).
- Not in scope: product code behaviour, the host cleanup hook (`/home/gha/bin/runner-job-cleanup.sh` belongs to the runner owner; do not try to reach the host), runner labels or repository variables.
- rc.7 wave rule: do not edit `CHANGELOG.md` or the root `README.md`. Put the changelog entry in the gaps file under "Changelog entry".

## Hand-off

1. Attach `BUG-261018-bisu6n_gaps.md`: each required item with file and line, the tests that prove it, the sweep list, the runners the gate jobs ran on, the changelog entry.
2. Run `task-board handoff BUG-261018-bisu6n --role developer`.
3. END YOUR TURN.
