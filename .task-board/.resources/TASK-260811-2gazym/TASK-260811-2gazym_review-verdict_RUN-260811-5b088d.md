# Independent reviewer verdict for TASK-260811-2gazym

Verdict: **changes requested -> to-dev**

This is ordinary, actionable implementation/test rework. It is not a stop-the-line boundary.

## Review authority and inputs

- Reviewer run: `RUN-260811-5b088d`
- Authoritative goal at the final directive checkpoint: `GOAL-260811-b2090a` revision 1
- Resolved scope: `TASK-260811-2gazym`
- Reviewed producer evidence: `TASK-260811-2gazym_rework-evidence_RUN-260811-a4d8a4.md`
- Producer evidence SHA-256: `075859fd165f935e3b45fdeaa0b138ac5c61e4d4a8fc9f87fd7f43df735f9eb6`
- Prior verdict: `TASK-260811-2gazym_review-verdict_RUN-260811-290cd4.md`
- Prior verdict SHA-256: `08c5db588bc14a70854dd39094f293fbc78926bb45c224382dc900365967fe91`
- Reviewed Git HEAD: `8ff4a238f7725bada3cfb8aa7c9c135698483caa`
- Independently reproduced artifactpolicy source fingerprint: `ca53bba924ed0cf8ecdf81be1a680cfe38bef04cfe9531ccb5adb01c02cba2d3`
- Independently reproduced 364-file all-Go fingerprint: `dcf6064e5a87cf1f09237789fe0456e09f9d66f7f683a6b7c5821ad560d78910`
- Independently reproduced exact corpus fingerprint: `87a5cb6afb1c120cf75979cccd57fe2702c9a7dd74bee22dfa80418e1f26750e`

No product code was modified by this reviewer.

## Acceptance blocker R1: the required external-package positive does not execute the selected toolchain

The production selector is materially improved and its API is closed: `SelectExternalToolchain` accepts the closed selector plus admitted dependencies, selection state is private, the root is centrally selected, descriptor fields are audit-only, and `AuthorizeSelectedAdapterExecution` performs a time-of-use recheck.

However, the claimed real-toolchain external integration in `internal/artifactpolicy/manager_external_test.go:84-104` never starts a process. It:

1. calls `AuthorizeSelectedAdapterExecution`;
2. increments a local `adapterStarts` integer when authorization returns nil;
3. checks that the returned path is outside the caller root; and
4. calls `os.Stat` on the executable.

There is no `exec.Command`, `exec.CommandContext`, or equivalent invocation anywhere in `internal/artifactpolicy`. Repository search finds the authorization API used only by that test and the mode-drift test. Consequently, the test proves path authorization and existence, but not the required manager-to-adapter execution boundary or that the executable actually run is the centrally selected, freshly rechecked tool.

The production link/special-node negative is also not exercised. `internal/artifactpolicy/conformance_test.go:128-138` constructs a test-only authorization record and flips `containedLinksValidated` or `ordinaryNodesValidated`; it does not create an escaping link or special node under a selected toolchain root and run the real fingerprinter. The arbitrary-root test uses a real copied tree, and the mode-drift test mutates a real selected file, but dependency drift is only changed to a nil boundary rather than a same-count/different-digest boundary. The T04 hard-link branch likewise supplies no output authority without constructing an actual hard link.

Required R1 evidence before the next review:

- From an external-package test, select through `curator-runtime-go-toolchain-v1`, authorize, then execute the exact returned real Go executable under a bounded context. Assert a successful, bounded `go version` result and bind the observed version/target back to the selected evidence. Increment a process-start counter only at the actual process-launch site.
- Exercise the production fingerprinter/recheck against a real escaping symlink and a real special node in a controlled selectable-root fixture, and prove zero process starts.
- Add a same-entry-count but different dependency-manifest digest negative, with zero process starts.
- Keep arbitrary caller roots/copies untrusted, add the concrete hard-link publication negative, and keep A08 manager output authority unavailable until `TASK-260811-27xisf`.

## R2 conclusion: bounded metadata repair is technically sound

Source inspection and sparse probes confirm that local/global PAX path and linkpath records, GNU long-name/long-link records, GNU/COFF ar string-table names (including unreferenced names), and BSD `#1/` names are bounded before value allocation/read using the full canonical `container!/member` path. Non-path PAX values are streamed in bounded chunks. The ar string-table index is keyed by exact record starts, so mid-record references cannot resolve. Invalid unreferenced names are validated in both archive order permutations. No R2 implementation blocker was reproduced.

## Acceptance evidence gap R3: the complete self-rehashed BSD forgery matrix is not present

The implemented BSD accounting is correct for the reviewed fixture: physical extended-name bytes plus member bytes are charged once, logical content size/hash remain member-only, and `archive-ar-v1` records declared size, member size, extended-name size/hash, and original name. The regular-node mutation suite rejects self-rehashed changes to extended-name size/hash, declared size, member size, and original name.

The required complete matrix is nevertheless missing:

- there is no BSD-specific self-rehashed canonical node-path forgery;
- there is no BSD-specific self-rehashed manifest-accounting forgery; and
- `TestBSDExtendedSymbolTableNameRemainsExactMetadata` only verifies a valid round trip. It does not forge and rehash BSD symbol-metadata name size/hash/original-name/path/accounting fields and prove rejection.

Generic codec mutation tests do not replace the required native-archive semantic evidence. Add these exact BSD regular-member and symbol-metadata mutations and assert `DecodeManifest` rejects each after recomputing the manifest digest.

## Preserved invariants and independent validation

- Exact conformance remains 182 cases: `A=14 C=91 F=61 T=15 V=1`; the canonical corpus test passed.
- Compiled dependency classes remain globally deny-dominant.
- F14 preserves exact raw-payload manifest binding while comparing the canonical logical projection across order permutations.
- `verified-binary-v1` and production A08 output authority remain unavailable.
- Kotlin/Gradle/Maven surfaces are absent from `internal/artifactpolicy`.

Independent gates, all passing:

- targeted R1-R3 tests: `go test -count=1 -v ./internal/artifactpolicy -run <review-pattern>` (4.604s);
- exact corpus: `go test -count=1 ./internal/artifactpolicy -run '^TestReusableArtifactManifestV1ConformanceCorpus$'`;
- independent JSON event count: `total=182 A=14 C=91 F=61 T=15 V=1`;
- package suite: `go test -count=1 ./internal/artifactpolicy/...` (33.857s);
- focused race suite including exact corpus: passed (10.011s);
- `go vet ./internal/artifactpolicy/...`;
- `go build ./internal/artifactpolicy/...`;
- `gofmt -l internal/artifactpolicy` (no output);
- `git diff --check`;
- pinned `/Users/iv/go/1.25.5/bin/golangci-lint run ./internal/artifactpolicy/...` (`0 issues`);
- `task-board validate` (`Board is valid. No issues found.`).

The producer's repository-wide lane was not duplicated because the independently reproduced all-Go fingerprint is unchanged, as explicitly directed.

## Route

Return to `to-dev`. After the real selected-tool execution integration and the missing production/fraud negatives are added, refresh task-scoped evidence and submit to a new independent reviewer cycle.

This reviewer supplies no `commit_ack`.
