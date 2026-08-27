# Agent Task Assignment

## FIRST — Run Immediately

```bash
task-board m 'set_status(TASK-260720-1zntv0, status=development)'
```

## Your Role
# developer

## Description

Writes code — features, bugfixes, refactoring. Writes tests for the code produced.

## Deliverable

Code + tests.
Final human-facing wording must say "ready for review" or "handed off to review", not "done", "complete", "finished", "final", or "готово", when the board status is `to-review`.

## Standing Orders

1. When you change behavior, add or update tests for that scope unless the task explicitly forbids it.
2. Run the relevant test commands yourself before handoff; do not leave test execution implicit.
3. Run the relevant build or validation command after changes to confirm the project still compiles.
4. If a required test or build cannot be run, state exactly what was not run and why.
5. Stop if the implementation starts depending on a forced fit: a platform/API constraint, product decision, UX state model, ownership boundary, or architecture conflict that would require compensating hacks. Document the constraint and options, then ask or mark the task blocked instead of adding more stubs, flags, priority rules, or tests around a broken assumption.
6. For board reads, use compact task-specific projections. A concrete assignment does not need routine `summary()`, `plan()`, `schema()`, or `{ full }`; request scoped schema only after an unknown call.

### Evidence Honesty Contract

1. Run each validation or gate command directly as a standalone process. Do not pipe it through `tee`; do not use a pipe chain unless `pipefail` is enabled and the gate command's real status is preserved.
2. Report the real exit code of every validation or gate command.
3. Report expected-red gates truthfully as failing: when a command is expected to fail (for example, `go test` in a package-less module), give its real non-zero exit code and a one-line expected-failure rationale; never present it as passing.
4. Check a checklist item tied to a command only after that exact command has actually run green with exit code 0. If it did not run or did not exit 0, leave the item unchecked.

## Status Transitions

- **start_status:** `development`
- **end_status:** `to-review` (review handoff, not accepted done)

## Constraints

Full read/write access does not authorize forced-fit workarounds. Tests and stubs may verify a valid design, but they must not be used to make an invalid product/API model appear acceptable.

## Skills

These role skill references are a lazy catalog, not a mandate to bulk-read every
body. Before technical work, identify the skills relevant to this task's concrete
scope and read those full skill bodies. Always read any skill explicitly required
by the task, user, or project instructions.

- **project-management**: `/Users/iv/.claude/skills/project-management/SKILL.md`
- **swiftui**: `/Users/iv/.claude/skills/swiftui/SKILL.md`
- **core-data**: `/Users/iv/.claude/skills/core-data/SKILL.md`
- **go-testing-tools**: `/Users/iv/.claude/skills/go-testing-tools/SKILL.md`
- **architecture-diagrams**: `/Users/iv/.claude/skills/architecture-diagrams/SKILL.md`

## Definition of Done

- [ ] Complete package graph and source inputs pass all rejection clusters
- [ ] Only the fixed go list and go build commands can execute
- [ ] Built outputs are verified but never launched during installation
- [ ] Readonly source, network denial, and supported host resource controls are executable platform gates
- [ ] Implementation matches AC
- [ ] Solution fits project architecture
- [ ] Tests green
- [ ] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches
- [ ] Board size is proportional to the spec and is the smallest decomposition that maps every requirement
- [ ] Every story and task traces to a concrete spec requirement; justified-gap elements also carry a self-verified gap record
- [ ] Beyond-literal-spec elements include a written justification naming the gap and the spec and out-of-scope checks performed before creation
- [ ] Research tasks cite an exact question the spec genuinely leaves open
- [ ] Dependencies linked
- [ ] Tasks are atomic — one clear deliverable each
- [ ] Completeness verified — nothing forgotten
- [ ] Any planning artifacts actually produced are linked as new task-scoped outcome resources; diagrams are strictly optional, never a standing deliverable
- [ ] Implement the identity-verified hidden manager worker, nonce-bound list/validate/permit/build state machine and exact package-independent entrypoint without package-controlled process inputs.
- [ ] Implement the exact mandatory portable controls and pre-worker failure boundary while ensuring unavailable inventory controls and all six deferred hardened guarantees never reject or create a containment claim.
- [ ] Implement rc5-native-control-inventory-v1 and closed capability-evidence-v1 for macOS and Windows with per-operation probes, stable contradiction errors and result-only exposure.
- [ ] Enforce fixed offline Go configuration, manager-selected graph, private staging, deadline/output/artifact bounds, worker-domain teardown, and pre/post worker/source/toolchain identity checks without ever launching the artifact.
- [ ] Pass focused adversarial and real-toolchain tests, full Go test/race/vet/build gates, native macOS validation and native Windows validation through ssh win, with Linux runtime honestly deferred.
- [ ] Attach revised task-scoped outcome with exact conformance pin, provenance, files, native capability evidence, commands/results and no-stage/no-commit proof, then hand off to a fresh independent reviewer.
- [ ] Code written per task description and AC
- [ ] Relevant tests written for new or changed behavior and passing
- [ ] Lint clean
- [ ] Relevant build/validation commands run after changes and build not broken
- [ ] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [ ] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
## Your Task

