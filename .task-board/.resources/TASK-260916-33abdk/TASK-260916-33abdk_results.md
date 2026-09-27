# TASK-260916-33abdk — Revision 3 re-apply and review response

## Normative basis

Pinned specification: curator-spec v1.0.0-rc.13, commit `23435129ebc4c29e5b7f75ec72a0aa0cd3f16065`.

- `protocol/environments.md` §7.4, “Codex seed MCP rule,” Revision B: provisioning removes the full top-level `mcp_servers` table and subtables, retains other TOML members, requires the resulting file to parse, reports stripped names once, and records names (including an empty snapshot) at provisioning. §7.4 also says an existing home keeps its bytes.
- §§7.7, 7.8, and 12 define the per-home status disposition and the profile MCP launch channel layered over the seeded base.
- §8.2 permits the closed `{revision, native_mcp_servers}` record on schema-1 and schema-2 managed Codex markers. Absence on a pre-rule home is reported as `mcp_seed_unstripped`; unknown fields remain rejected.

## Rules → production entries → vectors

| Rule | Production entry and regression | rc.13 result |
|---|---|---|
| Revision B strips native MCP tables at first provisioning, preserves other TOML members, records sorted names, warns once for non-empty snapshots, and refuses malformed TOML before publication. | `curator env resolve codex_cli --repair` → `cmdEnvResolve` → `envprofile.Resolve` → `gatherSeeds` → `stripCodexSeedMCPServers`. New CLI regression: `TestEnvResolveStripsAndReportsInlineNativeCodexMCPTable`. | Provisioning vectors: **4/7 driven**, **3/7 bounded** as Revision-A behavior, 0 known gaps, 0 skipped. Driven: `b-strips-servers-keeps-rest`, `b-without-servers-no-warning`, `b-subtable-only-form-stripped`, `b-empty-mcp-servers-table-no-warning`. Bounds: `a-copies-whole-with-servers`, `a-without-servers-no-warning`, `a-inline-table-form-inherited`. |
| Env status reports the shipped revision, sorted per-home names, and the correct disposition for B, A, and pre-rule homes. | `curator env status` → `cmdEnvStatus` → `StatusOf` / `codexSeedStatusWarnings` → `printEnvStatus`. CLI regressions cover stripped nested and inline declarations. | Posture vectors: **5/8 driven**, **3/8 bounded** as Revision-A manager behavior, 0 known gaps, 0 skipped. Bounds: `a-home-lists-ungoverned`, `pre-rule-home-unstripped-under-a`, `a-home-empty-snapshot-no-rows`. Combined seed vectors: **9/15 driven, 6 bounded, 0 gaps, 0 skipped**. |
| Marker v1 accepts the new record only in its closed schema shape; old pre-rule homes remain unmodified by metadata-only repair. | `envmarker.Parse` / `Validate`; CLI regression `TestEnvResolvePreservesPreRuleCodexSeedAndReportsUnstrippedHome` checks the marker and seeded TOML bytes, then checks `env status`. `TestEnvResolveKeepsSchema1BytesForMetadataOnly` checks a schema-1 marker that already has the required empty snapshot. | v1 schema cases: **60/61 driven**, 1 bound, 0 indexed skips. Both E3-owned valid-record rows pass and were removed from the ledger. Unknown-field and malformed-record cases remain rejected. Twelve published-but-unindexed fixtures are reported separately by the harness and are outside the indexed denominator. |
| Existing declared MCP surfacing continues to reach the launch configuration. | Existing production-entry consumers `TestInstallSurfacesMCPDeclarations` and `TestUpdateSurfacesCandidateMCPSet`. | Both passed in the pinned-root envprofile run; profile MCP surfacing from TASK-260910-gocke2 remains intact. |

The rc.13 suite was extracted from commit `23435129` under `$TMPDIR`; the pinned vector file contains 7 provisioning and 8 posture cases. Counts `environments-codex-seed/provisioning-cases=7` and `.../posture-cases=8` match that suite.

## Gap ledger and re-apply

On trunk `eca2bf27`, STORY-owned rows were **2 → 0**; total case rows were **69 → 67**, and non-comment ledger records including the header were **70 → 68**. Removed only:

- `agent-environment-marker-v1/schema-cases/valid-codex-seed-record.json`
- `agent-environment-marker-v1/schema-cases/valid-codex-seed-record-empty-snapshot.json`

The three-way patch applied cleanly except for the expected `conformance-gaps.tsv` conflict (git apply exit 1). Resolution retained trunk’s rows and removed only those two STORY rows. The two E3 count pins were computed from the pinned suite; trunk’s unrelated `environments-read-failure/restore=2` pin was retained. The final repository delta is exactly the 11 accepted rev2 paths. Trunk-side registry posture and manager read-failure changes in `cmd/curator/envstatus.go` and `internal/envprofile/status.go` were preserved.

