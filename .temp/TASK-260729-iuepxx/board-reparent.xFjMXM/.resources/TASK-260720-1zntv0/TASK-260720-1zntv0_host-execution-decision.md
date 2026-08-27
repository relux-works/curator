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
