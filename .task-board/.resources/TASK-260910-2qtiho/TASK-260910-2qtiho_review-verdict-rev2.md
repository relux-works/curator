# TASK-260910-2qtiho — review verdict, revision 2

Verdict: **changes_requested**, route **to-dev**. One remaining correction: F2 provenance vocabulary. F1 scenario pinning and the F2 twelve-gate inventory are fixed.

Reviewed curator-spec HEAD `e8b53a003256433761cebce6080d6a955d777f25` plus the uncommitted candidate. The attached revision-2 patch and `git diff HEAD` have stable patch-id `d0399de31edf57b13be39b9d12c5f96a62a9a1da`. No candidate or index edits, commits, or LOGBOOK edits. Regeneration used a disposable tracked-file copy. No run goal is active.

## Remaining F2 correction — exact provenance vocabulary (medium)

The revision-2 rework brief explicitly requires “provenance values (`profile`, `explicit`, `lock`, `shipped`)”. Candidate `profiles/manager.md:1396` instead defines `profile-default`, `explicit`, `locked`, `shipped`; `protocol/environments.md:2843` repeats these spellings, and `conformance/v1/vectors/security-posture.json:54` closes the vocabulary to those old spellings. The validator's `SECURITY_POSTURE_PROVENANCE` (`tools/validate.py:7503`) and derived expected rows enforce them. Producer evidence explicitly says “Provenance spellings unchanged”. This is internally consistent but does not implement the rework brief's closed vocabulary.

Correction: use exactly `profile`, `explicit`, `lock`, `shipped` for provenance on both status commands, throughout normative prose, CHANGELOG, vector provenance/profile_source/sources/row values, validator derivation/pins and tests. Do not rename configuration fields such as `system.locked`: those are not provenance values. Regenerate manifest/release digests and attach the updated spec patch/evidence. Add negative checks rejecting the superseded provenance spellings. Preserve unrelated content. This is routine rework, with no human decision or external blocker required.

## Per-deliverable review

| Requirement | Evidence and assessment |
|---|---|
| Patch equals worktree | Stable patch-id above matches. `git diff HEAD --check` passes. Base is own HEAD e8b53a0, not moved main. PASS. |
| Revision-2 scope | 17/24 patch sections byte-identical to rev1. Only manager, environments, posture vector, validator/tests and manifest/release digests changed. PASS. |
| Knob/lock | `profiles/manager.md:46`: “one closed top-level knob”; line 80: “only in the direction of `hardened`”. `manager-config-v2.schema.json:46` enum permissive/hardened; `system-config-v2.schema.json:18,43` lock admission and hardened-only value. PASS. |
| Effective defaults and exceptions | Manager §7.1 (`profiles/manager.md:1095`) specifies audit/registry strict, non-empty source/MCP allowlists, explicit-null refusal, transitive error, signers true, unreachable-registry error. “Precedence is lock, then explicit machine value, then the profile default”, except three refusals. PASS. |
| Two rollout revisions | `profiles/manager.md:1122`: “Revision A (this release)” admits permissive default and exactly-once warning/hint; “Revision B (a later release)” flips the default. PASS. |
| Registry §4 | `protocol/registry.md:135`: “revocation is network-dependent”; prominent `registry_unreachable_during_install` names registry/artifacts and warns under permissive/errors under hardened. PASS against settled brief. |
| Registry §8 | `protocol/registry.md:262`: “The grace widens the advisory-policy residual”; original grace unchanged. PASS. |
| SECURITY | `SECURITY.md:40` recommends hardened defaults; line 519 names network-dependent advisory revocation and availability tradeoff. PASS. |
| Environments rules/tables | `protocol/environments.md:523` escalates empty MCP allowlist when declarations exist; line 2611 refuses explicit null. Tables at §2.1/§9.7/§10.4 updated; §12.1 references profile, preserving gate defaults. No new environments lock key requiring §12.2 addition. PASS. |
| Complete status inventory | `profiles/manager.md:1389`: both commands “carry the same closed posture inventory”; twelve gates plus header, same order/value/provenance. `protocol/environments.md:2832` repeats complete inventory. Schema-1 limited to header+four manager gates and no env posture section. Inventory PASS; exact provenance FAIL (F2 above). |
| Currency | `profiles/manager.md:1421` makes mandatory-refusal contradictions non-current; explicit permitted knob opt-outs remain current. Environments §12 agrees. PASS. |
| Schemas/generator | Four manager cases and three system cases cover valid values, invalid values/types and forbidden system direction. Generator property/lock tests updated. Frozen v1 config schemas 2/2 byte-identical. PASS. |
| Vectors/scenario pins | 17 cases cover A/B, defaults, explicit/lock precedence, three refusals, permissive warning, registry warning/error, schema1, status contradiction and shipped revisions. `tools/validate.py:7583` pins schema/posture/precedence/branch inputs; production dispatch calls gate at main. Replayed three prior holes through main: 3/3 refused. Independent whole-case sweep: 272/272 refused. F1 PASS. |
| Both outputs pinned | 16 schema2 cases carry 13 rows on each command; schema1 carries 5 curator rows and null env rows. Validator compares both with derived complete lists. PASS (subject to provenance correction). |
| Manifest/CHANGELOG | New family registered; `CHANGELOG.md:10`: “S1/S3”, both revisions, residual and exact diagnostics. CHANGELOG provenance needs same F2 correction. |
| Scope/boundaries | 24 paths: spec prose, schema, conformance, release digest, spec generator/validator/tests. No product implementation, proposals or LOGBOOK changes. All 33 pre-existing vector files byte-identical to HEAD. PASS. |
| Empty curator CR | Exact curator base `640a9df19ba94295cce059562101599996761f4b` to tree `57d2578ea9d1644d8a8931fdcdfbcb8ce4e15545` diff is empty; curator worktree clean. This is the correct repository boundary: the leaf delivers curator-spec changes and task-scoped artifacts, not curator implementation. Empty curator delta is not a rejection reason and does not prove spec integration. |

