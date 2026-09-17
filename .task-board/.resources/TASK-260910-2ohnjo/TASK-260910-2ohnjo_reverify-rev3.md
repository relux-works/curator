# Re-verification — TASK-260910-2ohnjo rev3 (run of 2026-09-17)

Role: doc-writer. This run found the rev3 candidate already complete in the
Story worktree (`/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260910-1lf0m5/worktree`,
base `07e2b41`) with `TASK-260910-2ohnjo_spec-patch_rev3.patch` and
`TASK-260910-2ohnjo_evidence.md` attached. No file was edited in this run;
everything below is an independent re-run against the untouched candidate.

## Patch = worktree

`git diff 07e2b41 | git patch-id --stable` and the attached rev3 patch both
yield `3fcb1de6fe40f2753141b2f8e69fd3ea2c2a3be1`. Match.

## Gates re-run in this run (all exit 0, `set -o pipefail`, repo venv first on PATH)

1. `python3 tools/validate.py` → exit 0:
   `validated 60 schemas and 1048 vector files` (the S4 semantic consumer
   runs inside this entry point and accepts the published vector, including
   the `args-with-space` / `args-with-escaped-quote` positive cases).
2. `python3 -B -m unittest test_validate.EnvPassthroughVectorTests` (from
   `tools/`) → exit 0: `Ran 27 tests … OK`, including both round-1 review
   mutants (`test_review_mutant_enforce_absent_passes_is_rejected`,
   `test_review_mutant_surfacing_bytes_are_rejected`) and the six rev3
   parser tests (space arg, escaped quote, delimiter-like string, extra
   column, padded JSON, space/quote gate pass).
3. `go test ./tools/...` → exit 0:
   `ok github.com/relux-works/curator-spec/tools/generate-vectors`.
4. Literal `make regenerate-check` → exit 0 (`regen_check_exit=0`).
5. `git diff --check` → exit 0.

## Vector / manifest spot-check (this run)

`conformance/v1/vectors/environments-env-passthrough.json` holds 25 cases
across its families; `surfacing_cases` contains `args-with-space` and
`args-with-escaped-quote`; `conformance/v1/manifest.json` references the
env-passthrough vector.

## Accepted from attached evidence (not re-run here)

The full `python3 -B -m unittest discover` sweep (254 tests) and the literal
single-call `make validate` were quoted green in `TASK-260910-2ohnjo_evidence.md`
(rev2 literal run: 248 OK; rev3 split runs: 254 OK). This run re-ran the S4
subset (27/27 OK) plus the `validate.py` and Go gates rather than the ~9 min
full sweep, per the headless-run time bound.

## Scope

`git status --short` lists only the same 11 spec/schema/vector/manifest/
CHANGELOG/tooling paths; no implementation code, no new edits. Ready for
review round 3 against the rev3 patch.
