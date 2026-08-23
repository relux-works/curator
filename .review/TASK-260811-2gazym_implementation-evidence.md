# TASK-260811-2gazym implementation and reviewer-rework evidence

Status: developer evidence prepared for a new independent review

Run: `RUN-260811-c1163f`

Active goal at the last directive checkpoint: `GOAL-260811-caa9a9`
revision 1, resolved scope `TASK-260811-2gazym`

Review policy: required

Date: 2026-08-11

## Delivered boundary

The adapter-independent `internal/artifactpolicy` package contains 27 Go files
and 14,555 lines including tests and the reusable conformance corpus. It
provides:

- closed artifact-class, trust-role, decision, node-kind, source-profile,
  source-grammar, resolved-use, and stable diagnostic vocabularies;
- immutable byte and directory admission through dependency, external
  toolchain, local-output, and unavailable verified-binary APIs;
- package-private manager seals for toolchain and local-output authority, so
  caller-populated evidence cannot authorize execution or publication;
- structural ELF, PE/COFF, Mach-O/fat Mach-O, JVM/DEX, WebAssembly, LLVM,
  Python bytecode, V8 cache/snapshot, framework, and XCFramework detection;
- recursive ZIP/ZIP64, POSIX/PAX/GNU tar, gzip, and GNU/BSD/COFF native `ar`
  traversal, including byte-detected nested containers;
- NFC/UTF-8 virtual paths, portable collision keys, link/special-node denial,
  early checked limits, complete bounded finding evidence, and deterministic
  duplicate handling; and
- canonical `artifact-manifest-v1` encoding/decoding that binds policy,
  detectors, limits, origin, causal role evidence, raw payload, every logical
  node, observations, accumulated traversal accounting, every finding,
  decision, and final digest.

`verified-binary-v1` remains unavailable and returns
`artifact_binary_admission_unavailable`. Kotlin remains outside this cycle.
Compiled dependency bytes remain fail-closed regardless of suffix, package
metadata, adapter, copied toolchain bytes, or claimed output location.

The canonical codec golden is
`sha256:66b11740d0ed814eaee0a3d141778b9fb21719366ea550c8385145a3556f5d8b`.
The reusable conformance corpus contains 182 public-API cases, each with an
exact pinned full-manifest digest. Its source SHA-256 is
`49f796cd44fa89428b4b572d390bb49765119b12fedb428f7ae87ee45e61dd24`.

## Closure of RUN-260811-068fd0 findings

The complete changes-requested verdict remains preserved on the board as
`TASK-260811-2gazym_review-verdict_RUN-260811-068fd0.md`.

