# SNAP-ID revision 2 developer handoff

Task: TASK-261016-1v3gbc — implement-rc7-snapshot-identity-verifier.
Candidate: `ad0947421e20acb08c1d980f04d04cb3c82847c1` (detached author-signed CI snapshot, signature G).
Story checkpoint: `0d00a109ad1da77a5506747c89f24af9442eb341`; all changes remain UNCOMMITTED and the Story index is unchanged.
Current normative curator-spec main read at `15d970111277b1f68b3f7a4ddd63bc339b2c0a3f`: managed-home-snapshots sections 2, 4, 6 and 7, and CIP-0014 implementation/test plan. No leaf/normative conflict found. Protocol bytes equal the previously pinned document. Corpus and embedded schemas remain unchanged.

## Rework and review response

Revision 1 REJECT finding `unicode-casefold-renormalization` is addressed in this leaf. Collision keys now implement canonical caseless matching with NFD before and after full case folding. Folding can produce decomposed strings, so normalization only before folding was insufficient.

`TestReviewerGreekCollision/{iota,upsilon}/{directory,file}` now resides in the candidate under its original name. Each singleton tree verifies through public `Verify`; each combined tree must refuse with `environment_materialize_snapshot_mismatch/path-collision` and preserve the manifest bytes. On a private case-sensitive filesystem, all four cases reproduced the old successful-admission bypass (exit 1), then passed after the fix (exit 0). M12 restores the old normalization-before-fold-only algorithm, retaining collision checks. It is killed by that same named production-entry regression (child exit 1).

All reviewer test functions are retained under their original names in `reviewer_unix_test.go`: `TestReviewerNullChannel`, `TestReviewerManifestOnly`, `TestReviewerPortableLeftDrift`, `TestReviewerRecapture`, `TestReviewerFreeHunt`, `TestReviewerUnicodePositive`, `TestReviewerGreekCollision`. Only formatting and a Unix build constraint were added to the downloaded reviewer tests; no existing tests were removed or renamed.

Rework paths relative to revision 1: `internal/envsnapshot/metadata.go`, `internal/envsnapshot/controls_test.go`, `internal/envsnapshot/reviewer_unix_test.go`, `internal/envsnapshot/README.md`. No path outside the named leaf module changes in this rework. The whole candidate has 116 files under the original two declared directories; every live byte matches the CI snapshot. Fresh upstream main is 5 commits ahead of the Story checkpoint, with no overlap in either leaf directory. Branch/checkpoint changes remain the orchestrator's responsibility.

## Acceptance evidence and coverage map

**4 of 4 AC rows driven; 4 of 4 mandatory paired-test rows driven at the production library boundary.** No CLI, capture/publication, hosted-launch, archive or runtime qualification is claimed. The helper is explicitly reusable by those consumers; it has no current CLI caller. Test definitions remain uncommitted as required by the managed Change Request contract and are included in the candidate snapshot.

| AC row | Production call site | Named driving tests |
|---|---|---|
| Closed codecs, CCJ-1 + one LF and distinct hash domains | `EncodeManifest`, `DecodeManifest`, `EncodeDescriptor`, `DecodeDescriptor`, `FragmentDigest`, `Digest`, `Verify` | `TestPinnedSchemaCases`, `TestMetadataFramingAndHashDomains`, `TestLaunchHashIncludesLF`, `TestTypedEncodingRejectsInvalidUTF8`, `TestPinnedTrees` |
| Complete actual inventory, modes, collisions, references and closed bindings | `Verify` → `validateMetadata`, `verifyEntries`, `verifyFile` | `TestPinnedTrees`, `TestActualTreeRefusals`, `TestMetadataNeighbors`, `TestReviewerGreekCollision`, `TestMuseDerivedDirectoriesAndRootFact`, `TestPortableComponentSuffixes`, `TestPiAlternativeChannels`, `TestNoFollowSnapshotRoot`, `TestNoFollowBeforeParentSegment` |
| Exact versus portable identity; never recapture baseline | `Verify`, `PortableEqual`, `Verified.ManifestSHA256`, `Verified.ManifestBytes` | `TestPortableIdentity`, `TestBaselineNeverRecaptured`, `TestReviewerPortableLeftDrift`, `TestReviewerRecapture` |
| Fixed separate corpus; no released-fragment/archive extension or manifest-only attestation | codecs and `Verify`, `PortableEqual` | `TestCorpusPins` (95 published pins), `TestPinnedSchemaCases` (58 structural cases), `TestPinnedTrees` (19 semantic/tree vectors), `TestReviewerManifestOnly`, `TestPinnedSchemaCheckoutBytes` |

Brief gap: no catalog-keyed attack-surface table was supplied. This is not treated as a waiver; the map covers every supplied mandatory paired row in order using the coverage-map template.

