# Pre-implementation review v3 (final): the agent-environments capability

Consolidated synthesis, 2026-09-02. Merges the orchestrator's re-read (O)
with three independent fable-5 lens reviews: **A** threat model
(TASK-260901-2eu2qg), **B** operator journeys / day-2 (TASK-260902-2142et),
**C** implementation feasibility and cross-spec contract realism
(TASK-260902-13dvty). Subject: curator-spec `4d55698`, curator `66e34a23`,
curator-agent-launcher `6de42d8`, ax PR #1 `d7075e1`. *Verified* means
checked on this machine (claude 2.1.257/258, codex 0.151.0, pi 0.84.2,
gemini 0.54.4; macOS Keychain via `security`; git behavior on a scratch repo;
the curator, ax, and agents-management sources).

Severity is cost-of-being-wrong before implementation. **MUST** = resolve
before implementation touches the area; **NEXT** = same spec pass;
**LATER** = documented and deferrable.

## 1. Verdict

The core model is sound and survived every attack: the profile-influence
boundary, the closed adapter registry, byte-exact materialization with its
vectors, ledger/backup discipline, absence-vs-unreadable, always-strict
audit, fire-vs-manage, the transaction/journal engine as the switch
substrate. What is broken sits at the edges of the process and in facts we
recorded without testing: **how a session is actually launched** (no ax
input surface; agents-management has no interactive mode and would ship
`--dangerously-skip-permissions` into a terminal), **how bytes are acquired**
(`git archive` honors autocrlf and `export-subst`, so pins are
machine-dependent and Windows fails every profile — a defect the shipped
skills pipeline already carries), **what a fresh managed home is to the tool**
(not logged in, no config, first-run walls; Keychain keyed by macOS user),
**which adapter facts are true** (pi flags do not exist), **what a profile
session can run** (no commands), and **the lifecycle verbs never written**
(update, remove, unmanage, switch failure). Curator-side work fits the
existing architecture without structural change; the blocked planes are
external contracts, and one acquisition rule.

## 2. MUST — resolve before implementation

### M1. Execution ownership: no ax input surface, two argv owners [O-C1, C-C1]

