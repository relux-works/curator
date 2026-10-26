# Profile skill runtime

Profiles must install their skills' active commands. Slice A connects profile
requirements to the ordinary skill installer. Slice B gives each profile its
own command root. This note proposes a design for `TASK-261026-6b5ioc`.
It authorizes no implementation, installation, or enrollment.

## Accepted direction and role-profile sources

The owner accepts this design direction. Every spec MUST requires an
executable conformance test with a real install in an isolated home,
behind a CI gate. This direction is accepted; executable coverage and
runtime qualification remain pending. The coverage contract below makes
that distinction enforceable.

Public reusable role profiles belong in `swarma-*` repositories. Derive
them from the designated upstream role-profile set when its owner publishes
it. Keep personal and host policy in separate admitted overlays.
`relux-root-context` is not a required base. This PR creates no role package.

| Designated upstream | Public reference available for adoption |
| --- | --- |
| Product-keeper profile | Repository, package directory, signed tag, and full commit are pending upstream publication. |
| Project-orchestrator profile | Repository, package directory, signed tag, and full commit are pending upstream publication. |
| Shared modules for keeper, orchestrator, developer, tester, architect, reviewer, and researcher | The upstream release must identify the reusable modules and dependency pins. No source pin is inferred here. |

Record those references in the adoption review before publishing derived
profiles. An unpublished local prototype supplies no release pin.

