# TASK-260728-zb2s4z developer handoff evidence

Amend the unreleased rc.5 candidate so `go-v1` and `go-repository-v1`
normatively identify the portable `manager-worker-v1` execution policy on macOS
and Windows, with honest capability evidence and no hardened claim.

## 1. Workspace and provenance

- Assigned worktree:
  `/Users/iv/Developer/ReluxWorks/.temp/TASK-260728-zb2s4z/curator-spec-worktree`
- Base: fresh detached worktree created from the pinned predecessor base
  `57c1f56846d221ecc55786bd3c2467ec32f11730` (`git worktree add --detach`).
- Import: byte copy of the independently accepted `TASK-260728-3b8qym`
  candidate state (`rsync -a --delete --exclude=.git`). Immediately after
  import, `diff -r --exclude=.git` between source and destination returned
  exit 0, so the starting bytes were exactly the accepted candidate.
- Predecessor worktree
  `/Users/iv/Developer/ReluxWorks/.temp/TASK-260728-3b8qym/curator-spec-worktree`
  was read only. Re-checked after all work: HEAD still
  `57c1f56846d221ecc55786bd3c2467ec32f11730`, still 125 uncommitted paths, and
  its manifest digest is still
  `sha256:33fd7aed900ec4f9f9c72be6298823a16674116b18e3b8efdd2d147574dba2b8`.
- The assigned worktree HEAD remains `57c1f56846d221ecc55786bd3c2467ec32f11730`
  with zero commits after the pin (`git rev-list --count <pin>..HEAD` = 0) and
  an unstaged index (`git diff --cached --quiet` exit 0). Nothing was staged,
  committed, tagged, pushed, or published.
- The disposable clean release probe
  `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260728-zb2s4z/release-probe`
  is a scratch repository used only to exercise the real clean-checkout gates.
  Its synthetic commit `279551f0774f1baa466c386d7578422162d3eaf3` is not a
  protocol, implementation, or downstream pin.

## 2. Normative amendments

| Document | Change |
| --- | --- |
| `protocol/core.md` §4.2 | Source-aware Go now runs inside the fixed `manager-worker-v1` graph; package-independent probes stay direct; child-start rule scoped below the worker |
| `protocol/core.md` §4.2.1 (new) | Portable execution policy: fixed four-node graph, worker identity model, one-session list/validate/permit/build state machine, mandatory controls, available-native-control rule, six deferred hardened guarantees, capability-evidence rules, package-influence exclusions, non-aliasing rule |
| `protocol/core.md` §4.2.2 | Former §4.2.1 (schema-7 external repositories) renumbered; now also reuses the execution policy. Single cross-reference at §1.1 updated |
| `protocol/core.md` §9.1 | Policy object gains `execution_policy`; cache-key text states that a different or pre-revision execution contract misses and cannot alias; capability evidence excluded from the key |
| `protocol/core.md` §9.2 | Execution policies added to the non-aliasing list |
| `protocol/core.md` §9.3 | Readers reject absent or foreign execution-policy identities as a miss, never a hit or an upgrade |
| `protocol/core.md` §10 | Marker v3 records `execution_policy` explicitly and must match the receipt; marker v2 keeps its frozen shape and binds transitively through cache key and `receipt_sha256` |
| `protocol/core.md` §12.3 | A hardened profile needs a new policy identity, a new claim schema version, and its own vectors; no in-place upgrade |
| `profiles/manager.md` §2.2 | Probe/worker split for the five argument-vector forms; child-start rule scoped below the worker |
| `profiles/manager.md` §2.2.1 (new) | Ten-step worker session, failure-boundary rules, capability-evidence honesty, and the `build_execution_*` stable diagnostic table |
| `profiles/manager.md` §2.4 | Cache lookup rejects absent or foreign execution policy; capability evidence never participates |
| `profiles/manager.md` §11.8 | External private build explicitly reuses the worker session |
| `SECURITY.md` | New "Portable execution boundary" section: expanded TCB, worker threats and answers, package-influence exclusions, the six deferred guarantees, non-aliasing. Compiler-input section now separates mandatory controls from available native controls, and states that bounding is not containment |
| `decisions/0004` | Direct-Go process-graph clause explicitly superseded before publication; build input binds the execution-policy identity |
| `decisions/0005` | Frozen-set wording corrected to the exact frozen surfaces; both drivers named as running under one execution policy |
| `decisions/0006` (new) | Portable execution policy decision, including the rejected direct-Go, hardened-Linux-only, false-claim, capability-in-cache-key, and marker-v2-field alternatives |
| `COMPATIBILITY.md`, `CHANGELOG.md`, `RELEASE.md`, `README.md`, `cli/curator.md`, `conformance/README.md`, `schemas/v1/README.md` | Compatibility, release-checklist, CLI reporting, and suite documentation for the revision |
| `docs/portable-go-execution-policy.md` (new) | Authoring and implementer guidance, native-control matrix, and the "what this policy does not promise" list |

