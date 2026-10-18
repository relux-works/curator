# Review verdict — revision 6

```verdict-findings
{
  "verdict": "changes_requested",
  "revision": 6,
  "candidate_tree": "2fe068da665b203ed48e52cfcb7fba74de96e37e",
  "findings": [
    {
      "id": "unrelated-takeover-argument-discard",
      "row": "12",
      "invariant": "Takeover is admitted only for the five activating operations; env resolve requires --repair, including when help is requested.",
      "mechanism": "cmd/curator/command_help.go:152-154 admits all env resolve takeover arguments and delegates activation validation to cmdEnvResolve; cmd/curator/main.go:315 returns successful help before cmdEnvResolve at env.go:141 can enforce repair.",
      "reproductions": [
        {
          "test_file": "BUG-261018-uoisfe_review-probe-rev6.go",
          "test": "TestReviewNativeTakeoverHelpBoundary",
          "command": "go test -overlay .temp/review6/probe-overlay.json ./cmd/curator -count=1 -timeout=3m -v -run '^TestReviewNativeTakeoverHelpBoundary$'",
          "expected_failure": "exit 1 with three takeover-without-repair assertions: missing repair, --repair=false and --repair=0 each return 0 instead of 2; no-consent help controls return 0.",
          "observed_exit": 1,
          "log": "BUG-261018-uoisfe_review-evidence-rev6.txt:help-probe.log",
          "pinned_blobs": [
            "tree:2fe068da665b203ed48e52cfcb7fba74de96e37e"
          ]
        }
      ],
      "severity": "bypass",
      "repeat-of": "none"
    }
  ],
  "notes": [
    "The revision-1 list-wrapper mechanism is fixed. This is the same argument-discard class but a different mechanism (help bypasses activation validation); repeat-of is none by mechanism. Preserve the old regression and add the help witness plus a narrowing mutant in this leaf.",
    "No wall-clock review/free-hunt budget was supplied. All twelve rows were swept in order, persisted after each command; bounded free hunt then found the help bypass. No row is not-attacked.",
    "12 of 12 surface rows driven; 11 held, row 12 broken. All 5 AC rows exercised; AC5 is not fully satisfied because row12 is broken. A passing registered-command sweep is not proof of the activation/help cross-product.",
    "Group and interactive command positive controls in the producer sweep use help; actual no-flag controls exist for list/update/gc/version. This bound is explicit, not live-operation coverage for all commands.",
    "No full-suite replay in this review: reused the exact-tree local evidence, inspected the complete bundle and verified all 121 payload hashes. Targeted row checks, controls, exclusions and the new probe were rerun.",
    "Product tests pin HOME/USERPROFILE and native homes to temporary fixtures. The reviewer did not direct a product operation at real HOME. Toolchain/config reads by development tools are not claimed absent.",
    "Open spec questions remain Muse multi-root ownership, adapter-ledger coexistence, and opencode. No spec change requested."
  ],
  "surface_results": [
    {
      "row": "1",
      "result": "held",
      "detail": "Named attacks passed; see ordered sweep and command logs."
    },
    {
      "row": "2",
      "result": "held",
      "detail": "Named attacks passed; see ordered sweep and command logs."
    },
    {
      "row": "3",
      "result": "held",
      "detail": "Named attacks passed; see ordered sweep and command logs."
    },
    {
      "row": "4",
      "result": "held",
      "detail": "Named attacks passed; see ordered sweep and command logs."
    },
    {
      "row": "5",
      "result": "held",
      "detail": "Named attacks passed; see ordered sweep and command logs."
    },
    {
      "row": "6",
      "result": "held",
      "detail": "Named attacks passed; see ordered sweep and command logs."
    },
    {
      "row": "7",
      "result": "held",
      "detail": "Named attacks passed; see ordered sweep and command logs."
    },
    {
      "row": "8",
      "result": "held",
      "detail": "Named attacks passed; see ordered sweep and command logs."
    },
    {
      "row": "9",
      "result": "held",
      "detail": "Named attacks passed; see ordered sweep and command logs."
    },
    {
      "row": "10",
      "result": "held",
      "detail": "Named attacks passed; see ordered sweep and command logs."
    },
    {
      "row": "11",
      "result": "held",
      "detail": "Named attacks passed; see ordered sweep and command logs."
    },
    {
      "row": "12",
      "result": "broken",
      "detail": "See ordered sweep and command logs; help probe additionally breaks row 12."
    }
  ]
}
```

## Identity and validation

Base `5d8e6b07f7f485557d9bbb065c9692348ca065c2`; fresh fetched trunk `b9caa20901ee859eec20e649940693033e6297d4`. Candidate/upstream changed-path intersection is empty. Candidate tracked working blobs match the published tree; root README unchanged. Rework versus revision 1 is six CLI files; old list witness and list-only narrowing mutant passed before the ordered sweep. No production or tracked test file was edited.

