# TASK-260908-fdg3gn producer evidence

Candidate: uncommitted Story worktree at d97e6cb8 (SPEC5 E5 base). Ready for review.

## Scope and sources

Implemented `internal/systemprompt/systemprompt.go` and external-package API tests;
added `.scripts/systemprompt-mutants.py` and updated README with the pending main
call-site obligation. No main, SPEC, BuildLaunch, registry, managed-home/config,
LOGBOOK, control-root, CI, release, or runtime-install changes. No commits.

Read SPEC §5 at the candidate HEAD, production fragment types/registry,
curator-spec `protocol/environments.md` at d019f0e, and accepted A0 evidence
(`TASK-260909-1hznm7/accepted-A0-evidence.md`, originally TASK-260908-qblycn).
A0 Codex 0.153.4's per-invocation override remains docs-confidence; no native model
launch was performed. Exact TOML encoding is tested, including quotes, backslash,
control characters, DEL, Unicode, whitespace and shell-looking characters.
Pi 0.84.2 polymorphic path validation is explicitly required by environments §7.3.
There is no system-prompt variable descriptor in the closed registry, so Env is
empty; no synthetic adapter or future channel support was introduced.

`PrepareLaunch` is the reusable production boundary: it selects, validates the
selected Pi path, freshly calls `ProbeFiles`, and invokes pure `FormatWarnings`.
Callers supply a fragment validated by `fragment.Parse`; exported mutable fragment
structs are not a second untrusted protocol input to this package.

## AC coverage: 11 of 12 AC rows driven

Rows below decompose the supplied AC into independently observable behaviors.
Tests are candidate source to be committed by the parent, not producer commits.

| # | AC row | Production call site | Named test / bound |
|---:|---|---|---|
| 1 | Explicit opt-in only, no implicit argv/env | Select | TestSelectRefusalsAndNoOptIn |
| 2 | Exact semantics, non-file selection and unavailable refusal | Select | TestSelectRefusalsAndNoOptIn (Pi replace, Codex append, absent sections, invalid opt-in, opencode) |
| 3 | Channel argv/env encoding | Select / Selection.Argv / Selection.Env | TestSelectExactArgv; no declared variable input exists |
| 4 | Callable every-launch probe independent of descriptors | PrepareLaunch / ProbeFiles | TestPrepareLaunchLateFilesAndWarnings; TestProbeIndependentAndNoWrites |
| 5 | Absence distinguished from dangling, unreadable, nonregular and other errors | PrepareLaunch / ProbeFiles | TestPrepareLaunchFileFailures; TestProbeParentFailures |
| 6 | Same-semantics E5 suppression with validation retained | PrepareLaunch / FormatWarnings | TestPrepareLaunchLateFilesAndWarnings; TestPrepareLaunchFileFailures/APPEND_SYSTEM.md |
| 7 | Conditional discovery subject to native flags and trusted project | PrepareLaunch / FormatWarnings | TestPrepareLaunchLateFilesAndWarnings |
| 8 | Profile, kind, semantics, filename, replacement and cache/billing warnings | PrepareLaunch / FormatWarnings | TestAppliedReplacementWarning; TestPrepareLaunchLateFilesAndWarnings |
| 9 | Real filesystem and late changes | PrepareLaunch | TestPrepareLaunchSelectedPathChangesLate; TestPrepareLaunchLateFilesAndWarnings; TestPrepareLaunchFileFailures |
| 10 | No writes | Select / PrepareLaunch / ProbeFiles | TestProbeIndependentAndNoWrites; TestPrepareLaunchLateFilesAndWarnings (original bytes retained, only fixture entries exist) |
| 11 | No invented registry support | Select | TestSelectRefusalsAndNoOptIn/opencode; TestSelectExactArgv (empty Env); registry source unchanged |
| 12 | Reviewed signed delivery | Parent workflow | Bound: independent acceptance and signed delivery are parent-owned, not performed by this producer |

Main launch wiring coverage is explicitly **0 of 2 modes** (tracked/untracked).
This assignment requires a reusable API and excludes main integration. Neither
CLI stderr emission nor actual exec/ax application is claimed. README makes the
later every-launch invocation and warning emission obligations explicit.

