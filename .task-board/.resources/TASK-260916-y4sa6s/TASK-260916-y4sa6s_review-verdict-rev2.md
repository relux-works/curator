# TASK-260916-y4sa6s review verdict — revision 2

Verdict: **changes_requested**. Route: **to-dev**. R2–R7 are closed; R1 needs one focused conformance correction described below. No candidate files or original index modified; no commit or acknowledgement supplied.

## Candidate identity and scope

Reviewed curator-spec Story worktree `/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260916-ioemse/worktree`, HEAD `23dafa7`, attached revision-2 patch and producer evidence, campaign/producer/rework/review briefs, prior verdict, audit E1 at docs/security-audit-2026-09.md:222 and Appendix B:423. Applied project-management lifecycle and negative-evidence skill reference. Architecture-diagrams is not relevant.

Both `git diff origin/main | git patch-id --stable` and the downloaded patch yield **0f23fc0825477b0ffd18b20dd7bf2d4426a9356a**. New paths already have intent-to-add entries, so no index mutation was needed. `git diff --check` exits 0.

Important baseline distinction: current main and origin/main are `dced9b8317e0e8af79edf2d0539b32bd22b6c85b`, not the producer evidence's claimed `23dafa7`. Thus the requested full diff against origin/main includes reversals of later upstream changes. Patch identity passes, but integration must apply the task delta relative to its recorded story base, not blindly restore the old tree over current main. Actual working changes relative to HEAD remain the scoped E1 specification, schema, vector, generator/validator and tests, plus generated manifest/release hash. No proposal or manager implementation edit.

The exact curator CR base `aa46ecd80ad0b83853586454723ea99fb76977a9` to candidate `45aae0e7ba8fdc58a42eb553ded0726f3c68b8cf` diff is empty; curator `git status --short` is also empty. **An empty curator repository delta is correct for this leaf:** its normative deliverable lives in the separate curator-spec worktree and attached spec patch. This is not the reason for changes requested. R5's LOGBOOK edit is gone. No LOGBOOK was written by this reviewer; findings persist in this artifact and board notes under the campaign prohibition.

## Required correction — R1 remains partially open (medium)

The prose and URL/selector/reordering cases are corrected, but `tools/validate.py:5274` requires every MCP snapshot to have exactly all six keys (`transport`, `command`, `args`, `env_names`, `url`, `environments`). At :5296/:5301 it requires both optional arrays to be present. The fixture representation therefore cannot preserve field absence from the declaration that §9.2 now requires comparing byte-for-byte. `protocol/environments.md:1841` explicitly says “no field is narrowed out” and “absent ... differs from present.” The unchanged real MCP schema at `schemas/v1/agent-mcp-v1.schema.json:15`/:27 makes `env_names` and `environments` optional; environments.md:497 says “absent means every adapter.”

Independent reproductions, without editing files:

- A stdio server with no `environments` and one with `environments: ["codex_cli"]` both validate through the real Draft202012Validator/schema registry, and have different CCJ-1 bytes. Removing the selector from an existing delta fixture is rejected by the six-key shape guard before delta recomputation.
- A stdio server with absent `env_names` and the same server with `env_names: []` both validate. Canonical comparison returns true. A narrowed comparator that inserts `[]` for absent optional arrays returns false.
- That narrowed comparator **survives the entire published E1 vector gate**. The suite has **0 cases representing absent optional declaration fields**; all MCP snapshots carry both arrays. This is a measured blind spot, not evidence that the implemented `ccj1_bytes` comparison itself normalizes them.

Correct the snapshot representation/checker to preserve actual optional-field presence (prefer real server objects validated against the existing MCP schema, without padding irrelevant transport fields with null). Add absent-to-present and present-to-absent optional-field cases, including absent `env_names` versus explicit `[]`, with A warning + migration hint, B refusal, and flagged acceptance. Keep URL-only, selector-only and array-order cases. Add a narrowing test that normalizes absent optional fields and fails on the new cases. Do not change the settled canonical-byte rule or frozen MCP schema to accommodate the fixture abstraction. Update manifest, evidence and patch, then hand off for review.

## Requirement-by-requirement assessment

All file references below are relative to the spec worktree; quotes are exact excerpts.

