# Pre-implementation review v2: the agent-environments capability

Consolidated synthesis, 2026-09-02. Merges the orchestrator's own re-read
(v1) with three independent fable-5 lens reviews: **A** threat model
(TASK-260901-2eu2qg), **B** operator journeys / day-2 (TASK-260902-2142et),
**C** implementation feasibility (TASK-260902-13dvty). Subject: curator-spec
`4d55698`, curator `66e34a23`, curator-agent-launcher `6de42d8`, ax PR #1
`d7075e1`. Every claim marked *verified* was checked on this machine
(claude 2.1.257, codex 0.151.0, pi 0.84.2, gemini 0.54.4; macOS Keychain
via `security`; the curator reference implementation; the ax and
agents-management sources).

Source tags: O = orchestrator, A/B/C = lens. Severity is cost-of-being-wrong
before implementation. **MUST** = resolve before implementation touches the
area; **NEXT** = same spec pass; **LATER** = documented and deferrable.

## 1. Verdict in one paragraph

The core model is sound and survived every attack: the profile-influence
boundary, the closed adapter registry, byte-exact materialization with
vectors, ledger/backup discipline, absence-vs-unreadable, always-strict
audit, fire-vs-manage. What is broken is everything at the *edges of the
process*: how a session actually gets launched (two argv owners, no ax
interface), what a freshly provisioned managed home looks like to the tool
(not logged in, no config, first-run walls), how credentials really behave
(Keychain keyed by user; symlinks eaten by rename), which adapter facts are
true (pi flags do not exist), what commands a profile session can run
(none), and the lifecycle verbs that were never written (update, remove,
unmanage, switch failure). None of these invalidate the design; all of them
would have been discovered mid-implementation at the worst possible time.

## 2. MUST — resolve before implementation

### M1. Execution ownership: two argv owners, no interface [O-C1; C pending]

Decision 10 puts spawn in `agents-management` and session in `ax`; the
launcher SPEC composes an `agents-management` plan with the fragment and
then "always routes through ax". *Verified:* ax builds `SpawnPlan` inside
its provider plugins and derives resume argv from them; its command
vocabulary has no operation accepting an external plan, argv, or fragment;
the ax implementation does not import `agents-management` (go.mod: `jcs`
only). The ax PR describes a merge no external process can perform, and
the launcher SPEC lists the handoff shape as an open item.

**Resolution → Decision 0011.** ax owns spawn whenever present: one new ax
operation (`ax start <provider> [--curator-profile P | --curator-env E]
[--model --effort] -- <native args>`) whose owner calls `curator env
resolve --format json`, merges the fragment into its plugin's
`env_literals`, and records the `works.relux.curator.*` keys. The launcher
in tracked mode delegates; its own agents-management composition is the
untracked path. Revise ax PR #1 (add the operation section) and launcher
SPEC → 0.2. Whether ax plugins consult agents-management for admission is a
smaller follow-on decision.

### M2. Managed-home provisioning and credentials do not match reality [O-C5, O-H3, A-T1/T2/T3, B-F4]

*Verified on macOS:* the Claude credential is a Keychain item
`svce="Claude Code-credentials" acct="iv"` — keyed by macOS user, not by
config dir — so two `isolated` claude profiles share one account: the
headline "company beside personal" case cannot work on the flagship
tool/platform. *Also verified:* `CLAUDE_CONFIG_DIR=<fresh> claude config get
theme` → **"Not logged in · Please run /login"**: login *state* lives in
`.claude.json` inside the home; Keychain alone authenticates nothing. A
fresh `CODEX_HOME` is likewise not logged in until `auth.json` passthrough,
and loses `config.toml` (project trust, model, MCP) entirely; a fresh pi
dir loses `settings.json`/`models.json` and re-downloads its `bin/`,
`npm/`, `tools/` trees. File passthrough by symlink is destroyed by the
first temp+rename token refresh (`auth.json`, `.credentials.json`), and
credentials are excluded from drift, so the breakage is silent. opencode
`isolated` is a no-op (auth lives in `XDG_DATA_HOME`, never swapped) while
XDG seeds link the operator's real `~/.config` (gh, git, gpg) into the
"isolated" parent.

**Resolution (environments §7.4/§8.1 rewrite):** (a) introduce a
per-adapter **provisioning seed** class — non-credential, one-time copy at
provisioning, never refreshed, never hashed — enumerated per adapter
(claude: auth-relevant `.claude.json` members, `settings.json`; codex:
`config.toml`; pi: `settings.json`, `models.json`, tool trees or a
documented re-download); (b) passthrough *strategy* per adapter instead of
per-file symlinks: prefer keyring-backed modes (codex exposes
`cli_auth_credentials_store = file|keyring|auto`), shared-directory
passthrough where the tool rewrites in place, and a status check
`environment_passthrough_detached` for a link that became a file; (c) mark
`isolated` **unsupported** for claude_code/macOS and opencode in revision 1
(`environment_isolated_unsupported`), not silently shared; (d) a normative
"fresh-home state" paragraph naming what each adapter re-asks, and a
first-resolve notice; (e) narrow XDG seeding to a documented allowlist and
state that `XDG_DATA/STATE` stay ambient.

