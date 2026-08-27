# Agent Task Assignment

## FIRST — Run Immediately

```bash
task-board m 'set_status(TASK-260728-zb2s4z, status=reviewing)'
```

## Your Role
# reviewer

## Description

Reviews how a task was implemented and how the solution fits into the project. Does not modify code; records one of the explicit verdict branches below.

When the run is goal-bound, query `task-board spawn goal "$TASK_BOARD_RUN_ID"` before recording the verdict. The reviewer goal is role-derived as `reviewer_verdict/reviewer_verdict`, carries its immutable parent goal ID/revision, and is satisfied only by exactly one verdict branch with evidence. A provider exit or `reviewing` status is not a verdict.
The runner persists the branch from the accepted board status plus a new or updated task-scoped verdict artifact. Only persisted `accepted` can satisfy the parent delivery goal; `changes_requested` and `stop_the_line` finish the reviewer goal without accepting delivery.

## Deliverable

Verdict branches are explicit:

- accepted → `done`
- changes requested → `to-dev` for implementation rework or `analysis` for research/decision work, with verdict evidence for the next producer and another reviewer cycle
- genuine stop-the-line boundary → `blocked` only for a concrete external blocker or an unresolved human-only platform/product/architecture/tradeoff/approval decision, with evidence, failed assumptions/attempts, viable alternatives and tradeoffs, a recommendation, and the exact human decision or external input needed

Do not leave the task in `reviewing`, and do not use `blocked` for ordinary rework or a recoverable child/runtime failure.

For an enforced Bug/Story `done` transition, a reviewer-archetype run must not supply `commit_ack`. Record acceptance evidence and hand it to the commit-owning mover; after committing its scope, that mover makes the final `done` transition with `commit_ack=scope_committed`.

For board reads, use compact task-specific projections. A concrete review does not need routine `summary()`, `plan()`, `schema()`, or `{ full }`; request scoped schema only after an unknown call.

## Status Transitions

- **start_status:** `reviewing`
- **end_status:** no unconditional default; the reviewer must set exactly one verdict status: `done`, `to-dev`, `analysis`, or evidence-backed `blocked`

## Constraints

Does NOT modify code. Read-only access.
- Reviewer-archetype runs must not supply `commit_ack`; record acceptance evidence for the commit-owning mover, which commits then makes the final `done` transition with `commit_ack=scope_committed`.

## Skills

These role skill references are a lazy catalog, not a mandate to bulk-read every
body. Before technical work, identify the skills relevant to this task's concrete
scope and read those full skill bodies. Always read any skill explicitly required
by the task, user, or project instructions.

- **project-management**: `/Users/iv/.claude/skills/project-management/SKILL.md`
- **architecture-diagrams**: `/Users/iv/.claude/skills/architecture-diagrams/SKILL.md`

## Definition of Done

- [ ] Create a fresh task-scoped curator-spec worktree from the pinned predecessor base and import the accepted rc.5 candidate without mutating predecessor worktrees, staging, committing or publishing.
- [ ] Amend normative core, manager profile, SECURITY and decision text to define portable manager-worker-v1 capabilities and explicitly defer all six hardened guarantees to STORY-260728-327soo.
- [ ] Prove the manager-owned worker and fixed Go process graph remain identity-verified and cannot be selected or influenced by package-controlled executable, argv, environment, paths, flags, hooks, plugins or generators.
- [ ] Bind execution-policy identity and capability evidence into cache, receipt, marker and claim semantics so portable and future hardened outputs cannot alias.
- [ ] Regenerate executable positive and negative vectors, compatibility guards and authoring guidance while preserving schemas 1-5 bytes, closed schema-6/7 declarations, audit-before-cache/compiler and empty candidate platform claims.
- [ ] Run all schema, Python, Go, vet/gofmt, deterministic double-regeneration, legacy compatibility and clean rc.5 release gates with no skipped or fabricated evidence.
- [ ] Attach a task-scoped outcome containing exact base/import provenance, changed artifacts, mandatory versus unavailable controls, commands/results and independently recomputable new candidate digest; hand off to review without self-acceptance.
- [ ] Code written per task description and AC
- [ ] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [ ] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
- [ ] Implementation matches AC
- [ ] Solution fits project architecture
- [ ] Tests green
- [ ] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches
## Your Task

- **ID**: TASK-260728-zb2s4z
- **Title**: TASK-260728-zb2s4z: amend-portable-go-v1-host-execution-contract
- **Parent**: STORY-260728-10wxx2
### Description

Amend the unreleased rc.5 candidate so go-v1 normatively supports the maximum autonomously enforceable portable manager-owned execution profile on macOS and Windows without claiming the separately tracked hardened guarantees.
### Scope

Start from the independently accepted TASK-260728-3b8qym candidate state. Update protocol core, manager profile, security/threat text, decision record, go-v1 driver policy, cache/receipt/marker/claim identity, conformance vectors, compatibility guards, authoring guidance and rc.5 release metadata. Permit one identity-verified manager-owned worker boundary with package influence excluded. Define exact portable mandatory controls and capability evidence; explicitly exclude full network/filesystem/resource/executable-allowlist guarantees and point to STORY-260728-327soo. Keep schemas 1-5 byte compatible, schema-6/7 declaration shapes closed, external audit-before-cache/compiler unchanged, candidate platform claims empty and all release provenance honest. No Curator or csk code.
### Acceptance Criteria

