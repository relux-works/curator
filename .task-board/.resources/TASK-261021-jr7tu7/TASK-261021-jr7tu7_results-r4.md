# TASK-261021-jr7tu7 results r4

## Delivery

- Repository: curator-spec. New branch: `repository-transport-lanes-r4`.
- Base: current fetched `origin/main`, `76784377893ab578489f8ba098763b5937ed04fa`, containing the landed CIP-0016, CIP-0022, CIP-0002 and CIP-0023 work.
- Head: `dd2777743d60aac3ffd2011278d5b0725285cfb6`.
- Signed cherry-picks: `608e512cbd955f598226dbf4ba9bdebac528ffb2` and `dd2777743d60aac3ffd2011278d5b0725285cfb6`; both signatures `G`.
- Both cherry-picks applied without conflicts. Their patches are unchanged from the accepted range `e41f96387287f5e76604627f9e63662f7d6d86d4..71b24e02d2f9fbeebf70a6c881adc1ba1a1648ff`.
- Ordinary feature push exited 0; remote tracking head readback equals the tested head. No main push, force push, PR or board-status change.

## Files and integrity

The final five-file delta is unchanged: seven lane-side backlinks in `profiles/manager.md`, `protocol/core.md`, `protocol/environments.md` and `protocol/skillfile-sources.md`, plus the live draft-owned prose digest in `conformance/draft-sources-v2/manifest.json`. No additional change, URL grammar expansion or transport rule change.

`protocol/repository-transport.md`, `conformance/skillfile-sources-v1/manifest.json`, `conformance/candidate.json`, the core manifest and rc.14 release metadata are byte-identical to the new main. Recomputed 138 accepted-source and 1294 core inventory entries preserve all non-exempt main bytes and expected digests. The only pre-existing live prose exception remains `protocol/skillfile-sources.md`, which the draft manifest owns; its accepted validation uses rc.14 bytes. Candidate manifest pins match. Production gate/checker code and tests are unchanged.

## Commands, exits and claims

All evidence is for the exact delivery head. E is `~/Werkstatt/curator-dev4/curator-spec/.temp/TASK-261021-jr7tu7/r4/`. PY is the existing local validation venv, Python 3.14.8/jsonschema 4.25.1. Go is configured `CURATOR_GO`, Go 1.25.5. No installs.

| Command/check | Exit | Claim and evidence |
| --- | ---: | --- |
| `a2a inbox`; board assignment/preconditions query; five `resource get` calls; ack | 0 each | Latest addressed orchestrator request and assignment confirmed; all preconditions read. |
| `git fetch origin`; main fast-forward/ref equality; new branch; `git cherry-pick -S` accepted two commits | 0 each | New branch starts at current main; no conflict resolution or extra edit. |
| `git range-diff OLD_BASE..OLD_HEAD origin/main..HEAD` | 0 | Both accepted patches identical; output below and `range-diff.txt`. |
| `PY -B E/inventory.py` | 0 | Independently recomputed frozen inventory and verified preserved main bytes, digests and candidate pins. `inventory.log`, `frozen-inventory.json`. |
| `env -u GIT_DIR -u GIT_WORK_TREE PY -B E/check-editorial.py` | 0 | Five-file scope, seven valid backlink anchors, frozen/checker bytes equal main, no private identifiers in added prose. Real link checker accepts valid target and rejects planted missing target with intended diagnostic. `editorial.log`. |
| `env -u GIT_DIR -u GIT_WORK_TREE PY -B E/production-gate.py` | 0 | Unchanged production `release_gate.validate_skillfile_sources_manifest` accepts accepted corpus with only the draft-owned doc restored from rc.14; rejects transport newline with intended transport-digest error; rejects rewritten manifest with stale release pin; restored positive passes. `production-gate.log`. |
| Raw live draft production-gate call within that helper | expected refusal; wrapper 0 | Sole draft-owned Skillfile prose yields its exact digest-mismatch refusal. This does not claim strict production acceptance of raw draft prose. |
| `PY -B tools/verify_skillfile_sources_independence.py` with explicit spec Git context | 0 | 89/89 schema references byte-identical to rc.10, 34/34 cited clauses present; existing single exception unchanged. `prose.log`. |
| `env -u GIT_DIR -u GIT_WORK_TREE PY -B tools/validate.py` | 0 | 74 schemas and 1294 vectors validate. `validator.log`, `validation-exits.json`. |
| `env -u GIT_DIR -u GIT_WORK_TREE PY -B E/suite-groups.py N`, N=0..124 | 0 each | All 879 unique discovered Python tests ran exactly once, including all 40 release-gate tests. Every group has command, selected IDs, exit, count, duration and matching unittest OK summary. `suite-groups.json`, `python-001..125.{json,log}`, `validation-summary.json`. |
| `env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test ./tools/...` | 0 | Complete Go tools suite passes. `go.log`, `validation-exits.json`. |
| `PY -B E/final-check.py`; signature/ancestry/diff/status checks | 0 | Both introduced commits signed G; source and both clones clean at exact head; no whitespace errors. `validation-summary.json`. |
| `git push -u origin repository-transport-lanes-r4`; tracking-head readback | 0 each | Published head equals tested head. `push.log`. |

The three `make validate` recipes were fully executed in two disposable normal-.git checkouts at the exact head, with `TMPDIR=/tmp/cw-dev4` and inherited Git variables removed. Python discovery was partitioned into 125 calls to honor the ten-minute shell-call limit; this is complete recipe-equivalent validation, not a claim of a monolithic make invocation. Go caches stayed inside the lane. Slowest Python group: 259.947 seconds. Both clean validation clones were deleted after final verification. Logs remain in E.

This is an unchanged editorial reapplication. Existing executable negative/positive release regressions plus bounded link, transport-digest and release-pin controls cover the affected claims; no behavior or new gate was introduced.

## Range-diff

```text
1:  d63a1dd = 1:  608e512 docs: enumerate covered repository transport lanes
2:  71b24e0 = 2:  dd27777 docs: preserve frozen transport corpus in lane clarification
```

## Open decisions

None. Prior lane inventory and contradiction findings remain applicable because both accepted patches are unchanged. Editing frozen transport prose remains outside this rebase request and would require a separate owner-authorized revision. No native runtime qualification is claimed.
