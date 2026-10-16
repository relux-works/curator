# SNAP-ID developer review handoff

Task: TASK-261016-1v3gbc — implement-rc7-snapshot-identity-verifier.
Candidate: `5fd9bed1663a16a4914cc341949e69eb58a2b7b0` (detached, author-signed CI snapshot).
Story checkpoint remains `0d00a109ad1da77a5506747c89f24af9442eb341`; candidate changes remain UNCOMMITTED.
Normative curator-spec main read at `d2a942cf17e4eb017c45475442da1456d8dfbe6c`.
No leaf/normative conflict was found. The older brief baseline was not substituted for current normative main.

## Change and integration interface

Only `internal/envsnapshot/` and `internal/crossconformance/testdata/managed-home-snapshot-v1/` change (115 added files). No existing code, reviewer tests, released fragment, archive format, CLI spelling, root documentation, gate script or control-root file was edited.

Public helper entry points are `EncodeManifest`, `DecodeManifest`, `EncodeDescriptor`, `DecodeDescriptor`, `FragmentDigest`, `Verify(root, originalManifestSHA256)`, and `PortableEqual(left, right)`.
`Verified` returns copied immutable observations and LF-inclusive digests. `Verify` is the intended destination-idempotence and future hosted-verification seam. Consumers must supply their original admitted baseline and retain overlay identity separately. There is no baseline persistence or recapture operation here.

Codecs validate the exact embedded published schemas offline, then the verifier checks cross-file identity, channel registry semantics, path/component collisions and complete actual file/directory inventory. Actual files are independently opened through retained no-follow Unix directory handles and read/hash checked. Modes, links, special files, missing entries, extras and read errors refuse with `environment_materialize_snapshot_mismatch`, without raw path/content diagnostics. Core portable-component rules are checked in addition to the structural schema.

## Coverage and bounds

**4 of 4 leaf mandatory paired-test rows driven at the production library boundary.** Production CLI materialize, capture, publication, source admission, hosted attach/resume, runtime protection and archive transport execution: **0 claimed**. These are consuming leaves, not helper verification claims. A source provenance hash cannot by itself attest the source capture; the capture owner supplies it from verified inputs using the correct hash domain.

The brief provides no catalog-keyed surface table. This is an explicit brief gap, not a silent waiver. The map below covers every row of its actual mandatory paired-test table, in order, using the coverage-map template columns.

| Surface row (mandatory pair) | Attacking tests / production call sites | Killed narrowing mutants | Out-of-contract inputs (AC clause) |
|---|---|---|---|
| Canonical tree / closed metadata and semantic bindings | `TestPinnedSchemaCases`, `TestMetadataFramingAndHashDomains`, `TestTypedEncodingRejectsInvalidUTF8`, `TestPinnedMetadataSemantics`, `TestPinnedTrees`, `TestMetadataNeighbors`, `TestLaunchHashIncludesLF`, `TestPortableComponentSuffixes`, `TestPiAlternativeChannels`, `TestPinnedSchemaCheckoutBytes`; public codecs, `FragmentDigest`, `Verify` | M1, M2, M3, M4, M7, M10, M11 → named killers below | None claimed |
| Independent copy / actual inventory, mode, aliases and read refusal | `TestPinnedTrees`, `TestActualTreeRefusals`, `TestNoFollowSnapshotRoot`, `TestNoFollowBeforeParentSegment`, `TestMuseDerivedDirectoriesAndRootFact`; `Verify` | M5-mode, M5-extra, M9 → named killers below | None claimed |
| Portable equality / exact provenance and all remaining bindings | `TestPortableIdentity`; `Verify`, `PortableEqual`, `Verified.ManifestSHA256` | M8 → `TestPortableIdentity/version` | None claimed |
| Persisted baseline / refusal after attempted recapture | `TestBaselineNeverRecaptured`; `Verify` with original exact digest, copied `Verified.ManifestBytes` | M6 → `TestBaselineNeverRecaptured` | None claimed |