| Surface row (mandatory pair) | Attacking tests | Killed narrowing mutants | Out-of-contract inputs (AC clause) |
|---|---|---|---|
| Canonical tree / closed metadata and semantic bindings | `TestPinnedSchemaCases`, `TestMetadataFramingAndHashDomains`, `TestPinnedMetadataSemantics`, `TestPinnedTrees`, `TestMetadataNeighbors`, `TestLaunchHashIncludesLF`, `TestTypedEncodingRejectsInvalidUTF8`, `TestPinnedSchemaCheckoutBytes`, `TestReviewerGreekCollision`, `TestReviewerNullChannel` | M1, M2, M3, M4, M7, M10, M11, M12 → named killers below | None claimed |
| Independent copy / actual inventory, modes, aliases and failed reads | `TestPinnedTrees`, `TestActualTreeRefusals`, `TestNoFollowSnapshotRoot`, `TestNoFollowBeforeParentSegment`, `TestMuseDerivedDirectoriesAndRootFact`, `TestReviewerManifestOnly`, `TestReviewerFreeHunt` | M5-mode, M5-extra, M9 → named killers below | None claimed |
| Portable equality / exact provenance and all remaining bindings | `TestPortableIdentity`, `TestReviewerPortableLeftDrift`; public `Verify`, `PortableEqual`, `Verified.ManifestSHA256` | M8 → `TestPortableIdentity/version` | None claimed |
| Original baseline / refusal after attempted recapture | `TestBaselineNeverRecaptured`, `TestReviewerRecapture`; `Verify` with original digest | M6 → `TestBaselineNeverRecaptured` | None claimed |

No supplied row is declared out of contract. Explicit implementation bounds remain: Unix actual-tree backend only; Windows fails closed, and this is a remaining backend limitation rather than an assertion that Windows is outside the normative contract. Codecs and pinned checkout tests remain cross-platform. Source capture/admission, destination ownership and ACLs, secret detection, consistent capture, launch/runtime protection, authorization and overlay baseline retention belong to consuming leaves. `Verify` observes a tree; it cannot grant plan authority or a continuing lease against hostile writers. Decode-only operations never attest a directory; no manifest-only verification shortcut or baseline rewrite API exists.

## Executed narrowing controls

**13 of 13 controls killed; 0 survivors.** Each child directly runs `go test -count=1 -json -run <named test> <isolated package copy>` and exits 1. The harness requires a failure event for the named behavioral test, not a compile/setup failure. The original package is never mutated. M5/M6 execute the leaf's planned checker controls at the helper seam, without claiming future materialize CLI qualification.

| Mutant | What it narrows the gate to | Named test that fails | Survival bound |
|---|---|---|---|
| M1-one-extra-LF | Admit precisely a second metadata LF | `TestMetadataFramingAndHashDomains` | Killed; no survivor |
| M2-one-unknown-member | Admit `_unknown: true` | `TestMetadataFramingAndHashDomains` | Killed; no survivor |
| M3-launch-hash-without-LF | Also admit the descriptor hash excluding LF | `TestLaunchHashIncludesLF` | Killed; no survivor |
| M4-component-NFC-alias | Permit an NFC directory-prefix alias containing café | `TestMetadataNeighbors/NFD-components` | Killed; no survivor |
| M5-mode | Permit execute-bit loss on a 0700 asset while retaining other guards | `TestActualTreeRefusals/execute-mode` | Killed; no survivor |
| M5-extra | Permit just unlisted `home/extra` | `TestActualTreeRefusals/extra-file` | Killed; no survivor |
| M6-saved-manifest | Skip actual reads only when a prior expected digest is supplied | `TestBaselineNeverRecaptured` | Killed; no survivor |
| M7-wrong-fixed-layer | Also admit listed AGENTS.md as the named MCP layer | `TestMetadataNeighbors/wrong-fixed-layer` | Killed; no survivor |
| M8-portable-version | Exclude producer version as well as source fragment | `TestPortableIdentity/version` | Killed; no survivor |
| M9-two-hardlinks | Permit exactly two hardlinks | `TestActualTreeRefusals/hardlink-outside` | Killed; no survivor |
| M10-UTF8-path-only | Permit malformed UTF-8 only for .txt strings | `TestTypedEncodingRejectsInvalidUTF8` | Killed; no survivor |
| M11-schema-README-EOL | Keep original protection token and permit schema Markdown conversion | `TestPinnedSchemaCheckoutBytes/protected` | Killed; no survivor |
| M12-casefold-without-renormalization | Retain all collision checks but omit normalization after fold | `TestReviewerGreekCollision` (four variants) | Killed; no survivor |

M11 preserves the searched source token and drives actual Git checkout behavior. There is no source-text-only production gate. M12 was executed with case-sensitive TMPDIR so the four formerly admitted physical trees coexist; this is stronger evidence than a later entry-set refusal on a case-folding filesystem.

## Commands and real exits

Local environment: Go 1.25.5, darwin/arm64. Every listed test/build/lint/gate ran directly without tee or a pipeline; logs are preserved in the verification outcome. CASE_TMP below denotes the private task-local case-sensitive HFS+ mount. All child test invocations run synchronously.

