# TASK-260908-fdg3gn: a1-system-prompt-delivery — CR1 review

Verdict: accepted. No blocking findings in the assigned reusable API scope.
Candidate tree: 2ce6096f89e9dd0e401e45e53627edf5ceed3548.
Base/worktree HEAD: 84747c326eee9863ddfd7e86ac65be1056718fbc.
The run is not goal-bound (queried at start and before verdict).
Acceptance is for this candidate, not signed delivery or main integration.

## Scope and source checks

Reviewed all four changed paths against SPEC §5 including E5, curator-spec
protocol/environments.md at d019f0e §7.3, production fragment.Parse and its closed
registry, and attached accepted A0 TASK-260908-qblycn_a0-verification-findings.md.
Each changed working file matches its candidate-tree blob. The three producer
code/test/harness SHA-256 values match both publication evidence and the original
recovery manifest. README retains the composition API and its late-check
obligation as well as the new systemprompt obligation. HEAD..main count was 0;
this is local-ref evidence, not a fresh remote-main attestation.

Selector applies only explicit matching non-file semantics; unsupported/absent
channels refuse. Exact Codex TOML quoting and declared Pi/Claude path argv are
covered; no variable descriptor exists, Env is empty, and no registry extension
or provider-permission flag was added. Pi selected-path readability prevents
missing paths becoming literal prompts. PrepareLaunch invokes fresh ProbeFiles
and FormatWarnings; home probing is independent of descriptors. Suppression
does not bypass readability checks, and opposite-semantics files remain
conditional native-discovery candidates. Warnings preserve profile, semantics,
kind/filename, conditional replacement effects, caching and billing.

Caller search confirms main does not yet invoke these APIs. This is the explicit
assignment bound, not a claimed completed execution guard. Future tracked and
untracked launch wiring must invoke PrepareLaunch immediately before each
handoff/exec, refuse errors and emit all warning lines. No native argv inspection,
project discovery, file writes, second composer or native-source attestation was
introduced. Existing validated-fragment preconditions remain owned by Parse.

## AC coverage: 11 of 12 rows driven; one explicit delivery bound

Checked the producer's ratio before code review and independently matched rows
to the candidate tests. Tests are snapshotted candidate source awaiting the
producer/parent signed-delivery transaction; the reviewer made no commits.

| Row | Production entry | Named driver / bound |
|---|---|---|
| 1 opt-in only | Select | TestSelectRefusalsAndNoOptIn |
| 2 semantics/non-file/unavailable | Select | TestSelectRefusalsAndNoOptIn |
| 3 exact argv/env | Select, Selection.Argv/Env | TestSelectExactArgv |
| 4 every-call independent probe | PrepareLaunch, ProbeFiles | TestPrepareLaunchLateFilesAndWarnings; TestProbeIndependentAndNoWrites |
| 5 absence vs failure | PrepareLaunch, ProbeFiles | TestPrepareLaunchFileFailures; TestProbeParentFailures |
| 6 E5 suppression plus validation | PrepareLaunch, FormatWarnings | TestPrepareLaunchLateFilesAndWarnings; TestPrepareLaunchFileFailures |
| 7 conditional discovery | PrepareLaunch, FormatWarnings | TestPrepareLaunchLateFilesAndWarnings |
| 8 warning content | PrepareLaunch, FormatWarnings | TestAppliedReplacementWarning; TestPrepareLaunchLateFilesAndWarnings |
| 9 filesystem/late changes | PrepareLaunch | TestPrepareLaunchSelectedPathChangesLate; TestPrepareLaunchFileFailures; TestPrepareLaunchLateFilesAndWarnings |
| 10 no writes | PrepareLaunch, ProbeFiles | TestProbeIndependentAndNoWrites; TestPrepareLaunchLateFilesAndWarnings |
| 11 closed support | Select | TestSelectRefusalsAndNoOptIn; TestSelectExactArgv |
| 12 reviewed signed delivery | integration transaction | Reviewer acceptance recorded here; signing/publication/landing remain producer/parent-owned |

Main integration is 0 of 2 modes by the explicit API-only task scope. Codex
per-invocation model_instructions_file behavior remains A0 docs-confidence;
exact encoding is tested, no actual model launch was attempted.

## Independent verification and attacks

- go test ./internal/systemprompt -count=1 -v: exit 0, all named API/filesystem
  cases passed, no skips (non-root macOS arm64, Go 1.25.5).
- Eight discriminating narrowing mutants executed sequentially on an archive of
  the exact CR tree under .temp, never on the Story source. Every mutant ran the
  behavioral package suite with -count=1 and exited 1 at its named assertion.
- git diff --check: exit 0. Original Story files remain unchanged by review.
- Accepted the attached exact-CR runtime make check log, exit 0 (build, formatting,
  vet, tests and race); did not repeat the broad suite without a new concern.
- Inspected producer zip summary: 11 of 11 killed. M1/M4/M7 are accepted from
  producer evidence; the eight below were independently rerun.

| Mutant | Shape / weakened behavior | Named failing test |
|---|---|---|
| M2 | bypass: Codex append accepts replace | TestSelectRefusalsAndNoOptIn/codex_cli/section/append |
| M3 | absent usable channel treated as satisfied: file-only replace | TestSelectRefusalsAndNoOptIn/pi/section/replace |
| M5 | bypass: missing section disables registry probe | TestPrepareLaunchLateFilesAndWarnings/no-section |
| M6 | failure reported as absence: dangling target | TestPrepareLaunchFileFailures/APPEND_SYSTEM.md/dangling |
| M8 | failure admitted: permission-denied open | TestPrepareLaunchFileFailures/APPEND_SYSTEM.md/permissions |
| M9 | absence admitted: selected Pi prompt missing | TestPrepareLaunchFileFailures/selected/missing |
| M10 | unsupported fact: opposite semantics claimed suppressed | TestPrepareLaunchLateFilesAndWarnings/section |
| M11 | bypass: suppressed append file skips refusal | TestPrepareLaunchFileFailures/APPEND_SYSTEM.md/directory |

No survivors in this subset; this is not an exhaustive mutation census. Stable
FIFO/nonregular, dangling/loop and readable symlinks, missing files, file/parent
permissions, late creation/removal/replacement were driven through exported APIs.
Residuals: probe-to-open/exec races, deterministic syscall I/O/close failures,
Linux execution, native source selection and actual CLI warning delivery are not
proven here. Readable means regular and openable, not full content consumption.

No production static/source-token gate exists; the conditional token-preserving
static-gate requirement is inapplicable. The mutation script only injects changes
and invokes behavior tests. LOGBOOK writes were explicitly forbidden; this board
outcome is the required persistent substitute. Non-acceptance routing checklist
is inapplicable because this verdict accepts.

Operational read failures: initial resources(...) query was unsupported; recovered
with get(...){outcomeResources} and resource get. Assumed registry.go path did not
exist; rg located registry in fragment.go. No absence conclusion relied on either
failed read. Tool readiness and actual execution logs are in the attached review
zip. No source changes, commits, installs, CI, real models, managed homes or
private/control-root records were written.
