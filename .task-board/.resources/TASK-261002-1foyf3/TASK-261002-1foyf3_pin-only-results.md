# TASK-261002-1foyf3 — rc14-pin-and-v2-writer-cutover (pin-only rescope)

Scope: apply cutover-rescope.md. v2 writer and atomic profile migration are deferred to TASK-261003-1uzji7. Working-tree changes remain uncommitted for the developer handoff.

SPEC_PIN: `43bf0a2506d5c354a73bbc3ea4623d4653db10c7`. Fresh `git ls-remote` reports rc.14 tag object `661bead088186db70c9f57ad0b115305ec2bef35` and the same peeled commit. The withdrawn draft tag object is not the current release.
Checked-out core manifest SHA-256: `6f832d813efc768ea154a7d5076b512ab4be6aa9409d92e11469d21ea9bc69f5` (shasum exit 0).

EnableV2Writers remains false; WriteVersion defaults to v1. Removed the earlier flip-only marker, install, context and environment test adaptations, and the draft rc14_cutover_test.go. CodexSeedRevision and SecurityPostureRevision remain A. The rc.8 module release pin, CHANGELOG.md and LOGBOOK.md are byte-identical to HEAD.

Snapshot gap retained with the exact requested reason and owner TASK-261003-1uzji7. No gap removed. Snapshot production entry: gitops.Extract, used by internal/snapshot and internal/closure. Actual rc.14 tally: 0 driven, 1 known-gap, 0 bound, 0 skipped, 1 total (1/1 accounted). Both autocrlf settings preserve bytes and produce v1 hash `500ea934403d10a2a0b6b7e8874790e489ee002328d3dc0edbda2fe5be2bced0`; rc.14 expects v2 hash `ecca17aacc80360a390b2403016f71e04a29e5d46c6643b77a5ee77b47186a9e`.

Added a regression for the rc.14 gap owner, reason, exact count and tally. Existing tests reject missing classifications, unlisted failures, passing gaps and vanished cases. Historical vectors select frozen v1; rc.14 uses the actual writer selection.

| Validation | Exit | Duration (s) | syspolicyd crashes before → after |
| --- | --- | --- | --- |
| rescope-required | 0 | 3.92 | 379 → 379 |
| rescope-snapshot | 0 | 2.07 | 379 → 379 |
| rescope-build | 0 | 3.2 | 379 → 379 |
| rescope-lint | 1 | 29.76 | 379 → 379 |
| rescope-lint-fresh | 0 | 21.8 | 379 → 379 |
| rescope-rc13 | 1 | 2.56 | 379 → 379 |
| rescope-rc13-checkout | 0 | 2.74 | 379 → 379 |

Every Go build/test used -work and the shared host build lock. syspolicyd was running before and after each check. No cooldown was triggered. Each command ran directly as a standalone subprocess without a pipe or tee.

Initial lint exit 1 was stale cached diagnostics naming another worktree. Re-run with isolated GOLANGCI_LINT_CACHE and the CI-pinned golangci-lint v2.12.2 exited 0 with 0 issues.
Initial optional rc.13 compatibility exit 1 was invalid corpus materialization: git archive expanded export-subst fixture bytes. Re-materialized via an actual detached checkout at `23435129ebc4c29e5b7f75ec72a0aa0cd3f16065`, verified manifest `be11bb1e4c46f21fb5684d586f9c2a8b0d59f3b437bc7ea7aa5aa530fe4d47ca`, then re-ran the same full required package list successfully. No product code was changed to compensate for either local evidence issue.

Other validation: naming-gate.sh exit 0; git diff --check exit 0; changed Go files gofmt check exit 0; unchanged-switch/release-pin/CHANGELOG/LOGBOOK checks exit 0.

Hosted rc.14 default-pin matrix: NOT RUN for this uncommitted candidate. The most recent hosted green https://github.com/relux-works/curator/actions/runs/37141634038 is for 68210ecc with the previous rc.13 pin; it is not accepted as evidence for this draft. Publishing the candidate and obtaining the hosted rc.14 gate remains with integration/review. No earlier attached build or test evidence was reused.

Checklist rescope: the first handoff exited 1 because original checklist items 1 and 7 required the superseded writer flip and forbidden LOGBOOK.md edit. Replaced those obsolete entries through the board CLI with verification of the pin-only scope and recording findings in this outcome while leaving LOGBOOK.md unchanged. These replacement items are satisfied by the validation above. This administrative correction changes no repository code or validation input.

## Command output

### rescope-required

`go test -work ./internal/hashing ./internal/marker ./internal/conformancecoverage ./internal/interop/... -count=1`

```text
WORK=<retained-work>
ok  	github.com/relux-works/curator/internal/hashing	0.652s
ok  	github.com/relux-works/curator/internal/marker	2.046s
ok  	github.com/relux-works/curator/internal/conformancecoverage	0.864s
ok  	github.com/relux-works/curator/internal/interop	0.564s
ok  	github.com/relux-works/curator/internal/interop/environments	1.836s
```

### rescope-snapshot

`go test -work ./internal/interop/environments -run ^TestConformanceSnapshotAcquisition$ -count=1 -v`

