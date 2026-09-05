# Drafting report: `protocol/environments.md` revision 1.1 (TASK-260905-jb6rvg)

## Commit

Signed commit on the story worktree branch `task-board/story/STORY-260905-1xwg3d`
(base curator-spec main `ec695ba`), fast-forwarded onto `draft/environments-revision-1-1`
in `/Users/iv/Developer/ReluxWorks/.worktrees/curator-spec-env-1-1` (both at the same head):

```text
commit 4492b7e0b49b73383d43dd1bde702a7ca578681f
Good "git" signature for oparin@me.com with ECDSA key SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM
Author: Ivan Oparin <oparin@me.com>
    Rewrite environments.md revision 1.1 on the Decision 0012 model
 protocol/environments.md | 1818 +++++++++++++++++++++++++++++++++++-----------
```

One file changed. Schemas, vectors, manager §12, cli rows untouched (next batches). Not
pushed, no tag, no PR. The spawn assignment placed the run in the story worktree, the brief
in the draft worktree; the same commit object is on both branches.

## `make validate` tail (exit 0, venv `.temp/venv`, log `.temp/make-validate-02.log`)

```text
python3 tools/validate.py
validated 57 schemas and 780 vector files
python3 -B -m unittest discover -s tools -p 'test_*.py'
Ran 152 tests in 23.896s
OK
go test ./tools/...
ok  	github.com/relux-works/curator-spec/tools/generate-vectors	0.521s
```

(A first run, `.temp/make-validate-01.log`, exited 2 because my scratch copy of the old
text under `.temp/` was picked up by the repo-wide link checker; the scratch file was moved
out of the tree and the run repeated.)

## Byte-identity check

Sections compared heading-by-heading against `ec695ba`:

- **identical**: §1.2 (M3 byte-exactness), §5.2, §7 body, §7.2, §8 heading, §8.4, §9 heading, §10 heading.
- **changed although the 0012 table says *unchanged***, each because a binding review item
  anchors there (listed so the reviewer can judge each): §7.4 (M4, erratum item 3), §7.5 and
  §7.7 (M13 — the "(warning)" marker dropped, `shadow_acknowledged`), §7.6 (N5), §8.3 (M11
  versioned backups), §8.5 (only the `environment_backup_exists` row reworded for
  generations), §9.3 (N11), §9.5 (N8, N9), §10.4 (M10 diagnostics), §11 and §11.1 (M6
  trusted-location rule for umbrella providers), §7.1 (M4e/M14 XDG allowlist and
  reconciliation, N2 split-brain note), §5.7 (M5 and the MCP `PATH` warning need a table).

## Decision 0012 impact table → applied

