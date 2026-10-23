# Claude strict mode: fresh-session contract and qualification plan

TASK-261022-h1ttey, design only, 2026-10-23. No product code, adapter
advertisement or native-session qualification is delivered by this note.
The owner selected native-project for wave 1 and strict exclusion as a later,
per-runtime qualification, with fresh Claude sessions first. This plan does
not convert that sequencing decision into permission to launch an unqualified
strict tuple.

## Evidence and exact target

| Source | Identity and implication |
| --- | --- |
| Installed runtime, metadata observed for this task | Claude Code `2.1.309`, macOS `27.0`, `arm64`; executable SHA-256 `5eadc0009cdcdbc0d06e58110d7b2baf88ced6088b1c33a60a21217fe410619c`. Private-HOME `--version` and `--help` succeeded; executable hashes before and after help agreed. These are identity observations, not exclusion evidence. |
| [Existing native record](project-source-exclusion-claude.md) | Claude Code `2.1.305`, macOS arm64, SHA-256 `ea3ea3c08d2a9634a4651e687fcf900d410c489bce2b00f22bd7352b91ba704f`. Startup and selected later paths were measured; complete strict qualification was refused. |
| [Existing harness](../../.research/project-exclusion-claude/README.md) | `replay_test.go` pins the older digest and `TestIdentity` pins `2.1.305`. Running the installed binary against it must refuse, not qualify the new release. |
| [Adapter candidate](https://github.com/relux-works/curator/tree/89f0a20c78616525ff98a2e897efcbc353df2da7), TASK-261022-sg5y2x | Its `ClaudeProjectVersion` is `2.1.305`. Construction helpers exist, but `RequireStrict` always refuses and `ClaudeLaunchMode` returns no executable contribution. This is a candidate dependency, not code on this note's base. |
| Normative baseline | [Project context](https://github.com/relux-works/curator-spec/blob/2510171ca9e7a4a9d2b69e7b2e7ec23b6d11bb8a/protocol/project-context.md), especially sections 3, 5, 6 and 8; [diagnostic registry, section 10.6](https://github.com/relux-works/curator-spec/blob/2510171ca9e7a4a9d2b69e7b2e7ec23b6d11bb8a/protocol/environments.md#106-project-context-diagnostics). |

The old unsafe-origin resume and fork-resume witnesses motivate refusal; their
results do not establish the behavior of 2.1.309. The old late-hook negative
was non-discriminating: the hook did not fire even with project discovery
allowed. Late settings, plugins and MCP remain gaps. The adapter candidate's
[gap inventory](https://github.com/relux-works/curator/blob/89f0a20c78616525ff98a2e897efcbc353df2da7/.research/project-claude-native-project-gaps.md)
also identifies unqualified manager default deny and automatic-memory effects.
Neither supplying `dontAsk` nor setting the memory variable closes those gaps.

## Proposed strict launch contract

1. Admission explicitly selects `strict` in protected registration,
   composition and launch bindings. Missing certification refuses, including
   when project admission is off. Never switch to native-project or legacy
   launch after a strict failure. Mode changes require an explicit operator
   update with new identities and the existing busy-home rules.
2. The first proposed supported form is a fresh headless `--print` process,
   using only a structured manager launch request. Interactive/TUI, cloud,
   remote, background and detached forms remain unsupported pending separate
   qualification and the owner decisions below. This is a proposed restriction,
   not a claim that headless mode disables every reload or child path.
3. Fresh means no transcript, restored state or imported conversation is loaded.
   The manager reserves a new session identity and proves it is unused before
   launch, binding it to the admitted composition, executable and lease.
   A new process, new UUID, new HOME, or `--no-session-persistence` alone is not
   proof. A reused stable managed home must not cause implicit conversation
   recovery. Lost or unreadable session evidence refuses; it is not absence.
4. Refuse resume and fork-resume even for previously excluded-origin sessions
   in this first slice. Cover `-r`/`--resume`, `-c`/`--continue`,
   `--fork-session`, `--from-pr`, `--teleport`, `attach`, `respawn`, cloud or
   remote attachment and any equivalent restoration path. Reject caller-chosen
   existing `--session-id` values, aliases, equals forms and unknown entry
   values through typed parsing and a manager-owned argv allowlist. No raw
   passthrough argv or environment overrides. Native in-session history
   commands must also be proven unreachable or refused before restoration;
   checking the initial CLI alone is insufficient. Refusal creates no fresh
   fallback session, child, MCP connection, partial controls or lease takeover.
5. Native automatic instruction/configuration discovery must exclude project
   and ancestor `CLAUDE.md`, `CLAUDE.local.md`, `.claude/CLAUDE.md`, rules,
   compatibility `AGENTS.md`, nested and path-scoped instructions; project and
   local settings; native skill/agent/command roots; hooks, plugins and MCP;
   automatic memory. Cover startup, native tool traversal, explicit reload,
   late file creation and every reachable child path. Native control-file
   imports are refused under leaf 1, including inventoried targets. Additional
   directories and symlink aliases require evidence or an enforced refusal.
6. Approved Curator channels remain usable. The profile root stays first;
   approved project context is appended in normative order, with system parts
   kept in their separate additive channel. Whole skills and explicit approved
   MCP remain pinned and individually approved. No selected item disappears
   to make an exclusion test pass. A zero-server selection still supplies an
   explicit existing empty MCP file. New project-native hooks/plugins/settings
   delivery is not introduced by this plan. Synthetic managed hook/plugin
   controls below diagnose selectivity; they do not authorize new item types.
7. Manager permissions remain default deny, with exact manager grants and
   denies. Repository settings cannot expand or restrict that policy. Native
   automatic memory stays off, including load and write effects. Ordinary
   permitted reads of project data must still work: exclusion concerns native
   control discovery, not a claim that project bytes can never be read.
8. Controls, policy and materialization receipts are verified under the
   same-home lock. Reserve a prepared lease, durably transfer it to the actual
   session/process generation before provider start, retain active/unknown
   leases until authoritative reconciliation, and prevent updates while busy.
   Current admission and dispatch authority remain required. Replacing files
   cannot remove context already loaded in a running process.

The candidate recipe is the adapter's managed config home,
`--setting-sources user`, one explicit `--mcp-config` with
`--strict-mcp-config`, `--permission-mode dontAsk`, exact `--tools` ceiling,
manager MCP deny and `CLAUDE_CODE_DISABLE_AUTO_MEMORY=1`, plus the approved
instruction channels. It is a hypothesis to test as a complete launch recipe.
The installed help also advertises `--restricted`, `--bare` and `--safe-mode`;
none substitutes for evidence or for preserved approved positive controls.
Caller `--settings`, plugin paths/URLs, agent definitions, additional directories,
permission bypasses and alternative MCP/source flags cannot override the recipe.

Fresh entry does not solve late reloads. A reachable unqualified loader keeps
strict unsupported. A restriction may remove a path only when the real native
consumer can enforce it before the load and a bypass control proves that
boundary. If the runtime cannot enforce the restriction, retain refusal and
record the vendor or normative change needed; do not simulate enforcement.

## Qualification procedure for the installed build

This is future work in a separate tracked qualification change. Do not run
vendor sessions or alter the pin merely to complete this design task.

1. Freeze a private executable copy; record its release, SHA-256, platform,
   architecture, harness/source commit and launch recipe. Check the digest
   before and after the run. Any mismatch invalidates the run. Review an
   explicit harness update for 2.1.309: both digest and version assertions,
   collector identity, fixtures and expected-red controls must agree. Preserve
   the 2.1.305 record as historical evidence. A skipped identity probe is not
   qualification.
2. Reuse the harness's disposable ancestor/project tree, separate Git metadata,
   synthetic HOME/config/XDG state and system-only PATH. Run the frozen binary
   through `sandbox-exec` and an explicit environment allowlist. Permit only
   the exact synthetic loopback SSE provider, fixture writes and required
   synthetic stdio MCP processes. Deny other network, real native config,
   Keychain/security services, managed-system discovery and ancestor inputs
   outside the synthetic tree. No real credentials, login, external inference
   or reads of operator-native state. Revalidate this containment on 2.1.309.
3. Define foreign markers and forbidden effects first. For every row below,
   start with a fresh discovery-enabled control proving the planted native
   source activates. Run a fresh strict counterpart and a nearby approved
   positive control. Preserve identical relevant fixtures and fixed provider
   tool requests. Require successful request/init/terminal observations and
   the expected native tool verdicts, not model statements about obedience.
4. Run each row in three independent fresh fixture pairs, including both
   orders of baseline and strict. Every repetition must satisfy its exact
   assertions; one violation fails the surface. Give each process a 30-second
   deadline and each replay group at most four minutes; split groups before
   that limit. If a valid native control needs a longer bound, revise and
   record the bound before collecting qualification evidence. A timeout,
   crash, missing turn or unrelated permission error is a failed probe.
5. For late mutations, synchronize on an observed completed first turn, create
   the fixture change, invoke the exact supported reload/event, and observe a
   second native request/tool completion. For headless multi-turn probes, use
   the measured stream input/output transport (`--input-format stream-json`,
   `--output-format stream-json`, `--verbose`) and bind events to their native
   turn/session IDs. Qualify that transport as part of the proposed form.
   Use a third turn where deferred
   application is possible. Record event acknowledgements and timing. An
   arbitrary sleep or absence on one turn, including the old 800ms hook check,
   does not establish that a loader was exercised. If the foreign positive
   never activates, mark that route unqualified.
6. Run each row's narrowing control, one omitted/weakened gate at a time,
   against unchanged negative assertions. Require the intended foreign marker,
   native effect or forbidden transition to produce the expected assertion
   failure, while the approved neighbor succeeds. A skipped plant, setup
   failure, containment denial or generic nonzero exit is not a valid mutant.
7. Validate collection/publication: completed nonempty run required, ordinary
   and encoded private-locator plants rejected, sanitized identity and request/
   effect evidence retained. Publish only reviewed synthetic excerpts and
   receipts; keep no real paths, session identities or credentials in public
   artifacts. Record exact commands, exits and limitations for every group.

### Negative-first matrix

All rows inherit the paired control, three repetitions and intended-failure
rules above. "Remove exclusion" means a synthetic discovery-enabled run;
it never means weakening the real-host containment around a vendor process.

| ID / claim to prove | Forbidden outcome and observable witness | Nearby positive control | Narrowing control / coverage limit |
| --- | --- | --- | --- |
| Q1 startup instructions | Foreign root, ancestor, `.claude/CLAUDE.md`, local instruction, rules or AGENTS fallback marker in actual request bytes | Each planted source reaches baseline requests; managed profile root/system and approved additive project context reach strict requests in required order | Remove source exclusion; exercise each source independently, with AGENTS fallback lacking CLAUDE.md |
| Q2 nested traversal | Read of ordinary nested data triggers foreign nested CLAUDE, AGENTS or path rule injection | Same ordinary file contents returned in both recipes; managed instructions retained | Remove exclusion; test Read and each other enabled traversal tool separately, including late-created instructions |
| Q3 settings | Project/local model, permission, environment or tool-setting atoms change manager behavior | Baseline changes requested model and a real permitted tool result; strict retains manager model/grant/deny with real file effects | Remove source gate for each settings location; conflicting allow and deny cases, not just one Read restriction |
| Q4 skills, agents, commands | Foreign catalog metadata/body or executable command becomes available at startup or after explicit reload | Baseline foreign invocation works; approved whole skill and any supported managed agent/command actually run | Remove exclusion or expose the foreign root; include `.claude/skills`, compatibility roots, nested roots and `/reload-skills`; other reachable roots need their own witness |
| Q5 hooks | Foreign SessionStart or tool-event hook process starts or writes a marker | Baseline hook starts and writes; synthetic managed hook works where that channel is supported | Remove exclusion; distinguish startup from after-turn creation/modification plus an acknowledged reload/event |
| Q6 plugins | Foreign plugin catalog, hook, skill or process loads, including local marketplace registration | Baseline registered fixture plugin loads with a real hook/skill witness; managed fixture plugin still works where supported | Expose project plugin configuration; test startup and install/enable/update/reload routes reachable in the proposed form; absent fixture registration proves nothing |
| Q7 MCP | Unapproved server starts, connects, contributes tools or executes a call | Baseline project server starts/connects; approved explicit server initializes and permitted policy is preserved; empty explicit file connects to no server | Cross project empty/nonempty, sources all/user, strict flag off/on, approved empty/nonempty (16 cells); omit strict MCP in the discriminating all-source row |
| Q8 late settings/hooks/plugins/MCP | After-turn mutation, enablement or reload introduces any forbidden setting/effect from Q3/Q5/Q6/Q7 | Every claimed reload mechanism has an activated foreign baseline and a working managed counterpart; old/new approved states distinguished | Remove the relevant source boundary or bypass an enforced reload refusal; untriggered or unknown reloads remain unsupported, even if fresh startup passed |
| Q9 memory off | Seeded automatic memory reaches requests or automatic writes occur in native memory stores | Enabled-memory fixture proves load and write separately; disabled recipe still reads permitted ordinary data and managed instructions | Remove memory-off control; test each discovered synthetic scope and late/child path; a log saying memory is off is insufficient |
| Q10 manager default deny | Empty/exact built-in ceiling or MCP deny is bypassed by project allow/deny settings | Narrow Read grant returns real bytes, neighbor deny is observed; approved MCP connects while calls remain denied | Omit tool ceiling or weaken MCP deny separately; attempt Read/Write/Bash and synthetic MCP call, with sandbox-permitted fixture effects so native policy is the discriminator |
| Q11 imports, aliases, extra roots | Native import follows foreign or inventoried target; aliases/additional roots regain control discovery | Approved rendered context without imports works; ordinary prose references stay opaque | Bypass import rejection or expose each alias/root; accepted import syntax must be rejected, not merely ignored by a broken fixture |
| Q12 fresh versus history | Any resume/fork/import/recovery entry launches, allocates fallback session or loads seeded old history | Manager fresh launch succeeds without history; native unsafe-origin resume/fork baseline visibly retains seeded text on this exact release | Bypass typed entry gate; test CLI aliases, equals/short forms, existing IDs, in-session history commands and unknown entries; even safe-origin resume stays refused |
| Q13 children and form boundaries | A child, background or remote route drops exclusion, permissions, memory control or lease ownership | If admitted, actual child request/completion preserves approved inputs and lease; native baseline activates foreign child inputs | Drop child propagation or bypass form refusal; test every reachable Task/agent mode and reload path separately, or prove native child entry unavailable; old one-child evidence does not qualify all modes |
| Q14 authority, receipts and leases | Changed executable/control/source, lost authority, busy home, interrupted transfer or stale generation starts a process or partially publishes | Valid exact composition launches once with active reserved lease; authoritative termination releases it | Omit each validation in an isolated consumer fixture; assert no process/MCP/partial writes on refusal and no automatic new-home fallback; unreadable evidence remains unreadable |

Test comments must state these bounded claims next to the future executable
cases, including materially different per-case limits. Retain the prior
unsafe-history regression witness and its neighbors on the new binary.
Containment controls remain separate: a second synthetic loopback destination
and synthetic sensitive-file sentinel prove that the outer boundary catches a
known violation without ever running the vendor under weakened host protection.

### Replay and acceptance evidence

After the reviewed harness update, use the existing entry points with a frozen
binary and fresh roots. These commands describe a future run, not checks
performed for this note:

```sh
# Set CLAUDE_EXCL_BINARY to the private frozen copy and
# CLAUDE_EXCL_RUN_ROOT to a fresh task-local absolute .temp/ path.
# Use private Go caches/module state and the configured pinned CURATOR_GO.
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test \
  ./.research/project-exclusion-claude -count=1 -v -timeout=4m
# Use another fresh run root. Run through the lane's suite slot for batches.
sh .research/project-exclusion-claude/mutants.sh
python3 .research/project-exclusion-claude/collect.py \
  "$CLAUDE_EXCL_RUN_ROOT" .temp/strict-claude/observations.json
```

The later harness must expose bounded groups for the expanded matrix and fail
if any required row/repetition is skipped. Record queue time separately from
held execution when suite coordination applies. Full repository regression,
if required for that later code change, runs at its exact handover head in a
normal disposable validation checkout with a short temporary-directory path.

A qualification outcome must bind the frozen binary hash/version, OS release
and architecture, source/harness hashes, mode, form, entry and allowed tool/
child set, config/policy/materialization digests, event sequences, all row IDs,
repetitions, commands, exits, expected-red diagnostics and publication checks.
Separate native observations from consumer integration tests. No surface is
certified from a model answer, help text, schema field, opt-in skip or absence
of an activated baseline. All mandatory reachable surfaces and the actual
consumer boundary must pass before a protected adapter advertises strict.

## Adapter and consumer work after qualification

These are required follow-up changes to TASK-261022-sg5y2x and its consuming
launcher, not implementation in this task:

| Area in the pinned adapter candidate | Required change and acceptance boundary |
| --- | --- |
| `ClaudeProjectVersion`, `ClaudeCertificationReport`, `RequireStrict` | Select reviewed evidence for the exact release/platform and privately bound binary/architecture/form. Keep strict refused until all mandatory surfaces pass. Do not merely bump the version constant or replace the unconditional refusal. Historical `excluded-origin-resume` observations never enable resume in this slice. |
| `PrepareClaudeMode` and `ClaudeLaunchMode` | Implement strict preparation using protected composition's additive profile/project outputs, exact mode and full receipt; keep native-project's profile-only recipe separate. Validate structured fresh entry before publication; no executable contribution for unsupported routes. |
| `ClaudeEntry`, `Authority.ClaudeLaunch`, actual launcher parser | Retain explicit fresh-only semantics at both construction and final consumer. Replace raw native passthrough with owned argv/environment, reject all restoration/form bypasses, reserve unused native identity and test real launch behavior. An offline contributor check alone is insufficient. |
| `ClaudeControls`, `claudeOutputs`, `Environment`, MCP contribution | Apply the measured permission ceiling, deny rules, explicit empty/nonempty MCP and memory controls unchanged through every admitted lifecycle path. Preserve approved channels and credential ownership; MCP connection does not grant tool calls. |
| `PublishClaude`, receipt verification and session/lease integration | Re-resolve current protected authority under lock, verify runtime/control identities before start, publish transactionally through the existing lease owner, transfer prepared to active before exec and retain unknown state. No new writer fallback or automatic repair. |
| Certification/status and carrier integration | Negotiate a project-capable outer carrier before authority lookup. Keep frozen carriers closed. Publish only the existing `project-adapter/1` members and surface vocabulary; keep detailed form/digest evidence in protected implementation records unless a normative extension is accepted. |

The public adapter record has `kind`, `environment`, `provider_version`,
`platform`, `modes`, `surfaces`. It does not allocate `fresh`, `form`, digest or
new refusal-code fields. A first-slice form restriction must be enforced and
unambiguous to every consumer; if existing negotiation cannot express it
safely, obtain normative publication before advertising strict. Internal
`ClaudeDiagnostic.Refused` details can describe individual gaps without new
wire codes. Existing credential/version diagnostics retain their owners.

## Curator status and refusal behavior

Status remains read-only: report selected mode, admitted generation, candidate
drift, home/lease ownership, unknown/unreadable state and exact qualification
gaps without fetch, repair, approval or baseline changes. Until qualification,
strict status must communicate `project_context_unsupported`, the observed
runtime tuple and missing surfaces/forms, not a ready strict launch. Resume
refusal should explain that the operator must explicitly request a new fresh
session; never silently create one. If the operator explicitly selects
native-project, retain its exact notice `project sources not isolated` in
status and launch output, including admission-off cases.

Use the existing section 10.6 codes, without allocating exit numbers or
extending a closed error object:

| Condition | Existing diagnostic |
| --- | --- |
| Well-formed resume/fork or other unsupported form; missing exact qualification, permissions, memory-off or exclusion | `project_context_unsupported` |
| Malformed/unknown/inconsistent mode, admission, bindings or structured entry value | `project_context_invalid` |
| Checkout/CWD/worktree no longer matches registration | `project_registration_mismatch` |
| Missing/revoked/wrong-item approval or insufficient authority | `project_item_unapproved` |
| Changed/missing protected source, policy, adapter contract, control receipt or pinned executable bytes | `project_context_drift` |
| Required authority, executable identity or liveness evidence cannot be read | `project_context_unreadable` |
| Update/repair/switch attempts with active or unknown home lease | `project_view_busy` |
| Wrong session/owner/generation, stale lease or replayed transition | `project_lease_mismatch` |
| Unsafe/missing/excluded PATH or wrong executable selection | `project_path_refused` |

An installed version with no certified contract is unsupported; modification
of bytes bound to an existing protected contract is drift. Failed reads retain
the unreadable distinction. An unsupported outer-carrier version uses its
published version refusal before project lookup. Every refusal preserves
existing home/approval state and emits no success fragment, child, MCP
connection or partial publication; telemetry is redacted.

## Owner decisions to resolve before implementation/admission

1. Accept headless fresh sessions as the first strict form, leaving interactive,
   background, remote and all restore forms refused? Qualification must follow
   the actual launch form, not an assumption of interactive equivalence.
2. Which built-in tools and native child/agent paths must the first form admit?
   A smaller set is viable only with measured native enforcement. Unavoidable
   unqualified child or reload routes keep strict refused; this is not a waiver
   of lifecycle exclusion, manager permissions or memory-off.
3. If late hooks/settings/plugins/MCP cannot be activated and excluded on this
   release, should the work wait for a vendor fix or pursue a separate proposed
   consumer/runtime restriction with executable evidence? No silent relaxation
   of strict is available.
4. Can the existing carrier and protected adapter contract safely bind and
   communicate fresh/headless-only qualification to all consumers, or is a
   normative form negotiation extension required before advertisement?

These questions are also indexed in [unresolved questions](../../UNRESOLVED_QUESTIONS.md).
Until resolved and qualified, the proposed contract remains a plan and strict
launch remains refused.
