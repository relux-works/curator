# BUG-261018-dbeghc: Refresh onto 402fdbc9

Ready for review after freshness repair; acceptance and integration remain with the reviewer/orchestrator.

README inspection retained trunk's CLIHELP and SNAPPUB text and the existing managed-home mode note, each once in a sensible order. The README contains trunk's exact bytes plus the existing mode paragraph. No repository edits were needed in this round. All five other implementation/doc/test files match revision 1 byte-for-byte. No reviewer tests were removed or renamed. Reviewer launch previously refused `revision_base_superseded`; no code-review rejection or `repeat-of:` verdict exists.

`task-board worktree refresh-candidate BUG-261018-dbeghc` exited 0, advancing the checkpoint from `402fdbc9b77803cb2d6b11f4f1cfd83961e6897b` to `db65e39555b8dbac7e75f3b2d337daa33831f297`. That advance changes only CHANGELOG.md. Detached candidate `04f9d8d2f06d6ca2df2e0e1b06f78078445fca35` is signed (`G`), tree `8c29fe0bc3ad426e196062a7efa7c9d6a601658b`. The Story branch and real index remain at their recorded checkpoint; the six task files remain uncommitted for the handoff snapshot.

## Production coverage and bounds

**4 of 5 AC rows driven through production; AC5 is external validation below.** The current refresh brief has no surface table, a brief gap rather than catalog-wide coverage. The unchanged implementation coverage is:

| AC | Named tests | Production call site |
| --- | --- | --- |
| 1 | `TestManagedSeedPreservesNativePermissions/{0600,0400,0640,0644}` | `Resolve` -> `gatherSeeds` / `applyPlan` |
| 2 | `TestManagedHomeRepairTightensDirectories` (five adapters; new directories, stale/read-only/dry-run refusal, repair retaining bytes) | `Resolve` -> `verifyHomeObserved` / `managedDirectory` |
| 3 | `TestManagedAuthoredFilesArePrivate`, `TestManagedFallbackPermissions` | `Resolve` -> `claudeSeed` / `applyPlan`; `copyLinkFallback` -> `copyTree` |
| 4 | `TestManagedSeedRefusalsPreserved/{symlink,unreadable,fifo}` and six narrowing controls | `Resolve` -> `gatherSeeds`, seed writer and directory verifier |
| 5 | Full local inventory, full hosted workflow | Validation entry points, not product behavior |

`TestEnvResolveCredentialRecordProvisionAndRepair` additionally exercises actual `env resolve ... --repair` through `cli.cmdEnvResolve` -> `envprofile.Resolve`; mode assertions are supplied by the direct Resolve tests.

Out-of-contract rows remain those justified in the original brief: spec changes excluded by Scope; existing tool-owned seeds do not refresh on repair (section 7.4); native passthrough/link targets and takeover backups retain their ownership contracts. Unix tests establish POSIX modes on Linux and macOS; the Windows job retains existing DACL behavior and makes no POSIX-bit equivalence claim. Deterministic tests do not establish protection against concurrent source replacement/chmod. Rework delta is zero; existing README/docs changes outside internal/envprofile are the expressly required docs note.

## Narrowing controls

Linux and macOS hosted streams, plus the refreshed local suite, exercise all six controls. Each wrapper requires the child command's real exit 1, the named failing behavioral test and the intended permission reason. Compile/setup failures do not qualify. No survivors.

| Mutant | What it narrows the gate to | Named failing test | Verdict / bound |
| --- | --- | --- | --- |
| private-seed-0644 | Only source 0600 widens to 0644 | `TestManagedSeedPreservesNativePermissions/0600` | Killed: 0644, expected 0600 |
| claude-trust-0644 | Only Claude trust writer widens to 0644 | `TestManagedAuthoredFilesArePrivate` | Killed: .claude.json 0644, expected 0600 |
| single-fallback-0644 | Only single-file fallback widens to 0644 | `TestManagedFallbackPermissions/tree-false` | Killed: copy 0644, expected 0600 |
| executable-tree-fallback-0755 | Only a 0755 source bypasses the owner mask | `TestManagedFallbackPermissions/tree-true` | Killed: 0755, expected 0700 |
| codex-directory-drift-admitted | Only Codex directory drift is admitted | `TestManagedHomeRepairTightensDirectories/codex_cli` | Killed: wide home admitted |
| codex-directory-not-tightened | Only Codex protection is skipped | `TestManagedHomeRepairTightensDirectories/codex_cli` | Killed: repair reports permission drift |

The tree mutant preserves the mask token in its other branch. Behavioral tests, not a source-text checker, establish these kills. Bounds are these named permission classes, not every possible mode or filesystem race.