## Review-round regression and mutants

The review-round brief describes the rev2 verdict as a rejection, but the attached `TASK-260916-33abdk_review-verdict-rev2.md` is explicitly **ACCEPTED**, and no separate rejection resource is attached. This revision adds a named regression for the concrete §7.4 pre-rule-home edge: a schema-1 marker without `codex_seed_record` and a home that still has native MCP bytes must remain byte-stable on metadata-only repair and be reported as `mcp_seed_unstripped`.

| Mutant | Before regression | After regression |
|---|---|---|
| Narrow marker mutant: metadata-only repair backfills an empty revision-B record only when an existing Codex marker has no record. The new regression was temporarily renamed outside the selection. | `go test ./cmd/curator -run 'EnvResolve|Marker|Credential|Seed|Mcp|EnvStatus' -count=1` → **exit 0** (527.354s); mutant survived the existing focused selection. | `go test ./cmd/curator -run '^TestEnvResolvePreservesPreRuleCodexSeedAndReportsUnstrippedHome$' -count=1` → **exit 1**; the test observed the schema-1 marker rewritten with the injected empty record. Mutant killed. |
| Narrow strip mutant: report inline-table names but return the original `config.toml` when input uses `mcp_servers = { ... }`. | Current `TestEnvResolveStripsAndReportsInlineNativeCodexMCPTable` without mutation → **exit 0**. | With mutant: same named test → **exit 1** because `mcp_servers` and its command reached managed `config.toml`. Mutant killed. |
| Strip rule removed / B status report removed / provisioning warning suppressed. | Existing rev2 baseline evidence: strip-removal mutant survived `TestResolveProvisionRepair` (**exit 0**). | Existing rev2 evidence: strip removal was killed by `TestCodexSeedProvisioningAndStatus` (**exit 1**); B status-report removal by envprofile and CLI status tests (**exit 1**); warning suppression by three provisioning tests (**exit 1**). The independent accepted rev2 verdict records these runs; they were not rerun during this re-apply. |

Both temporary mutants were restored from `$TMPDIR` copies; no mutation remains in the worktree.

## Local validation

Commands ran as standalone processes. All conformance tests used `CURATOR_CONFORMANCE_ROOT=$TMPDIR/curator-spec-rc13/conformance/v1`.

| Command | Exit |
|---|---:|
| `go test ./cmd/curator -run '^TestEnvResolveStripsAndReportsInlineNativeCodexMCPTable$' -count=1` | 0 |
| `go test ./cmd/curator -run '^TestEnvResolvePreservesPreRuleCodexSeedAndReportsUnstrippedHome$' -count=1` | 0 |
| `env CURATOR_CONFORMANCE_ROOT=$TMPDIR/curator-spec-rc13/conformance/v1 go test ./internal/envprofile -run 'Seed|Mcp|MCP|Codex|Status|Guarded' -count=1 -v` | 0 (141.406s); vectors 4/7 + 5/8; stateread guard **374/374** reads covered, 260 via seam and 114 allowlisted |
| `env CURATOR_CONFORMANCE_ROOT=$TMPDIR/curator-spec-rc13/conformance/v1 go test ./cmd/curator -run 'EnvResolve|Marker|Seed|Mcp|EnvStatus' -count=1` | 0 (447.627s) |
| `env CURATOR_CONFORMANCE_ROOT=$TMPDIR/curator-spec-rc13/conformance/v1 go test ./internal/envmarker -run '^TestParseAuthoritativeEnvMarkerSchemaCases$' -count=1 -v` | 0 (60 driven, 1 bound) |
| `go vet ./internal/envprofile ./internal/envmarker ./cmd/curator` | 0 |
| `go build -o "$TMPDIR/curator-33abdk" ./cmd/curator` | 0 |
| `golangci-lint run ./internal/envprofile ./cmd/curator` | 0 (0 issues) |
| `git diff --check` | 0 |

An initial compile of the new marker regression failed (exit 1) because its test file lacked the `strings` import. The import was added; the named test and final focused command then passed. No full hosted/landing suite was run locally; `task-board handoff` will publish the candidate and run the configured hosted gate once. Its attached validation log is the authoritative gate result.

## Scope and release prep

`CHANGELOG.md` and `LOGBOOK.md` are unchanged. No files outside the Story worktree were changed; temporary archives, mutants, logs, and the build binary are under `$TMPDIR`. The worktree remains uncommitted for handoff.

## CHANGELOG entry (for release prep)

Managed Codex provisioning strips native `mcp_servers` from `config.toml`, records sorted native server names, and reports stripped or pre-rule servers in `env status`.

## Revision 4 (carry-forward republish)