Hosted run [39151661272](https://github.com/relux-works/curator/actions/runs/39151661272) reports completed/success, 20 successful jobs and two conditional skips. Its head `eecf6ae84dac3e03b70b0b031b947c0f550426b5` has signature G and the exact candidate tree. Configured validation ends with exit 0. Earlier infrastructure failures remain historical failures.

Local full-suite evidence is reused from signed `3e00fcc8abe75e411a88773c65c4042ae2448eb0`, the identical tree: 85 packages, 28 grouped lanes, 3494 top-level tests, with the producer-documented unchanged-package reuse bounds. The attached round-3 bundle has 121 intact payload hashes, a completed merged stream, and successful lint/vet/build ledgers. This review does not claim to have replayed those full checks.

## Ordered sweep

Paths without prefixes below are under `internal/envprofile/`. Every command uses `-count=1`; commands and full output are in the evidence bundle. Held means these attacks did not reproduce a violation, not proof of absence.

| Row | Result | Source/test location | Exact command (exit) |
|---|---|---|---|
| 1 | held | switch.go:169; native_skills_test.go:48 | `go test ./internal/envprofile -count=1 -timeout=3m -v -run ^TestNativeSkills(UseAndSync)$` (0) |
| 2 | held | managed.go:2568; native_skills_test.go:266 | `go test ./internal/envprofile -count=1 -timeout=3m -v -run ^TestNativeSkills(Repair)$` (0) |
| 3 | held | switch.go:800; native_skills_test.go:70 | `go test ./internal/envprofile -count=1 -timeout=3m -v -run ^TestNativeSkillsTakeover$/^foreign$` (0) |
| 4 | held | switch.go:811; native_skills_test.go:70 | `go test ./internal/envprofile -count=1 -timeout=3m -v -run ^TestNativeSkillsTakeover$/^(directory|file|store|relative-store)$` (0) |
| 5 | held | native_skills.go:31; native_skills_test.go:180 | `go test ./internal/envprofile -count=1 -timeout=3m -v -run ^TestNativeSkills(LinkedParent)$` (0) |
| 6 | held | switch.go:863; native_skills_test.go:317 | `go test ./internal/envprofile -count=1 -timeout=3m -v -run ^TestNativeSkills(ConvergenceAndSwitch|EmptyProfileLeavesUnnamedParent)$` (0) |
| 7 | held | native_skills.go:53; native_skills_test.go:200 | `go test ./internal/envprofile -count=1 -timeout=3m -v -run ^TestNativeSkills(Ledger|UnrecordedLedger)$` (0) |
| 8 | held | switch.go:863; native_skills_test.go:317 | `go test ./internal/envprofile -count=1 -timeout=3m -v -run ^TestNativeSkills(ConvergenceAndSwitch)$` (0) |
| 9 | held | switch.go:169; native_skills_test.go:372 | `go test ./internal/envprofile -count=1 -timeout=3m -v -run ^TestNativeSkills(InstallAndUpdate|ConvergenceAndSwitch)$` (0) |
| 10 | held | switch.go:1013; native_skills_test.go:424 | `go test ./internal/envprofile -count=1 -timeout=3m -v -run ^TestNativeSkills(BackupCollision)$` (0) |
| 11 | held | switch.go:707; native_skills_test.go:250 | `go test ./internal/envprofile -count=1 -timeout=3m -v -run ^TestNativeSkills(UnreadableMarker)$` (0) |
| 12 | broken: help bypass in free hunt | cmd/curator/main.go:283,315; cmd/curator/native_skills_flags_test.go:17 | `go test ./cmd/curator -count=1 -timeout=3m -v -run ^TestNativeSkills(CLIOperations|TakeoverFlagBoundary|TakeoverEveryRegisteredCommand|TakeoverArgumentDiscardRegression|TakeoverActivationBoundary)$` (0) |

The row-12 producer sweep exited 0; the additional public-run probe exited 1 and determines the final broken result. No absent row or unexecuted test is counted as held.

## Narrowing controls and exclusions

`go test ./internal/envprofile -count=1 -timeout=5m -v -run '^TestNativeSkillsNarrowingMutants$'` exited 0. All 12 plants produced expected behavioral child exit 1, including gate rows 3 dangling-foreign-consent, 4 plain-file-consent, 5 skills-parent-traversal, 6 unnamed-takeover-directory, 7 one-ledger-name and 11 directory-marker-is-absence. The CLI list-only control also exited 0 with both intended child failures. These controls do not cover the newly found help path.

`go test ./internal/envprofile -count=1 -timeout=2m -v -run '^Test(MusePublishedRegistryLayout|MuseDirectoriesAndIsolation|OpencodeFormFallback)$'` exited 0. The native sweep also checks opencode staging. Managed-home layout remains unchanged; managed Resolve gains only the required native repair hook.

## Reproduction and required rework

Materialize the attached probe as `.temp/review6/help_probe_test.go`. Create Go overlay `.temp/review6/probe-overlay.json` mapping the absolute path `cmd/curator/review_native_probe_test.go` to that scratch file. Run the finding command above against the candidate; the overlay only adds a test, preserving all candidate blobs.

Make activation validation hold before the help early return without duplicating guards that defeat existing mutants. Preserve successful help without takeover and all five activating paths. Add a named regression covering missing/false repair with help and a narrowing mutant for this bypass. Keep all accepted native rows, the list regression, and existing narrowing controls. No separate research/harness task is requested.

## Logbook

Revision 6 review: old list bypass fixed; ordered twelve-row sweep green on existing tests; free hunt reproduced a distinct help short-circuit bypass in row12. All controls completed. Verdict, probe and sanitized evidence attached before routing once to to-dev. No source edits or landing action.
