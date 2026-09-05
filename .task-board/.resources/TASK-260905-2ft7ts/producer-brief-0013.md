# Producer brief: Decision 0013 — execution ownership and launch plans

## Where and what

- Worktree `/Users/iv/Developer/ReluxWorks/.worktrees/curator-spec-decision-0013`,
  branch `draft/decision-0013-execution-ownership`, base `b4f29cd` (= curator-spec main).
- Create exactly one file: `decisions/0013-execution-ownership-and-launch-plans.md`.
  Touch nothing else in the repository. The story worktree task-board provisions
  for this run stays untouched; the document lives only in the worktree above.
- House style: mirror `decisions/0012-context-packages-and-semver-locks.md` and
  `decisions/0010-agent-environment-profiles.md` — Status, Context, Decision
  (numbered subsections), Rejected alternatives, Compatibility impact, Security
  impact, Consequences, Open questions. English. Dense normative prose; RFC 2119
  words only where the sentence is a rule. Every `§` citation names its document
  and is verified against the cited text at the cited commit before you write it.
- Deliverable: one signed commit (`git commit -S`; run `git log --show-signature -1`
  and paste the verification line into your report). Do not push, tag, open a PR,
  or mark the task done. Attach `TASK-260905-2ft7ts_drafting-report.md` as an
  outcome resource: file, commit hash, a table of every contract item below with
  the section that specifies it, every fact you verified (source + commit or
  installed binary version), and anything you could not verify. Then hand off.

## Sources (read before writing; never restate divergently)

Read-only checkouts:

- curator-spec main `b4f29cd` (`~/Developer/ReluxWorks/curator-spec`):
  `decisions/0010-agent-environment-profiles.md` Decision 6 (launcher boundary,
  four planes) and Decision 10; `decisions/0012-context-packages-and-semver-locks.md`
  Context (the paragraph naming the M1 resolution and "the next free number after
  reconciliation with the swift-driver draft's 0011"), Decision 6 (MCP fragment
  section, `env_names` union, "Under Option A the launcher, `curator-run`, is the
  single composer"), Decision 8 (fragment `profile.lock_sha256`, `precedence`,
  `mcp`); `protocol/environments.md` revision 1 §10 (`env resolve`,
  `launch-env-fragment-v1`, profile-influence boundary), §7.3 (channels), §11
  (umbrella discovery); `protocol/registry.md` §1 (CCJ-1 canonical JSON).
- The number reconciliation fact: `decisions/0011-swift-driver-pair.md` exists only
  on the unlanded branch `draft/TASK-260728-1yhuqi-swift-driver` (head `604d525`,
  checkout `~/Developer/ReluxWorks/.worktrees/curator-spec-draft-swift`); 0012
  landed on main skipping 0011. Verify with `git ls-remote --heads origin` and
  `ls decisions/` in both checkouts.
- Pre-implementation review v3: board resource `pre-implementation-review-v3.md`
  on STORY-260901-zddtn8 (`~/Developer/ReluxWorks/curator/.task-board/.resources/STORY-260901-zddtn8/`).
  Items M1, M2, M7, M8, M15 (ax refuse-on-drift), M16 (CCJ-1 digest) are this
  decision's scope. Quote their verified facts, do not re-derive them.
- curator-agent-launcher main `6de42d8` (`~/Developer/ReluxWorks/curator-agent-launcher/SPEC.md`,
  `0.1.2-draft`): §1 non-goals, §3 CLI, §4.1–4.5 composition, §7, §9 open items.