- **ID**: TASK-260720-1zntv0
- **Title**: implement-portable-go-v1-preflight-and-build
- **Parent**: STORY-260720-3plyvy
### Description

Implement the source-aware portable go-v1 pipeline that validates the complete active package graph with fixed manager-owned Go commands, builds into private staging, verifies the artifact and never transfers control to package code. Full fail-closed host sandbox guarantees are explicitly deferred to STORY-260728-327soo and do not block this task.
### Scope

Continue in internal/godriver using the trusted session and frozen build source. Use an identity-verified manager-owned worker boundary where needed; set CWD to canonical source_dir; pass only manager-defined argv and a fixed offline environment; mark snapshot/vendor inputs read-only where supported; write only manager-created operation-private Go cache, temp, telemetry and output roots by construction; bound time, captured output and artifact size; apply every reviewed native process-tree, memory, filesystem and network control available on the host; verify source/toolchain identities after execution; and report capabilities honestly. Parse the complete fixed go list stream, validate every package input, reject cgo/native/workspace/PGO/generator/external-link/input escapes, build with the normative internal-link argv and never execute the artifact. Do not publish caches, write markers or mutate live installs. The portable profile must not claim the six hardened guarantees tracked by STORY-260728-327soo and must not fail solely because a host lacks those hardened primitives.
### Acceptance Criteria

Valid standard-library and correctly vendored transitive fixtures build from the frozen snapshot with exact manager-owned argv/environment and verified artifact metadata. The portable profile proves package-influence exclusions, offline Go/module/VCS configuration, private manager-owned staging, bounded deadline/output/artifact size, pre/post source and toolchain identity checks, and every available reviewed native process-tree, memory, filesystem and network control. Diagnostics and evidence distinguish applied controls from unavailable hardened capabilities. Missing kernel-level total network denial, immutable read-only presentation, private-root-only write confinement, aggregate descendant process/memory/disk ceilings or exact descendant executable allowlisting does not reject an otherwise valid portable build and is never represented as a hardened claim. All package-controlled executable/argv/environment/output inputs, cgo/native/workspace/download/external-link/generator/PGO/input escapes, detected source or toolchain mutation, links, oversized output/artifacts, nonzero exits and built-artifact execution are rejected. Focused, full, race and vet tests pass; native macOS and Windows behavior is validated where available, with Linux compile/mock coverage until a Linux host is scheduled.

## Instructions

The following instructions have been attached to this task:

### portable-host-execution-accepted-spec.md
> Independent acceptance of portable manager-worker-v1 contract and exact candidate digest.

# TASK-260728-zb2s4z independent review cycle 2

## Verdict

**ACCEPTED.** Review findings R1 and R2 are closed. The amended rc.5 candidate
matches the task acceptance criteria and the curator-spec architecture. No
remaining product or specification finding requires rework.

## R1 — portable-versus-hardened boundary

The rework removes the contradictory implication that portable execution
provides kernel read-only presentation, total network denial, or an exact
executable allowlist.

- `protocol/core.md` sections 4.2 and 4.2.1, `profiles/manager.md` sections 2.2
  and 2.2.1, `SECURITY.md`, decisions 0004/0006, the author guide, CLI,
  compatibility text, changelog, release text, and the executable vector now
  consistently describe manager-selected mechanisms.
- `network: "none"` has one meaning: fixed offline Go module, proxy,
  checksum-database, and VCS configuration with no manager- or Go-initiated
  dependency/build networking. It explicitly does not claim kernel network
  containment.
- Frozen source/toolchain integrity is enforced by manager non-write rules and
  pre/post identity/currentness checks, not by claiming read-only presentation
  to descendants.
