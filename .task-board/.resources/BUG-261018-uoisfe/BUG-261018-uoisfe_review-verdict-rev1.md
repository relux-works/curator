# Review verdict — native profile skills, revision 1

Verdict: **changes_requested**, CR revision 1.

```verdict-findings
{
  "findings": [
    {
      "id": "unrelated-takeover-argument-discard",
      "row": "12",
      "invariant": "Round-2 row 12 requires --takeover acceptance on the five named operations and on nothing else.",
      "mechanism": "cmd/curator/command_help.go:30 discards all list arguments before calling cmdList; the producer surface map excludes this built-in command without an approved scope amendment.",
      "reproductions": [
        {
          "test_file": "BUG-261018-uoisfe_review-probe-rev1.go",
          "command": "CURATOR_TEST_CLI_HELPER=1 go test -overlay .temp/review-uoisfe/probe-overlay.json -count=1 -timeout=2m -v ./cmd/curator -run '^TestReviewNativeSkillsUnrelatedTakeover$/^list$'",
          "expected_failure": "go test exits 1: unrelated takeover admitted: exit 0, want 2. The same list call without takeover passes first.",
          "pinned_blobs": [
            "tree:97af0e0f6eb16747393683be175b7e297d3b60bf",
            "git:61da5a19975e49e54a83574aac884448e4851e92"
          ]
        }
      ],
      "severity": "bypass",
      "repeat-of": "none"
    }
  ],
  "notes": [
    "The update probe did not reach its negative assertion: its positive fixture failed with an empty skills-root path. It is not a second finding. Only the isolated list reproduction is blocking.",
    "This is a pre-existing parser behavior, but the explicit row-12 acceptance contract includes it. The observed effect is ignored consent syntax, not unauthorized takeover by list.",
    "No review/free-hunt wall-clock budget was specified. The ordered sweep and bounded additional probes completed in this turn.",
    "There is no new runtime source-text attestation gate. Mutants execute behavioral tests, with exact named exit-1 failures."
  ],
  "surface_results": [
    {
      "row": "1",
      "result": "held",
      "detail": "See ordered surface sweep and attached command evidence."
    },
    {
      "row": "2",
      "result": "held",
      "detail": "See ordered surface sweep and attached command evidence."
    },
    {
      "row": "3",
      "result": "held",
      "detail": "See ordered surface sweep and attached command evidence."
    },
    {
      "row": "4",
      "result": "held",
      "detail": "See ordered surface sweep and attached command evidence."
    },
    {
      "row": "5",
      "result": "held",
      "detail": "See ordered surface sweep and attached command evidence."
    },
    {
      "row": "6",
      "result": "held",
      "detail": "See ordered surface sweep and attached command evidence."
    },
    {
      "row": "7",
      "result": "held",
      "detail": "See ordered surface sweep and attached command evidence."
    },
    {
      "row": "8",
      "result": "held",
      "detail": "See ordered surface sweep and attached command evidence."
    },
    {
      "row": "9",
      "result": "held",
      "detail": "See ordered surface sweep and attached command evidence."
    },
    {
      "row": "10",
      "result": "held",
      "detail": "See ordered surface sweep and attached command evidence."
    },
    {
      "row": "11",
      "result": "held",
      "detail": "See ordered surface sweep and attached command evidence."
    },
    {
      "row": "12",
      "result": "broken",
      "detail": "See ordered surface sweep and attached command evidence."
    }
  ],
  "free_hunt": []
}
```

## Identity and freshness

CR base `5d8e6b07f7f485557d9bbb065c9692348ca065c2`; fresh fetched main `2e5ab6dab61fc78a525cbcb0de6c5fcc9f5ae882`. Candidate/upstream changed-path intersection is empty. Candidate tree matches signed snapshot `b32823ac920d8b0485c4a7a7111b48fff8e3ca2b` (signature `G`). All ten working-file blobs were compared byte-for-byte with the candidate. Root README is unchanged. No production or tracked test file was edited by this reviewer.

## Ordered surface sweep