| Section | Disposition | Where / what |
|---|---|---|
| §1 | rewritten | profile = root + lock + overlays; `git` carries `range`/`tag`/`revision`, `directory`; no `branch`; `path` root is `agent-context.json`; `local` unchanged |
| §1.1 | bytes change | ref-form row reworded; `context_range_conflict`, `context_version_mismatch`, `context_weights_duplicate`, `context_weight_unknown` added (the two remaining weight diagnostics sit in §6.1 per the §6 row) |
| §1.3 (new) | — | the lock (0012 D3) — placed as a subsection of §1 so the §1 rewrite can reference it |
| §1.4 (new) | — | versions, ranges, resolution (0012 D2), restated verbatim in semantics |
| §2 | rewritten | `agent-context.json`, `context/`, `CONTEXT.md`; `Profilefile.json` withdrawn |
| §2.1 | rewritten | `profile_index_invalid`/`profile_root_invalid` withdrawn; `context_manifest_invalid`; `mcp_declaration_invalid`, `mcp_package_not_allowed` |
| §2.2 (new) | — | MCP declaration package (0012 D6 package half) |
| §3 | rewritten | inline `context.modules`; entry shape, byte rules, applicability verbatim; `version: 1` gone; title renamed |
| §3.1 | bytes change | `context_manifest_invalid` |
| §4 | rewritten | per-package entries; profile identity = set the lock names |
| §5 body | rewritten | tuple (lock, precedence policy, env, form); emitted order; one chapter per member with applicable modules; `## Context: <name> <version>`; collision and joining rules verbatim; M5 size advisory paragraph |
| §5.1 | rewritten | `curator-root-context-v2` grammar of 0012 D8 |
| §5.2 | unchanged | byte-identical |
| §5.3 | bytes change | `<package-name>`; opencode rule stands; N14 docs-confidence sentence; N1 consequence paragraph |
| §5.4 | bytes change | header alone |
| §5.5 | bytes change | "every context member in emitted order"; `mcp/` sibling named |
| §5.6 | bytes change | tuple follows §5; MCP file is its own surface |
| §5.7 | unchanged → two rows added | `environment_context_size_exceeded` (M5), `mcp_command_unresolved` (0012 D6 warning needs a table) |
| §5.8 (new) | new | per-adapter MCP file, bytes, managed-home only, codex fixed location, trailing LF |
| §6 | rewritten | overlays as closure members; effective weight; two primitives; joint resolution; N10/N4 scope paragraph |
| §6.1 | rewritten | `environment_composition_invalid`, `context_weight_conflict`, `context_weights_not_root`; skill-divergence withdrawn |
| §7 body, §7.2 | unchanged | byte-identical |
| §7.1 | unchanged → changed | XDG paragraph rewritten (M4e, M14), opencode split-brain note (N2); adapter table itself identical |
| §7.3 | bytes change | `argument`, `with`; `semantics` system-prompt-only; M5 pi row; admission rule |
| §7.4 | unchanged → rewritten | M4 (a)(b)(c), N2 matrix, erratum item 3 |
| §7.5 | unchanged → changed | M13 |
| §7.6 | unchanged → changed | N5 |
| §7.7 | unchanged → changed | M13 row; new rows for M4/M14/N3/N5 |
| §7.8 (new) | new | the four MCP channel rows with evidence column |
| §7.9 (new) | — | N3 recorded versions, M5 per-adapter advisory, erratum fast path |
| §8.1 | bytes change | "the same lock's store entries"; M4(d) fresh homes; N6 two doors |
| §8.2 | rewritten | `profile` = name/root/kind/`lock_sha256`; `members`; `precedence` object; `mcp` surface; passthrough/seeds records |
| §8.3 | unchanged → changed | M11 versioned backups, retention, scrub, status visibility |
| §8.4 | unchanged | byte-identical |
| §8.5 | unchanged → one row reworded | backup generation wording |
| §9.1 | rewritten | one root; `--range|--tag|--revision`, `--directory`, `--as`; resolution+lock+audit of every member; detector scope re-targeted; `--use` bare; M9; N7 hint |
| §9.2 | bytes change → extended | "from the store entries its lock names"; M11 `use` failure, `update`, `remove`, `env unmanage` |
| §9.3 | unchanged → changed | N11 |
| §9.4 | rewritten | lock's skills; direct declarations into the lock; `default` local root 0.0.0; M6 commands; N4 hybrid |
| §9.5 | unchanged → changed | N8 trigger set; N9 heuristic and suspected external writer |
| §9.6 | rewritten | `agent-context.json` version `1.0.0`; `requires.skills` pinned by `revision` |
| §9.7 | bytes change | `profile_index_ambiguous` withdrawn; ref row reworded; new rows for M9/M11/M6/N9 |
| §10.1 | bytes change → extended | tuple; M10 read-only resolve, `--repair`, locks, residual; M16 CCJ-1 output; N7 shell; N12 `curator run`/opencode; 0013 D6.3/6.4 sentences |
| §10.2 | rewritten | `lock_sha256`; no `composition`; `precedence` object; `mcp` section; `argument`/`with`; M6 reserved `path_prepend` |
| §10.3 | bytes change | `env_names` paragraph |
| §10.4 | unchanged → two rows added | `environment_home_stale`, `environment_lock_unavailable` (M10) |
| §11, §11.1 | unchanged → changed | M6 `subcommand_provider_untrusted`; 0013 named |
| §12 | bytes change → extended | requirement/lock hash/GC roots; every status row named by M4/M13/M14/N2/N3/N5/N6/M11 |
| §12.1 (new) | — | M12 closed knob list, bootstrap shape |
| §12.2 (new) | — | M15 lockable keys and phasing sentence |
| §13 | rewritten | full surface list for the next batch |

## Review items → where and what

