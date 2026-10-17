# BUG-261017-ygfdal — review revision 1

Candidate tree: `103755136899324f1401da18113bd6cc5c17f3e7`.
Base: `612960ee90bedc7a4a113eda13aaa6cf2a2cbe9f`.
Hosted snapshot: `23a4f4ee7dee14319f6215df7073dcef5b2ba900` (identical tree).

Coverage was checked before source review: 4 of 5 AC rows have named behavioral drivers; AC5 has the same entry regression plus the explicitly external full-CI obligation. No AC is omitted. Call sites and tests are enumerated in the producer handoff. Review B's six numbered requirements form the surface table because the brief supplied no catalog-specific table or time budgets. Linux-only attack evidence is reused only after exact-tree and full-stream verification; local execution is explicitly distinguished.

## Surface sweep

| Row | Result | Attack and evidence |
| --- | --- | --- |
| 1 — affected binary entry points | held | Independently derived all five affected binaries from `go list -test -json ./...`, normalizing test-package variants. Matches cmd/curator, internal/install, internal/install/atomicity, internal/crossconformance, internal/scriptworker. Reran TestLandlockProbeModeTerminatesAffectedTestBinaries locally, exit 0, including missing directory, missing argument and extra argument on every binary. No stdout/stderr and no test execution. |
| 2 — recursion guard / hidden mode / environment | held | Reused exact hosted TestLandlockProbeChildCannotReprobe plus the malformed-child narrowing mutant: argc=2 guard bypass is killed by the named test. Production callers verified: inventory_unix.go -> probeScriptLandlock -> landlockProbeCommand; init -> RunLandlockProbe. Local binary attacks cover malformed requests; environment constructor remains byte-for-byte unchanged and is called by the actual probe constructor. No product marker added. |
| 3 — Linux descendant cleanup | held | Independently downloaded and parsed complete hosted stream: TestLandlockProbeCancellationKillsProcessGroup passes both deadline and explicit cancellation through the exact production command constructor. Reused cleanup-kills-group-only-on-deadline narrowing attack; behavioral diagnostic confirms a surviving descendant when only the direct child is killed. Positive normal exit also passes. This is hosted execution, not local Linux execution. |
| 4 — red baseline / green candidate | held | Independently reran the final regression with an overlay restoring baseline production blobs: exit 1, TestLandlockProbeModeTerminatesAffectedTestBinaries/internal/install/owned reports `probe ran tests or hung: context deadline exceeded` and actual unrelated test output. Candidate run exits 0. This is a bounded macOS entry-point regression, not Linux fork exhaustion. |
| 5 — probe result preservation / full gate | held | Exact hosted stream contains passing TestLinuxLandlockProbeFindsPresentControls and TestLinuxLandlockProbeChildAnswersAtTheBoundary, including existing missing-path absence control. Local off-Linux binary probes return absent silently. Kernel-without-Landlock execution was not newly reproduced; ABI/expected-absence code is unchanged. Fresh hosting API confirms run 38997554340 success on the exact tree, 20 successful jobs and 2 optional skips. |
| 6 — public text / scope | held | Scanned added lines for personal paths, addresses and IP patterns, then manually reviewed all eight-file additions; no host identifiers, personal names or client material found. Only product/test files and the required incident logbook entry changed. Runner registration names are confined to the requested evidence. |

## Free hunt

Built the actual production binary and drove five additional malformed/boundary argument vectors: missing directory, empty directory, empty surplus argument, repeated mode token, and 128 surplus values. Each exited 1 silently within a 3-second watchdog. Actual `--help` exited 0 without the hidden token. No findings. These off-Linux checks establish dispatch/silent refusal, not Linux syscall behavior. TestProductionBinaryDispatchesLandlockProbe independently reran successfully (63.990s including the existing shared host-toolchain lock wait).

## Evidence identity and limits

- Fresh origin/main fetch still resolves to the CR base, so no upstream overlap or convergence is required.
- All eight worktree candidate files match exact candidate blobs. Remote snapshot tree equals candidate tree. No product/test source was edited by this reviewer.
- Independently downloaded artifact 12087593346 from run 38997554340. SHA-256 of the complete served stream is `aa277143d2e1112ae4f32ab71acf0ed04a912cf0a8bf1f4b796850d221e1ae52`, matching attached producer evidence. All 83 started packages have terminal records; zero failing records; seven required named tests, twenty binary cases, and all three exact narrowing-kill diagnostics verified.
- Linux Test ran on registration `relux-rise1-curator-2`; Linux qualification and gate self-test jobs used `relux-rise1-curator-1` / `relux-rise1-curator-2`. Existing Linux Race used `GitHub Actions 1000035633`. This is reported as the existing workflow scope bound, not a claim that every Linux job is self-hosted. Task scope forbids workflow/host changes.
- The reviewer reran macOS dispatch, the baseline regression, production dispatch, direct argument attacks, production build and whitespace validation. The reviewer did not rerun the full gate or Linux-only tests; exact-tree hosted results and behavioral mutant outputs were verified and reused. No local Linux VM was available.
- Process-group coverage assumes descendants remain in the inherited group. A new-session malicious executable is outside the fixed self-probe model; the real init dispatcher runs no external executable.

## Verdict

```json
{
  "verdict": "accepted",
  "revision": 1,
  "candidate_tree": "103755136899324f1401da18113bd6cc5c17f3e7",
  "findings": [],
  "notes": [
    "The binary regression enumerates five current packages explicitly. Independent dependency inventory confirms 5 of 5 today; a future importer would require updating the test list even though shared init dispatch automatically protects it.",
    "No Linux-without-Landlock kernel was executed in this review; expected-absence implementation is unchanged, missing-path and off-Linux absence controls pass.",
    "Full gate and Linux-only attack results are exact-tree hosted evidence reused after independent API, digest and complete-stream verification."
  ]
}
```

Acceptance is limited to this candidate and these attacks; it does not claim absence of all defects. Accept through `accept_cr`; landing remains the producer/manual PR lifecycle.