## 3. Mandatory versus unavailable controls

Mandatory on every supported host; a host that cannot apply all of them rejects
before the worker or Go with `build_execution_control_unavailable`:

`fixed-offline-vendored-go`, `fixed-argument-vectors`, `fixed-empty-environment`,
`identity-verified-manager-owned-worker`,
`pre-launch-worker-identity-verification`, `post-exec-identity-reverification`,
`manager-private-staging-roots`, `manager-derived-output-path`,
`bounded-wall-clock-deadline`, `bounded-combined-output`,
`bounded-artifact-size`, `closed-standard-input-and-descriptors`,
`worker-domain-teardown`, `no-artifact-execution`,
`available-native-controls-applied`.

Native controls, applied when the host provides them and recorded as
unavailable otherwise:

| Control | macOS | Windows |
| --- | --- | --- |
| descendant-domain-termination | process group and session teardown | Job Object kill-on-close |
| active-process-count-limit | unavailable as a private domain | Job Object active-process limit |
| aggregate-memory-limit | unavailable as a private domain | Job Object process/job memory limit |
| per-file-size-limit | `RLIMIT_FSIZE` | unavailable as a private domain |
| inherited-handle-restriction | close-on-exec plus explicit descriptor release | explicit handle inheritance list |

Deferred to `STORY-260728-327soo`, never claimable under `manager-worker-v1`,
and never a reason to reject a portable build: `total-network-denial`,
`read-only-source-and-toolchain`, `private-build-root-only-writes`,
`hard-aggregate-descendant-resource-bounds`, `exact-executable-allowlisting`,
`fail-closed-capability-preflight`.

## 4. Wire and vector changes

- `schemas/v1/common.schema.json`: new closed `$defs/goExecutionPolicyV1`
  (`const: "manager-worker-v1"`). `goBuildPolicyV1`,
  `goRepositoryBuildPolicyV1`, `buildRecordV1WithReceiptVersion`, and
  `buildRecordV2` all require `execution_policy` and reference that one
  constant. `buildRecordV1` (marker v2) deliberately unchanged.
- `schemas/v1/conformance-claim-v3.schema.json`: each `build_drivers` item
  requires `execution_policy` bound to the same constant, so a hardened claim
  is structurally impossible in claim v3.
- New vector `conformance/v1/vectors/go-host-execution-policy.json` with the
  fixed process graph, 12 ordered session states, 15 mandatory controls, 5
  native controls with per-platform evidence, 6 deferred guarantees, 8
  package-influence surfaces, 12 identity/protocol negatives, 5
  capability-evidence cases, and 3 non-aliasing cache identities.
- New negative schema cases: `invalid-hardened-execution-policy` and
  `invalid-missing-execution-policy` for build receipt v1 and v2 and claim v3;
  `invalid-local-hardened-execution-policy`,
  `invalid-external-hardened-execution-policy`,
  `invalid-local-missing-execution-policy`, and
  `invalid-external-missing-execution-policy` for install marker v3.
- The generated `go-v1` receipt example now derives its cache key from its own
  CCJ-1 input instead of a literal, matching the receipt-v2 treatment.
- Suite inventory: 42 schemas, 422 manifest-listed files (was 411). Largest
  fixture is still `pack-index.json` at 18,377 bytes, below the 65,536-byte
  ceiling.

### Non-aliasing proof (independently recomputable)

Each key is `"sha256:" || lowercase_hex(SHA-256(CCJ-1(input)))` over the exact
input stored beside it in `cache_identity`:

