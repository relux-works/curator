# CR2 independent review — accepted

Task TASK-260908-ranc5y; CR-TASK-260908-ranc5y-2 revision 2. Reviewer RUN-260908-39de12.
Candidate tree 2b91d2d63d09c2f18e58b32211dded18f1a64107, base 25379f22245e3bd8d2b81b192287d04368bf42c7. Repository delta present. Exactly five expected files changed from CR1. No new findings; prior CR1 F1/F2 closed.

## Independent evidence

- `python3 .temp/review-rev2/derive.py`: exit 0. Independently read upstream index and verbatim fixtures using git show from curator-spec 87a0d0060bad64ab883d007dcdf35df7485368bf. 49/49 names, verdicts and bytes match. Manifest SHA256 540f9d32f14490c9293454f466c9e139cc112971a1147d038369702bd7984e60 matches the new test pin. Upstream absolutePath minLength 2/maxLength 4096 confirmed.
- `go test ./internal/fragment ./cmd/curator-run -run 'TestConformanceCorpus|TestExecutableUnicodePathBoundary|TestRejections' -count=1 -v`: exit 0. Unicode 4096 accepted into the existing not_implemented stage, 4097 refused with resolve_fragment_invalid. Real built main -> run -> Resolver.Resolve -> ExecRunner -> fake Curator -> Parse -> checkAbsolutePath. No path creation or model launch.
- Archived the frozen candidate into scratch, then ran `FRAGMENT_MUTANT_IDS='M27 M28' FRAGMENT_TEST_PATTERN='TestExecutableUnicodePathBoundary|TestConformanceCorpus' bash .scripts/fragment-mutants.sh .temp/mutants`: harness exit 0. M27 keeps the path gate and weakens max 4096 to 4097: test exit 1, TestExecutableUnicodePathBoundary/reject4097 fails. M28 removes only invalid-path-prepend-outside-root from the index: test exit 1, TestConformanceCorpus fails with 48 rows and mismatched manifest. 2/2 narrowing mutants killed.
- All six scratch mutation targets restored byte-for-byte. Managed source remained unchanged: `git diff 2b91d2d63d09c2f18e58b32211dded18f1a64107 --exit-code` and `git diff --check`: exit 0.

F1 counts already-validated UTF-8 scalars and preserves absolute, NUL and traversal checks. F2 attests an independently derived corpus before invoking production Parse; it no longer permits a narrowed evidence set. Optional harness selectors leave all 28 mutants and the full behavioral test pattern active by default. No static-only harness was introduced.

## Coverage and retained evidence

Focused rework: 2 of 2 rows driven (F1 executable boundary; F2 production Parse corpus and integrity gate). Overall 8 of 9 AC rows driven, row 9 explicitly an inspection bound:

| AC row | Production call site | Named driving test |
|---|---|---|
| 1 repair/order | run -> Resolve -> Request.Argv | TestArgvExactOrderAndRepair |
| 2 operand boundaries | Resolve -> ExecRunner | TestExecRunnerEndToEnd |
| 3 stderr | Resolve io.MultiWriter -> ExecRunner | TestResolveSuccessForwardsWarnings |
| 4 errors | Resolve -> curatorDiagnostic | TestNonZeroExitMapping |
| 5 closed schema | Resolve -> Parse -> checkAbsolutePath | TestConformanceCorpus, TestRejections, TestExecutableUnicodePathBoundary |
| 6 A0 digest | Parse -> Canonical/Digest | TestA0FragmentsParseAndDigestMatch |
| 7 canonical bytes | ParseJSON/Canonical used by Parse | TestCanonicalRules |
| 8 no retry/fallback | run -> Resolve | TestNonZeroExitNeverYieldsFragmentEvenWithValidStdout |
| 9 scope/E6 | SPEC/import diff | inspection bound |

Retained, not rerun: sound CR1 review evidence including independent A0 digest computation, subprocess/canonical/reader probes and selected attacks; producer original mutant evidence; CR2 exact-tree runtime make check (build/fmt/vet/test/race, attached validation log exit 0). Inspected the unchanged SPEC E6-only delta and no added dependency/stage. SPEC and all corpus files are unchanged from CR1. The scope remains fragment resolution.

Read actual task checklist: all 23 entries checked; reviewed applicable items against focused and retained evidence. Spawn goal reports not goal-bound; no directives. This acceptance is not landing; producer-side integration and parent signed PR lifecycle remain outstanding.

Operational notes: initial board projection used unsupported resources field and was corrected to outcomeResources; an archive command initially targeted a not-yet-created scratch cwd and never executed, then directory creation and execution succeeded; checklist CLI discovery returned unknown command, so used the actual task checklist projection. No failed read was treated as absence. No product edits, commits, CI, installs, daemon changes, real Curator/model calls, home writes, or private/control-root/logbook edits. This outcome also records F1/F2 closure for parent logbook propagation.
