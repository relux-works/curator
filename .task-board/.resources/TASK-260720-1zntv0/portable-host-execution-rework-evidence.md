# TASK-260728-zb2s4z rework 1 — developer handoff evidence

Closes review-1 findings R1 and R2 in the existing producer worktree. No
predecessor worktree, release, ref, tag, pin, or platform claim was touched, and
nothing was staged, committed, or published.

## 1. Workspace and provenance

- Assigned worktree:
  `/Users/iv/Developer/ReluxWorks/.temp/TASK-260728-zb2s4z/curator-spec-worktree`
  (unchanged from the reviewed run; no new worktree was created).
- HEAD is still exactly `57c1f56846d221ecc55786bd3c2467ec32f11730`,
  `git rev-list --count 57c1f568..HEAD` is `0`, and `git diff --cached --quiet`
  exits `0`. Nothing was staged, committed, tagged, pushed, or published.
- Accepted predecessor
  `/Users/iv/Developer/ReluxWorks/.temp/TASK-260728-3b8qym/curator-spec-worktree`
  was read only. Re-checked after all work: HEAD still `57c1f568…`, still 125
  uncommitted paths, manifest digest still
  `sha256:33fd7aed900ec4f9f9c72be6298823a16674116b18e3b8efdd2d147574dba2b8`.
- Disposable clean probe
  `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260728-zb2s4z/release-probe-r1-clean`
  is a scratch repository. Its synthetic commit
  `5a0e2ffc15fc063f3e4f24d3dbeee25fb3482f61` is not a protocol, implementation,
  or downstream pin.
- Two untracked scratch artifacts left behind by the earlier run — a stray
  `generate-vectors` binary at the worktree root and `tools/__pycache__` — were
  removed before the final probe, so the probe contains only candidate product
  bytes (514 tracked files).

## 2. R1 — portable deferral no longer contradicts enforcement text

Every absolute statement that implied kernel-enforced read-only source or exact
executable allowlisting is now stated as the portable mechanism it actually is,
with the stronger guarantee named separately.

| File | Change |
| --- | --- |
| `protocol/core.md` §4.2 | "only … may start" and "presented read-only to child processes" replaced by: within this boundary the manager and worker start no other program, each identity is verified, neither writes to the frozen snapshot or `GOROOT`, and both identities are re-verified after the last child exits |
| `protocol/core.md` §4.2.1 | New normative table pairing each portable mechanism with what it means, what it does not mean, and the single deferred guarantee it stops short of; explicit single failure boundary |
| `protocol/core.md` §9.1 | `network: "none"` defined as fixed offline Go module/proxy/checksum-database/VCS configuration and no manager- or Go-initiated network access — explicitly not a kernel network-denial claim |
| `profiles/manager.md` §2.2 | Same replacement, cross-referenced to §2.2.1 |
| `profiles/manager.md` §2.2.1 | Exact failure boundary paragraph; identity re-verification now covers worker, snapshot, and toolchain |
| `SECURITY.md` compile-only section | "A manager MUST reject … source/context separation … process graph" scoped to the mandatory portable control set, with the statement that the deferred guarantees never cause rejection |
| `SECURITY.md` portable-boundary section | Mechanism/deferred-guarantee table replaces the prose list; one failure-boundary sentence |
| `SECURITY.md` compiler-input section | "Source is read-only to children" replaced by the frozen-snapshot plus identity-re-verification mechanism |
| `decisions/0004` | Rejection list no longer says "any executable child outside …"; the superseded clause now names both boundaries |
| `decisions/0006` | Mechanism-versus-guarantee paragraph, exact failure boundary, and a new rejected alternative: keeping the absolute wording alongside the deferral |
| `docs/portable-go-execution-policy.md` | "What this policy does not promise" is now a two-column mechanism/guarantee table; new "one failure boundary" section |
| `CHANGELOG.md`, `COMPATIBILITY.md`, `cli/curator.md`, `conformance/README.md` | Same reconciliation in release, compatibility, CLI, and suite documentation |

The failure boundary is now stated identically in normative text, the manager
profile, `SECURITY.md`, decision 0006, the author guide, and the executable
vector:

- a mandatory portable control that cannot be applied rejects with
  `build_execution_control_unavailable` **before the worker starts** and
  publishes nothing;
- an unavailable inventory native control and the absence of any of the six
  deferred hardened guarantees never reject, never emit a diagnostic, and never
  block publication.

## 3. R2 — closed native-control and capability-evidence vocabulary

### Exhaustive versioned inventory

