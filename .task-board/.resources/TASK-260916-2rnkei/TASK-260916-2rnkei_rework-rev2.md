# Rework brief — TASK-260916-2rnkei, revision 2 (E3 codex seed)

Revision 1 was rejected with two corrections
(`TASK-260916-2rnkei_review-verdict-rev1.md`, F1–F2). Everything else passed;
keep it byte-identical.

## Corrections (both required)
- **F1 — semantic validator gate (high; producer rule 7 is not optional).**
  `tools/validate.py` registers no codex-seed family: a digest in the manifest
  is not coverage. Add a `validate_environments_codex_seed_vectors` gate
  reached from `main()` that pins every required provisioning/posture
  scenario to its discriminating inputs (native `mcp_servers` present/absent,
  seed rule revision A/B, manager revision vs home's recorded revision,
  retained TOML members and values, diagnostics, marker seed-record shape,
  posture rows) and refuses a corpus where a named case no longer exercises
  its branch; negative replacement tests in `tools/test_validate.py`
  (the reviewer's `b-strips-servers-keeps-rest` ← `b-without-servers-no-warning`
  replacement MUST be refused through `python3 tools/validate.py`). Add
  generated positive/negative `agent-environment-marker-v1` schema cases for
  the closed `codex_seed_record` member (valid record; unknown field; wrong
  revision spelling; names not an array of identifiers).
- **F2 — manager revision vs home revision (medium).** Distinguish the
  manager-shipped seed rule revision from the home's RECORDED seed revision:
  a home provisioned under revision A (servers inherited) and now served by
  a revision-B manager keeps its bytes, lists its inherited names as
  ungoverned, and reports `mcp_seed_unstripped` with the re-provision hint —
  not only a home whose marker predates the rule (absent record). Fix
  §7.4/§7.7/§8.2/§12 wording accordingly, clarify §7.8 ("under revision B the
  base carries none" applies to B-provisioned homes), keep CHANGELOG in sync,
  add the pinned B-manager/A-home vector case and its replacement test.

## Validation and handoff
`make validate` and the regeneration proof (exit codes); evidence "Revision
2" section; `TASK-260916-2rnkei_spec-patch_rev2.patch` = `git diff HEAD` of
the worktree (base `684c9f1`) with new files via `git add -N`; curator
repository delta stays EMPTY; `task-board handoff TASK-260916-2rnkei --role doc-writer`.
Worktree and rules unchanged.
