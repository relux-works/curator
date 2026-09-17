# TASK-260910-2ohnjo — review verdict revision 2

Verdict: **changes_requested**. Route to `to-dev`; do not accept CR revision 2.

Reviewed the curator-spec Story worktree against recorded base `07e2b41`, the campaign and rework briefs, attached producer evidence, rev1 verdict, and rev1/rev2 patches. The candidate was not edited. Scratch generation/mutation work used `/tmp/2ohnjo-review2-snapshot` and a copied Git index. No commits, pushes, branches, or LOGBOOK edits were made.

The curator CR's empty repository delta is expected: this leaf delivers a different repository's spec revision. The external candidate is an eleven-file delta, including spec-validation tooling explicitly authorized by rework rev2. Empty curator delta is not the reason for rejection; no curator implementation change or merged AC is claimed.

## Required correction

**Medium — semantic consumer rejects valid argument strings containing spaces.** `tools/validate.py:4530` splits a surfacing row with `row.split(" ")`, then requires exactly seven tokens. `s4_surfacing_row` emits compact JSON arrays, which preserve spaces *inside* string values. An allowed argument such as `"hello world"` therefore creates an eighth token and the parser returns `None`. The positive-case consumer then rejects its own correctly rendered row as violating the closed columns. Environments §2.3 requires JSON strings in args, not a ban on spaces inside strings; JSON serialization with no separator whitespace does not remove string content. This is an unintended narrowing introduced by rev2.

Parse the JSON array columns structurally (preserving spaces and escapes inside strings), while continuing to reject missing, reordered or extra columns. Add a positive vector/test containing spaces in an argument and retain both requested mutant refusals. Validate through `tools/validate.py` with refreshed integrity pins. Do not narrow the normative argument grammar to fit the parser.

The defect was reproduced through the real `python3 tools/validate.py` entry point in the scratch copy with all integrity pins refreshed. `schemas/v1/agent-mcp-v1.schema.json:19` permits string arguments (maxLength 8192); `hello world` is valid. See the transcript below.

## Per-item review

| Requirement | Evidence and assessment |
| --- | --- |
| Patch equals candidate | `git diff 07e2b41 \| git patch-id --stable` and attached rev2 patch both yield `78c86592a0024537c22842b25b1ef7c918e33538`. PASS. `main`/`origin/main` moved to `f544a01`; using that moving base would review unrelated landed changes. |
| §2.2 opt-in and explicit null | `protocol/environments.md:451`: “an absent knob is the empty list”; “an explicit `null` keeps meaning unbounded”. PASS. Reserved names remain excluded. |
| Empty allowlist warning | `protocol/environments.md:471`: install, update and status “MUST emit `mcp_package_allowlist_empty`”, stating “every declaration package in the closure is admitted”. PASS. |
| Closed surfacing columns | `protocol/environments.md:480`: “one surfacing row per MCP declaration package”; closed columns `package`, `version`, `transport`, `command`, `args`, `env_names`; compact JSON and exactly one LF. Normative PASS; new consumer narrowing described above. |
| Install/update sequence | `protocol/environments.md:480`, `:1649`, `:1701`: “after the audit gate passes and before the lock is published” with surfacing now update step 3 and publication step 4. Previous contradiction removed. Two `surfacing_order_cases` at vector `:275` cover install and update. PASS source review. |
| Two rollout steps | `protocol/environments.md:2193`: “Profile `s4-warn` (the warning release)” retains unbounded absence and migration hint “list the named variables to keep passing them after the flip”; “Profile `s4-enforce` (the flip release...)” drops unlisted names; “A manager MUST ship `s4-warn` before `s4-enforce`”. PASS. |
| Diagnostic closed sets | `protocol/environments.md:407`, `:1999`, `:2233` vicinity: `mcp_package_allowlist_empty`, `mcp_env_passthrough_unlisted`, `mcp_env_passthrough_dropped` are admitted and consistently spelled in prose, tables, vectors and changelog. No new knob or lock key. PASS. |
| Status posture | `protocol/environments.md:2302` vicinity: empty-allowlist warning, “active S4 profile” with “effective `passable_env_names`”, repeated declaration rows; warnings do not make a row non-current. PASS. |
| Defaults and lockable set | §12.1: `passable_env_names` default `[]`; package allowlist “empty (permits all, warned)”. §12.2 `:2381` explicitly includes both keys. Schema default `[]`. PASS. |
| Generator drift correction | `tools/generate-vectors/manager_config.go:22` emits `[]any{}`; four knob schema cases now generated around `:313`. CLOSED: independent `make regenerate-check` exits 0 on the restored candidate snapshot. |
| Semantic consumer | `tools/validate.py:4583` recomputes defaults, passed/dropped sets, diagnostics, surfacing bytes and schema cases; `main():4971` calls it. `tools/test_validate.py:1705` adds 21 tests including the two review mutants, narrowed lists and negative-observation repair. Previous uncalled-consumer defect corrected in source, but valid-space narrowing remains. |
| Contradictory fixture | New vector `:28` makes `non-empty-allowlist-silent` positive with diagnostic null; `:39` adds negative `non-empty-allowlist-warns` with emitted warning and conforming false. Consumer checks both. PASS. |
| Vector coverage | 23 cases: 7 default, 6 allowlist, 4 surfacing, 4 schema, 2 sequence. Both rollout profiles, explicit null/list, rejected passthrough, package refusal, missing/reordered columns and install/update order appear. Registered in manifest; four manager-config schema cases also present. Source coverage 23/23 cases connected to the consumer; independently executed negative mutations rejected 2/2, but the valid-space admission control failed (0/1 accepted). This is spec-vector validation coverage, not proof of manager runtime behavior. |
| CHANGELOG / future scope | `CHANGELOG.md:99` entry “S4 (MCP env passthrough and declaration surfacing)” names both rollout steps. Environments §10.3 leaves closed interpreter contract to a later revision. PASS. |
| Architecture and scope | Eleven paths: normative docs, schema, vectors/manifest, derived release pins, generator, validator and tests. Tooling changes explicitly required by rework. No product implementation, tags, proposal 0014–0018 content or unrelated source edits observed. Manager cross-references updated. |