Normative text and executable vectors distinguish portable manager-worker-v1 from future hardened execution; package-controlled executable, argv, environment, output path, flags, hooks, plugins and generators remain impossible; fixed offline vendored Go behavior, private staging, bounded time/output/artifact, pre/post identity verification and available native controls are mandatory; unavailable hardened capabilities do not reject portable builds and cannot appear as hardened claims. Execution-policy identity is included in cache, receipt, marker and claim semantics so aliases are impossible. All schema, Python, Go, deterministic regeneration, compatibility and clean rc.5 release gates pass; a new exact downstream candidate digest is independently recomputed.

## Instructions

The following instructions have been attached to this task:

### accepted-rc5-review.md
> Independent acceptance evidence and exact predecessor candidate digest.

# TASK-260728-3b8qym independent review cycle 2

## Verdict

ACCEPTED. No remaining findings.

## Rework closure

1. CCJ-1 oracles: independent jq canonicalization recomputed SHA-256(CCJ-1(receipt.input)) as sha256:012564909df8f333004eb5aec867210d8973c74d9b71948e1f4fdb0d00c76559 and SHA-256(CCJ-1(full receipt)) as sha256:d86ca882ff336480a277010e886bea46767bf9a23d928fe3b43ee1009c3bd0ed. Receipt, mixed marker, mixed plan, schema examples, Python validation, and Go validation agree. Targeted false-cache-key and false-receipt-hash mutation tests passed in both harnesses.
2. Whole-snapshot ordering: executable external-source-dry-run and external-audit-only vectors are non-mutating, claim source and audit coverage, and enforce exact acquisition -> graph/object proof -> LFS scan -> immutable snapshot -> whole-snapshot validation -> source digest -> descriptor validation -> independent audit before cache or compiler. The same checker covers cache hit, cache miss, repair, dry-run, and audit-only; syntax-only claims none of source, audit, cache, or mutation.
3. Pack negatives: the checksum case materializes base plus final-index-byte XOR and differs at exactly byte offset 1071; full PACK and index-v2 header, fanout, trailers, filenames, embedded pack checksum, and index checksum parsing proves the intended sole fault. The family case preserves exact valid SHA-1 bytes under a SHA-256 declaration and fails SHA-256 validation with the stable build_repository_local_object_format_unsupported error.

## Release qualification

- Assigned HEAD is exactly 57c1f56846d221ecc55786bd3c2467ec32f11730; index is unstaged and there are zero commits after the pin. Content matches the disposable clean probe byte-for-byte. Reviewer regeneration changed generated-file mtimes only, not bytes.
- Two consecutive clean make regenerate-check runs passed. Clean make release-check VERSION=1.0.0-rc.5 passed at disposable probe commit baed6d17303344c8c48dfdbb9cc6f6681aab1e1d.
- Full validation passed: 42 schemas, 411 manifest files, 18 Python tests, all Go tool tests, go vet, gofmt, compileall, go build, git diff --check, and tracked probe cleanliness.
- Independent manifest audit found 411 unique, complete, non-self-referential entries with exact file hashes. Manifest pin is sha256:33fd7aed900ec4f9f9c72be6298823a16674116b18e3b8efdd2d147574dba2b8. Release metadata SHA-256 is 1f4f6fcc7c57f26f050d730f7b84091cf3f58970871bd7f511bea03f4b2b8f31.
- Exact file SHA-256 values: receipt 28d3340295731b4271ceb002ef3fc063bbf187237f4012e2269b0c702bf78622; mixed marker 0088cb9536eaacea2c087efeb70a5bc453c923020c5c01d3a87146f6e85773e3; pack/index 09768bac46eb9966e76baf7fbe613ed87a88dd665962f4d3c18f20dc6a98791c; lifecycle 175d709f183775b22ed3db27bc923ca78d6394d93a13096b5d6890c509aab072. Largest fixture is 18377 bytes, below the 65536-byte ceiling.
- Exact downstream protocol pin is the manifest pin above through CURATOR_CONFORMANCE_ROOT; committed_release_pin_advanced remains false and the obsolete sha256:30a64ed0... pin is absent.
- Candidate claim-v3 claims are empty; macOS and Windows remain pending downstream native evidence; Linux is explicitly excluded until TASK-260728-1skseh. No manager implementation, generic driver, real remote contact, fabricated platform evidence, or schema rollback entered scope.
- 196 protected schemas 1-6, go-v1, receipt-v1, marker-v1/v2, claim-v1/v2, and generated legacy cases match the accepted predecessor byte-for-byte.

The rc.5 shared conformance corpus and candidate release layer satisfy the task acceptance criteria and fit the specification architecture.


### host-execution-decision.md
> Architecture analysis of manager worker and platform enforcement boundaries.

# TASK-260720-1zntv0 — go-v1 host-execution decision packet

Date: 2026-07-27
Role: solution-architect
Status: decision packet for review; implementation remains authorization-gated

## Decision summary

Authorize the manager-owned worker contract, but initially support real `go-v1`
builds only on an explicitly capability-gated hardened Linux profile.

The worker is an identity-verified re-execution of the installed manager, not a
package-selected executable or a general command launcher. It provides the
trusted pre-exec boundary that the current direct `os/exec` path lacks. One
worker session executes exactly one fixed `go list` command, waits while the
parent validates the complete package graph, and then—only after an authenticated
fixed permit—executes exactly one fixed `go build` command. The built output is
verified and returned to the manager without being started.

This contract is implementable fail-closed on a sufficiently provisioned Linux
host. Current supported public primitives do not provide an equivalent,
production-stable profile on macOS or Windows while retaining the exact
executable graph and hard filesystem, network, descendant-process, memory,
process-count, disk, time, and output gates. Those hosts must reject `go-v1`
before starting the worker or Go until separately reviewed native profiles meet
the same contract.

