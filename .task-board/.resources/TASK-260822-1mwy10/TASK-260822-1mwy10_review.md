# TASK-260822-1mwy10 review — manifest-schema-execution-policy

Verdict: **accepted**. Reviewer run RUN-260822-53fb45 (not goal-bound).

Subject: `spec/sw-schema` @ `ebfed81` in `curator-spec`, base `origin/main` = `b92b105`.
Reviewed read-only in a scratch detached worktree at the same commit, not the
implementer's worktree.

## Provenance

- `ebfed81` signature: **Good** ECDSA `SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM`,
  `oparin@me.com`, verified against `maintainers.allowed_signers`.
- Commit message carries no AI attribution (`co-authored-by|claude|anthropic|generated with` -> no match).
- Pushed: `origin refs/heads/spec/sw-schema` = `ebfed8171cd49eec0c8c010801929a01d2352569`. No PR, as instructed.

## Gates re-run independently at ebfed81 (fresh venv, jsonschema 4.25.1)

| Gate | Result |
|---|---|
| `python tools/validate.py` | exit 0 — `validated 52 schemas and 656 vector files` |
| `python -B -m unittest discover -s tools -p 'test_*.py'` | exit 0 — `Ran 94 tests ... OK` |
| `go test ./tools/...` | ok `tools/generate-vectors` |
| `go vet ./tools/...` | clean |
| `gofmt -l tools` | no output |
| `git diff --check` / `git show --check ebfed81` | clean (CI `formatting` job) |
| `go run ./tools/generate-vectors -root .` x2, each followed by `git diff --exit-code -- conformance/v1 release/*.json` | 0 on all four checks — double regeneration proven byte-clean |

Not run, same as the implementer reported: the `links` job (lychee, network) and
`release-provenance` (token, and `if: github.event_name != 'pull_request'`), and
`make release-check` (needs `VERSION`, release-time only). The implementer's
reported counts (52/656, 94 tests) reproduce exactly.

## Frozen-byte audit

`git diff --diff-filter=M b92b105..ebfed81` — **9 modified files, 185 added, 0 deleted**:

- `conformance/v1/manifest.json` and `conformance/v1/schema-cases/index.json` — inventories only.
  **No schema-1..7 or marker-v1..v3 case instance changed a byte.**
- `release/1.0.0-rc.8.json` — `candidate_protocol_pin.manifest_sha256` and
  `downstream_consumption.required_manifest_sha256`. This is **required**, not a
  freeze violation: `tools/validate.py` pins rc.5/rc.6/rc.7 by exact digest
  constants (`RC5/RC6/RC7_RELEASE_METADATA_SHA256`) and separately asserts that
  rc.8's pin *equals* the live suite manifest digest. rc.5/rc.6/rc.7 bytes are
  unchanged. `make regenerate-check` lists all four files for exactly this reason.
- The five source files: `schemas/v1/README.md`, `schemas/v1/common.schema.json`,
  `tools/generate-vectors/main.go`, `tools/generate-vectors/main_test.go`,
  `tools/test_validate.py`, `tools/validate.py`.

## Schema shape verified structurally

- `agent-skill-v8` vs `-v7` differ in exactly 4 lines: `$id`, `title`,
  `schema_version` const 7->8, `commandV7`->`commandV8`. Same for `csk-skill-v8`.
- `install-marker-v4` vs `-v3` differ in exactly 4 lines: `$id`, `title`,
  `schema_version` const 3->4, `skill_schema_version` const 7->8. The repo asserts
  this itself in `test_marker_v4_records_schema8_installations_under_v3_rules`,
  which strips those two properties and compares the whole documents.
- `scriptCommandV8` = frozen `scriptCommand` + two OPTIONAL fields, `additionalProperties: false`,
  `anyOf` path rule inherited, `dependentRequired` both directions.
- `commandV8` = `commandV7` with only the script branch swapped. `commandV7` still
  refs the frozen `scriptCommand`, so schema-7 meaning is untouched.

## Schema behaviour verified by independent probing (37 cases, not the repo's own vectors)

Every rejection path the prose claims holds, and the reverse also holds:

- schema 8 accepts: both fields; neither (declared-only); node-v1; unix-only; win-only.
- schema 8 rejects: `execution_policy` alone; `interpreter` alone; `null`; `"none"`;
  `false`; `manager-worker-v1`; `hardened-worker-v1`; `script-worker-v2`; `bash-v1`;
  pathless enforced command; the fields on a system, `go-v1`, or `go-repository-v1`
  command; the fields at manifest top level.
- schemas 1..7 reject both fields at top level and on a command. **Attribution proven**:
  for each of v1..v7 the identical instance *without* the schema-8 fields ACCEPTs, so the
  rejection is caused by the new fields and is not an incidental invalidity. The prose's
  split is exact — v2..v7 reject via `additionalProperties`, v1 (open top level) rejects
  via `tools/validate.py` wire semantics, message `execution_policy is legal only in
  manifest schema 8`.

