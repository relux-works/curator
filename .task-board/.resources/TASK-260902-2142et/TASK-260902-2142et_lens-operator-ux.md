# Lens B report — operator UX, day-2 operations, and failure modes

Task: TASK-260902-2142et (STORY-260901-zddtn8). Reviewer lens over the
agent-environments capability. Corpus: curator-spec decision 0010,
protocol/environments.md, manager.md §12 (+§4/§5/§6/§10), cli/curator.md, the
four v1 schemas, conformance/v1 environments vectors;
curator-agent-launcher SPEC.md 0.1.2-draft, README.md, cmd/curator-run stub;
agent-session-manager-spec SPEC.md v0.5.0 + PR #1 delta;
skill-agents-management SKILL.md; curator reference implementation (read-only).

Binary evidence gathered on this machine (2026-09-02): Claude Code 2.1.257,
codex-cli 0.151.0, pi 0.84.2, gemini 0.54.4; opencode not installed. Probes
that produced findings are quoted inline; each finding names its evidence
class (verified-on-binary, spec-text, or journey-analysis).

Severity vocabulary: critical | high | medium | low. Each finding carries:
area, exact quote/section, why it is a weakness, a concrete proposal, and
MUST-before-implementation or can-follow. The MUST list is consolidated at
the end, ranked by cost-of-being-wrong.

Revision 2 (reviewer run RUN-260901-2fae73, 2026-09-02): independent
re-verification pass over the same corpus confirmed the revision-1 findings
and added two binary/schema-verified facts — F22 (the §7.3 pi flag spellings
do not exist in pi 0.84.2; Decision 0010's "verified in 0.84.2" claim does
not reproduce) and the manager-config-v1 closed-schema evidence strengthening
F18. The MUST list is re-ranked accordingly (now 10 items).

---

## Journey 1 — first install on a used machine

### F1. HIGH — `environment_backup_exists` is a guaranteed dead end on any second takeover of the same path
- **Area**: onboarding/backup, environments §8.3, manager §12.2.
- **Quote**: "A backup, once written, is never overwritten: an operation that
  would replace an existing backup path MUST fail with
  `environment_backup_exists`." (§8.3) Backups "preserve each file's
  home-relative path" — i.e. the backup path for `~/.claude/CLAUDE.md` is
  always `.agent-environment-backup/CLAUDE.md`.
- **Why weak** (journey-analysis): the failure is reachable by construction,
  not by misuse. Cycle: onboard (backup written) → operator un-manages by
  hand or restores the backup → edits `CLAUDE.md` natively → any later
  takeover/onboarding of the same path computes the same backup path and MUST
  fail. No versioned backup naming, no documented recovery action, and no
  operator-facing text tells them the fix is "move the old backup aside
  yourself." The one safety property (never lose bytes) turns into a
  permanent refusal on the second management cycle of a machine — exactly the
  used-machine journey this section exists for.
- **Proposal**: version the backup path — spec text direction: "backups land
  in `.agent-environment-backup/<n>/` where `<n>` is the lowest unused
  non-negative integer for this operation; one operation writes one `<n>`
  tree; a path collision *within* one operation is still
  `environment_backup_exists`." Never-overwrite is preserved per entry;
  repeated cycles keep working; restore-latest is well-defined. At minimum,
  require the `environment_backup_exists` diagnostic to name the existing
  backup path and the exact manual action.
- **MUST before implementation** — backup layout is written into homes on
  first contact; changing it later orphans existing backups.

### F2. MEDIUM — "first profile operation that meets unmanaged state" leaves the onboarding trigger set open; read-only commands must be excluded explicitly
- **Area**: onboarding trigger, environments §9.5, manager §12.3.
- **Quote**: "On bootstrap, or on the first profile operation that meets
  unmanaged state, the manager: 1. Inventories … 3. Backs up, always …"
  (§9.5) vs §12: "Both commands follow the manager §10 discipline exactly:
  recompute and report, never mutate."
- **Why weak** (spec-text): `profile list` and `env status` are profile
  operations; on a used machine they will "meet unmanaged state." As written
  an implementer can legitimately wire onboarding (which writes backups) into
  a read-only command, contradicting §12, or exclude it and never trigger
  onboarding until `install` — divergent implementations either way.
- **Proposal**: one sentence in §9.5: "Only mutating profile operations —
  `install`, `use`, `sync`, takeover — trigger onboarding; read-only
  commands report unmanaged state and never begin it."
- **Can follow** (small text fix, but do it before someone implements the
  wrong branch — cheap to include in the pre-implementation batch).

### F3. MEDIUM — foreign-manager detection only sees symlinks; file-writing dotfile managers are absorbed silently and then fight Curator forever
- **Area**: onboarding, environments §9.5 step 1.
- **Quote**: "managed-surface paths that are already symlinks pointing
  outside the manager's store. The last is evidence of another manager and
  stops the operation…"
- **Why weak** (journey-analysis): chezmoi, home-manager in copy mode,
  dotbot copy mode, and ansible-managed dotfiles produce plain files. They
  pass detection as ordinary unmanaged files, get backed up and taken over —
  and the foreign tool's next apply overwrites Curator's surface, producing
  perpetual `environment_surface_drift` (linked mode: the symlink is replaced
  by a file) that the operator will read as Curator flakiness. The
  foreign-manager stop exists precisely to prevent this and only catches the
  symlink-shaped half of the ecosystem.
