# Review verdict — TASK-260910-2qtiho, revision 3

Verdict: accepted.
Reviewed CR-TASK-260910-2qtiho-3, spec base e8b53a003256433761cebce6080d6a955d777f25.

## Candidate and scope

The attached spec patch and `git diff HEAD` are byte-identical. Patch ID:
`d34f8329fc83f9b83a5d0b53f784bc0d7e4fcbea`; SHA256:
`94d759b8cc8565b7c0c285304f00dc85098905eb20f9dfb3d628c5ef7efda84b`.
Comparison is against HEAD, not moving main. All 33/33 pre-existing vector files match HEAD byte for byte.

The curator CR delta between `640a9df19ba94295cce059562101599996761f4b` and
`57d2578ea9d1644d8a8931fdcdfbcb8ce4e15545` is empty, as independently verified.
This is the correct repository outcome: this leaf authors curator-spec, whose candidate is the separate story worktree plus the task-scoped spec patch. No curator implementation change belongs to this task. Acceptance authorizes integration; it does not claim these uncommitted spec changes are already merged.

## Per-item review

| Requirement | Evidence and result |
|---|---|
| Closed knob and lock direction | PASS. profiles/manager.md:46: “one closed top-level knob, `security_posture`”; :49: “`permissive` and `hardened`”; :80: “lockable only in the direction of `hardened`”. manager-config-v2.schema.json:46 has the two-value enum and revision-A default; system-config-v2.schema.json:18 admits the lock key and :43 restricts its value to hardened. Frozen v1 schemas unchanged. |
| Effective defaults, precedence and three refusals | PASS. profiles/manager.md:1097 table supplies strict audit/registry, nonempty source/MCP allowlists, null refusal, error transitive policy, required signers and unreachable-registry error. :1106: “Precedence is lock, then explicit machine value, then the profile default”; :1109–1113 names the three exceptions and applies them however the value arrived. |
| Two rollout revisions and warning | PASS. profiles/manager.md:1122: “Revision A (this release)”; :1125: warning “exactly once per operation”; :1127 gives the migration hint naming the knob; :1130: “Revision B (a later release).” Default flip kept separate. |
| Twelve gates on both commands | PASS. profiles/manager.md:1390: “`curator status` and `env status` carry the same closed posture inventory”; :1403–1416 lists all twelve gates; :1418: “same gates in the same order with the same values and provenance”. protocol/environments.md:2833: “twelve posture rows ... identical in gate name, order, value, and provenance”. |
| Schema-1 bound and --check | PASS. profiles/manager.md:1420 bounds schema-1 to header plus four manager rows and no env posture section; :1424 identifies hardened contradictions. Explicit permitted opt-outs remain current (:1430). environments.md:2864 also declares contradictions non-current. |
| Provenance correction | PASS. manager.md:1397: “`profile` ... `explicit` ... `lock` ... `shipped`”; environments.md:2843 and CHANGELOG.md:30 agree. tools/validate.py:7503 is exactly that tuple. All 421/421 vector row values use this vocabulary; all 17/17 scenario machine/system/operation inputs are unchanged. `system.locked` configuration fields are preserved. Old spellings survive only as descriptive case names, configuration vocabulary, or intentional negative-test inputs, not admitted provenance values. |
| S3 residual and notice | PASS. registry.md:135 and SECURITY.md:519 explicitly state “revocation is network-dependent”; registry.md:145 requires notice naming artifacts, :152 closes the diagnostic to `registry_unreachable_during_install`, warning permissive/error hardened. Registry :259 retains seven-day grace and :262 cross-references the residual. No underlying cache-policy change. |
| Environment interactions and diagnostics | PASS. environments.md:523 escalates empty MCP allowlists with declarations; :2611 refuses explicit null at install/update/launch; diagnostic tables :454, :2330, :2643 carry exact names. :2948 references profile interaction while retaining existing knob defaults. §12.2 existing environment locks remain unchanged; the new top-level lock belongs to manager §1. |
| SECURITY and CHANGELOG | PASS. SECURITY.md:40 recommends hardened posture and names rollout; :519 names residual. CHANGELOG.md:12 begins “S1/S3”, names revisions, diagnostics, provenance and residual. |
| Schema cases, generator, manifest | PASS. Seven added schema cases cover valid postures, invalid type/value and forbidden permissive system direction. Generator and its assertions updated; manifest registers security-posture.json; regenerated pins verified. |
| Vector branch coverage and validator | PASS; see independent transcript below. Seventeen cases cover defaults, lock/explicit precedence, three refusals, warning/error notices, schema-1, both status commands and --check. validate.main dispatches validate_security_posture_vectors (tools/validate.py:8397); :8257 recomputes values and :8306 checks scenario pins. |
| Revision-3 scope | PASS. Six authored files changed from rev2 only for the rename and two rejection tests, plus manifest/release digest updates. Sixteen of 24 patch sections unchanged. No implementation, LOGBOOK, proposals or unrelated edits. |

## Validation method and anomaly

Shell: zsh with `set -o pipefail`; Python from `/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin`.
The three Makefile validate legs were independently executed, with unittest discovery split into bounded shards because one uninterrupted run can exceed the headless command limit. No producer test result substitutes for the independent rerun.