Because `1.0.0-rc.4` has not been released or pinned, retain the `go-v1`
identifier and add an explicit cache/receipt policy revision such as:

```json
{
  "execution_boundary": "manager-worker-v1"
}
```

If the direct-Go contract is released first, the same semantic change must use
a new driver identifier or another explicitly versioned policy revision that
cannot collide with prior cache and receipt inputs.

## Exact human decision requested

> Authorize protocol 1.0.0-rc.4 to permit one identity-verified manager-owned
> sandbox worker in the fixed go-v1 process graph, retain the `go-v1` identifier
> with a new `manager-worker-v1` execution-policy cache revision, and ship real
> builds only on the fail-closed hardened Linux profile; macOS and Windows must
> reject go-v1 until independently reviewed native profiles satisfy the same
> executable, filesystem, network, and resource gates.

## Why the current contract cannot satisfy the requirement

The go-v1 requirements jointly demand:

- one fixed `go list` and one fixed `go build`;
- an empty fixed environment and canonical source CWD;
- validated source and vendored inputs mounted or presented read-only;
- writes confined to manager-private cache, temp, telemetry, and output roots;
- network, module, workspace, toolchain-switching, and VCS denial;
- an exact executable graph consisting only of trusted Go/toolchain programs;
- hard bounds for deadline, output, artifact/disk, memory, and process use where
  the supported host claims those controls;
- complete package-graph validation between list and build; and
- no execution of a package-selected program or built artifact.

The current `HostPolicy` validates policy metadata and then delegates directly
to the ordinary OS executor. Its platform hooks do not install a sandbox;
deadline and combined output are bounded, while artifact/disk checks are
post-exit and memory/process controls are not enforced. A direct Go
`os/exec` launch has no portable, arbitrary pre-exec hook in which the child can
install all required restrictions before Go code runs.

This is not a missing test or a local implementation defect. It is a missing
trusted execution boundary in the normative process graph.

## Justified gap record

**Missing piece.** An enforceable, manager-controlled pre-exec sandbox boundary
for source-aware `go list` and `go build`.

**Requirement otherwise incomplete.** The task scope and AC require read-only
source/toolchain inputs, operation-private writes, network denial, bounded host
resources, no unexpected executable, and rejection before publication. The
candidate contract expresses the same requirements in `protocol/core.md §4.2`,
`profiles/manager.md §§2.2–2.4`, and the compile-only/compiler-input sections of
`SECURITY.md`.

**Consequence if omitted.** The manager either runs compiler activity without
the advertised native controls or rejects every real build. Post-exit checks
cannot undo source mutation, network access, process escape, or resource
exhaustion.

**Proposed closure.** Add one fixed, identity-verified manager worker to the
normative graph. The worker installs the native sandbox before invoking Go; the
parent retains graph validation and publication authority.

**Self-verification performed before proposing the addition.** Checked the
candidate `protocol/core.md §4.2`, `profiles/manager.md §§2.2–2.4`,
`SECURITY.md`, decision 0004, the closed Go policy schema, accepted contract
research, and the documented rejected/out-of-scope alternatives. The candidate
currently requires the manager to invoke Go directly and therefore does not
already authorize this boundary. It excludes package-selected/general
executables, not a fixed manager-owned security boundary. The worker remains
outside package control, but the process-graph change is normative and therefore
requires explicit authorization and cache/receipt separation. No reviewed
section supports claiming the same fail-closed implementation on current macOS
or Windows.

## The two viable contracts

### A. Fixed manager-owned worker

Normative graph:

```text
manager parent
  └─ same installed manager, hidden worker mode
       └─ fingerprinted <GOROOT>/bin/go
            └─ fingerprinted tools below <GOROOT>/pkg/tool
```

The worker is an implementation boundary, not a new user-visible command
surface. In Curator it can be an exact self-re-execution. In csk the normative
term must remain “identity-verified manager-owned worker,” because a Python
entrypoint may make the interpreter and installed package tree part of the TCB
unless csk ships a standalone worker.

Advantages:

- gives the child a safe place to enter namespaces/sandbox controls before Go;
- retains a closed, non-package-selected process graph;
- allows a single authenticated list→validation→build session;
- separates sandbox authority from package parsing and publication authority;
- can fail before Go when the host profile is unavailable.

Costs:

- amends the fixed process graph and cache/receipt semantics;
- expands the TCB to the worker protocol and identity checks;
- requires native platform adapters and capability probes;
- currently yields a positive real-build profile only on hardened Linux;
- requires csk to define a trustworthy worker identity instead of assuming a
  mutable Python module launch is equivalent to one installed executable.

### B. Preserve direct Go and narrow supported hosts

This retains the candidate graph:

```text
manager
  └─ fingerprinted <GOROOT>/bin/go
       └─ fingerprinted tools below <GOROOT>/pkg/tool
```

It is viable only where the manager is launched inside a sufficiently prepared
external containment domain, or where a platform-specific direct-child
mechanism installs every restriction without adding a helper. For Go on Linux,
providing arbitrary setup before `execve` would require a custom native
clone/fork/exec trampoline or equivalent runtime-specific work rather than the
ordinary Go executor. Externally containing the whole manager also complicates
its access to protected install/cache state and can require a split
manager/broker architecture.

Advantages:

- preserves the already documented executable graph;
- avoids a worker protocol and worker identity as new TCB components.

Costs:

- much narrower deployability and difficult capability assumptions;
- substantially higher native/runtime implementation risk;
- poor portability to csk;
- external-manager containment changes operational packaging and may give the
  build domain more manager access than intended;
