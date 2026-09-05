# Producer brief: `protocol/environments.md` revision 1.1 on the Decision 0012 model

## Where and what

- Worktree `/Users/iv/Developer/ReluxWorks/.worktrees/curator-spec-env-1-1`, branch
  `draft/environments-revision-1-1`, base = curator-spec main `ec695ba`.
- This batch edits exactly one file: `protocol/environments.md`, rewritten **in place**. It
  stays "revision 1" in its own header and identifiers; add a `1.1` note to its revision line
  or history block only if the document already carries one (check; do not invent a new
  versioning surface). Schemas, vectors, manager §12, and cli rows are the next batches —
  do not touch them; where the text names a schema or vector that the next batch will create
  or regenerate, name it exactly as Decision 0012's impact table names it.
- Deliverable: one signed commit (`git commit -S`; paste the `git log --show-signature -1`
  line). `make validate` must still pass (use a venv: `python3 -m venv .temp/venv &&
  .temp/venv/bin/pip install -r requirements-dev.txt`, then `PATH=$PWD/.temp/venv/bin:$PATH
  make validate`) — the link checker and structure checks cover this file. Do not push, tag, or
  open a PR. Attach `TASK-TASK-260905-jb6rvg_drafting-report.md` as an outcome resource (see Report).
  Then `task-board handoff TASK-260905-jb6rvg --role developer`. Never write LOGBOOK.md or anything into
  the control root; the story worktree task-board provisions stays untouched.

## Sources (read all before writing; never restate divergently)

curator-spec main `ec695ba`:
- `protocol/environments.md` — the landed revision-1 text you are rewriting; keep every rule
  the impact table marks *unchanged* byte-for-byte, and keep the document's voice.
- `decisions/0012-context-packages-and-semver-locks.md` — Decisions 1–9 are the model; the
  **Compatibility impact table** is the section-by-section contract for this rewrite (each row's
  disposition is binding: *rewritten*, *bytes change*, *unchanged*, *new*).
- `decisions/0013-execution-ownership-and-launch-plans.md` — Decision 6.3 (composition rule:
  `env_names`, the literal-wins collision rule), 6.4 (extension keys, `profile-pin` =
  `lock_sha256`), Decision 4 (stdin), Consequences; the `curator run` sentences in §10/§11
  must agree with it.
- `decisions/0010-agent-environment-profiles.md` with its Erratum (2026-09-05): every rule the
  erratum corrects must appear corrected here (pi channel row; `isolated` unsupported for
  claude_code/macOS and opencode).
- `protocol/core.md` §2 (identifiers), §4.4, §6.1–6.3, §7 (closure), §8 (content hash), §10;
  `protocol/registry.md` §1 (CCJ-1); `profiles/manager.md` §1 (system configuration, `locked`),
  §3.1 (reserved names), §6, §12 (read for consistency; do not edit).
- Board resource `pre-implementation-review-v3.md` on STORY-260901-zddtn8
  (`~/Developer/ReluxWorks/curator/.task-board/.resources/STORY-260901-zddtn8/`): items
  M4, M5, M6, M9, M10, M11, M12, M13, M14, M15, M16 and N1–N14 are binding on this text
  (M1/M2/M7/M8 landed as Decision 0013; M3 landed as §1.2). Lens evidence for verified facts:
  `~/Developer/ReluxWorks/curator/.task-board/.resources/TASK-260902-2142et/*.md` (operator
  journeys, F1–F22), `TASK-260902-13dvty/*.md` (feasibility), `TASK-260901-2eu2qg/*.md` if
  present (threat model).
- Installed binaries for any spelling you state as verified: `claude --help` (2.1.26x),
  `codex --help` (0.151.0), `pi --help` (0.84.2); record versions. Anything you cannot verify
  on this machine is labeled **docs-confidence** in the text's existing style (the §7.3
  discipline) — never asserted as verified.

## Settled operator decisions (record; do not reopen)

Semver only where a lock exists; core frozen (project lock is a separate open question).
Precedence = `winner` × `placement`. Weights: manifest → direct-requirer edge (must agree) →
root `weights` map. Overlays from `git` or `path`. MCP ships in this revision as a launch
channel into managed homes only; allowlist over MCP package canonical source identities.
Strict SemVer 2.0 `v`-tags without build metadata; npm-shaped ranges (node-semver semantics as
0012 Decision 2 fixes them) plus `latest`. environments.md stays revision 1; only the header
type line bumps to `curator-root-context-v2`. Execution ownership is Decision 0013 Option A.
csk cleanup is surface naming only.

