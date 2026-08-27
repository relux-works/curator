# TASK-260822-1mwy10 — manifest schema 8: `execution_policy` on script commands

**Repo:** `relux-works/curator-spec`
**Branch:** `spec/sw-schema` (pushed, no PR — TASK-260822-c0rxj7 merges)
**Commit:** `ebfed81` "Add manifest schema 8 for the script-worker-v1 opt-in", signed, no AI attribution
**Base:** `origin/main` = `b92b105` (decision 0009 landed)
**Worktree:** `/Users/iv/Developer/ReluxWorks/curator-spec/.temp/STORY-260822-3k3hbs/schema-worktree`

## 1. What schema 8 admits

`$defs.scriptCommandV8` in `schemas/v1/common.schema.json` is the schema-7
script command plus exactly two OPTIONAL fields:

| Field | Binding | Meaning |
|---|---|---|
| `execution_policy` | `$defs.scriptExecutionPolicyV1` = `{"const": "script-worker-v1"}` | the command is enforced |
| `interpreter` | `$defs.scriptInterpreterV1` = `{"enum": ["node-v1", "python3-v1"]}` | closed identifier the manager resolves |

`dependentRequired` binds them in both directions: a command declares both or
neither. `$defs.commandV8` is a new union over `scriptCommandV8`,
`systemCommand`, `buildCommandV6`, and `repositoryBuildCommandV1`.
`schemas/v1/agent-skill-v8.schema.json` and `csk-skill-v8.schema.json` are the
schema-7 manifests with `schema_version` 8 and `commands` pointing at
`commandV8`; they are byte-identical to each other apart from `$id` and
`title`.

New files:

- `schemas/v1/agent-skill-v8.schema.json`
- `schemas/v1/csk-skill-v8.schema.json`
- `schemas/v1/install-marker-v4.schema.json`

## 2. Resolutions consumed from TASK-260822-1l4r4f

- **Q1 (placement).** Per command, one OPTIONAL field, one closed value, no
  manifest-level default and no override resolution. Implemented literally.
- **Q2 (interpreter identity).** Neither posed option: the package names a
  closed identifier, the manager owns resolution. The schema therefore carries
  `interpreter` as an enum and never a path, digest, argv, or shebang.
- **Q3 (network host globs).** Reporting-only, so no schema change; the
  `capabilities.network` shape is untouched.
- **Q4 (evidence cadence)** and **Q5 (Windows scope)** produce no manifest
  surface. The `script-capability-evidence-v1` record is prose + vectors, like
  `capability-evidence-v1`, and correctly has no JSON schema file.

## 3. Interpreter closed set — the one call the analysis left open

The analysis left the exact interpreter set to "the prose task's call". Schema 8
admits `node-v1` and `python3-v1` and no shell identifier: neither a POSIX shell
nor PowerShell resolves on all three supported platforms, and admitting one is a
specification revision under core.md 12.3.

**This converged independently.** `origin/spec/sw-core-prose` (TASK-260822-1f533i,
commit `41cf556`) states "Protocol 1.0 admits exactly `python3-v1` and `node-v1`"
and defers `bash-v1`/`powershell-v1` for the same reason. The schema and the
normative prose agree without reconciliation. Their capability-default sentence
("the schema default for that field MUST NOT widen a derived control") also
matches the schema prose in `schemas/v1/README.md`.

## 4. Defaults and rejection paths (prose in `schemas/v1/README.md`)

Defaults:

- Absent `execution_policy` is the default and the **only** spelling of
  declared-only. The command keeps its exact schema-7 meaning.
- `$defs.capabilities` `default` annotations are Draft 2020-12 annotations, not
  applied values. They state the declared-only reading. Under `script-worker-v1`
  an absent capability field takes decision 0008 section 3's deny-by-default
  meaning, which is narrower than the annotation for `filesystem`. The schema
  deliberately does not resolve that difference; `protocol/` prose is the
  authority. (This is finding F1 of the analysis, recorded rather than papered
  over.)

Rejection paths, each with a generated conformance case:

