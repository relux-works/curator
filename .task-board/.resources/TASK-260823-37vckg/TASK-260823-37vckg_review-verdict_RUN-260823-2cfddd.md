# Review verdict: accepted

Task: `TASK-260823-37vckg`  
Reviewer run: `RUN-260823-2cfddd`  
Reviewed main commit: `95ca5ae837462463e84d27289c9ed6141f27c43d`  
Replacement PR: <https://github.com/relux-works/curator/pull/24>

## Findings

No blocking or rework findings.

## Scope and architecture

- PR 24 is the current-main extraction of PR 14 commit `d345420109a9d043546d7cdb7b78a13d0bc19137` and retains the same two-file scope and change volume: `.github/ci/root-artifacts.tsv` plus `internal/install/dryrun_conformance_test.go`.
- The resulting test file differs from the old PR 14 result in only two lines: the project and global cases preserve current-main `installPlatform()` calls instead of restoring the stale hard-coded `"unix"` value. This is the correct conflict adaptation.
- The new multi-project binding exercises the project planner once per canonically ordered project with one shared `FetchedRepos` set, matching the production `--all` operation shape. It uses the real trusted-toolchain/protected-cache boundaries and a refusing source-aware builder, so the dry-run assertions are non-vacuous.
- The added effect-surface witness test ensures every published forbidden effect can be observed after a real write, preventing a missing probe from silently passing.

## Validation

- Detached reviewer worktree was created from `origin/main`; both `HEAD` and `origin/main` resolved to merge commit `95ca5ae837462463e84d27289c9ed6141f27c43d`.
- Exact candidate regression executed on that merge commit:

  `CURATOR_CONFORMANCE_ROOT=/Users/iv/Developer/ReluxWorks/curator-spec/conformance/v1 go test ./internal/install -run '^TestAuthoritativeDryRunCasesMutateNothingPersistent$/^compiled-cache-miss-is-read-only$' -count=1 -v`

  Result: PASS; subtest PASS in 7.19s, package PASS in 7.485s.
- `git diff --check 17c9218^ 17c9218`: clean.
- `gofmt -l internal/install/dryrun_conformance_test.go`: no output.
- PR 24 merged at 2026-08-23T11:47:42Z. All executed PR CI checks passed: Test on Ubuntu/macOS/Windows, Race on Ubuntu/macOS, Lint, three Gate self-tests, Interop conformance gate, and Naming gate. The workflow-dispatch-only candidate matrix was skipped by design.
- Producer evidence additionally records an expected-red baseline with `published dry-run scope "multi-project" has no executable binding`, a green extracted named test, a green full candidate gate after materializing the test-tools submodule, and green lint/vet/build.
- The old PR 14 Windows check was not evidence against this patch: its log contains many unrelated platform failures and did not execute the candidate internal/install binding. Current replacement PR CI is fully green.

## Verdict

Accepted. The merged extraction fixes exactly the missing executable binding, fits the existing install test architecture, and satisfies the task acceptance criteria.