- The process graph is the fixed manager-selected four-node graph with
  per-program identity verification. Package bytes cannot select or add an
  executable, argv, environment value, path, flag, hook, plugin, generator, or
  build permit. The stronger kernel executable-path allowlist remains deferred.
- The failure boundary is identical in the normative and executable surfaces:
  an unavailable mandatory portable control rejects with
  `build_execution_control_unavailable` before worker launch and publishes
  nothing; an inventory control normatively unavailable for the platform, or
  any of the six deferred hardened guarantees, never rejects, warns, or blocks
  portable publication.

Exhaustive searches found no residual absolute “source is read-only to
children,” total-network-denial, or kernel executable-allowlist implication.
The remaining “start no program other than” wording is expressly scoped to
manager/worker selection and immediately distinguished from kernel allowlisting.

## R2 — native-control inventory and capability evidence

`rc5-native-control-inventory-v1` is an exhaustive, versioned authority over
exactly macOS and Windows and exactly five controls. Every control has a closed
per-platform `{availability, mechanism, unavailable_reason}` record. Inventory
membership, availability states, the one unavailable-reason vocabulary,
per-operation probe scope, and pre-worker-launch timing are pinned independently
by both Python and Go validators.

`capability-evidence-v1` is closed:

- record fields are exactly `record_version`, `execution_policy`, `platform`,
  and `controls`;
- entry fields are exactly `name`, `availability`, `status`, and `probed_at`;
- availability, status, timing, cardinality, exposure, exclusion, and
  consistency-rule vocabularies are closed;
- generated macOS and Windows examples are cross-checked against every
  per-platform inventory record;
- unknown, missing, duplicate, contradictory, wrong-version, wrong-policy, and
  deferred-guarantee evidence paths have stable errors and executable negative
  guards; and
- reporting remains result-only in dry-run plan, install, and status results,
  and is excluded from cache keys, receipts, markers, and claims. The portable
  execution-policy identity itself remains bound into all four identity
  surfaces.

The per-platform mechanisms are honest native primitives, not qualification
claims. Apple documents `RLIMIT_FSIZE` as a per-file limit inherited by created
processes, `killpg` as process-group signalling, and `setsid` as session/process
group creation. Microsoft documents Job Object active-process and job-wide
memory limits, kill-on-close termination, and explicit inherited-handle lists.
Those primitives support the inventory entries but do not imply any of the six
deferred hardened guarantees:

- https://developer.apple.com/library/archive/documentation/System/Conceptual/ManPages_iPhoneOS/man2/setrlimit.2.html
- https://developer.apple.com/library/archive/documentation/System/Conceptual/ManPages_iPhoneOS/man2/killpg.2.html
- https://developer.apple.com/library/archive/documentation/System/Conceptual/ManPages_iPhoneOS/man2/setsid.2.html
- https://learn.microsoft.com/en-us/windows/win32/api/winnt/ns-winnt-jobobject_basic_limit_information
- https://learn.microsoft.com/en-us/windows/win32/procthread/job-objects
- https://learn.microsoft.com/en-us/windows/win32/api/processthreadsapi/nf-processthreadsapi-updateprocthreadattribute

Candidate claim-v3 tuples remain empty. macOS and Windows remain pending
downstream native evidence, Linux remains excluded until
`TASK-260728-1skseh`, and `committed_release_pin_advanced` remains false.

## Identity, compatibility, and provenance

- Assigned and predecessor worktrees remain detached at
  `57c1f56846d221ecc55786bd3c2467ec32f11730`, with zero commits after the pin
  and clean indexes.
- Accepted predecessor manifest:
  `sha256:33fd7aed900ec4f9f9c72be6298823a16674116b18e3b8efdd2d147574dba2b8`.
- Accepted downstream candidate manifest:
  `sha256:58f8d2299c8f4a5ed78546913e567f637f1cc905dc5212b460f4097be7ff2af9`.
- Release metadata SHA-256:
  `b163f445f206c17dc618cc10b3957ca2f6f1b28607288162c0cfc5de02d83ee6`.
- Host-execution vector SHA-256:
  `c3d42f763afdcfa229430e7de5bb9f1e9f44607a7790aef6f4e0bf6d1bc644de`.
- The manifest has 422 unique, complete, non-self-referential entries and all
  independently recomputed file hashes match.