## Hosted validation

Full [workflow 39102151566](https://github.com/relux-works/curator/actions/runs/39102151566) succeeded on exact signed `04f9d8d2f06d6ca2df2e0e1b06f78078445fca35`: 20 successful jobs, two expected conditional skips (`Candidate suite (${{ matrix.os }})`, `Test (rose-air)`). `gh run watch 39102151566 --exit-status` exited 0. Exact-head/full-job verifier exited 0. Linux/macOS artifacts downloaded with exit 0; permission-stream verifier exited 0. The task-prefixed refresh-permission-evidence JSON preserves six suites and six named mutant failures per platform. Temporary gate branch was deleted only after verifying its exact head and signature.

## Logbook / diagnostics

- Initial refresh-candidate invocation lacked the required agent Git environment; stopped own process, exit 143. Sourced retry exited 0; no credential fallback used.
- Initial protocol pin probe ran a relative Git pathspec from the extracted subdirectory and selected zero files, exit 1. Root-relative check passed, exit 0: 1423 conformance blobs, 1295 under v1, zero mismatches.
- cli-13 initially and on its first replay exited 1 before named tests: TestMain's five-minute host-GOROOT lock acquisition timed out. No test coverage or mutant kill is claimed from that stream. Its original log is retained separately; exact replay exited 0 and is the accepted coverage.
- packages-01 and cli-01 initially exited 0 but began before checkpoint refresh ended. Those streams were superseded and replayed with exit 0 to remove VCS/build-metadata ambiguity; the build replay exited 0 for the same reason.
- No product gate, test assertion, timeout or source file was changed to accommodate validation. All fixture homes are disposable; no user home/session state was reset.

## Local validation

Darwin/arm64, Go 1.25.5, Node 25.6.1, pnpm 10.33.0, Rust 1.91.0. Protocol pin `43bf0a2506d5c354a73bbc3ea4623d4653db10c7`; all 1423 conformance blobs (1295 v1) match committed bytes. Fixture processes remove inherited Git override variables and use the isolated CLI/Pnpm PATH from revision 1.

| Command | Real exit | Result |
| --- | ---: | --- |
| 71 standalone `go test -json -count=1 -p 2 -timeout=8m` commands, exact package lists/masks in refresh-commands JSON | 0 each | 85/85 packages, 425/425 CLI, 363/363 install, 375/375 envprofile top-level test names |
| `python3 <scratch>/collect-refresh.py` | 0 | Exact package/test inventories; every accepted source stream has a terminal package verdict |
| `python3 <scratch>/collect-refresh.py --control-missing-batch` | 1 (expected red) | Rejects the specifically planted `known-missing-batch`; not reported as a passing gate |
| `bash .github/ci/platform-case-gate.sh <scratch>/local-effective.json <scratch>/local-platform` | 0 | Platform ledger accepts only classified skips |
| `go build -o <scratch>/curator ./cmd/curator` (build-02) | 0 | Refreshed candidate compiles |
| `go vet ./...` | 0 | Static validation |
| `GOLANGCI_LINT_CACHE=<scratch>/lint-cache golangci-lint run` | 0 | No lint issues |
| `git diff --check` and changed Go files' formatting check | 0 | Clean |
| `bash .github/ci/ledger-consistency.sh <scratch>/ledger` | 0 | Ledger consistency |
| `bash .github/ci/suite-plan.sh <verified-pin-root> <scratch>/plan` | 0 | 85 served packages, no local exclusions |
| `python3 <scratch>/check-integrity.py` | 0 | Candidate signature/tree, unchanged real HEAD/index, pinned bytes; local main not ahead |

The collector observed 12406 named pass/skip verdicts. Every package was replayed in this round; no revision-1 test stream substitutes for refreshed coverage. Preserved reviewer tests remain under their original names and green as part of that inventory. Full hosted evidence above independently covers Linux, macOS and Windows and toolchain/race/gate jobs.

Sanitized raw evidence archive retained at `.temp/BUG-261018-dbeghc/refresh-402fdbc9/BUG-261018-dbeghc_refresh-402fdbc9.zip` (1232776 bytes), SHA256 `e55af5b6f83e410cec7daf696ad262f1378d3035c713a33a4e2365049760e3c0`. The task-prefixed refresh-commands JSON binds exact command vectors, actual exits, raw source hashes and sanitized-file hashes. Board attachment supports plain-text evidence; raw archives remain task-local and are not committed. Redactions remove workspace/home/temp paths and configured human identity while retaining diagnostics and verdicts.

Handoff runs the configured landing gate again synchronously; its exact signed snapshot and result are retained by the Change Request validation resource. No review acceptance or integration is claimed here.
