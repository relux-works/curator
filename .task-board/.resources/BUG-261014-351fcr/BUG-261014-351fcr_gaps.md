# BUG-261014-351fcr — migrated-global-skill-audited-under-v1-hash

## Candidate and scope

Base/checkpoint: `84cf54b94670ae9881f91b1c147feb4b75afddee`, equal to freshly advertised and fetched `origin/main` at development start. Managed Story HEAD remains at that checkpoint; no producer commit, branch change, or integration was performed. The handoff owns candidate snapshotting.

The candidate changes only `internal/envprofile/envprofile.go` and adds `internal/envprofile/global_migration_hash_version_test.go`. Both migration audit selections use the existing `hashing.WriteVersion()` rule: the direct `contextaudit.DetectAtVersion` call and the migrated `contextresolve.Resolved.HashVersion` consumed by `strictAuditMember`. Existing lock serialization already writes schema 2/hash version 2 under v2 and required no change. No path exemption was added.

Candidate source SHA-256:

- `internal/envprofile/envprofile.go`: `3aa398279f16918fa25eb33d3916482dc9664cb820fd762ca77084cbbe50ab5c`
- `internal/envprofile/global_migration_hash_version_test.go`: `c6dd17cdd40952c88c7d0b894a25b761d75c7ac0dbcdd47f33b7844dabf80db2`

Hosted snapshot: `f34a0db8be5b13282f7bb4742b5d8a21ead874cf`. Its test source was compared byte-for-byte to the candidate. The candidate is frozen after mutation restoration; later edits are ignored task evidence only.

## AC coverage — 4 of 4 AC rows driven

| AC row | Named test | Production call site and assertion |
| --- | --- | --- |
| 1 — gzip and binary migrate under v2; lock records version 2 | `TestMigratedGlobalSkillBinaryWriteVersion` (`gzip-and-binary/v2`, both directories) | `EnsureDefaultWithPolicy -> ensureDefault -> migrateGlobalSkills -> DetectAtVersion / strictAuditMember`; reads the published lock through `readLock`, checks schema/hash 2, canonical skill source/exact commit, validates lock, and reads stored gzip/binary bytes unchanged. |
| 2 — same fixture under forced v1 refuses with opaque-file | `TestMigratedGlobalSkillBinaryWriteVersion` (all `/v1` cases) | Same production migration entry, with existing `hashing.EnableV2Writers` seam restored by cleanup. Requires `profile_source_invalid`, `opaque-file`, and the fixture directory in the refusal; requires both profile source and lock to remain unpublished. |
| 3 — no path exemption | `TestMigratedGlobalSkillBinaryWriteVersion` (`.task-board/attachments` and `assets/deep`, gzip alone, binary alone, both, v1/v2) | Same production migration entry. Isolated file cases prevent one opaque file from masking a file-type exemption. The narrowing directory-exemption mutant is killed by forced-v1 cases. Self-review confirms the production diff contains only two version selections. |
| 4 — network follow-up covered or separately filed | `TestMigratedGlobalSkillNetworkUndeclaredFollowUp` | `EnsureDefaultWithPolicy -> migrateGlobalSkills -> strictAuditMember -> audit.Gate`; v2 binary data passes the opaque guard, but URL-bearing `scripts/fetch.sh` produces the exact network finding below and publishes neither profile source nor lock. |

The brief supplies no attack-surface table: this is a brief gap, not an implicit waiver. The map above covers its four explicit AC rows; broader attack-surface completeness is not claimed.

## Bounds and out-of-contract rows

- The existing migration accepts tag/revision declarations and deliberately skips live branch declarations. The fixture first writes and parses a legacy `branch: main` declaration, then pins that branch's exact commit as `revision` before driving the migration. Direct migration of an unresolved branch is outside this leaf's scope: BF1 asks for the two audit-version corrections, matching `global add --revision`, and does not change source selection.
- Tests drive the public Go migration entry, not a spawned CLI binary. They use temporary homes, temporary repositories cloned to local bare repositories, and fake `example.com` source identities rewritten to local file remotes. They never contact the fixture URL or a live manager home.
- Capability policy remediation is out of contract under BF1 point 5: record the confirmed finding for a separate leaf, do not fix it here.
- No previously committed reviewer tests were renamed, edited, or removed. The package suite retains all existing envprofile tests; the v1 NUL guard regressions and the new tests are green. Local conformance-root and platform skips are explicitly declared below; hosted CI is the arbiter for those rows. New tests remain uncommitted for the managed handoff snapshot, as explicitly required.
- This is an initial producer candidate, not review rework. No lasting diff lies outside `internal/envprofile`. Mutations temporarily touched `internal/opaquescan` and `internal/audit` solely to prove their gates; both were restored byte-for-byte from pre-mutation copies.
- No new source-text inspection gate was introduced. For the existing URL-source detector, M4 preserves the URL regex, the `audit.capability.network-undeclared` finding token, and the URL-bearing fixture while changing admission behavior; the behavioral migration test fails. The narrowing mutants execute the behavioral migration tests, not a static checker.