`available_native_controls` is gone. The single authority is
`conformance/v1/vectors/go-host-execution-policy.json#native_control_inventory`,
version `rc5-native-control-inventory-v1`, with `exhaustive: true`, closed
`availability_states` (`available`, `unavailable`), a closed
`unavailable_reasons` vocabulary (`no-private-aggregate-domain`),
`probe_timing: pre-worker-launch`, and `probe_scope: per-operation`. Each control
carries a closed per-platform record `{availability, mechanism,
unavailable_reason}`:

| Control | macOS | Windows |
| --- | --- | --- |
| `descendant-domain-termination` | available: process group and session teardown | available: Job Object kill-on-close |
| `active-process-count-limit` | unavailable: `no-private-aggregate-domain` | available: Job Object active-process limit |
| `aggregate-memory-limit` | unavailable: `no-private-aggregate-domain` | available: Job Object process and job memory limit |
| `per-file-size-limit` | available: `RLIMIT_FSIZE` | unavailable: `no-private-aggregate-domain` |
| `inherited-handle-restriction` | available: close-on-exec plus explicit descriptor release | available: explicit handle inheritance list |

The open-ended mandatory control `available-native-controls-applied` was
replaced by `inventory-native-controls-applied`. Three mandatory controls were
added to name the portable mechanisms explicitly:
`fixed-manager-selected-process-graph`, `frozen-source-snapshot-integrity`, and
`closed-capability-evidence-record`. The mandatory set is now exactly 18 names
and is checked for exactness in both harnesses.

### Closed evidence record

`capability-evidence-v1` is normative in `protocol/core.md` §4.2.1,
`profiles/manager.md` §2.2.1, the author guide, and the vector's
`capability_evidence_record` section:

- record fields exactly `record_version`, `execution_policy`, `platform`,
  `controls`;
- one entry per inventory control, each exactly `{name, availability, status,
  probed_at}`;
- `availability` ∈ {`available`, `unavailable`}, `status` ∈ {`applied`,
  `unavailable`}, `probed_at` ∈ {`pre-worker-launch`};
- availability probed once per operation before worker launch — a host label, a
  build-time constant, a configuration value, or a cached result is not a probe;
- exposure exactly `dry-run-plan-result`, `install-result`, `status-result`;
- exclusion exactly `cache-key`, `conformance-claim`, `install-marker`,
  `receipt`;
- eight closed consistency rules, each bound to a stable diagnostic;
- generated per-platform examples derived from the inventory, so a record and
  the inventory cannot drift.

Error rules: contradictory availability/status, a missing, duplicated, extra, or
unknown entry, an unknown `record_version`, and un-probed availability are
`build_execution_capability_evidence_invalid`. A deferred guarantee named as an
entry, or a record `execution_policy` other than `manager-worker-v1`, is
`build_execution_hardened_claim_forbidden`.

### New executable negative guards

Vector `capability_evidence_cases` is now an exact 11-case inventory. New cases:
`available-control-cannot-be-reported-as-unavailable`,
`unknown-native-control-is-rejected`,
`missing-native-control-entry-is-rejected`,
`duplicate-native-control-entry-is-rejected`,
`unknown-evidence-record-version-is-rejected`, and
`hardened-execution-policy-in-evidence-record`. Both harnesses recompute the
expected diagnostic from the case fields instead of trusting the recorded value,
and both reject any case whose `expected_error` is
`build_execution_control_unavailable`, which is the "reporting fault became a
mandatory rejection" path.

New `deferred_capability_rejection_guards` gives each of the six deferred
guarantees `in_mandatory_controls: false`, `in_native_control_inventory: false`,
`in_capability_evidence_record: false`, `portable_rejection_code: null`, and
`build_permitted_when_absent: true`. The validators additionally prove the six
names appear in no mandatory control, no inventory entry, and no evidence
example.

New `failure_boundary` and `policy_semantics` sections make R1 executable: the
boundary has exactly three keys with fixed verdicts, and each of the six deferred
guarantees is answered by exactly one portable mechanism with a non-empty
`means` and `does_not_mean`.

New identity cases `post-build-source-snapshot-mutated` and
`unexpected-program-started-below-the-worker` make the two portable substitutes
for kernel read-only and allowlisting detectable rather than assumed. New session
state `parent-native-control-availability-probe` is ordered before
`parent-worker-identity-verification`.

### Release metadata

`release/1.0.0-rc.5.json` `execution_policy` now also records
`native_control_inventory_version` and `capability_evidence_record_version`.
`tools/release_gate.py` rejects a candidate that omits or drifts either;
`tools/test_release_gate.py` proves four new rejections alongside the existing
four.