- **Proposal**: (a) extend the inventory with a documented best-effort
  heuristic — presence of well-known foreign-manager state files
  (`~/.local/share/chezmoi`, `~/.config/home-manager`, …) elevates the notice
  from "unmanaged file" to "a dotfile manager appears to manage this machine;
  it will overwrite managed surfaces" without blocking; and (b) specify that
  repeated drift on the same surface (N repairs in M days is
  implementation-defined) SHOULD be reported as suspected external writer.
  Wording as SHOULD keeps the closed surface intact.
- **Can follow** (revision 2 acceptable; the drift path is safe, just noisy).

---

## Journey 2 — daily use: native tool vs `curator run` into a managed home

### F4. CRITICAL — the claude_code macOS credential passthrough row is factually wrong, and the spec nowhere admits that a fresh managed home replays login, onboarding wizard, per-project trust, and MCP approvals
- **Area**: credential passthrough + managed-home provisioning; environments
  §7.4, §8.1; decision 0010 Decision 7; manager §12.4.
- **Quote**: "`claude_code` | macOS: none (Keychain is ambient)" (§7.4) and
  "Default is `shared`: every profile home reuses the operator's existing
  authentication." (0010 Decision 7)
- **Evidence (verified-on-binary, Claude Code 2.1.257, macOS)**: with the
  operator fully logged in natively, `CLAUDE_CONFIG_DIR=<fresh dir> claude
  config get theme` prints **"Not logged in · Please run /login"** and
  creates a fresh `.claude.json` containing only
  `firstStartTime/machineID/userID/projects…`. The native `~/.claude.json`
  carries `oauthAccount`, `hasCompletedOnboarding`, `tipsHistory`, and
  per-project `hasTrustDialogAccepted`, `allowedTools`, `mcpServers`,
  `enabledMcpjsonServers`, `hasClaudeMdExternalIncludesApproved`. Login
  *state* (account record) lives in `.claude.json` inside the config dir;
  the Keychain item alone does not make a fresh home authenticated.
- **Why critical**: (1) The §7.4 row is a closed conformance-frozen surface
  ("MUST implement the complete closed revision-1 surface exactly") and it is
  demonstrably insufficient on the platform it names. The first
  `curator run claude_code --profile companyA` on a Mac lands the operator in
  `/login`, then the first-run wizard, then — per project — trust dialogs and
  MCP re-approval, for every profile × environment pair. (2) The same class
  hits codex: `CODEX_HOME/config.toml` holds the `projects` trust table,
  model choice, and MCP servers, and is neither a managed surface nor a
  passthrough entry, so every managed codex home re-asks project trust and
  loses the operator's config.toml entirely (fresh `CODEX_HOME` verified
  "Not logged in" pre-passthrough; `auth.json` passthrough fixes auth only).
  (3) Claude's `settings.json` (permissions allowlist, hooks, model,
  statusline) is likewise absent from every managed home. The spec's silence
  means the first real-world day-2 experience — a wall of repeated wizards —
  is undocumented, unwarned, and will be attributed to Curator.
- **Proposal**:
  1. Re-verify and rewrite the claude_code §7.4 row before freezing: macOS
     needs at minimum a **seeded** class — a new passthrough entry kind
     "seed-once at provisioning, never refreshed, never drift-checked" —
     covering the auth-relevant `.claude.json` members (or the whole file,
     seeded then owned by the tool). If seeding `.claude.json` is rejected,
     the row must say "macOS: interactive `/login` required once per managed
     home" so the claim is honest and the launcher/`env resolve` can print a
     first-use notice.
  2. Add a normative "fresh-home state" paragraph to §8.1 enumerating what a
     newly provisioned managed home will re-ask per adapter (login where
     passthrough is none/insufficient, first-run wizard, per-project trust,
     MCP approvals), and require the first `env resolve` of a newly
     provisioned home to report it (one-line notice, not a failure).
  3. Decide explicitly, per adapter, whether tool *settings* files
     (`settings.json`, `config.toml`) are (a) out of scope forever with the
     consequence documented, or (b) a named revision-2 surface ("Settings
     fragments" already sits in the 0010 phasing table — link the two so the
     day-2 gap is visibly owned).
- **MUST before implementation** — highest cost-of-being-wrong in this
  review: a frozen wrong adapter row plus an unowned first-run wall.

### F5. MEDIUM — two doors, two session histories for the SAME profile; the spec sells the split as a feature and never warns about the same-profile case
- **Area**: managed homes vs in-place; 0010 Decision 7; environments §8.1.
- **Quote**: "Session state living inside a profile home is a feature: a
  session created under profile P resumes under profile P's home" (0010 §7).
- **Why weak** (journey-analysis): with in-place mode active, a native
  `claude` session and a `curator run claude_code` session of the *same
  current profile* write to two different homes (`~/.claude` vs
  `<curator-home>/environments/<profile>/claude_code/`). `claude --resume`
  shows a different session list depending on which door was used; trust and
  MCP approvals diverge per door too (F4). Operators will lose yesterday's
  session and not know why. Nothing in 0010, environments.md, or
  cli/curator.md names this two-door split for the same profile.
- **Proposal**: an informative paragraph in cli/curator.md §Environment
  profiles and one sentence in 0010 Decision 6: "a native launch and a
  launcher launch of the same profile use different homes and keep separate
  session histories; pick one door per environment for daily work." Optional
  rev-2: `env status` reports per-home recent-session counts; launcher prints
  the managed-home path on first launch per home.
- **Can follow** (documentation), but the acknowledgment sentence should land
  with revision 1 text.

---

## Journey 3 — switching