- still does not produce a reviewed current macOS or Windows profile.

### Recommendation

Choose A. The worker is the smaller auditable security boundary and the only
option that cleanly places restrictions before compiler execution while keeping
package inputs out of process selection. Ship it honestly: positive real builds
only on a capability-gated hardened Linux profile, and deterministic
`host_profile_unsupported` rejection elsewhere.

Option B should be selected only if preserving the direct-Go graph is more
important than practical in-manager execution. It would require an explicit
deployment contract describing the external containment owner and would still
need a fresh security review.

## Worker contract

### Identity and TCB

The parent resolves the current executable to a canonical, regular installed
file, rejects symlink/reparse/link substitution as appropriate, records strong
file identity, and hashes the bytes. The child proves the same executable
identity and hash before accepting any work. Identity is rechecked at the
launch boundary to close replacement races.

The TCB is:

- the installed manager parent and worker bytes;
- worker framing, authentication, and state machine;
- platform capability probe and sandbox adapter;
- OS kernel sandbox/namespace/job primitives;
- fingerprinted Go binary and GOROOT tools;
- source/build-root canonicalization, fingerprint, and input validation;
- operation-private roots and artifact verifier; and
- policy/cache/receipt canonicalization.

For csk, the Python interpreter/runtime and installed csk package tree are also
TCB components unless a standalone immutable worker is distributed. That
choice must be explicit in the csk implementation contract.

### Session and protocol

1. Parent performs package-independent toolchain probes and freezes the source.
2. Parent capability-probes the named native profile. Failure is terminal before
   the worker or Go starts.
3. Parent opens anonymous inherited pipes/handles and launches the exact manager
   executable in a fixed hidden worker mode. No package file, environment value,
   PATH lookup, shell, or user option selects the executable or mode.
4. Parent sends a length-bounded canonical request containing a fresh nonce,
   exact identity expectation, CWD, fixed empty environment, immutable roots,
   private writable roots, limits, exact Go/tool identities, and both permitted
   Go argument vectors.
5. Worker validates all containment and identity data, closes unrelated
   descriptors/handles, installs the native sandbox, and acknowledges the
   nonce plus applied-profile evidence.
6. Worker runs exactly the fixed `go list` command and returns bounded output
   and exit metadata. It cannot proceed to build autonomously.
7. Parent parses the complete stream and applies every graph/input rejection.
8. Parent sends one authenticated fixed build permit. Worker runs exactly the
   fixed `go build` command or tears down on any other message.
9. Worker returns one bounded regular artifact through a manager-controlled
   channel or designated output root. Parent applies manager permissions,
   verifies size/type/identity, hashes it, and exposes metadata.
10. Parent kills/joins the full worker domain and discards all private state.
    Neither worker nor parent starts the artifact.

The state machine admits no retry command, arbitrary argv, shell, generator,
test, run, tool, VCS, module download, or additional executable request.

### Package influence exclusions

Package-controlled bytes may affect only validated compiler input and the
resulting artifact. They cannot select or modify:

- manager or worker executable, hidden mode, or identity;
- Go/tool executable paths;
- list/build argv, environment, CWD, or build tags;
- sandbox profile, namespaces, job, entitlements, handles, or limits;
- permitted roots or write destinations;
- worker protocol messages or the parent’s build permit;
- graph-validation result, artifact verifier, cache key, receipt, or marker;
- publication, artifact execution, or any post-build command.

Generator comments and PGO paths remain inert input and must not cause a command
to run. The exact ASCII `//go:cgo_import_dynamic` token and all other documented
native/cgo/syso/non-standard assembly rejection cases remain mandatory.

## Platform enforcement matrix

| Gate | Hardened Linux profile | macOS current public profile | Windows current stable profile |
|---|---|---|---|
| Filesystem read/write/execute | User/mount namespaces plus read-only binds/bounded tmpfs; Landlock as a second restriction layer | App Sandbox is entitlement/signing and user-authorization based; deprecated dynamic sandbox API is not a stable contract | AppContainer ACL/capability model can constrain access, but setup/packaging and exact descendant executable policy remain unresolved |
| Network | Isolated network namespace; Landlock TCP/UDP where required ABI exists | App Sandbox can deny network for a packaged entitled app, but arbitrary source/GOROOT access conflicts with the current CLI contract | AppContainer can default-deny network |
| Process tree | Delegated cgroup v2 plus PID namespace; tree kill through cgroup | `setrlimit` is not a private descendant domain; process count is UID-scoped | Job Object provides descendant grouping, kill-on-close, active-process limits |
| Memory/process | cgroup `memory.max` and `pids.max` | RSS limit is not a hard aggregate tree limit | Job Object group/process memory and active-process limits |
| Disk/artifact | Bounded tmpfs private roots plus bounded artifact streaming | Per-file `RLIMIT_FSIZE` is not aggregate private storage | No Job Object hard aggregate storage quota for arbitrary private roots |
| Deadline/output | Parent deadline, cgroup tree kill, bounded pipes | Parent deadline/output possible, but missing other fail-closed gates | Job termination and bounded pipes possible |
| Exact descendants | Landlock execute allowlist/read-only mount graph plus fixed worker state machine | No reviewed stable dynamic path/executable allowlist for this CLI design | Child-process policy is all-or-none; allowing Go tools does not itself restrict their exact paths |
| Result | Supported only after every feature probe succeeds | Must fail closed | Must fail closed |

### Linux profile requirements

The profile name must identify capabilities, not merely “Linux” or “Ubuntu.”
Before starting the worker it must prove:

