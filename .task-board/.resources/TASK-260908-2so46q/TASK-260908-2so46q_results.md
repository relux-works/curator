# TASK-260908-2so46q — partial implementation, release-blocked

## State and provenance

Candidate remains UNCOMMITTED in the assigned Story worktree at checkpoint
289ff42f037b9f86411fe7852000c466b3fe970d. HEAD..main was 0 when inspected.
No branch switch, commit, install, tag, release, runtime-home edit, control-root
source write, real ax invocation or provider execution was performed.

Fallback continues the existing Muse partial internal/plan files. Read the
original RUN-260909-b70807 and last successor RUN-260909-83a1ac attached logs:
both end in transport_connect_unreachable / meta response transport failure
after four provider attempts. c6584a and 752162 failures are accepted from the
provided fallback precondition, not independently rerun. This producer used
Codex gpt-6-astra medium per explicit fallback authorization. Initial source
copies are retained in the evidence archive. The inherited test file did not
compile (three range variables); its claimed 12/13 coverage and mutant results
were comments with no executed tests and have been removed.

Accepted inputs read: task's accepted-plan-contract/current-resume/fallback,
SPEC 0.3.0 sections 4.2–4.4, epic goal A1 and common producer rules, and accepted
A0 findings sections 3.2–3.5. A0 installed-native-tool/fragment observations are
accepted evidence, not rerun. This Go CLI targets Darwin/Linux; fresh checks
here use darwin/arm64 and the installed Go toolchain recorded in readiness.

## Candidate changes

- internal/plan/plan.go: retains and completes the partial API. Build calls the
  tagged vendorplugin.BuildLaunch Interactive, then the separately supplied
  Store.AvailabilityFor, using exact resolved Runtime/Model/managed Home.
  Empty Home/WorkDir and absent dependency functions refuse. DefaultDeps(nil)
  now leaves the reader absent so Build refuses instead of panicking.
- RefusedError unwraps module errors and adds --effort guidance for required
  effort; LimitedError retains the typed verdict and renders observation
  timestamps and nanosecond Until precision, which the partial code omitted.
- No legacy Pi plugin registration: its wrapper does not satisfy native Pi.
- internal/plan/plan_test.go: actual tagged Claude/Codex calls, request-shape
  capture, exact argv goldens, no-child marker, failure/call-count tests,
  real Store fixtures for absent/corrupt/limited and profile/model-group isolation.
- .scripts/plan-mutants.py: source-restoring behavioral harness.
- README.md: documents API, exact commands, evidence location and pending wiring.

## Coverage, from SPEC 4.4 plus explicit assignment gates

**12 of 16 AC rows driven at the production package API by named candidate
tests; 4 stated bounds below. 0 of 16 rows delivered as committed/accepted
coverage: this managed candidate is uncommitted and no complete CR is published.**
Tests of the package are not evidence that main calls it.

| Row | Required behavior | Production call site / named test | Result |
|---|---|---|---|
| 1 | §4.2 Claude/Codex runtime matches mapped system | plan.Build → vendorplugin.BuildLaunch; TestTaggedInteractivePlans | driven |
| 2 | Explicit resolved model/effort unchanged | plan.Build → SpawnRequest; TestSpawnRequestShape | driven |
| 3 | Managed Home passed; blank Home cannot use native default | plan.Build → SpawnRequest and Availability; TestTaggedInteractivePlans, TestRequiredInputs/home_empty, home_whitespace | driven |
| 4 | Supplied WorkDir unchanged, required | plan.Build → SpawnRequest; TestSpawnRequestShape, TestRequiredInputs/workdir_empty, workdir_whitespace | driven; caller choosing current cwd is row 14 |
| 5 | Supplied Env unchanged; tagged child Env preserves inherited marker | plan.Build → BuildLaunch; TestSpawnRequestShape, TestTaggedInteractivePlans | driven; os.Environ wiring is row 14 |
| 6 | Empty Composition, zero Run, no goal/budget/tier/assignment/profile/engine | plan.Build → SpawnRequest; TestSpawnRequestShape | driven |
| 7 | Named Interactive; exact argv without bypass; unattached stdin; no child | plan.Build → BuildLaunch; TestTaggedInteractivePlans | driven for available two adapters |
| 8 | Module refusal retained, missing-effort guidance, invalid/empty model/runtime, invalid effort, cancellation/nil registry terminal | plan.Build → BuildLaunch, RefusedError; TestTaggedAdmissionRefusals | driven |
| 9 | Separate single Store read keyed by exact runtime/model/Home | plan.Build → Availability; TestTaggedInteractivePlans; DefaultDeps → Store.AvailabilityFor in TestRealStoreIsolation | driven |
| 10 | Only Serviceable; Unknown/Limited/Unreachable/invalid state refused with full typed/text evidence | plan.Build → Serviceable, LimitedError; TestProviderVerdicts | driven |
| 11 | Reader error terminal, no retry/default/downgrade; no executable result | plan.Build → Availability; TestProviderReadError, TestTaggedAdmissionRefusals (exact single calls) | driven; reader-error branch uses injected error, not manufactured live-store failure |
| 12 | Determinate absence admits; corruption refuses; profile and model group isolate state | plan.Build → DefaultDeps → real Store.AvailabilityFor; TestRealStoreIsolation | driven |
| 13 | All three adapters at operator native-Pi tag | Actual v0.5.11 absent; TestPiReleaseBound only exercises unavailable runtime at v0.5.10, not native Pi | BLOCKED |
| 14 | Both main tracked/untracked entry paths use current cwd/os.Environ, before any execution | main remains not_implemented; TASK-260908-1o7i8y owns wiring | stated bound |
| 15 | Other module admission classes (model not driven, unsupported mode, unresolved vendor) through real registry fixtures | delegated to BuildLaunch; not independently exercised by this package suite | stated bound |
| 16 | Named committed tests, complete publication, independent Astra-medium acceptance, signed delivery | no complete CR until row 13; parent owns integration | pending |

