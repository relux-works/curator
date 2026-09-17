# TASK-260910-2ohnjo — reviewer verdict, CR revision 4

Verdict: accepted. Accept CR revision 4 and route to integrating; no merge or done claim.

Reviewed the curator-spec Story candidate at base `07e2b41892bf24c33b859072b2b920fe2d0a17fe`, attached spec patch rev3, producer evidence, previous verdict and the campaign/producer/rework/review briefs. Both S4 audit findings were read. Candidate files were not edited. Generation and mutation checks used `/tmp/s4-review4-snapshot`, copied from tracked candidate files with its own scratch index; no commit, branch change, push or LOGBOOK edit. The campaign expressly forbids LOGBOOK edits; this artifact records the review result.

The curator CR repository delta is empty, independently confirmed from its base/candidate OIDs. This is the right outcome in curator: this leaf is a curator-spec revision, not curator implementation. Its actual deliverable is the eleven-file external spec candidate plus attached spec patch, whose stable patch-id equals the live candidate: `3fcb1de6fe40f2753141b2f8e69fd3ea2c2a3be1`. Review uses the recorded base, not advanced origin/main (`0da4020`). Rebase and landing review remain separate; acceptance does not assert the merged acceptance criterion has already happened.

## Per-item findings

All paths below are relative to the curator-spec Story worktree.

| Requirement | Evidence / assessment |
| --- | --- |
| Opt-in default and explicit null | PASS — `protocol/environments.md:451`: “an absent knob is the empty list”; “an explicit `null` keeps meaning unbounded”. Reserved exclusions retained. |
| Empty package allowlist | PASS — `protocol/environments.md:471`: install, update and status “MUST emit” `mcp_package_allowlist_empty`, stating “every declaration package in the closure is admitted”. |
| Exact surfacing output | PASS — `protocol/environments.md:480`: “one surfacing row per MCP declaration package”; table at :488 closes package, version, transport, command, args, env_names; JSON strings and exactly one LF. Parser now preserves spaces within strings, as required by round 3; normative grammar unchanged. |
| Install/update sequence | PASS — `protocol/environments.md:480`, :1649, :1701: “after the audit gate passes and before the lock is published”; before materialization. Two order vectors at `conformance/v1/vectors/environments-env-passthrough.json:319` cover install and update. |
| Warn-first | PASS — `protocol/environments.md:2193`: impact “S4 passthrough default”; labelled `s4-warn` warning release and `s4-enforce` flip release; migration hint “list the named variables to keep passing them after the flip”; “MUST ship `s4-warn` before `s4-enforce`”. Explicit lists continue bounding under both profiles; explicit null remains unbounded. |
| Diagnostic closed sets | PASS — `protocol/environments.md:407`, :1999, :2233 and status :2320 consistently admit `mcp_package_allowlist_empty`, `mcp_env_passthrough_unlisted`, `mcp_env_passthrough_dropped`. Same spellings in vectors, validator constants and CHANGELOG. No new knob/lock key or open-ended extension added. |
| Status posture | PASS — `protocol/environments.md:2303`: warning, “active S4 profile” and “effective `passable_env_names`”, repeated declaration rows. :2318 warnings never make a row non-current. |
| Knob defaults / locks / schema | PASS — `protocol/environments.md:2358`: `[]`; :2359 “empty (permits all, warned)”. :2384 retains both lockable keys. `schemas/v1/manager-config-v2.schema.json:454` default `[]`; `profiles/manager.md:2429` matches. |
| Generator and schema cases | PASS — `tools/generate-vectors/manager_config.go:22`: `[]any{}`; :313 adds four generated knob schema cases. Regeneration gate independently green. Frozen schema-1 family unchanged. |
| Semantic vectors | PASS — 25/25 cases connected to consumer: 7 defaults, 6 allowlist, 6 surfacing, 4 schema, 2 order. `tools/validate.py:4612` recomputes semantics; `main():5000` calls consumer from the real validation entry point. `conformance/v1/manifest.json:4116` registers vector. Positive/negative default, explicit null/list, refusal, warning/silence, column grammar and schema branches inspected. |
| Contradictory fixture | CLOSED — vector :28 `non-empty-allowlist-silent` is positive with null diagnostic; :39 `non-empty-allowlist-warns` is negative with the emitted diagnostic. |
| Round 3 parser correction | CLOSED — `tools/validate.py:4523` uses `JSONDecoder().raw_decode` at exact ordered prefixes, compact reserialization and end-of-input check. Spaces, quotes and delimiter-like string contents preserved; missing, reordered and extra columns rejected. New vector :262 / :284 covers space and escaped quote. `tools/test_validate.py:1883` onward adds six tests; previous two mutants at :1737 / :1747 retained. |
| Revision scope | PASS — diff of rev2/rev3 patches changes only parser and its case inventory, two positive vectors, six tests, manifest/release pins. No normative changes or narrowing, no unrelated edits. All eleven candidate paths are docs/schema/vectors/pins or spec tooling expressly authorized by rework rev2. No product implementation, tags or proposals 0014–0018 touched. |
| CHANGELOG / future boundary | PASS — `CHANGELOG.md:99` names “S4”, both rollout steps and diagnostics. `protocol/environments.md:2220` leaves closed interpreter contract to a later revision. |