- required Landlock ABI and handled filesystem rights; network rights when used;
- unprivileged/delegated user, mount, network, and PID namespace availability;
- cgroup v2 delegation with memory and pids controllers and reliable tree kill;
- bounded tmpfs/private-root setup;
- no-new-privileges and any required syscall filter support;
- ability to mount/present frozen source, vendor tree, and GOROOT read-only;
- ability to allow execute only for the fixed manager, Go, and fingerprinted
  GOROOT tools; and
- a bounded artifact return path that never executes the artifact.

Any missing or administratively disabled primitive rejects the operation before
the worker/Go graph begins.

### macOS finding

The SDK documents the dynamic sandbox initialization API as deprecated, and its
named-profile interface does not provide the required stable dynamic path
contract. App Sandbox is a signed-entitlement model; arbitrary source access
generally requires user-selected authorization/security-scoped access, while a
child inherits static sandbox configuration. Resource limits do not form a hard
aggregate private worker domain for memory, process count, and storage.

Therefore exact manager re-execution is not a sufficient current fail-closed
macOS profile. A separately packaged/signed XPC/helper architecture or a future
reviewed public primitive would be a different contract, not an implementation
detail of this decision.

### Windows finding

Job Objects provide strong descendant grouping, termination, time, memory, and
active-process controls. AppContainer can default-deny filesystem/network
resources when ACLs/capabilities are set correctly. However the stable child
process policy is all-or-none rather than an exact permitted-descendant path
graph, and Job Objects do not supply a hard aggregate storage quota.

Microsoft now documents an experimental `CreateProcessInSandbox` family for
Windows 11, but it is explicitly experimental, dynamically obtained rather than
provided through a normal public header, and does not by itself document all
required resource and exact-descendant gates. It cannot be the production
rc.4 contract.

Therefore exact manager re-execution is not a sufficient current fail-closed
Windows profile.

## Required specification and artifact amendments

### curator-spec

1. **`protocol/core.md §4.2` — source-aware Go command**
   - Replace direct-Go-only wording with the exact
     parent→manager-worker→Go→GOROOT-tools graph.
   - Define worker identity, fixed hidden mode, authenticated bounded protocol,
     one-session list/validate/build state machine, and no package influence.
   - Keep the exact Go argv, fixed empty environment, canonical CWD, and
     package-graph rejections unchanged.
   - Keep package-independent probes direct and closed.
   - Require capability detection and fail-before-worker behavior.

2. **`protocol/core.md` cache/receipt sections**
   - Add `execution_boundary: manager-worker-v1` to the canonical Go build
     policy/input.
   - Recompute cache keys, build-receipt inputs/hashes, and all canonical
     preimages. Old candidate values must miss, never alias.

3. **`profiles/manager.md §§2.2–2.4`**
   - Specify the worker lifecycle, identity proof, list/build permit boundary,
     applied native-profile evidence, teardown, tree kill, and artifact return.
   - Add the supported profile table and deterministic unsupported-host error.
   - Forbid publication if capability proof or any control is absent.

4. **`SECURITY.md`**
   - Record the expanded TCB, package-influence exclusions, worker protocol
     threats, identity/replacement threats, and fail-closed platform support.
   - Preserve the statement that sandboxing does not make compiler input safe;
     it bounds the compile-only operation.

5. **Decision 0004**
   - Supersede the direct process-graph clause before rc.4 publication and
     record why the worker is a security boundary, not a command extension.

6. **`schemas/v1/common.schema.json`**
   - Add the new required const field to the closed Go policy object.
   - Build receipt input inherits the revised object. Marker v2 need not change
     structure if it already carries cache/receipt hashes, but its vectors and
     expected hashes must change.

7. **Conformance vectors and generation**
   - Regenerate indexes, manifests, expected canonical forms, cache/receipt
     hashes, and claim-suite digests.
   - Add positive and rejection vectors listed below.
   - Update compatibility, conformance README, changelog, and release evidence.

### Curator

- Replace the declarative `guardedHostPolicy`/direct OS-executor boundary with a
  worker client/session and a hidden fixed worker dispatcher.
- Add executable identity verification and race-resistant launch checks.
- Add native worker adapters and a named hardened-Linux capability probe.
- Keep one worker alive across list→parent validation→build; do not add a
  second list or build command.
- Return the artifact through bounded private staging/streaming; verify regular
  type, size, permissions, identity, and digest before exposing metadata.
- Add the policy revision to build metadata, canonical encoding, cache input,
  receipt input, and vector tests.
- Fail before worker/Go on unsupported profiles; publish no cache, marker,
  receipt, or artifact.
- Remove comments or interfaces that imply policy validation alone is native
  enforcement.

### csk

- Update the existing fixed go-v1 driver work item; do not add a parallel
  ceremonial driver.
- Do not implement source-aware Go with plain `subprocess.run`.
- Define the worker identity model explicitly:
  - either distribute a standalone immutable manager-owned worker; or
  - fingerprint and lock the Python interpreter/runtime plus installed csk
    package tree as the fixed TCB graph.
- Implement the same bounded protocol, state machine, policy serialization,
  Linux capability profile, artifact handling, and unsupported-host rejection.
- Consume the same regenerated canonical/cache/receipt vectors as Curator.

### Claims, pins, and release flow

No candidate, release, ref, pin, or provenance record should be edited until the
human decision is recorded and the normative spec/vectors land. After landing:

- the claim suite must name the new conformance digest and supported native
  profile evidence;
- Curator and csk pin the same immutable spec ref;
- cache and receipt vectors must prove old candidate policy inputs cannot
  collide; and
- release evidence must not claim positive macOS/Windows real-build support.

