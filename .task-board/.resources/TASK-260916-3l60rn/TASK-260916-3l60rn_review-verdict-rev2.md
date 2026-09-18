# TASK-260916-3l60rn — revision 2 review verdict

Verdict: **accepted**. Reviewed CR-TASK-260916-3l60rn-2, revision 2. This is acceptance for producer integration, not a claim that the revision is merged.

Candidate: curator-spec Story worktree `/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260916-wgt8vz/worktree`, HEAD `e8b53a003256433761cebce6080d6a955d777f25`. Read the campaign rules, producer and rework briefs, producer evidence and both spec patches, and audit E6/Appendix B. No implementation changes made by this reviewer. No LOGBOOK edit: the campaign explicitly prohibits one; review findings are persisted here.

## Per-item review

All line references below are relative to the spec worktree.

| Requirement | Evidence and assessment |
| --- | --- |
| Patch equals candidate | Attached rev2 patch and independently captured `git diff HEAD` are byte-identical (`cmp`, exit 0). Both stable patch IDs: `24bd6262fce17d670fa75cbf9b77fdfcf770aff1`. HEAD, not moved origin/main, is the review base. |
| Decision amendment | `decisions/0012-context-packages-and-semver-locks.md:127`: “Amendment (2026-09-18, E6)”; original passages retained with amendment markers. The entire Decision file is byte-identical between revisions 1 and 2. |
| Git-only MCP, direct refusal | `protocol/environments.md:526`: “An MCP declaration package MUST resolve from a `git` source”; line 530: “resolution MUST refuse”; line 533: “never admitted, never warned-through”. §2.1 line 456 tables `mcp_declaration_path_source_refused`, naming package and declaration. |
| E1 rationale | `protocol/environments.md:533`: “A `path` source has no canonical identity for the allowlist above and is never verified”; §3 line 641 repeats the rationale for system-module admission. No signer allowlist expansion. |
| E2 direct naming preserved | `protocol/environments.md:636`: “The direct-naming rule above applies as is”; transitive path lock explicitly described as unproducible through git-only `requires`. The transitive vector is an admission-model control, not a claim such a dependency resolves. |
| Boundary and timing | `protocol/environments.md:738`: “every `path` source the lock names”; line 743: “manager MUST verify the five boundary checks”. Ownership, permissions/DACL, containment, regular types, link safety are closed and enumerated. Timing is every resolve and again under mutation lock before mutating operations. §3 line 645 expressly says “before any materialization”. |
| Failure and pin baseline | `protocol/environments.md:762`: “no pin is recomputed against the live directory”; line 765: “entry-class `environment_store_untrusted`”; line 767: “resolve emits no fragment, status is non-current”; line 769: “There is no rebuild”. Operator repairs the directory; stored snapshot remains pinned. |
| R1 dry-run consistency | §4 line 772: “failure is not in the entry-rebuild branch”; line 775: “nothing — never `would-rebuild-untrusted-store`”. §10.1 line 2479 reports `environment_store_untrusted` with no rebuild; §10.4 line 2699 tables exactly that branch. Store-entry rules remain distinct. Pinned failed-directory, intact control, and forbidden-would-rebuild negative cover the correction. |
| Content-independent overlay/import boundary | §3 line 648: “The boundary check is on the directory, not on the content class”; §6 line 1126 applies the contract to path overlays; §9.6 line 2376: “The import is a `path` root”. Both no-system refusals and both trusted controls are present. |
| R2 adversarial proof | Production `validate.main()` rejects the reviewer’s `if not trusted and case["carries_system_modules"]` narrowing. The regression suite also fails `test_published_vector_passes` under that mutant. Three internally consistent passing replacements under required refusal names are rejected through main(), including both no-system branches and dry-run. See transcript. |
| Diagnostics and closed sets | Exact new spelling is `mcp_declaration_path_source_refused` in Decision, §2.1/§2.2/§13, CHANGELOG, vector diagnostics/cases and validator constant (`tools/validate.py:6434`). Reuses `environment_store_untrusted` and `context_system_module_transitive`; no new knob, lock key or wire/schema field. §1.1 and §3.1 existing domain tables remain unchanged; store refusal is tabled in §8.5 and §10.4 and referenced from §3/§4. §12.1/§12.2 need no new row because no new knob exists. |
| Posture and direct rollout | `protocol/environments.md:2884` extends the store-trust row to the source directory, “naming the failing check and the path”; line 781: extension “is direct (impact row \"E6\": under the hood)”. No warning profile is required by this brief. |
| Vectors and registration | New family has 5 MCP cases, 14 boundary cases and 3 dry-run cases. Coverage: 3/3 path MCP origins; 5/5 isolated boundary-check failures; 2/2 required no-system refusal contexts with trusted controls; 3/3 dry-run cases. Manifest line 4288 registers it. §13 line 3153 lists the surfaces. `tools/validate.py:7835` calls the E6 gate from main. This proves specification-vector semantics, not manager runtime behavior, which belongs to TASK-260916-yvxbs1. |
| Rule 7 | Input pins cover kind, role/origin, directness, content class, policy, and isolated failed check. `tools/test_validate.py:2452` onward tests no-system replacements; line 2573 tests the healed dry-run branch. Independently replayed replacements through main(), not only the helper. |
| Rev1 preservation | Reconstructed both attached patches against HEAD in scratch directories. Only R1/R2 protocol paragraphs, E6 CHANGELOG summary, new cases/tests/gate additions, manifest digest and rc.9 pins differ. All 5 MCP and 11 boundary cases from rev1 are unchanged; Decision and unrelated spec content unchanged. |
| Scope | Eight intended paths: CHANGELOG, Decision 0012, environments protocol, new vector, manifest, rc.9 pins, validator and validator tests. Validator code is explicitly required conformance tooling, not manager implementation. No schemas, manager profile, proposals 0014–0018 or LOGBOOK changed. |
| Empty curator CR | Independently inspected `git diff 640a9df19ba94295cce059562101599996761f4b 57d2578ea9d1644d8a8931fdcdfbcb8ce4e15545`: zero paths. This is correct: this leaf delivers a curator-spec revision in the separately designated spec worktree and attached spec patch. It must not change curator manager implementation. Acceptance is bound to those spec artifacts; the empty curator delta alone would not satisfy the task. |

## Independent validation

Shell: zsh with `set -o pipefail`. Python: `/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin/python3`; Go: installed toolchain. The `make validate` recipe was executed as its three constituent gates with unittest discovery partitioned into bounded batches, as required by the headless time limit. No producer test result substitutes for these reviewer runs. There is no claim that one monolithic `make validate` process completed.

Commands:

```
PATH="/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin:$PATH" python3 tools/validate.py
# validated 62 schemas and 1096 vector files
# exit 0

PATH="/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin:$PATH" python3 -B -m unittest discover -s tools -p test_validate.py -k PathKindAdmission
# Ran 36 tests in 0.024s; OK; exit 0

go test ./tools/...
# ok github.com/relux-works/curator-spec/tools/generate-vectors 7.150s
# exit 0
```

The full discovery batch driver discovers `tools/test_*.py`, recursively flattens the unittest suite, and partitions by test ID into `test_release_gate.*`, `*.WriteNofollowVectorTests.*`, and all remaining tests. All 443 discovered tests are included exactly once across these batches. Focused E6 tests above are an additional replay.


### Regeneration and byte preservation

Ran `make regenerate-check` on a byte copy of the candidate in `/tmp/e6-review-r2/regen`. First the inherited HEAD-based index reported the intended uncommitted patch (make exit 2, underlying diff exit 1); this is not generated drift. Comparing all tracked candidate files before and after generation yielded zero byte changes. Then initialized a scratch-only Git index holding the candidate's conformance/release bytes (no commit, no candidate index changes) and ran the same target:

```
go run ./tools/generate-vectors -root .
git diff --exit-code -- conformance/v1 release/1.0.0-rc.5.json release/1.0.0-rc.6.json release/1.0.0-rc.7.json release/1.0.0-rc.8.json release/1.0.0-rc.9.json
EXIT: 0
```

Direct `git cat-file --batch` comparison against HEAD (not `git archive`, which honors export-subst on its test fixture) proves 1095/1095 pre-existing conformance files byte-identical in the frozen candidate. No generated data was accepted solely on producer evidence.

### Adversarial probes

The attached production function was compiled in memory with only this replacement; candidate code was not edited:

```
if not trusted:
    expected_diag = E6_DIAG_UNTRUSTED
# becomes
if not trusted and case["carries_system_modules"]:
    expected_diag = E6_DIAG_UNTRUSTED
```

Invoked `validate.main()` using that function, then bound the test module to the same mutated validator instance and ran all 36 E6 tests. Output:

```
validation failed: path-kind case path-overlay-no-system-world-writable-untrusted: diagnostic is not the §4/§3 rule (None)
NARROWED main exit: 1
ERROR: test_published_vector_passes
Ran 36 tests in 0.758s
FAILED (errors=1)
```

Restored the normal function in memory. Substituted each refusal's input and reported result with an internally consistent passing scenario while retaining its required name. The on-disk corpus and manifest remained unchanged; `load_json` supplied only the changed E6 vector to main's real semantic gate:

```
REPLACEMENT path-overlay-no-system-world-writable-untrusted main exit: 1
# this branch must fail permissions
REPLACEMENT path-import-no-system-wrong-ownership-untrusted main exit: 1
# this branch must fail ownership
REPLACEMENT path-overlay-dry-run-untrusted-no-rebuild main exit: 1
# this branch must fail permissions
```

Replacement coverage: 3/3 newly required refusal scenarios rejected. The normal focused suite covers the remaining inventory and replacement shapes. Bounds: these probes validate the spec gate and its production call site; they do not exercise a manager implementation.

### Validation-run anomaly

The existing `WriteNofollowVectorTests.run_main_with_vector` temporarily rewrites its fixture, manifest, and rc.9 pins, restoring them in `finally`. Running a copying release-test batch concurrently produced one harness-interference failure: `test_rejects_changed_historical_rc7_release_metadata` expected its historical-metadata rejection but saw `conformance manifest hash is stale: vectors/environments-write-nofollow.json`. The 12-test nofollow batch subsequently passed (248.737s, exit 0), and `git diff HEAD` again matched the attached rev2 patch byte-for-byte. The affected release test was rerun after fixture restoration; its result is below. This transient failure is disclosed rather than characterized as an initially green suite.

### Final results

```
python3 -B /tmp/e6-review-r2/batch.py other
DISCOVERED 443 BATCH other 399
Ran 399 tests in 439.983s
OK
exit 0

python3 -B /tmp/e6-review-r2/batch.py nofollow
DISCOVERED 443 BATCH nofollow 12
Ran 12 tests in 248.737s
OK
exit 0

python3 -B /tmp/e6-review-r2/batch.py release
DISCOVERED 443 BATCH release 32
Ran 32 tests in 280.456s
31 passed; 1 interference failure described above
exit 1

python3 -B -m unittest discover -s tools -p test_release_gate.py -k test_rejects_changed_historical_rc7_release_metadata
Ran 1 test in 33.341s
OK
exit 0

python3 -B tools/validate.py  # final restored candidate
validated 62 schemas and 1096 vector files
VALIDATOR EXIT: 0
```

Final measured coverage: 443/443 discovered tests passed across the bounded runs and the one justified isolated rerun; no skipped tests and no producer-only validation accepted. Go gate and candidate-relative regeneration also exit 0. The inherited uncommitted candidate prevents a meaningful HEAD-relative regeneration diff, hence the explicit scratch candidate baseline above.

Final `git diff HEAD` SHA-256 and attached rev2 patch SHA-256 both:
`035d84d4bef144de7be6aca2389bae16b006dff8a875be5713ff0b223305d1df`.
The original eight-path footprint is restored; no reviewer repository edits remain. The run is not goal-bound (`task-board spawn goal "$TASK_BOARD_RUN_ID"`: “Active Goal: none”).

## Findings and routing

R1 and R2 are resolved. No remaining rework finding. Existing manager behavior is not claimed to implement this new specification; its implementation is explicitly the follow-on task. Merge/integration is still pending and belongs to the bound producer role.

Record acceptance using `accept_cr(TASK-260916-3l60rn, revision=2, evidence=TASK-260916-3l60rn_review-verdict-rev2.md)`. No `commit_ack`, no direct `done` transition.
