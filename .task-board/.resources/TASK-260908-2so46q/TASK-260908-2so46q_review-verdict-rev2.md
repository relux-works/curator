# CR2 independent review — accepted

TASK-260908-2so46q / CR-TASK-260908-2so46q-2 revision 2.
Exact candidate tree: 7a63d60ecef47701de12ffe720f45854dddc4f7b.
Base and unchanged Story HEAD: 289ff42f037b9f86411fe7852000c466b3fe970d.
All six candidate paths match the published tree byte-for-byte. Relative to
CR1 tree d339d0b244706bb04e7e004c48de19aff4d85845, only plan_test.go and
plan-mutants.py changed; production semantics, module pin and README are unchanged.
Review and attacks used a git-archive copy under .temp/review-cr2/candidate.
No Story source changes, commits, branch operations, hosted CI, real ax,
provider execution, installs, restarts, runtime-home or LOGBOOK writes.

## F1 resolved

TestModelNotDrivenBySystem drives plan.Build -> real v0.5.11
vendorplugin.BuildLaunch with pi-openai/gpt-6-astra/high. It requires RefusedError,
ErrModelNotDrivenBySystem via errors.Is, retained Cause and runtime/system/model
detail, zero returned plan, one Interactive launch call and zero limits reads.
The shipped P13 preserves the propagation gate but exempts exactly that error
class. Independently running the behavioral suite kills it: Go exit 1, named
TestModelNotDrivenBySystem fails. Original package suite exits 0. P1-P12 remain
present and independently killed too. This resolves the CR1 surviving mutant.

Unresolved-vendor bound is valid for this fixed mapped/default-registry scope:
v0.5.11 runtime.go frozen declarations assign known vendors to claude/codex and
all three native Pi runtimes. Sole frozen VendorUnresolved runtime muse is
unmapped and its system is not imported here. Registry.ResolveRuntime
(registry.go:644-669) checks system registration first. TestUnresolvedVendorScope
therefore correctly demonstrates ErrRuntimeSystemUnregistered through plan.Build,
not ErrRuntimeVendorUnresolved. spawn.go:294-326 would absorb the unresolved
vendor for the actual muse declaration with its nonempty museModels, if its
system were present. This is not a claim about arbitrary custom registries or
unresolved declarations without model rows.

Unsupported-mode bound is valid: plan.Build:222 fixes Interactive and exposes no
mode input; the exact three mapped systems declare it in their tagged
Capabilities. TestInteractiveDeclaredForMappedSystems checks that declaration;
TestTaggedInteractivePlans drives successful Interactive calls for all three
mapped environments. The declaration-only test is not counted as a behavioral
refusal test. ErrUnsupportedLaunchMode is unreachable in the stated fixed scope.
Neither bound is a synthetic positive admission or a reopened product decision.

## Coverage accounting

14 of 16 AC rows driven at production package API, with row 15's unreachable
subclasses explicitly bounded above. Row 14 is separate main integration;
row 16 is lifecycle/delivery, not executable behavior. Named tests are in the
published candidate snapshot, not yet committed: managed Story policy requires
producer-side signed integration after this acceptance. No 16/16 main or
committed-delivery claim is made.

| Row | Production call site | Named tests / evidence |
|---|---|---|
| 1 | plan.Build -> vendorplugin.BuildLaunch | TestTaggedInteractivePlans, TestNativePiRuntimes; mapped systems |
| 2 | plan.Build -> SpawnRequest | TestSpawnRequestShape; resolved model/effort unchanged |
| 3 | plan.Build -> SpawnRequest / Availability | TestRequiredInputs, TestTaggedInteractivePlans; managed Home |
| 4 | plan.Build -> SpawnRequest | TestSpawnRequestShape, TestRequiredInputs; supplied WorkDir |
| 5 | plan.Build -> BuildLaunch | TestSpawnRequestShape, TestTaggedInteractivePlans; supplied inherited Env |
| 6 | plan.Build -> SpawnRequest | TestSpawnRequestShape; empty Composition, zero Run, optional inputs unset |
| 7 | plan.Build -> BuildLaunch | TestTaggedInteractivePlans, TestNativePiRuntimes; Interactive argv goldens, no bypass/child, stdin unattached |
| 8 | plan.Build -> BuildLaunch / RefusedError | TestTaggedAdmissionRefusals; sentinel/detail, effort guidance, terminal refusals |
| 9 | plan.Build -> Availability | TestTaggedInteractivePlans, TestNativePiRuntimes, TestRealStoreIsolation; exact single runtime/model/Home read |
| 10 | plan.Build -> Serviceable / LimitedError | TestProviderVerdicts; all non-serviceable classes and structured/text evidence |
| 11 | plan.Build -> Availability / RefusedError | TestProviderReadError, TestTaggedAdmissionRefusals; terminal failures, no retry/downgrade |
| 12 | plan.Build -> DefaultDeps -> Store.AvailabilityFor | TestRealStoreIsolation; determinate absence, corruption, profile/model-group isolation |
| 13 | plan.Build -> tagged BuildLaunch | TestTaggedInteractivePlans, TestNativePiRuntimes; all 3 environments and 3 native Pi runtimes |
| 14 | main tracked/untracked callers | Stated bound: no current callers; TASK-260908-1o7i8y owns current cwd/os.Environ and orchestration |
| 15 | plan.Build -> BuildLaunch refusal propagation | TestModelNotDrivenBySystem + P13; TestUnresolvedVendorScope; unsupported-mode/unresolved-vendor bounds above |
| 16 | CR/review/signed integration | CR2 published; accepted by this review; integration remains producer-owned |

## Independent validation and accepted evidence

Fresh Go 1.25.5 darwin/arm64 package suite:
`go test ./internal/plan -count=1 -v` exit 0 (package-01.log).
Fresh `python3 .scripts/plan-mutants.py ../mutants` in scratch candidate exit 0
(mutants-01.log): 13/13 killed, each Go exit 1 with its named behavioral failure.
The harness runs the whole package behavioral suite with -count=1, not a static
checker. No production gate inspects source text; token-preserving source-checker
requirement is inapplicable. Both scratch and Story production bytes were verified
against the tree after attacks. `git diff --check` exit 0.

Fresh `go list -m -json github.com/relux-works/skill-agents-management` exit 0
resolves v0.5.11 without Replace; GOWORK empty. Pinative and vendor registration
imports are real. Signed tag identity is accepted from primary publication and
CR1 evidence, not independently re-fetched/re-verified here: tag object
0ea486e46765ecf12fff7c2ac526e12da02e95ed, peeled commit
a2a6e9f377f62a5872d99ecdfff0d1690e385f2a. No pseudo-version or workspace override.
Accepted A0 sections 3.2-3.5 and current SPEC4.4 were read; obsolete Pi-wrapper
limitations are superseded by real native-Pi positive tests.

Accepted, not rerun: attached change-request_rev2-validation.log records runtime
`make check` exit 0: build, fmt-check, vet, all 10 packages in tests and race.
R1-rework reports only narrow producer commands. Historical native-pi-validation
reported a manual full suite in addition to CR1 runtime full suite: that duplicate
remains acknowledged, not rewritten as one execution. Reviewer ran no full suite.
No changed production semantics warrant repeating all adjacent package validation.

All 19 checklist entries reviewed: package requirements satisfied; source-token
check is inapplicable; logbook facts are recorded here because control-root writes
are prohibited; the negative-verdict branch is inapplicable to acceptance.
Producer evidence and independent reviewer bundle are attached before accept_cr.
Verdict: accepted for owned package requirements. Route CR2 via accept_cr to
integrating, never done; signed integration and main wiring remain separate.
