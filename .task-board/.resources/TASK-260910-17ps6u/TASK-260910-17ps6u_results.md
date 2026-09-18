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

## Revision 2 (rework 1: hosted gate failure on TestMarkerRefusalSeparatesUnsupportedFromInvalid)

Cause: this leaf makes marker schema 5 readable (`marker.SupportedSchema` includes `SchemaV5`), so `{"schema_version":5,"name":"build-skill"}` is correctly `invalid-marker`; the cmd/curator test still used the literal 5 as "a schema from a newer manager". No production change; all revision-1 code kept as is.

Changed (only `cmd/curator/builds_test.go`):
- "schema from a newer manager" row now uses `marker.NewestSchemaVersion+1` (derived, no literal; no new exported API).
- Added row "readable draft schema 5 that is still not a valid marker" → `stateInvalidMarker`.
- Comment above the readable-schema rows updated to "Schemas 3, 4 and 5 are readable".
- All other rows unchanged.

Grep for other schema-5-as-newer assumptions (`grep -rnE '"schema_version": ?5|SchemaVersion: ?5\b|newer manager' --include='*_test.go' cmd internal`):
- cmd/curator/status_test.go:684-696 already derives from `marker.NewestSchemaVersion` (+1 for unsupported, exact for readable-but-invalid) — no change.
- internal/marker/schema_band_test.go already asserts the band against `NewestSchemaVersion` — no change.
- internal/install/install_test.go:836, internal/skillspec/{fallback,build,builddriver_conformance,parse}_test.go, internal/audit/vendor_test.go: `schema_version: 5` there is the *skill spec* (csk-skill.json) schema, not the marker schema — unrelated, no change.
- No other cmd/curator or repository test treated marker schema 5 as unsupported.

Commands (zsh, each run as a standalone process, real exit codes):
| command | exit |
|---|---|
| gofmt -l cmd/curator internal/marker | 0, no output |
| go vet ./cmd/curator | 0 |
| go test -p 1 ./cmd/curator -run 'TestMarkerRefusalSeparatesUnsupportedFromInvalid\|TestStatus.*Marker\|Marker' -count=1 -timeout=180s | 0 (3.2s) |
| go test -p 1 ./internal/marker -count=1 | 0 |
| go test ./cmd/curator -count=1 (full package, once, background to file) | 0 (671s) |

Not rerun in revision 2: the install/runtimestore/scopes package tests from revision 1 (no files in those packages changed in this revision; their revision-1 evidence stands). The hosted gate runs the landing suite once at handoff.

## Revision 3 (rework 2: Windows-only gate failure in the two new draft runtime tests)

Cause: `writeDraftScriptSkill` (internal/install/draftruntime_test.go) declared only `unix_path`; `runtimestore.commandRel` selects `WinPath` on windows, so the command path was empty and `validateScriptSpec` refused it ("path is not portable"). ubuntu/macos were green in gate run 35309565085. No production change; all revision-1/2 code kept as is.

Changed (only `internal/install/draftruntime_test.go`):
- `writeDraftScriptSkill` now declares both `"unix_path": "scripts/<cmd>.sh"` and `"win_path": "scripts/<cmd>.sh"` for every fixture command. The portability check (`identifiers.PortablePath`) is a path-shape rule, not an extension rule, so the same `.sh` file satisfies the Windows arm; materialization, runtime-store keys, marker v5 binding, refresh replacement and the no-links-in-protected-tree assertions now run on Windows (no `.cmd` file needed).
- The existing Windows skip around shim execution (POSIX scripts run via `exec.Command`) now uses the declared reason verbatim: `executes POSIX skill commands` (matches install_test.go:282/326). No skip class added; the skip still covers only the execution block, after every materialization assertion.
- Checked the other new tests of this leaf (gc_draft_test.go, marker_v5_test.go, sourcev1_test.go, remainder of draftruntime_test.go): no other `unix_path`-only command fixtures (the schema-8 fixture at draftruntime_test.go:379 already declares both), no `exec.Command`, symlink creation or executable-bit assertions, so no further skips were needed.

