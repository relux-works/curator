# BUG-261001-2n70px — install-skips-non-git-project-root

Ready for review. Changes are uncommitted in the assigned Story worktree.

## Behavior and production path

The compiled CLI drives main -> cli.install (cmd/curator/main.go) -> install.Project/projectAttempt -> gitignore.Ensure/Missing. Only Git exit 128 with its English repository-discovery absence diagnostic yields ErrNotRepository. Installation prints a clear notice and continues without initializing Git or writing .gitignore. LC_ALL=C makes classification independent of the operator's locale. Missing ignore entries (exit 1) retain the existing skip/fix policy; all other Git errors fail closed and the CLI returns 1. The additional schema-1 dev-substitution hygiene gate uses the same distinction.

docs/cli.md documents non-git product folders, the project-root Skillfile.lock.json, and machine-local source-bindings under the manager home. CHANGELOG.md has one Unreleased fix line.

## Evidence bound

Executed locally on darwin/amd64. Compiled CLI matrix coverage: **10/10 intended rows**, each with a fresh temporary HOME and manager configuration. Fixtures use real project add, project resolve, and install processes and a nested source repository. Asserted context bytes are SKILL.md and references/info.md; the package descriptor is not a project context output. Assertions also cover lock/binding locations, unchanged source content, allowed project surfaces, and no .git/.gitignore creation for non-git installs.

| Row | Expected curator install exit | Result |
| --- | --- | --- |
| Non-git product root | 0, installed bytes + notice | pass |
| Non-git product root --fix-gitignore | 0, installed bytes + notice, no ignore file | pass |
| Project inside parent work tree, ignored | 0, installed bytes, no non-git notice | pass |
| Git root, already ignored | 0, installed bytes, existing ignore file unchanged | pass |
| Git root, missing ignore entries | 0, skipped, no installed bytes (existing behavior) | pass |
| Git root, missing ignore entries --fix-gitignore | 0, repaired and installed | pass |
| Git root, negated ignore rule --fix-gitignore | 0, skipped, no installed bytes, ignore unchanged | pass |
| Corrupt .git pointing at missing gitdir | 1, refused, no installed bytes or ignore write | pass |
| Unexpected Git exit 2 carrying absence diagnostic | 1, refused, no installed bytes or ignore write | pass |
| Unexpected Git exit 128 carrying unrelated diagnostic | 1, refused, no installed bytes or ignore write | pass |

Additional internal/install positive scope: **2/2 schema-1 rows**, declared skill and local dev substitution, both install bytes and do not create Git metadata or ignore files. Existing spawn-error refusal, ignore skip/fix, ordinary install and frozen runtime materialization regressions passed.

Bound: this is targeted verification, not a claim that all repository tests passed. Linux and Windows were not run. The two shell-shim entry-point rows skip on Windows; the corrupt .git refusal row is portable. The filesystem-boundary discovery diagnostic is recognized in code but was not reproduced at a real mount boundary in this run. No earlier attached evidence was accepted as a substitute for running these commands.

## Commands and actual exits

All commands ran directly; log redirection did not change the command's exit status. No tee or gate pipelines were used.

| Command | Actual exit | Evidence / meaning |
| --- | --- | --- |
| go test ./internal/gitignore -count=1 | 0 | Full gitignore package before mutants |
| go test ./cmd/curator -run '^TestInstallGitignoreEntryPoint$' -count=1 -v | 1, 1, then 0 | First two development runs exposed fixture mistakes: unsupported --config flag, then an assertion for a package descriptor excluded from projected context. Corrected to CURATOR_CONFIG and actual context bytes. Third run passed 10/10 rows; entrypoint log attached. |
| go test ./cmd/curator -run '^(TestCLIEndToEndInstallStatusAndTamperCheck\|TestInstallFlags.*\|TestProjectResolveLocalCreatesLockThroughCLI)$' -count=1 -timeout=4m | 0 | Existing CLI regressions |
| go test ./internal/install -count=1 -timeout=8m | **1** | Failed at eight-minute timeout (480.567s); active test was TestDraftLocalRuntimeMaterializesFromFrozenSnapshot, running for five seconds at timeout. Timeout log attached. This broad suite is not passing evidence. |
| go test ./internal/install -run '^(TestProjectNonGitWithDevSubstitution\|TestGitignoreGateSkips\|TestProjectRefusesOnGitignoreSpawnFailure\|TestDraftLocalRuntimeMaterializesFromFrozenSnapshot\|TestEndToEndInstall)$' -count=1 -timeout=4m -v | 0 | Bounded relevant subset; includes timeout-active test, which passed on this isolated rerun |
| Old-skip mutant: go test ./cmd/curator -run '^TestInstallGitignoreEntryPoint$/^non-git$' -count=1 -v | **1** | Expected failure: curator install returned 0/skipped, and required SKILL.md bytes were absent. 1/1 positive row killed the mutant. |
| All-128 mutant: go test ./cmd/curator -run '^TestInstallGitignoreEntryPoint$/^(broken-git\|unexpected-exit-128)$' -count=1 -v | **1** | Expected failure: both broken-Git rows returned 0 and installed instead of refusing. 2/2 refusal rows killed the mutant. |
| cmp internal/gitignore/gitignore.go /tmp/BUG-261001-2n70px_gitignore.go.baseline | 0 | Both mutations restored from a saved copy, not git checkout |
| Post-restore targeted go test across cmd/curator, internal/install, internal/gitignore (exact command below) | 0 | Fresh validation.log attached |
| golangci-lint run ./cmd/curator/... ./internal/install/... ./internal/gitignore/... | 0 | Three runs, including after restoration; 0 issues |
| go vet ./cmd/curator ./internal/install ./internal/gitignore | 0 | Before mutants and after restoration |
| go build -o /tmp/BUG-261001-2n70px_curator ./cmd/curator | 0 | Before mutants and after restoration |
| git diff --check | 0 | Whitespace validation |