### F6. HIGH — `profile use` failure semantics are unspecified below the per-entry level: a mid-switch failure strands the machine half-switched with an ambiguous recorded current
- **Area**: switching, environments §9.2, manager §12.3.
- **Quote**: "re-materializes every in-place surface of every registered
  adapter … atomically per entry, under the manager-home mutation lock,
  journaled like any other manager-home transaction (manager §2.5); 2.
  updates the recorded current profile for the affected scope".
- **Why weak** (spec-text): atomicity is declared per *entry*; the switch as
  a whole spans multiple native homes **outside** the manager home, where the
  §2.5 manager-home journal cannot roll anything back. When adapter 3 of 4
  fails (`environment_surface_unmanaged_conflict` on a file someone created
  since onboarding, `environment_backup_exists` per F1, an I/O error), the
  spec does not say: does materialization continue for adapter 4? Is step 2
  (record current) executed, skipped, or executed per-scope? The machine ends
  with claude_code on B, codex on A, and a recorded current of — unspecified.
  `env status` will show non-current rows, but the operator's mental model
  ("current profile = X") has no defined truth value, and recovery is
  unnamed.
- **Proposal**: specify: (a) the switch attempts every entry, collecting
  per-adapter results (partial progress is reported, not hidden); (b) the
  recorded current updates only when every in-place surface of the scope
  materialized — otherwise the previous current stands and the diagnostic
  names every failed surface and the recovery (`profile sync` after the
  conflict is resolved); (c) mixed materialized pins are guaranteed
  non-current rows in `env status`. (a)+(b) makes "recorded current implies
  fully materialized" an invariant an implementation can test.
- **MUST before implementation** — transactional shape drives the
  implementation architecture and cannot be retrofitted.