*Verified:* ax's only session-creating operation is `ax start NAME --provider
ID [--profile standard|yolo] [--workspace PATH]` — no argv, no env literal,
no extension, no model, no effort. The Launch Plan is built inside ax at
creation and provider `launch` must receive exactly the record's plan. The
ax implementation does not import agents-management. So "always through ax"
fails every launch on a configured machine, and ax PR #1's "the launcher
merges the fragment into `env_literals`" names an actor with no interface.
Additionally, Launch Plan / SpawnPlan carry no stdin member, while
agents-management plans may transport effort/prompt via stdin.

Two viable resolutions — **decide as Decision 0011**:

- **Option A (lens C, minimal ax change):** extend `ax start` with
  `--launch-plan FILE|-` — a closed `{argv | argv_suffix, env_literals,
  extensions}` validated under ax §5.1 rules and embedded verbatim into the
  Session Record; provider plugins translate without rebuilding argv.
  Ownership: curator-run composes (agents-management plan + fragment +
  native args), ax validates and keeps the record immutable, the plugin
  translates. Stdin: refuse the ax route for non-empty stdin in revision 1,
  or grow SpawnPlan with an optional `stdin` — an explicit ax decision.
- **Option B (orchestrator v1, ax owns spawn):** `ax start` gains
  `--curator-profile P | --curator-env E [--model --effort]`; ax's owner
  calls `curator env resolve` itself and merges into its plugin's
  `env_literals`; the launcher delegates in tracked mode and composes only
  when untracked. Keeps ax free of a caller-supplied plan but puts
  Curator-awareness inside ax.

Recommendation: **Option A** — it preserves ax's plugin-owned resume argv,
keeps ax ignorant of Curator and agents-management, and makes one component
(curator-run) the single composer in both modes. Either way: revise ax PR #1
(operation section, stdin decision), launcher SPEC → 0.2, and state in the
launcher SPEC that the behavioral contract is unimplementable until the ax
change lands.

### M2. agents-management has no interactive launch mode [C-C2]

*Verified in `pkg/agentic`:* the closed LaunchMode set is `Exec`, `DryRun`,
`ManagedSession` — all tracked-assignment shapes. The claude argv site emits
`-p --output-format json --model … --dangerously-skip-permissions [goal]` for
every mode. Consuming `BuildPlan` as a value therefore yields a headless
print-mode run with permission bypass injected into the operator's
terminal — a safety-relevant wrong default — and appending native args
(`resume --last`) after it is incoherent. The launcher's two non-goals ("no
plan rebuilding", "exactly the tool they know") are jointly unsatisfiable
against the module as it exists.

**Resolution:** add `LaunchModeInteractive` to agents-management (per-system
argv containing only model selection and effort transport; no print mode,
output format, permission bypass, or assignment machinery; empty stdin);
launcher SPEC §4.1 names the mode it requests. Reject launcher-owned
interactive argv (a second flag-spelling site).

### M3. Snapshot acquisition is not byte-exact; Windows fails every profile [C-C3]

*Verified on a scratch repo:* `git -c core.autocrlf=true archive` emits CRLF
for a committed-LF blob, and `git archive` expands `export-subst`. Curator's
`gitops.Archive` runs plain `git archive`; audit hashes the *extracted* tree.
Consequences: on stock Git-for-Windows every git-sourced profile fails
`profile_module_bytes_invalid`; content hashes, `path`/`local` state hashes,
effective pins, and revocation identities become git-config- and
attribute-dependent — §5.6 cross-platform hash equality is false, and the
already-shipped skills pipeline carries the same defect.

**Resolution:** normative rule (core §6.2, or environments §1 if core is
frozen): snapshot production MUST reproduce the exact committed blob bytes;
working-tree conversion and attribute-driven processing MUST NOT alter or
omit entry bytes; add a vector with `* text=auto` and an `export-subst`
entry hashing to the raw-blob value. Reference implementation: extract from
the object database (`ls-tree -r` + `cat-file`), not `git archive`.

### M4. Managed-home provisioning and credentials do not match reality [O-C5/H3, A-T1/T2/T3, B-F4, C-H4/M1]

*Verified:* the Claude credential is Keychain item `Claude Code-credentials`
keyed by macOS account; a fresh login inside an `isolated` home writes the
same item, clobbering every other home's credential (an
`oauth.claude.profile.<64-hex>` account also exists — a newer per-profile
scheme worth investigating before lifting the restriction). A fresh
`CLAUDE_CONFIG_DIR` reports "Not logged in" (login state lives in
`.claude.json`); a fresh `CODEX_HOME` is not logged in until `auth.json`
passthrough and loses `config.toml` (project trust, model, MCP); a fresh pi
dir loses `settings.json`/`models.json` and re-downloads its tool trees.
Per-file symlink passthrough is severed by any temp+rename refresh (write
path per tool unverified — demand evidence). opencode `isolated` is a no-op
(auth in `XDG_DATA_HOME`), while XDG seeds link the real `~/.config` into
the "isolated" parent.

**Resolution (environments §7.4/§8.1 rewrite):** (a) a per-adapter
**provisioning seed** class — non-credential, one-time, never hashed,
enumerated per adapter; (b) passthrough *strategy* per adapter with the
pinned release's verified write behavior recorded (in-place vs
rename-over): prefer keyring-backed modes (codex `cli_auth_credentials_store
= file|keyring|auto`), directory-level passthrough for rename-over tools,
and a status liveness row `environment_passthrough_detached`; (c)
`isolated` declared **unsupported** for claude_code/macOS and opencode in
revision 1 with `environment_isolated_unsupported`, lifted only on positive
evidence; (d) a normative fresh-home paragraph and first-resolve notice;
(e) XDG seeding narrowed to an allowlist with `XDG_DATA/STATE` stated
ambient.

### M5. Adapter facts are wrong; the channel model cannot express reality [O-C4, B-F22, B-F12, C-H3, C-M3]

*Verified:* pi 0.84.2 rejects `--system-prompt-file` and
`--append-system-prompt-file` ("Unknown option"); the real flags are
`--system-prompt <text>` and `--append-system-prompt <text>`, the latter
polymorphic — a dead path is silently sent as literal prompt text (the
read-failure-as-absence class the protocol bans). Decision 0010's "verified
in 0.84.2" does not reproduce. Claude 2.1.258 recognizes both `-file`
flags. codex reads a global `AGENTS.md` from `CODEX_HOME` and ships
`project_doc_max_bytes = 32768`; cap applicability to the global doc is
unverified, and composition truncates the precedence-winning chapters first.

**Resolution:** recorded erratum on Decision 0010; rewrite the pi row from
evidence (append via `--append-system-prompt` with a path the launcher
verifies readable immediately before exec; drop pi flag/replace —
`SYSTEM.md` is its only replace path); add `argument: path | contents` to
`flag` descriptors and an admission rule requiring verified file-path
acceptance; add `root_context_size_advisory_bytes` per adapter with
`environment_context_size_exceeded` (warn); verify the codex cap.

### M6. Skill commands are unreachable in managed homes; the fragment cannot be widened later [O-C2, B-F16, A-T8]

*Verified:* `internal/globalbins` stages forwarding shims into one user-bin
dir (machine singleton); the fragment `env` enum is closed. A
`--profile companyB` session inherits the machine-current profile's shims —
wrong command version or none, silently. Skill-published bins on PATH can
also shadow `curator-run`/`curator-session` (profile *files* can poison
what profile *bytes* cannot).

**Resolution (before the fragment freezes):** profile-scoped command roots
below the environments root; fragment `path_prepend` (one manager-owned
path); `linked` switching re-points machine shims to the current profile;
umbrella providers resolved from a trusted location or identity-verified,
with managed/skill bin dirs forbidden from carrying `curator-*` names. If
revision 1 chooses "no commands in managed homes", say so — the schema
reservation is still required now.

### M7. Launcher plan/fragment ordering keys limit state to the wrong home [C-H1]

*Verified:* agents-management `LaunchRequest.Home` is load-bearing — on-disk
limit state is keyed by (provider, home), and the module warns that the key
must never move without a migration. The launcher builds the plan before it
has the fragment, so admission reads and writes the *native* home while the
process runs in the managed home; with `isolated` accounts profile A's
rate-limit evidence gates profile B.

**Resolution:** reorder launcher §4 — resolve the fragment first, pass its
managed-home path as `LaunchRequest.Home` normatively.

### M8. Bare `curator run <env>` cannot launch: nobody owns model/effort defaults [C-H2]

*Verified:* claude `Args` requires `Model.ID`; `ErrEffortMissing` refuses a
required-effort model; "no default is injected at any call site" is a
module invariant; the launcher SPEC defines no default surface. The
flagship one-liner exits with a plan refusal every time.

**Resolution:** the launcher SPEC owns default resolution: a closed
launcher machine-config mapping env-id → {model, effort}, with
`vendorplugin.Lineup`'s top admitted pair plus the vendor's recommended
effort as the documented fallback; precedence against flags named.
Interacts with M2 (interactive mode may make model optional per tool).

### M9. `context-secret-material` × pins: bypass or unescapable DoS [O-C6, A-T4, A-T5]

*Verified in the reference implementation:* `decideWithPins` returns
`DecisionAllow` for a pinned hash before evaluating findings — a routine
operator pin bypasses the always-on gate; the opposite reading makes a
false positive permanently uninstallable. No pattern classes, no precision
vectors, no span reporting; no install-time signal that a profile carries
`class: system` modules.

**Resolution:** the detector is **unpinnable**; scoped false-positive
waiver (file + span + reason, recorded); detector classes with
positive/negative vectors; reconcile the reference implementation; add the
always-warn surfacing class `context-system-module-present`.

### M10. `env resolve` mutates on the launch hot path without a lock or cost story [O-C3, A-T9, C-M2]

Declared pure, required to repair; currency as written re-hashes every
recorded surface per launch (tens of MB for a large skills tree); repair
contends on the single exclusive manager-home lock with no specified
outcome when `profile sync` holds it; the marker is not an integrity
binding.

**Resolution:** verification is lock-free and covers exactly the marker's
surfaces, with link-target identity into an immutable store entry
sufficient for symlinked surfaces; resolve reports `environment_home_stale`
without repairing unless `--repair`; repair takes the mutation lock with
bounded wait and a distinct lock-acquisition diagnostic; the repair→child
read race is a recorded residual; the launcher may re-check hashes before
exec.

### M11. Lifecycle verbs never written [O-H1/H2, B-F1/F6/F8/F9/F10, A-T7, C-L2]

- `profile update`: branch tracking is allowed but nothing moves a pin;
  specify re-resolve → strict re-audit (old pin stands on a blocking
  finding) → new store entry → re-materialize in-place scopes → managed
  homes stale → old entries GC-eligible; re-`install` of an installed
  source defined; skill-scope `update`/`upgrade` never move profile pins;
  the `default` local pin must not churn ax drift on every `global add`.
- `profile remove`: refuse while current in any scope or in any overlay;
  managed homes (operator session data) retained unless `--purge`; orphans
  reported.
- `profile use` failure semantics: attempt every entry, per-adapter results,
  record the new current only when the whole scope materialized.
- Backups: versioned `.agent-environment-backup/<n>/`, retention, scrub,
  status visibility (they hold prior secrets forever).
- `env unmanage [--restore-backups]` or documented manual rollback.

### M12. Composition, forms, targets, and `isolated` have no storage contract and no CLI [B-F18, C-L1, O-M6]

*Verified:* `manager-config-v1.schema.json` is `additionalProperties: false`
and carries none of these knobs. Manager-config schema 2 (or a versioned
`environments-config-v1`) naming every knob, informative CLI rows
(`profile compose`, `env config`), key names in manager §12; team
distribution stays per-machine with a documented bootstrap shape and an
owned rev-2 story.

### M13. Normative contradiction on the `--check` exit contract [O-M2, B-F13]

§7.5 warning vs §12 non-current for shadow-inert. Resolve to non-current
by default plus a per-path machine-config acknowledgment that downgrades
exactly that row.

### M14. XDG seed links are a point-in-time snapshot [B-F15, A-T1]

Reconcile against the operator's current XDG home on `sync`/`use`/repair;
record new seeds; `environment_seed_shadowed` for unrecorded entries; within
the M4 allowlist.

### M15. Enterprise lockability and resume-drift default [O-H9, A-T6, A-T11]

Extend manager §1 `locked` (composition policy, require-current-profile,
non-overridable skill class) or declare fleet enforcement out of revision 1
in the phasing table; ax PR #1: refuse on drift when the chain carries
`class: system` modules.

### M16. ax fragment digest over non-canonical bytes [C-M5]

`--format json` is not canonicalized, so `works.relux.curator.fragment-digest`
false-positives on any pretty-printer change. Declare the emitted JSON
CCJ-1 (registry.md §1) or key the digest over the CCJ-1 canonicalization —
fix in the ax PR before it merges.

## 3. NEXT — same spec pass

| # | Finding | Source | Resolution |
|---|---|---|---|
| N1 | opencode `referenced` form makes `opencode.json` instructions-only — MCP/providers lockout in managed homes | O-H5, B-F14 | Drop from revision 1 or state the consequence and defer |
| N2 | No per-adapter isolation matrix; opencode skills are structurally split-brain (machine-current skills, launched-profile context) | O-H6, A-T1, C-M4 | Normative table; standing `env status` note for opencode |
| N3 | No tool-version compatibility surface | O-H7, B-F11 | Recorded verified version per adapter; best-effort detected version in status; erratum fast path |
| N4 | Hybrid scope unreconciled with profiles | O-H8, B-F17 | project > hybrid > current profile of the scope; hybrid never composes |
| N5 | Xcode `auto` first-writes a foreign app's security surface; embedded MCP/commands ungoverned | A-T10, O-M5 | One-time consent before first write; status names the ungoverned surfaces; verify the agents read the files |
| N6 | Same profile, two doors, two histories | O-H4, B-F5 | Native home for the machine-current profile in `linked` mode unless `--isolated-home`, or a loud documented split |
| N7 | Windows: `path` installs from autocrlf checkouts; `--format shell` POSIX-only | O-H10, B-F19 | Fix hint / normalize flag for `path`; `--format pwsh` or json-only guidance |
| N8 | Onboarding trigger includes read-only commands | B-F2 | Only mutating profile operations trigger onboarding |
| N9 | Foreign-manager detection sees symlinks only | B-F3, O-M4 | Heuristic notice + repeated-drift "suspected external writer" |
| N10 | Composition covers skill union only | O-M3 | Specify `agents`, `locale`, hybrid targets |
| N11 | Scoped current cannot be cleared | B-F7 | `--clear` or naming the default removes the scope record |
| N12 | `curator run opencode` → `env_unsupported` while opencode is a rev-1 adapter | O-M8 | Add the plugin or mark the row |
| N13 | Decision 0010 carries the superseded sequencing sentence and the false pi verification | O-M9, B-F22 | Recorded erratum |
| N14 | claude referenced-form approval claim rests on bundle reading, not a behavioral test | C-L3 | Keep behind the pinned-release gate; extend the §7.3 wording to §5.3 |

## 4. LATER

Generation-header token cost [O-M1]; concurrency statements [O-M7]; launcher
stub reports 0.1.0-draft vs SPEC 0.1.2 [B-F20]; warning fatigue for
deliberate pi file channels [B-F21]; team-recommended machine settings as a
displayed, explicitly applied suggestion [O-M6]; `default` local pin churn
handled as one transaction per operation [C-L2].

## 5. What held under attack (do not reopen)

Env-var injection boundary (identifier grammar + shell quoting), umbrella
dispatch immune to profile bytes, no-templating IR, byte-exact determinism
with its 15 vectors (`monolithic-claude-code` expected bytes match §5/§5.1
exactly), `path` snapshot discipline, absence-vs-unreadable throughout,
always-strict profile audit, system-prompt inert-by-default, fire-vs-manage,
the two-mode requirement, the onboarding import design, ax PR #1 extension
key shape, the transaction/journal engine as the substrate for per-entry
atomic multi-adapter switching, the conservative GC extension.

## 6. Feasibility (lens C)

Curator-side revision 1 fits the existing architecture without structural
change; aggregate size is comparable to the manager §11 external-repository
capability. Nothing curator-side is blocked except M3 (acquisition
byte-exactness). The blocked planes are external: M1 (ax input surface) and
M2 (interactive launch mode), plus the launcher fixes M7/M8.

| Spec area | Existing base (`internal/…`) | Fit | Size |
|---|---|---|---|
| Profilefile/context.json validation | `manifest`, `skillspec`, `protocoljson` | clean new package | M |
| Profile store (§4) | `runtimestore` pattern | new package, pattern reuse | S–M |
| `git`/`path`/`local` acquisition | `gitops` + `snapshot` (M3 fix required); `path` copy new | mostly reuse | S + M3 |
| Always-strict audit + detector | `audit` pipeline | clean extension; detector quality is the work | S–M |
| Deterministic materialization (§5) | none — pure emitter driven by shipped vectors | new package | M |
| Adapter registry generalization (§7) | `adapters` (skills table + ledger) | generalize; marker and ledger stay separate | M |
| Environment marker (§8.2) | `marker` bands v1–v4 | sibling marker type | S |
| Modes + per-entry atomic switch | `transaction` (ordered targets, journal, rollback) + `managerlock` | **strong fit** — `profile use` is one Plan of per-entry targets | M |
| Onboarding detect/backup/takeover/import | nothing comparable | largest greenfield chunk | L |
| `env resolve` + fragment | `config`, `marker` reads; repair via `transaction` | S once M10 settles | S–M |
| `env status` matrix | status patterns | extension | S–M |
| GC live roots | `scopes/gc.go` conservative mark-sweep | clean extension | S |
| Umbrella discovery | `cmd/curator` dispatch | trivial | S |
| opencode XDG seeds | none | small, precise rules | S |

## 7. Action plan

1. **Decision 0011 — execution ownership** (M1, M2, M7, M8): choose Option
   A or B; specify the ax operation and the stdin decision; add
   `LaunchModeInteractive` to agents-management; launcher SPEC 0.2 with
   fragment-first ordering and default ownership; revise ax PR #1 (operation
   section, refuse-on-drift for system modules, CCJ-1 digest).
2. **Acquisition byte-exactness** (M3): normative rule + vector in
   curator-spec; reference implementation switches to object-database
   extraction — this also fixes the shipped skills pipeline and should land
   first as its own PR.
3. **Decision 0010 erratum** (N13): pi verification, sequencing sentence,
   credentials claim.
4. **environments.md revision 1.1 + schemas + vectors + manager §12 + CLI**,
   one producer/reviewer batch: M4–M6, M9–M15, N1–N14.
5. **Verification sprint** before any adapter freezes: Keychain
   `oauth.claude.profile.*` scheme; `.claude.json` seed shape; codex
   global-doc cap; codex/pi `auth.json` write mode; fresh-home first-run per
   tool with seeds applied; Xcode embedded agents honoring root context;
   opencode XDG on Windows; claude referenced-form approval behavior.
6. **Implementation, staged, each with its own conformance subset:** (a)
   acquisition fix, profile store, `git`/`local`, monolithic, `linked`
   switching with the M11 transactional shape, global-scope migration; (b)
   managed homes with seeds and passthrough strategies, read-only resolve,
   untracked launcher with `path_prepend` and interactive plans; (c)
   composition, `path`/import, onboarding, config schema 2 + CLI; (d) ax
   integration once Decision 0011 and the ax operation land.

## Appendix — finding cross-reference

O-C1→M1 · O-C2→M6 · O-C3→M10 · O-C4→M5 · O-C5→M4 · O-C6→M9 · O-H1/H2→M11 ·
O-H3→M4 · O-H4→N6 · O-H5→N1 · O-H6→N2 · O-H7→N3 · O-H8→N4 · O-H9→M15 ·
O-H10→N7 · O-M1→LATER · O-M2→M13 · O-M3→N10 · O-M4→N9 · O-M5→N5 ·
O-M6→M12/LATER · O-M7→LATER · O-M8→N12 · O-M9→N13 · A-T1→M4/M14 ·
A-T2/T3→M4 · A-T4/T5→M9 · A-T6→M15 · A-T7→M11 · A-T8→M6 · A-T9→M10 ·
A-T10→N5 · A-T11→M15 · B-F1→M11 · B-F2→N8 · B-F3→N9 · B-F4→M4 · B-F5→N6 ·
B-F6→M11 · B-F7→N11 · B-F8/F9/F10→M11 · B-F11→N3 · B-F12→M5 · B-F13→M13 ·
B-F14→N1 · B-F15→M14 · B-F16→M6 · B-F17→N4 · B-F18→M12 · B-F19→N7 ·
B-F20/F21→LATER · B-F22→M5 · C-C1→M1 · C-C2→M2 · C-C3→M3 · C-H1→M7 ·
C-H2→M8 · C-H3→M5 · C-H4→M4 · C-M1→M4 · C-M2→M10 · C-M3→M5 · C-M4→N2 ·
C-M5→M16 · C-L1→M12 · C-L2→LATER · C-L3→N14.
