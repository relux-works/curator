# Producer brief: environments schemas and conformance vectors

## Setup
- Base: fresh `origin/main` of `~/Developer/ReluxWorks/curator-spec` — MUST be `c3b29b1f7f37829fd4d0c50b2023efa2feb4c615` or later (contains protocol/environments.md). Fetch and verify.
- Worktree `~/Developer/ReluxWorks/.worktrees/curator-spec-env-schemas`, new branch `draft/environments-schemas`. Work only there; shell tooling; signed commits.

## Task
Deliver the machine surfaces `protocol/environments.md` promises, following the repo's existing schema/vector conventions exactly (study how agent-skill/skillfile schemas and conformance/v1 are wired: naming, manifest, tools/validate.py, tools/generate-vectors, Makefile targets):
1. JSON Schemas under `schemas/v1/`: `profilefile-v1.schema.json`, `context-manifest-v1.schema.json` (module entries: path, environments selector, class root|system), `agent-environment-marker-v1.schema.json`, `launch-env-fragment-v1.schema.json`. Strict (additionalProperties per repo convention), matching the normative prose exactly — where prose and your schema disagree, the prose wins; file a note, do not silently diverge.
2. Conformance vectors under `conformance/v1/`: positive + negative schema-cases for each schema; determinism vectors producing exact expected bytes for: generation header (7-line grammar), monolithic join, chapter composition (incl. empty chapter), zero-applicable-modules output, referenced-form layout paths, managed opencode.json CCJ-1 bytes (+ trailing LF rule), system-prompt output (no header). Deterministic regeneration: run the generator twice, byte-identical.
3. Wire into the validate/CI flow the same way existing vectors are; `make validate` green.
4. Update the vector manifest and anything the repo's conventions require; do NOT touch CHANGELOG, decisions/, protocol prose (except a trivial §13 pointer ONLY if the repo convention demands file references — prefer none).

## Deliverables
Signed commit(s); board resource `env-schemas-notes.md` (file list, any prose ambiguity found, twice-run regeneration evidence); handoff to-review.

## Do not
Push; tag; touch profiles/manager.md or cli/curator.md (a sibling story owns them); mark done.