Each refusal is asserted against the stable mismatch code, with specific bounded reasons where applicable, beside a positive control. Actual-tree negatives preserve the manifest baseline; the baseline-regression test also checks the newly observed target metadata is not rewritten by repeated rejection. Codecs have no external write side effects. Every test has a brief claim comment. The corpus includes 58 indexed structural cases and 19 semantic/tree vectors; exact counts prevent accidental case loss. `TestCorpusPins` checks 95 referenced corpus/schema files. The remaining protocol-document reference was read as normative source; it is not claimed as an executable fixture.

**Platform limit:** Unix actual-tree verification is implemented. Other platforms, including Windows, refuse rather than infer private ACL or execute semantics from synthetic Go permission bits. This is a remaining platform backend limitation, not an assertion that Windows is out of the normative contract. Cross-platform codecs compile; hosted workflow results below describe only the jobs actually run. No Windows tree-verification success or runtime/ACL qualification is claimed.

The verifier is a read-only observation, not a continuing lease against hostile writers. Capture locking/revalidation, destination ownership/ACL admission, live-home closure selection, secret admission, relocation, tool execution and overlay persistence are not implemented or attested by this helper. No `manifest-only` directory-verification shortcut exists. Required source/discovery closure selection belongs to the capture owner; this helper verifies the entire supplied inventory and its declared channel references.

## Narrowing controls

Every child command is `go test -count=1 -json -run <named test> <isolated package copy>` and fails with exit **1**. The control harness exits **0** only after the named test emits a failure event; a compilation/setup failure cannot count as a kill. Source originals remain untouched. These are behavioral controls; no production gate attests behavior by searching source text. The mutation placement search preserves the target token while changing behavior and executes the behavioral suite.

| Mutant | What it narrows the gate to | Named test that fails | Survival bound |
|---|---|---|---|
| M1-one-extra-LF | Admit precisely a second metadata presentation LF | `TestMetadataFramingAndHashDomains` | Killed; no survivor |
| M2-one-unknown-member | Admit the otherwise forbidden `_unknown: true` member | `TestMetadataFramingAndHashDomains` | Killed; no survivor |
| M3-launch-hash-without-LF | Also admit a launch digest excluding presentation LF | `TestLaunchHashIncludesLF` | Killed; no survivor |
| M4-component-NFC-alias | Permit an NFC directory-prefix alias containing café | `TestMetadataNeighbors/NFD-components` | Killed; no survivor |
| M5-mode | Verify bytes while permitting execute-bit loss on 0700 assets | `TestActualTreeRefusals/execute-mode` | Killed; no survivor |
| M5-extra | Permit just the unlisted `home/extra` member | `TestActualTreeRefusals/extra-file` | Killed; no survivor |
| M6-saved-manifest | Trust metadata only when a prior expected digest is supplied | `TestBaselineNeverRecaptured` | Killed; no survivor |
| M7-wrong-fixed-layer | Permit a named MCP layer to target listed AGENTS.md instead | `TestMetadataNeighbors/wrong-fixed-layer` | Killed; no survivor |
| M8-portable-version | Also discard producer-version difference in portable comparison | `TestPortableIdentity/version` | Killed; no survivor |
| M9-two-hardlinks | Admit exactly two hard links instead of requiring independence | `TestActualTreeRefusals/hardlink-outside` | Killed; no survivor |
| M10-UTF8-path-only | Permit malformed UTF-8 only in strings ending in .txt | `TestTypedEncodingRejectsInvalidUTF8` | Killed; no survivor |
| M11-schema-README-EOL | Keep schema JSON byte-protected but permit text conversion of schema Markdown | `TestPinnedSchemaCheckoutBytes/protected` | Killed; no survivor |

The M4 control proves semantic collision rejection (the named refusal reason), including on filesystems that fold normalization and cannot represent the two physical aliases independently. It does not claim creation of that impossible physical tree on a folding filesystem.

**12 of 12 controls killed; 0 survivors.** M5/M6 are the planned CIP-0014 checker classes, now executed at this leaf's public verifier seam. This does not relabel them as future `MaterializeCLI_*` qualification.

## Verification

Current candidate commands and observed exit codes:

