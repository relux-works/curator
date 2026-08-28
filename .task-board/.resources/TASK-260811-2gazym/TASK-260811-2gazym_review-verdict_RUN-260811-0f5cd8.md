# Reviewer verdict for TASK-260811-2gazym

Verdict: **changes requested -> to-dev**

This artifact records exactly one reviewer branch. The findings are ordinary,
autonomously repairable implementation rework; they are not a Stop-The-Line
boundary.

## Goal and reviewed scope

- Reviewer run: `RUN-260811-0f5cd8`
- Authoritative reviewer goal before verdict: `GOAL-260811-a51200`
  revision 1
- Resolved scope: `TASK-260811-2gazym`
- Reviewed producer run: `RUN-260811-799b42`
- Reviewed revision: `8ff4a238f7725bada3cfb8aa7c9c135698483caa`
- Reviewed package: `internal/artifactpolicy`
- Reviewed package source fingerprint:
  `dff80fb36b9324684551a1cc8a037f75552e3accad1bf64961b0072ccfa928a4`
- Reviewed implementation evidence resource:
  `TASK-260811-2gazym_implementation-evidence.md`
- Prior changes-requested baseline:
  `TASK-260811-2gazym_review-verdict_RUN-260811-353071.md`

## Acceptance blockers

### R1 — production trust authority still accepts caller-minted causality

The production `ManagerVerifier` does not prove the causal facts required to
establish `external_toolchain` or `local_build_output` trust roles.

1. `NewManagerVerifier` is exported and creates a fresh authorization seal from
   caller-populated `ManagerVerifierConfig`. Its dependency roots are optional,
   and its toolchain policies accept caller-provided selector, root, executable,
   version, and platform fields (`manager_verifier.go:50-106`). The resulting
   root fingerprint is useful integrity evidence, but it does not prove that a
   central Curator policy selected the root outside the dependency closure.
2. `BeginLocalOutputVerification` creates an empty directory and hashes the
   caller's `LocalOutputPlan`; it does not bind a protected-executor action,
   selected toolchain, observed process, declared reads, or observed writes
   (`manager_verifier.go:220-257`).
3. `VerifyAndProtectLocalOutput` validates only the file shape and the
   caller-supplied expected hash/size. It then unconditionally mints
   `observedProduction=true`, `preexistingInputExcluded=true`,
   `expectationIndependentlyDerived=true`, `completeInputMatched=true`, and
   `protectedPublicationValidated=true` (`manager_verifier.go:285-388`).
4. The external-package positive test demonstrates the bypass. It constructs
   compiled object bytes before the session, supplies its own expected digest
   and arbitrary plan digests, writes the bytes directly with `os.WriteFile`,
   and receives `ALLOW_OUTPUT` plus cache-publication authorization
   (`manager_external_test.go:92-135`). No declared build action produced the
   file. This is the normative T04 case: copying a pre-existing binary into the
   nominal output directory must return
   `artifact_local_output_unreceipted`.
5. The exact-corpus T04 branch does not close the gap. It constructs a
   package-private test authorization and manually flips
   `preexistingInputExcluded=false` (`conformance_test.go:164-176`), so it
   tests record validation rather than the public production path that cannot
   observe the distinction.

Impact: an adapter can present an arbitrary fingerprinted directory as a
toolchain, or copy dependency/ambient compiled bytes into a fresh staging root,
then obtain execution/publication authority. That violates the causal trust
role invariant and the explicit T01-T04 boundary before adapter execution or
cache publication.

Required rework:

- Make authority creation manager-owned and unavailable to adapter-controlled
  configuration. Bind external toolchains to a centrally selected policy and
  the complete captured dependency boundary.
- Issue local-output authority only from an opaque protected-executor receipt
  that proves the declared action, selected toolchain, exact input/read set,
  process and write observations, and protected publication. If that receipt
  belongs to `TASK-260811-27xisf`, keep the positive output seam fail-closed
  until it is supplied rather than inferring production from a file appearing
  in staging.
- Add external-package negative tests showing that (a) a caller-created
  toolchain root cannot establish `ALLOW_TOOLCHAIN`, and (b) directly copying a
  pre-existing object into manager-created staging cannot establish
  `ALLOW_OUTPUT`, adapter execution, or cache publication.