- Exact delta against the accepted predecessor is **98 modified, 13 added, and
  0 removed files**. The producer outcome's “15 added” count was stale after
  its documented removal of two scratch artifacts; this reviewer artifact
  records the corrected independently recomputed count. The 13 added files are
  ten execution-policy negative schema cases, the host-execution vector,
  decision 0006, and the portable policy author guide.
- Manifest schemas 1-7, receipt v1/v2, marker v1/v2/v3, claim v1/v2, and
  `curator-build-v1` promised protected surfaces compare byte-for-byte with the
  accepted predecessor. Under `schemas/v1`, only `README.md`,
  `common.schema.json`, and `conformance-claim-v3.schema.json` differ as
  intended.
- External acquisition/lifecycle, pack/index, and claim-v3 qualification
  artifacts compare byte-for-byte. Audit-before-cache/compiler ordering,
  offline errors, operator-owned signing, and empty candidate claims remain
  intact.

Independent CCJ-1 recomputation matched every policy-separation oracle:

- portable cache key:
  `sha256:529370122ae11e2e961d5265b1a020e046bcd43165b2eb96b05e73a51187ac9b`;
- reserved hardened key:
  `sha256:13736230d33ce59de7f7323dcd4cffd510655ad8dabd5ee9e8b6cb182ec70037`;
- pre-revision key:
  `sha256:3fcd714a40e8918eb67dbd35d435875dcce6c9047da811a1fa26626e5e57be48`;
- receipt-v2 cache key:
  `sha256:07dd911a7edc29b906a021aa6e1449632ce91c2e5a3eb0ea4f851cb84fe5c492`;
- full receipt-v2 hash:
  `sha256:11d2bf4df52638ef353b3286c426261eac2a73b0b64a32f85d78c04490072cea`.

Portable, reserved-hardened, and pre-revision identities are distinct. Receipt
and marker values agree, and frozen marker v2 remains transitively bound through
its cache key and receipt hash.

## Independent gates

Assigned worktree, read-only:

- `python -B tools/validate.py`: 42 schemas and 422 vector files, exit 0;
- Python unit suite: 22 tests, exit 0;
- focused dishonest-evidence mutation suite: exit 0;
- `go test ./tools/...`: exit 0;
- focused portable-policy and identity-binding Go tests: exit 0;
- `go vet ./tools/...`: exit 0;
- gofmt check: exit 0;
- `git diff --check`, clean index, and zero-commit checks: exit 0.

The system `python3` preflight lacked the pinned `jsonschema` dependency. The
Python gates were therefore rerun with the existing validation environment
containing the exact `requirements-dev.txt` dependency
`jsonschema==4.25.1`; they passed. This was an environment/tool-entrypoint
issue, not a candidate failure.

Disposable byte-identical clean probe:

- two consecutive `make regenerate-check` runs: exit 0;
- `make release-check VERSION=1.0.0-rc.5`: exit 0;
- Python compileall and Go generator build: exit 0;
- post-gate Git status: clean;
- post-gate recursive comparison with the assigned candidate: byte-identical.

The probe's local synthetic commit was used only to exercise the clean-checkout
release gate. No candidate or predecessor file was staged or committed, and no
ref, tag, release, downstream pin, platform claim, remote, or publication was
created.

## Reviewer boundary

No candidate, predecessor, schema, vector, tool, release, or product-code file
was modified during review.



### portable-host-execution-rework-evidence.md
> Implementation-facing native inventory, evidence record, vectors and gate details.

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





## Board CLI Reference

You have access to `task-board` CLI for managing your work. All writes use the `m` (mutation DSL) subcommand:

```bash
# FIRST/LAST lifecycle commands:
# Use explicit status changes when the task flow requires it.
task-board m 'set_status(TASK-260720-1zntv0, status=analysis)'       # analyst-style work
task-board m 'set_status(TASK-260720-1zntv0, status=development)'    # implementation / testing work
task-board m 'set_status(TASK-260720-1zntv0, status=reviewing)'      # reviewer handoff
task-board m 'set_status(TASK-260720-1zntv0, status=blocked)'        # when blocked
task-board m 'set_status(TASK-260720-1zntv0, status=to-review)'      # when your work is ready for review

# Track progress with checklist
task-board m 'check_item(TASK-260720-1zntv0, item=1)'                        # check item N
task-board m 'add_checklist_item(TASK-260720-1zntv0, text="Write tests")'    # add checklist item

# Add notes about your progress, decisions, blockers, or review findings
task-board m 'set_notes(TASK-260720-1zntv0, text="your note here")'

# Save short text directly as outcome resources
task-board m 'add_resource(TASK-260720-1zntv0, name=TASK-260720-1zntv0_results.md, content="...", type=outcome, description="Description")'

# Attach an existing local file (screenshots, PDFs, logs, archives, research docs)
task-board resource add TASK-260720-1zntv0 ./path/to/file --type outcome --name TASK-260720-1zntv0_artifact.bin -d "Description"
```