Revision 3 was ACCEPTED on content. Trunk moved to `aa7d8d09`; the orchestrator
ran `worktree converge STORY-260916-1i1gfo` and the accepted delta is carried
uncommitted in the Story worktree. This revision republishes that delta after
the carry-forward checks below. No product-code change was made in revision 4.

### Carry-forward verification (per path of `TASK-260916-33abdk_change-request_rev3.patch`)

Method: the patch's `index <pre>..<post>` lines give the rev3 base and
post-image blob for each tracked path. `git rev-parse HEAD:<path>` tells
whether trunk touched the path since the rev3 base; `git hash-object <path>`
tells whether the worktree file is byte-identical to the rev3 post-image.
New-file post-images were checked the same way (`0e9f70df`, `9c87de88`).

Byte-identical to revision 3, trunk untouched (HEAD == pre-image,
worktree == post-image):

- `.github/ci/conformance-case-counts.tsv` (pre `d9edd023`, work `abe4c703…`)
- `.github/ci/conformance-gaps.tsv` (pre `b23b0080`, work `008af583…`)
- `.github/ci/root-artifacts.tsv` (pre `243e24ad`, work `af1c942d…`)
- `cmd/curator/env_credential_marker_test.go` (pre `3fc216eb`, work `33cf1451…`)
- `internal/envmarker/envmarker.go` (pre `3d92a7ee`, work `c9395ef0…`)
- `internal/envprofile/managed.go` (pre `42c0dc15`, work `3e4d621f…`)
- `internal/envprofile/status.go` (pre `efd9dee3`, work `29886966…`)
- `internal/envregistry/envregistry.go` (pre `d146834f`, work `03f98cba…`)
- `cmd/curator/envstatus_test.go` (new file, work `0e9f70df…` == post-image)
- `internal/envprofile/codex_seed_test.go` (new file, work `9c87de88…` == post-image)

Intersecting path (trunk moved under the delta; three-way merge result keeps
both sides, no conflict markers):

- `cmd/curator/envstatus.go`: trunk `b5bd638e → e9e687b6` refactored
  `attachProviderPosture` to call the new `providerPostureForConfig` /
  `providerPostureWithPath` (umbrellas, STORY-260916-2otjbn); rev3 adds the
  `codex-seed:` revision line and the `codex-seed-record:` home block in
  disjoint hunks. Worktree `e157c62c` contains both: rev3 lines 39/84/86
  present, trunk lines 181/232/238 present, and
  `git diff HEAD -- cmd/curator/envstatus.go` is exactly the rev3 hunk.

No `<<<<<<<` / `>>>>>>>` markers in the worktree (the only `=======` hits are
pre-existing comment rules in `.github/ci/platform-cases.tsv`, untouched by
this task).

### CHANGELOG policy

`TASK-260916-33abdk_change-request_rev3.patch` contains no CHANGELOG.md /
LOGBOOK.md hunk. `git status --short -- CHANGELOG.md` and
`git diff --name-only HEAD -- CHANGELOG.md` are both empty: the file equals
trunk's. No stray root `TASK-*.md` / `BUG-*.md`, `test/` or ledger paths were
added. The entry text stays in the results resource only:

## CHANGELOG entry (for release prep)

Managed Codex provisioning strips native `mcp_servers` from `config.toml`, records sorted native server names, and reports stripped or pre-rule servers in `env status`.

### Stale-snapshot check

`git diff --name-only HEAD -- . ':!.task-board'` (exit 0) lists exactly the 9
tracked rev3 paths; `git status` adds only the 2 untracked rev3 new test
files above (hash-verified). No other file is modified, deleted, or untracked:
the worktree reverts nothing of trunk. Not a stale snapshot — publish proceeds.

### Revision 4 validation (focused bounded run, standalone processes)

Conformance root: fresh shallow checkout of `relux-works/curator-spec` tag
`v1.0.0-rc.13` at `/tmp/curator-spec-rc13`; HEAD `23435129…` == SPEC_PIN and
`conformance/v1/manifest.json` digest `be11bb1e…` matches
`.github/workflows/ci.yml`. (Outside the worktree scratch per policy; the
prior run's `$TMPDIR` checkout no longer exists.)

| Command | Exit |
|---|---:|
| `env CURATOR_CONFORMANCE_ROOT=/tmp/curator-spec-rc13/conformance/v1 go test ./internal/envprofile -run 'Seed\|Mcp\|MCP\|Codex\|Status\|Guarded' -count=1` | 0 (179.254s) |
| `env CURATOR_CONFORMANCE_ROOT=/tmp/curator-spec-rc13/conformance/v1 go test ./cmd/curator -run 'EnvResolve\|Marker\|Seed\|Mcp\|EnvStatus\|Umbrella\|Provider' -count=1` | 0 (536.984s) |

The `Umbrella|Provider` mask additionally covers the trunk-side
`providerPostureForConfig` refactor on the intersecting path; the envprofile
mask covers the rev3 seed/status scope.