## Validation executed by this producer

| Command | Real exit | Evidence |
|---|---:|---|
| `go test ./internal/systemprompt -count=1 -v` | 0 | test-01.log |
| `python3 .scripts/systemprompt-mutants.py .temp/TASK-260908-fdg3gn/mutants` first attempt | 1 | Harness anchor M6 matched twice; stopped and restored source; not a product gate pass |
| Same mutant command after anchor correction | 0 | mutants-02.log and mutants/*.log; 11 expected-red behavioral runs each exit 1 |
| `go test ./internal/systemprompt -count=1 -cover` | 0 | test-02.log; 91.3% statement coverage |
| `make check` | 0 | check-01.log; build, fmt-check, vet, tests, race |
| `git diff --check` | 0 | run after edits and self-review |

No permissions tests were skipped on this non-root macOS run. Tests explicitly
skip chmod-based denial under root rather than misrepresent denial as proven.
Both supported host targets are Go CLI targets; the configured local macOS check
was run, not iOS. Linux execution was not independently run here.

The first required board status command returned exit 1 because an estimate was
missing. Set estimate=5 then entered development successfully. An attempted
compact resources query returned exit 1 for an unknown field; used scoped schema
recovery and local attached evidence. These are operational failures, not green
gates. No new logbook write was made because the task explicitly forbids it.

## Narrowing-mutant evidence

Each mutant executes the behavioral package suite with `-count=1`; original bytes
are restored, never reset from Git. No production gate inspects source text, so
the additional token-preserving source-inspection-gate requirement is inapplicable.
The mutant harness itself uses anchors only to inject behavior changes; it does
not substitute a static checker for behavioral execution.

| Mutant | Narrowing | Named failing test | Exit | Survivor bound |
|---|---|---|---:|---|
| M1 | apply append without opt-in for Claude only | `TestSelectRefusalsAndNoOptIn/claude_code/section/` | 1 | None: killed |
| M2 | admit Codex append using replace descriptor | `TestSelectRefusalsAndNoOptIn/codex_cli/section/append` | 1 | None: killed |
| M3 | admit matching replace file as an opt-in no-op | `TestSelectRefusalsAndNoOptIn/pi/section/replace` | 1 | None: killed |
| M4 | admit append with absent section | `TestSelectRefusalsAndNoOptIn/pi/absent/append` | 1 | None: killed |
| M5 | skip registry probe only without section | `TestPrepareLaunchLateFilesAndWarnings/no-section` | 1 | None: killed |
| M6 | treat dangling target as absence | `TestPrepareLaunchFileFailures/APPEND_SYSTEM.md/dangling` | 1 | None: killed |
| M7 | admit directory only at regular-file check | `TestPrepareLaunchFileFailures/APPEND_SYSTEM.md/directory` | 1 | None: killed |
| M8 | admit permission-denied open only | `TestPrepareLaunchFileFailures/APPEND_SYSTEM.md/permissions` | 1 | None: killed |
| M9 | admit missing selected Pi prompt path | `TestPrepareLaunchFileFailures/selected/missing` | 1 | None: killed |
| M10 | claim suppression for opposite-semantics home file | `TestPrepareLaunchLateFilesAndWarnings/section` | 1 | None: killed |
| M11 | bypass file failure only for suppressed append file | `TestPrepareLaunchFileFailures/APPEND_SYSTEM.md/directory` | 1 | None: killed |

## Bounds and self-review

No surviving mutants in the executed set. Not an exhaustive mutation census:
post-open object replacement, syscall I/O faults, close failures, and adversarial
probe-to-open/probe-to-exec races are not deterministically injected. Readability
means a regular file that can be opened, not full content consumption; no writes
occur. Symlinks to readable regular files are valid, dangling links are failures.
Only registry-named Pi home files are observed. No project probe, trust decision,
complete native-source enumeration, absent-home-implies-no-customization claim,
or double-application assertion is made. The invalid-descriptor defensive arms
are unreachable for fragment.Parse inputs and do not add registry support.

Self-review: read the implementation/tests and README, verified source restoration,
reviewed candidate scope with git status, and ran diff whitespace validation.
Independent review, main integration, and signed delivery remain the parent lane.