Post-restore command:

```sh
go test ./cmd/curator ./internal/install ./internal/gitignore -run '^(TestInstallGitignoreEntryPoint|TestCLIEndToEndInstallStatusAndTamperCheck|TestInstallFlags.*|TestProjectResolveLocalCreatesLockThroughCLI|TestProjectNonGitWithDevSubstitution|TestGitignoreGateSkips|TestProjectRefusesOnGitignoreSpawnFailure|TestDraftLocalRuntimeMaterializesFromFrozenSnapshot|TestEndToEndInstall|TestMissing.*|TestEnsure.*)$' -count=1 -timeout=4m -v
```

## Findings and task-specific instruction

The old implementation conflated all Git exit errors with missing ignore entries. Corrupt Git metadata can report “not a git repository: <gitdir>”; that is distinct from repository-discovery absence and must refuse. Both status and diagnostic are necessary: exit 2 carrying the absence text still refuses.

LOGBOOK.md was not touched, as required by the sole current nongit-brief.md instruction. Findings are recorded in this task outcome and board notes. The generic logbook checklist item is not applicable under that explicit task instruction; it does not attest a logbook write.

Remaining validation limitation: the full internal/install suite did not reach a green result within the bounded run. Its actual failure is attached; relevant scoped checks and the test active at timeout were rerun green. No full cmd/curator or repository-wide suite was run.

## Revision 3 (republish)

Trunk moved to f0119a8b while rev2 was green; the orchestrator converged the Story worktree and carried this delta over without conflict. Rev3 republishes that delta unchanged in behavior.

Converge check: `git diff CHANGELOG.md` shows only this task's fix line added under `### Fixed`; trunk's 2772iz askpass line (line 7) is intact; `grep` for conflict markers across CHANGELOG.md, docs/cli.md, internal/gitignore/gitignore.go, internal/install/install.go found none. Diff stat: 9 modified files plus the new `cmd/curator/install_gitignore_test.go` (untracked, part of the delta).

Re-ran on darwin with real exit codes (each command direct, no tee/pipes through the gate):

| Command | Actual exit | Evidence / meaning |
| --- | --- | --- |
| go test ./cmd/curator -run InstallGitignore -count=1 | 0 | ok in 101.291s; entry-point rows green |
| go test ./internal/gitignore -count=1 -v | 0 | 11/11 --- PASS, ok in 1.768s |
| go test ./internal/install -count=1 (single full-suite attempt) | **1** | panic: test timed out after 10m0s; the in-flight test (TestDraftLocalEnforcedCommandRefused, 43s at dump) passes alone in 52s (exit 0). The suite is too slow for one bounded shell call, not hung on this delta — a prior run's attached install-suite-timeout.log shows the same property. Not passing evidence; see chunked rerun below. |
| go test ./internal/install -run 'Test[A-C]' -count=1 -timeout 540s | 0 | ok in 74.168s |
| go test ./internal/install -run 'TestD(e\|i\|o)\|TestDry' -count=1 -timeout 540s | 0 | ok in 20.415s |
| go test ./internal/install -run 'TestDraft[A-G]' -count=1 -timeout 540s | 0 | ok in 439.170s |
| go test ./internal/install -run 'TestDraft[H-Z]' -count=1 -timeout 540s | 0 | ok in 144.975s |
| go test ./internal/install -run 'Test[E-P]' -count=1 -timeout 540s | 0 | ok in 490.956s |
| go test ./internal/install -run 'Test[R-Z]' -count=1 -timeout 540s | 0 | ok in 406.856s |
| bash .github/ci/gate-selftest.sh | 0 | 301 passed, 0 failed |
| golangci-lint run ./internal/gitignore/... ./internal/install/... ./cmd/curator/... | 0 | 0 issues |
| go vet ./internal/gitignore ./internal/install ./cmd/curator | 0 | clean |

Chunk coverage: the six `-run` masks partition all 337 tests listed by `go test -list` (verified zero unmatched names); every chunk log was swept for `--- FAIL`/`FAIL` with no hits. So rev3 upgrades the rev2 limitation: the full internal/install suite is now green, verified in bounded chunks. Mutant, docs, and CHANGELOG evidence from rev2 stands unchanged and was not re-derived; no code, test, doc, or ledger file was modified in rev3.