| Item | Section(s) | What the text now says |
|---|---|---|
| M4a | §7.4 | provisioning seeds per adapter table (one-time, never hashed, marker-recorded); `.claude.json`/`config.toml` shapes docs-confidence |
| M4b | §7.4, §7.7, §12 | strategy column (ambient / `directory` / `keyring-preferred` / `file-link`), write behavior recorded verified/unverified, `environment_passthrough_detached` liveness row, `--repair` re-links |
| M4c | §7.4, §7.7 | `isolated` unsupported for `claude_code`/macOS and `opencode` → `environment_isolated_unsupported`; lifted only on positive evidence |
| M4d | §8.1 | normative fresh-home paragraph (provisioning order, one transaction) and first-resolve notice |
| M4e | §7.1, §12.1 | `xdg_seed_allowlist` (default git, gh, ssh); `XDG_DATA_HOME`/`XDG_STATE_HOME` ambient |
| M5 | §7.3, §7.9, §5, §5.7 | pi row from evidence (append via `--append-system-prompt` path, polymorphic, launcher readable-check; no flag replace; `SYSTEM.md` only); claude both `-file` flags verified 2.1.261; `argument` admission rule; `root_context_size_advisory_bytes` per adapter, `environment_context_size_exceeded`; codex global cap docs-confidence |
| M6 | §9.4, §10.2, §11, §11.1, §9.7 | decided: no commands in managed homes in rev 1; `path_prepend` reserved never emitted; shims follow machine-current and re-point on `linked` switch; `curator-*` forbidden in skill/managed bin dirs (`environment_reserved_command_name`); providers refused from manager-published dirs (`subcommand_provider_untrusted`) |
| M9 | §9.1, §9.7, §12.1, §13 | detector unpinnable; scoped waiver `{pin,file,span,reason}` in `secret_material_waivers`; `context_secret_waiver_unmatched`; detector classes with named positive/negative vectors in `vectors/context-detectors.json`; `context-system-module-present` always-warn |
| M10 | §10.1, §10.4 | resolve read-only, lock-free verification over marker surfaces, link-target identity suffices; `environment_home_stale` fail-closed without `--repair`; repair under the mutation lock with bounded wait, `environment_lock_unavailable`; residual recorded; launcher MAY re-check; lock classes named (mutation lock, status read; no per-home lock) |
| M11 | §9.2, §8.3, §9.7 | `profile update` six-step order; re-install = update; skill-scope update/upgrade never move the lock; `default` re-keyed once per operation; `profile remove` refusal (`profile_in_use`), homes retained unless `--purge`, orphans reported; `profile use` attempts every entry, `profile_use_partial`, current recorded only on full scope; backups `<n>/`, retention, scrub, status visibility; `env unmanage [--restore-backups]` |
| M12 | §12.1 | closed knob table (20 knobs) for `manager-config` schema 2; per-machine bootstrap shape |
| M13 | §7.5, §7.7, §12, §12.1 | shadow-inert non-current by default; `shadow_acknowledged` per `{env, path}` downgrades that row |
| M14 | §7.1, §7.7 | reconciliation on `sync`/`use`/`--repair`; new seeds recorded; `environment_seed_shadowed`; inside the allowlist |
| M15 | §12.2 | decided: lockable `overlays_allowed`, `precedence`, `mcp_package_allowlist`, `passable_env_names`, `require_current_profile`, `isolation` (shared-direction only); non-overridable skill class declared out of rev 1 with rationale |
| M16 | §10.1 | `--format json` is CCJ-1 + LF; digest over the CCJ-1 bytes (0013 D6.4) |
| N1 | §5.3 | decided: keep opencode `referenced`; consequence stated (no other `opencode.json` config; MCP unaffected via `OPENCODE_CONFIG`; monolithic is the supported shape otherwise) |
| N2 | §7.4, §7.1, §12 | isolation matrix table; standing `env status` opencode split-brain note |
| N3 | §7.9, §7.7, §12 | recorded verified release per adapter; detected release in status; `environment_tool_version_unverified`; erratum fast path |
| N4 | §9.4, §6 | project > hybrid > current profile of the scope; hybrid never composes, never re-renders on switch |
| N5 | §7.6, §7.7, §12, §12.1 | one-time consent (`environment_target_consent_required`, `targets.<id>.consented`); status names ungoverned MCP/commands; agents-read-the-files docs-confidence |
| N6 | §8.1, §12 | decided: loud documented split (two doors), no native-home routing, no `--isolated-home` in rev 1 |
| N7 | §9.1, §10.1 | CRLF `path` install gets a fix hint (no normalize flag); `--format shell` POSIX-only, Windows uses json, no pwsh |
| N8 | §9.5 | only mutating operations trigger onboarding; read-only commands never write or prompt |
| N9 | §9.5, §9.7 | dotfile-manager heuristic (`environment_foreign_manager_suspected`); repeated-drift suspected external writer |
| N10 | §6 | decided: `agents`/`locale` scoped out with reason (machine preference applies to the closure); hybrid never composes |
| N11 | §9.3 | `--clear` or naming the machine default removes the scope record |
| N12 | §10.1 | `opencode` is `env_unsupported` for `curator run` (launcher diagnostic); `env resolve opencode` fully specified |
| N13 | §7.3, §7.4 | erratum items 1 and 3 carried; item 2 (path in rev 1) is the document's standing state |
| N14 | §5.3 | claude referenced approval docs-confidence behind the pinned-release gate |