**12 of 12 rows driven; 11 held, 1 broken.** Producer coverage names all rows, but its row-12 bound excludes an explicitly required built-in refusal. A declared implementation limit does not amend the brief. Rows 1–11 held only for the attacks named here, not as proof of absence. Each sequential row command exit/log was appended to sweep.jsonl before starting the next row.

| Row | Surface | Result | Test location and production entry | Command / evidence | Exit |
|---|---|---|---|---|---:|
| 1 | Absent target | held | `internal/envprofile/native_skills_test.go:48`; UseWithPolicy / SyncWithPolicy -> switch.go:690; `TestNativeSkillsUseAndSync` | `go test -count=1 -timeout=3m -v ./internal/envprofile -run ^TestNativeSkills(UseAndSync)$`; row-01.log | 0 |
| 2 | Recorded manager link | held | `internal/envprofile/native_skills_test.go:266`; Resolve / StatusOf -> native_skills.go:65 / managed.go:2648; `TestNativeSkillsRepair` | `go test -count=1 -timeout=3m -v ./internal/envprofile -run ^TestNativeSkills(Repair)$`; row-02.log | 0 |
| 3 | Foreign symlink | held | `internal/envprofile/native_skills_test.go:70`; UseWithPolicy -> switch.go:792; `TestNativeSkillsTakeover` | `go test -count=1 -timeout=3m -v ./internal/envprofile -run ^TestNativeSkillsTakeover$/^foreign$`; row-03.log | 0 |
| 4 | Other unrecorded entry | held | `internal/envprofile/native_skills_test.go:70`; UseWithPolicy -> switch.go:810; `TestNativeSkillsTakeover` | `go test -count=1 -timeout=3m -v ./internal/envprofile -run ^TestNativeSkillsTakeover$/^(directory|file|store|relative-store)$`; row-04.log | 0 |
| 5 | Path components | held | `internal/envprofile/native_skills_test.go:180`; UseWithPolicy -> inPlaceSkillsLedger -> managedPath; `TestNativeSkillsLinkedParent` | `go test -count=1 -timeout=3m -v ./internal/envprofile -run ^TestNativeSkills(LinkedParent)$`; row-05.log | 0 |
| 6 | Unnamed entries | held | `internal/envprofile/native_skills_test.go:317`; SyncWithPolicy / UseWithPolicy -> switch.go:849; `TestNativeSkillsConvergenceAndSwitch` | `go test -count=1 -timeout=3m -v ./internal/envprofile -run ^TestNativeSkills(ConvergenceAndSwitch|NativeSkillsEmptyProfileLeavesUnnamedParent)$`; row-06.log | 0 |
| 7 | Adapter ledger | held | `internal/envprofile/native_skills_test.go:200`; UseWithPolicy / StatusOf -> native_skills.go:31 / status.go:1069; `TestNativeSkillsLedger` | `go test -count=1 -timeout=3m -v ./internal/envprofile -run ^TestNativeSkills(Ledger|NativeSkillsUnrecordedLedger)$`; row-07.log | 0 |
| 8 | Old staging | held | `internal/envprofile/native_skills_test.go:317`; SyncWithPolicy -> switch.go:849; `TestNativeSkillsConvergenceAndSwitch` | `go test -count=1 -timeout=3m -v ./internal/envprofile -run ^TestNativeSkills(ConvergenceAndSwitch)$`; row-08.log | 0 |
| 9 | Profile switch | held | `internal/envprofile/native_skills_test.go:372`; Install / UpdateWithPolicy / UseWithPolicy -> materializeOne; `TestNativeSkillsInstallAndUpdate` | `go test -count=1 -timeout=3m -v ./internal/envprofile -run ^TestNativeSkills(InstallAndUpdate)$`; row-09.log | 0 |
| 10 | Backup generation | held | `internal/envprofile/native_skills_test.go:424`; UseWithPolicy -> switch.go:1013; `TestNativeSkillsBackupCollision` | `go test -count=1 -timeout=3m -v ./internal/envprofile -run ^TestNativeSkills(BackupCollision)$`; row-10.log | 0 |
| 11 | Unreadable marker | held | `internal/envprofile/native_skills_test.go:250`; UseWithPolicy -> switch.go:698 envmarker.Read; `TestNativeSkillsUnreadableMarker` | `go test -count=1 -timeout=3m -v ./internal/envprofile -run ^TestNativeSkills(UnreadableMarker)$`; row-11.log | 0 |
| 12 | Flag acceptance | broken | `cmd/curator/native_skills_test.go:112; cmd/curator/command_help.go:30`; run -> commandRegistry / profile and env handlers; `TestNativeSkillsTakeoverFlagBoundary` | `go test -count=1 -timeout=3m -v ./cmd/curator -run ^TestNativeSkills(TakeoverFlagBoundary|NativeSkillsCLIOperations)$`; row-12.log | 0 |