## The rewrite (section by section = the 0012 impact table; plus the review items placed where they belong)

Apply the 0012 impact table row by row. Then fold in the review items at these anchors
(each is a MUST; the resolution text in the review is the requirement — specify it, do not
paraphrase it weaker):

- **§1 / §1.1**: 0012 row. Keep the M3 `§1.2` byte-exactness subsection intact.
- **§2, §2.1, §3, §3.1, §4**: 0012 rows (`agent-context.json`, `context/`, `CONTEXT.md`,
  inline manifest, diagnostics renames, per-package store entries).
- **§5, §5.1–5.8**: 0012 rows; §5.8 new (MCP launch-channel output per adapter, managed homes
  only, `codex_cli` fixed location `<home>/curator-mcp.config.toml`, CCJ-1 bytes where JSON,
  the trailing-LF rule as §5.3's opencode file has it). **M5** size advisory:
  `root_context_size_advisory_bytes` per adapter with `environment_context_size_exceeded`
  (warn) — put the rule in §5.6/§5 and the per-adapter value in §7's registry; codex global
  `AGENTS.md` cap 32,768 is **unverified for the global doc** — state it as docs-confidence
  pending the verification sprint.
- **§6, §6.1**: 0012 rows (joint resolution, weights, two primitives, `context_weight_conflict`,
  `context_weights_not_root`, withdrawal of `environment_composition_skill_divergence`).
  **N10**: composition covers `agents`, `locale`, and hybrid targets — specify or explicitly
  scope out with a reason. **N4**: hybrid scope reconciled — project > hybrid > current profile
  of the scope; hybrid never composes.
- **§7, §7.3, §7.8**: 0012 rows (descriptor `argument` + `with`, `semantics` system-prompt-only;
  §7.8 the four MCP rows). **M5** pi row rewritten from evidence: append via
  `--append-system-prompt` with a path the launcher verifies readable immediately before exec,
  `argument: contents`-class polymorphism recorded; no flag replace for pi (`SYSTEM.md` only);
  claude rows keep both `-file` flags (verified 2.1.26x); add the admission rule requiring
  verified file-path acceptance for `argument: path` descriptors. **M4** per adapter:
  (a) **provisioning seeds** class — non-credential, one-time, never hashed, enumerated per
  adapter (`.claude.json` shape and codex `config.toml` trust/model entries are docs-confidence
  until the sprint — say so); (b) **passthrough strategy** per adapter with the pinned
  release's write behavior recorded as verified/unverified (in-place vs rename-over), preferring
  keyring-backed modes (codex `cli_auth_credentials_store = file|keyring|auto`) and
  directory-level passthrough for rename-over tools, plus the status liveness row
  `environment_passthrough_detached`; (c) `isolated` **unsupported** in this revision for
  `claude_code` on macOS and for `opencode` (`environment_isolated_unsupported`), lifted only
  on positive evidence; (d) a normative fresh-home paragraph and first-resolve notice; (e) XDG
  seeding narrowed to an allowlist with `XDG_DATA`/`XDG_STATE` stated ambient. **N2**: a
  per-adapter isolation matrix (normative table) and the standing `env status` note for
  opencode's split-brain skills. **N3**: recorded verified tool version per adapter; best-effort
  detected version in status; erratum fast path. **N12**: `opencode` row marked
  `env_unsupported` for `curator run` until an agents-management plugin exists (Decision 0013
  D6.4 table). **N14**: claude referenced-form approval stays behind the pinned-release gate;
  extend the §7.3 wording to §5.3.
- **§8.1, §8.2**: 0012 rows (marker `profile` = root, kind, lock hash; `members`, `precedence`,
  `mcp` surface). **M13**: `--check` exit contract resolved to non-current by default for
  shadow-inert, plus a per-path machine-config acknowledgment that downgrades exactly that row
  (`shadow_acknowledged` — name it in §8/§12 and the manager knob for the next batch).
  **M14**: XDG seed reconciliation on `sync`/`use`/repair; record new seeds;
  `environment_seed_shadowed` for unrecorded entries; inside the M4 allowlist.
- **§9.1**: 0012 row (one root per install, `--range|--tag|--revision`, resolution + lock +
  audit of every member, detector scope, `--use` takes no name). **M9**: the
  `context-secret-material` detector is **unpinnable**; scoped false-positive waiver (file +
  span + reason, recorded); detector classes with positive/negative vector names (for the next
  batch); the always-warn surfacing class `context-system-module-present`. **M6**:
  profile-scoped command roots below the environments root; fragment `path_prepend` (one
  manager-owned path); `linked` switching re-points machine shims to the current profile;
  umbrella providers resolved from a trusted location or identity-verified; managed/skill bin
  dirs MUST NOT carry `curator-*` names — if this revision chooses "no commands in managed
  homes", say so and still reserve `path_prepend` in the fragment. **N8**: only mutating
  profile operations trigger onboarding. **N9**: foreign-manager heuristic notice and
  repeated-drift "suspected external writer". **N7**: Windows — `path` installs from autocrlf
  checkouts get a fix hint or normalize flag; `--format shell` POSIX-only with a `pwsh` or
  json-only guidance sentence.
- **§9.2**: 0012 row. **M11**: `profile update` (re-resolve → strict re-audit, old pin stands on a
  blocking finding → new store entries → re-materialize in-place scopes → managed homes stale →
  old entries GC-eligible; re-install of an installed source; skill-scope `update`/`upgrade`
  never move profile pins; `default` local pin must not churn ax drift on every `global add`);
  `profile remove` (refuse while current in any scope or in any overlay; managed homes
  retained unless `--purge`; orphans reported); `profile use` failure semantics (attempt every
  entry, per-adapter results, new current recorded only when the whole scope materialized);
  versioned backups `.agent-environment-backup/<n>/` with retention, scrub, and status
  visibility; `env unmanage [--restore-backups]`. **N11**: scoped current can be cleared
  (`--clear` or naming the default removes the scope record). **N6**: native home for the
  machine-current profile in `linked` mode unless `--isolated-home`, or a loud documented split
  — decide and state.
- **§9.4, §9.6, §9.7**: 0012 rows. **N5**: Xcode `auto` first write needs one-time consent;
  status names the ungoverned surfaces; the "agents read the files" claim is docs-confidence
  pending the sprint.
- **§10.1–10.3**: 0012 rows + Decision 0013. **M10**: resolve is read-only — verification is
  lock-free over exactly the marker's surfaces (link-target identity into an immutable store
  entry suffices for symlinked surfaces); `environment_home_stale` reported without repair
  unless `--repair`; repair takes the mutation lock with bounded wait and a distinct
  lock-acquisition diagnostic; the repair→child read race is a recorded residual; the launcher
  MAY re-check hashes before exec. Lock classes named. **M16**: the fragment's CCJ-1
  canonicalization is the digest base (Decision 0013 D6.4). **M6** `path_prepend` member in
  the fragment. **N1**: opencode `referenced` form makes `opencode.json` instructions-only —
  drop from this revision or state the consequence and defer; decide and state.