The reviewer role permits publishing verdicts, findings, and handovers
through authorized review channels. It grants no product-integration
authority to push product changes, merge, or deploy. Those permissions
require separate authority. Enrollment approval is another permission;
neither a reviewer label nor permission to post a verdict grants it.

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
schema through curator-spec before implementation. Published lock schemas
1 and 2 remain byte-identical. Schema 4 is unreleased and may still grow,
as stated in
[`cips/CIP-0021-profile-skill-directories.md:86-103`](https://github.com/relux-works/curator-spec/blob/6c7ce40301e254e14e55ebd29405c90fe00c79c6/cips/CIP-0021-profile-skill-directories.md#L86-L103).
Preserving today's schema-4 shape is this design's compatibility
recommendation, not an existing release freeze. Prefer a new carrier because
the complete activation graph changes lock interpretation and reader
requirements. A separate identity makes reader opt-in explicit and keeps
existing schema-4 directory readers from accepting incomplete runtime state.
Extending unreleased schema 4 remains possible pending the spec owner's
allocation decision. This design chooses the separate successor and adds
no field to `requires.skills`.

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
in the protected build cache. These shims serve only the machine-current
profile. Inactive profiles receive retained runtime generations without
shell publication. Managed-home launches of other profiles receive no
profile-bound commands from Slice A. An accidentally inherited machine shim
does not satisfy that other profile's requirement. Adoption of those
launches requires qualified Slice B routing.

Existing destinations and receipt bindings are
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

### Isolated single-profile admission

Slice A qualifies one current profile in an isolated routing domain.
A routing domain contains a private manager home and private user-bin
directory with one supervised consumer population. Its fixed PATH already
names that private directory. Slice A never edits machine PATH, shell
startup files, or a shared executable artifact. It publishes new immutable
artifacts and replaces only domain-owned command shims.
User-bin selection must explicitly choose the private domain directory.
Refuse a PATH-scan result outside that boundary. The canonical bin, cache,
and forwarding ledger must also belong exclusively to the domain.

Require an exclusive routing witness before any live publication. An
isolation supervisor must bind the witness to the domain's filesystem
identities, fixed PATH, current profile, and admitted consumer set. It must
exclude consumers outside that domain and serialize consumer admission
against routing mutations. An OS-isolated disposable test environment can
supply this boundary. An ordinary home path, owner acknowledgement, process
scan, or empty lease file cannot prove exclusive routing.

The mutator holds the domain's admission lock through revalidation,
publication, and recovery. Acquire that lock and validate the witness before
entering `beginOperationValidated` or ordinary runtime-journal recovery.
Revalidate it after acquiring the manager-home lock. Every writer of the
domain's shared shims must use this same admission boundary or refuse;
gating only the profile CLI would leave another publisher able to bypass it.
Before first activation, it requires no consumers
of another profile. Update, sync, and repair require every admitted consumer
to use the same current profile. Machine `use` of another profile requires
the previous consumer set to drain before changing routing. New consumers
cannot enter during these checks. Any other live profile, unknown consumer
state, missing witness, or mismatched boundary refuses before publication.
The supervisor retains the lock while recovering durable routing intent;
it admits no consumer into a partially published generation.

Current code supplies no such Slice A admission witness. Until this gate
exists, qualify A only in a supervised disposable domain. Do not adopt A
on a host with multiple live profiles sharing the command namespace.
Inactive preparation can retain generations, but cannot bypass the gate
when it later activates. Managed-home launches of another profile are
outside A's admitted population. Qualified Slice B is required before
multi-profile adoption. Slice A requires no launcher change; its isolation
boundary is supplied outside the launcher.

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
| First install or install with `--use` | Prepare the runtime generation first. Require the isolated single-profile admission witness before any live publication. Publish domain-owned machine shims and the current pointer together after successful scope activation. |
| Update | Resolve and build a successor before replacing the lock. Retain the prior lock and runtime generation. Require the admission witness for current-profile publication; refuse another live profile or unknown consumers. |
| Machine `use` | Verify or rebuild the pinned generation. Require the admission witness and a drained old-profile consumer set before changing profile. Repoint only domain-owned shims with the current pointer after scope success. |
| Scoped `use` | Materialize the selected adapter scope. Leave machine shims unchanged, as revision 1 requires. |
| `sync` | Rebuild pinned outputs without resolving references. Require the admission witness before current routing or live surfaces change. Refuse if another profile can consume those commands. Leave inactive routing alone. |
| Repair through `env resolve` | Restore pinned outputs under explicit repair authority. Require the admission witness before live publication, including recovery. Preserve all routing on a failed or unknown admission check. |
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

Slice B also requires a launcher change. At curator-run source
`89dd2d124fb9fd2151a73c22fef3508119786362`,
[`internal/fragment/fragment.go:549-563`](https://github.com/relux-works/curator-agent-launcher/blob/89dd2d124fb9fd2151a73c22fef3508119786362/internal/fragment/fragment.go#L549-L563)
parses and validates `path_prepend`. It remains inert. The launch regression
at [`cmd/curator-run/reviewer_reserved_test.go:11-43`](https://github.com/relux-works/curator-agent-launcher/blob/89dd2d124fb9fd2151a73c22fef3508119786362/cmd/curator-run/reviewer_reserved_test.go#L11-L43)
asserts that reserved metadata leaves PATH unchanged. The D-PATH target
contract also distinguishes that current behavior in
[`SPEC.md:2156-2166`](https://github.com/relux-works/curator-agent-launcher/blob/89dd2d124fb9fd2151a73c22fef3508119786362/SPEC.md#L2156-L2166).
These are source findings, not new launcher qualification.

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

For interactive native and hosted launches, curator-run must apply the
verified fragment's `path_prepend`, bound to the selected profile lock.
Construct one effective PATH after verification and admission. Pass that
exact value to the provider process and record it in launch provenance,
with the fragment digest, lock pin, and runtime generation. A later ambient
PATH change must not cause recomputation. The hosted session-manager launch
plan must carry that same PATH and its binding through to execution.
Serializing the field without applying it does not satisfy this contract.

One-shot D-PATH needs a separate owner decision. Its target order is the
protected tool directory, admitted base PATH, then protected dispatcher.
It currently excludes reserved fragment data. Recommend admitting a
verified lock-bound profile root after protected tools and before admitted
base PATH, subject to the existing heavy-command and project-dispatch
controls. The spec owner and curator-run lane must decide whether to admit
that root and its position. Until then, preserve the existing one-shot
recipe and make no one-shot profile-command claim.

The Slice B review list includes the Curator runtime/store owner, the
curator-spec owner, the curator-run lane, and the hosted session-manager
owner. Curator owns command publication and liveness. The curator-run lane
owns PATH application and provenance. The session-manager owner owns
transporting and enforcing the PATH in hosted launch plans. Slice A has
no launcher implementation dependency.

## Acceptance and forbidden outcomes

The implementation must provide executable witnesses for these claims.
This design change runs document checks only; the following are future
acceptance requirements. Fixtures must use isolated manager homes and local
signed repositories. No test may enroll a live identity or touch native
operator configuration.

### Exact artifacts and refusal preimages

A version string cannot identify executable bytes. For every successful
install, independently resolve the published forwarding and canonical shim
chain to its executable. Compare the reached target's regular-file bytes
with the admitted artifact SHA-256, execution receipt, and runtime marker.
Check that the marker names the expected lock hash and the receipt names
the selected source/commit/directory and trusted build inputs. Script
commands additionally bind the interpreter and script-runtime digests.
The command invocation records the reached target identity. A matching
`--version` result is only supplemental evidence.

Plant a second executable with the same version output and different
bytes. Redirect the shim or substitute that artifact between planning and
publication. Revalidation must refuse the identity mismatch before
publication or execution. A separate post-publication plant must fail the
independent byte witness; the harness never invokes that altered artifact.
The valid receipt-bound artifact is the positive control. Preserve the
previously working command and its exact target bytes on refused updates.

Capture preimages before each refusal test. The preimage set contains old
lock bytes, runtime marker and receipt bytes, current/scoped pointer bytes,
prompt tree contents, canonical and forwarding shim bytes or link targets,
and resolved executable bytes. Record file kinds, modes, and absence too.
Hashing a shim alone cannot prove its target remained unchanged.
Selected and unrelated declarations have separate preimage inventories.

Use these paired controls on the real install/update entry points:

| Pair ID | Admitted control | Refused or interrupted control and required state |
| --- | --- | --- |
| `A-BUILD` | A trusted go-v1 package compiles and reaches its receipt-bound artifact. | The next pin has a compiler error. Every live preimage remains old; no new installation marker or receipt claims success. |
| `A-AUDIT` | A compilable fixture declares a covering network capability and passes strict audit. | The same compilable fixture omits that capability. Require `audit.capability.network-undeclared`, zero compiler calls, and every live preimage old. |
| `A-SIGN` | The required source signer verifies for the selected pin. | An invalid or unaccepted signature under configured required policy refuses before publication. Every live preimage remains old. |
| `A-COLLISION` | Distinct active command names publish successfully. | A selected command collides with another active or retained provider. Require the collision diagnostic, zero publication, and every live preimage old. |
| `A-INTERRUPT` | Uninterrupted activation publishes a coherent new generation. | Inject interruption before durable intent, after prepare, after each swap, and at pointer publication. Check the recovery states below at every boundary. |

Before durable intent, the only allowed state is the complete old preimage
set. A failed commit that completes compensation restores that same set.
After durable intent, a pending journal can temporarily contain staged
old/new targets, but the isolation supervisor admits no consumer. Successful
recovery yields the complete journal-selected new generation or the complete
compensated old generation, including shims and reached executable digests.
No successful result may expose a mixed runtime generation.

Assert the recovery result from the durable phase, with no further injected
fault during that recovery. The existing phase dispatch is
[`internal/transaction/engine.go:397-433`](https://github.com/relux-works/curator/blob/f52b8bdd05d0705632b0142d31580e015aba2369/internal/transaction/engine.go#L397-L433).

| Interruption phase | Exact result after successful recovery |
| --- | --- |
| No durable journal, or `preparing` | The entire old runtime/record preimage set remains; discard incomplete staging. |
| `prepared` or `committing` | Resume the admitted transaction to the entire new runtime/record set, including shims and current pointer. |
| `cleanup` | Keep the entire new set and finish backup cleanup. |
| `rolling_back` or `rollback_cleanup` | Restore the entire old set and finish rollback cleanup. |

An ordinary commit error durably selects rollback. Recovery corruption,
unknown bytes, or a new admission failure keeps the domain closed and
reports failed recovery; it never reports a coherent completed install.
The adapter exceptions below remain separate from this runtime phase table.

Keep adapter exceptions explicit in these tests. A handled unmanaged
surface conflict retains the new lock and verified runtime generation,
restores touched in-place surfaces, preserves the unmanaged surface, and
keeps old current pointers and machine shims. Other documented partial
adapter results keep the old machine routing and report their exact per-entry
state. Retain the prior generation and pending recovery references. Tests
must assert those specified exceptions rather than accept arbitrary mixtures.

| Witness | Required observation and forbidden outcome |
| --- | --- |
| Profile-only `full` | A clean profile installs its prompt and compiled CLI. Invoke the published shim and verify its exact executable, receipt, and lock-bound generation through the preceding byte-identity procedure. No ambient binary may supply the result. |
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
| Strict audit failure | Run `A-AUDIT` with a fixture that would otherwise compile. Require the blocking audit diagnostic before compilation and assert the entire old preimage set. |
| Interrupted publication | Inject failure at journal prepare, each target swap, and current-pointer publication. Recover to a coherent generation or the documented partial-adapter state. Preserve rollback references until recovery completes. |
| Inactive lifecycle | Installing/updating an inactive profile builds its generation without changing current routing. Later machine `use` selects its commands; scoped `use` does not. |
| Exclusive Slice A routing | In a supervised isolated domain, one current profile installs/updates successfully. Attempt first activation, update, machine use, sync, repair, and recovery with a second live profile or unknown admission state. Each refuses before live publication. Verify the second profile's command resolution, shim targets, executable bytes, PATH, and process identity remain unchanged. Removing the conflicting consumer under the admission lock enables the valid control. |
| Unrelated global declaration | Pin an unrelated `task-board` declaration at `0.29.3` while its installed fixture reports `0.29.2`. Install profile-only Memory. Assert byte-identical global `Skillfile.json`, task-board marker, binary, shims, service configuration, and daemon identity. Assert zero task-board fetch/build/service-control calls. Preserve the observed drift. |
| Enrollment readiness | Installed Memory without enrollment reports `installed, not enrolled` and successful install. Unknown probe evidence stays unknown. Missing artifacts still fail installation integrity. |
| Batched owner request | Several declared identities produce one stable request and one approval. Repeated spawns produce no duplicate requests. A new identity or changed pin invalidates the old batch. Approval alone never reports enrollment completion. |
| Slice B version isolation | Two profiles pin different Memory versions. Resolve and invoke each through the real verified launch fragment. Each sees its own command even while the other is machine-current. |
| Slice B negative controls | Tamper with a root, fragment pin, receipt, or owner/mode. Refuse instead of falling through to machine `memory`. A live snapshot/lease prevents GC; releasing the last reference permits GC. |
| Slice B launcher delivery | Launch real isolated native and hosted interactive providers. Verify each child's effective PATH, reached artifact, provenance binding, and hosted launch plan. An inert `path_prepend`, omitted hosted PATH, mismatched lock, or ambient PATH recomputation must fail the witness. One-shot expectations remain pending the D-PATH decision. |

Fault injection must assert the intended diagnostic and post-failure state.
Pair refusal cases with valid controls. A source scan or two planners
agreeing cannot prove that the installed command resolves correctly.

### MUST traceability and CI gate

The accepted direction requires executable coverage for every MUST.
The following catalog covers every uppercase MUST, including MUST NOT,
in core §4.4 and environments §9.4 at spec `6c7ce403`.
Compound clauses retain all their assertions. Source ranges are
[`protocol/core.md:1138-1191`](https://github.com/relux-works/curator-spec/blob/6c7ce40301e254e14e55ebd29405c90fe00c79c6/protocol/core.md#L1138-L1191)
and [`protocol/environments.md:3088-3206`](https://github.com/relux-works/curator-spec/blob/6c7ce40301e254e14e55ebd29405c90fe00c79c6/protocol/environments.md#L3088-L3206).
The row IDs below are proposed stable obligation/vector IDs.

Reserve the suite entry point
`TestProfileSkillRuntimeConformance` in
`internal/crossconformance/profile_skill_runtime_test.go`. Each row names
an exact subtest group. Each case in its controls column names a concrete
subtest below that group. For example, `C44M01/inside` expands to
`TestProfileSkillRuntimeConformance/C44M01/inside`. The proposed vectors
live under spec `conformance/v1/profile-skill-runtime/`, keyed by those IDs.
These test and vector allocations do not exist in this PR.

Every row is `PENDING`. None counts as passing coverage. Each group must
invoke the real Curator CLI install/migration path in a fresh isolated
manager home and inspect actual publication or refusal. Parser-only unit
tests can supplement that witness. They cannot replace the real install.

| Obligation / exact subtest group | Spec clause | Cases and observable assertions | Status |
| --- | --- | --- | --- |
| `C44M01` | `core.md:1155-1158` | `inside` installs a matching package; `escape`, `linked`, `missing`, and `wrong_name` refuse a selected dependency before publication. Draft schema-9 selection is explicitly opted in for these vectors. | `PENDING` |
| `C44M02` | `core.md:1168-1169` | `independent` validates, installs, locks, and audits two distinct packages in one snapshot; `blocked_sibling` refuses the blocked package even if its sibling has an admitted audit record. | `PENDING` |
| `C44M03` | `core.md:1171` | `normalized` records `.` for root selection and the selected relative directory in dependency lock/install identities; `missing_directory` detects an omitted or incorrect identity field. | `PENDING` |
| `C44M04` | `core.md:1172` | `agreement` binds package and lock directories; `disagreement` refuses a planted mismatch while preserving old outputs. | `PENDING` |
| `C44M05` | `core.md:1173-1176` | `audit_bound` admits the audited tuple; `audit_replay` rejects reuse for another directory at the same source/commit. An attestation cannot substitute for local per-package audit. | `PENDING` |
| `C44M06` | `core.md:1189-1191` | `modern_writer` performs a real install without creating `agents/runtime.json`; `legacy_writer_plant` proves the witness rejects a writer that creates that legacy form; `legacy_read` reads an existing legacy fixture without authoring that form. | `PENDING` |
| `E94M01` | `environments.md:3101-3106` | `selected_surfaces` installs the selected directory's `SKILL.md` and package content into managed and in-place surfaces; `sibling_root` detects a substituted repository-root or sibling projection. | `PENDING` |
| `E94M02` | `environments.md:3115-3119` | `publish_order` runs global add and install for each selected profile and observes immutable entries and extended lock before in-place writes; `early_surface_plant` detects publication in the wrong order. | `PENDING` |
| `E94M03` | `environments.md:3121-3128` | `handled_conflict` keeps the new lock/store, restores touched surfaces, preserves the conflicting unmanaged surface and all other mutable state, and names the conflict; `ordinary_failure` restores old state under the ordinary rollback rule. | `PENDING` |
| `E94M04` | `environments.md:3148-3149` | `ordinary_names` installs allowed commands; `reserved_payload` refuses a `curator-*` executable in every managed skill bin or skill-published PATH root before exposure. | `PENDING` |
| `E94M05` | `environments.md:3150-3152` | `reservation_diagnostic` asserts exactly `environment_reserved_command_name` on the refused install; `allowed_neighbor` succeeds with a neighboring nonreserved name. | `PENDING` |
| `E94M06` | `environments.md:3176-3180` | `root_migration` preserves legacy root addressing; `nested_decoy` proves migration neither infers a directory from layout nor rewrites legacy records to select one. | `PENDING` |
| `E94M07` | `environments.md:3190-3197` | `own_capability` admits migration with the member's own covering declaration and exercises loopback/reserved-label exemptions; `root_capability` refuses borrowing capabilities from the synthesized root or global declarations. | `PENDING` |
| `E94M08` | `environments.md:3198-3200` | `covered_contact` migrates; `absent_manifest`, `absent_network`, and `network_none` require blocking `audit.capability.network-undeclared` for an uncovered contact. | `PENDING` |
| `E94M09` | `environments.md:3200-3201` | `diagnostic_identity` asserts the refused migration names skill, undeclared host, and detecting file; `truncated_diagnostic_plant` proves omission of any field fails the witness. | `PENDING` |
| `E94M10` | `environments.md:3202-3204` | `diagnostic_remedy` asserts the covering-host-glob guidance, correct `agent-skill.json` or legacy `csk-skill.json` spelling, and unmigrated alternative; `missing_remedy_plant` fails that witness. | `PENDING` |
| `E94M11` | `environments.md:3205-3206` | `preserve_capabilities` refuses uncovered migration with unchanged declarations and a blocking finding; `synthesized_or_downgraded_plant` fails if capabilities are invented or severity is weakened. | `PENDING` |

Normative rules in those sections also use declarative wording. The next
rows cover that behavior and this design's additional obligations. Their
subtest naming and `PENDING` meaning match the MUST catalog.

| Obligation / exact subtest group | Source or design section | Cases and observable assertions | Status |
| --- | --- | --- | --- |
| `C44R01` | Core §4.4 exact refs, modes, commands | `full`, `context`, `runtime_all`, `runtime_narrow`, and `union` install their exact surfaces; `branch_range`, `duplicate`, `empty_narrow`, `system_narrow`, and `missing_command` refuse with the expected validation diagnostic. | `PENDING` |
| `C44R02` | Core §4.4 unification and prerequisites | `same_tuple_refs` unifies aliases resolving to one tuple; `source_or_directory_conflict` refuses incompatible tuples; `system_present` and `system_missing` check prerequisites; `legacy_requirement` reads the legacy command form with its migration warning and no shim. | `PENDING` |
| `E94R01` | Environments §9.4 direct declarations and scopes | `direct_constraint` combines machine and profile requirements; `incompatible_direct` refuses; `selected_profiles` exercises current, explicit-profile, and all-profile selection; `takeover_flag` refuses unsupported global-operation takeover. | `PENDING` |
| `E94R02` | Environments §9.4 sync and commands | `sync_registered` actualizes all installed profiles/participating targets from pins; `unreadable_lock` reports `environment_store_untrusted`; `machine_shims` serves only machine-current commands; `other_managed_profile` supplies no profile-bound runtime in A. | `PENDING` |
| `E94R03` | Environments §9.4 hybrid scope | `hybrid_unchanged` keeps hybrid nodes outside locks/switches and renders once; `project_precedence` checks project, hybrid, applicable profile order; `profile_switch` preserves hybrid bytes. | `PENDING` |
| `E94R04` | Environments §9.4 default and lossless migration | `default_local` creates the local hash-pinned root without context and preserves installed behavior; `unreadable_record` stops before writes with `environment_import_lossy` naming the skill; `consent_bypass` cannot drop the skill. | `PENDING` |
| `AISO` | Isolated single-profile admission | `exclusive` permits domain-owned publication; `second_live`, `unknown`, and `admission_race` refuse first activation/update/use/sync/repair/recovery and preserve the other consumer's exact resolution/bytes. | `PENDING` |
| `ABYTES` | Exact artifacts and refusal preimages | `receipt_bound` verifies the reached artifact; `same_version_other_bytes` refuses; `build`, `audit`, `signature`, `collision`, and `interruption` run all paired controls and preimage/recovery assertions above. | `PENDING` |
| `APINS` | Planning and lifecycle | `inactive`, `pinned_sync`, `repair`, and `rollback` retain graph/receipts and GC roots; `edge_tamper`, `moved_ref`, and `legacy_enablement` prevent implicit resolution or incomplete-plan adoption. | `PENDING` |
| `ADRIFT` | Ownership and command destinations | `memory_only` retains the unrelated task-board declaration and all installed bytes with zero fetch/build/service calls; `overlap_conflict` refuses an incompatible selected constraint. | `PENDING` |
| `AREADY` | Per-instance readiness | `not_enrolled`, `unknown`, and `enrolled` preserve installation semantics; `batch_reuse`, `batch_change`, and `approval_only` enforce exact-set approval without enrollment side effects. | `PENDING` |
| `BPATH` | Slice B and launcher delivery | `native`, `hosted`, and `two_versions` verify child PATH, reached bytes, provenance and hosted-plan bindings; `inert`, `bad_pin`, `omitted_path`, `ambient_recompute`, and `gc_live` detect forbidden outcomes. `one_shot` remains blocked on the D-PATH decision. | `PENDING` |

Propose a required `profile-skill-runtime-conformance` CI gate for candidate
implementation PRs and adoption. Wire it into `.github/workflows/ci.yml`
through proposed `.github/ci/profile-skill-runtime-gate.sh`. Its manifest
records the spec commit,
source section/range, normalized clause digest, obligation ID, test/vector
IDs, phase, supported platforms, and evidence location. Phase A includes
all existing-spec rows and all `A` rows. Phase B adds `BPATH`. The reviewed
gate configuration fixes those sets; a manifest cannot drop an obligation
by relabeling its phase. Enumerate MUST occurrences
from the pinned source, including MUST NOT, independently of that manifest.
Compare the resulting clause set and every compound assertion against the
catalog. A changed spec pin requires a reviewed catalog update. Imported
normative obligations need their own mappings before claiming coverage.

The gate consumes test discovery and actual Go JSON results. Require every
mapped subtest and case to execute successfully on each declared platform.
Missing, renamed, skipped, disabled, failed, or unreported tests fail the
gate. Missing/duplicate mappings, unknown IDs, stale clause digests, pending
rows, missing positive or negative controls, and absent real-install
evidence also fail it for every obligation in the selected phase. The
reviewed CI configuration fixes the required platform set too. Platform
exclusions require an explicit spec applicability ruling; a skip cannot
satisfy an applicable obligation.
Evidence binds the implementation SHA, suite/spec digest, isolated-home
identity, invoked CLI, exit status, diagnostic, and resulting byte inventory.

Gate controls must delete a mapping, disable one mapped subtest, force a
skip, and replace a real install with a parser-only check. Each control
must make the gate fail for that intended coverage defect. Keep an intact
positive run beside those controls. Also run the behavioral plants named
in the catalog against their actual witnesses. A missing plant execution
or unrelated harness failure never proves detection.

This document PR does not implement that gate or any listed runtime test.
It can pass document checks while the matrix stays pending. No Slice A
conformance claim or adoption may pass until the implementation gate is
required and green with all A and existing-spec rows executed. Slice B
needs the additional launcher/session-manager gate and completed `BPATH`
cases before its separate adoption claim.

## Owner decisions

| Decision | Recommendation |
| --- | --- |
| Accepted design and coverage direction | The owner accepts this design direction and every spec MUST becoming an executable real-install conformance test in an isolated home behind CI. Runtime acceptance remains blocked on the pending catalog and required gate. |
| Accepted reusable profile location | Publish reusable role profiles in `swarma-*`, derived from the designated upstream role-profile release. Complete public pins before adoption. Keep personal/host overlays separate. |
| Slice A scope | Repair `requires.skills` only in an enforceably isolated single-current-profile domain. Refuse unknown/shared routing. Multi-profile adoption waits for qualified B. Add no dependency family or installation hook. |
| Lock allocation | Prefer a separate explicit successor skill plan. Published schemas 1-2 stay frozen; preserving unreleased schema 4 is a compatibility proposal. The spec owner allocates the carrier and canonical fields before implementation. |
| Reconciliation boundary | Permit profile-owned closure installation with unrelated direct declarations retained unchanged. Specify drift reporting and refusal on overlapping incompatible pins. Never use full global reconciliation as the profile bridge. |
| Transaction semantics | Compile before lock publication. Share one runtime/record journal. Preserve documented adapter partial-success and handled-conflict exceptions. |
| Command scope | A publishes only the isolated machine-current profile's command shims. Other managed-home profiles receive no profile-bound CLI from A. B requires curator-run PATH application, provenance, and hosted-plan changes plus launcher conformance. |
| Enrollment readiness | Adopt separate installation and enrollment dimensions. Curator exposes observations and one identity-list approval request; the skill owns enrollment and custody. Allocate the readiness probe/report contract jointly with skill owners. |
| Approval authority | Bind each approval to the exact batch, pins, requested authority, and expiry. Treat added identities as a new owner request. Never authorize enrollment from a role name. |
| Memory adoption | Require a reviewed signed Memory release and an operator-adopted source signer policy. A successful package check alone does not qualify a release or enroll an instance. |
| Slice B carrier and liveness | Version the verified fragment and snapshot bindings together. Give the launcher responsibility for binding verification and Curator responsibility for generation publication and GC accounting. |
| One-shot D-PATH | Decide whether the protected recipe admits a verified profile root. Recommend placement after protected tools and before admitted base PATH, preserving dispatcher/heavy-command controls. The curator-run lane and spec owner review the decision. |
| Delivery order and review | Resolve remaining spec details, implement A, and pass the required real-install coverage gate before isolated adoption. Require Curator, spec, curator-run, and hosted session-manager review and qualified B before multi-profile adoption. Reviewers may publish verdicts and handovers without acquiring product-integration or enrollment authority. |