| Finding | Implementation and adversarial proof |
| --- | --- |
| R1: rejected-node semantics forgeable | `DecodeManifest` independently derives class, variant, selected detector, decision, and rule from a closed detector-result/fact schema for admitted and rejected nodes. It validates complete entry evidence for ZIP, tar, gzip, and `ar`, including indexed ELF/PE/Mach-O structure. `TestCodecDerivesRejectedAndAdmittedNodeSemantics` self-consistently rehashes contradictory ELF, Mach-O, PE, JVM, Wasm, LLVM, text, ZIP-entry, detector-result, class, and rule mutations and requires rejection. |
| R2: truncated findings forgeable | `FindingsSummary.Evidence` retains every complete canonical finding while `Diagnostics` remains the cap-bounded display prefix. Decode requires `Total == len(Evidence)`, `Recorded == min(Total, max_recorded_findings)`, an exact recorded prefix, unique finding identities, a recomputed complete-set digest, semantic binding for hidden findings, and finding coverage for every rejecting node. `TestFindingsSummaryRequiresCapSaturationAndCompleteVerifiableEvidence` uses 1,001 findings with a compiled rejection hidden after the 1,000-record prefix and rejects premature truncation, invented counts/duplicates, omission, and semantic drift. |
| R3: native-archive metadata omitted | GNU `/`, `/SYM64/`, `//`, BSD `__.SYMDEF`, and COFF second-linker/import metadata are structurally validated, hashed, and manifested as canonical `native.library.static` metadata nodes. `TestNativeArchiveMetadataIsCanonicalAndRoleBound` proves dependency rejection and sealed toolchain/output decisions for GNU, BSD, and COFF-import shapes. A08 and C04 corpus branches include those forms. |
| R4: conflicting duplicates order-dependent | ZIP, tar, and `ar` walkers preflight the closed entry limit, retain and inspect every duplicate occurrence, and assign canonical occurrence identities/paths after sorting content and metadata identity. `TestDuplicateMembersAreInspectedIndependentlyOfPhysicalOrder` permutes source and ELF duplicates for all three formats and requires identical logical nodes, complete finding evidence, primary code, and no authorization. Six F02 corpus cases pin the distinct raw manifests while comparing identical logical projections. |
| R5: link/load use ignored for benign bytes | Resolved use evidence is part of the source detector contract. A `link_or_load` edge with otherwise benign text resolves to `opaque.unknown` and rejects with `artifact_type_ambiguous`; an `execute` edge still admits a declared interpreted script. `TestResolvedTextUsesAreFailClosedWithoutBreakingScripts` covers neutral and deny-suffix names plus the script case, and the same branches are reusable corpus inputs. |
| R6: gzip padding inflates ratio budget | Gzip traversal meters bytes actually consumed by the first member with a counting byte reader and adjusts the streaming ratio ceiling from those bytes. Trailing padding or a concatenated stream cannot credit the first stream's decompression budget. `TestGZIPExpansionUsesOnlyFirstStreamCompressedBytes` proves both high-ratio forms stop at `max_expansion_ratio`; F08 publishes both exact cases. |

The final codec hardening also validates every detector fact field rather than
only the class discriminator. Unknown, duplicate, malformed, out-of-range, or
structurally inconsistent facts make a rehashed manifest invalid.

## Required vector evidence

`conformance.Cases()` publishes 182 reusable inputs covering A01-A08,
C01-C12 including C01a-C01f, F01-F14, T01-T05, and current-capability V01.
Every case invokes a public role-specific admission API and pins class, node
decision, manifest decision, primary diagnostic, canonical path,
authorization absence/presence, round-trip bytes, and exact manifest digest.
Every negative asserts that neither adapter-execution nor cache-publication
authority is available. C12 compares the entire canonical leaf record across
Go, Rust, Node, SwiftPM, and Python-reference descriptors.

The option-1 F14 clarification remains intact: raw-order permutations have
different exact payload/full-manifest identities but identical canonical
logical evidence projections. Current full-manifest digests are:

- `F14/z-first`:
  `sha256:02076c7b888ea56ec893f773e6d2035fdd25eba08032bab8f8dfb47c2cb7ca49`
- `F14/a-first`:
  `sha256:064f09122ffafc3438057d5c0782df505b1713bd9c8a92e5bfb21c9228c0cec9`

Conflicting duplicate ZIP, tar, and `ar` permutations likewise retain distinct
raw/full identities while their node and complete-finding projections compare
equal.

## Current-source validation

Every green gate below ran as a direct standalone process after the last
behavior change. No gate was piped through `tee` or a status-obscuring command.

