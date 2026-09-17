# Rework brief — TASK-260910-2ohnjo, revision 3 (answers review verdict rev2)

Read `TASK-260910-2ohnjo_review-verdict-rev2.md` first; it is the authority for
this round. Everything it marks PASS/CLOSED stays exactly as it is (rollout
profiles, defaults, surfacing order, generator, fixtures, vectors). Work in the
same curator-spec Story worktree
(`/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260910-1lf0m5/worktree`,
uncommitted rev2 edits are still there); rules in
`remediation-spec-producer-rules.md` (attach `TASK-260910-2ohnjo_spec-patch_rev3.patch`).

## The one correction (medium)
`tools/validate.py` (≈ line 4530, the surfacing-row consumer) splits a row with
`row.split(" ")` and requires exactly seven tokens, so a valid argument
containing a space (e.g. `"hello world"`, permitted by
`schemas/v1/agent-mcp-v1.schema.json` string args) breaks the row into eight
tokens and the consumer rejects a correctly rendered positive row. Parse the
JSON-array columns structurally — tokenize the row by the closed column
layout, decoding the `args` and `env_names` JSON arrays with a JSON parser so
spaces and escapes inside strings are preserved — while still rejecting
missing, reordered or extra columns. Add a positive surfacing vector whose
`args` contains a string with a space (and one with an escaped quote if the
grammar allows it), keep both previously requested mutant refusals in
`tools/test_validate.py`, and refresh the vector integrity pins
(`make regenerate` / manifest). Do not narrow the normative argument grammar
in `protocol/environments.md` §2.3 to fit the parser.

## Validation and handoff
`make validate` and `make regenerate-check` (repo venv on PATH,
`set -o pipefail`, quote outputs and exit codes). Attach
`TASK-260910-2ohnjo_spec-patch_rev3.patch`, update
`TASK-260910-2ohnjo_evidence.md` (closure with file:line, transcripts), tick
the checklist items you satisfy, then
`task-board handoff TASK-260910-2ohnjo --role doc-writer`.
