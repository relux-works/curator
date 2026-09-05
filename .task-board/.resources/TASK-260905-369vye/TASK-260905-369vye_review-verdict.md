# TASK-260905-369vye review cycle 1 — manager §12, cli rows, manager-config schema 2

Verdict: **ACCEPT** (no blocking or major finding). Change Request `CR-TASK-260905-369vye-1` revision 1.
repeat-of: none

## Subject and the empty repository delta

- Reviewed work: `9af8af8` vs base `a68559b` (`git diff a68559b..9af8af8`, 54 files, +5757/-200).
- The Change Request's base OID is `9af8af8` and its candidate tree `0095f8ee…` equals `9af8af8^{tree}`, so `repository_delta=empty`.
  This is the right outcome: the leaf's second brief (`producer-brief-manager-publish-cr.md`) was a no-edit run whose sole purpose
  was to verify the orchestrator's squashed commit and publish it as the CR. The substance under review is the commit `9af8af8`
  itself, which is one signed commit past `a68559b`. Nothing further was expected in the tree.
- Commit: author `Ivan Oparin <oparin@me.com>`; `git log --show-signature -1` = Good SSH signature (`%G?` = `U`, no principal
  file at the tmp allowed_signers path — same validity the base commits carry; an environment matter, not a change defect).
  `git show --stat` carries no `LOGBOOK.md`, `.temp/`, or `__pycache__` path.

## Gates (rerun by the reviewer at 9af8af8, venv `.temp/venv`)

```text
Ran 166 tests in 29.697s
OK
go test ./tools/...
ok  	github.com/relux-works/curator-spec/tools/generate-vectors	0.467s
validate exit=0
go run ./tools/generate-vectors -root .
git diff --exit-code -- conformance/v1 release/1.0.0-rc.5.json … release/1.0.0-rc.9.json
regenerate-check exit=0
```

Logs: `.temp/review-369vye/validate-01.log`, `regenerate-check-01.log`.

## Mutants run against schema 2 (scratch, schema restored after each)

| # | Mutation of `manager-config-v2.schema.json` | `tools/validate.py` |
|---|---|---|
| M1 | rename `backup_retention` → `backup_retention_days` | FAIL (valid.json rejects the table knob) |
| M2 | `overlay_default_weight` default 1000 → 100 | FAIL (default cross-check names §12.1) |
| M3 | open `$defs.environments` (`additionalProperties: true`) | FAIL (`invalid-unknown-environments-field`) |
| M4 | widen `precedence.winner` enum with `first` | **PASS — survived** (finding F1) |
| M5 | delete `overlay.oneOf` | FAIL (`invalid-overlay-two-requirement-forms`) |
| M6 | `schema_version` const 2 → any integer | FAIL (`invalid-schema-version-1`) |
| M7 | `backup_retention` minimum 0 → -1 | FAIL (`invalid-backup-retention-negative`) |

Six of seven narrowing/widening mutants are caught; the widening mutant M4 is not (see F1).

## Knob → property (every §12.1 row checked byte for byte against the schema)

| §12.1 knob | Property / grammar | Default | OK |
|---|---|---|---|
| `current_profile` | identifier \| null | `null` | ✓ |
| `scoped_current` | identifier → identifier | `{}` | ✓ |
| `overlays.<profile>` | identifier → array of closed `{source, exactly one of range\|tag\|revision, directory?, weight≥0}` | `{}` | ✓ |
| `overlay_default_weight` | nonNegativeSafeInteger | `1000` | ✓ |
| `overlays_allowed` | boolean | `true` | ✓ |
| `precedence.winner` | enum `higher-weight`,`lower-weight` (closed object) | `higher-weight` | ✓ |
| `precedence.placement` | enum `winner-last`,`winner-first` | `winner-last` | ✓ |
| `forms.<env-id>` | identifier → enum `monolithic`,`referenced` | `{}` | ✓ |
| `system_prompt_files.<profile>.pi` | identifier → closed `{pi: off\|append\|replace}` | `off` | ✓ |
| `targets.<target-id>.participation` | closed `{participation: auto\|off\|enabled}` | `auto` | ✓ |
| `targets.<target-id>.consented` | boolean | `false` | ✓ |
| `isolation.<profile>.<env-id>` | identifier → identifier → enum `shared`,`isolated` | `{}` | ✓ |
| `xdg_seed_allowlist` | unique array of entry names (no separators, not `.`/`..`/`opencode`) | `["git","gh","ssh"]` | ✓ |
| `passable_env_names` | identifierSet \| null | `null` | ✓ |
| `mcp_package_allowlist` | unique array of nonEmptyString | `[]` | ✓ |
| `shadow_acknowledged` | array of closed `{env, path}` | `[]` | ✓ |
| `secret_material_waivers` | array of closed `{pin (commit hex), file, span[2], reason}` | `[]` | ✓ |
| `backup_retention` | integer ≥ 0 | `5` | ✓ |
| `require_current_profile` | identifier \| null | `null` | ✓ |
| `in_place_mode.<env-id>` | identifier → enum `linked`,`copied` | `{}` | ✓ |

