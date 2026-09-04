# env-schemas-notes — TASK-260901-1u50cr

Environments JSON Schemas and conformance/determinism vectors per
`protocol/environments.md` at `c3b29b1`, delivered on branch
`draft/environments-schemas` in worktree
`~/Developer/ReluxWorks/.worktrees/curator-spec-env-schemas` (base =
`origin/main` = `c3b29b1f7f37829fd4d0c50b2023efa2feb4c615`, verified equal at
verification time).

Signed commit: `cef93fbd95b43a13932bb0d8b397e177c5301045`
("Author the environments schemas and conformance vectors"). Good ECDSA
signature by the configured git signing key (the local `allowedSignersFile`
points at a stale temp path, so principal matching is unavailable locally;
the signature itself verifies).

Run provenance: three earlier spawn runs for this task (RUN-260901-8f5870,
-6623ac, -f27a3d) died with exit 1 on API network errors (`ENOTFOUND`), one
of them after authoring commit `cef93fb` but before any board evidence
landed. This run (RUN-260901-510d74) independently re-verified the full
deliverable at that commit and completed the lifecycle. No code changes were
needed; the only working-tree cleanup was removing an untracked
`tools/__pycache__/`.

## File list (89 files changed, +4410/-3)

Schemas (`schemas/v1/`), all strict Draft 2020-12 (`additionalProperties:
false`, closed enums/consts, `$ref` into `common.schema.json`):

- `profilefile-v1.schema.json` — §2: `version` const 1, non-empty `profiles`
  with identifier member names and portable-path values.
- `context-manifest-v1.schema.json` — §3: `version` const 1, REQUIRED
  possibly-empty `modules`; entries carry REQUIRED portable `path`, OPTIONAL
  non-empty unique `environments` selector, OPTIONAL `class` root|system.
- `agent-environment-marker-v1.schema.json` — §8.2: git/local profile pin
  branches (`commit` XOR `state_sha256`), composition⇔precedence
  `dependentRequired` in both directions, mode enum, per-surface entries
  (`paths`, optional `form`, `content_sha256`), `copy_fallback` required for
  `linked` mode and rejected otherwise, `seed_links` only on `managed-home`.
- `launch-env-fragment-v1.schema.json` — §10.2: `fragment` const, closed
  environment enum, pinned profile/composition members, closed `env`
  variable-name set, `system_prompt` with the four channel-descriptor kinds
  (`flag`/`config-key`+`key`/`variable`+`variable`/`file`+`filename`), closed
  semantics enum; empty `channels` admitted (opencode declares no channels
  in revision 1).

Cross-field rules JSON Schema cannot express are in `tools/validate.py`
`validate_wire_semantics` (aliased/nested profile roots, duplicate module
paths, sorted marker surface keys), matching the repo's existing convention
for manifest-schema semantics.

Conformance (`conformance/v1/`):

- `schema-cases/` — 57 new registered cases across the four schemas
  (16 valid, 41 invalid), one negative per rejection branch, wired into
  `schema-cases/index.json`.
- `vectors/environments.json` — 4 header cases (single-profile, composed
  default precedence, composed earlier-overrides-later, local state pin) and
  11 materialization cases (monolithic incl. selector exclusion, composed
  chapters incl. the empty chapter, zero-modules plain and composed,
  no-context-directory, referenced claude_code, referenced opencode incl.
  zero-modules with empty `instructions`, system-prompt composed and
  none-applicable) with per-file sha256 and the §5.6 surface hash.
- `expected/environments/` — exact expected bytes for every materialized
  file, cross-checked against the vector inventory in both directions.
- `manifest.json` + `release/1.0.0-rc.9.json` — the candidate protocol pin's
  `manifest_sha256` follows the regenerated manifest, exactly as the
  existing `regenerate-check` Make/CI target requires.

Tooling:

- `tools/generate-vectors/environments.go` (+`main.go` wiring) — the Go
  generator for the vector and expected bytes.