- Route the reusable T04 vector through that production path; package-private
  boolean mutation may remain a codec/unit test but cannot be the conformance
  proof.

### R2 — archive metadata is allocated before the closed leaf limit

The raw tar walker validates the metadata member bounds and then calls
`readExactAt` for PAX/GNU metadata before applying `checkLeaf`
(`containers.go:543-570`). `readExactAt` allocates the entire declared payload
with `make([]byte, size)` (`detect.go:443-455`). A valid raw tar containing a
257 MiB PAX or GNU metadata member therefore allocates, parses, and hashes that
member before reporting the 256 MiB `max_single_leaf_bytes` limit.

The native archive string-table path has the same ordering: it reads the full
`//` table at `containers.go:1365-1371`, while its leaf check begins at line
1380. Existing focused tests cover metadata counting, malformed metadata, and
entry-count limits, but do not exercise an oversized PAX/GNU or native-archive
metadata payload.

Impact: declared metadata sizes do not provide the required early refusal, and
an attacker-controlled payload can force a policy-forbidden large allocation
and parsing work before the deterministic limit diagnostic. This conflicts
with the closed resource-limit contract, which requires declared sizes to be
used for early refusal and actual streamed counts to tighten the result.

Required rework:

- Apply the declared single-leaf and applicable emitted-byte preflight before
  reading, parsing, or hashing physical tar and native-archive metadata.
- Stream metadata hashing/parsing where practical; no path should allocate a
  policy-forbidden leaf merely to discover that it exceeds the policy.
- Add oversized PAX, GNU long-name/link, and native-archive string-table cases
  that assert the first deterministic diagnostic and prove bounded traversal.

## Verified repairs and non-blocking findings

The current revision does close most of the prior RUN-260811-353071 rework:

- canonical-tree manifests independently rederive raw/canonical identity and
  reject missing or contradictory mandatory descriptors;
- physical PAX/GNU tar headers are otherwise manifested and charged, including
  invalid/repeated metadata and the 100,001-entry case;
- pinned GNU dynamic PIE, static PIE, and shared-object fixtures are genuine
  compiled artifacts with recorded provenance;
- the pinned JVM fixture is a structurally valid Java 17 class and loads in
  JShell;
- duplicate ELF execute/link-load evidence is resolved before SONAME-based
  classification;
- the global compiled dependency deny, unavailable `verified-binary-v1`, and
  Kotlin exclusion remain intact; and
- all 182 exact A/C/F/T/V corpus branches are present with pinned canonical
  outcomes and digests. Their green result does not override R1 because the
  T04 branch does not exercise production causality.

## Independent validation

The reviewer ran the current sources, not cached producer output:

| Gate | Result |
| --- | --- |
| Focused R1-R5 adversarial tests, including canonical-tree identity, physical tar metadata, pinned JVM/GNU fixtures, duplicate ELF uses, and external manager verification | pass |
| Exact reusable/production conformance tests | pass |
| Independent corpus enumeration | `total=182 A=14 C=91 F=61 T=15 V=1` |
| `go test -short -count=1 ./internal/artifactpolicy/...` | pass, 8.949s |
| Targeted `go test -race` over the exact corpus and critical rework tests | pass, 6.592s |
| `go vet ./internal/artifactpolicy/...` | pass |
| `go build ./internal/artifactpolicy/...` | pass |
| Pinned `golangci-lint run ./internal/artifactpolicy/...` | pass, `0 issues.` |
| `gofmt -l internal/artifactpolicy` | pass, no files listed |
| `git diff --check` | pass |
| Pinned fixture inspection with `file` plus JVM loading in JShell | pass |
| `task-board validate` | pass |

The reviewer did not start another serialized full-repository lane after these
semantic blockers were established. The producer's current-source full test,
vet, build, and lint evidence remains useful mechanical evidence, but a green
repository gate cannot establish the missing causal facts or undo the
pre-limit allocation ordering.

## Route

Route `TASK-260811-2gazym` to `to-dev`. After R1-R2 are repaired, refresh the
task-scoped implementation evidence, ensure the exact 182-case corpus exercises
the production role path, rerun the focused/race/full/vet/build/lint gates, and
hand the task to a new independent reviewer cycle.

No product code was modified by this reviewer. As a reviewer-archetype run,
this verdict supplies no `commit_ack`.