| Identity | Cache key | Schema valid |
| --- | --- | --- |
| portable `manager-worker-v1` | `sha256:529370122ae11e2e961d5265b1a020e046bcd43165b2eb96b05e73a51187ac9b` | yes |
| reserved `hardened-worker-v1` | `sha256:13736230d33ce59de7f7323dcd4cffd510655ad8dabd5ee9e8b6cb182ec70037` | no |
| pre-revision rc.4 (no execution policy) | `sha256:3fcd714a40e8918eb67dbd35d435875dcce6c9047da811a1fa26626e5e57be48` | no |

The third value is exactly the accepted rc.4 candidate `go-v1` cache key, so
the change is proved to be a miss rather than an alias. Receipt v2 recomputes
to cache key
`sha256:07dd911a7edc29b906a021aa6e1449632ce91c2e5a3eb0ea4f851cb84fe5c492` and
receipt hash
`sha256:11d2bf4df52638ef353b3286c426261eac2a73b0b64a32f85d78c04490072cea`,
both carried unchanged by the mixed marker and mixed plan.

## 5. Compatibility guards

- `TestRC4CompiledArtifactsRemainByteFrozen` retains every rc.4 artifact whose
  bytes this revision does not touch. The `go-v1` receipt example was removed
  from that map because its bytes MUST change; the removal is compensated by a
  new dedicated guard, not dropped.
- New `TestGoV1ExecutionPolicyRevisionCannotAliasRC4` pins the old rc.4 example
  digest `93217cf1ce2965435042f8e20ebfec45498bae67a128cc16f38b3f4c8b64ecab` and
  asserts it is no longer produced, that the receipt key is the CCJ-1 digest of
  its own input, and that portable, hardened, and pre-revision keys are
  pairwise distinct.
- New `TestExecutionPolicyIsBoundIntoReceiptMarkerAndClaim` checks receipt v1
  and v2, the mixed marker, the claim example, the closed `$defs`, and that
  marker v2's frozen record shape did not acquire the field.
- New `TestPortableGoHostExecutionPolicyContract` checks the whole vector,
  including session ordering, exact mandatory-control and deferred-guarantee
  inventories, per-platform native evidence, and pre-worker versus pre-compiler
  failure boundaries.
- Python `validate_go_host_execution_policy` and
  `validate_local_go_receipt_oracles` mirror those checks;
  `test_portable_execution_vector_rejects_dishonest_evidence` runs 13 targeted
  mutations (hardened claim permitted, unavailable control reported as applied,
  capability evidence entering the cache key, unavailable control rejecting a
  build, deferred guarantee claimed or blocking, package influence reaching the
  worker or becoming expressible, identity failure still publishing, permit
  before graph validation, mandatory control downgraded to optional, aliased
  hardened key, hardened policy marked schema valid) and requires each to be
  rejected.
- `tools/release_gate.py` now fails a candidate whose release metadata names an
  unknown portable policy, claims the hardened profile, omits the deferring
  story, or omits the block; `test_release_gate.py` proves all four rejections.

### Preserved invariants (verified by byte comparison against the predecessor)

Manifest schemas 1-5, `agent-skill-v6`/`csk-skill-v6`, `agent-skill-v7`/
`csk-skill-v7`, `curator-build-v1`, `build-receipt-v1`, `build-receipt-v2`,
`install-marker-v1/v2/v3`, `conformance-claim-v1/v2` schema files, and the
`agent-skill-v6`, `csk-skill-v6`, `install-marker-v2`, and
`conformance-claim-v2` valid cases are byte-identical. The
`external-repository-lifecycle`, `external-repository-acquisition`,
`pack-index`, and `conformance-claim-v3-qualification` artifacts are
byte-identical, so whole-snapshot audit-before-cache/compiler ordering and the
empty candidate platform claims are untouched. No file was removed; 98 files
changed and 13 were added.

## 6. Exact rc.5 candidate identity

- Protocol version: `1.0.0-rc.5`
- New downstream candidate protocol pin (SHA-256 of
  `conformance/v1/manifest.json`):
  `sha256:bfe49f254332cb6f38d47b015679b4e6a4ec46eb13207dfa06e94171651b9124`
  (predecessor was `sha256:33fd7aed...`)
- Release metadata SHA-256:
  `ead54508d5435a6e05f96e8fc399d5acfebab9c813a1e704faf4c8acec309766`
- `release/1.0.0-rc.5.json` now carries
  `execution_policy.portable = "manager-worker-v1"`,
  `hardened_profile_claimed = false`,
  `hardened_profile_owner = "STORY-260728-327soo"`, and the superseded rc.4
  cache key as negative evidence.
