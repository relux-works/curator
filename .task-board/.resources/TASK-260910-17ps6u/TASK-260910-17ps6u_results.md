# TASK-260910-17ps6u — materialize local runtime and command dependencies (handoff evidence)

Candidate: uncommitted working tree of `.temp/STORY-260910-20sx61/worktree` on top of checkpoint 7ce27b20c43baf489d4d11e9c55dc05eb2dc108e (hwxr26). Shell: zsh, `set -o pipefail`, every gate run as a standalone process with real exit codes.

## Provenance
The earlier muse run RUN-260918-a711fb wrote most of this tree before its provider died (402). This run (claude-fable-5-1, RUN-260918-7ca748) reviewed it against the spec, completed it, fixed lint (renamed `marker.MarkerPackage`/`MarkerCommit` to `marker.Package`/`marker.Commit` for the revive stutter rule), and produced all evidence below.

## Changed paths
 M internal/install/global.go
 M internal/install/install.go
 M internal/install/targets.go
 M internal/marker/marker.go
 M internal/marker/schema_band_test.go
 M internal/scopes/gc.go
?? internal/install/draftruntime.go
?? internal/install/draftruntime_test.go
?? internal/marker/marker_v5_test.go
?? internal/runtimestore/sourcev1.go
?? internal/runtimestore/sourcev1_test.go
?? internal/scopes/gc_draft_test.go

## What it implements (skillfile-sources.md protocol §"schema-2 installations", lines 180-248)
- `internal/runtimestore/sourcev1.go`: distinct `source-v1` runtime-store namespace; key = `source-v1-<SHA-256(CCJ-1(package))>` — never the bare snapshot digest, never a Git commit. Malformed digests refused.
- `internal/install/draftruntime.go` + wiring in `install.go`/`targets.go`/`global.go`: on the draft lane, `projectAttempt` derives runtime keys from the consumed lock before staging (unhashable member fails before the first live mutation); `stageRuntimeAndShims` takes an optional key map and materializes each local-snapshot member's declared runtime roots from the immutable frozen snapshot under `home/runtime/<skill>/<source-v1 key>` through the existing `PrepareScriptRuntime` copy path (bytes copied, no links); shims/commands/capabilities/dependency/system-command/script-policy gates unchanged. Global scope passes nil (byte-identical). Local-snapshot members write marker schema 5 (`package` + `lock_sha256`), Git draft members keep their accepted legacy marker shape (interim boundary, documented in code).
- `internal/marker/marker.go`: schema 5 read/write/validate/currentness per the spec migration table (legacy source fields forbidden, attestation/substituted forbidden for local-snapshot, disjoint package arms, lock-bound currentness). `internal/scopes/gc.go`: GC marks draft runtime leaves by the source-v1 key derived from the marker package; unhashable package is reported as uncertain, never silently unmarked.
- Hybrid-store root set on the draft lane comes from lock selection indices (collections carry no manifest declaration names).

## Shared edits / scope note
No files owned by TASK-260910-dufdai (buildcache, buildsource, buildrepo, godriver) touched. `internal/marker` and `internal/scopes` are outside the literal scope line but are the spec-mandated carriers of the "protected runtime-store" semantics (marker-5 package + source-v1 keys followed by status/GC); edits are additive and draft-only. hwxr26's source-audit gate (`draftaudit*.go`) untouched. Marker-5 build entries do NOT yet bind receipt v3 — that is dufdai's scope; `validBuildState` carries v5 through the legacy arm with a comment naming the sibling.

## Tests (production entry: install.Project with DraftSourcesV1=true)
Positive: TestDraftLocalRuntimeMaterializesFromFrozenSnapshot (provider+consumer with skill dependency; runtime under source-v1 key; no links in protected tree; runtime root excluded from context; v5 marker binds package+lock; adapter mirror serves installed context; shim reaches runtime store not live tree; shims actually execute on POSIX; reinstall reports up-to-date). TestDraftLocalRefreshReplacesFrozenRuntime (live edit stays pinned; explicit refresh moves the key and publishes a new frozen tree; marker rebound).
Negative (fail closed, no side effects under .agents or home/runtime): tampered frozen runtime → source_snapshot_changed; missing snapshot → source_snapshot_unavailable; enforced script policy → script_execution_policy_unsupported; missing system command; absent provider / provider without script command; invalid capability → source_member_invalid at resolve. Marker: 12 malformed-identity rows refused, legacy-vs-v5 currentness, arm confusion refused. Runtimestore: malformed digest rows. GC: live draft leaf retained, stale commit leaf swept.
Windows: POSIX execution skipped with declared platform-control reason "executes POSIX skill commands only on unix runners".

## Commands and exit codes (all run by me in this session)
| command | exit |
|---|---|
| go build ./... | 0 |
| go vet ./internal/install/ ./internal/marker/ ./internal/runtimestore/ ./internal/scopes/ | 0 |
| gofmt -l (touched pkgs) | 0, no output |
| golangci-lint run (touched pkgs) after rename | 0 |
| go test -p 1 -count=1 ./internal/runtimestore/ ./internal/marker/ | 0 |
| go test -p 1 -count=1 ./internal/scopes/ ./internal/closure/ ./internal/sourcelock/ | 0 |
| go test -p 1 -count=1 -run TestDraft ./internal/install/ (rerun after rename) | 0 (197s) |
| go test -p 1 -count=1 -run 'TestLegacyInstallUntouchedWhenDraftOff\|TestRuntimeLauncher\|TestRuntimeOnly\|TestGlobalInstall$\|TestProjectInstall\|Audit' ./internal/install/ | 0 (199s) |

Not run: the full install package and the full repository suite (host stalls; the remote gate runs the landing suite once at handoff). The legacy install evidence above is a bounded subset; byte-identity of legacy goldens with the switch off rests on TestLegacyInstallUntouchedWhenDraftOff plus the nil-key/nil-lock paths being unchanged.

## Narrowing mutants (all killed, files restored, rebuild verified)
| mutant | test that killed it | exit |
|---|---|---|
| M1 SourceV1Key drops the source-v1 namespace (bare digest leaf) | TestSourceV1KeyShape, TestSourceV1DirLayout | 1 |
| M2 marker.Current ignores lock_sha256 | TestMarkerV5Currentness | 1 |
| M3 install draftRuntimeKey ignores the frozen key map (falls back to commit) | TestDraftLocalRuntimeMaterializesFromFrozenSnapshot | 1 |
| M4 GC marks draft markers by commit instead of source-v1 key | TestCollectRetainsDraftSourceV1Runtime | 1 |

## Findings / anomalies
- No `logbook` CLI on this host and LOGBOOK.md edits are forbidden by campaign rules; findings recorded here and in board notes instead.
- Bound: Git draft members still run commit-keyed runtime leaves with legacy markers; migrating them to marker-5/source-v1 is deferred to the Story integration leaf (spec requires marker 5 for every schema-2 installation; only local-snapshot members reach it in this leaf).