| Command | Exit | Evidence |
| --- | ---: | --- |
| `gofmt -l internal/artifactpolicy` | 0 | no files listed |
| `go test -short -count=1 ./internal/artifactpolicy/...` | 0 | artifactpolicy 9.032s |
| `go test -count=1 -run '^TestDirectoryEntryLimitStopsTheLiveWalker$' ./internal/artifactpolicy` | 0 | normative 100,001-entry live-tree rejection and canonical decode passed in 27.466s |
| `go test -short -count=10 ./internal/artifactpolicy/...` | 0 | ten repeated focused runs passed in 83.489s |
| `go test -race -short -count=1 -coverprofile=/tmp/TASK-260811-2gazym-cover.out ./internal/artifactpolicy/...` | 0 | race passed; artifactpolicy coverage 74.8%, 86.101s |
| `go test -short -count=1 -run 'TestReusableArtifactManifestV1ConformanceCorpus\|TestConformanceCorpusBuildersAreDeterministic' ./internal/artifactpolicy` | 0 | all 182 exact vectors and deterministic builders passed in 0.940s |
| `go vet ./internal/artifactpolicy/...` | 0 | no findings |
| `go build ./internal/artifactpolicy/...` | 0 | package compiled |
| `/Users/iv/go/1.25.5/bin/golangci-lint run ./internal/artifactpolicy/...` | 0 | `0 issues.` |
| `GOOS=linux GOARCH=amd64 go test -c -o /tmp/TASK-260811-2gazym-artifactpolicy-linux-amd64.test ./internal/artifactpolicy` | 0 | Linux/amd64 test binary compiled |
| `GOOS=windows GOARCH=amd64 go test -c -o /tmp/TASK-260811-2gazym-artifactpolicy-windows-amd64.test.exe ./internal/artifactpolicy` | 0 | Windows/amd64 test binary compiled |
| `go test -count=1 ./...` | 0 | every repository package passed; `cmd/curator` 362.269s, artifactpolicy 129.622s, closuregraph 15.806s, godriver 66.492s, install 114.216s, install/atomicity 114.775s, transaction 65.533s |
| `go vet ./...` | 0 | no findings |
| `go build ./...` | 0 | repository compiled |
| `/Users/iv/go/1.25.5/bin/golangci-lint run ./...` | 0 | `0 issues.`; one non-failing processor warning referenced a deleted stale integration-worktree path under `/private/tmp` |
| `git diff --check` | 0 | no tracked whitespace errors |
| `task-board validate` | 0 | `Board is valid. No issues found.` |

The final exclusive repository test ran only after no other `go test`,
`golangci-lint`, or `go build` process was active. The SHA-256 aggregate over
all tracked/untracked Go sources plus `go.mod` and `go.sum` was
`b8dd52be77db0829d0792769fbce1a96b935861d900d8417bf0a356ef9d0da40`
both immediately before and immediately after the full suite, and still matched
after full vet/build/lint. No sibling source drift occurred during or after the
accepted gate snapshot.

## Truthful development-loop failures

The following red gates are recorded as failures, not acceptance passes:

- Focused package runs exited 1 while the expanded manifest schema and
  182-case corpus still carried stale exact digests. A temporary, gated golden
  printer exposed the new digests; the printer was removed, every value was
  pinned, and the unchanged corpus command now exits 0.
- One new ZIP mutation test initially exited 1 because it rewrote the declared
  size to the existing value `13`; changing it to the contradictory value `12`
  made the negative meaningful and the same focused command exited 0.
- The first pinned focused lint exited 1 with nine findings (six checked
  integer conversions and three error-string capitalization findings). After
  explicit bounds/string comparisons, the next lint exited 1 with three more
  checked conversions and two capitalization findings. Those were repaired
  without suppressions; current focused and full lint exit 0 with zero issues.
- The first exclusive `go test -count=1 ./...` exited 1 after 326s. Only
  `TestDirectoryEntryLimitStopsTheLiveWalker` failed: the new full-finding
  validator rejected a legitimate `max_entry_count` finding attached directly
  to the rejecting canonical-tree root. All other packages passed. Pre/post
  source fingerprints were identical at
  `e2647e42c77a96dbe4b96e3597a2431782350b9778670fc929f880e55376b2be`.
- The first exact-node repair made the short suite exit 1 because it also
  treated diagnostic size/original-name fields as node content identity for
  structural ZIP/tar findings. The final rule instead binds decision,
  ancestry, optional class/variant/detector, and optional content hash. The
  focused suite, exact live-tree regression, and new full repository retry all
  then exited 0.

These failures were recoverable implementation feedback. None is reported as
a passing checklist gate.

## Scope and review focus

No existing Go path was weakened. Concurrent `internal/closuregraph` work
belongs to `TASK-260811-i3154q` and was preserved. No Kotlin,
verified-binary, ecosystem-adapter, protected-executor, or cache implementation
is claimed here.

Independent review should concentrate on full rejected-node semantic
derivation, complete hidden-finding verification, native archive metadata,
conflicting duplicate projections, resolved-use ambiguity, actual gzip
compressed-byte metering, the 182 pinned public-API vectors, and the sealed
causal authorization boundary.