## Required tests and CI evidence

### Protocol/identity rejection

- worker path/hash/file-identity mismatch, symlink/reparse/hardlink substitution,
  and replacement race;
- malformed, oversized, replayed, wrong-nonce, out-of-order, or extra protocol
  message;
- mutation of argv, env, CWD, roots, limits, executable paths, or profile;
- build permit before successful complete list validation;
- any retry, extra executable, shell, VCS, module, generator, PGO, test, run, or
  tool request.

### Native enforcement

- source and GOROOT writes fail;
- outside-root reads and escaped embed/module/vendor inputs fail;
- TCP and UDP network attempts fail;
- an unexpected executable and surviving descendant fail;
- fork/process-count, memory, private-disk, deadline, combined-output, and
  artifact-size limits terminate the full domain;
- artifact symlink/link/device/non-regular/oversize cases fail;
- no built output is started.

### Functional/package graph

- valid standard-library and correctly vendored transitive fixtures;
- every existing non-main/multiple-root/module/vendor/workspace/toolchain/cgo/
  native/syso/assembly/embed/directive/poisoned-input rejection;
- exactly one fixed list and one fixed build appear in interaction evidence;
- artifact metadata matches the verified bytes and manager permissions.

### CI split

- **Hardened Linux native lane:** positive real-toolchain builds plus all native
  enforcement/rejection tests. The runner must prove every profile capability.
- **Generic Linux/macOS/Windows lanes:** unit, parser, canonical/vector, worker
  protocol, identity, and compile tests.
- **macOS/Windows native behavior:** deterministic
  `host_profile_unsupported` before worker or Go, with no publication.

## Minimal board decomposition

No new speculative story or duplicate quality/documentation task should be
created before authorization. The smallest traceable plan reuses existing
ownership:

| Existing board owner | Atomic deliverable after authorization | Requirement trace |
|---|---|---|
| curator-spec work under `STORY-260720-35dck7` (prior owners `TASK-260720-1nvomm`, `TASK-260720-17llva`, `TASK-260720-12iigs`, `TASK-260720-1s1vr6`, `TASK-260720-cw39jh`; integration gate `TASK-260720-3ag6pi`) | Land the authorized worker graph, policy revision, schema, vectors, and claim/release semantics as one coherent normative revision | `protocol/core §4.2`; cache/receipt sections; `profiles/manager §§2.2–2.4`; `SECURITY.md`; closed Go policy schema |
| `TASK-260720-1zntv0` | Implement Curator worker boundary, Linux profile, artifact verification, and unsupported-host rejection without changing the fixed Go commands | This task’s Scope and AC: native gates, exact graph, full list validation, one staged artifact, never execute output |
| `TASK-260720-2g21eg` | Implement the same authorized go-v1 contract and explicit manager-worker identity model in csk | go-v1 cross-manager conformance and the revised canonical policy/vectors |
| Existing downstream staging/install/claim tasks | Consume only a verified artifact and revised receipt/cache evidence; never execute it during installation | This task’s “no publication/no execution” AC and claim/pin protocol gates |

Dependencies already place staging behind the Curator build task. The normative
spec revision must also block both manager implementations and release claims.
Creating more tasks now would either duplicate these owners or presume the
unmade human contract decision.

## Completeness audit

- Every task requirement cluster is mapped: fixed commands/env/CWD; full graph
  parsing; path/fingerprint/native-input rejections; read-only source; private
  writes; network denial; executable graph; resource bounds; artifact
  verification; non-execution; tests and CI.
- Both and only both documented viable contracts were compared.
- The added boundary carries a self-verified justified-gap record.
- The open question is exact and human-owned: whether to authorize the worker
  graph and Linux-only initial positive profile.
- No diagram is required; the four-node process graph and state sequence are
  clearer in text.
- No product code, candidate, release, ref, pin, provenance, or implementation
  artifact is changed by this packet.

## Residual risks after the recommended decision

- Kernel or runner capability drift can turn a nominal Linux lane unsupported;
  every operation must probe, not trust CI labels.
- Manager binary replacement/identity races require platform-specific
  race-resistant handles and tests.
- Landlock and namespace rules must cover both read/write and execute behavior;
  one mechanism alone is not the whole profile.
- cgroup I/O throttling is not a storage quota; bounded private tmpfs is required
  for aggregate disk control.
- The csk Python TCB may be materially larger than Curator’s self-reexec TCB.
- A future macOS/Windows implementation may require packaging or helper choices
  that constitute another normative architecture decision.
- The worker bounds compiler behavior but does not establish that adversarial Go
  source is semantically safe; the package-input rejection policy remains
  mandatory.

## Primary references

Project sources:

- Candidate `curator-spec/protocol/core.md §4.2`
- Candidate `curator-spec/profiles/manager.md §§2.2–2.4`
- Candidate `curator-spec/SECURITY.md`
- Candidate `curator-spec/decisions/0004`
- Candidate `curator-spec/schemas/v1/common.schema.json`
- Accepted contract research:
  `TASK-260720-1nvomm_accepted-contract.md`
- Curator `internal/godriver/host_policy*.go`, `executor.go`, and `build.go`

Official platform references:

- Linux Landlock:
  <https://cdn.kernel.org/doc/html/latest/userspace-api/landlock.html>
- Linux cgroup v2:
  <https://docs.kernel.org/admin-guide/cgroup-v2.html>
- Linux tmpfs:
  <https://cdn.kernel.org/doc/html/latest/filesystems/tmpfs.html>
- Linux namespaces:
  <https://man7.org/linux/man-pages/man2/unshare.2.html>