Schema-1 file and cases: byte-unchanged (`git diff --stat` empty for `manager-config-v1.schema.json` and its cases).
`schema1-rejects-environments` (valid: false under schema 1) and `invalid-schema-version-1` (schema-1 file rejected by
schema 2) both exist and pass; COMPATIBILITY paragraph matches that behaviour.

## Manager §12 (Decision 0012 impact rows)

- 12.1 bytes change ✓ (MCP channel rows, reserved `curator-mcp`, forms, XDG seeds, targets, shadowing, tool release).
- 12.2 bytes change ✓ (marker contents per env §8.2, `in_place_mode`, claude_code copy exception, backups + `backup_retention`, drift).
- 12.3 rewritten ✓ (install grammar with one root / `--directory` / `--as` / name-less `--use` / `--range latest` / no `--branch`;
  overlays + precedence knobs; `profile use` shims/partial/`require_current_profile`; scoped `--clear`; `profile update` order;
  `profile remove [--purge]` + orphans; `env unmanage [--restore-backups]`; lock-writing skill scope; `default` migration; onboarding/import).
- 12.4 unchanged → **extended**: correct. Environments §12.1 places `isolation.<profile>.<env-id>` under §7.4 and §12.2 makes it
  lockable; the brief's knob rule obliges the manager to state it, and §7.4 adds the liveness row and seeds. The 0012 row is the
  stale one (already filed under TASK-260905-2tqh59 item 1). Not re-raised.
- 12.5 bytes change ✓ (tuple, `mcp` section, lock-free read, `--repair` with 1–60 s bounded wait, formats/digest,
  `passable_env_names`, `system_prompt_files.<profile>.pi`, `curator run` pointer).
- 12.6 bytes change ✓ (detector scope incl. `agent-mcp.json` args/url, unpinnable, waivers, `context-system-module-present`, `path` snapshot).
- 12.7 bytes change ✓ (`profile list` columns, `env status` rows per env §12, currency rule, warnings, GC roots incl. retained previous lock).
- §12.2 lockable keys named in the §12 intro; `isolation` only toward `shared`; credential rule restated as a citation.
- Every one of the 42 diagnostic codes named in manager §12 exists in environments.md. Citation spot-checks (>10): bounded wait
  1–60 s (§10.1), `APPEND_SYSTEM.md`/`SYSTEM.md` (§5.5), `curator-mcp.config.toml` (§5.8), `.agent-environment-backup/<n>/` and
  retention semantics (§8.3), old lock retained until next GC (§9.2), `profile remove` retention/orphans (§9.2), `env unmanage`
  clears current (§9.2), re-install-as-update (§9.1), `seeded_projects` (§8.2), `profile_use_partial` (§9.3), `curator run`
  `--repair`/provider mapping (§9.2, §10) — all consistent.

## CLI rows

Install grammar, `profile list` columns, `profile use --clear`, `profile update [<name>|--all]`, `profile remove <name> [--purge]`,
`env unmanage [--restore-backups] [--env] [--target]`, `env backups scrub [--older-than <days>]`, `env resolve … [--repair]`,
`env status` rows, informative `profile compose` / `env config`, `curator run` provider pointer — all present, one line per
command, flags spelled as environments.md spells them. Examples block carries the four required lines plus `--use`, `--clear`,
`env unmanage`.

## Findings (none blocking)

**F1 — minor — `tools/validate.py` / schema-2 cross-check.** A widening mutant survives: adding `first` to the
`precedence.winner` enum leaves `make validate` green (M4). The cross-check compares knob names and literal defaults against
the §12.1 table but not enum value sets, and each enum has exactly one negative case (`"winner": "later"`), so any widened
enum passes. Shape: gate widened, not detected. Fix: extend the §12.1 cross-check to the backticked value lists of the
`Values` column (all enums are literal there), or add a negative per enum that asserts the enum set equals the table set.
repeat-of: none.

**F2 — minor — `conformance/v1/schema-cases/manager-config-v2/`.** The brief asks one negative per value grammar; the
`overlay.range` grammar (`versionRange` pattern), `overlay.tag` (`gitRefName`), and empty `overlay.source` have no negative
case (`invalid-overlay-revision-grammar` covers only `revision`). Fix: three generated cases in `manager_config.go`.
repeat-of: none.

**F3 — nit — `profiles/manager.md` §12.5 and `cli/curator.md`.** "`curator run` … under Decision 0013 Decision 6.4 — it
always resolves with `--repair`": the provider column is D6.4, but the always-`--repair` rule lives in environments §9.2
step 5 and the §10 `curator run` paragraph (0013 never says `repair`). Cite environments for that half.
repeat-of: none.

**F4 — observation.** Drafting-report inconsistencies 1–5 acknowledged as TASK-260905-2tqh59; not re-raised.

## Disposition

ACCEPT: `accept_cr(TASK-260905-369vye, revision=1, evidence=TASK-260905-369vye_review-verdict.md)`. F1–F3 are follow-ups for
the next batch, not rework of this revision; the schema matches §12.1 byte for byte at this revision and both gates are green.