The prose's capability paragraph is factually correct: `$defs.capabilities` does carry
`default` annotations `network: "none"`, `filesystem: "repo"`, `exec: "none"`,
`secrets: "none"`, `env_read: []`.

## Cross-branch coherence (checked directly, not taken on trust)

`origin/spec/sw-core-prose` @ `41cf556` (TASK-260822-1f533i), written independently:

- core.md:211 — "Selection is an OPTIONAL per-command field on a script command of
  manifest schema 8 or later" — matches the schema's per-command, no-manifest-default choice,
  which resolves decision 0008 open question 1.
- core.md:256-257 — "An enforced script command names a closed interpreter identifier.
  Protocol 1.0 admits exactly `python3-v1` and `node-v1`" and core.md:282 defers
  `bash-v1`/`powershell-v1` for platform-resolution reasons — matches
  `$defs.scriptInterpreterV1` and the README's rejection prose. Resolves open question 2.
- core.md:417 — declared host globs mean "recorded and reported", not "any filtering,
  allowlisting, resolution failure, or denial" — matches the README's reporting-only
  reading. Resolves open question 3.
- core.md:~300 — "a capability field the manifest does not contain derives the
  deny-by-default meaning ... and the schema default for that field MUST NOT widen a
  derived control" — matches the README's annotation-vs-derivation paragraph.

No reconciliation is needed between the two branches. The `interpreter` field is not an
unauthorised addition beyond decision 0008: it is the normative resolution of that
decision's open question 2, and both halves of the story landed on the same answer.

## Owed deltas — verified genuinely still owed, not stale notes

Confirmed absent on `origin/spec/sw-core-prose@41cf556`:

- core.md:157 still reads "exactly one of `agent-skill-v1.schema.json` through" v7.
- core.md:182 schema table stops at row 7.
- core.md:186 still reads "schemas 2 through 7 reject unknown fields".
- core.md:1539/1554 still read "MUST read marker schemas 1, 2, and 3" and
  "Marker v3 permits `skill_schema_version` through 7".
- `COMPATIBILITY.md` has no rc.9 paragraph (already inside TASK-260822-c0rxj7 AC).

The implementer's handoff notes are accurate and actionable. `CHANGELOG.md` has no
Unreleased section, so nothing is owed there before release.

## Advisory for the downstream tasks (not defects, not blocking)

1. **TASK-260822-3nvx91 (module roots into schema 8).** The README says module roots
   "extends the build-command branch reached from `$defs.commandV8`". `commandV8`'s build
   branches currently `$ref` `buildCommandV6` and `repositoryBuildCommandV1`, both **shared
   with the frozen `commandV6`/`commandV7`**. That extension must add a new `$def`, never
   edit `buildCommandV6` in place, or it silently mutates schemas 6 and 7. The repo's own
   legacy-case regeneration tests would catch it, but the wording invites the wrong edit.
2. **TASK-260822-3nvx91 / TASK-260822-c0rxj7.** Whichever of the two schema-8 branches
   lands second must regenerate: this branch has already generated 185 schema-8 case files
   and the rc.8 candidate pin for a schema 8 that does not yet carry module roots.
3. **TASK-260822-f4qv7w.** Confirmed the two hooks the implementer flagged:
   `mixed_build_cases` (`tools/generate-vectors/main.go:1187`) has rows only through
   `schema7-*` and no schema-8 row, and `install-marker-v4` has 27 structural schema cases
   (mirroring v3 exactly) but no expected marker fixture under `conformance/v1/expected/`.
   Neither misstates schema 8; both are coverage gaps in the correct downstream scope.

## AC

| AC / DoD | Status |
|---|---|
| Schema files committed on the story branch | met — `ebfed81` on `spec/sw-schema`, signed, pushed |
| Validation prose with defaults and rejection paths | met — `schemas/v1/README.md`, every claim probe-verified |
| Schema validation gate passes | met — reproduced independently, exit 0 |
| Numbering coordination with STORY-260822-1pm1c9 recorded | met — in board notes *and* durably in the README "Schema-version numbering" section |
| Tests for new/changed behaviour, passing | met — `TestSchemaV8WireSurfacesAreClosedAndVersioned`, `TestGeneratedSchemaV8CasesCoverTheScriptWorkerOptIn`, 3 new python tests; 94 pass |
| Lint clean / build not broken | met — gofmt, go vet, `git diff --check` all clean |
| Implementation matches AC | met |
| Solution fits project architecture | met — additive versioned wire family, frozen predecessors, closed constants, generator-owned vectors |

Reviewer supplied no `commit_ack`; the scope is already committed and pushed at `ebfed81`.