## 4. Test and validator changes

- `tools/validate.py`: new constants
  (`NATIVE_CONTROL_INVENTORY_VERSION`, `CAPABILITY_EVIDENCE_RECORD_VERSION`,
  `UNAVAILABLE_NATIVE_CONTROL_REASON`, `NATIVE_CONTROL_INVENTORY`,
  `CAPABILITY_EVIDENCE_RECORD_FIELDS`, `CAPABILITY_EVIDENCE_ENTRY_FIELDS`,
  `CAPABILITY_EVIDENCE_CASES`, `POLICY_SEMANTIC_KEYS`); the inventory block now
  compares each per-platform record to a normative expectation; three new
  functions `validate_capability_evidence_record`,
  `validate_capability_evidence_cases`, `validate_execution_failure_boundary`.
- `tools/test_validate.py`:
  `test_portable_execution_vector_rejects_dishonest_evidence` grew from 13 to 40
  targeted mutations. Every one is asserted to be rejected. New mutations cover
  inventory loss/drift/non-exhaustiveness, contradictory platform availability,
  per-host probing, a deferred guarantee entering the inventory, an open record
  field or probe time, evidence entering the receipt, a lost consistency rule, a
  dropped or contradictory example, a hardened example policy, each of the four
  new negative cases becoming acceptable, a reporting fault becoming a mandatory
  rejection, a dropped case, every failure-boundary inversion, a deferred
  guarantee gaining a rejection code or blocking a build, a lost rejection guard,
  an undefined `network=none`, a lost portable mechanism, and two mechanisms
  answering one guarantee.
- `tools/generate-vectors/main_test.go`:
  `TestPortableGoHostExecutionPolicyContract` now pins the exact mandatory-control
  count, the exact inventory with per-platform records, and the exact evidence
  case inventory, and recomputes each expected diagnostic. Two new helpers
  `assertCapabilityEvidenceRecord` and `assertExecutionFailureBoundary` mirror the
  Python guards.
- `tools/generate-vectors/main.go`: new constants and helpers
  (`availablePlatform`, `unavailablePlatform`, `capabilityEvidenceRecord`,
  `evidenceRule`, `evidenceCase`, `withEvidenceRecordVersion`,
  `withEvidenceRecordExecutionPolicy`) emitting the new sections.

## 5. Preserved evidence (re-verified after the rework)

Every value the independent review recomputed is byte-identical:

| Artifact | SHA-256 |
| --- | --- |
| portable cache key | `sha256:529370122ae11e2e961d5265b1a020e046bcd43165b2eb96b05e73a51187ac9b` |
| reserved hardened key | `sha256:13736230d33ce59de7f7323dcd4cffd510655ad8dabd5ee9e8b6cb182ec70037` |
| pre-revision rc.4 key | `sha256:3fcd714a40e8918eb67dbd35d435875dcce6c9047da811a1fa26626e5e57be48` |
| receipt-v2 cache key | `sha256:07dd911a7edc29b906a021aa6e1449632ce91c2e5a3eb0ea4f851cb84fe5c492` |
| receipt-v2 hash in the mixed marker | `sha256:11d2bf4df52638ef353b3286c426261eac2a73b0b64a32f85d78c04490072cea` |
| `expected/external-repository/build-receipt-v2.json` | `3fb1ad89bd3085c3862450a4fd7356ffe00f2f0a030bcec2d2a37f34b7bfb2a5` |
| `expected/external-repository/install-marker-v3-mixed.json` | `9c8c6bc7c63038217f022141412e8c194069dfedef64b14f9bed0c6ef01d5877` |
| `schema-cases/build-receipt-v1/valid.json` | `1a887eb6bb436a3491250b0814dded2a1b1d108640ba67837ba9e89b1183daf3` |
| `schemas/v1/common.schema.json` | `aa927e2149399f3128af0c5ce1c872da6fa5e7ac09c9a2cb687dc5c58e199501` |
| `schemas/v1/conformance-claim-v3.schema.json` | `7b8fcdb5e8f608fd3b7bdbcfe4be5059b84c8eedf5e7a612dcb7253cf17052c8` |