Decision 0013: D4 (stdin is the launcher's; no environments.md surface), D6.3 and D6.4 are
reflected in §10.1 `curator run`, §10.3 `env_names` bounding, §1.3 `profile-pin` = lock hash,
§10.1 fragment-digest base. Decision 0010 erratum: items 1 and 3 corrected in §7.3/§7.4.

## Decisions taken where the brief said "decide and state"

- **M6**: no commands in managed homes in rev 1, `path_prepend` reserved. Rationale: the 0013
  launcher composes environment from fragment `env` only; emitting a member it would ignore
  silently is worse than a stated limitation, and the schema reservation cannot be retrofitted.
- **M10 stale semantics**: no fragment without `--repair` (fail-closed). Rationale: a launcher
  running an agent in a home the manager knows is wrong, silently, is the class of failure the
  protocol bans; `curator run` passes `--repair`.
- **M15**: six lockable keys; non-overridable skill class out of rev 1 because joint resolution
  already makes a disagreeing overlay a hard conflict.
- **N1**: keep opencode referenced, state consequence; MCP reaches opencode via `OPENCODE_CONFIG`
  so the lockout is providers/theme/keybinds only.
- **N6**: loud split, not native-home routing — §10.3 requires fragment values below the
  environments root and §5.5/§5.8 surfaces live only in managed homes.
- **N7**: fix hint only, no normalizing install flag (§3 has no normalization path); no pwsh.
- **N10**: `agents`/`locale` scoped out; a per-package member returns under its own review.
- **Diagnostic placement**: weight diagnostics split §1.1 (`_duplicate`, `_unknown`) / §6.1
  (`_conflict`, `_not_root`) so both 0012 rows are satisfied literally.
- **Backups**: `environment_backup_exists` now fires only on an existing next generation, which
  removes the F1 dead end without a new diagnostic.
- **Default XDG allowlist** `git`, `gh`, `ssh`: the three entries the lens reports named as
  the operator tools an agent session actually invokes.
- **Repair lock wait bound**: "at least one second, at most sixty", implementation-documented.
- **Seed-unreadable**: `environment_seed_unreadable` stops provisioning (absence-vs-unreadable).

## Facts labeled docs-confidence in the text

- Claude referenced-form external-include approval exemption (§5.3).
- Per-adapter MCP server-object shapes for `claude_code` and `opencode`; codex layer-file grammar and
  that an `mcp_servers`-only layer composes cleanly (§5.8, §7.8).
- codex `-c model_instructions_file=<path>` per-invocation application (§7.3).
- claude Linux `.credentials.json` write behavior; codex `cli_auth_credentials_store` modes and
  `auth.json` write behavior; pi `auth.json` write behavior (§7.4).
- `.claude.json` seed authenticating a fresh home; codex `config.toml` member shapes; pi
  `settings.json`/`models.json` shapes (§7.4).
- opencode configuration merge order and every opencode fact (not installed here) (§7.8, §7.9).
- Xcode embedded agents reading root context and skills at the declared homes (§7.6).
- codex `project_doc_max_bytes` applying to the global `AGENTS.md`; every other adapter's
  size advisory adopts 32768 as a default with no published cap (§7.9).
- `oauth.claude.profile.<64-hex>` Keychain account (§7.4, explicitly unrelied-upon).

## Verified on this machine (2026-09-05)

- Claude Code **2.1.261**: `--system-prompt-file`, `--append-system-prompt-file`, `--mcp-config`,
  `--strict-mcp-config` present in `claude --help`.
- codex-cli **0.153.2** (brief said 0.151.0; the installed binary moved): `-p, --profile` "Layer
  $CODEX_HOME/<name>.config.toml on top of the base user config"; `-c, --config <key=value>`.
- pi **0.84.2**: `--system-prompt <text>`, `--append-system-prompt <text>` ("Append text or file
  contents"); no `-file` variants; no MCP flag.
- opencode: not installed (`command not found`).

## Notes for the reviewer

- The rewrite keeps every "held under attack" rule of review §5 (boundary, closed registry,
  no templating, byte-exact determinism, always-strict audit, inert system prompt,
  fire-vs-manage, two modes, onboarding import).
- Duplicate table rows that pre-existed and were kept: `context_manifest_invalid` (§2.1/§3.1),
  `environment_import_lossy` (two §9.7 rows), `environment_surface_unmanaged_conflict`
  (§5.7/§8.5), `environment_unknown` (§7.7/§10.4), `profile_unknown` (§1.1/§10.4).
- Names introduced for the next batches: `vectors/context-detectors.json` and its case names;
  knob names of §12.1; `subcommand_provider_untrusted`, `environment_reserved_command_name`.
