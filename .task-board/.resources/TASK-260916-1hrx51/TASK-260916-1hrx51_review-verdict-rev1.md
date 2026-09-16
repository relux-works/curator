# TASK-260916-1hrx51 — review verdict, revision 1

Verdict: **changes_requested**. Route to `to-dev`; do not accept CR-TASK-260916-1hrx51-1 revision 1.

## Findings and concrete corrections

1. **R1 — normative closed lock set contradicts the new schema (required).** `profiles/manager.md:51–62` says `locked` “contains only” the enumerated keys, ending with “`environments.require_current_profile`, and `environments.isolation`”, followed by “No other `environments` knob is lockable or carried by the system file.” It omits `environments.transitive_system_modules`. In contrast, `protocol/environments.md:2344–2355` adds that key and permits only `error`, and `schemas/v1/system-config-v2.schema.json:23,51–53` accepts it. A manager following §1 would reject precisely the strict-machine configuration this revision introduces. Add the key to manager §1's closed enumeration and mirror the error-only direction, keeping waivers un-lockable. Extend a consistency test to catch drift of this second normative enumeration; the current gate compares environments §12.2 with the schema but misses manager §1.
2. **R2 — required RFC 2119 formulation missing (required).** The new admission block `protocol/environments.md:510–535` and direction rule at `2352–2355` contain no RFC 2119 requirement keywords, although task DoD item 1 explicitly requires them. The current text says “is skipped”, “is the resolution error”, and “no lock is written or changed”. Express the implementer obligations explicitly with MUST/MUST NOT: admitted-only output, drop plus warning, strict refusal and unchanged lock, and error-only locking. Keep the settled behavior unchanged.

These are routine specification rework, not an external blocker or a decision requiring a human. Reattach a refreshed spec patch/evidence and hand off for another reviewer cycle.

## Exact candidate and empty curator delta

Reviewed curator-spec worktree `/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260916-2d9coh/worktree`, against `origin/main`, and the board's `TASK-260916-1hrx51_spec-patch_rev1.patch`. Read the campaign rules, producer brief, evidence and patch, then audit E2 and Appendix B. The curator-repository CR delta is intentionally empty: this leaf delivers external curator-spec text/schema/conformance changes, not curator implementation. That is appropriate to this leaf and is not itself a finding; it does not prove external integration or justify acceptance while R1/R2 remain.

Used a temporary Git index to include untracked candidate files without modifying its real index. Both complete candidate diff and attached patch have stable patch ID:

`d0b6f56569cb506c7dfda0ba7fd7c7af02858e69`

No candidate code, text, index, branch or commits were changed by review. Regeneration ran only in a disposable byte copy with its own temporary Git baseline. No product implementation or proposals 0014–0018 were changed by the producer; Go/Python changes are spec conformance tooling. The rc.9 JSON delta only refreshes two generated manifest pins, not a release/tag.

## Brief/DoD inspection

| Requirement | Evidence (candidate file:line, quoted text) | Result |
| --- | --- | --- |
| Direct/transitive admission | `protocol/environments.md:510`: “root itself”, “active overlay”, “root's or an active overlay's `requires.contexts` entry”; `:515`: “transitive packages admitted by a `system_module_waivers` entry” | Behavior matches settled decision; R2 wording correction required. §2 package shape need not change. |
| Drop warning and byte exclusion | `protocol/environments.md:521`: “skipped at materialization with the warning `context_system_module_dropped`”; `:523`: “materialized bytes are exactly the admitted modules' bytes” | Present, table at :843; vector byte comparison passes. |
| Strict resolution error | `protocol/environments.md:526`: “resolution error `context_system_module_transitive`”; :529 “no lock is written or changed” | Present in §3.1 :545 and §5.7 :844. |
| Always-warn finding | `protocol/environments.md:532`: “unaffected by admission”; :1632 “always-warn, never-blocking” | Preserved over all members; no audit-class widening. |
| Fragment and resume drift | `protocol/environments.md:798`: “fragment's `system_prompt` presence”; :800 “`works.relux.curator.system-modules` extension key”; :801 “`ax` resume still refuses on drift” | Present; §10.2 :2140 follows admitted set. Launcher implementation not tested here. |
| Knobs/defaults/waiver scope | `protocol/environments.md:2315`: `transitive_system_modules`, `drop`, `error`, default `drop`; :2316 `system_module_waivers`, list of `{ package, reason }`, empty; :2327 “portable identifier” naming a lock member | Exact spellings match machine schema :494/:501 and vectors. Closed waiver object and required fields in schema :307. |
| Lock direction and closed set | `protocol/environments.md:2353`: “lockable only in the direction of `error`”; :2354 “`system_module_waivers` is not lockable” | Environments/schema agree; manager §1 contradicts them (R1). |
| Rollout | `protocol/environments.md:803`: “default `drop` policy is non-breaking”; :806 “no warn-first split applies”; :807 “`error` is opt-in strictness” | Matches explicit brief exemption; two rollout profiles are not required for E2. |
| Status posture | `protocol/environments.md:2259`: “`transitive_system_modules` value with every dropped system module by package and path”; :2274 drop warning never makes row non-current | Mirrored at `profiles/manager.md:2518,2532`. |
| Conformance surface | `protocol/environments.md:2389`: “system-module admission cases”; vectors `environments.json:1066,1182,1310,1439,1576` | All 5/5 delivered branch cases inspected; details below. |
| Changelog | `CHANGELOG.md:36`: “E2 direct-only `class: system` modules”; :49 “no warn-first split applies” | Finding, knobs, diagnostics, rollout and posture described. |
| Evidence/scope | Task evidence and matching spec patch attached; only spec/conformance tooling and generated pins changed | Appropriate external spec delivery. Evidence says eleven schema cases; actual index adds seven cases, plus four expected byte files. Correct that count when refreshing evidence. |

