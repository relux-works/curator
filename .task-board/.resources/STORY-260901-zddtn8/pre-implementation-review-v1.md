# Pre-implementation review: the agent-environments capability

Orchestrator synthesis, 2026-09-02. Subject: everything landed for the
capability as of curator-spec `4d55698`, curator `66e34a23`,
curator-agent-launcher `6de42d8`, and the open ax proposal
`agent-session-manager-spec#1` (`d7075e1`). Method: full re-read of Decision
0010, `protocol/environments.md`, manager §12, the launcher SPEC, and the ax
delta; targeted verification against the installed binaries (claude 2.1.25x,
codex 0.151.0, pi 0.84.2) and the ax/agents-management sources; three
independent fable-5 lens reviews (threat model, operator journeys,
feasibility) run in parallel — their reports land as an addendum.

Severity is cost-of-being-wrong before implementation: **C** must be
resolved before a line of Go is written for that area; **H** should be
resolved in the same spec pass; **M** can follow.

## 1. Critical

### C1. The execution plane has two argv owners and no interface between them

Decision 10's four-plane map says the spawn plane is `agents-management`
(`BuildLaunch` → binary/argv/env/stdin) and the session plane is `ax`. The
launcher SPEC §4.1/§4.5 composes an `agents-management` plan with the
Curator fragment and, when `ax` is configured, "routes the composed launch
through ax's instrumentation" — always, with no untracked fallback. The ax
PR says "the launcher merges the fragment env into the Launch Plan and
resulting SpawnPlan env_literals".

Verified facts that break this:

- ax's `SpawnPlan` is produced by ax **provider plugins** (SPEC §7.5
  `launch`/`resume`/`fork` → `SpawnPlan`), and `resume-plan` derives argv
  from the plugin so that resume fidelity holds. Nothing in ax's command
  vocabulary (`resume`, `fork`, `pane`, `takeover`, `stop`, `sync`,
  `sessions`, `session clone|set-profile`, `materialize`, …) accepts an
  external plan, an external argv, or an env fragment. The launcher SPEC
  itself lists the handoff argv shape as an open item (§9).
- The ax implementation does not consume `agents-management` (`go.mod`
  requires only `jcs`). So when tracked, argv comes from ax's plugin; when
  untracked, from `agents-management`. Two builders, two goldens, and the
  launcher's `--model/--effort` admission story applies to one of them.

As written, "always through ax" is unimplementable and the ax PR describes a
merge no external process can perform.

**Proposal (decide as Decision 0011, then revise ax PR #1 and launcher SPEC
0.2):** ax owns spawn whenever it is present. Concretely: ax gains one
operation, e.g. `ax start <provider> [--curator-profile <name>|--curator-env
<env-id>] [--model … --effort …] -- <native args>`, whose owner calls
`curator env resolve --format json` itself (this is the D10 direction: ax
calls Curator, never the reverse), merges the fragment into its plugin's
`SpawnPlan.env_literals`, and records the `works.relux.curator.*` keys. The
launcher in tracked mode becomes a thin delegator to that operation; its own
composition (agents-management plan + fragment + exec) is the untracked
path only. Whether ax's plugins should consult `agents-management` for
admission and effort vocabulary is a separate, smaller decision. The
alternative — ax accepting an external Launch Plan — is rejected because it
breaks ax's plugin-derived resume/fork argv.

### C2. Skill commands are invisible inside managed homes and unscoped under switching

Today global skills publish command shims to `<curator-home>/global/bin`,
`env.sh` exports `CSK_GLOBAL_ROOT` and prepends that bin, and forwarding
shims land in a user bin dir on `PATH` (manager §3, `globalbins`). The
capability makes the global skill set profile-scoped (§9.4) but never
mentions commands: the fragment carries only the home variable, so a
`curator run` session sees the profile's `skills/` context but none of its
commands; and under `profile use`, which profile's shims occupy the single
machine-global bin and the forwarding shims is unspecified.

