# Unattended lane launch and pre-approved project trust

- **Status:** Design note, proposed. No implementation is delivered or claimed.
- **Tasks:** TASK-261026-p2itpj (launch parameters), TASK-261026-9rlqfj (project trust and threat model).
- **Decision:** operator. Section 4 lists the open questions.

This note defines the launch profile for unattended agent lanes and an
owner-approved policy that pre-approves Codex project trust and hook trust.
It closes with the threat model for both. It is written against these
baselines, and every "exists" claim below refers to them:

| Component | Baseline |
| --- | --- |
| Curator | `main` at `77f8d047` |
| curator-spec | [`6c7ce403`](https://github.com/relux-works/curator-spec/tree/6c7ce40301e254e14e55ebd29405c90fe00c79c6) |
| curator-run (`curator-agent-launcher`) | [`60143fec`](https://github.com/relux-works/curator-agent-launcher/tree/60143fec205e21a7b7ec98674f6e48cd646db22a) |
| Codex CLI | 0.163.0, observed through `--help` and binary strings only |

Spec and launcher paths below are relative to those two repositories at those
commits. Neither the launcher nor the spec was executed for this note.

## Terms

A lane is one agent work stream with its own identity, checkout and task
queue. An orchestrator lane coordinates a project; a worker lane implements
or reviews one task at a time. CIP-0024 calls the durable form of a lane a
station (`cips/CIP-0024-agent-stations.md`); this note uses "lane" because it
also covers today's hand-started sessions.

An unattended lane runs without a human at its terminal. Nobody answers an
approval dialog, a trust prompt or a login request during its session.

An owner-approved lane is a lane whose launch profile the operator has
approved, including its permission mode. Approval belongs to the operator
alone; an agent or a relayed message cannot grant it.

A front door is the entry that starts a lane's provider process. Section 1.7
maps the launch profile onto three of them.

## Current state

Two Codex orchestrator lanes start from a hand-written script. In substance
it runs the following:

```sh
cd <project checkout>
exec env -u SSH_AUTH_SOCK -u A2A_SESSION_ID \
  A2A_AGENT=<lane identity> \
  CODEX_HOME=<curator-managed codex_cli home> \
  PATH=<fixed list> \
  codex -m <model> -c model_reasoning_effort=high \
    --dangerously-bypass-approvals-and-sandbox \
    -c shell_environment_policy.inherit=core \
    "<lane brief>"
```

The script removes two inherited variables, names the lane, points Codex at a
Curator-managed home and disables every approval and sandbox control.
Worker launchers add an explicit `shell_environment_policy.set.*` list and
scrub secret variables before the spawn.

Each lane carries its own copy of these decisions. Nothing records which
lanes may bypass the sandbox, which variables were removed, or why. The
Codex project trust prompt and hook trust prompt remain a human step. An
agent must never answer either prompt through terminal keystrokes, so a lane
whose checkout is not yet trusted stalls until the operator intervenes.

## 1. Launch profile

The launch profile is the complete, reviewable set of parameters that start
one lane. Each parameter below has a stated reason and an owner. The goal is
that a lane launch is reproducible from a record, and that a hand-written
script is no longer needed.

| Parameter | Value for an owner-approved lane | Owner of the value |
| --- | --- | --- |
| Permission mode | `yolo` (Codex: approval and sandbox bypass) | Operator, per lane |
| Shell environment policy | `inherit=core` plus an explicit `set` list | Operator, per lane class |
| Process environment | Allowlisted base; operator SSH agent and provider tokens removed | Launcher |
| Git identity and signing | Human author name and email; agent signing key through a dedicated agent socket | Operator; key custody per CIP-0024 |
| State directories | Per-lane `TMPDIR`, build caches and launcher state; Curator-managed `CODEX_HOME` | Launcher and Curator |
| Model and effort | Explicit values, admitted by machine locks | Operator, per lane |
| Brief | Lane brief as the initial prompt | Raising orchestrator or operator |

### 1.1 Approval and sandbox mode

An unattended lane cannot answer approval dialogs, so it runs either with
approvals off inside a sandbox or with both bypassed. Today's lanes bypass
both, because builds, tests, Git signing and a2a delivery need the network,
the agent socket and paths outside the checkout.

curator-run already exposes this choice. `--permissions yolo` and its exact
alias `--yolo` select the bypass mode, and agents-management owns the
release-specific Codex spelling ([Decision 0018](https://github.com/relux-works/curator-spec/blob/6c7ce40301e254e14e55ebd29405c90fe00c79c6/decisions/0018-curator-run-permission-interface.md),
choices 1 and 6). A fleet-wide force-`native` lock refuses any visible
`yolo`. `--dangerously-bypass-hook-trust` is never mapped (Decision 0018,
Compatibility and security impact). This note keeps that rule: hook trust is
handled by section 2, never by a bypass flag.

The resolved default differs from Decision 0018's text. The launcher SPEC
records operator decision Q-D3 (2026-10-05): untracked silence resolves to
`yolo` on every stdio shape, and hosted launches follow the same ladder
(`SPEC.md` §4.3 item 4 and §4.6; `cmd/curator-run/main.go`,
`resolvePermission`). Decision 0018 still says headless silence resolves to
`native`; the SPEC marks the matching amendments as pending. The consequence
for this note is direct: on a host without the force-`native` lock, every
lane that names no permission mode runs with the sandbox bypassed.

Proposed rule: a lane may run `yolo` only when an owner-approved lane record
names it. Hosts that run unattended lanes set the launcher-global default
to `native`, so silence never bypasses. The lane record then carries
`permissions: yolo` with the approving operator and time. A flag on the
command line does not carry approval, because any process that can invoke
curator-run can type it. Enforcing the rule against a same-UID agent needs
a protected record (section 2.1); until then it constrains cooperative
launchers only.

### 1.2 `inherit=core` and the explicit set list

Codex applies `shell_environment_policy` to the commands its shell tool
runs. It does not change the environment of the Codex process itself. Two
layers therefore exist, and the launch profile controls both:

- The process environment holds what the Codex process and its own child
  services see. Section 1.3 controls it.
- The shell tool environment holds what model-issued commands see. The
  shell policy controls it.

`inherit=core` starts the shell tool environment from Codex's built-in core
set of basic variables; its exact membership belongs to the per-release
qualification. Every other value a command needs arrives through an explicit
`shell_environment_policy.set.<NAME>` entry. Worker lanes set these:

| Name | Value | Reason |
| --- | --- | --- |
| `HOME` | Lane home | Tools that read `~` stay inside lane state |
| `TMPDIR` | Directory inside the worker directory | No temporary files shared with other lanes |
| `GOCACHE` | Per-run cache directory | No cache poisoning across lanes or runs |
| `GIT_AUTHOR_NAME`, `GIT_AUTHOR_EMAIL`, `GIT_COMMITTER_NAME`, `GIT_COMMITTER_EMAIL` | Configured human author | Commits keep the human author (section 1.4) |
| `GIT_CONFIG_COUNT` with `user.signingkey`, `gpg.format`, `gpg.ssh.program` | Agent signing public key and SSH signer | Signing uses the agent key, never the operator's interactive signer |
| `GIT_SSH_COMMAND`, `SSH_AUTH_SOCK` | Dedicated agent socket and key | Push and signing reach only the lane's agent socket |

A `set` entry assigns a value; it never appends. CIP-0002 records the same
fact for `PATH` (`cips/CIP-0002-project-context-in-managed-launches.md`,
D-PATH). A `set.PATH` entry would therefore replace the launcher's checked
PATH transform, and the launch profile must not contain one once
`checked-d-path/1` is in use (`protocol/project-context.md` §7).

The shell policy is a Codex setting with no Curator owner today. Profile
packages cannot supply it: the profile Codex settings allowlist excludes
sandbox, approval, project trust and hook keys
(`protocol/profile-codex-settings.md` §1). Proposed: the lane launch profile
carries the shell policy as launcher-emitted `-c` pairs, owned by the
operator per lane class, and qualified per Codex release like the
permission mapping.

### 1.3 Environment scrubbing

The process environment of a lane must not contain the operator's
credentials. Two classes matter most. The operator's `SSH_AUTH_SOCK` gives
push and signing rights under the operator's identity. Provider tokens
(`OPENAI_API_KEY`, `CODEX_API_KEY`, `ANTHROPIC_*` and similar) give direct
model and account access. A stale `A2A_SESSION_ID` from the parent session
also misroutes a2a delivery, so the script removes it; the session manager
sets the correct value once the native session exists (`FRONT-DOOR.md` §6.3).

Scrubbing by denylist fails open: a secret with an unlisted name passes.
The launch profile therefore starts from an allowlisted base and adds the
section 1.2 values. The allowlist names `PATH`, `HOME`, `USER`, `LOGNAME`,
`SHELL`, `LANG`, `LC_*`, `TERM` and the values the launcher itself owns.
The FRONT-DOOR secret check stays as a second guard on what a record adds
(`FRONT-DOOR.md` §5.3), and its own text calls it "a guard against
mistakes, not a secret classifier".

Today no front door scrubs. Direct curator-run passes `os.Environ()` to the
provider in both modes (`internal/plan/plan.go`, `Request.Env`). The hosted
Werkstatt route adds the record's `child_env` after the secret check and the
reserved-name list, which refuses `CODEX_HOME`, `OPENAI_API_KEY`,
`CODEX_API_KEY` and the XDG names (`FRONT-DOOR.md` §9.2). How the session
manager builds the provider's base environment is the manager's contract and
was not verified for this note. The target one-shot exec route resets the
environment in the root launcher before it drops privileges (`README.md`,
Target one-shot exec mode).

### 1.4 Agent Git identity and signing

A lane commits under the configured human author name and email, and signs
with an agent key. The author fields keep repository history attributed to
the accountable human. The agent key makes every agent commit verifiable
and distinguishable from the operator's own signatures, because
`allowed_signers` lists the agent key separately.

The signing key stays in a dedicated SSH agent. The lane receives only the
socket path and the public key path through `GIT_CONFIG_COUNT` and
`GIT_SSH_COMMAND` (`FRONT-DOOR.md` §6.3). A commit key never signs release
tags; release signing uses a separate key named by `RELEASE_SIGNING_KEY`.
CIP-0024 decides the custody target: one SSH agent per station, a key file
the station cannot read, a socket only its UID can open, and later a signing
service that keeps keys out of stations entirely
(`cips/CIP-0024-agent-stations.md`, owner decision 1). A lane that shares
the operator's UID does not meet that custody rule. Section 3 records it as
a residual risk.

### 1.5 Per-lane state directories

Each lane owns its temporary directory, build caches, launcher state and
a2a identity. Shared directories let one lane read another's intermediate
files, reuse a poisoned build cache or delete state another lane still
uses. The section 1.2 `set` list covers `TMPDIR` and `GOCACHE`; Werkstatt
records name `GOCACHE` and `GOPATH` under the Werkstatt directory
(`FRONT-DOOR.md` §6.3). Launcher state for each Werkstatt lives in its own
record, journal and lock files (`README.md`, Werkstatt lifecycle).

`CODEX_HOME` comes from Curator's fragment and is not a lane choice. Today a
managed home exists per profile and environment, so two lanes using the same
profile share one Codex home, its sessions, its trust table and its login.
The Werkstatt home of CIP-0002 is allocated per checkout, profile and
environment under `environments/project-views/` (`protocol/project-context.md`
§5; [project leases](../project-leases.md)). That allocation gives each lane
its own native home once the native-project or strict route is supported.

### 1.6 Model and effort

The script spells the model and effort as native Codex flags. The launch
profile carries them as values instead. curator-run admits `--model` and
`--effort` against machine locks (`SPEC.md` §4.3), and a Werkstatt record
pins both at create or restart (`FRONT-DOOR.md` §6.1). A per-task effort
travels with the task. Values let the launcher refuse an unadmitted model
before start, and let a status read show what each lane runs.

### 1.7 Front doors

The same launch profile maps onto three front doors. "Exists" means the
behavior is present at the baseline commits; "proposed" means this note asks
for it; "target" means an accepted design that is not implemented.

| Parameter | Direct `curator run` | Hosted lane (`curator-run werkstatt up`) | Dispatcher one-shot exec |
| --- | --- | --- | --- |
| Status of the door | Exists | Exists; delegated route signed off for Claude and Codex (`FRONT-DOOR.md` §5.8) | Target, not implemented (`SPEC.md` §10) |
| Permission mode | Exists: flag, profile, global default; silence resolves `yolo` (Q-D3). Proposed: global `native` default and lane-record approval | Exists: same ladder, no per-record member. Proposed: `permissions` member in the lane record | Target: the plan carries the mode; the executor never infers a default |
| Sandbox | Bypassed by `yolo` | Exists: record `sandbox` is `none` or `workspace-write` | Target: separate agent account and zero egress (R-NET-TOOL) |
| Shell environment policy | Proposed: launcher-emitted `-c` pairs | Proposed: same pairs in the hosted request | Proposed: same pairs in the sealed plan |
| Process environment | Exists: full caller environment inherited. Proposed: allowlisted base | Exists: record `child_env` with secret check and reserved names; manager base not verified | Target: root launcher resets the environment |
| Git identity and signing | Proposed: from the lane record | Exists: through `child_env` (`FRONT-DOOR.md` §6.3) | Target: per-account keys; signing service later |
| `A2A_AGENT`, `A2A_SESSION_ID` | Caller-supplied | Exists: launcher sets `A2A_AGENT`; manager sets `A2A_SESSION_ID` | Target: per-station identity |
| State directories | Proposed: per-lane `TMPDIR` and caches | Exists: per-Werkstatt launcher state; caches through `child_env` | Target: per-account home |
| `CODEX_HOME` | Exists: Curator fragment, per profile | Exists: Curator fragment; reserved from `child_env` | Target: Werkstatt home per station |
| Model and effort | Exists: `--model`, `--effort` with admission | Exists: pinned in the record | Target: in the plan |
| Project trust and hook trust | Proposed: section 2 | Proposed: section 2 | Proposed: section 2, checked before the plan is sealed |

The hosted lane is the intended home for orchestrator lanes. It already owns
liveness, resume, doorbells and the record that a launch profile needs.
The proposals in this table add the missing members to that record rather
than introduce a second launch path.

## 2. Pre-approved project trust

Codex asks before it trusts a project directory, and asks again before it
runs a project hook it has not seen. Both prompts protect the operator; both
stall an unattended lane. A pre-approved trust policy lets the operator
answer them ahead of time for named repositories, and leaves them in place
for everything else.

This design builds on the workspace trust ledger of
[CIP-0020](https://github.com/relux-works/curator-spec/blob/6c7ce40301e254e14e55ebd29405c90fe00c79c6/cips/CIP-0020-workspace-trust-ledger.md),
whose A1 portion is adopted in `protocol/environments.md` §7.10 and
`decisions/0026-workspace-trust-ledger-and-bindings.md`. It reuses the
ledger's record fields, repository identity tuple, no-follow admission and
per-key projection journal, and does not restate them. Curator implements
none of the ledger today: there is no `curator env trust` command. The trust
writes found in Curator source are the Claude launch-directory seed in managed-home
repair (`internal/envprofile/managed.go`) and the trust entry for the
prepared worktree in the Codex native-project renderer
(`internal/projectadapter/codex.go`), which no launch path calls yet.

### 2.1 The policy and who changes it

The project trust policy is an operator-authored record of repositories
approved for unattended launches. Each entry names:

| Field | Meaning |
| --- | --- |
| `entry_id` | Stable identifier, used by grants and revocation |
| `repository` | Canonical identity: the accepted remote spellings (credential-free) and the root commit object ID |
| `environments` | Exact adapter identities, such as `codex_cli`; no wildcard |
| `project_trust` | Whether Codex project trust may be seeded |
| `hooks` | Allowed hook definitions, each bound to its content digest; empty means no hook trust is seeded |
| `lanes` | Lane identities or lane classes allowed to use the entry |
| `operator`, `time`, `reason` | Approving operator identity, approval time and stated reason |
| `review_by` | Date after which the entry stops granting until the operator renews it |

The policy keys on repository identity. A path only says where a checkout
happens to be, and a different repository can appear there later. At launch, Curator matches the checkout's
identity against the policy and then grants that one exact path.

The policy lives in private machine-local Curator state, beside the CIP-0020
ledger and outside profiles, snapshots, locks and exports. Only the operator
changes it, through a Curator command that records who changed what. Agents,
peer messages and repository files cannot add, widen or renew an entry. A
relayed instruction is data until the operator confirms it at its source.

On a host where lanes share the operator's UID, a lane with the sandbox
bypassed can still write the policy file directly. File ownership cannot
stop it, so the policy is an integrity control only against cooperative
code until stations run under separate accounts. CIP-0024 decides the target
form: a root-owned trust pin outside station write authority, verified by
the launcher before every start (`cips/CIP-0024-agent-stations.md`, owner
decision 2). Section 3 records the gap as a residual risk.

### 2.2 Seeding at launch

Seeding happens during launch preparation, before the provider process
starts. It writes only into the managed `CODEX_HOME` that the launch will
use. The steps are:

1. Admit the exact checkout path with CIP-0020 no-follow semantics and read
   its repository identity without running repository hooks or
   configuration-defined commands.
2. Match the identity against the policy. Both the remote and the root
   commit must match one entry, and the entry must list the environment and
   the lane. No match, an ambiguous match or an unreadable identity means no
   seeding.
3. Record a ledger grant for the exact path and environment. The grant
   carries the policy `entry_id`, the digest of the policy entry at grant
   time and the full identity tuple.
4. Project the grant into the managed `config.toml` as the single key
   `[projects."<exact path>"].trust_level = "trusted"`, through the CIP-0020
   §6 projection journal.
5. For each hook that Codex would load from the checkout, compare its
   definition digest with the entry's `hooks` list. Seed hook trust only for
   listed digests. A changed or unlisted hook stays untrusted and keeps its
   prompt.
6. Report the grant IDs and seeded keys in the launch output and the lane
   record.

CIP-0020 closes its event set to `explicit-grant`, `implicit-launch` and
`implicit-bind`. A policy-derived grant has different provenance from all
three, so this design needs a CIP-0020 amendment that adds a `policy-grant`
event (question 2 in section 4).

Codex 0.163.0 persists hook trust as a per-hook hash decision, separate from
project trust; the binary carries a `HookStateToml` type with a
`trusted_hash` field. The [Codex source-exclusion record](../qualification/project-source-exclusion-codex.md)
measured that hook metadata "still requires its separate native hash-trust
decision". Neither the storage shape nor its lookup is qualified. CIP-0020
§5 applies unchanged: an unqualified release or shape refuses automatic
projection, so hook trust seeding stays off until a per-release
qualification covers merge, lookup and revocation.

### 2.3 Recording, revocation and audit

Every seeding leaves three records: the ledger grant, the per-key journal
entry in the managed home and the line in the launch output. The ledger and
journal are the authority; the launch line is a convenience for whoever
reads the lane log.

Removing or editing a policy entry revokes every grant derived from it.
Curator then removes only the keys it inserted and still owns, as CIP-0020
§6 requires. A key that changed after insertion is an ownership conflict:
Curator keeps it, revokes the ledger authorization and reports the vendor
cleanup as pending. Revocation does not stop a running session; the next
launch preparation refuses.

A grant also lapses on its own when the identity tuple at the path changes
or `review_by` passes. Resolve and repair never renew a lapsed or revoked
grant. A read-only listing shows policy entries, their grants, the managed
homes each grant touched and any pending cleanup. An unreadable policy,
ledger or journal is an error, never an empty result.

### 2.4 Unknown projects keep the human step

A checkout with no matching entry gets no seeded trust. For an attended
launch, Codex shows its own prompt, and the operator answers it. For an
unattended launch, preparation refuses with CIP-0020's `workspace_untrusted`
diagnostic and its explicit-grant hint, before any provider process starts.
A refusal is visible in the lane's status; a session blocked on a prompt that
nobody sees is not. No agent answers a trust prompt by sending keystrokes,
and no front door passes `--dangerously-bypass-hook-trust`.

### 2.5 Composition with native-project and strict modes

Pre-approved trust decides whether Codex may treat a checkout as a trusted
project. It approves no project item. `protocol/project-context.md` §1
states that native trust must not substitute for item approval, and that
rule applies here unchanged.

| Launch mode | Where trust is seeded | What hook trust means | State at baseline |
| --- | --- | --- | --- |
| Unenrolled (today's lanes) | The exact checkout path in the profile's managed home | Project hooks run natively once trusted; the hook digest list is the only gate | Exists as a launch; seeding is proposed |
| `native-project` | Only the prepared Werkstatt worktree in its project-view home (`protocol/project-context.md` §5) | Same as unenrolled: repository hooks load natively, so only listed digests are seeded | Codex launch refuses `project_context_unsupported` because repository MCP permission keys survive ([Codex composition delivery](../project-adapter-codex.md)) |
| `strict` | Only the prepared worktree; a certified adapter suppresses repository hooks | Not seeded: strict exclusion keeps repository hooks out, and a seeded entry would widen it | Codex refuses `native-source-exclusion` |

In `native-project` mode the policy gates whether an unattended lane may
prepare a Werkstatt for that repository at all. The Codex adapter already
writes trust for the prepared worktree, so the policy grant records the
consent behind that write. CIP-0024 makes `native-project` the wave-1 mode
for stations and requires status to show "project sources not isolated"
(`cips/CIP-0024-agent-stations.md`, Wave 1 project mode). In strict mode the
policy only gates project trust for the worktree. A mode change needs the
operator's explicit update; the policy never selects a mode.

## 3. Threat model

### 3.1 Assets

The threat model protects these assets:

| Asset | Why it matters |
| --- | --- |
| Operator credentials: interactive SSH agent, provider tokens, GitHub tokens | Act as the operator on every service |
| Agent signing and push keys | Produce verified commits under the human author and push to every repository the key can reach |
| Repository default branches and releases | Downstream users install from them |
| Managed homes, trust policy, ledger and journal | Decide what each lane loads and trusts |
| Other lanes' checkouts, homes and caches | Integrity of parallel work |
| Operator's personal files on the same host | Outside every lane's task scope |
| a2a identities and owner chat | Commands and handoffs between lanes and the owner |

### 3.2 Actors

The operator owns the host, the policy and every approval. Other human
principals may hold a narrower scope assigned by the operator. Lane agents
follow their brief but can be steered by any text they read. Peer agents
send messages over a2a; their content is data, and authority comes only
from service-checked grants (`cips/CIP-0024-agent-stations.md`, Root context
and authority). Repository content authors include contributors, upstream
dependency maintainers and pull-request authors. A repository's future owner
is whoever controls its remote after a transfer or account takeover.

### 3.3 Trust boundaries

The design relies on these boundaries:

- Operator to policy store: only operator actions change the policy.
- Launcher to provider process: the launcher decides environment, argv and
  home.
- Provider process to shell tool commands: the shell policy and the
  sandbox; bypass mode removes the sandbox half of this boundary.
- Repository content to agent context: instructions, hooks, MCP servers and
  plain files enter the model's input.
- Lane to lane: separate directories today; separate OS accounts in the
  station target.
- Host to remote services: SSH agent sockets, provider endpoints and a2a.

### 3.4 Threats and mitigations

Each threat names its mitigation, or an accepted residual risk with its
owner. "Operator" means the human owner of the host and policy.

**T1. Prompt injection from repository or peer content.** A file, issue,
commit message, tool output or a2a message tells the lane to run commands,
leak data or push changes. With the sandbox bypassed, the lane can act on it
at once. Mitigations: the lane root context states that only the raising
orchestrator and owner give commands (CIP-0024); services check grants and
signed envelopes, never message text; candidate code from others reaches
reviewers as inert diffs (`cips/CIP-0002-project-context-in-managed-launches.md`,
R-TRANSFER); pushes land only through signed review and green checks.
Residual: prose cannot contain injection, and a lane that obeys injected
text acts with its full section 3.5 capability. Owner: operator, accepted
per approved lane.

**T2. Malicious or changed hooks in a trusted project.** A commit adds or
edits a hook in a repository whose project trust is pre-approved. Codex runs
project hooks once the project is trusted and the hook is trusted.
Mitigations: hook trust is seeded only for listed definition digests, so a
changed hook keeps its prompt and an unattended launch refuses; no front
door maps `--dangerously-bypass-hook-trust`; strict mode does not seed hook
trust at all. Residual: an approved hook can still run code that changes
between launches, for example a script it calls; the digest covers only the
hook definition. Owner: operator, who approves hook
entries.

**T3. Trust seeded for the wrong repository through a path or remote
alias.** A different repository appears at an approved path, or a fork,
mirror or `insteadOf` rewrite makes a foreign remote look approved.
Mitigations: the policy keys on remote and root commit together, so a
replacement repository at the same path fails the root commit and a fork
with the same root fails the remote; CIP-0020 no-follow admission refuses
symlinked paths and rechecks identity before each use; the policy lists
accepted remote spellings explicitly, and Curator compares the configured
URL without applying `insteadOf`, refusing when a rewrite rule matches; SSH
host aliases count only when the entry lists them. Seeds copied from a
native config can carry broad ancestor entries, including the filesystem
root, and CIP-0020 records that it cannot assert how a vendor treats
ancestor trust. Proposed: Curator reports such entries in managed homes as
ungoverned trust. Residual: a byte-identical clone at an approved identity
is indistinguishable (Decision 0026, Security impact). Owner: Curator
maintainers for the matching rules; operator for the entries.

**T4. Credential exposure through the inherited environment.** A provider
token or the operator's SSH agent socket reaches the lane, and a command
prints or uses it. Mitigations: the allowlisted base environment of section
1.3; `inherit=core` for the shell tool; the FRONT-DOOR secret check and
reserved names on records; credentials delivered only through Curator's
admitted channels (`protocol/project-context.md` §1). Residual: direct
curator-run inherits the full caller environment until the allowlist is
implemented, so the caller must scrub; a same-UID lane can also read
credential files that UID can read, such as other managed homes' login
files. Owner: launcher maintainers for the allowlist; operator for the
same-UID exposure until separate accounts exist.

**T5. Cross-lane interference on a shared host and UID.** One lane edits
another lane's checkout, managed home, trust entries, hooks or caches, or
exhausts the host. Mitigations: per-lane state directories and per-run
caches; per-Werkstatt homes and leases with busy refusal
([project leases](../project-leases.md)); one Git checkout per lane.
Residual: the same UID gives no OS boundary, and a bypassed sandbox lets any
lane write anything another lane owns. CIP-0024 and CIP-0011 move each
station to its own account; until then CIP-0024 requires an interim Seatbelt
deny profile for the operator's personal data. Owner: operator, accepted
until stations run under separate accounts.

**T6. Stale trust after a repository changes owners.** A repository is
transferred, renamed or its account is taken over. The remote URL may still
resolve and the root commit is unchanged, so identity still matches.
Mitigations: `review_by` forces periodic renewal; every grant names its
policy entry, so revoking the entry revokes every derived grant; branch
protection and signed review limit what reaches a trusted checkout through
normal flow. Residual: identity checks cannot detect a change of control
behind an unchanged URL and history. Owner: operator, who reviews entries
on their renewal date and revokes on any ownership change.

**T7. Policy or ledger tampering by a lane.** A lane with the sandbox
bypassed edits the policy, the ledger or a managed `config.toml` directly
and trusts a repository the operator never approved. Mitigations: the
launcher reports projected keys and ungoverned trust, so drift is visible;
the target root-owned trust pin with a launcher digest check before every
start (CIP-0024). Residual: on a shared UID, the policy is advisory to the
lane it governs. Owner: operator, accepted until the trust pin exists.

### 3.5 What the sandbox bypass gives an attacker

The sandbox bypass is the largest residual risk in this design. An attacker
who controls repository content, or any text the lane reads, and persuades
the lane to act on it gets everything the lane's OS user can do, without an
approval step:

- run any command as that user, with the lane's full process environment;
- read every file that user can read, including other lanes' checkouts,
  every managed home on that UID and their login files, and the trust
  policy and ledger;
- write the same files, including other lanes' hooks, shell startup files
  and trust entries, which persists the compromise beyond the session;
- reach the network without restriction;
- sign commits with the agent key under the human author name, and push to
  every repository the agent key can push to;
- send a2a messages as the lane identity.

Nothing in section 2 reduces this list. Pre-approved trust removes a
prompt; it adds no containment. The controls that bound the damage sit
outside the lane: separate OS accounts per station, a signing service that
keeps keys out of the lane, push keys scoped to the lane's repositories,
branch protection with signed review before landing, and egress limits.
The operator accepts this risk per lane by approving `yolo` for it, and the
lane record must show that approval.

## 4. Open questions for the owner

1. **Default permission mode on agent hosts.** Q-D3 resolves silence to
   `yolo`, while Decision 0018 says `native`. Recommendation: set the
   launcher-global default to `native` on hosts that run unattended lanes,
   carry `yolo` only in owner-approved lane records, and file the pending
   Decision 0018 amendment so spec and launcher agree.
2. **Grant event for policy-derived trust.** Recommendation: amend CIP-0020
   with a `policy-grant` event that names the policy entry and its digest,
   rather than minting `explicit-grant` records on the operator's behalf.
3. **Hook trust seeding.** Recommendation: allow it only for digest-listed
   hooks after per-release qualification of the Codex hook state; keep every
   other hook on the human prompt.
4. **Unknown project in an unattended launch.** Recommendation: refuse
   before start with `workspace_untrusted`, rather than start a session that
   waits on a hidden prompt.
5. **Policy integrity before separate accounts.** Recommendation: keep the
   policy in Curator private state with operator-only commands now, accept
   the same-UID tamper risk (T7) explicitly, and move it under the CIP-0024
   root-owned trust pin with the first station pilot.
