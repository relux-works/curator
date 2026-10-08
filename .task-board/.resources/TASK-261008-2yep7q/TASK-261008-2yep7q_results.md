# TASK-261008-2yep7q results (developer): re-apply accepted 2am4qa rev2 onto curator-spec main (rc.14)

## 1. What conflicted / moved since the rev2 base (3d2c611, 2026-09-24)

- Namespace rename: `schemas|conformance/draft-sources-v1` became
  `schemas|conformance/skillfile-sources-v1`. All rev2 draft-namespace paths were
  remapped (file locations and schema `$id`s). `git apply --3way` handled
  CHANGELOG/manager/skillfile-sources/schemas-v1-README/validate.py mechanically;
  core.md, test_validate.py and every renamed path were ported by hand.
- Core marker v5 was minted since (framing v2, rc.14-era), so bare "marker v5" is
  now ambiguous. Rev2's marker v5 is the source-extension marker v5; every ported
  sentence now qualifies the lineage (core vs source-extension).
- Core section 10 was rewritten around framing v2; rev2's section-10 hunks were
  re-anchored, not applied verbatim.
- New freeze policy post-rev2: `validate_released_schema_immutability` (introduced
  by b1a2efb) requires every `schemas/**/*.schema.json` byte-identical to tag
  v1.0.0-rc.13. In-place edits to install-marker-v5, skillfile-v2,
  skillfile-lock-v1, source-types-v1 are therefore forbidden (the repo's own
  ReleasedSchemaImmutabilityTests prove the gate bites).
- rc.14 is tagged (v1.0.0-rc.14). `release/1.0.0-rc.14.json` is frozen and pins the
  138-file skillfile-sources manifest. This change is the first post-tag suite
  change, so 3 RC14 release-gate tests fail (see section 4).
- "rc.9" wording throughout rev2 updated to rc.14; "draft namespace" wording updated
  to source-extension namespace. Schema 9 itself stays unreleased/opt-in (draft).

## 2. Changes versus rev2 (policy-required deviations only, plus mechanical remaps)

- NEW schema revision (schema policy requires it; the brief allows it, "say which"):
  skillfile-sources `install-marker-v6` = marker v5 bytes + `schema_version` 6 +
  `skill_schema_version` max 9 (3-line delta, asserted by recipe and README audit).
  The frozen v5 band (max 8) cannot record a schema-9 installation and no other
  marker permits 9, so the amendment is unrecordable without it.
  - Vectors (`manifest-dependency-directories.json`) expect `marker_schema_version`
    6 where rev2 said 5; the recipe checks 6.
  - New cases: `install-marker-v6/{valid,invalid}.json`, plus
    `install-marker-v5/invalid-skill-schema-version-9.json` proving frozen v5
    rejects 9 (the negative evidence for minting v6).
  - New narrowing mutant test: v6 rewritten with max 8 is rejected by the recipe.
  - The core.md / schemas-README forward pointer now reads: lock-v2 and audit-v2
    adopt `hash_version`, and the schema-9 `install-marker-v6` gains it, in
    follow-up TASK-260930-3ny11n (that task is backlog; v6 is unreleased, so it
    can still adopt `hash_version` before any tag freezes it).
- DROPPED rev2's `$ref`-sharing refactors of frozen skillfile-v2, skillfile-lock-v1,
  source-types-v1. Byte-freeze forbids them and they are zero-semantic-delta:
  verified all four inline directory grammars JSON-equal to
  source-types `$defs/directory`. Core section 4.4 now says schema 9 "reuses the
  shared `directory` definition" instead of "the JSON Schemas share one
  definition". The recipe enforces the v9 `$ref` only.
- Recipe renames: `draft_schema_registry` -> `skillfile_sources_schema_registry`,
  `validate_draft_source_schemas` -> `validate_skillfile_sources_schemas`,
  `DRAFT_SUITE` -> `SOURCES_SUITE` (+ `SOURCES_SCHEMAS`). The returned paths map
  covers the sources dir only: both namespaces now contain
  `install-marker-v5.schema.json`, so a shared name map would shadow.