- `tools/validate.py` `validate_environment_vectors` — an independent Python
  reimplementation of the §5 byte rules (header grammar, part joining,
  chapter parts, referenced layout, opencode CCJ-1 + trailing LF,
  header-free system output, §5.6 content hash) compared byte-for-byte
  against the generated expected files, with exact case inventories so a
  dropped case fails closed. Wired into `main()`, so every `make validate`
  and CI run executes it — this is the production call site for the
  negative tests.
- `tools/test_validate.py` `EnvironmentVectorTests` — fail-closed negative
  coverage that narrows the gates: dropped case, precedence mutation, pin
  mutation, CRLF module bytes, missing trailing LF, selector widening,
  surface-hash mutation, absence↔written contradictions, opencode.json
  trailing-LF tamper, stale expected-file inventory, plus accept/reject
  pairs for each `validate_wire_semantics` rule.
- `tools/generate-vectors/environments_test.go` — grammar, determinism, and
  case-coverage tests on the Go side.
- `conformance/README.md`, `schemas/v1/README.md` — convention docs updated.

No CI edits were needed: `.github/workflows/ci.yml` already runs
`tools/validate.py`, unittest discovery, `go test ./tools/...`, and the
regenerate + `git diff --exit-code` determinism proof, so the new surfaces
are exercised on all three OS runners automatically.

## Prose ambiguities found

1. **"7-line grammar" (producer brief) vs §5.1** — the §5.1 grammar block is
   6 lines uncomposed (`<!--`, marker, `profile:`, `generated:`, `notice:`,
   `-->`) and 6+N+1 when composed (N `compose:` lines + one `precedence:`).
   Prose wins; the vectors follow §5.1 exactly, and the brief's "7-line" is
   read as informal (the display block with one compose line elided).
2. **Fragment `environment` ↔ `env` variable-name binding** — §10.2 says
   `env` maps "each registry-declared variable name for the environment",
   but the structural schema admits e.g. `environment: claude_code` with
   `CODEX_HOME`. The binding is machine-registry semantics (§7.1), analogous
   to the other cross-field rules this repo keeps out of the structural
   schemas; the variable-name value space itself is closed to the four
   registry names. Not a silent divergence — flagged here for the reviewer;
   if the reviewer wants it structural, a per-environment `if/then` on
   `propertyNames` expresses it.
3. **Marker `copy_fallback`** — §8.2 says the record notes "for a `linked`
   home — whether any entry fell back". The schema reads this as: REQUIRED
   on every surface entry when `mode` is `linked`, rejected otherwise
   (required-arrays-present-even-when-empty discipline applied to the flag).
   Negative cases cover both directions.

## Verification evidence (this run, 2026-09-01)

All commands run from the worktree with its `.venv` active (repo needs
`jsonschema==4.25.1` from `requirements-dev.txt`; system python lacks it —
`make validate` fails on import without the venv):

- `make validate` — exit 0: `validated 57 schemas and 766 vector files`;
  `Ran 147 tests ... OK`; `go test ./tools/...` ok.
- `go test -count=1 ./tools/...` — exit 0 (uncached).
- Twice-run determinism: `go run ./tools/generate-vectors -root .` run
  twice; sha256-over-sorted-sha256s of every file under `conformance/v1` and
  `release/` identical across runs:
  `a84f1c14e9eaa515870745f96341b12d6c09510d1e6dd987b1183110cb7728fc`, and
  `git diff --exit-code` against the committed bytes clean (NO-DRIFT), i.e.
  regeneration reproduces the commit byte-for-byte.
- `gofmt -l tools` — empty; `git diff --check` — clean.
- Expected bytes eyeballed against prose for the flagship cases: composed
  empty chapter, opencode CCJ-1 `{"instructions":[...]}` + exactly one
  trailing LF, header-free system prompt, zero-modules header-only file,
  and all four header cases including the `state sha256:` pin.

## Not done / out of scope

- No push, no tag (per brief). Branch `draft/environments-schemas` stays
  local in the worktree.
- `profiles/manager.md`, `cli/curator.md`, CHANGELOG, decisions/, protocol
  prose untouched (verified via commit file list); no §13 pointer edit was
  needed — the section's existing wording already names `schemas/v1/` and
  the vector surfaces without per-file references.