| Requirement | Evidence and assessment |
|---|---|
| Dated Decision amendment, retained history, §8 pointer | decisions/0012-context-packages-and-semver-locks.md:69 “Amendment (2026-09-17, E1)”; :100 “(Decision 8)”; :604 amendment marker. Pass. |
| Signer rule in amendment | Decision:92 “before the candidate enters the lock — either suffices”. Pass. |
| latest residual / strict tags | Decision:122 “`latest` stays `*`”; environments.md:254 “policy covers a *moved* tag, not a *new* one”. Pass. |
| §1.3 lock invariant | environments.md:194 “The lock is a record, not a signature”; resolution pointer added, context-lock-v1 schema diff empty. Pass. |
| §1.4 verification | environments.md:328 “before the candidate enters the lock”; :331 “signature of the commit it peels to MUST verify”; :332 “only the commit signature can satisfy the check”. Pass. |
| Fail closed / path / no downgrade | environments.md:335 “MUST leave the old lock in place”; :348 “MUST NOT silently select a lower candidate”; :351 “are never verified”. Pass. |
| Three resolution diagnostics | environments.md:109–111 and :339–345 spell `context_source_unsigned`, `context_source_signer_rejected`, `context_source_signers_missing`, matching validator constants and vectors. Pass. |
| Delta before publish, closed line forms | environments.md:1787 “resolved-version delta of the candidate lock against the old lock”; :1813–1815 added/removed/moved grammar includes version and pin. Pass. |
| System inventory trigger | environments.md:1832 “(`path`, `environments` selector, bytes)”; “admission under section 3 does not narrow the trigger”. Pass. |
| R1 canonical MCP trigger | environments.md:1837 “the CCJ-1 bytes”; :1841 “Any byte difference”; tools/validate.py:5325 uses `ccj1_bytes`. Prose and present-field comparisons pass; absent-field conformance gap above. |
| R1 URL / selector / reorder | vector:985/1007 `changed-mcp-url` and confirmed; :1029/1051 selector and confirmed; :1073 `mcp-env-names-reorder-triggers`. Negative tests test_validate.py:2291–2322. Pass these branches. |
| Two rollout revisions + migration hint | environments.md:1848 “Revision A (warning release)”; :1853 “Revision B (flip release)”; :1861 “MUST ship revision A before revision B”. A hint names B refusal and flag; e1_revision_outcomes:5380 pins it. Pass. |
| R2 reinstall flag | environments.md:1744 “`profile install` accepts”; :1869 “identical per-invocation semantics”; cli/curator.md:30 includes `[--confirm-system-delta]`; vector:1371/1395 flag/no-flag reinstall cases. Pass. |
| Per-run / --all / no config consent | environments.md:1862 “no configuration knob may pre-confirm it”; :1863 “confirms every profile of the run”; two --all vectors. Pass. |
| Update diagnostic table | environments.md:2147–2148 has `profile_update_system_delta` with migration hint and `profile_update_confirmation_required`. Pass. |
| §12 signer posture | environments.md:2565 “signer-verification posture per lock member's source”; three states and machine require value; :2599 “`unknown` when it cannot”. Pass, including settled unknown behavior. |
| R7 confirmation posture | environments.md:2606 “`A-warning`”; :2608 “`B-flip`”; behavior names warn/proceed versus refusal unless flagged, no preconfirmation. Two vector rows :1420/1427. Pass; B-flip consistently derives from the rollout's “flip release” label. |
| §12.1 knobs/defaults | environments.md:2657/2658 `source_signers.<source>` default `{}` and `require_source_signers` default `false`; schemas and generated defaults match. Pass. |
| Closed entry shape | environments.md:2671 “`type` exactly `ssh` or `gpg`”; :2678 prohibits cross-fields; manager schema:352 closed union. Pass. |
| R3 fingerprint | manager schema:372–374 minLength/maxLength 40 plus uppercase pattern; system schema:58 references it. New newline invalid cases in both configurations, real-entry regression test test_validate.py:3311. Pass. |
| R6 SSH identity | environments.md:2679 “key type plus its base64 key material”; :2680 “trailing comment is not part”; oracle:5024 compares first two tokens; vector:419/447 accepts changed comment and refuses different material. Pass. |
| §12.2 fleet lock/direction | environments.md:2699 admits both keys; :2713 “only to `true`”; :2717 machine list “is ignored” for system-named source; unnamed source takes machine list. profiles/manager.md:57/67/73 agrees. Pass. |
| Both config schemas and lockable set | manager schema:602/:616 knobs; system schema:26/:27 locked names and :58/:59 properties (`enum: [true]`). No extra configuration preconfirmation. Pass. |
| R4 revision forgery | validator:5094 forbids tag signature for revision; test_validate.py:2354 replays forgery. Reviewer separately replayed it through published gate and observed rejection. Pass. |
| R5 empty curator delta | Exact CR object diff and current curator status both empty, no LOGBOOK change. Pass. |
| §13 conformance + manifest | environments.md:2780 names `vectors/environments-source-signers.json`; :2802 requires warning/migration hint; manifest registers file and new schema cases. Pass presence, R1 coverage caveat. |
| Vector branches | Inspected 57/57 E1 cases: 19 verification, 6 merge, 5 posture, 21 delta, 2 --all, 2 reinstall, 2 confirmation posture. Gates recompute expectations; R1 optional-presence branch is not represented (0). |
| Schema cases retain purpose | 101/101 changed preexisting config cases are identical semantically after removing only the two E1 knobs/lock entries. Pass under explicit regeneration exception. |
| Existing vectors | 29/30 existing files byte-identical against story HEAD; sole changed file manager-config-v2.json adds knobs/new cases under exception. Other apparent origin/main reversals are upstream baseline advancement. |
| CHANGELOG | CHANGELOG.md:10 “E1: per-source signer allowlist”; :32–43 names both revisions, canonical trigger, reinstall, identity, residual. Pass. |
| Outcome resources | Producer evidence and full revision-2 patch downloaded through resource CLI; this task-scoped verdict is attached before routing. |

