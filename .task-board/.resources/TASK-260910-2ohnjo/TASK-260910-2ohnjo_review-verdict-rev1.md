# TASK-260910-2ohnjo — review verdict revision 1

Verdict: **changes_requested**. Route to `to-dev`. Do not accept CR revision 1.

Reviewed the exact curator-spec Story worktree at `/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260910-1lf0m5/worktree`, against origin/main (07e2b41), the campaign rules, producer brief, producer evidence and attached spec patch, and both audits' S4 finding. No candidate files or index were modified. No commits, branches, pushes, or implementation edits were made. Scratch verification used `/tmp/TASK-260910-2ohnjo-review-snapshot` with a temporary Git index holding the candidate baseline; no scratch commit was made.

The curator CR's empty repository delta is expected and is not itself a defect: this leaf delivers a curator-spec revision in a different repository, through its Story worktree and attached spec patch. There is a real eight-file spec delta. Acceptance would concern that external candidate, not claim a curator implementation change or a merged revision. The external candidate has the defects below, so the empty CR is not accepted.

## Required corrections

1. **High — generated source and deliverables disagree.** `tools/generate-vectors/manager_config.go:22` still says `"passable_env_names": nil`, despite `schemas/v1/manager-config-v2.schema.json:454` and environments §12.1 requiring `[]`. Independent `make regenerate-check` exits 2 against a candidate-baseline scratch index: regeneration restores the old manager-config vector bytes (digest `c3eb7e6e...`), removes all four new schema cases, and changes manifest/release pins. This is actual generator drift, not the expected diff from an uncommitted candidate to main. Resolve the generator/default/case ownership follow-up before presenting a commit-ready revision, regenerate and attach the resulting patch. Route any needed code-owned tooling work to its producer; the reviewer must not edit implementation or approve red generation as an unrelated follow-up.
2. **High — contradictory normative update order.** `protocol/environments.md:481-484` says surfacing MUST happen “before the lock is published or any surface is (re-)materialized”; `:1714-1717` says “After publishing the new lock” for update. A single update cannot satisfy both. Specify one consistent sequence in §2.3 and §9.2, preserving the required pre-materialization output, and cover update order in a vector.
3. **High — S4 behavior vectors are not semantically checked.** The 20 entries in `vectors/environments-env-passthrough.json` are parsed and integrity-hashed by `tools/validate.py:814-844`; no dedicated consumer is wired into the validation entry point (`:4601-4610`). Existing `validate_environment_vectors` reads only `vectors/environments.json` (`:4303-4307`). In a scratch copy I changed the enforcing absent-knob case to pass `FIGMA_API_KEY` and drop nothing, and replaced the first expected surfacing bytes with `INVALID\n` while retaining its original expected length/hash. After refreshing only file/manifest/release integrity pins, `python3 tools/validate.py` still exited 0, “validated 60 schemas and 1048 vector files”. Thus the entry point does not catch either tested violation (0/2 mutations rejected); this is not evidence of launch behavior. Wire executable semantic verification into the spec validation tooling through the appropriate producer and demonstrate rejection of these mutations and narrowed-list bypasses. The separate four manager-config schema cases do have an existing semantic consumer; do not conflate that with coverage of the new file.
4. **Medium — contradictory negative fixture.** `vectors/environments-env-passthrough.json:27-35` marks `non-empty-allowlist-silent` as `conforming: false` while `diagnostic: null` describes the correct silent behavior; its reason instead describes incorrectly emitting the warning. Make the observation match the negative verdict (or make silence a positive and add an emitting negative), then assert it through the consumer.

## Brief conformance review

All file references below are relative to curator-spec; quotes are from the reviewed candidate.

