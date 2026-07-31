# TASK-260729-2afulh — review verdict

## Verdict branch: ACCEPTED

Accept the investigation and its routing, then mark the task `done`.

This acceptance means the separately scoped fixture-trim prototype was reproduced,
measured, and adjudicated correctly. It does **not** approve integrating the
14-file patch as the atomicity timeout solution. The fixture-only option is
rejected because its valid count-one race gate took 493 seconds, above the
strict 480-second ceiling and with negative margin. Preserve the evidence and
continue through production-side successor `TASK-260729-365r5r`.

## Independent review evidence

### Scope and manifests

- The six recorded source/baseline/final manifests inspected by the reviewer
  each contain 391 entries and are path-sorted.
- The accepted `TASK-260729-rfrdfo` prototype manifest is byte-identical to this
  task's baseline manifest.
- The tester's freshly regenerated baseline and final manifests are
  byte-identical to the producer's recorded baseline and final manifests.
- Re-running the literal integrity checker over the tester manifests exited 0:
  `modified=14 added=0 deleted=0 unexpected=0 forbidden_touched=0`,
  `ALLOWLIST_SIZE=14 MODIFIED=14`, `INTEGRITY_OK`.
- Baseline-to-final evidence contains exactly one newly changed path:
  `internal/install/atomicity/fixture_test.go`. Direct comparison of the current
  task trees confirms that diff. The measurement-only
  `zz_measure_test.go` is absent from the deliverable.
- The evidence bundle SHA-256 was independently verified as
  `6ab49fff580eb03b5bb8608f5f856fc64195c6c0be36579d4d5af239702dbb23`.

The literal 14-file scope is the inherited 13 paths:

1. `internal/install/cache_conformance_test.go`
2. `internal/install/commit_test.go`
3. `internal/install/diagnostics_test.go`
4. `internal/install/dryrun_conformance_test.go`
5. `internal/install/generation_test.go`
6. `internal/install/install_test.go`
7. `internal/install/maintenance_test.go`
8. `internal/install/private_test.go`
9. `internal/install/registry_e2e_test.go`
10. `internal/install/revalidation_test.go`
11. `internal/install/stage_test.go`
12. `internal/install/atomicity/activation_test.go`
13. `internal/install/atomicity/commit_atomicity_test.go`

plus:

14. `internal/install/atomicity/fixture_test.go`

### Assertion neutrality and isolation

- Before/after counts are identical for assertions, tests, subtests,
  `t.Parallel`, `t.TempDir`, `t.Setenv`, skips, and timeout tokens.
- The only mechanical count change is the intended fixture write:
  `e.write(` 8 to 7.
- `activation_test.go`, `commit_atomicity_test.go`, `doc.go`, and inherited
  Patch A's `internal/install/install_test.go` compare byte-identically.
- The only remaining atomicity-package `references/info.md` mention is the
  explanatory comment. Whitelist coverage remains in
  `internal/whitelist/whitelist_test.go` and
  `internal/interop/golden_test.go`.
- Source inspection confirms rollback state uses recursive whole-tree digests,
  while scenario class partitions, distinct global user homes, environment
  isolation, and transaction assertions are unchanged.

### Measured staging reduction

The same measurement hook is byte-identical in both scratch arms and absent
from the deliverable. It observed staging entries and non-empty chunk syncs:

| scenario | phase | entries | chunks | staging-path saves |
| --- | --- | ---: | ---: | ---: |
| project | baseline | 24 -> 20 | 14 -> 12 | 100 -> 84 |
| project | upgrade | 34 -> 28 | 19 -> 16 | 140 -> 116 |
| global | baseline | 26 -> 22 | 16 -> 14 | 110 -> 94 |
| global | upgrade | 25 -> 21 | 16 -> 14 | 107 -> 91 |

The save counts are exact staging-path arithmetic over the measured values, not
a claimed direct hook: source inspection confirms `stageTarget` calls
`saveJournal` three times per staging entry and `copyStagingFile` calls it
twice per copied chunk, so `3*entries + 2*chunks` reproduces every row.

### Focused gates and barrier

- Format, build, and vet each have real exit 0.
- Focused atomicity non-race has real exit 0 in 273 seconds.
- Count-one atomicity race repetition 1 has real exit 0 in 493 wall-clock
  seconds (`ok ... 492.231s`) and no `DATA RACE`.
- Each executed gate has its own two-scan `BARRIER_OK` record.
- Race repetition 2 has an empty partial log and no exit/seconds file, so it is
  correctly excluded from evidence. Repetition 3 has no artifact because it was
  not run.
- `DRIVER-STOP-REASON` and the annotated manual `DRIVER-DONE` accurately record
  the cooperative stop after the first valid race result made the strict
  predicate false. They do not claim natural completion of all repetitions.
- The immutable inherited baseline remains exit 0 at 286 seconds non-race and
  593/561/564 seconds for the three count-one race repetitions. Patch A remains
  unchanged and green by that inherited evidence.

## Acceptance decision

The fixture removal produces a real 12.5–17.6% staging-work reduction and
improves the isolated race runtime, but 493 seconds still misses the required
`<=480s` gate by 13 wall-clock seconds (12.231 seconds by the test-reported
duration). The task explicitly permits rejecting the trim when it lacks
defensible margin. The correct outcome is therefore:

- accept this task as a complete, evidence-backed investigation;
- reject and archive the fixture-only patch as the timeout solution;
- do not integrate it on the strength of this task;
- continue the production optimization through `TASK-260729-365r5r`.

The reviewer ran no Go/build/test/vet/format command and modified no prototype
or product file.
