# CR1 independent review — changes_requested

Task: TASK-260908-ranc5y. CR-TASK-260908-ranc5y-1 revision 1.
Candidate: 4f04bf708eb70afe019399c10586334f362099fe; base: 25379f22245e3bd8d2b81b192287d04368bf42c7.
Reviewer run: RUN-260908-a67608. repeat-of: none.
Repository delta is present. Candidate remains uncommitted and byte-identical to the frozen tree; no product source was modified. Mutants ran only in an archived scratch copy.

## F1 — schema character bound implemented as a byte bound (medium)

Location: internal/fragment/fragment.go:284–287, checkAbsolutePath.
Shape: refusal contradicts normative source; Unicode boundary untested.
The pinned launch-env-fragment-v1 schema absolutePath uses minLength=2/maxLength=4096 (JSON string characters). The implementation explicitly says bytes and uses len(s). A valid 4096-character path containing non-ASCII characters fails with resolve_fragment_invalid.

Reproduced through the built production executable -> run -> Resolver.Resolve -> ExecRunner -> real fake-curator subprocess -> Parse -> checkAbsolutePath. With a valid pi fragment and home '/' + 4095 'é' characters, Curator exits 0 but the launcher rejects. '/' + 4095 ASCII characters resolves; both 4097-character variants reject. No filesystem path is accessed or model launched. The issue applies to env, system_prompt.path, mcp.path, and path_prepend via the shared validator.

Required rework: implement the schema character bound (UTF-8 validity is already enforced), correct the comment, and add named production-boundary tests accepting the Unicode limit and refusing one character beyond it. Keep NUL/traversal/absolute-path guards intact. Include a narrowing numeric mutant and demonstrate the named boundary test fails.
repeat-of: none.

## F2 — corpus coverage attestation permits silently dropped negatives (medium)

Location: internal/fragment/fragment_test.go:42, TestConformanceCorpus.
Shape: coverage ratio not enforced; narrowed evidence set accepted.
The current 49 rows and fixture bytes DO match curator-spec 87a0d0060bad64ab883d007dcdf35df7485368bf, independently compared. However the test accepts any corpus of at least 40 rows. Removing only invalid-path-prepend-outside-root.json from the index leaves 48/49 rows and TestConformanceCorpus still passes (exit 0). Therefore the reported source-derived completeness can silently regress.

Required rework: pin and assert the exact independently derived indexed fixture set and expected verdicts (and fixture integrity as appropriate), not a minimum count or implementation-derived enumeration. Add a narrowing corpus mutant that removes one normative negative and requires the behavioral corpus test to fail. Do not invent additional channel restrictions.
repeat-of: none.

## Coverage and sound evidence retained

Producer reports 9/9 AC rows. More precisely, 8 of 9 rows have named driving tests; row 9 (SPEC-only delta and dependency/stage scope) is a stated inspection bound. Row 5's schema acceptance-boundary coverage is incomplete as F1 demonstrates; a named test is not proof of complete acceptance semantics.

| Row | Production call site | Candidate tests / evidence |
|---|---|---|
| 1 repair/argv order | run -> Resolve -> Request.Argv | TestArgvExactOrderAndRepair, TestExecRunnerEndToEnd, TestRunProductionResolverAgainstFakeCurator |
| 2 argv boundaries | Resolve -> ExecRunner | TestArgvExactOrderAndRepair, TestExecRunnerEndToEnd |
| 3 stderr | Resolve io.MultiWriter -> ExecRunner | TestResolveSuccessForwardsWarnings, TestExecRunnerEndToEnd |
| 4 errors | Resolve -> curatorDiagnostic | TestNonZeroExitMapping, TestExecRunnerNonZeroAndInvalidOutput, TestExecRunnerMissingBinaryAndCancellation |
| 5 closed schema | Resolve -> Parse -> ParseJSON/fromValue | TestConformanceCorpus, TestRejections, TestReaderRejections; F1/F2 remain |
| 6 A0 digest | Parse -> Canonical/Digest | TestA0FragmentsParseAndDigestMatch, TestDigestInvariantUnderPrintingAndSensitiveToBytes |
| 7 canonical bytes | Canonical/ParseJSON called by Parse | TestCanonicalRules, TestReaderRejections, TestReaderAccepts |
| 8 no retry/fallback | run -> Resolve | TestNonZeroExitNeverYieldsFragmentEvenWithValidStdout, TestInvalidOutputIsFragmentInvalidNeverAbsence |
| 9 scope/E6 | SPEC and imports | inspection bound; only E6 paragraph, tied to A0 accepted evidence and STORY-260908-2utz8k; no new module dependency or later stage |

Independently ran:
- go build -o .temp/review-ranc5y/curator-run ./cmd/curator-run — exit 0.
- go test -count=1 ./internal/fragment ./cmd/curator-run -run 'Test(ExecRunner|RunProduction|RunInformational|RunUsage|CanonicalRules|Reader|A0|Conformance|Optional)' — exit 0.
- python3 .temp/review-ranc5y/probe.py — exit 0, records F1, 49/49 fixture/source equality, all 3 independently computed A0 digests equal recorded values, opencode inside/outside path_prepend behavior correct for protocol fixture layout.
- python3 .temp/review-ranc5y/mutants.py — exit 0 as harness; duplicate-environment admission killed by TestRejections (test exit 1); diagnostic token-preserving mid-line match killed by TestNonZeroExitMapping (test exit 1); corpus-drop-one-negative survived (test exit 0).
- git diff 4f04bf708eb70afe019399c10586334f362099fe --exit-code — exit 0 after probes.

Accepted existing evidence, not rerun: CR1 exact-tree make check validation (build/fmt/vet/test/race exit 0), producer installed Curator probes, producer 26 fragment / 14 CLI mutant results. Independently reran the selected behavioral attacks above. No hosted CI, actual Curator invocation, ax/model launch, managed-home write, installation, branch mutation, or control-root/logbook edit.

Initial local corpus-comparison probe failed due to expecting an object rather than the normative index array; corrected the scratch probe and reran successfully. No failed read was treated as absence. task-board spawn goal reports this run is not goal-bound; directives empty.

Route to to-dev for these two localized fixes, then another independent reviewer cycle. This is ordinary rework, not a human-only blocker. Parent retains signed PR / exact-head landing ownership. This verdict also records the review findings for parent logbook propagation without touching the prohibited control root.