| Requirement | Evidence / assessment |
|---|---|
| Patch equals exact candidate | Stable patch IDs of attached spec patch and read-only reconstruction of `git diff origin/main` plus the untracked new vector both equal `43fbf25a3cd7b06ce2c1859a1fa3b6a01c9e278b`. No `git add -N` was needed in the candidate. PASS. |
| §2.2 opt-in default, explicit null | `protocol/environments.md:451-455`: “an absent knob is the empty list” and “an explicit `null` keeps meaning unbounded”. PASS. Existing reserved-name exclusion remains in force. |
| Empty package allowlist warning | `protocol/environments.md:472-475`: install, update and status “MUST emit `mcp_package_allowlist_empty`, stating that every declaration package in the closure is admitted.” PASS. Empty still permits all. |
| Surfacing exact closed columns / stdio command and args | `protocol/environments.md:478-509`: “one surfacing row per MCP declaration package”, ascending package-name byte order; closed columns `package`, `version`, `transport`, `command`, `args`, `env_names`; one LF; compact JSON arrays, HTTP `command=- args=[]`. Present, but update timing FAILS finding 2. |
| Install and update integration | `protocol/environments.md:1649-1652` requires install output before lock publication; `:1714-1717` requires update output after publication. Contradiction with §2.3; finding 2. |
| §10.3 two labelled rollout steps | `protocol/environments.md:2193-2220`: “Profile `s4-warn` (the warning release)” retains absent/unbounded, warns with variable names and knob plus “list the named variables to keep passing them after the flip”; “Profile `s4-enforce` (the flip release...)” drops unlisted names; “A manager MUST ship `s4-warn` before `s4-enforce`”. PASS in normative intent. Explicit lists remain bounds in warn profile; explicit null remains unbounded. |
| Diagnostics admission and identical spelling | `protocol/environments.md:407-409`, `:1999`, `:2233-2234`, `:2319-2321`: exact diagnostics `mcp_package_allowlist_empty`, `mcp_env_passthrough_unlisted`, `mcp_env_passthrough_dropped`, also used identically in new vectors and CHANGELOG. No new knob or lock-key spelling. PASS. |
| §12 env status posture | `protocol/environments.md:2302-2307`: warning row, “active S4 profile ... with the effective `passable_env_names`”, repeated declaration rows for each current scope. `:2319-2322` warnings never make a row non-current. PASS. |
| §12.1 defaults / §12.2 closed lockable set | `protocol/environments.md:2358-2359`: default `[]`; “empty (permits all, warned)”. `:2381-2384` already explicitly admits `passable_env_names` and `mcp_package_allowlist`. Schema default changed to `[]` at `schemas/v1/manager-config-v2.schema.json:454`. PASS; generator disagrees (finding 1). |
| Vectors + manifest | `conformance/v1/manifest.json:4115` registers new file. New file has 7 default-resolution, 5 allowlist, 4 surfacing and 4 schema cases (20 total); separate manager-config file adds 4 real schema cases at `:1489` onward. Both rollout profiles, explicit null, narrowed list, allowlist rejection, output ordering, missing/reordered columns appear. FAIL semantic exercise / one contradictory case, findings 3–4. |
| Exact output bytes | Independently recomputed both positive output lengths and hashes: 126 and 159 bytes, both match declared SHA-256. Their declared output is correct for these examples; validator does not enforce it (finding 3). |
| CHANGELOG S4 and rollout | `CHANGELOG.md:99-120`: “S4 (MCP env passthrough and declaration surfacing)” names `s4-warn` then `s4-enforce`, migration hint, diagnostics, default and surfacing. PASS. |
| Consistency / scope | Changes limited to environments, manager cross-reference, schema, vectors, manifest, CHANGELOG and derived rc.9 manifest pins. `profiles/manager.md:2203-2206,:2429` updated to avoid stale default/warning. Derived `release/1.0.0-rc.9.json` pin changes are necessary manifest consistency, not a release publication. No tools/implementation or proposals 0014–0018 changed. PASS scope, FAIL generator consistency. Manager implementation remains separately scoped and is not claimed to implement the new revision. |
| Future interpreter contract | `protocol/environments.md:2222-2224`: “is a later revision, not this one”. PASS settled scope. |
| Validation / ready for integration | See independent transcript below. Generation red; not commit-ready. No claim that AC “merged” has occurred. |

## Independent validation transcript

Candidate shell: `/bin/bash`, `set -o pipefail`, workdir as above.
Command: `PATH="/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin:$PATH" make validate`.
The literal Makefile gate was started in a tracked foreground exec session and awaited to completion before verdict attachment. No producer test result is substituted for this run.

```text
python3 tools/validate.py
validated 60 schemas and 1048 vector files
python3 -B -m unittest discover -s tools -p 'test_*.py'
...................................................................................................................................................................................................................................
----------------------------------------------------------------------
Ran 227 tests in 180.117s

OK
go test ./tools/...
ok  	github.com/relux-works/curator-spec/tools/generate-vectors	(cached)
exit 0
```

`make validate` completed successfully: 60 schemas, 1048 vector files, 227 Python tests and the Go tools suite (cached). This does not override the red regenerate-check or the measured semantic blind spot.

Generation check: `/bin/bash`, `set -o pipefail`, `make regenerate-check` in the byte-copied scratch candidate with its baseline staged in a temporary index (the Makefile mutates generated files, so running it in the review worktree would violate read-only review).

```text
go run ./tools/generate-vectors -root .
git diff --exit-code -- conformance/v1 release/1.0.0-rc.5.json release/1.0.0-rc.6.json release/1.0.0-rc.7.json release/1.0.0-rc.8.json release/1.0.0-rc.9.json
[manager-config-v2 defaults revert [] -> null; four added cases removed; manifest and rc.9 pins change]
make: *** [regenerate-check] Error 1
exit 2
```

`git diff --check` on original candidate: exit 0. Original positive surfacing hash checks: 2/2 pass. Mutation probe described in finding 3: validator exit 0 (0/2 violations caught). No real manager launch behavior was run; manager implementation is outside this spec review. This bounds what the passing gates establish.

## Lifecycle and notes

`task-board spawn goal "$TASK_BOARD_RUN_ID"` reported “Active Goal: none (run is not goal-bound)”. No directives were recorded. Exactly this verdict resource will be attached before routing to `to-dev`. Ordinary producer/tooling rework is required, not an external blocker or a human-only decision. Findings are also recorded in task notes; no LOGBOOK.md is edited because the campaign explicitly forbids those edits. Another producer revision and independent reviewer cycle are required.
