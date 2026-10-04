# BUG-261004-2v9pbz — expanded-snapshot-budget-bypass-repeated-blobs

Verdict: accepted, revision 1. No blocking findings. Route through accept_cr to integrating; producer integration remains required.

Reviewed base ca1b776fb580ec0cee0173bf150daf063023aeaa and exact candidate tree 438f9c753fdbc238e7125a5440ce73710b98ca77. Fresh origin HEAD advertised main at the base OID; exact-ref fetch matched. Only the three declared paths changed. No LOGBOOK.md, CHANGELOG.md, or unrelated changes. Reviewer made no changes to the repository code; base and mutation probes used isolated exports.

## Swept surfaces and acceptance criteria

| Surface | Result and evidence |
| --- | --- |
| Per-path emission accounting (AC1) | reserveFile checks files, aggregate content, and canonical framing before content copy/file append. Cache object counts/bytes remain independent. Subtractive arithmetic avoids overflow. Production local exact/one-over limits pass; after-append mutant fails. |
| Canonical allocation (AC1/5) | Header and every record reserved before framing; exact-sized allocation prevents capacity growth. Frozen one-blob/empty-tree bytes and SHA1/SHA256 parity pass. |
| Expanded tree DAG and cancellation (AC2/3) | Entries charged before path retention, including empty directories and aliases. Exact 128/127 and empty 64/63 cases pass. Context checked at tree and each entry; cached-walk cancellation test passes. |
| Local production entry (AC3) | Real AdmitLocal fixtures: one blob admitted, 64 distinct refused, 64 aliases refused, repeated tree DAG refused with build_repository_incomplete_source. Base admits both adversarial alias cases and fails assertions. |
| Network production entry (AC4) | AcquireNetwork invokes proveRepository (admission.go); AdmitLocal invokes same function (local.go). Real Git with POSIX test transport runs one/64 aliases; base admits 64 and fails, candidate refuses. External network transport itself is not exercised. |
| Conformance/scope (AC5) | Nine targeted buildrepo rc.14 conformance tests pass, zero skips. Exact candidate hosted gate succeeds. Docs describe stricter bounds as required by manager profile 11.5. |

Coverage: 5/5 acceptance criteria reviewed; 2/2 admission entry points exercised; 3/3 targeted mutants killed. These ratios describe this bounded review, not all possible resource exhaustion attacks. Cancellation timing and partial internal emission are observed at the walker because discarded partial state is not exposed by admission APIs; production call wiring independently inspected.

## Independent commands and exits

All Go checks used GOFLAGS=-work, -count=1 and ./internal/buildrepo. Attached review-tests log contains full output with private WORK paths removed.

- Base export plus only the two compatible new entry test functions and their fixture: `go test ./internal/buildrepo -run 'Test(AdmitLocalExpandedSnapshotBudget|AcquireNetworkExpandedSnapshotBudget)$' -count=1 -v`, exit 1. Failures are admission of local aliases, repeated-tree DAG and network aliases; controls pass. No production code backported.
- Exact candidate: `go test ./internal/buildrepo -run 'Test(AdmitLocal.*Budget|AcquireNetworkExpandedSnapshotBudget|SnapshotBudget)' -count=1 -v`, exit 0; seven top-level tests, zero skips.
- Required mutant: move reserveFile after content copy/file append. `-run '^TestSnapshotBudgetRefusesBeforeEmission$'`, exit 1: four files emitted instead of three for byte/framing/file bounds.
- Required mutant: reserve only uncached blob OIDs. `-run 'Test(AdmitLocal|AcquireNetwork)ExpandedSnapshotBudget$'`, exit 1: local aliases, DAG and network aliases wrongly admitted.
- Reviewer-added mutant: weaken expanded-entry comparison from >= to >. `-run '^TestAdmitLocal(ExpandedTreeEntryBudget|EmptyTreeDAGEntryBudget)$'`, exit 1: both one-over cases wrongly admitted. Positive exact-limit controls stay green.
- Exact candidate conformance: `-run 'Test(NetworkAndLocalSHA1SHA256RawObjectParity|RawObjectAndLFSPinnedConformanceFixtures|PackIndexConformanceAndExactSSHWrapper|ExternalRepositoryAcquisitionConformance|LocalPackedHeadAndSnapshotMaterialization|LocalConfigAndAdministrationAdversarialBoundaries|ReleasedSkillBuildSchemaCases|CanonicalRepositorySourceVectors|ReleasedSourceIdentityVectors)$'`, exit 0; nine top-level tests, zero skips. Selected CURATOR_CONFORMANCE_ROOT manifest SHA256: 6f832d813efc768ea154a7d5076b512ab4be6aa9409d92e11469d21ea9bc69f5 (rc.14).
- Exact base/candidate `git diff --check`: exit 0. gofmt check: exit 0, no output.

## Hosted gate and evidence reuse

Board resource BUG-261004-2v9pbz_change-request_rev1-validation.log records `sh scripts/remote-gate.sh`, exit 0, required command shards 1/1 green. Independently queried GitHub: run 37171394793 completed successfully at head a68987c8c4e83fa3ce7ca379970553f48751ddad; its Git tree is exactly 438f9c753fdbc238e7125a5440ce73710b98ca77.

https://github.com/relux-works/curator/actions/runs/37171394793

Hosted lint, Linux/macOS race, Linux/macOS/Windows tests, interop, naming and gate self-tests succeed. Rose-air and optional candidate-suite jobs are skipped; do not claim those ran. This hosted result supersedes the earlier producer outcome's pending statement. Reviewer did not rerun full suite, lint or package build locally; hosted exact-tree evidence is accepted for broad validation. Local verification is limited to the commands above. No heavy OOM stress or external server test was run.

Run goal queried: not goal-bound. No directives recorded. Findings persist in board evidence rather than LOGBOOK.md under binding produce mode.
