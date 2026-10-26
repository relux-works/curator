# Profile skill runtime

Profiles must install their skills' active commands. Slice A connects profile
requirements to the ordinary skill installer. Slice B gives each profile its
own command root. This note proposes a design for `TASK-261026-6b5ioc`.
It authorizes no implementation, installation, or enrollment.

## Contract and implementation gap

The protocol already defines the dependency mechanism. At curator-spec
[`6c7ce40301e254e14e55ebd29405c90fe00c79c6`](https://github.com/relux-works/curator-spec/tree/6c7ce40301e254e14e55ebd29405c90fe00c79c6),
environments §9.4 requires profiles to use the existing skill closure, audit,
build, and runtime machinery. The skill set includes profile requirements
and machine direct declarations. See
[`protocol/environments.md:3088-3106`](https://github.com/relux-works/curator-spec/blob/6c7ce40301e254e14e55ebd29405c90fe00c79c6/protocol/environments.md#L3088-L3106).
No new context-manifest dependency family is needed.

Skill requirements retain core §4.4 activation semantics. Manifest schema 2
also permits a skill package directory. The selected directory must stay
inside the snapshot without traversing links. See
[`protocol/environments.md:543-568`](https://github.com/relux-works/curator-spec/blob/6c7ce40301e254e14e55ebd29405c90fe00c79c6/protocol/environments.md#L543-L568)
and [`protocol/core.md:1178-1183`](https://github.com/relux-works/curator-spec/blob/6c7ce40301e254e14e55ebd29405c90fe00c79c6/protocol/core.md#L1178-L1183).

| Requirement mode | Context projection | Exported script/build commands |
| --- | --- | --- |
| `full`, the default | Active | All active |
| `context` | Active | None from this edge |
| `runtime` | Inactive from this edge | All, or the nonempty `commands` selection |

Activation combines incoming edges. A `context` edge and a narrowed `runtime`
edge activate both surfaces with the selected commands. A separate `full`
edge activates every exported command. Narrowing limits its own edge; it
does not revoke another consumer's requirement. System commands remain
executable prerequisites checked through the ordinary pipeline.

The following evidence uses Curator main at
[`f52b8bdd05d0705632b0142d31580e015aba2369`](https://github.com/relux-works/curator/tree/f52b8bdd05d0705632b0142d31580e015aba2369).
These are source findings. No profile installation established runtime behavior.

| Seam | Current implementation and gap |
| --- | --- |
| Declaration parsing | [`internal/contextpkg/contextpkg.go:319-401`](https://github.com/relux-works/curator/blob/f52b8bdd05d0705632b0142d31580e015aba2369/internal/contextpkg/contextpkg.go#L319-L401) accepts `mode`, runtime `commands`, and schema-2 `directory`. |
| Resolver input | [`internal/envprofile/gitsource.go:893-909`](https://github.com/relux-works/curator/blob/f52b8bdd05d0705632b0142d31580e015aba2369/internal/envprofile/gitsource.go#L893-L909) carries those fields into context requirements. |
| Skill loading | [`internal/envprofile/gitsource.go:344-360`](https://github.com/relux-works/curator/blob/f52b8bdd05d0705632b0142d31580e015aba2369/internal/envprofile/gitsource.go#L344-L360) returns an empty resolver package for repository-root skills. Directory-selected skills expose command names, without their ordinary skill dependency graph. |
| Command validation | [`internal/contextresolve/contextresolve.go:1058-1069`](https://github.com/relux-works/curator/blob/f52b8bdd05d0705632b0142d31580e015aba2369/internal/contextresolve/contextresolve.go#L1058-L1069) checks requested names only when the loader supplied `SkillCommands`. It constructs no activation plan. |
| Lock | [`internal/contextlock/contextlock.go:43-63`](https://github.com/relux-works/curator/blob/f52b8bdd05d0705632b0142d31580e015aba2369/internal/contextlock/contextlock.go#L43-L63) records source, commit, directory, and requirers. It has no activation edges, modes, or command selections. |
| Install | [`internal/envprofile/envprofile.go:881-936`](https://github.com/relux-works/curator/blob/f52b8bdd05d0705632b0142d31580e015aba2369/internal/envprofile/envprofile.go#L881-L936) resolves, audits/stores, publishes records, then optionally activates. It never invokes the command build planner. |
| Audit/store | [`internal/envprofile/envprofile.go:1519-1541`](https://github.com/relux-works/curator/blob/f52b8bdd05d0705632b0142d31580e015aba2369/internal/envprofile/envprofile.go#L1519-L1541) strictly audits and stores package inputs. This step does not compile commands. |
| Skill surfaces | [`internal/envprofile/managed.go:279-303`](https://github.com/relux-works/curator/blob/f52b8bdd05d0705632b0142d31580e015aba2369/internal/envprofile/managed.go#L279-L303) projects every lock skill. [`internal/envprofile/skill_projection.go:32-66`](https://github.com/relux-works/curator/blob/f52b8bdd05d0705632b0142d31580e015aba2369/internal/envprofile/skill_projection.go#L32-L66) excludes runtime/build roots. Neither path applies runtime-only context suppression. |
| Ordinary installer | [`internal/install/global.go:88-101,149-157`](https://github.com/relux-works/curator/blob/f52b8bdd05d0705632b0142d31580e015aba2369/internal/install/global.go#L88-L157) builds its closure from the global `Skillfile.json`. Profile-only requirements do not enter it. |
| Global continuation | [`internal/envprofile/global.go:91-147`](https://github.com/relux-works/curator/blob/f52b8bdd05d0705632b0142d31580e015aba2369/internal/envprofile/global.go#L91-L147) separately resolves profile members and publishes/materializes them. It does not supply the missing executable closure. |
| Launch fragment | [`internal/envprofile/managed.go:2561-2579`](https://github.com/relux-works/curator/blob/f52b8bdd05d0705632b0142d31580e015aba2369/internal/envprofile/managed.go#L2561-L2579) emits no `path_prepend`. |

The missing bridge has two outputs. It must construct ordinary skill nodes
with their incoming activation edges. It must then publish commands and
context projections from the same admitted plan. Fixing only compilation
would retain runtime-only prompt leakage and lose narrowing on later repair.

## Slice A: planning and lock persistence

An activation edge records one consumer's requirement. A runtime generation
records the installed outputs of an admitted skill graph. Both terms refer
to proposed structures below. Existing lock schemas have neither structure.

Choose a successor profile lock with an explicit skill plan. Allocate its
schema through curator-spec before implementation. Preserve frozen lock
schemas 1, 2, and 4. The precedent for a closed successor is
[`cips/CIP-0021-profile-skill-directories.md:86-103`](https://github.com/relux-works/curator-spec/blob/6c7ce40301e254e14e55ebd29405c90fe00c79c6/cips/CIP-0021-profile-skill-directories.md#L86-L103).
This changes lock persistence; it adds no field to `requires.skills`.

The successor lock must retain these inputs:

| Input | Required binding |
| --- | --- |
| Package identity | Kind, skill name, canonical source, full commit, selected directory, and source content identity. Root selection normalizes the omitted directory internally. |
| Dependency graph | Every transitive skill member and every consumer/provider edge, including machine-origin direct constraints and their provenance. |
| Activation | Explicit defaulted `mode` on every edge, plus sorted, unique runtime command narrowing. Absence of narrowing means all exported commands. |
| Resolution evidence | Requested reference and its resolved immutable pin, with the evidence needed by existing source admission and moved-reference checks. |
| Plan semantics | A versioned activation interpretation and canonical ordering, covered by the lock hash. |

The planner derives edges from the pinned manifests. It validates the stored
edges against those manifests on reuse. Machine-origin edges derive from
recorded direct declarations. Tampering, a missing node, or a disagreement
refuses reuse. Platform-specific receipts and executable hashes belong in
the runtime generation marker, bound to the lock hash. They do not enter
the portable lock.

Manifest reconstruction alone is insufficient for the current locks. It can
recover context-origin modes, but cannot supply missing transitive skill
pins without resolving their tags or ranges again. `required_by` also loses
machine-origin activation details. The explicit plan makes inactive installs
and later repair use identical inputs. This justifies a successor despite
its reader and migration cost.

Legacy locks remain readable for inspection and existing context behavior.
Runtime enablement requires an explicit `profile update` or reinstall that
produces the successor. `use`, `sync`, status, and repair must never migrate
by resolving fresh references. The protocol already forbids implicit
resolution in [`protocol/environments.md:268-272`](https://github.com/relux-works/curator-spec/blob/6c7ce40301e254e14e55ebd29405c90fe00c79c6/protocol/environments.md#L268-L272).
Report the required update when a legacy lock cannot prove its runtime plan.

### Bridge into the ordinary closure

Provide a typed closure entry point for resolved requirements. Each input
contains the locked source, commit, directory, consumer, mode, and command
selection. Do not synthesize a `Skillfile.json`: its direct declarations
enter as `full`, which would erase narrowing. The current entry point shows
this behavior at [`internal/closure/closure.go:155-176`](https://github.com/relux-works/curator/blob/f52b8bdd05d0705632b0142d31580e015aba2369/internal/closure/closure.go#L155-L176).

Load root-addressed and directory-selected skills through the same package
reader. Expand ordinary skill dependencies before finalizing the successor
lock. Resolve one source/directory/commit per skill name across profile
requirements and admitted direct constraints. Reject incompatible pins.
Keep the existing context resolver's composition and authority gates.

Compute context and commands with the shared closure functions. Their
existing union rules are at
[`internal/closure/closure.go:52-91`](https://github.com/relux-works/curator/blob/f52b8bdd05d0705632b0142d31580e015aba2369/internal/closure/closure.go#L52-L91).
Only context-active nodes reach the context projection publisher. Only active
script/build commands reach runtime planning. Validate each narrowed name
as an exported script/build command before invoking `ActiveCommands`.
Its current set computation alone can silently omit an unknown name.

Extract the ordinary build and target planner behind this bridge. Retain
strict profile audit, source admission, trusted build drivers, protected
cache verification, and receipt checking. go-v1 retains its fixed toolchain
and build invocation; profile packages gain no shell build hooks. The driver
contract is [`protocol/core.md:647-690,714-724`](https://github.com/relux-works/curator-spec/blob/6c7ce40301e254e14e55ebd29405c90fe00c79c6/protocol/core.md#L647-L724).
Existing private staging and commit seams are
[`internal/install/global.go:331-357,390-410,416-474`](https://github.com/relux-works/curator/blob/f52b8bdd05d0705632b0142d31580e015aba2369/internal/install/global.go#L331-L474).

Keep the shared planner below both callers. `internal/install` already
imports `internal/envprofile`; importing it back would create a cycle.
Extract closure-to-runtime preparation into a lower package, provisionally
`internal/skillruntime`. `envprofile` owns profile lifecycle and adapter
publication. The shared package supplies prepared build targets and receipt
references to the existing transaction coordinator. The ordinary Skillfile
installer must consume the same planner after extraction.

Preserve source-signer policy semantics. Configured allowlists and required
signers must pass before publication. An absent allowlist in permissive
posture currently skips verification; this design does not turn it into
mandatory verification. See
[`internal/contextresolve/contextresolve.go:719-758`](https://github.com/relux-works/curator/blob/f52b8bdd05d0705632b0142d31580e015aba2369/internal/contextresolve/contextresolve.go#L719-L758).
Signature rejection tests must configure a policy that requires verification.

### Ownership and command destinations

Slice A retains machine-current command routing. Canonical shims land in
`<manager-home>/global/bin/<command>`. The selected safe user-bin directory
contains forwarding shims to those canonical shims. Compiled targets remain
in the protected build cache. Existing destinations and receipt bindings are
[`internal/install/global.go:301-302,545-550`](https://github.com/relux-works/curator/blob/f52b8bdd05d0705632b0142d31580e015aba2369/internal/install/global.go#L301-L550),
[`internal/globalbins/stage.go:32-50,111-130`](https://github.com/relux-works/curator/blob/f52b8bdd05d0705632b0142d31580e015aba2369/internal/globalbins/stage.go#L32-L130),
and [`internal/install/targets.go:294-339`](https://github.com/relux-works/curator/blob/f52b8bdd05d0705632b0142d31580e015aba2369/internal/install/targets.go#L294-L339).

Profile operations must not call `install.Global`. They must not write the
operator's global `Skillfile.json`. A profile operation resolves and builds
its profile-derived skill closure. It carries unrelated machine direct
constraints forward without reconciling their installed generations.
It may read their declarations to check constraints. Reading a declaration
does not authorize fetching, rebuilding, deleting, or restarting its package.

Preserve unrelated installed targets as opaque owned entries. Merge the
selected command delta with those entries under the manager lock. Pass the
complete retained command inventory to forwarding publication; otherwise
its stale-removal logic could delete unrelated commands. Never feed the
entire manifest to the compiler to obtain that inventory.

If a selected dependency also has a machine direct declaration, enforce
that declaration's exact constraint. It has become part of the selected
closure. Refuse an incompatible requirement before publication. Preserve
unrelated pre-existing drift and report it separately; do not claim that
the whole machine manifest has converged. This operation boundary needs
spec clarification alongside the successor lock.

Preflight active command collisions across selected and retained providers.
Apply platform path and case-alias rules. Refuse an unmanaged destination
without overwriting it. Refuse `curator-*` command names with
`environment_reserved_command_name`, as required by
[`protocol/environments.md:3148-3152`](https://github.com/relux-works/curator-spec/blob/6c7ce40301e254e14e55ebd29405c90fe00c79c6/protocol/environments.md#L3148-L3152).
Reuse the ordinary collision checks at
[`internal/closure/closure.go:227-246`](https://github.com/relux-works/curator/blob/f52b8bdd05d0705632b0142d31580e015aba2369/internal/closure/closure.go#L227-L246).

### Lifecycle and failure boundary

Stage every selected build before publishing profile records. Bind the
source record, successor lock, runtime generation marker, and command
targets to one planned operation. Revalidate input hashes, admission policy,
destination preimages, and cache receipts under the manager lock before
commit. A concurrent change discards the stale plan and triggers bounded
replanning or an explicit retry diagnostic.

Use the existing transaction engine for mutable runtime targets and records.
Do not nest independent profile and skill commits. The current profile
journal covers manager records, while adapter surfaces have specified
partial-success behavior. These boundaries are explicit in
[`internal/envprofile/lock.go:87-96,162-174`](https://github.com/relux-works/curator/blob/f52b8bdd05d0705632b0142d31580e015aba2369/internal/envprofile/lock.go#L87-L174)
and [`internal/envprofile/switch.go:14-23`](https://github.com/relux-works/curator/blob/f52b8bdd05d0705632b0142d31580e015aba2369/internal/envprofile/switch.go#L14-L23).
Extract a shared prepared runtime plan and one journal coordinator. Keep
adapter backup generations and partial-switch diagnostics intact.

| Path | Proposed behavior |
| --- | --- |
| Inactive install | Resolve, audit, compile, and retain a complete runtime generation. Publish its lock after successful staging. Do not change machine shims, current pointers, or another profile's surfaces. |
| First install or install with `--use` | Prepare the runtime generation first. Materialize adapter surfaces through their existing scope rules. Publish machine shims and the machine-current pointer together only after successful machine-scope activation. |
| Update | Resolve a successor candidate. Build before replacing the installed lock. Preserve the previous lock and runtime generation as a rollback slot. Update machine shims only for the machine-current profile. |
| Machine `use` | Read the installed lock and verify or rebuild its pinned runtime generation. Repoint shims with the current pointer after the scope succeeds. |
| Scoped `use` | Materialize the selected adapter scope. Leave machine shims unchanged, as revision 1 requires. |
| `sync` | Rebuild missing pinned runtime outputs and context projections for installed locks. Publish shell shims only for the machine-current profile. Do not resolve new references. |
| Repair through `env resolve` | Verify the existing lock and runtime generation. Restore missing or drifted managed outputs from pinned inputs under explicit repair authority. Leave inactive profiles' machine routing alone. |
| Status/readiness | Inspect records and outputs without compiling, publishing, resolving, or enrolling. Report incomplete or untrusted runtime evidence. |

A failed compile leaves the prior profile usable. It changes no live lock,
current pointer, context surface, shim, or enrollment state. Private staging
can be removed. Verified immutable cache entries may remain unreferenced
until ordinary GC; they are never evidence of a completed install.

A publication error follows journal rollback and recovery. Before durable
intent, interruption preserves the old generation. After durable intent,
recovery completes or compensates the journal before another mutation.
Retain the prior generation while the journal or rollback slot references
it. Recovery must not compile or enroll to complete an admitted transaction.
Read-only status reports pending recovery instead of performing it.

Installed inactive profiles also keep their runtime generations live.
Record receipt and cache references for every installed generation before
enabling cache pruning on the new path. GC must account for profile roots,
rollback slots, and pending journals in Slice A. Slice B adds snapshot and
active-launch references to that same accounting.

Adapter conflicts retain their specified behavior. For `global add/install`
and the specified update ordering, the protocol permits a new lock to remain
published while conflicting surfaces stay at their prior state. Preserve
that explicit exception in
[`protocol/environments.md:3115-3128`](https://github.com/relux-works/curator-spec/blob/6c7ce40301e254e14e55ebd29405c90fe00c79c6/protocol/environments.md#L3115-L3128).
Every such lock must already have a verified runtime generation. A partial
activation keeps old machine routing and reports the mismatch for `sync`.
This exception never permits publishing a lock before a successful compile.

## Per-instance readiness and enrollment

Software installation and enrollment have separate states. Swarm Memory
supplies the worked example: the `memory` skill exports a go-v1 `memory`
command. Its skill contract requires an explicit home, identity, and grants.
See Memory at `9ee7b511a2e3c68ea31e4eb363f4a99b9dfbe6c2`,
[`agent-skill.json:1-14`](https://github.com/relux-works/swarma-memory/blob/9ee7b511a2e3c68ea31e4eb363f4a99b9dfbe6c2/agent-skill.json#L1-L14)
and [`SKILL.md:14-21`](https://github.com/relux-works/swarma-memory/blob/9ee7b511a2e3c68ea31e4eb363f4a99b9dfbe6c2/SKILL.md#L14-L21).
Installing that package supplies no identity or authority.

Define profile readiness per skill instance. An instance binds an
operator-declared identity to a profile lock and a skill. Curator must not
infer identities from process names or allocate a new one for each spawn.
The following report vocabulary is proposed; it is not implemented today.

| Installation evidence | Enrollment evidence | Profile readiness result |
| --- | --- | --- |
| Verified context/runtime outputs | Skill needs no enrollment | `ready` |
| Verified outputs | Skill reports enrolled for this instance | `ready` |
| Verified outputs | Skill reports enrollment absent | `installed, not enrolled`; installation remains successful |
| Verified outputs | No supported probe, unavailable state, or unreadable evidence | `installed, enrollment unknown`; include a bounded reason |
| Missing or untrusted required outputs | Any | `install incomplete` or `install untrusted`; enrollment cannot hide the failure |

Expose both dimensions in human and structured readiness output. Include
profile lock hash, skill pin, declared identity reference, observation time,
evidence source, and reason code. Include no key material, tokens, or grants
payloads. A report for one identity must not stand in for another identity.
An enrollment probe failure must not become an assertion of enrollment.

Use only a versioned, skill-owned read-only readiness contract. Curator must
select its probe from the verified installed package, with bounded output
and execution. It must not run an arbitrary `PATH` command or a manifest
shell hook. Until that contract exists, report enrollment as unknown.
Probe/schema allocation and whether probing requires explicit invocation
remain owner decisions. This note defines no Memory enrollment protocol.

Collect pending instances into one owner request. The request exposes a
deduplicated identity list, each profile/skill pin, and the requested
enrollment contract reference. Bind its batch identifier to that exact set
and an expiry. One owner approval covers the displayed identity list.
Repeated spawns reuse the pending request instead of asking again.

Revalidate the batch before handing approval to the skill's enrollment
workflow. Changed identities, pins, requested authority, or expiry require a
replacement batch. Approval never covers future unnamed spawns. Curator
reports pending, approved, declined, expired, and observed completion states;
it must not mark approval as completed enrollment. Concurrent reporters
deduplicate by the same batch identifier. Enrollment execution, per-instance
state homes, key custody, and grants remain the skill's contract.

Readiness observation never creates a Memory home, generates a key, enrolls,
or restarts a daemon. `installed, not enrolled` is a defined non-fatal result.
It does not break `profile install`, `update`, `sync`, or `use`. A consumer
may separately require enrollment for a specific task under its own policy.

## Slice B: command roots bound to profile locks

Revision 1 explicitly limits commands to machine-current shims. Managed-home
launches cannot claim profile-specific command resolution. The reserved
extension is `path_prepend` in
[`protocol/environments.md:3137-3147`](https://github.com/relux-works/curator-spec/blob/6c7ce40301e254e14e55ebd29405c90fe00c79c6/protocol/environments.md#L3137-L3147).
Slice A preserves that limit.

Slice B proposes an immutable root beneath
`<manager-home>/environments/<profile>/commands/<lock-hash>/<platform>/bin`.
Each root contains exactly the lock's active command shims. Its runtime
marker binds the lock, source identities, receipts, and executable targets.
The existing machine-current forwarding path can reference that generation.

A verified launch fragment prepends that root through `path_prepend`.
The launcher must validate profile pin, marker, owner, mode, path boundaries,
and referenced runtime generation before changing `PATH`. Missing or
untrusted roots refuse launch when profile commands are required. Ambient
machine shims must never satisfy that requirement silently. Keep the
`curator-*` reservation on every published command root.

Snapshots must capture the runtime generation binding. A retained snapshot,
active launch lease, rollback slot, or pending transaction keeps that
generation and its artifact receipts live. GC removes a generation only
after all such roots disappear. A running session retains its original
command resolution when a profile updates. Two installed profiles can then
run different Memory versions concurrently.

The owner must settle the successor fragment schema, exact root grammar,
snapshot carrier, and lease/GC protocol before Slice B implementation.
Also settle Windows command lookup, project/hybrid precedence, and whether
missing required commands always refuse launch. Test the final launcher
boundary rather than relying on a printed `PATH` value.

## Acceptance and forbidden outcomes

The implementation must provide executable witnesses for these claims.
This design change runs document checks only; the following are future
acceptance requirements. Fixtures must use isolated manager homes and local
signed repositories. No test may enroll a live identity or touch native
operator configuration.

| Witness | Required observation and forbidden outcome |
| --- | --- |
| Profile-only `full` | A clean profile installs its prompt and compiled CLI. Invoke the installed shim and check its version witness. No pre-existing binary may supply the result. |
| Profile-only `context` | The prompt appears. No runtime command target, build invocation, or forwarding shim appears from that edge. |
| Profile-only `runtime` | The CLI works. No prompt directory appears from that edge in managed or in-place skill surfaces. |
| Narrowed runtime | A two-command provider exports only the selected command. The other command is absent from targets and build calls. A separate `full` edge has a positive union control. |
| Transitive closure | A provider's skill dependency gets the declared activation and a locked pin. Sync with moved remote tags retains those pins and selections. |
| Missing selected command | Refuse with the command validation diagnostic before build/publication. Preserve all preimages. A neighboring exported name succeeds. |
| Invalid build | Reject an invalid driver declaration or source root before publication. A valid trusted-driver fixture builds. |
| Bad pin or signature | Reject mismatched immutable inputs and invalid required signer evidence. Configure `source_signers` for the signature case. Permissive absent-policy behavior has its own control. |
| Directory escape | Reject `..`, absolute paths, and linked path components before build/publication. A portable selected directory succeeds. |
| Collision | Reject two active providers, retained unrelated providers, reserved names, and unmanaged destinations. Verify the intended collision diagnostic and unchanged destinations. |
| Compile failure | Update a working profile to a fixture with a compile error. Old lock, current pointer, prompt, and command still work byte-identically. No new receipt or marker claims installation success. |
| Interrupted publication | Inject failure at journal prepare, each target swap, and current-pointer publication. Recover to a coherent generation or the documented partial-adapter state. Preserve rollback references until recovery completes. |
| Inactive lifecycle | Installing/updating an inactive profile builds its generation without changing current routing. Later machine `use` selects its commands; scoped `use` does not. |
| Unrelated global declaration | Pin an unrelated `task-board` declaration at `0.29.3` while its installed fixture reports `0.29.2`. Install profile-only Memory. Assert byte-identical global `Skillfile.json`, task-board marker, binary, shims, service configuration, and daemon identity. Assert zero task-board fetch/build/service-control calls. Preserve the observed drift. |
| Enrollment readiness | Installed Memory without enrollment reports `installed, not enrolled` and successful install. Unknown probe evidence stays unknown. Missing artifacts still fail installation integrity. |
| Batched owner request | Several declared identities produce one stable request and one approval. Repeated spawns produce no duplicate requests. A new identity or changed pin invalidates the old batch. Approval alone never reports enrollment completion. |
| Slice B version isolation | Two profiles pin different Memory versions. Resolve and invoke each through the real verified launch fragment. Each sees its own command even while the other is machine-current. |
| Slice B negative controls | Tamper with a root, fragment pin, receipt, or owner/mode. Refuse instead of falling through to machine `memory`. A live snapshot/lease prevents GC; releasing the last reference permits GC. |

Fault injection must assert the intended diagnostic and post-failure state.
Pair refusal cases with valid controls. A source scan or two planners
agreeing cannot prove that the installed command resolves correctly.

## Owner decisions

| Decision | Recommendation |
| --- | --- |
| Slice A scope | Repair the existing `requires.skills` contract. Add no `requires.tools` family and no installation hook. Curator owns the bridge and runtime transaction. |
| Lock allocation | Adopt an explicit successor skill plan. Assign its schema and canonical fields in curator-spec first. Require explicit update for legacy runtime enablement. |
| Reconciliation boundary | Permit profile-owned closure installation with unrelated direct declarations retained unchanged. Specify drift reporting and refusal on overlapping incompatible pins. Never use full global reconciliation as the profile bridge. |
| Transaction semantics | Compile before lock publication. Share one runtime/record journal. Preserve documented adapter partial-success and handled-conflict exceptions. |
| Command scope | Ship Slice A with machine-current commands. Make simultaneous role command isolation depend on Slice B and its launcher conformance tests. |
| Enrollment readiness | Adopt separate installation and enrollment dimensions. Curator exposes observations and one identity-list approval request; the skill owns enrollment and custody. Allocate the readiness probe/report contract jointly with skill owners. |
| Approval authority | Bind each approval to the exact batch, pins, requested authority, and expiry. Treat added identities as a new owner request. Never authorize enrollment from a role name. |
| Memory adoption | Require a reviewed signed Memory release and an operator-adopted source signer policy. A successful package check alone does not qualify a release or enroll an instance. |
| Slice B carrier and liveness | Version the verified fragment and snapshot bindings together. Give the launcher responsibility for binding verification and Curator responsibility for generation publication and GC accounting. |
| Delivery order | Review the spec delta and this design before implementation. Implement and qualify Slice A before adopting profile-only CLI dependencies. Qualify Slice B before promising concurrent version isolation. |