## Mutation evidence

Every mutant command was run directly with `-count=1`; expected-red results are failures with exit code **1**, not passing gates. No mutant survived.

| Mutant | What it narrows or changes | Named failing test | Result / survival bound |
| --- | --- | --- | --- |
| M1 — revert direct migration detector to `hashing.VersionV1` | Wrongly rejects the NUL-bearing migrated class under v2 while retaining the audit | `TestMigratedGlobalSkillBinaryWriteVersion/assets/deep/gzip-and-binary/v2` | Exit 1: `profile_source_invalid`, blocking `opaque-file`; killed. |
| M2 — revert migrated Resolved.HashVersion to `hashing.VersionV1` | Wrongly rejects that same v2 class through the second strict audit | `TestMigratedGlobalSkillBinaryWriteVersion/assets/deep/gzip-and-binary/v2` | Exit 1: `audit.opaque.nul-byte`; killed independently of M1. |
| M3 — skip NUL paths only under `.task-board/attachments/` in `opaquescan.walk` | Retains v1 NUL scanning everywhere else while admitting that directory's opaque files | `TestMigratedGlobalSkillBinaryWriteVersion/.task-board/attachments/gzip-and-binary/v1` (also gzip-only and binary-only) | Exit 1: migration unexpectedly returns nil; killed narrowing mutant, gate remains present. |
| M3b — skip only the exact `.task-board/attachments/attachment.gz` path in `opaquescan.walk` | Retains scanning for every other file while admitting exactly this opaque file | `TestMigratedGlobalSkillBinaryWriteVersion/.task-board/attachments/gzip-only/v1` | Exit 1: migration unexpectedly returns nil; killed. The gzip-plus-binary case still refuses, demonstrating why the isolated-file row is necessary. |
| M4 — allow only `downloads.example.com` in the network detector | Retains undeclared-host refusal for every other host while admitting this undeclared host | `TestMigratedGlobalSkillNetworkUndeclaredFollowUp` | Exit 1: migration unexpectedly returns nil; killed narrowing mutant, gate remains present. |

Commands for M1–M3b: `go test ./internal/envprofile/... -run '^TestMigratedGlobalSkillBinaryWriteVersion$' -count=1`.
Command for M4: `go test ./internal/envprofile/... -run '^TestMigratedGlobalSkillNetworkUndeclaredFollowUp$' -count=1`.
Raw focused and mutation logs, their digests and observed exit codes are attached as `BUG-261014-351fcr_test-evidence.txt`.

Logs: `.temp/BUG-261014-351fcr/mutant-direct-v1-01.log`, `mutant-resolved-v1-01.log`, `mutant-board-exemption-01.log`, `mutant-board-file-01.log`, and `mutant-network-host-01.log` in the same directory.

## Confirmed follow-up / logbook entry

`TestMigratedGlobalSkillNetworkUndeclaredFollowUp` is green because it explicitly expects the current refusal. Its production operation fails with:

```text
profile_source_invalid: migrated global skill "binary-skill" audit blocked: audit blocked: binary-skill: high audit.capability.network-undeclared - contacts downloads.example.com, which no capabilities.network entry covers (file: scripts/fetch.sh)
```

`strictAuditMember` constructs an audit subject with `capabilities.ImplicitNone()`. Under v2, binary content no longer hides this independent strict capability finding. The orchestrator should file a separate leaf to resolve the migrated-skill capability contract; no capability declarations, detector exemptions, or policy changes are introduced here. This outcome is also recorded in task notes; no control-root logbook was edited.

## Validation (own execution; no inherited results)