## Spawn Run Control

Tracked background spawn runs expose `TASK_BOARD_RUN_ID` in the child environment.
If your work is long-running, check for operator directives at safe checkpoints:

```bash
task-board spawn status "$TASK_BOARD_RUN_ID"
task-board spawn directives "$TASK_BOARD_RUN_ID"
```

Current runtimes do not support direct inbound push into your active session.
Treat directives as cooperative checkpoint signals:
- persist your current notes/artifacts before acting on `cancel`-style requests
- only honor pause/reroute intent at a safe checkpoint
- if no directive is present, continue normally

## IMPORTANT: Saving Results

When you produce work products (research documents, design docs, screenshots, logs, archives, implementation notes), you MUST save them as outcome resources with names that include the task ID:

```bash
task-board m 'add_resource(TASK-260720-1zntv0, name=TASK-260720-1zntv0_results.md, content="...", type=outcome, description="Description")'
task-board resource add TASK-260720-1zntv0 ./path/to/file --type outcome --name TASK-260720-1zntv0_artifact.bin -d "Description"
```

If you revise the same artifact later, use `task-board m 'update_resource(...)'` or `task-board resource update ...` instead of creating a silent overwrite.

If you discover important findings, decisions, anomalies, regressions, or non-obvious constraints while working, record them in `logbook` as well as on the board.

This ensures your results persist on the board and are accessible to other agents and the coordinator. Spawn completion is expected to produce at least one new task-scoped outcome artifact before the task can cleanly remain in `to-review`.

## Stop-The-Line: No Forced Fits

Do not keep implementing when autonomous work starts requiring a forced fit. A forced fit is any path where the task conflicts with a platform/API constraint, product decision, UX state model, ownership boundary, or architecture, and the remaining "solution" is mostly compensating hacks.

Warning signs:
- each fix needs another flag, stub, priority rule, mock-only behavior, or special-case test
- the tests can pass only because the test harness avoids the real platform behavior
- the implementation depends on an assumption you can no longer defend
- the user-facing behavior cannot be described cleanly without contradicting the product model

When this happens, stop product-code changes before adding another workaround layer. Attach or note:
- the constraint and evidence
- the failed assumptions/attempts
- the viable options and tradeoffs
- the recommended option
- the exact human/product/architecture decision needed

Then set the board item to `blocked` and ask only for that exact decision or external input. This stop applies only to a concrete external blocker or an unresolved human-only platform/product/architecture/tradeoff/approval decision; recoverable failures and ordinary rework stay autonomous. Tests and stubs are not proof that a forced-fit design is correct; use them only after the state model and platform assumptions are valid.

## Completion Discipline

Keep working until the task reaches a terminal handoff for your role. If no objective blocker remains, do not stop while the board item is still parked in `analysis`, `development`, `testing`, or `reviewing`.

Before your final status change:
- satisfy the task acceptance criteria and relevant checklist items
- attach outcome evidence for the work you produced
- run the relevant verification commands when the task changes code, tests, docs, or config

Use `blocked` only for either a concrete external blocker you cannot resolve autonomously or an unresolved human-only platform/product/architecture/tradeoff/approval decision. Record the constraint, evidence, failed assumptions/attempts, viable alternatives and tradeoffs, recommendation, and exact human decision or external input needed. Recoverable failures and ordinary rework are not `blocked`.

Status language is literal:
- `to-review` means your role has handed work to review; it does not mean the board task is accepted or done.
- In your final response, say "ready for review" or "handed off to review" when the final board status is `to-review`.
- Do not say "done", "complete", "finished", "final", or "готово" as the overall task state unless the board status is actually `done`.

## LAST — Run For Role Handoff

When you have completed all role work and the task is ready for its role handoff, run this as your **final board command**:

```bash
task-board handoff TASK-260720-1zntv0 --role developer
```

## Working Directory
Board directory: `/Users/iv/Developer/ReluxWorks/curator/.task-board`

Work in the project root. Do not modify board files directly — always use the `task-board` CLI.
