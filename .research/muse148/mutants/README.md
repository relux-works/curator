# Narrowing controls

These patches apply to the recorded replay source, not runtime code. They
retain the gate and weaken one boundary. Apply only in disposable task-local
copies. Never patch the candidate or run the vendor matrix with the offline
safety policy mutant.

```sh
cp -R .research/muse148 .temp/TASK-261016-i5gwaf/mutant-new
git apply --directory=.temp/TASK-261016-i5gwaf/mutant-new .research/muse148/mutants/checker-one-skill.patch
go test ./.temp/TASK-261016-i5gwaf/mutant-new -run '^TestObservationChecker$' -count=1 -v
```

Use a new copy for each control. `checker-one-skill.patch` must fail in
`TestObservationChecker` with `checker admitted one foreign skill`.
`snapshot-extra-entry.patch` must fail in `TestFrozenSnapshot/extra` with
`extra snapshot admitted`; copying `testdata` is necessary for this control.
`offline-loopback.patch` must fail in `TestMuse148OfflineBoundary` with
`network deny exit=0: MUSE148_NETWORK_CONTROL`. That test needs the documented
`MUSE148_BINARY`, `MUSE148_CURATOR` and a fresh `MUSE148_RUN_ROOT`; it starts
only the synthetic loopback endpoint and curl, never Muse/model inference.
Initial missing-fixture/unsupported-policy attempts do not count as kills.

All three corrected controls exited 1 for the intended assertion. The exec
omission plant uses `MUSE148_OMIT_EXCLUSION=exec` with the matrix test; the
interactive omission is an explicitly reported survivor.

Review rework adds five controls; each keeps every other rejection intact:

| Patch | Run mask | Required named failure |
|---|---|---|
| `checker-duplicate-count.patch` | `^TestReviewContradictoryObservation$` | `/count-zero-then-one`: admit only `user_skills=0 user_skills=1` |
| `checker-duplicate-outcome.patch` | `^TestReviewContradictoryObservation$` | `/outcome-failed-then-ready`: admit only failed followed by ready |
| `snapshot-go-listed-link.patch` | `^Test(FrozenSnapshot\|FrozenSnapshotSourceInputs)$` | Both `/listed-symlink` cases: follow only the listed root |
| `snapshot-reference-listed-link.patch` | `^TestReferenceSnapshotPhysicalClosure$` | `/listed-symlink`: follow only the listed root |
| `snapshot-reference-dangling-link.patch` | `^TestReferenceSnapshotPhysicalClosure$` | `/unlisted-dangling-link`: omit only the synthetic hidden link |

The reference controls need `MUSE148_SPEC` and `MUSE148_SPEC_PYTHON` as in
the probe README. Every final-source control exited **1** at its named
assertion; regular-file positive controls passed. These are behavioral suites,
not source-token scans. No new control survived. Exact commands, exits and
logs are in [review rework evidence](TASK-261016-i5gwaf_review-rework.md).