| Rejected | Mechanism |
|---|---|
| `execution_policy`/`interpreter` on a system or build command | that command's closed surface |
| `execution_policy` without `interpreter`, or the reverse | `dependentRequired` |
| `manager-worker-v1` or `hardened-worker-v1` as a script policy | the const — identities never alias |
| `script-worker-v2` or any successor | the const — needs its own identity and revision |
| `null`, `"none"`, `false` | the const — no second spelling of declared-only |
| an interpreter outside the closed set, including every shell | the enum |
| enforced command with neither `unix_path` nor `win_path` | inherited `anyOf` |
| the schema-8 surface in manifest schemas 1–7 | `additionalProperties` for 2–7; `tools/validate.py` wire semantics for schema 1, which keeps its deployed top-level extension behavior |

## 5. Install marker v4 (necessary consequence, not extra scope)

`install-marker-v3` pins `skill_schema_version` to `{"const": 7}` and marker v2
to 0–6, so a schema-8 installation had no marker version to record it — schema 8
would have been uninstallable. `install-marker-v4.schema.json` is marker v3 with
`schema_version` 4 and `skill_schema_version` 8 and **no other difference**
(asserted by a test that strips those two properties and compares the documents).
Every marker-v3 build-record rule applies unchanged and is proven by the same
generated case set.

## 6. Schema-version numbering coordination (STORY-260822-1pm1c9)

**Decision: one shared bump. Schema 8 carries both decision 0008
(`execution_policy` on script commands) and decision 0009 (first-party module
roots on local `go-v1` build commands). Module roots takes no separate version.**

Why shared rather than sequential 8 and 9:

1. One protocol release has never carried two manifest schema versions. rc.4
   shipped schema 6, rc.5 shipped schema 7.
2. The two surfaces are disjoint: script commands versus local build commands.
   Neither constrains the other.
3. Sequential numbering forces any manifest wanting both features onto schema 9,
   so schema 8 would be born already superseded.