- Go Linux process attributes:
  <https://go.dev/src/syscall/exec_linux.go>
- Apple App Sandbox:
  <https://developer.apple.com/documentation/security/protecting-user-data-with-app-sandbox>
- Apple App Sandbox file access:
  <https://developer.apple.com/documentation/security/accessing-files-from-the-macos-app-sandbox>
- Apple resource limits:
  <https://developer.apple.com/library/archive/documentation/System/Conceptual/ManPages_iPhoneOS/man2/setrlimit.2.html>
- Windows Job Objects:
  <https://learn.microsoft.com/en-us/windows/win32/procthread/job-objects>
- Windows AppContainer isolation:
  <https://learn.microsoft.com/en-us/windows/win32/secauthz/appcontainer-isolation>
- Windows process attribute list:
  <https://learn.microsoft.com/en-us/windows/win32/api/processthreadsapi/nf-processthreadsapi-updateprocthreadattribute>
- Experimental Windows CreateProcessInSandbox:
  <https://learn.microsoft.com/en-us/windows/win32/secauthz/createprocessinsandbox>



### host-execution-review.md
> Independent review of the prior strict host-execution decision.

# TASK-260720-1zntv0 — independent host-execution decision review

Date: 2026-07-27  
Role: reviewer  
Reviewed artifact: `TASK-260720-1zntv0_host-execution-decision.md`  
Reviewed SHA-256: `e0fb6218cb0bd573db496bca30346277f8f9623a40d2b42d2b0814ecec291978`

## Verdict

**Stop-the-line boundary confirmed.**

The architecture decision packet is accepted as a sound, decision-ready
analysis artifact. The implementation task is not accepted: its current direct
Go host policy still cannot satisfy the native isolation and resource-control
acceptance criteria. Route `TASK-260720-1zntv0` to `blocked` until the exact
human architecture decision below is recorded.

This is not ordinary implementation rework. The current accepted candidate
requires direct execution of the fingerprinted Go binary and forbids any other
process in the graph, while the missing filesystem, network, descendant, and
aggregate resource gates require a trusted pre-exec containment boundary. The
remaining choice changes the normative process graph, supported-host promise,
cache/receipt identity, conformance vectors, and both manager implementations.

## Independent findings

### Accepted contract and current implementation

- Candidate `protocol/core.md §4.2` and `profiles/manager.md §2.2` require the
  manager to invoke fingerprinted Go directly and allow only Go plus
  fingerprinted GOROOT tools to start.
- Candidate `SECURITY.md` requires rejection rather than approximation when the
  fixed environment, network denial, source/context separation, native target,
  or process graph cannot be enforced, and requires filesystem/network plus
  host-supported resource limits.
- Candidate `protocol/core.md` explicitly permits either a new driver identifier
  or an explicit versioned cache-key policy when sandbox-relevant behavior
  changes. There is no local `1.0.0-rc.4` tag/ref, so retaining `go-v1` with a
  mandatory `execution_boundary: manager-worker-v1` policy revision is
  consistent only if the normative candidate, schemas, vectors, receipts, and
  claims are regenerated before release.
- Current Curator `guardedHostPolicy` validates metadata and delegates directly
  to `OSExecutor`. Unix and Windows platform hooks return `nil`; memory and
  process limits are not enforced; disk/artifact bounds are post-exit checks.
  The earlier implementation rejection remains valid.

### Recommended manager-worker contract

The fixed manager-owned worker is a legitimate security boundary rather than a
package-selected command surface:

```text
manager parent
  -> identity-verified manager worker
      -> fingerprinted Go
          -> fingerprinted GOROOT tools
```

The packet correctly excludes package influence over worker identity/mode,
argv, environment, CWD, roots, limits, permits, publication, and artifact
execution. Its single-session state machine preserves exactly one list and one
build, with parent validation between them. It correctly keeps csk's Python
interpreter and installed package tree in the TCB unless csk distributes a
standalone immutable worker.

The implementation and normative text must bind worker launch to an already
verified executable identity (descriptor/handle-backed execution or an exact
platform equivalent). A path recheck plus a later hash is not sufficient to
close replacement races. The packet already makes race-resistant launch,
identity mismatch, substitution, and replacement-race tests mandatory; this is
an implementation acceptance gate, not a reason to reject the decision packet.

### Platform conclusion

- **Hardened Linux can be positive only by capability, never by OS label.**
  Landlock can restrict filesystem execute/read/write and propagates policy to
  descendants; current ABIs add TCP and UDP rules. Cgroup v2 supplies a hard
  `memory.max`, hard subtree `pids.max`, contained delegation, and tree-wide
  `cgroup.kill`. User/mount/network/PID namespaces and a bounded tmpfs can close
  the remaining filesystem, network, descendant, and aggregate-disk surfaces.
  Any absent ABI, controller, delegation, namespace, mount, or syscall-control
  prerequisite must reject before worker or Go.
- **macOS is not a current positive profile.** App Sandbox is a signed,
  entitlement-based packaging contract. Child inheritance carries static
  rights rather than post-launch PowerBox rights, and current Apple
  documentation says user-selected access cannot be used to execute programs
  outside the app bundle/container/group. The external fingerprinted GOROOT
  model therefore does not fit without bundling/packaging changes. `setrlimit`
  is per-process (and `RLIMIT_NPROC` is UID-scoped), not an aggregate private
  descendant domain; it does not close the required memory/process/disk model.
