# Producer brief: environments 1.1 batch 3 — manager §12, cli rows, manager-config schema 2

## Where and what

- Worktree `/Users/iv/Developer/ReluxWorks/.worktrees/curator-spec-env-manager`, branch
  `draft/environments-manager-cli-1-1`, base = curator-spec main `a68559b` (environments.md 1.1
  and its schemas/vectors landed).
- Scope: `profiles/manager.md` §12 (12.1–12.7 per the Decision 0012 impact table rows for
  manager §12, plus the knobs of environments §12.1/§12.2 and every lifecycle verb of
  environments §9 as the manager profile states them); `cli/curator.md` profile/env rows
  (install grammar `--range|--tag|--revision` with one root per install and `--use` taking no
  name; `profile list` columns; `profile update`, `profile remove [--purge]`, `env unmanage
  [--restore-backups]`, `profile use --clear`; informative `profile compose` and `env config`
  rows; `env resolve … [--repair]`; `env status` rows for the new liveness/shadow/seed notes;
  the `curator run` provider table pointer to Decision 0013 D6.4); `schemas/v1/manager-config-v2.schema.json`
  (schema 2 = schema 1 plus one closed `environments` object carrying exactly the environments
  §12.1 knobs with their value grammars and defaults; `schema_version` const 2; readers of
  schema 1 keep working — a compatibility note in `COMPATIBILITY.md` and `CHANGELOG.md`
  `## Unreleased`), `conformance/v1/schema-cases/manager-config-v2/` (positive: every knob
  present, minimal; negative: one per closed-object rule and one per value grammar),
  `conformance/v1/vectors/manager-config.json` extended for schema 2 through
  `tools/generate-vectors` + `tools/validate.py` (+ tests), `conformance/README.md` bullet,
  `conformance/v1/manifest.json` and the rc.9 pin regenerated.
- Do not edit `protocol/environments.md` or the batch-2 schemas; record any inconsistency you
  find as an exact sentence in the report for a follow-up.
- Deliverable: signed commit(s); `make validate` and `make regenerate-check` green (venv under
  `.temp/venv`); no push/tag/PR; `TASK-260905-369vye_drafting-report.md` attached (item → file/section
  table, knob → schema property table, gate tails); `task-board handoff TASK-260905-369vye --role developer`.
  Never write into the control root.

## Sources

`protocol/environments.md` at `a68559b` (§7, §8, §9, §10, §12 are the manager's normative
source; §12.1 table is the knob list, §12.2 the lockable subset); `decisions/0012` Compatibility
impact rows "manager §12.x" and "`cli/curator.md` profile rows"; `decisions/0013` D6.4 (the
`curator run` provider column, `--repair`); `profiles/manager.md` §1 (system configuration and
`locked`), §3.1, §5, §6 and the existing §12 text (keep its voice; rewrite only the rows the
impact table names as *rewritten*/*bytes change*); the existing `manager-config-v1.schema.json`
and its cases as the shape to extend; `cli/curator.md` table conventions.

## Rules

- The manager text never restates environments.md normatively — it cites sections and states
  the manager-side obligations (locks, transactions, ledger/backup discipline, status rows).
- Knob names, value spellings, and defaults MUST match environments §12.1 byte for byte; the
  schema encodes the grammars (e.g. `precedence.winner` enum, `overlays.<profile>` item object
  with exactly one of `range|tag|revision`, `secret_material_waivers` item shape,
  `shadow_acknowledged` item shape, `backup_retention` integer ≥ 0).
- CLI rows keep the table's one-line-per-command form; examples block extended with `profile
  install <url> --range '^1.2'`, `profile update`, `profile remove --purge`, `env resolve
  claude_code --repair --format json`.
- COMPATIBILITY.md: one paragraph — schema 2 is additive; a schema-1 file remains valid; readers
  reject an unknown `schema_version` explicitly.
