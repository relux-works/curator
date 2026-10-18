# Review verdict — revision 7

```verdict-findings
{
  "verdict": "accepted",
  "revision": 7,
  "candidate_tree": "769581956b4fd7bfd02c2567d7465b60335f5b4b",
  "findings": [],
  "notes": [
    "Held records only the named attacks; no proof of absence is claimed.",
    "No review/free-hunt budget supplied. Sequential sweep persisted each row before the next; additional bounded hunt completed.",
    "Group/interactive command positive controls use help; no live login/UI operation coverage claimed.",
    "Full local validation was inspected and reused, not independently replayed in full. See exact identity and reuse bounds below."
  ],
  "surface_results": [
    {
      "row": "1",
      "result": "held",
      "command": [
        "go",
        "test",
        "./internal/envprofile",
        "-count=1",
        "-timeout=3m",
        "-v",
        "-run",
        "^TestNativeSkillsUseAndSync$"
      ],
      "exit": 0
    },
    {
      "row": "2",
      "result": "held",
      "command": [
        "go",
        "test",
        "./internal/envprofile",
        "-count=1",
        "-timeout=3m",
        "-v",
        "-run",
        "^TestNativeSkillsRepair$"
      ],
      "exit": 0
    },
    {
      "row": "3",
      "result": "held",
      "command": [
        "go",
        "test",
        "./internal/envprofile",
        "-count=1",
        "-timeout=3m",
        "-v",
        "-run",
        "^TestNativeSkillsTakeover$/^foreign$"
      ],
      "exit": 0
    },
    {
      "row": "4",
      "result": "held",
      "command": [
        "go",
        "test",
        "./internal/envprofile",
        "-count=1",
        "-timeout=3m",
        "-v",
        "-run",
        "^TestNativeSkillsTakeover$/^(directory|file|store|relative-store)$"
      ],
      "exit": 0
    },
    {
      "row": "5",
      "result": "held",
      "command": [
        "go",
        "test",
        "./internal/envprofile",
        "-count=1",
        "-timeout=3m",
        "-v",
        "-run",
        "^TestNativeSkillsLinkedParent$"
      ],
      "exit": 0
    },
    {
      "row": "6",
      "result": "held",
      "command": [
        "go",
        "test",
        "./internal/envprofile",
        "-count=1",
        "-timeout=3m",
        "-v",
        "-run",
        "^TestNativeSkills(ConvergenceAndSwitch|EmptyProfileLeavesUnnamedParent)$"
      ],
      "exit": 0
    },
    {
      "row": "7",
      "result": "held",
      "command": [
        "go",
        "test",
        "./internal/envprofile",
        "-count=1",
        "-timeout=3m",
        "-v",
        "-run",
        "^TestNativeSkills(Ledger|UnrecordedLedger)$"
      ],
      "exit": 0
    },
    {
      "row": "8",
      "result": "held",
      "command": [
        "go",
        "test",
        "./internal/envprofile",
        "-count=1",
        "-timeout=3m",
        "-v",
        "-run",
        "^TestNativeSkillsConvergenceAndSwitch$"
      ],
      "exit": 0
    },
    {
      "row": "9",
      "result": "held",
      "command": [
        "go",
        "test",
        "./internal/envprofile",
        "-count=1",
        "-timeout=3m",
        "-v",
        "-run",
        "^TestNativeSkills(InstallAndUpdate|ConvergenceAndSwitch)$"
      ],
      "exit": 0
    },
    {
      "row": "10",
      "result": "held",
      "command": [
        "go",
        "test",
        "./internal/envprofile",
        "-count=1",
        "-timeout=3m",
        "-v",
        "-run",
        "^TestNativeSkillsBackupCollision$"
      ],
      "exit": 0
    },
    {
      "row": "11",
      "result": "held",
      "command": [
        "go",
        "test",
        "./internal/envprofile",
        "-count=1",
        "-timeout=3m",
        "-v",
        "-run",
        "^TestNativeSkillsUnreadableMarker$"
      ],
      "exit": 0
    },
    {
      "row": "12",
      "result": "held",
      "exit": 0,
      "log": "row-12.log",
      "command": [
        "go",
        "test",
        "./cmd/curator",
        "-count=1",
        "-timeout=4m",
        "-v",
        "-run",
        "^TestNativeSkills(CLIOperations|TakeoverFlagBoundary|TakeoverEveryRegisteredCommand|TakeoverArgumentDiscardRegression|TakeoverActivationBoundary|TakeoverRegistryNarrowingMutant)$"
      ]
    }
  ]
}
```