| Command | Exit | Evidence |
|---|---:|---|
| `TMPDIR="$CASE_TMP" go test -count=1 -run '^TestReviewerGreekCollision$' -v ./internal/envsnapshot` BEFORE fix | 1 | `greek-before-01.log`: expected red; all four cases admit colliding trees, singleton controls pass |
| Same command AFTER fix | 0 | `greek-after-01.log`: four intended refusals, unchanged manifests |
| `TMPDIR="$CASE_TMP" go test -count=1 -race -coverprofile=.temp/SNAPID-rework/coverage-01.out ./internal/envsnapshot` | 1 | `focused-race-01.log`: HFS+ rewrites composed café filename to decomposed spelling; exact inventory refusal in `TestReviewerUnicodePositive/home/café/asset`. Not green or reused evidence |
| `go test -count=1 -race -coverprofile=.temp/SNAPID-rework/coverage-02.out ./internal/envsnapshot` | 0 | `focused-race-02.log`: native filesystem, all tests including reviewer regressions; 87.8% statements |
| `TMPDIR="$CASE_TMP" go test -tags snapshotcontrols -count=1 -run '^TestNarrowingControls$' -v ./internal/envsnapshot` | 0 | `controls-01.log`: 13 named controls killed, each child expected red exit 1 |
| `golangci-lint run --build-tags snapshotcontrols ./internal/envsnapshot/...` | 0 | `lint-01.log`: 0 issues |
| `go vet ./internal/envsnapshot` | 0 | `vet-01.log` |
| `go build -o .temp/SNAPID-rework/curator-build-01 ./cmd/curator` | 0 | `build-01.log` |
| `GOOS=windows GOARCH=amd64 go test -c -o .temp/SNAPID-rework/envsnapshot-windows.test.exe ./internal/envsnapshot` | 0 | `windows-compile-01.log`: compile only |
| `sh scripts/remote-gate.sh` | 0 | https://github.com/relux-works/curator/actions/runs/38855636613 on exact candidate |

The unchanged gate's detached `commit-tree` operation runs through the previously reviewed task-local Git adapter to preserve configured author identity, explicitly sign and verify the snapshot; it changes neither source tree nor CI inputs. No unsigned snapshot is published and no commit is created on the Story branch. The directly executed gate returned exit 0; provider conclusion is success on this exact signed candidate. The two conditional skipped jobs are not claimed as executed evidence.

No older green is substituted for the changed candidate. Previous revision handoff and review remain historical evidence. The case-sensitive image was detached after the regression and controls; only this run's scratch image was touched. The general suite is run on the native filesystem, because HFS+ rewrites a valid composed test filename. No product inventory comparison was weakened to accommodate that fixture behavior.

## Task-scoped logbook

- Read the previous REJECT and retained every supplied reviewer test name in candidate source.
- Reproduced all four Greek successful-admission bypasses before changing production code.
- Changed canonical caseless keys within the leaf, without weakening exact path inventory or adding special-case spellings.
- Added M12 restoring the former algorithm; named behavioral regression kills it on a case-sensitive tree.
- Recorded the HFS+ filename-spelling fixture failure honestly; the native race suite is green, and the targeted Greek regression/control evidence retains the case-sensitive environment.
- Verified all 116 candidate file bytes equal the signed CI snapshot; only the declared leaf directories change, and only four module paths differ from revision 1.
- Fresh upstream code has no overlap; no worktree branch, checkpoint or index changes were made. No control-root file was edited directly.
- Early task query syntax and one unquoted API-path shell error were corrected; failed reads were not treated as absence or evidence.


## Current hosted results and handoff checks

Workflow [38855636613](https://github.com/relux-works/curator/actions/runs/38855636613) reports success on exact candidate `ad0947421e20acb08c1d980f04d04cb3c82847c1`. Directly executed `sh scripts/remote-gate.sh` returned exit **0**. Twenty jobs are successful, including both Linux/macOS Race jobs, all three Test jobs, Lint, protocol conformance, naming, gate controls and all toolchain drivers. Two conditional jobs are skipped (self-hosted Test job and Candidate suite); neither is represented as run or passing. No local mirror or substituted hosted status was used.

| Hosted Test job | Distinct passing SNAP-ID test/subtest names | Bound |
|---|---:|---|
| Linux | 178 names | Complete Unix helper suite, including all four Greek regression variants and all retained reviewer tests |
| macOS | 178 names | Complete Unix helper suite, including all four Greek regression variants and all retained reviewer tests |
| Windows | 103 names | Codecs, pinned schema checkout, corpus and explicit unsupported-platform refusal; no successful actual-tree backend claim |

All three package streams contain a package-level pass and zero package fail/skip events. `TASK-261016-1v3gbc_hosted-package-rev2.json` preserves exact named events and source artifact references; duplicate served streams are not counted twice. Both Linux/macOS Race jobs report success. All old test functions remain present.

Source-identity readback verifies 116 live files match their immutable CI blobs and hashes, signature G, unchanged Story HEAD/index and no unexpected worktree paths. Rework remains confined to four module paths. Normative protocol byte comparison also returned exit 0. No product source changed after validation; later edits were task-local evidence and board mutations only.

The handoff preconditions are evidenced: every supplied paired row is mapped to a killed mutant; landing properties are green; reviewer tests are retained and green; no supplied row is silently excluded; rework paths are bounded. Platform and consuming-leaf limits remain as declared above. The candidate is ready for review.