### F7. LOW — no specified way to clear a scoped current back to "follows machine default"
- **Area**: scoped switching, environments §9.3.
- **Quote**: "A scoped switch records a per-scope current profile. `env
  status` and `profile list` MUST surface every scope whose current profile
  differs from the machine default."
- **Why weak** (spec-text): after `profile use companyA --env claude_code`,
  the operator wants the scope to *follow the machine default again* —
  distinct from "set scope to the profile that happens to be the default
  today" (the record would silently diverge on the next machine-level
  switch… or not; unspecified whether a scope record equal to the default is
  kept). Implementers will guess divergently.
- **Proposal**: specify that a scoped switch naming the machine-default
  profile removes the scope record (scope follows the default thereafter),
  or add `profile use --env <id> --clear`. One sentence either way.
- **Can follow**.

---

## Journey 4 — upgrading a branch-tracking profile

### F8. HIGH — no `profile update`/`upgrade` lifecycle exists: nothing is specified to move a branch-tracking pin, and the skill-side `update`/`upgrade` commands are ambiguously adjacent
- **Area**: profile lifecycle; environments §9 (whole), cli/curator.md.
- **Quote**: environments §1: "A directly installed profile MAY track a
  branch … and the resolved commit is recorded as the **effective pin**."
  §9.4: "`profile sync` re-materializes every installed profile" (from the
  store — same pin). cli/curator.md: "`curator update` | Fetch configured
  source repositories"; "`curator upgrade` | Fetch only the selected
  dependency closure, then install". No profile command re-resolves a ref.
- **Why weak** (spec-text + journey-analysis): branch tracking is explicitly
  allowed, and the *only* specified path that could ever advance the pin is
  re-running `profile install <url>` — whose semantics against an
  already-installed repository (same names, new commit) are themselves
  unspecified: does it replace the pin? Refuse? Install duplicates? Every
  question an operator asks in week 2 — "how do I get the new company
  context?", "what re-materializes?", "what happens to my managed-home
  sessions?", "do overlay pins in existing markers/fragments go stale?" — has
  no answer. Also unstated: whether skill-side `update`/`upgrade` touch
  profile sources (they must not, but nothing says so).
- **Proposal**: specify `profile update [<name> | --all]`: re-resolve each
  declared ref (branch → new head; tag under the unchanged strict-tag
  policy; `revision` and `path`/`local` are no-ops reported as pinned),
  strict re-audit of the new snapshot (§9.1 unchanged — a blocking finding
  leaves the old pin in place and reports it), store the new entry, update
  the effective pin, then re-materialize in-place surfaces for every scope
  currently on that profile and mark managed homes stale for repair-on-next-
  resolve. Environment-owned mutable state is untouched (restate §7.4/§9.2
  discipline). State that re-`install` of an installed source is equivalent
  to `update` for that source (or refuse with a pointer — pick one). State
  that skill-scope `update`/`upgrade` never move profile pins. Old store
  entries become GC-eligible when unreferenced.
- **MUST before implementation** — a rev-1 capability that advertises branch
  tracking but cannot ever advance a branch is broken on arrival; and the
  GC live-root story depends on knowing when old pins die.

### F9. HIGH — no profile removal lifecycle; managed homes full of session history have no specified fate, and composition/current references can dangle
- **Area**: profile lifecycle; environments §9, §12; cli/curator.md.
- **Quote**: there is no removal operation anywhere in environments §9,
  manager §12.3, or the cli/curator.md command table. §12 GC: "its live roots
  additionally include … every managed home and in-place surface set
  referenced by a valid environment marker" — the marker lives *inside* the
  managed home, so an orphaned home self-roots forever.
- **Why weak** (journey-analysis): operators leave companies. The journey
  "remove the companyA profile" has to answer: (a) refuse or cascade when
  companyA is a current profile in some scope or named in another profile's
  overlay list (`environment_composition_invalid` would start firing on
  activation); (b) what happens to
  `<curator-home>/environments/companyA/*/` — these hold **operator data**
  (session logs, history) that "is never touched by materialization,
  refresh, switch, or garbage collection" — deleting silently is data loss,
  keeping silently is an invisible disk leak that GC is spec-barred from
  collecting; (c) what un-declares it from machine config. Nothing answers
  any of these.
- **Proposal**: specify `profile remove <name>`: refuses while the profile is
  any scope's current or appears in any overlay declaration, naming every
  referrer; on success removes the install declaration, leaves managed homes
  in place, and prints their paths with a session-data notice; an explicit
  `--purge` deletes the homes after printing what will be lost (mutable
  state included) — the one deliberate deletion door, with the backup
  discipline not applying to tool-owned mutable state. Store entry becomes
  GC-eligible; orphaned managed homes whose profile is uninstalled stop
  being GC live roots *only* under `--purge` (otherwise retained and
  reported by `env status` as orphaned).
- **MUST before implementation** — GC live-root rules are being implemented
  from this text; retrofitting removal changes them.

### F10. MEDIUM — un-managing entirely (rollback to native) is a journey with no owner
- **Area**: lifecycle; environments §8.3/§9.5; cli/curator.md.
- **Quote**: backups exist (§8.3) and are never overwritten or collected,
  but no operation restores them; nothing removes markers or managed
  surfaces from native homes as an end state.
- **Why weak** (journey-analysis): "I tried Curator, I want my machine back"
  must be answerable without archaeology. Today it is: manually delete
  symlinks, manually copy `.agent-environment-backup/*` back, manually
  delete markers — undocumented, error-prone, and F1 makes the next attempt
  at management jam. Backups also live *inside* the native homes; an
  operator who deletes `~/.claude` wholesale to "reset" destroys the backups
  with it.
- **Proposal**: either specify a minimal `env unmanage [<env-id>]` (remove
  marker-recorded surfaces, offer `--restore-backups`, leave mutable state
  and unmanaged files, print every action) in revision 1, or explicitly
  defer it with a named story and add an informative "manual rollback"
  section to cli/curator.md so the journey has *written* steps. Silence is
  the only unacceptable option.
- **Can follow** (with the documented-manual-steps half done now).

---

## Journey 5 — tool auto-updates changing home layouts

### F11. MEDIUM — adapter facts are verified against pinned tool releases, but tools self-update weekly and no surface records or checks tool-version compatibility
- **Area**: adapter registry; environments §7, 0010 Decision 3.
- **Quote**: "the exact spellings verify against the pinned tool releases
  before the conformance vectors freeze" (§7.3); local evidence in 0010 is
  "Claude Code 2.1.251" — the machine already runs 2.1.257.
- **Why weak** (journey-analysis): every adapter fact — home-relative
  targets, `.claude.json` gating, `AGENTS.override.md` shadowing, flag
  spellings — is a fact about a tool version, and the tools ship weekly with
  auto-update on. When a vendor moves a surface (pi's discovery chain has
  already changed across its short life), the failure is maximally silent:
  Curator writes a file the tool no longer reads, and `env status` reports
  all-current while the agent sees nothing. That is the shadow-inert failure
  class with zero diagnostic.
- **Proposal**: (a) add an informative recorded-verified-version note per
  adapter in the registry (spec appendix, not wire data); (b) `env status`
  SHOULD report the detected tool binary version per adapter when
  discoverable, explicitly best-effort and non-normative; (c) name the
  revalidation procedure: a discovered vendor layout change is a
  specification erratum with a defined fast path (registry row fix +
  vector re-freeze), not an ordinary revision. Keeps the closed set closed
  while giving day-2 operators a signal.
- **Can follow** (revision 2), but record the intent in revision-1 text so
  implementers leave room for the status column.

### F12. MEDIUM — composed root context vs tool-side size caps: codex ships a 32 KiB instructions cap and materialization has no size surface at all
- **Area**: materialization/adapters; environments §5, §7.
- **Quote**: nothing — no size word exists in environments.md. Verified on
  binary: codex 0.151.0 embeds default `project_doc_max_bytes = 32768`
  (strings of the shipped binary; config key present in ConfigToml).
- **Why weak** (verified-on-binary + journey-analysis): composition is a
  headline rev-1 feature — base + overlays + chapters + header concatenated.
  A company base plus a personal overlay walks past 32 KiB without effort.
  Where a tool caps or truncates its root/instructions read, the truncated
  tail is precisely the *later* profiles — under default
  `later-overrides-earlier` precedence, the content the operator declared to
  *win* is what silently disappears. No materialization warning, no status
  column, no adapter field acknowledges caps. (Whether codex applies this
  exact key to the `CODEX_HOME/AGENTS.md` global doc needs implementation-
  time verification — the proposal below covers the research either way.)
- **Proposal**: add an OPTIONAL adapter-registry field
  `root_context_size_advisory_bytes` recorded from the tool's documented cap;
  materialization and `env status` warn (never fail) when composed output
  exceeds it, naming the adapter, the size, and the cap. Add per-tool
  truncation semantics to the open-question research list that gates the
  vector freeze.
- **Can follow**, but the registry field should be reserved in revision-1
  schema thinking now — the registry is a closed surface (see F16 for the
  same structural point about the fragment).

---

## Journey 6 — `env status --check` in CI-like use

### F13. HIGH — shadow-inert is simultaneously a warning (§7.5) and a non-current `--check` failure (§12): a direct normative contradiction that makes `--check` permanently red for a legitimate pi setup
- **Area**: status; environments §7.5 vs §12; manager §12.7.
- **Quote**: §7.5/§7.7: "declared shadowing path exists (**warning**) |
  `environment_shadowing_path_present`". §12: "A drifted, missing,
  **shadow-inert**, or unreadable state is non-current" and "`--check`
  returns non-zero when any row is non-current."
- **Why weak** (spec-text): the same machine state is classified as a
  warning by one section and as a check-failing non-current row by another.
  Concretely: a pi user keeping a hand-written `AGENTS.override.md` — the
  tool's own supported personal-override mechanism, deliberately unmanaged,
  deliberately never touched by Curator — fails `env status --check`
  forever, with no acknowledgment mechanism, no scoping flag, and no way to
  make CI or a cron health check green again short of deleting their own
  file. `--check` (the one scriptable gate, exit-code contract in
  cli/curator.md) becomes unusable on that machine, so operators stop
  running it — the worst outcome for a drift-detection feature.
- **Proposal**: resolve the contradiction in one direction and add the
  acknowledgment valve: shadow-inert **is** non-current by default (the
  surface genuinely is inert), the §7.7 row drops the "(warning)" marker for
  consistency, and machine configuration MAY record a per-path shadowing
  acknowledgment ("this override is deliberate") that downgrades exactly
  that row to a reported-but-current warning. Deliberate overrides become an
  operator decision with a record, `--check` stays usable, and the default
  stays fail-closed.
- **MUST before implementation** — it is a normative contradiction on the
  frozen status/exit-code contract; whichever way an implementer guesses,
  half the spec text is violated.

---

## Journey 7 — opencode

### F14. MEDIUM — the referenced form makes `opencode.json` instructions-only forever, which silently forbids every other opencode config (MCP included) in that managed home
- **Area**: referenced form, environments §5.3; manager §6 (opencode MCP
  surface is `opencode.json`).
- **Quote**: "The managed `opencode.json` is fully manager-authored and its
  bytes are exact: … the object whose single member, `instructions`, is the
  ordered list … — no other member" (§5.3).
- **Why weak** (spec-text + journey-analysis): `opencode.json` is the tool's
  *entire* configuration surface — including its MCP table (`mcp`, per
  manager §6). In a referenced-form managed home the operator can never add
  MCP servers, theme, providers, or keybinds: editing the managed file is
  drift, and repair reverts it. The unmanaged-file fallback protects the
  *native* home (good — `environment_form_unavailable` → monolithic), but
  inside a managed home the manager authored the file first, so the lockout
  is total and nothing in the spec states the consequence. The operator
  discovers it as "my MCP config keeps disappearing" — repair eating their
  edit is working-as-specified and reads as data loss.
- **Proposal**: minimum: a consequence sentence in §5.3 — "a referenced-form
  managed opencode home cannot carry any other `opencode.json`
  configuration; where the operator needs tool configuration in that home,
  the monolithic form is the supported shape" — plus the same sentence in
  manager §12.1. Better: demote opencode `referenced` to revision 2 pending
  a merge-safe design (e.g. manager-authored `instructions` member inside an
  otherwise operator-owned file is explicitly *not* possible under the
  ledger discipline — say so and own it).
- **Can follow** (monolithic is the default form), but the consequence
  sentence belongs in revision-1 text.

### F15. HIGH — XDG seed links are a point-in-time snapshot: config dirs created after provisioning never reach managed opencode homes, so XDG tools inside launched sessions silently see empty or split-brain config
- **Area**: opencode XDG mechanism, environments §7.1.
- **Quote**: "the manager seeds it with symlinks to every entry of the
  operator's effective XDG config home except `opencode/`. Seed links are
  recorded in the environment marker; refresh adds missing **recorded**
  seeds and removes recorded seeds whose target no longer exists, and MUST
  NOT touch an entry the marker does not record."
- **Why weak** (spec-text): the reconciliation rule only heals seeds that
  were recorded at provisioning time. `~/.config/gh/` created next month
  (operator installs and authenticates the gh CLI) is never seeded: inside
  every opencode-launched session, `gh` resolves `XDG_CONFIG_HOME` to the
  managed parent, finds no `gh/`, and behaves logged-out — or worse, writes
  a *fresh* config into the managed parent, which then shadows the
  operator's real one forever (the marker doesn't record it, so refresh must
  not touch it, and no diagnostic exists for "unrecorded entry in the
  managed parent shadows an operator entry"). Every XDG-conforming tool the
  agent invokes is exposed; the failure is per-tool, silent, and looks like
  the tool's bug.
- **Proposal**: change refresh to *reconcile against the operator's current
  XDG home*: add-and-record seeds for entries newly present in the
  operator's config home (still excepting `opencode/` and never replacing an
  existing non-seed entry in the managed parent — that case is reported as
  `environment_seed_shadowed` or similar, named, not merged). Run
  reconciliation on `sync`, `use`, and `resolve`-repair. This keeps the
  MUST-NOT-touch-unrecorded discipline for tool-owned entries while closing
  the staleness hole for operator entries.
- **MUST before implementation** — the seeding rule as written bakes the
  failure into the closed rev-1 opencode adapter; the fix changes marker
  semantics (recorded seed set becomes dynamic), which is wire-adjacent.

---

## Journey 8 — skills commands and shims inside managed-home launches

### F16. HIGH — the profile's command surface is unreachable (or wrong-profile) inside managed-home launches, and the frozen fragment schema has no room to ever fix it
- **Area**: profile-scoped skills × launcher; environments §9.4, §10.2/§10.3;
  cli/curator.md "Developer shell"; grounded in reference implementation
  `internal/globalbins` (forwarding shims staged into one user-bin dir).
- **Quote**: "Global installation publishes forwarding shims to a safe
  existing user-bin directory when possible" (cli/curator.md). The fragment:
  "`env` maps each registry-declared variable name for the environment to a
  managed-home path" (§10.2) — nothing else; schema `propertyNames` enum is
  exactly the four home variables.
- **Why weak** (journey-analysis): skills are not only context — they carry
  commands. The global-scope shims on `PATH` are a machine-level singleton
  that can only track *one* profile's rendered command set (presumably the
  machine current). Launch `curator run codex_cli --profile companyB` while
  the machine current is `personal`: the child inherits `PATH`, so the agent
  — whose materialized companyB context references companyB's skill
  commands — either finds nothing on PATH or, worse, finds `personal`'s
  bytes under the same command name and runs the wrong version silently.
  That is precisely the class of silent-wrong-tool failure this protocol
  exists to prevent, and neither environments.md, manager §12, nor the
  launcher SPEC mentions PATH, bins, or shims for managed homes even once.
  And because `launch-env-fragment-v1` is a closed object with a closed
  `env` name enum, the fix cannot be added later without a schema-breaking
  widen.
- **Proposal**: decide now, one of: (a) each managed home gets a
  manager-owned `<home>/bin` carrying that profile's forwarding shims, and
  the fragment gains a declared, closed `path_prepend` member (single
  manager-owned absolute path below the environments root — the §10.3
  boundary argument extends cleanly: profile bytes still cannot name a
  variable or escape the root); the launcher prepends it. Or (b) revision 1
  explicitly declares profile skill *commands* unavailable in managed-home
  launches — a named limitation in §9.4 and the launcher SPEC, with the
  shims documented as following the machine current profile only — so the
  gap is a documented tradeoff instead of a discovered one.
- **MUST before implementation** — the fragment schema freezes at revision
  1; option (a) is impossible to retrofit compatibly, so the decision
  itself cannot wait even if the answer is (b).

### F17. MEDIUM — hybrid scope is never reconciled with profile-scoped globals
- **Area**: scopes; manager §4.3 vs environments §9.4.
- **Quote**: manager §4.3: "Precedence is project, hybrid, global." §9.4:
  "The existing machine-global skill scope becomes profile-scoped." Hybrid
  appears nowhere in environments.md or manager §12.
- **Why weak** (spec-text): "global" is now a moving target — per-scope
  current profiles can differ per environment. Is hybrid precedence "project,
  hybrid, *current-profile-of-the-applicable-scope*"? Do hybrid-only closure
  nodes re-render when the profile switches (they render "once in a machine
  store with the machine locale")? Can a hybrid manifest target a managed
  home's project? Unstated; implementers will wire something.
- **Proposal**: one paragraph in §9.4: hybrid scope is orthogonal to
  profiles; precedence is project > hybrid > the current profile's global
  scope as resolved per §9.3 scoping; hybrid manifests never participate in
  profile switching or composition. (Or defer hybrid×profiles explicitly.)
- **Can follow** (one paragraph, but write it before the skills-resolution
  code path is built).

---

## Journey 9 — team distribution and machine configuration

### F18. HIGH — composition, precedence, forms, target participation, and `isolated` auth have no CLI surface at all: a headline rev-1 feature is operable only by hand-editing an implementation-specific config file
- **Area**: machine configuration; environments §6, §7.2, §7.6, §7.4;
  cli/curator.md command table.
- **Quote**: "A machine MAY declare, per installed profile, an ordered
  **overlay list** … The declaration lives in machine configuration only"
  (§6). The cli/curator.md table contains `profile
  install|list|use|sync`, `env resolve|status` — no command declares an
  overlay, a precedence, a form, a target enable, `system_prompt_files`, or
  an `isolated` credential pair.
- **Why weak** (journey-analysis): the composition walkthrough in 0010
  ("activating companyA on one machine composes personal into it") cannot be
  performed by any specified command. Same for "referenced form for
  claude_code", "target off for Xcode", "isolated auth for companyA ×
  claude_code". `~/.curator/config.json` is explicitly "not protocol wire
  identifiers" and implementation-specific — so the only path to rev-1's
  differentiating features is editing an undocumented file. Worse
  (verified-on-schema): `manager-config-v1.schema.json` is
  `"additionalProperties": false` and carries **none** of these keys (its
  full property set: schema_version, skills_root, default_agents,
  preferred_locale, adapter_mode, worktree_alias_pattern, projects,
  allowed_sources, audit, audit_registries, disable_builtin_registries) —
  so a conforming implementation cannot even store composition, precedence,
  form, target participation, `system_prompt_files`, `isolated` pairs, or
  per-scope currents in the versioned config object; it is forced to invent
  an unversioned side file, violating the repo's own strict-surface
  discipline and creating migration debt on day one. For a team
  ("here is our recommended setup"), there is additionally no
  export/apply/doctor story: composition/forms/targets are per-machine and
  unversioned by design, but the design never says how a team distributes a
  recommended machine setup safely (a README of hand steps is where this
  lands today).
- **Proposal**: (a) version the machine-config surface first — either
  manager-config schema 2 with an `environments` object or a separate
  versioned `environments-config-v1` object under `schemas/v1/` naming every
  knob (composition + precedence, form per env, target participation,
  `system_prompt_files`, credential mode per pair, machine and per-scope
  currents) with strict validation — then extend cli/curator.md (informative,
  like the rest of the table) with `curator profile compose <name>
  [--overlays o1,o2] [--precedence …] [--clear]` and `curator env config` (or
  equivalent) covering form, target participation, `system_prompt_files`, and
  credential `shared|isolated` — and name the corresponding machine-config
  keys in manager §12 so implementations converge; (b) for teams, add an informative
  paragraph: machine settings are deliberately per-machine; the supported
  distribution shape is the profile repository plus a documented
  `curator`-command bootstrap script — and reserve a rev-2
  "machine-settings fragment" story in the 0010 phasing table so it is
  visibly owned.
- **MUST before implementation** (part a) — composition/forms/isolation are
  rev-1 surfaces; shipping them without an operator door means every early
  adopter script hard-codes a config file layout that was never a contract.

---

## Journey 10 — Windows

### F19. LOW — `--format shell` is POSIX-only; Windows operators (PowerShell) cannot eval a fragment
- **Area**: `env resolve`, environments §10.1.
- **Quote**: "`--format shell` prints one POSIX `export NAME='value'` line
  per variable with single-quote escaping."
- **Why weak** (spec-text): the four variables inject identically on Windows
  (0010 open question 5), but the eval-oriented output format has no
  PowerShell shape, so the documented "a script to `eval`" journey dies on
  the platform where it is most needed (no direnv culture).
- **Proposal**: add `--format pwsh` (`$env:NAME = '…'`) or document
  explicitly that Windows automation must consume `--format json`.
- **Can follow** — and note that the remaining Windows substance (symlink
  privilege → `copy_fallback` marker records; `XDG_CONFIG_HOME` semantics on
  Windows; claude_code credential shape) is already correctly parked in 0010
  open question 5 with the vector freeze held on it; no new finding there.

---

## Journey 11 — system-prompt channels verified against binaries

### F22. HIGH — the §7.3 pi flag spellings do not exist in pi 0.84.2; Decision 0010's "verified in 0.84.2" claim does not reproduce, and applying the declared replace channel would silently ship a file path as the entire system prompt
- **Area**: system-prompt channels; environments §7.3; decision 0010 Decision 2;
  launcher SPEC §5.
- **Quote**: §7.3 pi row: "`flag`/`append`: `--append-system-prompt-file`;
  `flag`/`replace`: `--system-prompt-file`". Decision 0010: "`pi` takes the
  same flags and additionally reads agent-dir `APPEND_SYSTEM.md` … (verified
  in 0.84.2)".
- **Evidence (verified-on-binary, pi 0.84.2, this machine)**: `pi --help`
  declares `--system-prompt <text>` ("System prompt (default: coding
  assistant prompt)") and `--append-system-prompt <text>` ("Append **text or
  file contents** to the system prompt"). Neither `--system-prompt-file` nor
  `--append-system-prompt-file` exists. The append channel exists under a
  different spelling with path-or-text heuristic semantics; the replace
  channel has **no file-taking flag spelling at all**.
- **Why weak**: this is the negative-evidence shape *capability claim that
  does not reproduce* inside an accepted decision — the exact version the
  claim names is the version that falsifies it. And the failure mode is not a
  clean error: if the launcher applies the declared replace descriptor by
  passing the fragment's system-prompt *path* to the text-taking
  `--system-prompt` flag, the model's entire system prompt becomes the
  literal string `/…/.agent-context/system-prompt.md`. The launch succeeds,
  the §5.2 warning prints, and the customization is garbage — no diagnostic
  can catch it. (Contrast claude_code: `--system-prompt[-file]` and
  `--append-system-prompt[-file]` both verified present on 2.1.257, so the
  claude rows hold.) The §7.3 "spellings verify before the vectors freeze"
  hedge covers the spelling; it does not cover a decision text asserting a
  verification that never happened, and it does not cover the
  text-flag-misfire class.
- **Proposal**: (1) correct the Decision 0010 verification claim explicitly
  (record the correction; do not silently edit an accepted decision). (2)
  Rewrite the §7.3 pi row from binary evidence: `flag`/`append` spelling
  `--append-system-prompt` with a note that the tool heuristically accepts a
  path or literal text; drop the pi `flag`/`replace` channel — `file`/`replace`
  `SYSTEM.md` remains pi's only replace path — or mark it
  unverified-absent. (3) Add a normative admission rule to §7.3: a
  `flag`-kind channel is admissible only with verified evidence that the flag
  accepts a **file path**, because a text-flag misfire is silent by
  construction.
- **MUST before implementation** — the row is closed conformance surface (the
  same class as F4), the launcher's §5 application logic is built directly on
  it, and the false verification claim should be corrected immediately.

---

## Cross-cutting / consistency

### F20. LOW — launcher stub reports specification version 0.1.0-draft while SPEC.md is 0.1.2-draft
- **Area**: curator-agent-launcher; SPEC.md §8 vs cmd/curator-run/main.go.
- **Quote**: SPEC.md: "the current version is **`0.1.2-draft`**"; "the stub
  reports the specification version only." main.go: `specVersion =
  "0.1.0-draft"`; README.md: "The specification is `0.1.0-draft`."
- **Why weak** (verified-on-source): the one thing the stub exists to report
  is stale, and the README disagrees with the SPEC it sits beside. Trivial,
  but it is a live example of exactly the version-skew class F11 worries
  about.
- **Proposal**: bump the stub const and README line; consider deriving the
  README status line from the SPEC header at release time.
- **Can follow**.

### F21. LOW — the launcher's unsuppressible per-launch system-prompt warning will train pi file-channel users to ignore stderr
- **Area**: launcher SPEC §5.2.
- **Quote**: "The warning is not suppressible in revision 1."
- **Why weak** (journey-analysis): for an operator who *deliberately*
  configured `system_prompt_files: append` at machine level, every launch
  prints the same three-part warning forever. Warnings that never change
  stop being read; the one day the warning differs (an unmanaged SYSTEM.md
  appeared — the actual attack/mistake case §5.1 exists for) it is
  camouflaged in the noise.
- **Proposal**: keep rev-1 as specified (deliberately conservative — fine),
  but note in §5.2's open items a rev-2 acknowledgment: a machine-config
  acknowledgment of a *specific* (profile, environment, channel, file-hash)
  tuple suppresses only that exact warning; any change in the tuple
  resurfaces it.
- **Can follow**.

---

## MUST-before-implementation list (ranked by cost-of-being-wrong)

1. **F4** — claude_code macOS passthrough row is verified-wrong on the
   shipping binary; passthrough tables are closed conformance surface.
   Re-verify, add the seeded class or the honest interactive-login wording,
   and own the fresh-home first-run wall in normative text.
2. **F22** — the §7.3 pi flag rows are verified-wrong on pi 0.84.2 and
   Decision 0010 asserts a verification that does not reproduce; correct the
   claim, rewrite the row from binary evidence, and add the
   flag-must-accept-a-file-path admission rule before the launcher's §5
   application logic is built on it.
3. **F16** — decide the managed-home command/PATH story now: the frozen
   `launch-env-fragment-v1` schema cannot be widened compatibly later, so
   even the "documented limitation" answer must be chosen before freeze.
4. **F13** — shadow-inert warning-vs-non-current contradiction: two
   normative sections disagree about the `--check` exit contract; resolve
   and add the acknowledgment valve.
5. **F8** — specify `profile update`: branch tracking is allowed but no
   operation can ever move a pin; also bounds the GC lifetime of store
   entries.
6. **F9** — specify `profile remove`: GC live-root rules and the
   session-data fate of managed homes depend on it.
7. **F6** — specify whole-switch failure semantics for `profile use`
   (per-adapter results; current recorded only for fully-materialized
   scopes); drives implementation architecture.
8. **F15** — XDG seed reconciliation: the as-written rule bakes silent
   stale-seed failures into the closed rev-1 opencode adapter and touches
   marker semantics.
9. **F18(a)** — schema + CLI surface for
   composition/precedence/form/target/isolated: the versioned
   manager-config object is closed (`additionalProperties: false`) with none
   of these keys, so rev-1 headline features currently have neither a
   storage contract nor an operator door.
10. **F1** — versioned backup paths: the never-overwrite rule as written
    guarantees a dead end on the second management cycle; backup layout is
    written into homes on first contact.

Can-follow (with revision-1 text touch-ups noted inline): F2, F3, F5, F7,
F10, F11, F12, F14, F17, F19, F20, F21.

## What this review did NOT re-litigate

Accepted decisions left alone (no concrete failure found): the two-mode
(in-place + managed-home) requirement; the ax always-through-when-configured
no-fallback posture; the umbrella subcommand trust model; system-prompt
inert-by-default and launcher-only activation; no-templating IR;
determinism/byte-equality discipline (the environments.json vectors — 4
header + 11 materialization cases including referenced/opencode/system-prompt
and zero-module shapes — cover the deterministic surfaces well); the
`path`-source snapshot discipline; the onboarding import design (§9.6 reads
complete and its absence/failed-read separation is consistently carried).

## Evidence appendix (binary probes, this machine, 2026-09-02)

- `claude --version` → 2.1.257; `codex --version` → codex-cli 0.151.0;
  `pi --version` → 0.84.2; `gemini --version` → 0.54.4; opencode absent.
- `CLAUDE_CONFIG_DIR=<fresh> claude config get theme` → "Not logged in ·
  Please run /login"; fresh home contains `.claude.json` (keys:
  firstStartTime, firstStartVersion, machineID, …, userID, projects),
  `backups/`, `projects/`, `sessions/`. Native `~/.claude.json` carries
  `oauthAccount`, `hasCompletedOnboarding`, per-project
  `hasTrustDialogAccepted`, `allowedTools`, `mcpServers`,
  `enabledMcpjsonServers`, `hasClaudeMdExternalIncludesApproved`.
- `CODEX_HOME=<fresh> codex login status` → "Not logged in".
- `strings $(which codex)` → `project_doc_max_bytes = 32768` (default),
  `model_instructions_file` present in the config-key table.
- `claude --help` → `--system-prompt[-file]`, `--append-system-prompt[-file]`
  both present (spec §7.3 spellings hold on 2.1.257).
- `pi --help` → `--system-prompt <text>` ("System prompt (default: coding
  assistant prompt)") and `--append-system-prompt <text>` ("Append text or
  file contents…"); no `--system-prompt-file`, no
  `--append-system-prompt-file` (F22 — the §7.3 pi spellings do NOT hold on
  0.84.2, the version Decision 0010 claims verified them).
- `~/.pi/agent/` holds `auth.json`, `settings.json`, `models.json`,
  `sessions/`, and pi-provisioned `bin/`, `npm/`, `tools/` trees — a fresh
  `PI_CODING_AGENT_DIR` loses settings/model config and re-downloads
  tooling per managed home (extends F4's fresh-home class to pi).
- `~/.codex/config.toml` carries `[projects."…"] trust_level` rows and
  `sessions/` holds history under `CODEX_HOME` (F4/F5 grounding).
- `schemas/v1/manager-config-v1.schema.json` → `"additionalProperties":
  false`; property set contains no environments key (F18 grounding).
- curator reference implementation: no environments/profile code present
  (internal/ has no profile, envhome, or fragment packages; no
  CLAUDE_CONFIG_DIR/Profilefile references) — review is pre-implementation,
  as intended. `internal/globalbins` confirms machine-singleton forwarding
  shims (F16 grounding).