No remaining rework findings. Review coverage concerns spec conformance artifacts, not actual curator launch behavior; implementation is the separate task. Existing passing items were re-inspected; no producer test result substitutes for this review's gates.

## Independent validation

Shell: zsh, `set -o pipefail`. Exact command from candidate worktree:

```sh
PATH="/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin:$PATH" make validate
```

```text
python3 tools/validate.py
validated 60 schemas and 1048 vector files
python3 -B -m unittest discover -s tools -p 'test_*.py'
..............................................................................................................................................................................................................................................................
----------------------------------------------------------------------
Ran 254 tests in 348.319s

OK
go test ./tools/...
ok  	github.com/relux-works/curator-spec/tools/generate-vectors	0.831s
make_validate_exit=0
```

All three make validate gates independently rerun to completion in this review; no prior evidence substituted.

Generation command used identical PATH and `set -o pipefail`, from the scratch candidate copy, because this gate writes generated files:

```sh
PATH="/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin:$PATH" make regenerate-check
```

```text
go run ./tools/generate-vectors -root .
git diff --exit-code -- conformance/v1 release/1.0.0-rc.5.json release/1.0.0-rc.6.json release/1.0.0-rc.7.json release/1.0.0-rc.8.json release/1.0.0-rc.9.json
regenerate_check_exit=0
```

Independent full-entry-point mutant checks ran `python3 /tmp/s4-review4-mutants.py` in the scratch copy. Each refreshed vector and manifest/release integrity pins before `python3 tools/validate.py`; rejected 2/2:

```text
unlisted-passed exit=1  validation failed: env-passthrough case s4-enforce-absent-drops-all: passed/dropped are not the s4-enforce rule ([]/['FIGMA_API_KEY'])

invalid-output exit=1  validation failed: env-passthrough case single-stdio-declaration: expected_bytes are stale

```

Independent parser admission/refusal probes passed 5/5: exact spaced args accepted, exact escaped-quote args accepted, missing column rejected, reordered columns rejected, extra column rejected. The baseline vector suite drives these positives through the real validation entry point. Unit tests additionally cover delimiter-like strings, separator padding and narrowed explicit allowlists. This bounds the claimed coverage to those cases, not all possible input strings.


`git diff 07e2b41 --check` passed (exit 0). An initial whitespace-check invocation accidentally used the curator cwd and could not resolve the spec base; discarded and rerun successfully in the spec worktree. No candidate edits resulted.

`task-board spawn goal "$TASK_BOARD_RUN_ID"` returned “Active Goal: none (run is not goal-bound)”. No commit_ack. Verdict evidence is attached before accept_cr; acceptance routes to integrating, leaving integration to the tracked producer.