### M3. Adapter facts are wrong and the channel model cannot express reality [O-C4, B-F22, B-F12]

*Verified:* pi 0.84.2 has `--system-prompt <text>` and
`--append-system-prompt <text>` ("text or file contents"); the §7.3
`--*-file` spellings do not exist, and Decision 0010 asserts a
"verified in 0.84.2" that does not reproduce. Applying the declared replace
channel would pass a *path string* as the entire system prompt — a silent
misfire the launcher's warning would even bless. Claude's `-file` spellings
do hold on 2.1.257. codex ships `project_doc_max_bytes = 32768`; whether it
caps the `CODEX_HOME/AGENTS.md` read is unverified, and a composed profile
walks past it with the *later* (precedence-winning) chapters truncated
first.

**Resolution:** correct the decision claim as a recorded erratum; rewrite
the pi row from evidence (`--append-system-prompt`, path-or-text; drop the
flag/replace channel — `SYSTEM.md` is pi's only replace path); add
`argument: path | contents` to `flag` descriptors and an admission rule
that a `flag` channel is admissible only with verified evidence it accepts a
file path; add an adapter `root_context_size_advisory_bytes` with
`environment_context_size_exceeded` (warn, never fail); verify the codex cap
against the global doc.

### M4. Skill commands are unreachable in managed homes; the fragment cannot be widened later [O-C2, B-F16, A-T8]

Global skills publish forwarding shims into one user-bin dir (verified:
`internal/globalbins`, a machine singleton). The fragment carries only the
home variable and its `env` name enum is closed. A `curator run codex_cli
--profile companyB` session inherits the *machine-current* profile's shims:
the agent runs the wrong command version silently or none at all. Nothing
in environments.md, manager §12, or the launcher SPEC mentions PATH.
Separately (A-T8), skill-published bins on PATH can shadow `curator-run` /
`curator-session`: profile *bytes* cannot steer dispatch, profile
*materialized files* can.

**Resolution (decide before the fragment freezes):** profile-scoped command
roots below the environments root, ledgered; fragment gains a closed
`path_prepend` member (one manager-owned path — the §10.3 boundary holds);
`linked` switching re-points the machine bin and forwarding shims to the
current profile; umbrella providers resolve from a fixed trusted location
or verify identity, and managed/skill bin dirs are forbidden from carrying
`curator-*` names. If the answer is "no commands in managed homes in rev 1",
say it in §9.4 and the launcher SPEC — but the fragment reservation must
be made now either way.

### M5. `context-secret-material` × pins: bypass or unescapable DoS [O-C6, A-T4]

*Verified in the reference implementation:* `decideWithPins` returns
`DecisionAllow` for a pinned hash before evaluating findings — a routine
operator pin bypasses the one always-on gate. The opposite reading (pins
never override) makes a false positive on free-form markdown permanently
uninstallable. The detector has no pattern classes, no precision vectors,
no span reporting.

**Resolution:** `context-secret-material` is **unpinnable**; add a scoped
false-positive waiver (file + span + reason, recorded in the audit record);
define detector classes with positive/negative vectors; reconcile the
reference implementation. Also add an always-warn surfacing class
`context-system-module-present` (A-T5) so system-prompt bytes never enter a
machine without an install-time provenance line.

### M6. `env resolve` mutates on the launch hot path without a lock story [O-C3, A-T9]

Declared pure, required to repair. Repair is a manager-home write under
running sessions of the same home, with no lock cited (contrast `profile
use`), racing concurrent resolves, and the marker is explicitly not an
integrity binding.

**Resolution:** resolve is read-only and reports `environment_home_stale`;
repair is explicit (`--repair` / `profile sync`) under the mutation lock,
atomic per entry, with the repair→child-read race recorded as a residual;
the launcher may optionally re-check surface hashes before exec.

### M7. Lifecycle verbs that were never written [O-H1/H2, B-F1/F6/F8/F9/F10, A-T7]

- `profile update`: branch tracking is allowed but nothing can move a pin;
  re-`install` of an installed source is unspecified. Specify re-resolve →
  strict re-audit (old pin stands on a blocking finding) → new store entry
  → re-materialize in-place scopes → managed homes stale → GC-eligible old
  entries; the `default` local pin must not churn ax drift on every
  `global add` (record a context-only pin or exclude skill-state).
- `profile remove`: refuse while current in any scope or named in any
  overlay; leave managed homes (operator session data) unless `--purge`;
  orphaned homes reported, not silently rooted forever.
- `profile use` failure semantics: attempt every entry, report per-adapter
  results, record the new current only when the whole scope materialized —
  "recorded current implies fully materialized" as a testable invariant.
- Backups: versioned `.agent-environment-backup/<n>/`; the never-overwrite
  rule as written dead-ends the second management cycle; add retention,
  scrub, and status visibility (they hold prior secrets forever).
- `env unmanage [--restore-backups]` or a documented manual rollback.

### M8. Composition, forms, targets, and `isolated` have no storage contract and no CLI [B-F18, O-M6]

*Verified:* `manager-config-v1.schema.json` is `additionalProperties:
false` and carries none of: overlay lists, precedence, per-env form,
target participation, `system_prompt_files`, credential mode, per-scope
currents. A conforming implementation cannot store revision-1's headline
knobs in the versioned config and has no operator command for any of them.

**Resolution:** manager-config schema 2 with an `environments` object (or
a separate versioned `environments-config-v1`) naming every knob, plus the
informative CLI rows (`profile compose`, `env config`, or equivalents) and
the machine-config key names in manager §12. Team distribution stays
per-machine by design, with a bootstrap-script shape documented and a rev-2
"machine-settings fragment" story owned in the phasing table.

### M9. Normative contradiction on the `--check` exit contract [O-M2, B-F13]

§7.5 makes a shadowing path a warning; §12 makes shadow-inert non-current
and `--check` non-zero. A pi user with a deliberate `AGENTS.override.md`
fails CI forever. Resolve one way (non-current by default) and add a
per-path acknowledgment in machine config that downgrades exactly that row.

### M10. XDG seed links are a point-in-time snapshot [B-F15, A-T1]

Refresh heals only *recorded* seeds; `~/.config/gh` created next month never
reaches managed opencode homes, and a tool writing a fresh config into the
managed parent shadows the operator's real one with no diagnostic.
Reconcile against the operator's current XDG home on `sync`/`use`/repair,
record new seeds, and report `environment_seed_shadowed` for unrecorded
entries — within a seeding allowlist (M2e).

### M11. Enterprise lockability [O-H9, A-T6]

A personal overlay under the default `later-overrides-earlier` shadows a
mandated company skill with a warning; manager §1 `locked` cannot lock
composition, current profile, or target participation. Either extend the
lockable set (composition policy, require-current-profile, non-overridable
skill class) or declare fleet enforcement out of revision-1 scope in the
phasing table. Related: resume drift should default to **refuse** when the
chain carries `class: system` modules (A-T11; change in ax PR #1 text).

## 3. NEXT — same spec pass

| # | Finding | Source | Resolution |
|---|---|---|---|
| N1 | opencode `referenced` form makes `opencode.json` instructions-only — total lockout of MCP/providers in managed homes, unavailable in place for anyone with a config | O-H5, B-F14 | Drop from revision 1 (monolithic stays) or state the consequence and defer to a merge-safe design |
| N2 | No per-adapter isolation matrix; operators will over-trust "profile" | O-H6, A-T1 | Normative table: what a profile isolates vs shares, per adapter |
| N3 | No tool-version compatibility surface; tools auto-update weekly | O-H7, B-F11 | Recorded verified-version per adapter; `env status` best-effort detected version; erratum fast path for vendor layout changes |
| N4 | Hybrid scope unreconciled with profile-scoped globals | O-H8, B-F17 | One paragraph: project > hybrid > current profile of the applicable scope; hybrid never composes |
| N5 | Xcode `auto` first-writes a foreign app's security surface on unrelated ops; embedded MCP/commands ungoverned | A-T10, O-M5 | One-time consent before first write; `env status` states the ungoverned surfaces; verify the embedded agents read the files |
| N6 | Same profile, two doors, two histories (native vs launcher) | O-H4, B-F5 | Either resolve for the machine-current profile in `linked` mode returns the native home unless `--isolated-home`, or a loud documented split |
| N7 | Windows: `path` installs from autocrlf checkouts fail every module; `--format shell` is POSIX-only | O-H10, B-F19 | Fix hint / explicit normalize flag for `path` installs; `--format pwsh` or json-only guidance |
| N8 | Onboarding trigger includes read-only commands as written | B-F2 | Only mutating profile operations trigger onboarding |
| N9 | Foreign-manager detection sees symlinks only; copying dotfile managers fight forever | B-F3, O-M4 | Best-effort heuristic notice + repeated-drift "suspected external writer" report |
| N10 | Composition specifies skill union only; `agents`, `locale`, hybrid targets unresolved | O-M3 | Specify per field |
| N11 | Scoped current cannot be cleared back to "follows default" | B-F7 | `--clear` or naming the default removes the scope record |
| N12 | `curator run opencode` → `env_unsupported` while opencode is a rev-1 adapter | O-M8 | Add the agents-management plugin or mark the row |
| N13 | Decision 0010 carries the superseded sequencing sentence and the false pi verification | O-M9, B-F22 | Recorded erratum/amendment, not a silent edit |

## 4. LATER — documented and deferrable

Generation-header token cost (~100 tokens per session; shorten the notice,
short URL token) [O-M1]; concurrency statements (many sessions share one
managed home; `profile use` never affects managed-home sessions) [O-M7];
launcher stub reports 0.1.0-draft while SPEC is 0.1.2 [B-F20]; unsuppressible
warning fatigue for deliberate pi file channels — rev-2 tuple acknowledgment
[B-F21]; team-recommended machine settings as a displayed, explicitly
applied suggestion [O-M6].

## 5. What held under attack (do not reopen)

Env-var injection boundary (identifier grammar + shell quoting), umbrella
dispatch immune to profile bytes, no-templating IR, byte-exact determinism
and its 15 vectors, `path` snapshot discipline, absence-vs-unreadable
discipline throughout, always-strict profile audit, system-prompt
inert-by-default, fire-vs-manage split, the two-mode requirement, the
onboarding import design.

## 6. Feasibility lens (C)

*Pending — this section is filled in when TASK-260902-13dvty reports.*

## 7. Action plan

1. **Decision 0011 — execution ownership** (M1). Nothing launcher- or
   ax-related is implemented before it. Deliverables: decision text, ax PR
   #1 revision (operation section + refuse-on-drift for system modules),
   launcher SPEC 0.2 (tracked mode delegates).
2. **Decision 0010 erratum** (N13): pi verification claim, sequencing
   sentence, credentials claim — recorded amendment.
3. **environments.md revision 1.1 + schemas + vectors**, one producer/
   reviewer batch, itemized: M2 provisioning seeds + passthrough strategies +
   `environment_passthrough_detached` + `environment_isolated_unsupported` +
   fresh-home paragraph + XDG allowlist; M3 channel `argument`, admission
   rule, corrected pi row, size advisory; M4 command roots + fragment
   `path_prepend` + umbrella hardening; M5 detector classes + waiver +
   `context-system-module-present`; M6 read-only resolve + `--repair` +
   locks; M7 update/remove/switch/backup/unmanage; M8 config schema 2 +
   CLI rows; M9 shadow acknowledgment; M10 seed reconciliation; M11
   lockability or explicit deferral; N1–N13.
4. **Verification sprint** before any adapter freezes: Claude Keychain and
   `.claude.json` seeding shape; codex global-doc cap; codex/pi `auth.json`
   write mode; fresh-home first-run per tool with seeds applied; Xcode
   embedded agents honoring root context; opencode XDG on Windows; the
   claude `-file` flags on the pinned release. Record as a board evidence
   resource; adjust the registry.
5. **Implementation, staged, each with its own conformance subset:** (a)
   profile store, `git`/`local`, monolithic, `linked` switching with the
   M7 transactional shape, global-scope migration; (b) managed homes with
   seeds and passthrough strategies, read-only resolve, untracked launcher
   with `path_prepend`; (c) composition, `path`/import, onboarding,
   config schema 2 + CLI; (d) ax integration once Decision 0011 lands.

## Appendix — finding cross-reference

O-C1→M1 · O-C2→M4 · O-C3→M6 · O-C4→M3 · O-C5→M2 · O-C6→M5 · O-H1/H2→M7 ·
O-H3→M2 · O-H4→N6 · O-H5→N1 · O-H6→N2 · O-H7→N3 · O-H8→N4 · O-H9→M11 ·
O-H10→N7 · O-M1→LATER · O-M2→M9 · O-M3→N10 · O-M4→N9 · O-M5→N5 · O-M6→M8 ·
O-M7→LATER · O-M8→N12 · O-M9→N13 · A-T1→M2/M10 · A-T2/T3→M2 · A-T4→M5 ·
A-T5→M5 · A-T6→M11 · A-T7→M7 · A-T8→M4 · A-T9→M6 · A-T10→N5 · A-T11→M11 ·
B-F1→M7 · B-F2→N8 · B-F3→N9 · B-F4→M2 · B-F5→N6 · B-F6→M7 · B-F7→N11 ·
B-F8/F9/F10→M7 · B-F11→N3 · B-F12→M3 · B-F13→M9 · B-F14→N1 · B-F15→M10 ·
B-F16→M4 · B-F17→N4 · B-F18→M8 · B-F19→N7 · B-F20/F21→LATER · B-F22→M3.