- agent-session-manager-spec (ax) main v0.5.0 (`~/Developer/ReluxWorks/agent-session-manager-spec/SPEC.md`):
  §5.1 Session Record and the closed Launch Plan (`argv`, `cwd_workspace_id`,
  `cwd_relative`, `env_names`, `env_literals`, `contains_secrets`, `extensions`
  with their limits: argv 1..128 elements, each 1–4,096 bytes, 65,536 total;
  env_names 0..64; env_literals 0..64 × 4,096 bytes, non-secret); §7.5 `launch`
  operation (`{session_record, workspace_paths, execution_profile, launch_plan,
  terminal}` → `SpawnPlan`) and the `SpawnPlan` row (`argv, cwd, env_names,
  env_literals, native_session_id, profile_mapping, extensions`); §13.1
  direct-session launch (plan built at creation, step 4 "Call provider `launch`
  and validate its argv/env-name plan"); §13.10 resume; §14.1 command surface
  (`ax start NAME --provider ID [--profile standard|yolo] [--workspace PATH]`).
  And PR #1 (`draft/curator-environment-integration`, head `d7075e1`, open):
  `git diff main...origin/draft/curator-environment-integration -- SPEC.md` —
  three extension keys `works.relux.curator.profile-name|profile-pin|fragment-digest`,
  the §7.5 paragraph "the launcher merges the env variables … into `env_literals`",
  the §13.10 drift paragraph, the §14 `curator session` informative note.
- skill-agents-management main `944c7b4` (`~/Developer/ReluxWorks/skill-agents-management`):
  `pkg/agentic/system.go` (`LaunchMode` = Exec | DryRun | ManagedSession;
  `LaunchRequest` with `Home` "load-bearing: on-disk limit state is keyed by
  (provider, home)"; `EffortTransport` None | Argv | Stdin; `StdinPayload`),
  `pkg/agentic/plan.go` (`Plan`, `BuildPlan(registry, req, mode)`),
  `pkg/agentic/systems/claude/args.go` (`-p --output-format json --model …
  [--effort …] --dangerously-skip-permissions` for every mode),
  `pkg/agentic/systems/codex/args.go` (`exec -m`, `-c model_reasoning_effort=…`),
  `pkg/agentic/systems/pi/` (stdin effort transport), `docs/architecture.md`
  invariants (plans as values; no injected default; vendor owns vocabulary).

## Settled operator decisions (record; do not reopen or weigh alternatives as open)

Execution ownership is **Option A** of review M1: `curator-run` is the single
composer in both tracked and untracked modes. `ax start` gains
`--launch-plan FILE|-` accepting a closed caller-supplied plan
`{argv | argv_suffix, env_literals, extensions}`; ax validates it under its §5.1
rules and embeds it verbatim into the immutable Session Record; provider plugins
translate it without rebuilding argv. `SpawnPlan` gains an optional `stdin`
member (Option A's stdin question is decided: grow the plan, do not refuse).
agents-management gains `LaunchModeInteractive`. The launcher SPEC goes to 0.2.
The ax change is delivered as a revision of PR #1 and is never merged by us.

## Contract items — every one must be specified normatively

### 1. Numbering and status
- Title `Decision 0013: execution ownership and launch plans`. Status: Proposed
  2026-09-05, draft for review. State the reconciliation: 0011 is reserved by the
  swift-driver draft on its branch (cite branch and head), 0012 is landed, 0013 is
  the next free number; this decision is the one 0012's Context defers to.
- Acceptance authorizes: launcher SPEC 0.2, the agents-management
  `LaunchModeInteractive` change, the revision of ax PR #1, and the
  `curator run` rows of the environments.md revision 1.1 batch — each separately
  tracked. Changes nothing in frozen `protocol/core.md`; changes no environments.md
  text itself (the 1.1 batch does), but fixes what that batch must say about
  `curator run` (Decision 6 below).

### 2. Ownership model (Option A) — the four planes with one composer
- Roles, closed: Curator resolves (fragment); agents-management builds an
  interactive plan value; `curator-run` composes exactly one launch plan from
  (interactive plan + fragment channels + native args after `--`); ax validates,
  records immutably, and launches through its plugin; the plugin translates a
  caller-supplied plan into its `SpawnPlan` without rebuilding argv.
- Why A over B (ax owning spawn): record the review's reasoning verbatim in
  Rejected alternatives — plugin-owned resume argv preserved, ax stays ignorant
  of Curator and agents-management, one component composes in both modes.

### 3. `ax start --launch-plan FILE|-` (the ax operation; to be carried by PR #1)
- Surface: `ax start NAME --provider ID --launch-plan FILE|- [--profile standard|yolo] [--workspace PATH]`;
  `-` reads stdin (the plan document, not the child's stdin). Mutually exclusive
  with `--task-board` in this revision (state it; task-board launches keep
  building their plan inside ax).
- Document shape, closed, JSON, rejected on unknown members:
  `schema` (fixed string — choose `ax-launch-plan-request-v1` or justify another),
  exactly one of `argv` (complete argv, element 0 the provider executable as the
  plugin would resolve it — or specify that element 0 is the plugin's own
  executable spelling and the caller supplies argv[1..]: decide and say why) or
  `argv_suffix` (appended after the plugin's own base argv), `env_literals`
  (map, same grammar and limits as §5.1), `stdin` (see item 4; absent = none),
  `extensions` (reverse-DNS keys, §1.6 rules). `env_names` is NOT caller-supplied
  from the fragment side in this revision except through the composer's own
  allowlist addition: specify how the fragment's `mcp.env_names` union reaches the
  Session Record — as `env_names` entries the composer adds, so the child resolves
  values destination-locally (0012 Decision 6 says the composer adds them to the
  plan's environment-name allowlist). Therefore the document carries `env_names`
  too; say so and keep the §5.1 disjointness rule.
- Validation: every §5.1 limit and secret rule applies to the caller-supplied
  members; a violation refuses `ax start` before any Session Record exists with a
  typed error (name it, e.g. `launch_plan_invalid`, with a `field` member). The
  plan is embedded verbatim into the Session Record's `launch_plan` — argv, or
  base argv + suffix as resolved at creation, must be recorded as the final
  argv so resume replays exactly what launched; specify that the record stores
  the resolved final form plus the caller document under an extension key
  (name it under `works.relux.curator.*`? No — it is ax-generic: use an ax key,
  e.g. `dev.ax.launch-plan-request`; decide, justify).
- Plugin contract: `launch` receives the record's plan as today; a plugin that
  cannot honor a caller-supplied argv (its manifest declares no such capability)
  refuses with a typed capability error before process creation. Add a
  capability name to the §7.3 manifest capability set (propose
  `caller_launch_plan`); say how `resume` uses the recorded final argv (plugin
  resume argv stays plugin-owned; the caller's `argv_suffix` and env are replayed
  from the record).
- Execution profile interaction: `--profile yolo` remains ax's; the caller plan
  MUST NOT carry a permission-bypass flag itself — state that the composer never
  emits one and ax MAY refuse a known bypass spelling (list none normatively;
  leave the list to ax).

### 4. `SpawnPlan.stdin` and the Launch Plan `stdin`
- Add optional `stdin` to both the §5.1 Launch Plan and the §7.5 `SpawnPlan`:
  `{encoding: "utf-8"|"base64", bytes: string}` (or a single string with a
  declared encoding — decide), bounded (propose 65,536 bytes; justify against
  argv's 65,536 total), non-secret under the same rule as `env_literals`, recorded
  immutably, replayed on resume only when the plugin declares it (state the
  default: a resume does not replay stdin). Absent means the child's stdin is
  the terminal. Say which agents-management transports need it (pi's stdin
  effort transport; interactive mode declares empty stdin per M2 — so in
  revision 1 the composer sends `stdin` only when the interactive plan's
  `StdinPayload.Attached` is true; specify the mapping).

### 5. `LaunchModeInteractive` (agents-management)
- A fourth `LaunchMode`: per-system argv containing only model selection and
  effort transport; no print/headless mode, no output-format flag, no permission
  bypass, no goal or assignment-prompt machinery, no budget flag; `StdinPayload`
  empty unless the system's effort transport is stdin (then exactly the effort
  encoding). Composition prefix (MCP) is NOT part of the interactive argv (the
  fragment's MCP channel is the composer's; say so). `Home` and `WorkDir` carry
  as today. A system that does not declare the mode → `ErrUnsupportedLaunchMode`.
- Per-system argv is the system plugin's to spell (single construction site
  invariant); this decision fixes the grammar constraints, not the spellings,
  and requires argv-parity goldens per system for the new mode. Model is
  required as today; effort follows the vendor's `EffortSupport`.
- Name the consumer: launcher SPEC 0.2 §4.1 requests this mode by name; the
  launcher never spells provider flags (M2's "reject launcher-owned interactive
  argv").

### 6. Launcher SPEC 0.2 (curator-agent-launcher) — what 0013 fixes
- Ordering: fragment first (§4.3 before §4.1); the fragment's managed-home path
  for the environment becomes `LaunchRequest.Home` (M7); `WorkDir` = cwd.
- Default model/effort ownership (M8): the launcher owns default resolution:
  a closed launcher machine-config mapping env-id → {model, effort}; fallback:
  `vendorplugin.Lineup`'s top admitted pair with the vendor's recommended effort;
  precedence: `--model`/`--effort` flags > machine config > lineup fallback; the
  resolved pair is printed on stderr at launch. Verify the `Lineup` name in the
  agents-management sources before citing it (grep `Lineup`); if the identifier
  differs, cite the real one.
- Tracked mode delegates: with ax configured, `curator-run` composes the closed
  plan document and invokes `ax start NAME --provider ID --launch-plan -`; the
  session name derivation is the launcher's (propose `<env-id>-<profile>-<utc-stamp>`
  or `--name`; decide); the extension keys of PR #1 are set by the composer
  (`profile-name`, `profile-pin` = `lock_sha256` under 0012 — reconcile PR #1's
  "commit or state hash" with 0012's lock identity: profile-pin becomes the lock
  sha256; say so), plus `fragment-digest` over the CCJ-1 canonicalization of the
  fragment (M16). Untracked mode execs the same composed argv/env/stdin directly.
  `ax_handoff_failed` stays terminal (no untracked fallback).
- Composition rule, closed: final argv = interactive plan argv ++ system-prompt
  channel flags (opt-in, §5) ++ MCP channel flags (fragment `mcp` descriptor,
  0012 Decision 6) ++ native args after `--`; env = inherited ⊕ plan env ⊕
  fragment `env` (fragment wins on its own names; unchanged §4.4); env_names
  allowlist = fragment `mcp.env_names`; stdin = interactive plan stdin. Order is
  contract because everything after the goal/last flag is user turn for some
  tools — state the general rule and leave per-tool verification to SPEC 0.2.

### 7. ax PR #1 revision items (this decision names them; the PR carries them)
- §7.5 paragraph rewrite: no "the launcher merges … into `env_literals`" actor
  without an interface — replace with the `--launch-plan` operation.
- §13.10 drift: refuse (not warn) by default when the Session Record's recorded
  chain carries `class: system` modules (M15); keep warn-and-continue otherwise;
  the strict mode stays. Specify how the record knows: a fourth extension key
  `works.relux.curator.system-modules` (boolean) set by the composer — or fold
  into fragment data; decide and justify.
- `fragment-digest` keyed over CCJ-1 canonical bytes (registry.md §1), not the
  pretty-printed `--format json` output (M16).
- The three existing keys remain; `profile-pin` semantics updated to the 0012
  lock identity.

### 8. Compatibility, security, consequences, open questions
- Compatibility: ax minor version bump (v0.6.0 candidate — say the PR proposes,
  the ax maintainer decides); agents-management minor (new mode, additive);
  launcher 0.2.0-draft; environments.md unaffected; core frozen.
- Security: the profile-influence boundary (environments §10.3) extends to the
  plan document — profile bytes never reach argv except through registry-declared
  channel descriptors with manager-owned paths; the composer is the only writer
  of `argv_suffix`; no permission bypass in interactive mode; secrets excluded by
  §5.1; stdin bounded and non-secret.
- Open questions: session-name derivation; plugin-side handling of a caller argv
  colliding with plugin base flags; whether task-board launches adopt
  `--launch-plan` later; stdin replay on resume per provider.

## Constraints
- No edits outside the one new file. No pushes. One signed commit.
- Verify before asserting: every §, every identifier (`Lineup`, `StdinPayload`,
  `ErrUnsupportedLaunchMode`, capability set spelling in ax §7.3), every limit.
  Label anything not verified as "docs-confidence" in the report, not in the decision.
- Board: `TASK_BOARD_DIR` is set for you; use `task-board m 'add_resource(TASK-260905-2ft7ts, …type=outcome…)'`
  then `task-board handoff TASK-260905-2ft7ts --role developer`. Never write
  LOGBOOK.md or anything into the control root.
