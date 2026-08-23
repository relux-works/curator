# TASK-260822-3nvx91 results

## Delivery

- Worktree: `/Users/iv/Developer/ReluxWorks/curator-spec/.temp/STORY-260822-1pm1c9/prose-worktree`
- Branch: `spec/module-roots-prose`
- Starting commit: `ebfed81` (`spec/sw-schema`, manifest schema 8)
- Decision baseline followed: amended decision 0009 at `be7861c`
- Numbering coordination: one shared manifest bump. Decision 0009 extends `$defs.commandV8` with `$defs.buildCommandV8`; no schema 9 was created.

The branch now specifies schema-8 `modules` on local `go-v1` build commands, the declaration/directive bijection, admitted unversioned directory replacement form, containment and scan rules, and pre-build diagnostics. The JSON Schema, validator, generator, tests, schema cases, suite inventory, and rc.8 candidate pin were updated together. Schemas 1 through 7 and their generated case directories remain byte-unchanged.

Thirty schema-8 cases were added: 15 canonical `agent-skill-v8` cases and the byte-equivalent 15 `csk-skill-v8` cases. They cover declared, empty, and absent lists plus duplicate, dot, parent escape, absolute, backslash, Windows device name, wrong type, null, wrong command kind, and top-level rejection.

## Important findings

1. A local Go probe proved that `vendor/modules.txt` distinguishes unversioned-left and versioned-left replacements. An unversioned directive emits a separate `# module => ../dir` annotation, while a versioned-left directive emits only `# module v0.0.0 => ../dir`. The initial prose ignored this distinction and would have accepted a forbidden versioned-left directive. Section 4.2.3 now requires annotation reconciliation and rejects an unmatched two-token left side.
2. The generator writes current cases but does not remove stale files. A red iteration that temporarily added legacy case names left those files on disk and therefore polluted the manifest inventory. They were explicitly removed before the clean regeneration. Legacy schema 1-7 case inventories remain frozen; their `modules` rejection is covered by `test_pre_schema8_manifests_reject_module_roots_surface` instead.

## Validation evidence

Green gates, each run directly as a standalone process:

- `go test ./tools/generate-vectors` — exit 0.
- task-scoped venv `python tools/validate.py` — exit 0; 52 schemas and 686 vector files.
- task-scoped venv `python -B -m unittest discover -s tools -p 'test_*.py'` — exit 0; 95 tests.
- `go test ./tools/...` — exit 0.
- `test -z "$(gofmt -l tools)"` — exit 0.
- `git diff --check` — exit 0.
- `lychee --no-progress --max-retries 3 --retry-wait-time 2 --accept 200,206,429 '**/*.md'` — exit 0; 40 OK, 0 errors, 1 excluded.
- deterministic regeneration digest comparison over `conformance/v1` and rc.5-rc.8 release JSON — exit 0; before and after SHA-256 both `6175eeb18e6a1f4885ac82114a96421a4e7115e0a7e6af379ecffeeafca58544`.

Expected/recoverable red evidence:

- Pre-regeneration `go test ./tools/generate-vectors` — exit 1 because the newly asserted module-root cases did not exist yet.
- First post-regeneration generator test — exit 1 because temporary additions changed frozen schema 1-7 case inventories; corrected by removing those legacy vector additions.
- First schema/Python suite attempt with system Python — exit 1 because `jsonschema` was not installed. Created `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260822-3nvx91/venv` and installed exactly `requirements-dev.txt` (`jsonschema==4.25.1`).
- First venv schema/Python retry — exit 1 because the generator manifest still listed stale legacy files from the red iteration; explicit cleanup and regeneration resolved it.

The exact CI `git diff --exit-code` regeneration check is not meaningful on an intentionally uncommitted delivery because it compares generated changes against `HEAD`. Determinism was instead proven by the identical before/after full generated-tree digest above.

## Version-control handoff

The reviewer accepted the technical implementation and requested delivery-only
rework. Under the task's explicit maintainer pre-authorization, the reviewed
scope was committed and pushed without AI attribution:

- Commit: `61ab80154c6aa8a83a33f2f2bbd8ec6e3dc1df50`
- Subject: `Specify declared first-party module roots`
- Signature: good (`G`), signer `oparin@me.com`
- Branch: `origin/spec/module-roots-prose`
- Remote verification: `git rev-parse HEAD`, the tracking ref, and
  `git ls-remote` all resolved to the same full commit hash.

Post-commit gates, each run directly as a standalone process:

- `PATH=<task-venv>/bin:$PATH PYTHONDONTWRITEBYTECODE=1 make validate` — exit 0;
  52 schemas, 686 vector files, 95 Python tests, and Go tool tests passed.
- `test -z "$(gofmt -l tools)"` — exit 0.
- `git diff --check` — exit 0.
- `lychee --no-progress --max-retries 3 --retry-wait-time 2 --accept 200,206,429 '**/*.md'` — exit 0; 40 OK, 0 errors, 1 excluded.
- `PATH=<task-venv>/bin:$PATH PYTHONDONTWRITEBYTECODE=1 make regenerate-check` — exit 0.
- `git show --check --oneline HEAD` — exit 0.
- Post-regeneration worktree and index checks — exit 0; branch clean and tracking
  `origin/spec/module-roots-prose`.
