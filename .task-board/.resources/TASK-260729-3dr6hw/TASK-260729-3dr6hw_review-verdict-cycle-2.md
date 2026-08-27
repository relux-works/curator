# TASK-260729-3dr6hw review verdict — cycle 2

Verdict: CHANGES REQUESTED to analysis. No code was edited and no Go command was run.

## Independently verified

- Verifier3 records `cmd/curator` passing at 557.779s, `internal/install` timing out at 603.306s with `TestStrictRegistryPolicyFailsUnknown` active for 3s, and `internal/install/atomicity` timing out at 603.701s with both sweep scenarios active for 8m28s. No DATA RACE marker is present.
- Static file/source order confirms `TestStrictRegistryPolicyFailsUnknown` is test 73 of 107, after 72 completed tests.
- The atomicity alarm positions, current-tree non-race subtest timings, 88 parallel / 19 sequential install split, and the 13-file producer allowlist match the cited sources. The 19 exclusions correctly cover 11 `t.Setenv` users, 7 `afterDocumentOpen` users, and the helper-process test.
- Distinct global user homes are compatible with `globalbins.Select`, which prefers each scenarios own `<userHome>/.local/bin` when it is on PATH.
- The projected plans preserve the unchanged 600s package timeout and provide plausible margin: install 220–340s and atomicity 340–460s, subject to focused measurement and verifier rerun.

## Required corrections

1. Section 8.3 has an impossible rsync expectation. `internal/install/cache_conformance_test.go` and `internal/install/dryrun_conformance_test.go` are already candidate-only and already appear as `*deleting` in the 23-line accepted-worktree delta. Modifying them cannot add `>fcsT....` entries. Correct the post-patch expectation to 11 new modified-file entries, for 34 lines total: the original 20 `*deleting` entries remain, the original 3 modified entries remain, and 11 allowlisted files that exist in both trees become modified. State explicitly that the pre/post SHA-256 manifest is the integrity proof for the two candidate-only test files.

2. Sections 2.2 and 6 miscount `saveJournal` call sites. Static `rg` finds 16 in `internal/transaction/engine.go` and 7 in `internal/transaction/staging.go`, for 23 total, not 24. The listed staging anchors themselves contain only seven entries. Correct every count while retaining the verified two copy-loop sites at staging.go:141 and staging.go:161.

3. Patch B1 is internally inconsistent unless it introduces a separate injection selector. Today `scenario.classes` drives both the fault-injection loop and the post-sweep full-class coverage assertion. The plan says that field must remain the full 7/5-class list while the loop becomes one class. Specify the literal data/function change, for example `injectClasses []string` or `failClass string`, with the injection loop using that selector and the final assertion continuing to use full `scenario.classes`. Add the new field/helper names to the function-level producer allowlist.

4. Resolve the wording contradiction between no assertion removed and the documented cross-class-chain reduction. Reviewer sanction: the partition is acceptable because every injection still checks whole-state digest equality, committed-class cutoff, reverse rollback, journal cleanup, and a full successful install covering all classes. Describe the cross-class sequencing check as intentionally retired defence in depth rather than claiming every test invariant is byte-for-byte unchanged.

## Routing

This is research artifact rework, so route to `analysis`, not `to-dev` and not `blocked`. Preserve all other cycle-2 findings, timings, allowlists, commands, and risk disclosures. After updating the existing diagnosis outcome, return for another independent reviewer cycle.