**Proposal:** (a) profile-scoped command roots below the environments root
(`<environments-root>/<profile>/bin`), ledgered; (b) the fragment gains a
closed `path_prepend` member (the profile's bin root) that the launcher
merges into `PATH` — registry-derived, never profile bytes, so §10.3 holds;
(c) in-place `linked` switching re-points the machine-global bin and
forwarding shims to the current profile's command set under the existing
ledger; (d) `env status` reports command currency. This touches §9.4, §10.2,
§10.3, manager §3/§12, and the fragment schema.

### C3. `env resolve` is declared pure but mutates on the launch hot path

§10.1 and manager §12.5 call resolution a pure function, then require it to
verify *and repair* the managed home — re-materializing surfaces and links —
before emitting a fragment. Repair is a manager-home write: it needs the
mutation lock (manager §2.5), so every launch contends with installs and GC,
and it swaps files under any session already running in that home (the same
home serves every concurrent session of a profile).

**Proposal:** resolve is read-only and reports currency: current → fragment;
non-current → `environment_home_stale` with the drift facts, no fragment,
unless `--repair` is passed (then the repair runs under the lock, atomically
per entry, with a warning that live sessions in the home are affected). The
launcher and ax choose the policy (`--repair` by default for interactive
launches is acceptable; the point is that the write is explicit and
lock-scoped). Specify lock class for resolve (shared) vs repair (mutation).

### C4. Adapter facts already wrong or unmodeled: pi flags, flag argument kinds, size ceilings

Verified on this machine: pi 0.84.2 exposes `--system-prompt <text>` and
`--append-system-prompt <text>` (the latter "text or file contents"); there
are no `-file` variants, yet §7.3 declares `--append-system-prompt-file`
and `--system-prompt-file` for pi, and the launcher would exec a
non-existent flag. Claude Code's option list shows only `--system-prompt`
and `--append-system-prompt`; the `-file` spellings appear only inside the
`--bare` help text and need confirmation. Both point at a modeling gap: a
`flag` descriptor needs an `argument` field — `path` or `contents` — so the
launcher knows whether to pass the file path or read the bytes.

Separately, codex reads `project_doc_max_bytes` (default 32 KiB) over its
instruction documents; whether the `CODEX_HOME/AGENTS.md` bytes count
toward it is unverified, and a composed profile (header + chapters + two
overlays) can exceed it silently. **Proposal:** verify and record per-adapter
root-context size ceilings in the registry; materialization warns
`environment_context_size_exceeded` when output exceeds the adapter ceiling.

### C5. Credential passthrough by symlink is fragile and `isolated` is unproven for the flagship tool

Codex and pi refresh `auth.json`; a tool that writes atomically (temp file +
rename) replaces the *symlink* in the managed home with a regular file: the
managed home keeps a stale private copy, the native home stops receiving
refreshes, and nothing in §7.4/§8.4 detects it (passthrough entries are
excluded from drift). For `claude_code` on macOS the spec offers `isolated`
as "the supported shape for genuinely separate accounts", but Keychain
storage is process-global unless the entry is keyed by config dir — not
verified — so isolation may be impossible exactly where it is advertised.

**Proposal:** per-adapter passthrough *strategy*, not just entries: prefer
keyring-backed modes where the tool offers them (codex exposes
`cli_auth_credentials_store = file|keyring|auto`), symlink only for tools
verified to write in place; add `environment_passthrough_detached` to status
(link became a file); mark `claude_code`/macOS `isolated` as reserved pending
verification rather than supported.

### C6. `context-secret-material` is REQUIRED, blocking, undefined, and un-overridable

§9.1/§12.6 make profile audit always-strict and add a blocking detector whose
pattern classes are not specified; no conformance vectors bound its
precision, and no operator pin path is named. A profile that documents a
token *format* (an onboarding guide, a sample `.env` block) fails
installation with no recourse — an install DoS on exactly the enterprise
profiles this is meant to serve.

**Proposal:** normative detector classes with positive/negative vectors
(canonical key formats, `BEGIN … PRIVATE KEY`, bearer-token shapes; not bare
words), findings that name module and line, and the core §7 operator pin
extended to profile findings — a pin is explicit and recorded, so strictness
is preserved while recovery exists.

## 2. High

- **H1. No removal or un-managing lifecycle.** `profile remove` is absent;
  managed homes hold session history; backups are never overwritten and
  never collected, so a second onboarding on the same home dies with
  `environment_backup_exists`. Specify `profile remove [--purge-state]`,
  `env unmanage` (restore native files from the recorded backup), and
  timestamped backup roots or `backup prune`.
- **H2. No upgrade lifecycle.** Branch-tracking profiles have no `profile
  update/upgrade`: when does the pin move, what re-materializes, what
  happens to live sessions and to ax drift (every pin move flips every
  session's `profile-pin`). The `default` local profile's state hash moves
  on every global skill operation, so ax will warn on resume after any
  `global add`. Specify the upgrade transaction and exclude skill-only
  state changes from the drift comparison (or record a separate
  context-only pin).
- **H3. Managed homes start as fresh installs.** A new `CLAUDE_CONFIG_DIR`
  has no `.claude.json`: first launch replays onboarding, theme prompts,
  per-project trust dialogs, and MCP approvals — per profile. Codex and pi
  have equivalents. Define a per-adapter **provisioning seed** class
  (non-credential, one-time, never hashed) copied at provisioning, with the
  exact entries enumerated per adapter, distinct from credentials (§7.4)
  and from XDG seeds (§7.1).
- **H4. Same profile, two homes.** Native `claude` and `curator run
  claude_code` under the machine-current profile use different homes and
  therefore different session histories. Either resolve for the current
  profile in `linked` mode returns the native home (empty `env`) unless an
  `--isolated-home` is requested, or document the split loudly.
- **H5. opencode `referenced` form takes over `opencode.json`.** The
  managed file is fully manager-authored with a single `instructions`
  member, so it displaces every other opencode setting in a managed home,
  and in place it is unavailable for anyone with an existing config. Drop
  the opencode referenced form from revision 1 (monolithic stays).
- **H6. No per-adapter isolation matrix.** opencode skills stay at the
  native `~/.agents/skills`, XDG seeding links the whole `~/.config`,
  Claude auth is ambient — a normative table "what a profile isolates and
  what it shares, per adapter" prevents operators from over-trusting
  "profile".
- **H7. No tool-version compatibility surface.** Adapters are verified
  against pinned releases; tools auto-update and can move their home
  layout. Adapters declare a tested version range; `env status` reports
  `environment_tool_version_untested`; never blocks.
- **H8. Hybrid scope and project registrations under profiles** are
  unaddressed (manager §4.3 hybrid manifests are machine-local).
- **H9. Enterprise lockability.** Manager §1 `locked` keys cannot lock
  composition, forms, `system_prompt_files`, or target participation, so a
  mandated company profile is overridable by a personal overlay with
  `later-overrides-earlier`. Extend the lockable set or state the gap.
- **H10. Windows realities.** `path` installs copy working trees: an
  autocrlf checkout yields CRLF modules and `profile_module_bytes_invalid`
  on every module (git snapshots are immune, path copies are not) — add a
  fix hint or an explicit `--normalize-line-endings` for `path` installs;
  opencode's `XDG_CONFIG_HOME` behavior on Windows is unverified.

## 3. Medium

- **M1.** The generation header (URL + long notice) is inside the prompt of
  every session; ~100 tokens per tool per session. Shorten the notice; the
  URL can be a short fixed token.
- **M2.** §12 treats a shadow-inert row as non-current while §7.5 makes
  shadowing a warning — `--check` will fail forever for a pi user with a
  personal `AGENTS.override.md`. Pick one (warning that does not flip
  currency).
- **M3.** Composition specifies skill-set union only; `agents`, `locale`,
  and hybrid `targets` in composed Skillfiles are unresolved.
- **M4.** Foreign-manager detection sees symlinks only; a copying manager
  is treated as unmanaged files — document as a known limitation.
- **M5.** Xcode secondary targets write into another application's support
  directory on docs-confidence; keep `auto`, but add an explicit
  verification item that the embedded agents honor `CLAUDE.md`/`AGENTS.md`
  there before shipping the target.
- **M6.** Team distribution: composition, forms, targets, and
  `system_prompt_files` are per-machine and unversioned. A profile-carried
  *suggested* machine config that Curator displays and applies only under
  an explicit flag keeps the profile-influence boundary while giving teams
  a one-command setup.
- **M7.** Concurrency statements missing: multiple sessions share one
  managed home (tool-level concurrency applies); `profile use` never
  affects managed-home sessions.
- **M8.** `curator run opencode` fails `env_unsupported` (no
  agents-management plugin) while Curator lists opencode as a revision-1
  adapter — either add the plugin before release or document the gap in
  the adapter table.
- **M9.** Decision 0010 still carries the superseded "path delivered
  between revisions 1 and 2" sentence (review minor); amend.

## 4. What holds up well

The profile-influence boundary, the closed adapter registry, the
byte-exact materialization surface with its vectors, the ledger/backup
discipline, the absence-vs-unreadable separation, the strict-audit stance,
and the fire-vs-manage split all survived three adversarial cycles each and
re-reading; none of the findings above argues for reopening them.

## 5. Recommended sequence

1. **Decision 0011 — execution ownership** (C1): ax owns spawn when present;
   define the ax operation; revise ax PR #1 (add the operation section) and
   launcher SPEC 0.2 (tracked mode delegates). Nothing else in the launcher
   should be implemented before this.
2. **environments.md revision 1.1 batch** (one producer/reviewer cycle):
   C2 command roots + `path_prepend`; C3 read-only resolve + explicit
   repair + lock classes; C4 `argument` on flag descriptors, verified
   spellings, size ceilings; C5 passthrough strategies +
   `environment_passthrough_detached` + isolated scoping; C6 detector
   classes + vectors + pins; H1/H2 lifecycles; H3 provisioning seeds; H5
   drop opencode referenced; H6 isolation matrix; M2/M3/M7 fixes; schema +
   vector updates.
3. **Verification sprint** (cheap, high value, before implementation):
   claude `-file` flags and Keychain keying; codex global `AGENTS.md`
   cap; codex/pi `auth.json` write mode (rename vs in-place); managed-home
   first-run behavior per tool; Xcode embedded agents honoring root
   context; opencode XDG on Windows. Record as a board evidence resource and
   adjust the registry.
4. **Implementation, staged:** (a) profile store, `git`/`local` kinds,
   monolithic form, `linked` in-place switching, migration of the global
   scope; (b) managed homes, read-only resolve, untracked launcher; (c)
   composition, `path`/import, onboarding; (d) ax integration once Decision
   0011 and the ax operation land. Each stage gets its own conformance
   subset so partial progress is measurable without claiming the
   capability.