## Validation and evidence bounds

All checks below were independently rerun; no producer result substitutes for reviewer execution. Spec-policy vectors use declared signature-validity observations, not cryptographic Git fixtures: downstream manager enforcement remains outside this leaf. The E1 gate is registered at tools/validate.py:6377 in the real validation entry.

Full make validate: shell `/bin/zsh`, `set -o pipefail`, requested venv on PATH, `PYTHONDONTWRITEBYTECODE=1` to avoid new worktree cache files. Three-gate transcript appended after completion.

Generator: copied tracked candidate files into a disposable directory, initialized a local index and staged the candidate as the comparison baseline (no commits/branches). Ran literal `make regenerate-check` there. This is the equivalent baselined proof requested by revision 2, not a claim that comparing the original uncommitted tree to HEAD is green. Output:

```
go run ./tools/generate-vectors -root .
git diff --exit-code -- conformance/v1 release/1.0.0-rc.5.json release/1.0.0-rc.6.json release/1.0.0-rc.7.json release/1.0.0-rc.8.json release/1.0.0-rc.9.json
exit=0
```

No human-only decision or external blocker is needed. This is ordinary conformance rework. Preserve all passing closures and correct the remaining R1 fixture bound.

## Independent adversarial transcript

```text
absent selector real declaration schema errors: []
present selector real declaration schema errors: []
canonical byte difference: True
published gate rejects before delta recomputation: source-signers case changed-mcp-selector: an mcp snapshot is { transport, command, args, env_names, url, environments }
absent-selector snapshots in delta fixtures: 0
R4 forgery rejected: source-signers case revision-selection-verifies-commit: a revision selection carries no tag, so tag-only signature evidence cannot satisfy the check
NARROWED canonical comparator (absent optional array -> []) survives complete published E1 gate
Both absent-env_names and explicit-empty declarations validate against real MCP schema
Canonical trigger: True normalized comparator trigger: False
```

Final spec diff SHA-256 and attached patch SHA-256 both remain `4ea5349088b529c7715cd0bef4dc16ff6ee4449110f3a6f77e52f1409313047e`. Before verdict, `task-board spawn goal "$TASK_BOARD_RUN_ID"` returned `Active Goal: none (run is not goal-bound)`.

## Completed independent validation

```bash
set -o pipefail
PATH="/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin:$PATH" PYTHONDONTWRITEBYTECODE=1 make validate 2>&1 | tee /tmp/y4sa6s-review-validation.log
```

```text
python3 tools/validate.py
validated 60 schemas and 1089 vector files
python3 -B -m unittest discover -s tools -p 'test_*.py'
...................................................................................................................................................................................................................................................................................................................................................
----------------------------------------------------------------------
Ran 339 tests in 459.481s

OK
go test ./tools/...
ok  	github.com/relux-works/curator-spec/tools/generate-vectors	10.155s
EXIT_CODE=0
```

All three gates passed independently: 60 schemas / 1089 vector files, 339 Python tests, Go tooling tests. Passing corpus checks do not close the demonstrated R1 absent-field blind spot. No process is left running.
