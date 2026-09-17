# TASK-260916-2rnkei evidence — E3 codex provisioning seed (spec revision 1)

Finding: `docs/security-audit-2026-09.md` E3 (Medium) + Appendix B (E3 confirmed).
Worktree: `/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260916-1i1gfo/worktree`
(branch `task-board/story/STORY-260916-1i1gfo`, base `684c9f1`).
Role: doc-writer. No implementation code touched; no `tools/` edits (role boundary, see §6).

## 1. What changed per file

- `protocol/environments.md`
  - §7.4 seeds table, `codex_cli` row: rewritten — revision A copies
    `config.toml` whole, revision B copies every top-level member except
    `mcp_servers`; evidence cell kept honest (whole-copy listing now scoped
    to revision-A/pre-rule behavior as **verified** 0.153.2; remaining
    member shapes still **docs-confidence**).
  - §7.4 new block "**Codex seed MCP rule**": the two explicitly labelled
    rollout revisions (A MUST ship before B), provisioning MUSTs
    (`mcp_native_servers_ungoverned` with migration hint under A,
    `mcp_native_servers_not_inherited` reported once under B), the warning
    fires exactly when the snapshot is non-empty, names-only snapshot rule,
    provisioning-only scope, pre-rule homes keep bytes and report
    `mcp_seed_unstripped` with the re-provision hint.
  - §7.7 diagnostics table: +3 closed rows (`mcp_native_servers_ungoverned`,
    `mcp_native_servers_not_inherited`, `mcp_seed_unstripped`).
  - §7.8: closed per-adapter residual table (home MCP configuration under
    the channel) + launch-set definition sentence. opencode states the
    merge-order residual only.
  - §8.2 marker: `codex_seed_record` bullet — closed object
    `{ revision, native_mcp_servers }`, `revision` exactly `A`|`B`,
    snapshot ascending byte order, names only; absent everywhere except
    rule-provisioned managed `codex_cli` homes; absence on such a home =
    pre-rule (`mcp_seed_unstripped`).
  - §12: status enumeration gains the active codex-seed revision
    (`A`|`B`, §7.4) with behaviour + per-home record rows; the three new
    diagnostics join the warnings-never-non-current list; new informative
    codex-seed row paragraph (no configuration knob selects the revision).
  - §13: `vectors/environments-codex-seed.json` family enumeration +
    revision-conformance sentences incl. MUST NOT claim B while inheriting.
- `schemas/v1/agent-environment-marker-v1.schema.json`: additive optional
  `codex_seed_record` (`revision` enum `A`/`B`, `native_mcp_servers`
  unique non-empty strings, `additionalProperties: false`); non-managed
  modes forbid it (same `false` pattern as `seeds`/`seeded_projects`).
  No existing instance changes meaning: all prior schema cases still
  validate (proven by `make validate`, 62 schemas green).
- `conformance/v1/vectors/environments-codex-seed.json` (new, hand-written
  family in the `environments-source-signers.json` shape): 7 provisioning
  cases (A-whole-copy, B-strip incl. sub-table-only native form, A/B
  no-servers negatives, A inline-table form, B empty-table negative) and
  7 posture cases (A ungoverned, B not-inherited, pre-rule under A and B,
  A/B empty-snapshot negatives, non-codex negative). TOML fixtures are
  valid TOML (parsed with `tomllib` in the probe below); seeded-member
  expectations are parsed-shape assertions (serialization of the kept
  members is deliberately unspecified — see §6).
- `conformance/v1/manifest.json`: +1 entry (new file digest), written by
  `make regenerate` (sanctioned writer), not by hand.
- `release/1.0.0-rc.9.json`: 2 pin fields refreshed by `make regenerate`.
- `CHANGELOG.md`: Unreleased → Added entry "E3: …" naming both rollout
  revisions, the three diagnostics, the marker record, the §7.8 table, the
  vectors, and the manager/README follow-up (`TASK-260916-33abdk`).

## 2. Why

Settled brief decisions honoured without reopening: strip at provisioning,
warn-first; revision A keeps whole-copy + warning + ungoverned status rows;
revision B strips `mcp_servers` + one report + not-inherited rows; seed rule
at provisioning only; names-only snapshot in the marker seed record.

## 3. Validation transcript (final tree)