Row 12 additionally ran `TestNativeSkillsCLIOperations` separately (all five operations; cli-operations.log, exit 0), then the independent list refusal probe (flag-list.log, intended test exit 1). The loop masks did not select the second alternation on rows 6/7; the explicit bounds replay below supplies those extra tests. No absent test is counted as executed.

## Controls, exclusions, and validation

- `go test -count=1 -timeout=4m -v ./internal/envprofile -run '^TestNativeSkillsNarrowingMutants$'`: exit 0, all 12 controls killed by their named behavioral exit-1 failures; mutants.log. Gate rows 3/4/5/6/7/11 respectively: dangling-foreign-consent, plain-file-consent, skills-parent-traversal, unnamed-takeover-directory, one-ledger-name, directory-marker-is-absence. These narrow exceptions, not whole-gate deletions.
- `go test -count=1 -timeout=2m -v ./internal/envprofile -run '^Test(MuseDirectoriesAndIsolation|NativeSkillsEmptyProfileLeavesUnnamedParent|NativeSkillsUnrecordedLedger)$'`: exit 0, bounds.log. Muse managed isolation retained; opencode staging is asserted by row 1. Managed-home planning/layout is unchanged by the diff; Resolve now additionally repairs current native state as requested.
- All exercised fixtures pin HOME/USERPROFILE and native adapter roots to temporary directories. The overlay adds a test only; production blobs stay exact.
- Full local suite was **reused, not rerun**: attached validation stream has 80 distinct passing packages and 5 no-test package skips, covering all 85 served packages; zero package failures. Producer documents refreshed envprofile/CLI groups and unchanged-input reuse for the other packages. Raw archive exists and SHA-256 matches `ca67a410e3e171a6180fe6d03f2d7ef54c1042c77004fb19295f52fb8f0c2209`. Command exits, platform skip classifications, lint/vet/build evidence remain in the producer validation-metadata and validation-r2 resources.
- Fresh `gh run view 39094506341 --json headSha,status,conclusion,jobs`: exit 0, exact signed SHA, completed/success, 20 successful jobs, 2 conditional skips. No infrastructure failure or synthetic status. The hosted gate does not detect the missing list refusal.

## Required rework

Make the built-in unrelated command boundary reject --takeover, preserving the valid argument-free commands. Commit a named regression test through CLI run for this list witness and relevant neighboring wrappers, plus a narrowing mutant that reintroduces one ignored-argument exception. Keep the five authorized takeover paths green. Re-run affected validation and publish the next candidate. Do not split this finding into a separate harness task.

## Reproduce the independent witness

Download the attached review-probe-rev1.go to `.temp/review-uoisfe/flag_probe_test.go`. Create `.temp/review-uoisfe/probe-overlay.json` with a Go overlay Replace mapping from the absolute `cmd/curator/review_native_probe_test.go` path to that scratch file. Run the finding command from the exact candidate checkout. The overlay adds only the test and leaves every candidate production blob unchanged.

Recording note: the initial generic JSON fence was parsed as legacy by the board. This resource now uses the required `verdict-findings` fence; the same reviewer re-records the same finding set, without changing the review outcome.