## Independent verification state

The following were launched in `/bin/bash` with `set -o pipefail` and repo venv first on PATH:

```
PATH="/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin:$PATH" make validate
PATH="/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin:$PATH" make regenerate-check
```

The generation gate runs only in the scratch copy with a copied index, preserving read-only candidate review. A temporary host execution stall delayed all commands, including shell-only probes. After recovery, the initial scratch regeneration overlapped the mutation script and saw mutated fixture pins; that contaminated result was discarded. After the mutation script finished and restored the original vector, regeneration was rerun sequentially:

```text
go run ./tools/generate-vectors -root .
git diff --exit-code -- conformance/v1 release/1.0.0-rc.5.json release/1.0.0-rc.6.json release/1.0.0-rc.7.json release/1.0.0-rc.8.json release/1.0.0-rc.9.json
regen_check_exit=0
```

Scratch mutation command: `python3 /tmp/2ohnjo-mutants.py`. Each case regenerates manifest/release pins with `go run ./tools/generate-vectors -root .` then invokes `python3 tools/validate.py`; surfacing cases also refresh expected byte length and SHA-256. Completed output:

```text
unlisted-passed exit=1  validation failed: env-passthrough case s4-enforce-absent-drops-all: passed/dropped are not the s4-enforce rule ([]/['FIGMA_API_KEY'])

invalid-output exit=1  validation failed: env-passthrough case single-stdio-declaration: expected_bytes are stale

valid-space-argument exit=1  validation failed: env-passthrough case single-stdio-declaration: recomputed row 'mcp-declaration figma-devmode 1.2.0 stdio command=npx args=["-y","figma-developer-mcp","--stdio","hello world"] env_names=["FIGMA_API_KEY"]' violates the closed columns

```

Thus the two required review mutants are rejected 2/2, closing the old uncalled-validator gap. The valid-space admission control fails, proving the remaining narrowing. The scratch script restored the vector; sequential regeneration restored derived pins. No candidate files were changed; its stable patch ID remained `78c86592a0024537c22842b25b1ef7c918e33538` after these checks. `git diff 07e2b41 --check` exits 0.

The rev1/rev2 patch comparison confirms the previously passing normative/default/schema/changelog changes are retained; changes are confined to the four requested correction areas and derived pins. No unrelated regression was observed beyond the new parser narrowing. Product launch behavior is not tested in this spec-only review.

The literal `make validate` completed independently with exit 0. The long elapsed Python test time includes the host stall; the exec session was retained and awaited to completion, never left running at turn end. No producer result was substituted. Go reports cached success, as shown.

```text
python3 tools/validate.py
validated 60 schemas and 1048 vector files
python3 -B -m unittest discover -s tools -p 'test_*.py'
........................................................................................................................................................................................................................................................
----------------------------------------------------------------------
Ran 248 tests in 1255.393s

OK
go test ./tools/...
ok  	github.com/relux-works/curator-spec/tools/generate-vectors	(cached)
make_validate_exit=0
```

Both baseline gates are green; they do not invalidate the separate failing valid-space admission control. The remaining correction is the parser and its missing positive coverage, not generator drift or an unresolved validation result.

`task-board spawn goal "$TASK_BOARD_RUN_ID"` returned `Active Goal: none (run is not goal-bound)` before verdict preparation. No commit_ack is supplied. Acceptance and integration are not claimed. Ordinary validator rework routes to `to-dev`, not blocked.
