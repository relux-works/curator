# TASK-261021-krcy1c results r2

## Delivery

- Repository: curator-spec.
- Branch: `cip02-ready-for-acceptance`.
- Signed head: `f202f41ba4718b82cdfefac28c8dfaa9af3a7de5` (`git log -1 --format=%G?`: `G`).
- Appended one commit to reviewed r1 `752307c6a825c07256521e818fef25ba5ae78fc7`; ordinary push succeeded (exit 0), remote tracking readback equals the new head. No force push, PR, main push or board status change.
- Only changed file: `cips/CIP-0002-project-context-in-managed-launches.md` (+96/-3).

## F1 delta

Context items now inherit context-package manifest schema 2, context lock schema 4, validation, explicit module selection, selectors/classes, weights/topological ordering and exact environments rendering. The payload binds both snapshot and retained lock digests; all pinned closure manifests and bytes are retained before approval. Root/module changes require fresh approval. Cross-item additive ordering is explicit.

Bare instruction files without a manifest and native import dialects are deferred. Missing or outside-snapshot modules, duplicate paths/roles/JSON keys, invalid module bytes and native imports refuse. Nested skill/MCP/native-setting activation cannot inherit authority from context approval. Three future vector families describe selection binding, package validation and closure rendering, with nearby positives and narrowing controls. All other candidate sections, dated owner decisions, Draft status and remaining owner questions are unchanged.

## Validation and claims

Evidence directory: `~/Werkstatt/curator-dev4/curator-spec/.temp/TASK-261021-krcy1c/r2/` (abbreviated E below). Python is the existing task-local Python 3.14.7/jsonschema 4.25.1 environment; Go is the configured CURATOR_GO (1.25.5). No installation was needed.

| Check / command | Exit | Claim and evidence |
|---|---:|---|
| `python -B E/check-docs.py` | 0 | Only this CIP changes; original operator input, dated decisions, operative disposition, questions and launch-route sections are byte-preserved; six new link targets/anchors resolve; added prose hygiene passes. `docs.log`. |
| `python -B E/context-witness.py` | 0 | Actual schema-2/schema-4 readers, content-pin framing and environment rendering oracle accept two packages with identical two Markdown bodies but different explicit selections. Snapshot/item/approval/output digests differ; exact expected rendered bytes independently checked and unselected body absent. `context-witness.log`, `.json`, `.py`. |
| Same witness: manifest-exclusion narrowing control | 0 | Intentionally omitting the manifest from inventory and content pin creates identical approval identity for different output; demonstrates the original underbinding failure. This is a design witness, not a production project checker mutation. |
| Same witness: inherited negative controls | 0 | Real path schema rejects parent/absolute entries; real byte checker rejects CRLF/missing or extra trailing LF beside valid LF; duplicate JSON class key rejects with intended diagnostic. Missing-entry, duplicate path/role and inside/outside native-import cases are enumerated design witnesses only. |
| `python -B tools/validate.py` in final-head normal checkout | 0 | Existing spec validation remains green: 74 schemas, 1294 vector files. `validator-exit-check.log`, `validation-exits.json`. |
| All discovered `python -B -m unittest` cases, split by E/suite-groups.py | 0 each | 879/879 tests in 73 groups; same unique discovery list, none omitted or duplicated. Every group records command-selected IDs, count, exit and duration; each log has the matching unittest total and OK. `suite-groups.json`, `python-01..73.{log,json}`, `validation-summary.json`. |
| `env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test ./tools/...` | 0 | Full existing Go tool test package passes, including vector generation regression coverage. `go.log`, `go-exit-check.log`, `validation-exits.json`. |
| `python tools/verify_skillfile_sources_independence.py` | 0 | Existing independence gate passes: 89/89 schema references, 34/34 cited clauses; the one existing allowlisted exception remains. `prose.log`. |
| `git diff --check HEAD^ HEAD`; signature; branch ancestry; final checkout status | 0 | No whitespace defects, signature G, r1 remote head remains ancestor, source and both validation clones clean. Final clone HEADs both equal delivered head. `validation-summary.json`. |

The three recipes printed by `make -n validate` were executed completely at the exact delivered head in disposable normal-.git clones under `.val/`, with `TMPDIR=/tmp/cw-dev4` and inherited GIT_DIR/GIT_WORK_TREE removed. Python discovery was split per the standing long-suite instruction; this is recipe-equivalent validation, not a claimed monolithic `make validate` invocation. Two clones prevented fixture mutations from overlapping. The initial 40-test release-gate group took 1040.57s, exceeding the estimate; remaining groups were reduced to 10 tests. Both clones have been removed after clean HEAD/status verification. Go caches remained in this Werkstatt.

Operational notes: a relative local-clone fetch path failed and was corrected to an absolute path before final-head checkout. A summary assertion initially assumed OK was the last log line; existing fixtures print after unittest's summary, so the summary now checks the exact OK line and matching test count. All actual test-group exit codes were zero. Validator/Go completion was repeated once to persist explicit exit-code records after a tool-output truncation; both exit 0. These are harness observations, not product failures.

## Limits and open decisions

This is a Draft-only contract change. No runtime implementation, schemas or checked-in executable project vectors were added. The new project vector families specify future acceptance work; this run does not qualify native adapters, import suppression or project admission rejection. Existing executable inherited checks and the bounded selection/rendering witness cover only their stated claims.

The unchanged later owner questions are profile root/system replacement and private memory retention after explicit opt-in. Leaf 1 remains additive with memory off. Engineering qualification gates remain open as already documented in the candidate. No new owner decision is needed for F1.
