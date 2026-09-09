# Independent review — changes requested

Task TASK-260908-2so46q, CR-TASK-260908-2so46q-1 revision 1.
Reviewed tree d339d0b244706bb04e7e004c48de19aff4d85845 against base
289ff42f037b9f86411fe7852000c466b3fe970d. All six workspace candidate paths
match that tree byte-for-byte. HEAD remains the base; HEAD..main count 0.
No source edits, commits, branch operations, runtime-home edits or real ax.
Reviewer attacks were performed only in a git-archive scratch copy.

## F1 — required admission refusal coverage is incomplete (blocking)

internal/plan/plan.go:223 propagates BuildLaunch errors; the candidate
TestTaggedAdmissionRefusals omits ErrModelNotDrivenBySystem. This is a
reachable requirement of SPEC 4.4/6, not future main integration: the real
v0.5.11 rejects pi-openai / gpt-6-astra / high with this sentinel.

R1 narrows the existing propagation gate to:
`if err != nil && !errors.Is(err, vendorplugin.ErrModelNotDrivenBySystem)`.
All other error classes remain rejected. Running the complete candidate
package behavioral suite with -count=1 -v exits 0: this narrowing survives.
A scratch-only TestReviewerModelNotDriven calls production plan.Build with
DefaultDeps and that exact triple. It passes against original code (exit 0)
and fails against R1 (exit 1: incompatible model admitted or evidence lost:
<nil>). This proves a test gap, not a claim that the unchanged implementation
currently admits the request. Candidate tests cannot defend this required gate.

Rework: add the real tagged mismatch case to the shipped admission tests,
assert RefusedError, preserved sentinel/detail, zero returned plan, one launch
call and zero limits calls; ship a narrowing mutant for this class. Resolve
row 15's remaining unresolved-vendor and unsupported-mode omissions through
real admission fixtures, or give a precise evidence-backed unreachable bound
for the fixed Interactive/mapped-system scope. Merely delegating admission to
BuildLaunch is not evidence that the wrapper propagates those refusals.
Do not add synthetic positive Pi admission or broaden main integration scope.

## Coverage accounting

13 of 16 producer AC rows driven at the production package API; row 15 is
incomplete owned coverage, row 14 is legitimately separate main wiring, row
16 remains delivery/acceptance pending. Named candidate tests are not yet
committed; managed worktree policy correctly requires them uncommitted here.

| Rows | Production call site | Candidate tests / assessment |
|---|---|---|
| 1, 7, 13 | plan.Build -> real vendorplugin.BuildLaunch | TestTaggedInteractivePlans, TestNativePiRuntimes: all 3 environments and all 3 native Pi runtime goldens |
| 2, 4, 6 | plan.Build -> SpawnRequest | TestSpawnRequestShape, TestRequiredInputs: explicit pair, supplied cwd, zero optional fields |
| 3 | plan.Build -> SpawnRequest / Availability | TestRequiredInputs, TestTaggedInteractivePlans: managed Home and blank rejection |
| 5 | plan.Build -> BuildLaunch | TestSpawnRequestShape, TestTaggedInteractivePlans: supplied Env preserved; current os.Environ belongs to row 14 |
| 8 | plan.Build -> BuildLaunch / RefusedError | TestTaggedAdmissionRefusals: listed covered classes and effort guidance |
| 9 | plan.Build -> Availability | TestTaggedInteractivePlans, TestNativePiRuntimes, TestRealStoreIsolation: exact query triple |
| 10 | plan.Build -> Serviceable / LimitedError | TestProviderVerdicts: non-serviceable states and structured/text evidence |
| 11 | plan.Build -> Availability / RefusedError | TestProviderReadError, TestTaggedAdmissionRefusals: terminal errors and call counts |
| 12 | plan.Build -> DefaultDeps -> Store.AvailabilityFor | TestRealStoreIsolation: absence, corruption, profile/model-group isolation |
| 14 | main tracked/untracked callers | No callers yet; TASK-260908-1o7i8y owns this explicitly deferred integration |
| 15 | plan.Build -> BuildLaunch error propagation | F1: missing reachable mismatch refusal; other named classes unproven |
| 16 | CR/review/signed delivery | CR published; acceptance withheld; producer integration remains pending |

## Validation and provenance

Independent commands: go test ./internal/plan -count=1 -v exit 0;
R1 candidate behavioral suite exit 0 (survivor); reviewer probe on R1 exit 1
(expected failure); reviewer probe on baseline exit 0; git diff --check exit 0.
Tools: installed Go 1.25.5 on darwin/arm64; this is a Go CLI, not an iOS target.
The exact commands and probe source are included in the attached reviewer bundle.

Accepted, not rerun: configured runtime change-request_rev1-validation.log
records make check (build, formatting, vet, all tests and race) exit 0.
Producer native-pi-validation separately reports a manual make check exit 0:
there was therefore a manual full suite plus the runtime full suite, not a
single overall full-suite execution. Future producer runs must use the
configured runtime check without manually duplicating it. Reviewer ran no
full make check. Producer mutants-03 summary and each named failure log were
inspected: 12/12 existing mutants killed, all Go exits 1, harness reported 0.
Those twelve successful attacks do not cover the additional surviving R1.
No production gate scans source text; source-token attack rule is inapplicable.

Fresh go list -m -json resolves real v0.5.11 without Replace; GOWORK empty.
Fresh exact remote tag read exits 0 and matches tag object
0ea486e46765ecf12fff7c2ac526e12da02e95ed and peeled commit
a2a6e9f377f62a5872d99ecdfff0d1690e385f2a. Cryptographic signature verification
is accepted from the explicit primary publication evidence, not independently
rerun. Real pinative and google registration imports are present. No bare
BuildPlan call or permission bypass exists in the candidate production code.
A0 sections 3.2–3.5 and current SPEC 4.4 were read; historical v0.5.10 Pi
limitations are superseded by the actual v0.5.11 positive evidence.

Architecture is appropriate and existing tests are green, but F1 prevents
acceptance. Verdict: changes requested; route to to-dev. No external blocker,
no reopened Pi MCP product decision, no accept_cr, no done transition.