## Negative evidence

12/12 mutants killed by named behavioral test failures (each Go test exit 1,
harness exit 0); see mutants-02/summary.tsv and per-mutant logs. P1/P2 narrow
whitespace validation; P3/P4 admit a missing dependency through a default for Codex only;
P5 accepts only missing-effort BuildLaunch errors; P6 admits Unknown;
P7 accepts exactly one reader-error text. P8–P10 substitute a Home, Model or
Runtime in the limits query. P11/P12 substitute effort for exactly one required
model or invalid word. No production gate scans source text, so token-preserving
source-checker attack is inapplicable. Harness anchors edit source but acceptance
requires the named behavioral suite failure, never an anchor/token check.
The harness restores byte copies after each mutant; no Git reset/restore is used.

## Fresh validation (this producer)

| Command | Exit | Evidence |
|---|---:|---|
| Initial go test ./internal/plan -count=1 | 1 | baseline-01.log; inherited syntax error |
| First completed suite | 1 | test-02.log; fixture assumed wrong module model-group key |
| go test ./internal/plan -count=1 -v | 0 | test-03.log; fixture corrected from actual module GroupTable |
| python3 .scripts/plan-mutants.py .temp/TASK-260908-2so46q/mutants-02 | 0 | mutants-02.log; 12/12 killed |
| go test ./internal/plan -count=1 -race -cover | 0 | race-02.log; restored final candidate |
| go vet ./internal/plan | 0 | vet-01.log |
| go build ./internal/plan | 0 | build-01.log |
| go test ./internal/diagnostics ./internal/mapping -count=1 | 0 | adjacent-01.log |
| git diff --check and gofmt -l internal/plan | 0 | final-checks.log; no format output |

No full make check or hosted CI was run; configured runtime check is not duplicated.
No unrelated suite success is claimed. Package statement coverage was 90.4% in
race-01; final race-02 records the value after strengthening the shape test.

## External blocker and resumption

Exact command: git ls-remote https://github.com/relux-works/skill-agents-management.git
refs/tags/v0.5.11 'refs/tags/v0.5.11^{}'. Fresh read exit 0, zero advertised refs
(tag-01.log, repeated with command/exit metadata in tag-02.log). This proves the
requested tag absent at those reads; it does not claim native support absent
from upstream main. Existing v0.5.10 BuildLaunch refuses pi-native as undeclared.

Failed assumption: legacy system pi or system id pi-native can supply the required
native-Pi runtime on v0.5.10. Neither is accepted evidence. No synthetic runtime,
pseudo-version, replace, workspace override or agent-created tag was attempted.

Clean progress is preserved. Recommendation: operator publishes the intended
v0.5.11 native-Pi release. Then the next producer verifies advertised/peeled
identity, pins the real tag, uses its declared native runtime/model/effort,
registers its real system, replaces the refusal-only Pi bound with native tagged
plan/golden tests, reruns narrow gates, publishes the complete CR and routes the
independent Astra-medium reviewer. Main wiring remains a separate task.
Keeping v0.5.10 proves two adapters but cannot discharge the three-adapter AC;
substituting unreleased code or legacy wrapper would violate the accepted scope.
Exact external input needed: operator-created actual native-Pi v0.5.11 tag.

## Logbook handoff

Parent may append: resumed failed Muse producer under authorized Codex fallback;
repaired uncompilable partial tests and unverified coverage claims; real plan and
limits API tests now pass and 12 narrowing mutants fail as expected. Native-Pi
release remains the only blocker to all-three package validation. No control-root
LOGBOOK.md writes were made by this run. Leave task blocked, candidate uncommitted,
and do not route independent acceptance until complete publication is possible.

Final race-02 passed with 90.4% statement coverage. Candidate path hashes are in
candidate-sha256.txt. The attached patch is preservation evidence only, not a
Change Request or an accepted delivery. Handoff was deliberately not invoked:
its AC preconditions are unmet and current-resume forbids complete publication
until the required operator tag exists.