- Verbatim from rev2: core section 4 intro/table/gates (modulo namespace/rc), the
  section 4.4 normative text (modulo the one definition->grammar sentence), the
  section 7 unification rule, all three manager.md additions (modulo marker-v6
  wording), all skillfile-sources.md additions except the marker-writer sentence
  (v6) and the marker-5 migration table row (kept "1 through 8": marker 5 is
  frozen; v6 documented in a new paragraph instead), all 22 v9 schema cases, the
  vectors file (modulo marker 6), the vector-semantics recipe, the 4 mutant tests.
- v9 schema files kept byte-identical to rev2 except the `$id` namespace (including
  the odd but resolving `../v1/../v1/common.schema.json` refs: fidelity over cleanup).
  The independence gate normpaths them; all 152/152 refs byte-identical to rc.10.
- CHANGELOG entry placed under Unreleased/Added (rev2's hunk targeted the old
  Unreleased section, which has since become rc.14).

## 3. Files

- New: `schemas/skillfile-sources-v1/{agent-skill-v9,csk-skill-v9,install-marker-v6}.schema.json`;
  `conformance/.../manifest-dependency-directories.json`; 22 v9 schema cases;
  3 marker cases.
- Modified: `protocol/core.md`, `protocol/skillfile-sources.md`, `profiles/manager.md`,
  `schemas/{skillfile-sources-v1,v1}/README.md`,
  `conformance/.../{README,index}.json`, `tools/validate.py`, `tools/test_validate.py`,
  `CHANGELOG.md`; regenerated `conformance/.../manifest.json` (138 -> 167 files) and
  the `conformance/candidate.json` skillfile pin.
- Untouched: `conformance/v1/*`, `schemas/v1/*.schema.json`, `release/*` (rc.14 corpus
  and release records byte-identical). No LOGBOOK edits (none in repo; brief forbids).

## 4. Validation (command, exit code, tail)

- `python -B tools/validate.py` -> exit 0 (ran twice, final state):
  `validated 73 schemas and 1294 vector files`
- `python -B -m unittest discover -s tools -p 'test_validate.py'` -> exit 0:
  `Ran 567 tests in 656.281s / OK` (pytest is not installed in the dev venv; the
  brief allows the Makefile target, which uses unittest)
- New `ManifestDependencyDirectoryTests` alone -> exit 0: `Ran 5 tests / OK`
- Full `unittest discover -s tools -p 'test_*.py'` -> exit 1: 674/677 pass. The 3
  red are one root cause, all in `ProtocolRC14ReleaseGateTests`, which run the
  rc.14 release gate against a copy of the live tree:
  - `test_accepts_complete_rc14_artifact_set` (ERROR),
  - `test_rejects_removed_lifecycle_case_with_updated_manifest_hash` (FAIL),
  - `test_rejects_stale_compiled_fixture_with_updated_manifest_hash` (FAIL),
  all with `ReleaseFailure: 1.0.0-rc.14 metadata does not pin the accepted source
  suite separately`: the gate requires the live skillfile manifest sha to equal the
  frozen `release/1.0.0-rc.14.json` pin, and the regenerated 167-file manifest
  legitimately differs. No in-scope fix exists: advancing the pin means editing a
  tagged release record (forbidden by `validate_released_record_immutability` and
  `TestRC14ReleaseMetadataPinsCandidateAndPreservesSourceBaseline`), and
  re-pointing the gate means rc.15 release prep (release-process ownership, out of
  scope for this re-application). RECOMMENDATION: the orchestrator sequences rc.15
  prep with landing; the reviewer judges fidelity on validate.py + test_validate +
  suite gates, all green.
- `python tools/verify_skillfile_sources_independence.py` -> exit 0:
  `152/152 byte-identical to v1.0.0-rc.10`, `31/31 present`, gate passed
- `go test ./tools/...` -> exit 0: `ok .../tools/generate-vectors 9.354s`
- Regenerate idempotence (`go run ./tools/generate-vectors -root .` twice) ->
  exit 0, all four regenerate-check paths byte-stable; `conformance/v1` and
  `release/` carry zero modifications vs HEAD
- Embedded suite spec command (conformance README, incl. new v5->v6 audit) ->
  exit 0: `146/146` cases, `11/11` wire schemas
- `git diff --check` -> clean; `gofmt -l tools` -> clean
- Work left uncommitted in the story worktree per instructions.