## Identity and freshness

Base `5d8e6b07f7f485557d9bbb065c9692348ca065c2`; fresh fetched protected trunk `b9caa20901ee859eec20e649940693033e6297d4`. Candidate/upstream changed-path intersection is empty. All 17 candidate working blobs equal the published tree. Root README and CHANGELOG unchanged. Signed snapshot `8c5043959ce201bacb9ee73ce8d66d0d3a1f10e4` has exact candidate tree and verified signature G. Repository delta is present. No product/source/test file was edited by this reviewer.

## Prior finding and ordered sweep

Revision-6 `unrelated-takeover-argument-discard` is fixed: `cli.run` invokes the single admission check before help returns. True repair and false dry-run are required. The original help witness, full help matrix, list regression and both help narrowing plants passed before the native sweep. The four-file rework delta was inspected first.

Coverage: **12 of 12 surface rows driven; 5 of 5 AC rows driven**. The producer coverage map matches the named public entries exercised: UseWithPolicy, SyncWithPolicy, Install/Update, Resolve, StatusOf and CLI run.

Paths without a prefix are under internal/envprofile. Commands use count=1. Every log is in BUG-261018-uoisfe_review-evidence-rev7.txt.

| Row | Result | Production/test location | Command | Exit |
|---|---|---|---|---|
| 1 | held | switch.go:690; native_skills_test.go:48 | `go test ./internal/envprofile -count=1 -timeout=3m -v -run '^TestNativeSkillsUseAndSync$'` | 0 |
| 2 | held | managed.go:2568; native_skills_test.go:266 | `go test ./internal/envprofile -count=1 -timeout=3m -v -run '^TestNativeSkillsRepair$'` | 0 |
| 3 | held | switch.go:800; native_skills_test.go:70 | `go test ./internal/envprofile -count=1 -timeout=3m -v -run '^TestNativeSkillsTakeover$/^foreign$'` | 0 |
| 4 | held | switch.go:811; native_skills_test.go:70 | `go test ./internal/envprofile -count=1 -timeout=3m -v -run '^TestNativeSkillsTakeover$/^(directory|file|store|relative-store)$'` | 0 |
| 5 | held | native_skills.go:31; native_skills_test.go:180 | `go test ./internal/envprofile -count=1 -timeout=3m -v -run '^TestNativeSkillsLinkedParent$'` | 0 |
| 6 | held | switch.go:863; native_skills_test.go:317,500 | `go test ./internal/envprofile -count=1 -timeout=3m -v -run '^TestNativeSkills(ConvergenceAndSwitch|EmptyProfileLeavesUnnamedParent)$'` | 0 |
| 7 | held | native_skills.go:53; native_skills_test.go:200,452 | `go test ./internal/envprofile -count=1 -timeout=3m -v -run '^TestNativeSkills(Ledger|UnrecordedLedger)$'` | 0 |
| 8 | held | switch.go:863; native_skills_test.go:317 | `go test ./internal/envprofile -count=1 -timeout=3m -v -run '^TestNativeSkillsConvergenceAndSwitch$'` | 0 |
| 9 | held | switch.go:690; native_skills_test.go:372 | `go test ./internal/envprofile -count=1 -timeout=3m -v -run '^TestNativeSkills(InstallAndUpdate|ConvergenceAndSwitch)$'` | 0 |
| 10 | held | switch.go:1013; native_skills_test.go:424 | `go test ./internal/envprofile -count=1 -timeout=3m -v -run '^TestNativeSkillsBackupCollision$'` | 0 |
| 11 | held | switch.go:707; native_skills_test.go:250 | `go test ./internal/envprofile -count=1 -timeout=3m -v -run '^TestNativeSkillsUnreadableMarker$'` | 0 |
| 12 | held | cmd/curator/main.go:283; command_help.go:128; native_skills_flags_test.go:17,102,127 | `go test ./cmd/curator -count=1 -timeout=4m -v -run '^TestNativeSkills(CLIOperations|TakeoverFlagBoundary|TakeoverEveryRegisteredCommand|TakeoverArgumentDiscardRegression|TakeoverActivationBoundary|TakeoverRegistryNarrowingMutant)$'` | 0 |