- **Windows is not a current stable positive profile.** Job Objects provide
  descendant containment, kill-on-close, accounting, memory/time and active
  process controls, while AppContainer provides default-deny file/network
  isolation. Stable child-process mitigation is all-or-none, not an exact
  executable allowlist, and Job Objects do not impose a hard aggregate private
  storage quota. Microsoft's newer CreateProcessInSandbox API is explicitly
  experimental, Windows-11-only, dynamically resolved with no public header,
  and cannot be the rc.4 production contract.

The packet is therefore right to specify deterministic
`host_profile_unsupported` before worker/Go on macOS and Windows, while keeping
their parser/protocol/vector/compile lanes.

## Alternatives and tradeoffs

1. **Authorize the manager worker and initially support hardened Linux.**
   This is the recommended option. It adds a small auditable TCB component,
   enables pre-exec enforcement, and preserves package non-influence, at the
   cost of a normative graph/policy revision and Linux-only positive native
   builds.
2. **Preserve direct Go and narrow deployment to externally contained or custom
   native-launch profiles.** This keeps the existing graph but shifts
   containment ownership outside the manager or requires a higher-risk native
   process subsystem. It remains difficult to port to csk and still provides no
   reviewed positive macOS/Windows profile.

Environment-only denial, permission-bit checks, mock-only evidence, and
post-exit scans are not viable alternatives because they cannot undo a source
write, network connection, escaped process, or resource exhaustion.

## Board and provenance integrity

- The reviewed resource is attached to `TASK-260720-1zntv0`, has a task-scoped
  name, and its content hash is recorded above.
- The canonical Curator checkout has no tracked or staged changes.
- The task implementation worktree files were last changed on 2026-07-21; the
  candidate specification worktree files were last changed on 2026-07-20/21;
  the decision artifact was created on 2026-07-27. No product, candidate,
  release, claim, pin, ref, or provenance file changed during this analysis.
- Existing owner `TASK-260720-2g21eg` covers the csk driver; the cited
  curator-spec owners and downstream staging dependency exist. No speculative
  duplicate task is needed before authorization.
- Global `task-board validate` still reports 45 pre-existing board issues
  (legacy `EPIC-260712` broken links/status mismatches, one unsupported
  container link, and two orphan resources). None is introduced by or local to
  this task-scoped decision resource, but the board is not globally clean.

## Exact human decision required

> Authorize protocol 1.0.0-rc.4 to permit one identity-verified manager-owned
> sandbox worker in the fixed go-v1 process graph, retain the `go-v1` identifier
> with a new `manager-worker-v1` execution-policy cache revision, and ship real
> builds only on the fail-closed hardened Linux profile; macOS and Windows must
> reject go-v1 until independently reviewed native profiles satisfy the same
> executable, filesystem, network, and resource gates.

## Primary evidence

- Linux Landlock: https://cdn.kernel.org/doc/html/latest/userspace-api/landlock.html
- Linux cgroup v2: https://docs.kernel.org/admin-guide/cgroup-v2.html
- Apple App Sandbox: https://developer.apple.com/documentation/security/protecting-user-data-with-app-sandbox
- Apple sandbox file access: https://developer.apple.com/documentation/security/accessing-files-from-the-macos-app-sandbox
- Apple `setrlimit(2)`: https://developer.apple.com/library/archive/documentation/System/Conceptual/ManPages_iPhoneOS/man2/setrlimit.2.html
- Windows Job Objects: https://learn.microsoft.com/en-us/windows/win32/procthread/job-objects
- Windows AppContainer isolation: https://learn.microsoft.com/en-us/windows/win32/secauthz/appcontainer-isolation
- Experimental Windows sandbox creation: https://learn.microsoft.com/en-us/windows/win32/secauthz/createprocessinsandbox





## Board CLI Reference

You have access to `task-board` CLI for managing your work. All writes use the `m` (mutation DSL) subcommand:

```bash
# FIRST/LAST lifecycle commands:
# Use explicit status changes when the task flow requires it.
task-board m 'set_status(TASK-260728-zb2s4z, status=analysis)'       # analyst-style work
task-board m 'set_status(TASK-260728-zb2s4z, status=development)'    # implementation / testing work
task-board m 'set_status(TASK-260728-zb2s4z, status=reviewing)'      # reviewer handoff
task-board m 'set_status(TASK-260728-zb2s4z, status=blocked)'        # when blocked
task-board m 'set_status(TASK-260728-zb2s4z, status=to-review)'      # when your work is ready for review

# Track progress with checklist
task-board m 'check_item(TASK-260728-zb2s4z, item=1)'                        # check item N
task-board m 'add_checklist_item(TASK-260728-zb2s4z, text="Write tests")'    # add checklist item

# Add notes about your progress, decisions, blockers, or review findings
task-board m 'set_notes(TASK-260728-zb2s4z, text="your note here")'

# Save short text directly as outcome resources
task-board m 'add_resource(TASK-260728-zb2s4z, name=TASK-260728-zb2s4z_results.md, content="...", type=outcome, description="Description")'

# Attach an existing local file (screenshots, PDFs, logs, archives, research docs)
task-board resource add TASK-260728-zb2s4z ./path/to/file --type outcome --name TASK-260728-zb2s4z_artifact.bin -d "Description"
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
task-board m 'add_resource(TASK-260728-zb2s4z, name=TASK-260728-zb2s4z_results.md, content="...", type=outcome, description="Description")'
task-board resource add TASK-260728-zb2s4z ./path/to/file --type outcome --name TASK-260728-zb2s4z_artifact.bin -d "Description"
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

## Working Directory
Board directory: `/Users/iv/Developer/ReluxWorks/curator/.task-board`

Work in the project root. Do not modify board files directly — always use the `task-board` CLI.