| Command | Exit | Evidence |
|---|---:|---|
| `go test -count=1 -race -coverprofile=.temp/TASK-261016-1v3gbc/coverage-current.out ./internal/envsnapshot` | 0 | `focused-12.log`; 86.9% statements |
| `go test -tags snapshotcontrols -count=1 -run '^TestNarrowingControls$' -v ./internal/envsnapshot` | 0 | `controls-06.log`; all 12 child test commands exit 1 as intended |
| `golangci-lint run --build-tags snapshotcontrols ./internal/envsnapshot/...` | 0 | `lint-06.log` |
| `go vet ./internal/envsnapshot` | 0 | `vet-04.log` |
| `go build -o .temp/TASK-261016-1v3gbc/curator-build-04 ./cmd/curator` | 0 | `build-04.log` |
| `GOOS=windows GOARCH=amd64 go test -c -o .temp/TASK-261016-1v3gbc/envsnapshot-windows-latest.test.exe ./internal/envsnapshot` | 0 | `windows-compile-04.log`; compilation only |
| `go test -count=1 -run '^TestPinnedSchemaCheckoutBytes$' -v ./internal/envsnapshot` | 0 | `checkout-regression-01.log`; real Git protected checkout and unprotected CRLF witness |
| `sh scripts/remote-gate.sh` | 0 | Workflow https://github.com/relux-works/curator/actions/runs/38841469667 on candidate 5fd9bed1663a16a4914cc341949e69eb58a2b7b0 |

The existing repository gate creates an unsigned detached commit with a synthetic author. This invocation uses a temporary Git command adapter only for that `commit-tree` call: it retains configured human identity, signs explicitly, verifies the object, then lets the unchanged gate push/run/cleanup normally. The Story branch and index remain untouched; the gate snapshot signature is verified as G. No unsigned candidate was published. The gate's source tree, CI commands and job configuration are unchanged by this execution adapter.

Prior evidence is not substituted for the current candidate: early package compile/test iterations failed during implementation, then were corrected. First repository-gate run 38834414874 on older candidate 27b492711660252e78f97c2f69d5dcb67c78ada5 was deliberately cancelled after the UTF-8 encoding issue was found; its local gate process exited 137. It is not passing evidence. Earlier local green runs are superseded by the exact commands above. Diagnostic shell calls that did not retain the child test status are not passing evidence.

## Task-scoped logbook

- Initial development transition refused because estimate was absent. Applied the leaf's size 3 estimate, then entered development.
- Fresh fetch confirmed Story checkpoint equals origin/main before source work. No branch switch, commit or rebase was performed.
- Read normative snapshot sections 2/4/6/7, CIP-0014 and referenced path/channel rules on current curator-spec main; pinned its separate corpus unchanged.
- Avoided a forced Windows ACL/execute-mode inference by preserving a structured refusal and explicitly recording that backend limit.
- Tests account for the platform test allocator's linked temp spelling without relaxing production no-follow checks; explicit ancestor-link and link-before-parent regressions remain.
- Removed an overconstraint during self-review: pi channels are alternatives and cannot require one inert prompt path to equal two fixed filenames.
- Refused typed invalid UTF-8 before JSON serialization so an encoder cannot silently normalize path identity. Added its named regression and narrowing control.
- Source/metadata verification does not imply source admission, baseline authority or launch/runtime protection. Those integrations remain consuming-leaf work.

The candidate is ready for review; all limitations above remain part of the handoff.

## CI registration regression and bounded gate control

Hosted run 38834659070 on prior candidate `37a4424677a7e8253641fa48e9962a276f4da3e0` had a real macOS platform-case gate failure: the default opt-in `TestNarrowingControls` registration skipped with an unrecognised reason. That failure is superseded, never counted as green. Remaining jobs were cancelled after diagnosis; the gate process returned exit **1**. The correction is confined to the control test build tag and package README: `snapshotcontrols` explicitly includes the runner; the default suite registers no unexecuted control test. Neither the shared platform gate nor its skip whitelist was weakened.

