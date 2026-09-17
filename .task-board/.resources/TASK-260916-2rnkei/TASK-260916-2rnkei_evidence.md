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

---

# Revision 2 (rework after `TASK-260916-2rnkei_review-verdict-rev1.md`, F1–F2)

Everything from revision 1 that passed is kept; only the two verdict
corrections change the tree. Scope note: F1 explicitly requires
`tools/validate.py`, `tools/test_validate.py` and the Go generator edits,
so revision 2 touches those files — the rev-1 "role boundary" reading of
the doc-writer constraint is withdrawn for this rework, since the campaign
rule 7 gate is a required conformance artifact, not manager
implementation. No manager implementation code is touched.

## R2.1. F2 — manager revision vs home revision

- §7.4: the seed-rule paragraph now distinguishes the manager-shipped
  revision from the home's recorded seed revision. An A-provisioned home
  under a B manager keeps its bytes, lists its recorded names as
  ungoverned, and reports `mcp_seed_unstripped` with the re-provision
  hint; an empty recorded snapshot warns nothing (no inherited server to
  re-provision away). A marker that predates the rule (absent record)
  reports the same warning with the same hint.
- §7.7: the `mcp_seed_unstripped` row covers both predicates (absent
  record, or `A` record with non-empty snapshot under a B manager).
- §7.8: the residual row now says a home *provisioned under revision B*
  carries none in its base, with the A-home-under-B-manager residual
  stated in parentheses.
- §8.2: the recorded `revision` never changes after provisioning; an `A`
  record under a B manager keeps ungoverned rows and adds
  `mcp_seed_unstripped`.
- §12: the posture enumeration and the codex-seed row paragraph state the
  mismatch case (older recorded revision than shipped revision).
- §13 and CHANGELOG name the new posture case and the gate.
- Vector: new posture case `a-home-unstripped-under-b` (shipped B,
  `codex_cli`, `A` record with `[figma, gh]` → row B, diagnostics
  `[mcp_native_servers_ungoverned, mcp_seed_unstripped]`, listed
  ungoverned, repair hint true, current).

## R2.2. F1 — semantic validator gate and schema cases

- `tools/validate.py`: new `validate_environments_codex_seed_vectors`,
  registered in `main()`. Per provisioning case it parses the
  `native_config_toml` fixture with `tomllib` and recomputes the seeded
  members *and values* (new `seeded_members` expectation, see below),
  snapshot names, diagnostic, hint, and the closed seed-record shape; per
  posture case it recomputes the row from the shipped revision, the
  adapter (closed set reused from `ENVIRONMENT_HOME_VARIABLES`), and the
  recorded revision. Every named case is additionally pinned to its
  branch: revision letters, servers present/absent/empty, the TOML form
  (subtable / subtable-only / inline / absent / empty-table, read from
  the raw fixture text because the parsed document cannot tell inline
  from dotted form), the exact native top-level member set, and the
  shipped/recorded/snapshot tuple — so an internally consistent
  replacement under the same name fails.
- `tools/test_validate.py`: new `CodexSeedVectorTests`, 32 tests: the
  published vector passes; the reviewer's exact
  `b-strips-servers-keeps-rest` ← `b-without-servers-no-warning`
  replacement; the F2 `a-home-unstripped-under-b` ← `a-home-lists-ungoverned`
  and ← `b-home-lists-not-inherited` replacements; narrowing tests per
  refusal clause (revision flip, retained table/value/diagnostic/hint/
  record/row mutations, invalid TOML, non-table `mcp_servers`, unknown
  adapter, dropped cases, revision-string mutation); derivation corners
  beyond the corpus (empty A snapshot under B warns nothing; a B record
  never reports unstripped); and a `main()` registration guard.
- `tools/generate-vectors/environments.go`: 7 new generated
  `agent-environment-marker-v1` schema cases for the closed
  `codex_seed_record` member — `valid-codex-seed-record`,
  `valid-codex-seed-record-empty-snapshot`,
  `invalid-codex-seed-record-unknown-field`,
  `invalid-codex-seed-record-revision` (revision `C`),
  `invalid-codex-seed-record-names-member` (empty-string member),
  `invalid-codex-seed-record-names-not-array`,
  `invalid-codex-seed-record-on-linked-home` — emitted by `make
  regenerate` with index and manifest entries.
- Vector shape addition: each provisioning `expected` gains
  `seeded_members` (the retained parsed values), recomputed by the gate.
  This is what pins "retained values": a fixture value edited without
  its expectation fails; a consistent value edit still exercises the
  branch and passes, which is correct — the gate refuses branch
  collapse, not fixture bytes (bytes are pinned by git and the manifest
  digest).

## R2.3. Validation transcript (final tree)

Shell `bash`, workdir = worktree. Python via
`/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin`
(first on `PATH`). The `make validate` recipe was run as bounded
sequential calls (the suite's ~13 min wall exceeds one headless call);
every step below is the recipe's own command with its real exit code:

- `python3 tools/validate.py` → `validated 62 schemas and 1101 vector
  files`, exit 0 (16.7 s).
- `python3 -B -m unittest` split by class/module, all OK, exit 0 each:
  31 + 25 + 27 + 106 (incl. the 32 new `CodexSeedVectorTests`) + 55 +
  63 + 78 = 385/385 (rev-1 353 + 32 new).
- `go test ./tools/...` →
  `ok github.com/relux-works/curator-spec/tools/generate-vectors`, exit 0
  (2.8 s).
- `make regenerate` → exit 0. Manifest diff: +7 schema-case entries,
  index digest refresh, codex-seed digest refresh; rc.9: the 2 pin
  lines; `conformance/v1/schema-cases/index.json`: +7 entries. All 31
  HEAD-tracked vector files byte-identical (the changed vector file is
  the new E3 family itself); all 974 non-index HEAD-tracked
  schema-case files byte-identical.
- Regeneration proof: candidate committed to a scratch baseline,
  `go run ./tools/generate-vectors -root .` +
  `git diff --exit-code -- conformance/v1 release/1.0.0-rc.*.json` →
  exit 0 (generator idempotent). Note: `make regenerate-check` run
  directly in the uncommitted worktree reports the worktree-vs-index
  delta (exit 1) — that is the expected uncommitted-tree behavior, the
  same reason the round-1 reviewer ran the check against a scratch
  baseline.
- Production-entry mutant probes (scratch copy, mutant applied to the
  vector *file*, manifest refreshed with the repo generator, then the
  real `python3 tools/validate.py`): F1 reviewer's replacement →
  `validation failed: codex-seed case b-strips-servers-keeps-rest:
  native top-level members are not the pinned set (...)`, exit 1; F2
  replacement → `validation failed: codex-seed case
  a-home-unstripped-under-b: revision_shipped does not match the pinned
  one (B)`, exit 1. Replacement rejection at the production entry: 2/2.
- `git diff --check` → exit 0. Closed spellings verified identical in
  text, §7.7 table, schema, vectors, gate, and CHANGELOG; no
  open-ended wording in added lines. Curator repository delta stays
  EMPTY (all work is in the curator-spec story worktree).