| Command | Real exit code | Evidence |
| --- | ---: | --- |
| Initial focused regression command before the test lookup correction | 1 | `focused-regression-01.log`: compile failure (`Lock.Member` undefined); corrected to existing `Lock.Find`; superseded by subsequent successful execution. |
| Focused regression command below | 0 | `focused-regression-02.log`: all 12 version/path/file cases and the follow-up test green. Full package run below covers the later lookup simplification. |
| `go build ./...` | 0 | `go-build-01.log` |
| `go vet ./...` | 0 | `go-vet-01.log` |
| `gofmt -l internal/envprofile/envprofile.go internal/envprofile/global_migration_hash_version_test.go` and `test ! -s` on its captured output | 0 | `gofmt-01.log` empty |
| `git diff --check` | 0 | Self-review; two-file scoped candidate. |
| `go test ./internal/envprofile/... -count=1 -v` | 0 | `envprofile-suite-01.log`: package PASS, 214.281s; conformance-root and platform skips are declared below. |
| `sh scripts/remote-gate.sh` | 0 | `remote-gate-01.log`; [hosted run 38517800551](https://github.com/relux-works/curator/actions/runs/38517800551), snapshot `f34a0db8be5b13282f7bb4742b5d8a21ead874cf` |

Focused regression command:

```sh
go test ./internal/envprofile/... -run '^TestMigratedGlobalSkill(BinaryWriteVersion|NetworkUndeclaredFollowUp)$' -count=1 -v
```

Local platform/toolchain: Go 1.25.5, darwin/arm64. The project is a cross-platform Go CLI, not an iOS app. R223 was followed: local execution was build, vet and the focused envprofile package; the full suite, race/property gates and lint belong to the hosted gate. The hosted gate is green on the exact snapshot; its full configured test/property suites, race jobs and lint passed. No prior green full-suite evidence was reused or substituted; this run executed the gate itself. Local-only conformance/platform skips remain declared, with hosted platform coverage supplied by this gate.

## Local package skips (unrun locally; hosted coverage required)

- `TestEnvironmentsCodexSeedVectors` — CURATOR_CONFORMANCE_ROOT is not set.
- `TestEnvironmentsEnvPassthroughVectors` — CURATOR_CONFORMANCE_ROOT is not set.
- `TestDotfileManagerVectors` — CURATOR_CONFORMANCE_ROOT is not set.
- `TestMusePublishedLinkStates` — CURATOR_CONFORMANCE_ROOT is not set.
- `TestMusePublishedRegistryLayout` — CURATOR_CONFORMANCE_ROOT is not set.
- `TestPathKindMCPVectorsDriveInstallEntry` — CURATOR_CONFORMANCE_ROOT is not set.
- `TestPathKindBoundaryVectorsDriveProductionEntries` — CURATOR_CONFORMANCE_ROOT is not set.
- `TestPathKindDryRunVectorsDriveReadOnlyResolve` — CURATOR_CONFORMANCE_ROOT is not set.
- `TestFoldedPathsAreSourceInvalid` — per-directory case sensitivity folds entries on this host; the collision gate runs on case-sensitive lanes.
- `TestSystemDeltaVectors` — CURATOR_CONFORMANCE_ROOT is not set.
- `TestEnvironmentsReadFailureVectors` — CURATOR_CONFORMANCE_ROOT is not set.
- `TestUpdateConfirmationPostureVectorAtStatus` — CURATOR_CONFORMANCE_ROOT is not set.
- `TestSignerPostureVectorsAtStatus` — CURATOR_CONFORMANCE_ROOT is not set.
- `TestStoreBoundaryResolveVectorsDriveResolve` — CURATOR_CONFORMANCE_ROOT is not set.
- `TestStoreBoundaryDryRunVectorsDriveResolveRepairPlan` — CURATOR_CONFORMANCE_ROOT is not set.
- `TestStoreBoundaryRepairVectorsDriveResolveRepair` — CURATOR_CONFORMANCE_ROOT is not set.
- `TestStoreBoundaryStatusVectorsDriveStatus` — CURATOR_CONFORMANCE_ROOT is not set.
- `TestEnvironmentWriteNofollowVectors` — CURATOR_CONFORMANCE_ROOT is not set.

## Hosted terminal result

[Run 38517800551](https://github.com/relux-works/curator/actions/runs/38517800551): **success**. `sh scripts/remote-gate.sh` returned **0** on its first attempt. Head: `f34a0db8be5b13282f7bb4742b5d8a21ead874cf`. Raw output is attached as `BUG-261014-351fcr_remote-gate.log` (SHA-256 `348fa0d72143f0afc4e7673a0afe7d3e52e1693aa2f66f1573c6d73bb3badd6e`).

| Hosted job | Outcome |
| --- | --- |
| Naming gate | success |
| Gate self-test (ubuntu-latest) | success |
| Race (macos-latest) | success |
| Gate self-test (windows-latest) | success |
| Test (macos-latest) | success |
| Go 1.26.0 driver (ubuntu-latest) | success |
| Race (ubuntu-latest) | success |
| Interop conformance gate | success |
| Go 1.27.1 driver (ubuntu-latest) | success |
| Go 1.26.0 driver (windows-latest) | success |
| Lint | success |
| Test (ubuntu-latest) | success |
| Go 1.27.1 driver (macos-latest) | success |
| Go 1.25.5 driver (ubuntu-latest) | success |
| Go 1.26.0 driver (macos-latest) | success |
| Go 1.25.5 driver (windows-latest) | success |
| Go 1.25.5 driver (macos-latest) | success |
| Gate self-test (macos-latest) | success |
| Go 1.27.1 driver (windows-latest) | success |
| Test (windows-latest) | success |
| Candidate suite (${{ matrix.os }}) | skipped |
| Test (rose-air) | skipped |

The candidate-dispatch suite and the self-hosted lane were skipped by the push workflow; they are unrun, not successes. The configured push landing gate passed all 20 applicable jobs, including Windows/macOS/Ubuntu tests, both race jobs, lint, driver checks, and gate properties. No remote check or status was synthesized.