4. Sequential numbering doubles the cost twice over: the legacy rejection matrix
   (every schema below N rejects N's surface) would need a second full pass, and
   the install-marker band would need marker v4 *and* v5, because
   `skill_schema_version` is a `const` per marker version.
5. No freeze is violated. `COMPATIBILITY.md` freezes *released* bytes
   ("A release never redefines old schema bytes in place"). Schema 8 has not
   shipped in any release; rc.8's frozen artefacts are untouched by this commit.

**What TASK-260822-3nvx91 has to do:** add the `modules` list to the build-command
branch reached from `$defs.commandV8` by adding a new `buildCommandV8` def and
swapping the `#/$defs/buildCommandV6` branch of `$defs.commandV8` for it. Do NOT
extend `buildCommandV6` in place: `$defs.commandV6` and `$defs.commandV7` both
reference it, so mutating it would silently give schemas 6 and 7 a schema-8
field. Do not create `agent-skill-v9`. Mirror the `legacyV8SchemaExamples` pattern in
`tools/generate-vectors/main.go` for the schema-1-through-7 rejection cases of
the `modules` field, and extend the `range(1, 8)` loop in `tools/validate.py`.

## 7. Deltas this branch deliberately did NOT make (owned elsewhere)

`protocol/core.md` is owned by TASK-260822-1f533i and `origin/spec/sw-core-prose`
does not yet contain these. They are required for the spec to be self-consistent
with schema 8 and should land with TASK-260822-c0rxj7 or as a follow-up:

1. **Section 4 preamble:** `agent-skill-v7.schema.json` → `agent-skill-v8.schema.json`,
   and `csk-skill-v7.schema.json` → `csk-skill-v8.schema.json`. Today core.md says
   a manifest conforms to "exactly one of v1 through v7", which excludes schema 8.
2. **Section 4 schema table:** add `| 8 | opt-in script-worker-v1 execution policy and its closed interpreter identity |`.
3. **Section 4 version gates:** "schemas 2 through 7 reject unknown fields" →
   "2 through 8"; add "Schema 1 through 7 MUST reject `execution_policy` and
   `interpreter` at the top level and on every command."
4. **Section 10 install markers:** "Managers supporting schema 7 MUST read marker
   schemas 1, 2, and 3" → schema 8 / marker schemas 1–4; "marker schema 3 for
   schema 7 installation mutations" gains "and marker schema 4 for schema 8
   installation mutations"; add "Marker v4 permits `skill_schema_version` 8 and
   otherwise carries marker v3's meaning unchanged."
5. **COMPATIBILITY.md** (already in the landing task's AC): the paragraph naming
   manifest schema 8, install marker v4, and the shared-bump numbering decision
   of section 6 above.

`profiles/manager.md` and `SECURITY.md` are TASK-260822-3fkfmf's; no schema-driven
delta is owed there beyond what that branch already carries.

## 8. Gates — commands run and real exit codes

All run in the worktree at commit `ebfed81`, each as a standalone process.
Python 3.14 venv at `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260822-1mwy10/venv`
with `jsonschema==4.25.1` from `requirements-dev.txt` (the system interpreter has
no `jsonschema`).

| Command | Exit | Result |
|---|---|---|
| `python3 tools/validate.py` | 0 | validated 52 schemas and 656 vector files |
| `python3 -B -m unittest discover -s tools -p 'test_*.py'` | 0 | Ran 94 tests, OK (91 before this change) |
| `go test ./tools/...` | 0 | ok |
| `gofmt -l tools` | 0 | no output |
| `go run ./tools/generate-vectors -root .` (run 1) | 0 | — |
| `git diff --exit-code -- conformance/v1 release/1.0.0-rc.5.json release/1.0.0-rc.6.json release/1.0.0-rc.7.json release/1.0.0-rc.8.json` | 0 | clean |
| `go run ./tools/generate-vectors -root .` (run 2) | 0 | — |
| same `git diff --exit-code` | 0 | clean — double regeneration proven per GOVERNANCE |

Not run: the `links` (lychee) and `release-provenance` CI jobs, which need network
and a GitHub token; `make release-check`, which requires a `VERSION` and is a
release-time gate. Nothing in this change adds an external link.

## 9. Frozen-bytes evidence

`git diff --name-only -- conformance/v1/schema-cases/` reports exactly one
modified file, `index.json`. All 186 other case changes are new files. No schema-1
through schema-7 case instance changed a byte, and `commandV7` still selects the
untouched `$defs.scriptCommand`, which is asserted by
`TestSchemaV8WireSurfacesAreClosedAndVersioned`. `release/1.0.0-rc.5.json`,
`rc.6`, and `rc.7` are unchanged; only `rc.8`'s candidate pin moves, which is
what that generated field is for.

## 10. Tests added

Go, `tools/generate-vectors/main_test.go`:

- `TestSchemaV8WireSurfacesAreClosedAndVersioned` — closed top level on all three
  new schemas; agent/csk schema-8 parity modulo `$id`/`title`; frozen
  `scriptCommand` and `commandV7`; policy is a single `const` and not an enum;
  interpreter set is exactly the reviewed pair; `scriptCommandV8` has exactly
  five properties and both `dependentRequired` directions; `commandV8` union
  order; marker v4 `schema_version` 4 / `skill_schema_version` 8.
- `TestGeneratedSchemaV8CasesCoverTheScriptWorkerOptIn` — 25 asserted generated
  case names incl. every rejection path in section 4; csk/agent case-set parity;
  marker v4 carries marker v3's branches; the enforced case really is the opt-in
  and the declared-only case carries neither field.
- Extended `TestLegacyManifestSchemaCaseNamesAndValiditySurviveRegeneration` and
  `TestGeneratedManifestV6CasesCoverBuildRejections` with the four
  `invalid-v8-*` rejection cases.

Python, `tools/test_validate.py`:

- `test_pre_schema8_manifests_reject_script_execution_surface` — 2 prefixes ×
  7 versions × 4 placements.
- `test_schema8_admits_enforced_scripts_and_keeps_schema7_rules` — enforced and
  mixed manifests accepted; the schema-7 repository bijection still enforced.
- `test_marker_v4_records_schema8_installations_under_v3_rules` — v3 conditionals
  hold, and v4 equals v3 once the two version properties are removed.

Generated cases: 186 new files. `agent-skill-v8` and `csk-skill-v8` get 51 cases
each (19 script-worker branches plus the inherited external-repository set),
`install-marker-v4` gets 27, and every schema from 1 to 7 gets the four
`invalid-v8-*` rejections under both the canonical and the legacy filename.

## 11. Follow-ups for TASK-260822-f4qv7w (conformance vectors)

`conformance/v1/vectors/` behavioural coverage stays that task's. Two concrete
hooks this change leaves open:

- `mixed_build_cases` in `tools/generate-vectors/main.go` maps `manifest_schema`
  to `marker_version`; it has no `schema8` row yet. Adding one records the
  schema-8 → marker-v4 binding behaviourally.
- `install-marker-v4` currently has only structural schema cases, not an expected
  marker fixture under `conformance/v1/expected/`.