- **§12**: 0012 row (lock hash, GC live roots) + the M4/M13/M14/N2/N3 status rows named above.
  **M15**: extend the manager §1 `locked` set (composition policy, require-current-profile,
  non-overridable skill class) — name the knobs here for the next batch — or declare fleet
  enforcement out of this revision in a phasing sentence; decide and state. **M12**: name every
  machine-configuration knob this revision needs (composition, forms, targets, `isolated`,
  seeds allowlist, passable-names allowlist, MCP package allowlist, shadow acknowledgment,
  backup retention, defaults) in one closed list so `manager-config` schema 2 (next batch) can
  carry them; team distribution stays per-machine with a documented bootstrap shape.
- **§13**: 0012 row — the complete conformance-surface list for the next batch (schemas
  `agent-context-v1`, `agent-mcp-v1`, `context-lock-v1`, rewritten marker and fragment
  schemas; vector families: version/range parsing, resolution incl. conflict/downward
  re-selection/prerelease, lock canonicalization and hashing, weight ordering under both
  primitives, MCP bytes per adapter, detector classes, the byte-exact snapshot).

Every diagnostic this rewrite adds, renames, or withdraws must appear in the section's
diagnostics table and nowhere else; keep the absence-vs-unreadable discipline; keep every
"held under attack" rule of review §5 (profile-influence boundary, closed registry, no
templating, byte-exact determinism, always-strict audit, inert system prompt, fire-vs-manage,
two-mode requirement, onboarding import design).

## Report

`TASK-TASK-260905-jb6rvg_drafting-report.md`: commit + signature line; a table with one row per 0012
impact-table row (section → disposition applied → where) and one row per M4–M16 / N1–N14 item
(item → section(s) → what the text now says, one line); every decision you had to take where
the brief said "decide and state" with a one-line rationale; every fact labeled
docs-confidence; `make validate` tail.