- Downstream environment remains `CURATOR_CONFORMANCE_ROOT`;
  `committed_release_pin_advanced` remains `false`.
- Claim-v3 candidate claims remain empty. macOS and Windows remain
  pending downstream native evidence; Linux remains excluded until
  `TASK-260728-1skseh`.
- Other exact file digests: `go-host-execution-policy.json`
  `cc0fa96c0a27558292c49949b046b5cdc3a9e6d78ce25d29afc678a88a28f43f`;
  `build-receipt-v1/valid.json`
  `1a887eb6bb436a3491250b0814dded2a1b1d108640ba67837ba9e89b1183daf3`;
  `build-receipt-v2.json`
  `3fb1ad89bd3085c3862450a4fd7356ffe00f2f0a030bcec2d2a37f34b7bfb2a5`;
  `install-marker-v3-mixed.json`
  `9c8c6bc7c63038217f022141412e8c194069dfedef64b14f9bed0c6ef01d5877`;
  `common.schema.json`
  `aa927e2149399f3128af0c5ce1c872da6fa5e7ac09c9a2cb687dc5c58e199501`.

## 7. Gates and exact exit codes

Every command below was run directly as a standalone process, not through
`tee` or a pipe. Green gates in the assigned worktree:

| Command | Exit |
| --- | --- |
| `python tools/validate.py` (task venv) — `validated 42 schemas and 422 vector files` | 0 |
| `python -B -m unittest discover -s tools -p 'test_*.py'` — 22 tests, OK | 0 |
| `go test ./tools/...` | 0 |
| `go vet ./tools/...` | 0 |
| `test -z "$(gofmt -l tools)"` | 0 |
| `python -m compileall -q tools` | 0 |
| `go build -o <task-temp>/generate-vectors-bin ./tools/generate-vectors` | 0 |
| `git diff --check` | 0 |
| `git diff --cached --quiet` | 0 |
| `git rev-list --count 57c1f56...:HEAD` = 0 | 0 |

Deterministic regeneration in the assigned worktree: three consecutive
`go run ./tools/generate-vectors -root .` runs produced the identical aggregate
digest `f04f76f12bf50a809b5a9dfa5f163f8cfe76b525ec54a9aec9dd8f6fd33020b3` over
every file under `conformance/v1` and `release/`.

Clean disposable release probe (byte-identical to the assigned worktree before
and after the gates, `diff -r --exclude=.git` exit 0 both times):

| Command | Exit |
| --- | --- |
| `make regenerate-check`, first run | 0 |
| `make regenerate-check`, second consecutive run | 0 |
| `make release-check VERSION=1.0.0-rc.5` — validation, Python, Go, regeneration, exact metadata pin, execution-policy honesty, clean checkout, version, candidate gate | 0 |
| `git status --porcelain` after the gates — zero lines | 0 |

Expected-red gates, reported truthfully as failures:

1. `go test ./tools/...` immediately after adding `execution_policy` to the
   schema: exit 1, three failures —
   `TestRC4CompiledArtifactsRemainByteFrozen` (the `go-v1` receipt example
   digest changed from `93217cf1...` to `1a887eb6...`),
   `TestBuildReceiptV1SchemaIsStrictAndPortable` (policy property count 11 to
   12), and `TestGeneratedBuildReceiptAndMarkerV2Cases` (two new negative
   cases). All three are the intended effect of the authorized revision and
   were closed by updating the guards, not by weakening them.
2. `gofmt -l tools` after the first generator edit: non-empty output naming
   `tools/generate-vectors/main.go`; fixed with `gofmt -w tools`.

Not run, and why: no native macOS or Windows manager execution was exercised.
This task is specification-only and adds no manager implementation, so there is
no worker binary to run. Native qualification remains downstream, which is why
the candidate still emits zero claim-v3 tuples.

## 8. Boundaries preserved

- No Curator or csk implementation code, no manager, no worker binary.
- No new package-controlled field in any manifest or descriptor; the build
  command surfaces are byte-identical.
- No generic language driver, no fallback, no package-selected program, argv,
  environment, output, credential, helper, filter, signer, hook, plugin, or
  generator.
- Independent audit-before-cache/compiler ordering, fail-closed offline
  behavior, and operator-owned signing are unchanged.
- No real remote was contacted and no platform evidence was fabricated.
- No source repository, predecessor worktree, release, ref, tag, or downstream
  pin was modified.