## Controls, free hunt and exclusions

`CURATOR_TEST_CLI_HELPER=1 go test ./cmd/curator -count=1 -timeout=3m -v -run '^(TestReviewNativeTakeoverHelpBoundary|TestNativeSkillsTakeoverHelpCrossProduct|TestNativeSkillsTakeoverHelpNarrowingMutant|TestNativeSkillsTakeoverArgumentDiscardRegression)$'` exited 0 (rework.log). The help matrix covers 595 rows/2975 consent invocations. Both help mutants fail with the intended actual forbidden exit 0, not a compile failure.

`go test ./internal/envprofile -count=1 -timeout=5m -v -run '^Test(NativeSkillsNarrowingMutants|MusePublishedRegistryLayout|MuseDirectoriesAndIsolation|OpencodeFormFallback)$'` exited 0 (controls.log). All twelve narrowing plants were killed by their named behavioral witnesses. Required gate rows: 3 dangling-foreign-consent; 4 plain-file-consent; 5 skills-parent-traversal; 6 unnamed-takeover-directory; 7 one-ledger-name; 11 directory-marker-is-absence. The list-only mutant also passed in row 12. This is behavioral execution, not a static-token-only control.

Free hunt: `CURATOR_TEST_CLI_HELPER=1 go test -overlay .temp/review-r7/probe-overlay.json ./cmd/curator -count=1 -timeout=2m -v -run '^TestReviewR7RepeatedActivation$'` exited 0. The attached scratch probe exercises repeated conflicting boolean aliases, both last-value directions, prefix help and help topics, plus unchanged fixture state. No finding reproduced. The overlay adds only a reviewer test against exact candidate source blobs.

Native fixture `nativeSkillsFixture` pins temporary operator/native homes; CLI cases call `profileHome` and `pinOperatorHome`. No product invocation was directed at real HOME. Development tools' ordinary config/cache reads are not claimed absent. Muse/opencode regression checks passed; managed-home layout remains unchanged (managed.go only adds native repair hooks). Open spec questions remain Muse multi-root ownership, adapter-ledger arbitration and opencode global ownership.

## Validation evidence reused and checked

Hosted run https://github.com/relux-works/curator/actions/runs/39158440563 freshly reports completed/success on exact signed head `8c5043959ce201bacb9ee73ce8d66d0d3a1f10e4`: 20 successful jobs, two workflow-conditional skips. No infrastructure failure is represented as passing.

Downloaded board resources BUG-261018-uoisfe_evidence-manifest-r7.json and BUG-261018-uoisfe_validation-logs-r7.txt; all 71 payload SHA-256 values match. Complete merged stream contains zero fail events and 85 terminal packages. Producer current receipts include 15 fresh CLI groups, envprofile groups 8/11, three install binary consumers and full scriptworker/scriptpolicy/crossconformance; other 79 packages and unchanged envprofile/install cases are reused with source/test/configuration/environment identities from signed 3e00fcc8abe75e411a88773c65c4042ae2448eb0. The four-file identity delta matches git; changed consumers were rerun. 3497 top-level names include documented skips, not an assertion that skipped cases executed. Build/vet/lint/format/platform-case checks carry exit 0; missing-group control has intended exit 1. Historical exploratory failures are separated from current receipts. The reviewer independently reran targeted sweep, controls and free hunt, not the entire composed suite.

## Logbook

Revision 7 closes the prior help short-circuit finding with one pre-dispatch admission rule. All twelve rows held, required narrowing controls killed, free hunt empty. Exact signed hosted gate and local composed evidence verified. Acceptance means ready for manual PR landing, not landed or done. No branch mutation or landing was performed.
