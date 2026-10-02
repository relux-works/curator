# BUG-261002-ot3ea1 — global-upgrade-sweeps-in-use-builds

Developer handoff: ready for review. Repository changes remain uncommitted in the assigned Story worktree.

## Implementation

The incident's sweep is `internal/buildcache.Store.Sweep`. Before this change it retained marker/journal references and publications within the default 24-hour grace period; it had no daemon-specific process check. The new process snapshot protects every executable path beneath a build directory, independently of command name, publication age, or daemon registration. This covers both protected namespaces: `go-v1` and `go-v1-receipt-3`, including retired-tree cleanup. Existing references, grace, protected-boundary validation, and handle-relative deletion remain in place.

The production global path is `cmdGlobal(upgrade)` → `runGlobalInstallMode` → `install.Global` → `runCommit` → `collectAfterCommit` → `scopes.Collect` → `Store.Sweep`. Standalone `run(gc)` reaches the same collector via `collectUnderLock`.

Process enumeration uses macOS `kern.proc.all` and the saved executable path from `kern.procargs2`, Linux `/proc/<pid>/exe`, and Windows Toolhelp process enumeration plus `QueryFullProcessImageName`. Kernel processes and proven zombies/exits are excluded; unreadable or malformed observations retain all protected builds and produce a warning. Canonical paths handle symlink aliases; Windows comparisons ignore casing; separator boundaries prevent sibling-prefix matches. `Store.ExecutablePaths` provides the injectable process-table dependency; production defaults to native enumeration.

Chosen design: native liveness retention, with the existing publication grace unchanged. No additional previous-build-per-command grace was added: age does not bound runner or daemon lifetimes.

The macOS parser reads saved exec_path rather than argv[0], following [Apple's process argument layout](https://github.com/apple-oss-distributions/adv_cmds/blob/main/ps/print.c). Windows uses the image query's limited-information access described in [Microsoft's process access documentation](https://learn.microsoft.com/en-us/windows/win32/procthread/process-security-and-access-rights).

One CHANGELOG entry was added. No LOGBOOK was written, as explicitly required by sweep-brief.md. Findings are recorded here and in board notes.

## Executed verification

Every command was run by this developer. No previously attached evidence was accepted. Go/build/lint processes ran directly, without tee or a pipeline, under the directory mutex at `~/.mini-build-lock`; the wrapper returned the command's real exit status. The host syspolicy service was checked at checkpoints and was running each time observed (successive crash count moved from 354 to 355).

| Command | Real exit code | Evidence / scope |
| --- | --- | --- |
| `go test ./internal/buildcache ./internal/scopes -count=1 -v -timeout=3m` | 0 | packages.log; entire two package suites, after mutant restoration |
| `go test ./cmd/curator -run '^TestGC' -count=1 -v -timeout=4m` | 0 | cli-gc.log; 6/6 GC entry-point tests, after restoration |
| `golangci-lint run ./internal/buildcache/... ./internal/scopes/... ./cmd/curator/... --timeout=4m` | 0 | lint.log; 0 issues |
| `GOOS=windows go vet ./internal/buildcache ./internal/scopes ./cmd/curator` | 0 | windows-vet.log; cross-platform source/test compilation and vet |
| `GOOS=linux go vet ./internal/buildcache ./internal/scopes ./cmd/curator` | 0 | linux-vet.log; cross-platform source/test compilation and vet |
| `go build -o .temp/BUG-261002-ot3ea1/curator ./cmd/curator` | 0 | build.log; native CLI build |
| `git diff --check` | 0 | diff-check.log |
| `cmp internal/buildcache/collect.go .temp/BUG-261002-ot3ea1/collect.go.original` | 0 | exact source restoration verified |

The liveness matrix passed 14/14 rows: in-use, another file beneath the build, unused, sibling prefix, enumeration error, active daemon, and unused daemon, in each of the two namespaces. The scope collector passed 2/2 new rows (in-use with an unused neighbor removed; enumeration failure retaining both with a warning). Additional tests cover partial/malformed tables, executable symlink aliases, saved macOS executable paths, and a forged argv[0].

The real CLI test launches the installed cached binary, waits for its readiness signal, removes its install marker, ages its cache entry by 30 days, and invokes `run(gc)`. Native enumeration encountered an unreadable macOS process (input/output error on the recorded rerun), so this entry-point test proves fail-safe retention, not successful full native enumeration. The deterministic collector test proves ordinary path-based retention and unused-build removal through `scopes.Collect` and the real Store.

Bounds: the retained-path decision uses one process snapshot per sweep; process starts after that snapshot are outside that observation. External repository artifact GC (`buildrepo.Collect`) was not changed. Full native enumeration on this host is unknown; an inaccessible process conservatively prevents protected-cache cleanup. Native Linux and Windows runtime tests were not executed because this run has only a macOS host; their sources/tests were cross-vetted. The CLI suite was deliberately restricted to GC tests, not the entire command package. Five conformance tests skipped because CURATOR_CONFORMANCE_ROOT was unset; the Windows case-comparison test skipped on macOS. No full conformance qualification is claimed.

## Mutants

Both mutants were run as standalone `go test ./internal/buildcache -run '^TestSweepLiveProcessBuilds$' -count=1 -v -timeout=2m` processes. Source was copied aside before mutation, restored from that copy, and byte-compared before green validation.

| Mutant | Real exit code | Observed expected failure |
| --- | --- | --- |
| Remove `buildInUse(...)` from the production candidate retention condition | 1 | 6/6 in-use, other-file and daemon rows failed across both namespaces; mutant-delete.log |
| Narrow the production retention condition to only the `tb-sessiond` executable | 1 | 4/4 runner and other-file rows failed; both daemon rows still passed; mutant-daemon-only.log |

These are failing gates with expected-red rationales, not passing test suites. Deletion proves the check is effective; daemon-only narrowing proves it covers ordinary runners and any executable file beneath a build.

## Earlier diagnostic runs

- Initial `go test ./internal/buildcache -count=1 -timeout=3m`: exit 0.
- Initial CLI single-test run (`^TestGCRetainsLiveProcessBuild$`, count=1, verbose, timeout=3m): exit 1 because the test context canceled the child before cleanup. Fixed by explicit stdin-close/wait cleanup.
- Initial `go test ./internal/scopes -count=1 -timeout=3m`: exit 1 because native enumeration uncertainty correctly retained the orphan the pre-existing test expected to sweep. The fixture now supplies an explicit empty process table through the real cache dependency.
- First combined package retry (`./internal/buildcache ./internal/scopes`, count=1, timeout=3m): exit 1 due to two undefined store variables in the conformance fixture refactor. Corrected; later package runs exit 0.
- Earlier GC-prefix run: exit 0. Earlier combined package run without verbose output: exit 0.
- The first deletion-mutant command used the mistyped singular mask `^TestSweepLiveProcessBuild$`; exit 0 with zero tests. This result was rejected as evidence. The corrected plural mask executed all rows and exited 1 as required.

The validation archive contains the recorded green-suite logs, expected-red mutant logs, build/lint/vet logs, build-lock wrapper, saved original sweep source, and source-sha256.json for all changed/new repository files. Empty build/vet logs correspond to the explicitly recorded exit codes above.
