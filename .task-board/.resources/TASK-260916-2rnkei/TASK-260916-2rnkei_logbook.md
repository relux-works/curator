# Logbook — TASK-260916-2rnkei (spec-codex-seed-mcp-residual, E3)

No board logbook facility exists in this environment and the campaign rules
forbid `LOGBOOK.md` edits, so this task-scoped outcome resource is the
durable logbook record. Full detail: `TASK-260916-2rnkei_evidence.md` §5.

## 2026-09-17 — decisions

- Strip-at-provisioning, warn-first, exactly as the brief settles it:
  revision A (whole-copy + `mcp_native_servers_ungoverned` + migration
  hint) MUST ship before revision B (strip + one
  `mcp_native_servers_not_inherited` report). No knob selects the revision
  (manager-shipped, like the update-confirmation revision).
- Seed record as additive optional `codex_seed_record` on
  `agent-environment-marker-v1`, NOT a new `marker-v2`: v1 is unreleased
  batch (absent at rc.8, mutated in `fcdb9ba`, not in the README frozen
  list) — same practice as E1's `manager-config-v2` additions. Old readers
  keep `version: 1` semantics.
- Semantic (parsed-member) strip rule, not byte-exact: seeded file MUST
  parse as TOML with every top-level member except `mcp_servers`; kept
  member serialization is implementation-defined. Byte-excision rejected
  as brittle to specify (multi-line inline tables, dotted keys).
- Warning fires exactly when the provisioning-time name snapshot is
  non-empty (covers the empty-`[mcp_servers]`-table edge: strip, record,
  no warning).

## 2026-09-17 — findings / anomalies

- New schema-cases for the new marker field are impossible without Go
  generator edits (`schema-cases/index.json` is generator-rewrite-owned);
  none added — existing marker cases pass unchanged. Out of role scope
  (doc-writer, read-only on code), same for a `tools/validate.py` gate
  over the new vector file (S4/E1/S6 added gates; suggested follow-up).
- E7 adjacency: the new §7.8 residual table also states the
  `--strict-mcp-config` asymmetry E7 (`STORY-260916-33vuzm`,
  `TASK-260916-2x2f7h`) plans a row for — that task can extend or no-op.
- No `profiles/manager.md` / `cli/curator.md` edits: no sentence there
  becomes false and E3 adds no flags (S4/E1 precedent).
- Validation: `make validate` exit 0 (62 schemas, 1094 vector files, 353
  unittest, go ok); `make regenerate` exit 0 with all pre-existing
  generated files byte-identical.