```text
WORK=<retained-work>
=== RUN   TestConformanceSnapshotAcquisition
=== RUN   TestConformanceSnapshotAcquisition/byte-exact-snapshot
=== RUN   TestConformanceSnapshotAcquisition/byte-exact-snapshot/autocrlf=true
    snapshot_acquisition_test.go:172: autocrlf=true: content hash sha256:500ea934403d10a2a0b6b7e8874790e489ee002328d3dc0edbda2fe5be2bced0, want sha256:ecca17aacc80360a390b2403016f71e04a29e5d46c6643b77a5ee77b47186a9e
=== RUN   TestConformanceSnapshotAcquisition/byte-exact-snapshot/autocrlf=false
    snapshot_acquisition_test.go:172: autocrlf=false: content hash sha256:500ea934403d10a2a0b6b7e8874790e489ee002328d3dc0edbda2fe5be2bced0, want sha256:ecca17aacc80360a390b2403016f71e04a29e5d46c6643b77a5ee77b47186a9e
=== NAME  TestConformanceSnapshotAcquisition
    coverage.go:236: published cases snapshot-acquisition/cases: 0 driven, 1 known-gap, 0 bound, 0 skipped, 1 total
--- PASS: TestConformanceSnapshotAcquisition (0.52s)
    --- PASS: TestConformanceSnapshotAcquisition/byte-exact-snapshot (0.51s)
        --- PASS: TestConformanceSnapshotAcquisition/byte-exact-snapshot/autocrlf=true (0.09s)
        --- PASS: TestConformanceSnapshotAcquisition/byte-exact-snapshot/autocrlf=false (0.10s)
PASS
ok  	github.com/relux-works/curator/internal/interop/environments	0.848s
```

### rescope-build

`go build -work -o .temp/TASK-261002-1foyf3/curator-pin-build ./cmd/curator`

```text
WORK=<retained-work>
```

### rescope-lint

`golangci-lint run --timeout=7m`

```text
level=warning msg="[runner] Can't process results by generated_file_filter processor: can't filter issue &result.Issue{FromLinter:\"gosec\", Text:\"G115: integer overflow conversion int -> byte\", Severity:\"high\", SourceLines:[]string(nil), Pkg:(*packages.Package)(0x27d7bdfd2ea0), Pos:token.Position{Filename:\"<other-worktree>\", Offset:0, Line:168, Column:33}, LineRange:(*result.Range)(nil), HunkPos:0, SuggestedFixes:[]analysis.SuggestedFix(nil), ExpectNoLint:false, ExpectedNoLintLinter:\"\", WorkingDirectoryRelativePath:\"../../STORY-260923-1lu2o3/worktree/internal/closuregraph/plan_test.go\", RelativePath:\"../../STORY-260923-1lu2o3/worktree/internal/closuregraph/plan_test.go\"}: failed to get doc (strict) of file <other-worktree> failed to parse file: open <other-worktree> no such file or directory"
level=warning msg="[runner/source_code] Failed to get line 66 for file <other-worktree> failed to get file <other-worktree> lines cache: can't get file <other-worktree> bytes from cache: can't read file <other-worktree> open <other-worktree> no such file or directory"
../../STORY-260923-1lu2o3/worktree/tools/goreleaserconfig/gate.go:66:15: G304: Potential file inclusion via variable (gosec)
1 issues:
* gosec: 1
```

### rescope-lint-fresh

`golangci-lint run --timeout=7m`

```text
0 issues.
```

### rescope-rc13

`go test -work ./internal/hashing ./internal/marker ./internal/conformancecoverage ./internal/interop/... -count=1`

```text
WORK=<retained-work>
ok  	github.com/relux-works/curator/internal/hashing	0.348s
ok  	github.com/relux-works/curator/internal/marker	1.752s
ok  	github.com/relux-works/curator/internal/conformancecoverage	0.831s
ok  	github.com/relux-works/curator/internal/interop	1.192s
--- FAIL: TestConformanceSnapshotAcquisition (0.27s)
    --- FAIL: TestConformanceSnapshotAcquisition/byte-exact-snapshot (0.27s)
        snapshot_acquisition_test.go:98: fixture subst.txt on disk (sha256:a1ab8edbd48667c39da619a7cc4bad17fe5e96eb218da72de150ae7fa4e93849, 65 bytes) does not match the vector (sha256:ec9a6c8c260b3977dad2f1cba09414599879ac5c3e7353c3205a40d5fbe122bc, 40 bytes); the checkout normalized it
    coverage.go:234: published-case coverage: failing published case snapshot-acquisition/cases/byte-exact-snapshot is not listed in the gap ledger: case test failed
    coverage.go:236: published cases snapshot-acquisition/cases: 0 driven, 0 known-gap, 0 bound, 0 skipped, 0 total
FAIL
FAIL	github.com/relux-works/curator/internal/interop/environments	1.883s
FAIL
```

### rescope-rc13-checkout

`go test -work ./internal/hashing ./internal/marker ./internal/conformancecoverage ./internal/interop/... -count=1`

```text
WORK=<retained-work>
ok  	github.com/relux-works/curator/internal/hashing	0.323s
ok  	github.com/relux-works/curator/internal/marker	1.562s
ok  	github.com/relux-works/curator/internal/conformancecoverage	0.825s
ok  	github.com/relux-works/curator/internal/interop	1.138s
ok  	github.com/relux-works/curator/internal/interop/environments	2.187s
```