## Adversarial evidence

Attached `TASK-260910-2qtiho_review-probe-rev2.py` adapts the prior in-memory MCP rewrite to the new row fields. It recomputes internally consistent expected outputs and drives `validate.main()` for all three prior holes. Only loading the posture vector is replaced in memory; other main checks run normally against unchanged disk files. Thus refusal is by scenario validation, not a manifest checksum mismatch. Whole-case sweep separately drives the family gate.

```
validation failed: security-posture case refusal-mcp-allowlist-empty-with-declarations must leave machine.security_posture absent
validation failed: security-posture case revision-A-default-permissive-status must leave machine.audit absent
validation failed: security-posture case revision-B-default-hardened-flip-install must carry machine.schema_version == 2
Whole-case substitution sweep: 272/272 refused; main-entry replays: 3/3 refused
```

Probe exit 0. Bound: exhaustive substitutions within the 17-case corpus, not exhaustive arbitrary inputs or downstream manager execution. The spec expressly assigns manager execution downstream.

## Independent validation

Shell `/bin/zsh`, `set -o pipefail`, repository venv on PATH:
`PATH="/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin:$PATH" make validate`.
The monolithic invocation was deliberately interrupted (exit 130) to switch to bounded calls; it is not claimed green. The three Makefile legs were then executed independently, with Python discovery partitioned using the attached review shard harness. No producer-only test result substitutes for the independent checks below.

Validator entry (`python3 tools/validate.py`), exit 0:
```
validated 62 schemas and 1103 vector files
```
Go leg (`go test ./tools/...`), exit 0:
```
ok github.com/relux-works/curator-spec/tools/generate-vectors 5.048s
```
Regeneration (`make regenerate-check` from `/tmp/qtiho-review2/regen`, a disposable candidate copy with comparison index), exit 0:
```
go run ./tools/generate-vectors -root .
git diff --exit-code -- conformance/v1 release/1.0.0-rc.5.json release/1.0.0-rc.6.json release/1.0.0-rc.7.json release/1.0.0-rc.8.json release/1.0.0-rc.9.json
```

Additional F2 inventory probe: removing each individual posture row from every command output was refused, **421/421** omissions. This proves the inventory correction and does not resolve the exact provenance spelling mismatch.

### Completed Python replay

Discovery replay: **426/426 tests passed**, including **11/11** scenarios in the repeated-main test. Every bounded call exited **0**. Ordinary shards ran from the candidate with the repository venv: `python -B /tmp/qtiho-review2/shards.py 0`, then `1`, then `2`. The final test ran whole with `PYTHONPATH=tools python3 -B -m unittest test_validate.WriteNofollowVectorTests.test_substituted_scenario_rejected_through_main`; no scenario partitioning was needed.

shard0:
```text
Tests 142
..............................................................................................................................................
----------------------------------------------------------------------
Ran 142 tests in 329.894s

OK
```

shard1:
```text
Tests 142
..............................................................................................................................................
----------------------------------------------------------------------
Ran 142 tests in 295.269s

OK
```

shard2:
```text
Tests 141
.............................................................................................................................................
----------------------------------------------------------------------
Ran 141 tests in 361.758s

OK
```

heavy:
```text
.
----------------------------------------------------------------------
Ran 1 test in 266.332s

OK
```

All three Makefile validation legs passed in the bounded replay; regeneration passed. The initial interrupted monolithic make invocation is not a passing invocation. No producer-only evidence was substituted. Final candidate/attached-patch SHA-256: `bc2363422da4f59340f41cb90fc1167be4bfe648386bd94bd534e7e2fbdbcc0f`. Findings are recorded on the board and in this verdict; campaign prohibition on LOGBOOK.md edits is respected. Exactly one verdict: **changes_requested / to-dev**.