Commands (zsh, each run as a standalone process, real exit codes, macOS arm64):
| command | exit |
|---|---|
| go test -p 1 ./internal/install -run 'TestDraftLocal(RuntimeMaterializesFromFrozenSnapshot\|RefreshReplacesFrozenRuntime)' -count=1 -timeout=180s | 0 (19.4s) |
| go vet ./internal/install | 0 |
| gofmt -l internal/install | 0, no output |
| go test -p 1 ./internal/install -run 'TestDraft' -count=1 -timeout=180s | 0 (180.0s wall, all draft install tests) |

Windows is verified only by the hosted gate (no Windows host here); the handoff triggers it once. Not rerun in revision 3: runtimestore/scopes/marker/cmd tests (no files in those packages changed in this revision; revision-1/2 evidence stands).

## Revision 4 (rework 3: review F1 — malformed v5 package union accepted as current)

Cause (reviewer's finding, confirmed): `validV5Package` checked decoded values; the shared `Package` struct knows every arm's fields and JSON `null`/`""` collapse to zero values, so a local-snapshot package carrying `source:null`, `repository:null`, `directory:null` or `commit:null` was readable and current.

Changed (production, only `internal/marker/marker.go`):
- New `validV5PackageShape(raw["package"])` runs in `validV5Identity` BEFORE the decoded-value check. It validates the raw package object against a closed per-arm member set with JSON types (`v5PackageArms`: local-snapshot = {kind, snapshot}; network-git = {kind, repository, commit, directory}; configured-git = {kind, source, commit, directory}), requires every member present with its type (null is neither absent nor typed → refused), refuses any member outside the arm even when null/empty, and closes the nested commit object the same way (`v5CommitShape` = {object_format, hex}, both strings). `kind` must be a JSON string naming a known arm.
- Duplicate keys and trailing data: `protocoljson.Validate` already runs over the whole marker document at the top of `marker.Read` (same validator the accepted sibling hwxr26 uses), so duplicate keys inside `package`/`commit` are refused there; the new regression rows exercise that path. No hand-rolled duplicate detection added; no foreign fields are dropped silently.
- The decoded-value check (`validV5Package`) is retained as the second layer (digest/hex grammars, directory portability, `.git` suffix).

Tests (`internal/marker/marker_v5_test.go`): `TestMarkerV5PackageClosedShape` — 39 rows through the production entry `marker.Read` + `marker.Current` on a rewritten on-disk document (bypasses `Write`, so the bytes are what a foreign/hostile writer produces). Valid controls per arm (local, network-git, configured-git). Negative rows: every foreign-arm member with null AND empty value for all three arms, required member null/missing, unknown member, `kind` null/number, package null/array, commit null/string/open (extra member)/missing member/hex null, and duplicate-key rows for package and commit. Each negative row asserts `Read == nil`, `Current == false` against the original expectation, and `Current == false` against an expectation mirroring the malformed value. The reviewer's four probe scenarios (source/repository/directory/commit null on local-snapshot) are rows `local-*-null`.

Narrowing mutant M5: restore decoded-value validation only (drop `validV5PackageShape` from `validV5Identity`, keep everything else) — `go test -p 1 ./internal/marker -run TestMarkerV5PackageClosedShape -count=1 -timeout=120s` → exit 1, 14 rows fail (all null/empty foreign-member rows across the three arms). File restored from backup, rebuild verified.

Commands (zsh, each a standalone process, real exit codes, this session):
| command | exit |
|---|---|
| go test -p 1 ./internal/marker -count=1 -timeout=120s | 0 (1.175s) |
| go test -p 1 ./internal/install -run 'TestDraftLocal\|TestLegacyInstallUntouchedWhenDraftOff' -count=1 -timeout=240s (background to file) | 0 (20.979s) |
| go vet ./internal/marker ./internal/install | 0 |
| gofmt -l internal/marker internal/install | 0, no output |
| mutant M5 (see above) | 1 (killed) |

No other scope changes; runtimestore, install, scopes and cmd/curator files untouched in this revision. Windows and the full suite are verified only by the hosted gate at handoff.
