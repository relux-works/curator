# Narrowing controls

Apply each patch only to a new disposable task-local copy of this module.
Keep the candidate unchanged. Each mutation preserves the guard and admits
one prohibited class. Run with `-count=1`; cached tests are not evidence.

| Patch / plant | Narrowing | Named failing test |
|---|---|---|
| `checker-one-skill.patch` | Admit exactly one foreign skill when zero is required | `TestObservationChecker` |
| `checker-duplicate-count.patch` | Admit zero followed by contradictory one | `TestReviewContradictoryObservation/count-zero-then-one` |
| `checker-duplicate-outcome.patch` | Admit failed followed by ready | `TestReviewContradictoryObservation/outcome-failed-then-ready` |
| `snapshot-extra-entry.patch` | Ignore exactly `home/extra` | `TestFrozenSnapshot/extra` |
| `snapshot-go-listed-link.patch` | Follow only the listed root link | `TestFrozenSnapshot/listed-symlink`, `TestFrozenSnapshotSourceInputs/listed-symlink` |
| `snapshot-reference-listed-link.patch` | Reference wrapper follows only the listed root | `TestReferenceSnapshotPhysicalClosure/listed-symlink` |
| `snapshot-reference-dangling-link.patch` | Reference wrapper drops only the hidden dangling link | `TestReferenceSnapshotPhysicalClosure/unlisted-dangling-link` |
| `intake-one-foreign-build.patch` | Accept exactly one synthetic wrong digest | `TestVendorIntakeBoundary/wrong-digest` |
| `offline-loopback.patch` | Permit only outbound localhost TCP | `TestMuse1411OfflineBoundary` |
| `offline-one-keychain-file.patch` | Permit only the synthetic Keychains file | `TestMuse1411OfflineBoundary` |
| `MUSE1411_OMIT_EXCLUSION=exec` | Omit only exec exclusion | `TestMuse1411LoaderMatrix/exec/linked/skills/exclude-true` |
| `MUSE1411_OMIT_EXCLUSION=interactive` | Omit only interactive exclusion | Survivor: broken vendor exclusion already retains one skill; proves no interactive protection |

```sh
cp -R .research/muse1411 .temp/TASK-261021-mw4laf/mutant-new
git apply --directory=.temp/TASK-261021-mw4laf/mutant-new .research/muse1411/mutants/checker-one-skill.patch
go test ./.temp/TASK-261021-mw4laf/mutant-new -run '^TestObservationChecker$' -count=1 -v
```

The offline mutant must run **only** `TestMuse1411OfflineBoundary`, never
the vendor matrix. It touches only a synthetic local endpoint and synthetic
files. The wrong-digest control likewise invokes only that boundary test;
it never executes its synthetic file as a vendor binary.

Every listed failure needs a real exit 1 and its named assertion. A setup
failure or an unapplied patch is not a kill. Exact results and survivor bounds
are recorded in the task handoff alongside sanitized logs.