## Vector and regeneration checks

Opened every new materialization case (5/5): direct admits sysmid/sysroot; drop skips sysleaf with named warning; error refuses sysleaf/90-system.md and writes no file; waiver admits sysleaf; overlay edge makes sysleaf direct. The direct and transitive-drop expected outputs are each 40 bytes with identical SHA-256 `537ebd5eed3dcaa2dec64e79ccbbdc325cdb35e49652250494ea873ba46d4a3a`. Waived and overlay cases include the leaf and are 61 bytes. The five cases are in the manifest-bound vector file; four expected files are registered at manifest lines 140–152. No existing expected byte output changed.

Measured schema preservation: 911/911 existing index entries retain their schema/validity meaning; 88/88 modified existing fixture objects differ only by the two new environment knobs or the new locked key. Seven new schema cases cover valid waiver, invalid mode, invalid package grammar, missing reason, unknown waiver field, forbidden system `drop` direction, and invalid system mode. All were read through the candidate index and generator diff; regeneration is exact.

Generator changes are scoped to admission, the two knobs, direction and related fixtures. Python validator independently computes the admission set and expected bytes. Negative tests reject missing dropped records, absent strict refusal, false refusal under drop, widened enum, changed default and broadened system direction. Review does not claim curator runtime resolution/install/update coverage: this is a spec-only task; runtime implementation belongs to TASK-260916-55g9dg. Existing tests also do not establish preservation of the manager §1 closed list (R1).

## Independent validation

Shell: zsh, `set -o pipefail`; candidate worktree above. Exact command:

```sh
PATH="/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin:$PATH" make validate
```

```text
python3 tools/validate.py
validated 60 schemas and 1058 vector files
python3 -B -m unittest discover -s tools -p 'test_*.py'
...........................................................................................................................................................................................................................................
----------------------------------------------------------------------
Ran 235 tests in 206.503s

OK
go test ./tools/...
ok  	github.com/relux-works/curator-spec/tools/generate-vectors	(cached)
EXIT=0
```

Full `make validate` exit 0. Go reported cached test results; the Go command was rerun, but this does not claim an uncached execution. Candidate and regenerated disposable copy have zero byte differences across tracked/untracked nonignored files.

Disposable-copy command: `make regenerate-check`, after byte-copying all tracked/untracked nonignored candidate files and establishing an isolated temporary Git baseline. This is not a commit in either project worktree.

```text
go run ./tools/generate-vectors -root .
git diff --exit-code -- conformance/v1 release/1.0.0-rc.5.json release/1.0.0-rc.6.json release/1.0.0-rc.7.json release/1.0.0-rc.8.json release/1.0.0-rc.9.json
EXIT=0
```

`git diff --check origin/main`: exit 0. Existing expected outputs: zero changed paths.

All validation above was rerun by this reviewer; no producer test result substitutes for it. No implementation runtime test was run.

## Lifecycle / logbook

`task-board spawn goal "$TASK_BOARD_RUN_ID"` reports “Active Goal: none (run is not goal-bound)”. Campaign instructions prohibit LOGBOOK.md edits, and no logbook CLI is installed; findings are persisted in this task-scoped outcome and task notes instead. Candidate remains untouched. Verdict is changes_requested, routed to `to-dev` after this evidence is attached.