Shell: `bash`, workdir = worktree above. Python deps via project-local
`uv venv` + `uv pip install -r requirements-dev.txt` (`.venv` removed
after the gate; it never entered the diff).

- `make regenerate` → exit 0. Manifest diff exactly +4 lines (new entry);
  rc.9 diff exactly the 2 pin lines; every other generated file
  byte-identical (`git diff --stat` shows only the 5 intended tracked
  files + 1 new vector file).
- `make validate` → exit 0, full gate on the final tree:
  - `python3 tools/validate.py` → `validated 62 schemas and 1094 vector
    files`, exit 0 (run standalone as well as under make).
  - `python3 -B -m unittest discover -s tools -p 'test_*.py'` → `Ran 353
    tests ... OK`, exit 0 (via make).
  - `go test ./tools/...` → `ok .../tools/generate-vectors`, exit 0.
- Independent vector probe (`/tmp/e3_vector_probe.py`, throwaway): parses
  every `native_config_toml` fixture with `tomllib`, recomputes names and
  kept members, and asserts all 14 cases' expectations incl. every
  negative (null diagnostic, empty rows, adapter scoping) → all passed.
- Closed-set spelling check: all three diagnostics and
  `codex_seed_record`/`native_mcp_servers`/`revision` A-B appear
  identically in text, §7.7 table, schema, vectors, and CHANGELOG (grep
  table in run notes; no collisions pre-existed).

## 4. Deliberately out of scope

- Implementation (`TASK-260916-33abdk`), E7 launcher notes, S4 (landed),
  opencode merge order beyond stating the residual.
- No `profiles/manager.md` / `cli/curator.md` edits: no manager.md sentence
  becomes false (S4/E1 precedent touches only conflicting claims or new
  flags; E3 adds neither), and manager.md's diagnostic mirrors are already
  selective (S4 left them so).
- No new machine knob / lockable key: the brief names none; the revision is
  manager-shipped like the update-confirmation revision ("no configuration
  knob selects the revision").

## 5. Notes for the reviewer (Codex) and the manager task

- Marker schema versioning: `agent-environment-marker-v1` is NOT
  byte-frozen — it did not exist at rc.8 and was mutated during the 1.1
  batch (`fcdb9ba` "schema minors"); the frozen list in
  `schemas/v1/README.md` does not name it. The additive optional field
  follows the unreleased-batch practice (same as E1's `manager-config-v2`
  knob additions). No `marker-v2` minted; old readers keep `version: 1`
  semantics and §8.2 already tolerates unknown fields ("MUST NOT infer
  newer semantics from unknown fields").
- New schema-cases for the new field are impossible without Go generator
  edits (`schema-cases/index.json` is generated and rewrite-owned); none
  added. Existing marker cases still pass unchanged (regeneration proof).
  The record's accept/reject shape is covered normatively by §8.2 prose +
  schema + the vector fixtures.
- No `tools/validate.py` semantic gate for the new vector file: S4/E1/S6
  added gates, but this run's doc-writer role is read-only on code.
  The vectors are normative data for the manager implementation and the
  reviewer; a machine gate is follow-up work (suggested, not required).
- Semantic (parsed-member) strip rule, not byte-exact: the seeded file MUST
  parse as TOML carrying every top-level member except `mcp_servers`;
  comment/ordering preservation of kept members is implementation-defined.
  A byte-excision rule was rejected as brittle to specify normatively
  (multi-line inline tables, dotted keys).
- E7 adjacency: the §7.8 residual table also states the
  `--strict-mcp-config` asymmetry E7 (`STORY-260916-33vuzm`) plans a row
  for; `TASK-260916-2x2f7h` can extend or no-op without conflict.
- `listed_as` values in posture cases (`ungoverned`/`not-inherited`/
  `unknown`/`none`) are case scaffolding, not new closed vocabulary.

## 6. Checklist mapping

Items 1–7 satisfied (normative rule + spellings; warn-first revisions +
posture row; vectors + manifest + green gates quoted above; CHANGELOG +
patch + evidence; docs consistent; no discrepancies; outcome resources
attached). Item 8 (logbook) left unchecked: no board logbook facility
exists and campaign rules forbid `LOGBOOK.md` edits, so decisions are
recorded here and in task notes instead.