The production `.github/ci/platform-case-gate.sh` was rerun on a focused real `go test -json` stream and a task-local ledger for this package: exit **0**, zero skips. Planting the former named skip into that stream makes the same command fail with exit **1** and `FAIL skip with an unrecognised reason ... TestNarrowingControls`. This is a bounded check of this package's registration and skip detection, not a substitute for the repository-wide ledger or hosted gate.

Exact focused commands: `go test -count=1 -json ./internal/envsnapshot`; then `CI_PLATFORM_CASES=.temp/TASK-261016-1v3gbc/focused-platform-cases.tsv bash .github/ci/platform-case-gate.sh .temp/TASK-261016-1v3gbc/focused-current.json .temp/TASK-261016-1v3gbc/platform-positive-04` (exit 0), and the same gate with `focused-planted-skip-04.json` / `platform-negative-04` (expected red, exit 1). Evidence: `platform-positive-04.log`, `platform-negative-04.log`.

## Windows checkout regression

Hosted run 38836557736 on prior candidate `e095dcbccc3122f65df7f33ea5c479802d5e6c87` returned **failure**; `sh scripts/remote-gate.sh` returned exit **1**. Nineteen jobs passed, two conditional jobs skipped, and `Test (windows-latest)` failed. Its Go JSON names `TestCorpusPins` in `internal/envsnapshot`: `pin mismatch: schemas/managed-home-snapshot-v1/README.md`. The platform-case gate itself returned 0; `go test` returned 1. This is a repository checkout configuration failure, not an external-CI failure. No mirror or waived check is claimed.

The correction is leaf-local `.gitattributes` with `schemas/** -text`. Schema and corpus bytes are unchanged. Existing root attributes already protect testdata. `TestPinnedSchemaCheckoutBytes` now drives actual `git hash-object`, `git update-index` and `git checkout-index` under `core.autocrlf=true`, without commits. Its positive control retains the three independently pinned snapshot schema assets; the unprotected witness must convert them to CRLF. M11 retains the original `schemas/** -text` token, adds a later override permitting conversion only for schema Markdown, and fails the named protected test (exit 1), while other schema assets remain protected. Its harness executes real Git checkout behavior; a token-presence check would miss this weakening.

The older Linux/macOS artifacts contain 157 distinct passing envsnapshot test/subtest names each, plus package-level pass (aggregate/served copies of the same stream are not counted twice), and no package fail/skip. They remain historical evidence only: the current candidate adds the checkout regression and configuration, the current hosted gate results below are the current evidence. Source verification logic is unchanged by the checkout repair.

## Current hosted gate and unchanged candidate

Workflow [38841469667](https://github.com/relux-works/curator/actions/runs/38841469667) completed with provider conclusion **success** at exact candidate `5fd9bed1663a16a4914cc341949e69eb58a2b7b0`. The directly executed `sh scripts/remote-gate.sh` returned exit **0**. Twenty jobs report success. Two report skipped: the self-hosted macOS Test job and Candidate suite. Neither is represented as an executed or passing test. No external-CI mirror or status substitution was used.

Package observations below come from each job's `test/go-test.json` artifact stream; the duplicate served stream is not counted again. All three contain an explicit package-level pass and zero package fail/skip events. The exact named test events are persisted in the verification outcome.

| Current hosted job | Distinct passing test/subtest names | Bound |
|---|---:|---|
| Test (ubuntu-latest) | 160 names | Codecs, real checkout and Unix actual-tree tests |
| Test (macos-latest) | 160 names | Codecs, real checkout and Unix actual-tree tests |
| Test (windows-latest) | 103 names | Codecs, real checkout, pinned bytes, and explicit unsupported-backend refusal; no actual-tree success claim |

`TestCorpusPins` and `TestPinnedSchemaCheckoutBytes` are green on Windows at this exact head. `TestUnqualifiedPlatformRefuses` is also green; the Windows actual-tree backend remains the previously declared implementation limit. Linux and macOS Race jobs report success. All previously committed repository tests remain unchanged.

Final source verification returned exit **0**: all 115 live files match recorded source hashes and the immutable CI Git blobs; candidate signature is G; Story checkpoint and index remain unchanged. Only the declared two leaf directories change. Candidate source remains UNCOMMITTED for the managed Change Request snapshot.