Byte comparison against the accepted predecessor: only
`schemas/v1/common.schema.json`, `schemas/v1/conformance-claim-v3.schema.json`,
and `schemas/v1/README.md` differ under `schemas/`. Manifest schemas 1–7
(`agent-skill-v1`…`v7`, `csk-skill-v1`…`v7`), `build-receipt-v1/v2`,
`install-marker-v1/v2/v3`, `conformance-claim-v1/v2`, and `curator-build-v1` are
byte-identical. `external-repository-acquisition.json`,
`external-repository-lifecycle.json`, the pack/index fixtures, and
`conformance-claim-v3-qualification.json` are byte-identical, so
audit-before-cache/compiler ordering, offline external-source errors, signing
boundaries, and the empty candidate platform claims are untouched. 98 files
changed and 15 added versus the predecessor; none removed.

Suite inventory unchanged at 42 schemas and 422 manifest-listed files. The
largest shared external-repository fixture is still `pack-index.json` at 18,377
bytes, below the 65,536-byte ceiling that `tools/validate.py` applies to
`conformance/v1/fixtures/external-repository/*.json`.

Candidate claim-v3 claims remain empty; macOS and Windows remain pending
downstream native evidence; Linux remains excluded until `TASK-260728-1skseh`;
`committed_release_pin_advanced` remains `false`.

## 6. New exact rc.5 candidate identity

Bytes changed, so the digests moved:

- new downstream candidate protocol pin (SHA-256 of
  `conformance/v1/manifest.json`):
  `sha256:58f8d2299c8f4a5ed78546913e567f637f1cc905dc5212b460f4097be7ff2af9`
  (previous producer state was `sha256:bfe49f25…`; accepted predecessor was
  `sha256:33fd7aed…`)
- release metadata SHA-256:
  `b163f445f206c17dc618cc10b3957ca2f6f1b28607288162c0cfc5de02d83ee6`
- `conformance/v1/vectors/go-host-execution-policy.json` SHA-256:
  `c3d42f763afdcfa229430e7de5bb9f1e9f44607a7790aef6f4e0bf6d1bc644de`

Recompute the pin with:
`shasum -a 256 conformance/v1/manifest.json`.

## 7. Gates and exact exit codes

Every command was run directly as a standalone process, not through `tee` or a
pipe chain.

Assigned worktree:

| Command | Exit |
| --- | --- |
| `python -B tools/validate.py` — `validated 42 schemas and 422 vector files` | 0 |
| `python -B -m unittest discover -s tools -p 'test_*.py'` — 22 tests, OK | 0 |
| `go test ./tools/...` | 0 |
| `go vet ./tools/...` | 0 |
| `test -z "$(gofmt -l tools)"` | 0 |
| `python -m compileall -q tools` | 0 |
| `go build -o <task-temp>/generate-vectors-bin ./tools/generate-vectors` | 0 |
| `git diff --check` | 0 |
| `git diff --cached --quiet` | 0 |
| `git rev-list --count 57c1f568..HEAD` = 0 | 0 |

Deterministic regeneration in the assigned worktree: three consecutive
`go run ./tools/generate-vectors -root .` runs produced the identical aggregate
digest `e1629c33f9fdfcc48cf98214efd60e533aa0ba1d537f52c802540b0a9d9f0b4e` over
every file under `conformance/v1` and `release/`.

Disposable clean probe (byte-identical to the assigned worktree before and after
the gates, `diff -r --exclude=.git` exit 0 both times, 514 tracked files):

| Command | Exit |
| --- | --- |
| `make regenerate-check`, first run | 0 |
| `make regenerate-check`, second consecutive run | 0 |
| `make release-check VERSION=1.0.0-rc.5` — validation, Python, Go, regeneration, exact metadata pin, execution-policy/inventory/record honesty, clean checkout, version, candidate gate | 0 |
| `git status --porcelain` after the gates — zero lines | 0 |

Expected-red gates during the rework, reported truthfully as failures:

1. `go test ./tools/...` after replacing the native-control section but before
   updating the Go oracle: exit 1, one failure —
   `TestPortableGoHostExecutionPolicyContract` at `main_test.go:740`, because
   `entry_count` decodes as `json.Number` and the `!= float64(1)` comparison
   misclassified a valid case. Closed by comparing the decoded value as text,
   not by weakening the assertion.
2. `gofmt -l tools` after the first generator edit: non-empty output naming
   `tools/generate-vectors/main.go`; fixed with `gofmt -w tools`.

Not run, and why: no native macOS or Windows manager execution was exercised.
This task is specification-only and adds no manager implementation, so there is
no worker binary to run and no native control to probe for real. Native
qualification remains downstream, which is why the candidate still emits zero
claim-v3 tuples and why the inventory availability recorded here is a normative
specification statement rather than a measured platform claim.

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
  pin was modified. This handoff makes no acceptance judgement about its own
  work.