An initial concurrent run exposed that existing tests temporarily rewrite source fixtures. Its results were discarded. Processes were interrupted; exact candidate bytes were verified against the attached patch before continuing. Independent shard reruns use separate temporary candidate copies. No candidate change is part of this review. Regeneration and disk mutation probes also use temporary copies; the candidate index is untouched. No LOGBOOK edit was made, as the campaign explicitly forbids it; this artifact records the anomaly.

The review is bounded to normative text, schemas and structural conformance validation. It does not prove downstream manager execution; manager.md:1155 explicitly assigns that separately.

## Independent transcript

The Makefile's three validation legs all exit 0 (full unittest discovery partitioned, 428/428 tests). They were not run as one uninterrupted positive `make validate` invocation; the bounded sequential-call requirement necessitated splitting its Python leg. Each shard had its own isolated candidate, so their temporary fixture writes could not interfere.

Commands (venv bin prepended to PATH, `set -o pipefail`):
```sh
python3 tools/validate.py
python3 -B /tmp/qtiho-review3/shards3.py 0
python3 -B /tmp/qtiho-review3/shards3.py 1
python3 -B /tmp/qtiho-review3/shards3.py 2
python3 -B /tmp/qtiho-review3/shards3.py fullheavy
go test ./tools/...
make regenerate-check
```
Validator and Go tests ran in the story worktree. Shards ran in `/tmp/qtiho-review3/run{0,1,2,heavy}` reconstructed from HEAD plus the exact patch (Git archive export-subst fixture restored to raw blob bytes). Regeneration proof ran in an exact temporary candidate with a temporary Git index recording the candidate, avoiding index writes in the review worktree.

### validate.log

```text
validated 62 schemas and 1103 vector files
```

Exit 0.

### isolated0.log

```text
Tests 143
...............................................................................................................................................
----------------------------------------------------------------------
Ran 143 tests in 269.446s

OK
```

Exit 0.

### isolated1.log

```text
Tests 142
..............................................................................................................................................
----------------------------------------------------------------------
Ran 142 tests in 290.544s

OK
```

Exit 0.

### isolated2.log

```text
Tests 142
..............................................................................................................................................
----------------------------------------------------------------------
Ran 142 tests in 258.307s

OK
```

Exit 0.

### isolatedheavy.log

```text
Tests 1
.
----------------------------------------------------------------------
Ran 1 test in 298.085s

OK
```

Exit 0.

### clean-regenerate.log

```text
go run ./tools/generate-vectors -root .
git diff --exit-code -- conformance/v1 release/1.0.0-rc.5.json release/1.0.0-rc.6.json release/1.0.0-rc.7.json release/1.0.0-rc.8.json release/1.0.0-rc.9.json
```

Exit 0.

### isolatedprobe.log

```text
validation failed: security-posture case refusal-mcp-allowlist-empty-with-declarations must leave machine.security_posture absent
validation failed: security-posture case revision-A-default-permissive-status must leave machine.audit absent
validation failed: security-posture case revision-B-default-hardened-flip-install must carry machine.schema_version == 2
Whole-case substitution sweep: 272/272 refused; main-entry replays: 3/3 refused
```

Exit 0.

### mutant.log

```text
python3 tools/validate.py
validation failed: security-posture case locked-value-beats-explicit curator status rows do not follow their inputs
make: *** [validate] Error 1
```

Expected `make validate` exit 2; probe accepted only because the intended row-check diagnostic fired.

Go output (exit 0):
```text
ok github.com/relux-works/curator-spec/tools/generate-vectors 2.416s
```
The disk mutant's preceding `make regenerate` completed successfully; it updated digest evidence so rejection proves the semantic gate. The superseded spellings' 8/8 negative subcases passed in the discovery shards. The three former substitution holes were rejected through main (3/3), and the exhaustive distinct-case sweep refused 272/272. These ratios cover the named vector inputs and structural validator, not runtime manager behavior.

Final `cmp` between `git diff HEAD` and the attached patch exits 0; `git diff HEAD --check` exits 0. All 33/33 pre-existing vectors are byte-identical. No extra files or code changes remain in either story worktree.

Run goal query: `Active Goal: none (run is not goal-bound)`. Checklist was inspected and is fully checked. No directives recorded.


## Revision-3 negative coverage

The two new test methods at tools/test_validate.py:3151 and :3173 each exercise four independently reset mutations (vocabulary, profile_source, sources, status-row source): 8/8 superseded-spelling locations are required to fail. The full-suite execution below includes both methods. The separate disk probe changes the `audit-mode` curator-status row in `locked-value-beats-explicit` from `lock` to `locked`, regenerates hashes, then runs the real `make validate` entry point. It fails with the posture-row diagnostic, not a stale digest failure.

Replay script: task outcome `TASK-260910-2qtiho_review-probe-rev2.py` executed unchanged against revision 3. Bounded discovery runner: attached `TASK-260910-2qtiho_review-test-shards-rev3.py`; fullheavy runs the original heavy test without dividing its 11 scenarios. All shards use separate candidate copies.

## Findings

No outstanding product/spec findings. Revision 2 F1 and the twelve-gate inventory remain intact; revision 3 closes the provenance correction. Accept revision 3 via accept_cr; leave integration to the exact producer role/archetype.
