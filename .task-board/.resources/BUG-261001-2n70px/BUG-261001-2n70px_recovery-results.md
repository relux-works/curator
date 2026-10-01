# BUG-261001-2n70px — install-skips-non-git-project-root

Ready for review. The existing uncommitted implementation was inspected and
validated in this recovery run. No commits, branch changes, or LOGBOOK.md writes.

## Implementation and recovery

Production path: compiled curator main -> cmdInstallMode -> install.Project /
projectAttempt -> gitignore.Ensure / Missing. Exit 128 plus the stable English
repository-discovery absence diagnostic returns ErrNotRepository. The installer
prints a notice and continues materializing declared skill bytes. Exit 1 retains
the Git workspace skip/fix policy; other Git failures refuse with CLI exit 1.
LC_ALL=C stabilizes the diagnostic. The schema-1 dev-substitution gate also
recognizes the non-repository outcome. docs/cli.md documents non-git product
folders and the project-root lock / manager-home source bindings; CHANGELOG.md
contains the Unreleased fix line.

The previous configured handoff gate failed with exit 1. Its Windows go test
passed (exit 0), but the platform-case gate rejected the unregistered skip reason
for the two POSIX Git shim rows (exit 1). The existing recovery edits register
only that exact reason and enumerate the entry-point rows in the platform
ledger. Gate self-tests permit those two skips on Windows while refusing skips
of positive non-git bytes, corrupt Git metadata, or Unix unexpected-exit rows.

## Actual commands in this run

All commands ran directly as standalone processes, without tee or gate pipes.
Redirection retained the command's real exit code. Raw logs are attached under
task-scoped recovery names.

| Command | Actual exit | Evidence |
| --- | --- | --- |
| go test ./cmd/curator -run '^TestInstallGitignoreEntryPoint$' -count=1 -v | 0 | 10/10 compiled CLI rows |
| Targeted regression command below | 0 | All three requested packages; raw recovery-targeted-tests.log |
| Old-skip mutant: go test ./cmd/curator -run '^TestInstallGitignoreEntryPoint$/^non-git$' -count=1 -v | **1** | Expected failure: install returned 0/skipped but SKILL.md bytes were absent; 1/1 positive row killed mutant |
| All-128 mutant: go test ./cmd/curator -run '^TestInstallGitignoreEntryPoint$/^(broken-git\|unexpected-exit-128)$' -count=1 -v | **1** | Expected failure: install returned 0 and published bytes instead of refusing; 2/2 negative rows killed mutant |
| cmp internal/gitignore/gitignore.go .temp/BUG-261001-2n70px-evidence/gitignore.baseline | 0 | Restoration checked after each mutant; restored from saved copy |
| Post-restoration command below | 0 | Fresh three-package validation; raw recovery-post-restore-tests.log |
| golangci-lint run ./cmd/curator/... ./internal/install/... ./internal/gitignore/... | 0 | Zero issues |
| go vet ./cmd/curator ./internal/install ./internal/gitignore | 0 | Scoped vet |
| go build -o .temp/BUG-261001-2n70px-evidence/curator ./cmd/curator | 0 | Production CLI compiled |
| bash .github/ci/ledger-consistency.sh .temp/BUG-261001-2n70px-evidence/recovery-ledger | 0 | 493 ledger rows checked across linux/darwin/windows build inventories |
| bash .github/ci/gate-selftest.sh | 0 | Entire self-test command passed; 4/4 new skip-policy rows passed, including three refusal controls |
| bash .github/ci/no-broad-suppression.sh | 0 | Narrow suppressions only |
| bash -n .github/ci/gate-selftest.sh | 0 | Shell syntax |
| test -z "$(gofmt -l internal/gitignore/gitignore.go internal/gitignore/gitignore_test.go internal/install/install.go internal/install/gitignore_spawn_test.go cmd/curator/install_gitignore_test.go)" | 0 | Formatting |
| git diff --check | 0 | Whitespace |
| git diff --quiet -- LOGBOOK.md | 0 | Explicit task prohibition honored |

Targeted regression command:

```sh
go test ./cmd/curator ./internal/install ./internal/gitignore -run '^(TestInstallGitignoreEntryPoint|TestCLIEndToEndInstallStatusAndTamperCheck|TestInstallFlags.*|TestProjectResolveLocalCreatesLockThroughCLI|TestProjectNonGitWithDevSubstitution|TestGitignoreGateSkips|TestProjectRefusesOnGitignoreSpawnFailure|TestDraftLocalRuntimeMaterializesFromFrozenSnapshot|TestEndToEndInstall|TestMissing.*|TestEnsure.*)$' -count=1 -timeout=4m -v
```

Post-restoration command:

```sh
go test ./cmd/curator ./internal/install ./internal/gitignore -run '^(TestInstallGitignoreEntryPoint|TestProjectNonGitWithDevSubstitution|TestGitignoreGateSkips|TestProjectRefusesOnGitignoreSpawnFailure|TestMissing.*|TestEnsure.*)$' -count=1 -timeout=4m -v
```

Coverage: 10/10 CLI rows (two non-git positives, five Git workspace regression
rows, three broken/unexpected Git refusals); 2/2 schema-1 install rows (declared
and substituted); 4/4 new CI skip-policy self-test rows. Positive rows verify exact SKILL.md and references/info.md
bytes, notice, lock/binding locations, unchanged source content, allowed project
surfaces, and no .git or .gitignore creation. Mutants were rerun here; no prior
passing artifact substitutes for the commands above. Source hashes accompany
this outcome.

Bound: execution was on macOS. Cross-platform inventories and synthetic gate
streams are not Linux/Windows runtime execution. The filesystem-boundary
discovery diagnostic was not reproduced on a real mount boundary. Full repository
tests were not rerun locally; the previously recorded full internal/install
timeout remains a failure, not passing evidence. The configured remote matrix
has not yet produced a green result for the recovery candidate.

The generic logbook checklist is N/A because nongit-brief.md explicitly forbids
LOGBOOK.md changes. Findings are recorded in this outcome and board notes.
