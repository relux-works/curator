# F1/F2 rework — ready for review

Preserved CR1 tree 4f04bf708eb70afe019399c10586334f362099fe; initial diff against it was empty. Candidate remains uncommitted. Only F1/F2 changed: internal/fragment/fragment.go, fragment_test.go, cmd/curator-run/main_test.go, .scripts/fragment-mutants.sh, README.md. SPEC and all fixture bytes remain unchanged from CR1.

F1: checkAbsolutePath now counts validated UTF-8 Unicode characters with utf8.RuneCountInString; min 2/max 4096, absolute/NUL/traversal guards retained. TestExecutableUnicodePathBoundary builds the actual executable and drives main -> run -> Resolver.Resolve -> ExecRunner -> fake curator -> Parse -> checkAbsolutePath with 4096/4097-character paths. Both processes exit 1: accepted input reaches the existing not_implemented refusal, oversized input reaches resolve_fragment_invalid. No home is accessed and no model is launched.

F2: independently selected 49 rows by schema name from curator-spec commit 87a0d0060bad64ab883d007dcdf35df7485368bf conformance/v1/schema-cases/index.json using git show. Every fixture byte matches that commit. In upstream order, each manifest line is basename TAB lowercase valid boolean TAB lowercase SHA256 of verbatim fixture LF. Manifest SHA256: 540f9d32f14490c9293454f466c9e139cc112971a1147d038369702bd7984e60. TestConformanceCorpus pins this hash and 49 rows before driving Parse. This attests exact names, verdicts, ordering and content independently of implementation enumeration. Evidence archive includes upstream-derived manifest.

## Commands and actual exits

- go version / git --version readiness: 0.
- Independent Python upstream index and fixture comparison via git show: 0; 49/49 equal.
- go test ./internal/fragment ./cmd/curator-run -run 'TestConformanceCorpus|TestExecutableUnicodePathBoundary' -count=1 -v: 0 before mutants and 0 after restoration. Executable test invokes go build -o <temporary>/curator-run . successfully.
- FRAGMENT_MUTANT_IDS='M27 M28' FRAGMENT_TEST_PATTERN='TestExecutableUnicodePathBoundary|TestConformanceCorpus' bash .scripts/fragment-mutants.sh .temp/f1-f2/mutants: harness exit 0.
- git diff --check: 0.
- Initial handoff returned exit 0/to-review but no new validation artifact was found. Ran make check directly once: exit 0 (build, fmt-check, vet, test, race); attached standalone log.

| Mutant | Narrowed gate | Named failing test | Actual test exit / verdict |
|---|---|---|---|
| M27 | upper bound 4096 -> 4097 characters | TestExecutableUnicodePathBoundary/reject4097 | 1 / KILLED |
| M28 | drops only invalid-path-prepend-outside-root.json index row | TestConformanceCorpus | 1 / KILLED |

No survivors among these 2 mutants. Existing 26 fragment and 14 CLI mutants retained as CR1 evidence, not rerun. The harness permits optional explicit test/ID selection and otherwise preserves its full-suite default.

## Coverage and bounds

Focused rework: 2 of 2 rows driven: R1 production main through checkAbsolutePath, R2 production Parse after corpus attestation. Overall retained review mapping is 8 of 9 AC rows driven, with row 9 a scope/SPEC inspection bound (not a behavioral test):

| AC row | Production call site | Named test |
|---|---|---|
| 1 repair/order | run -> Resolve -> Request.Argv | TestArgvExactOrderAndRepair |
| 2 argv operands | Resolve -> ExecRunner | TestExecRunnerEndToEnd |
| 3 stderr | Resolve io.MultiWriter -> ExecRunner | TestResolveSuccessForwardsWarnings |
| 4 errors | Resolve -> curatorDiagnostic | TestNonZeroExitMapping |
| 5 schema | Resolve -> Parse -> fromValue/checkAbsolutePath | TestConformanceCorpus, TestExecutableUnicodePathBoundary |
| 6 A0 digest | Parse -> Canonical/Digest | TestA0FragmentsParseAndDigestMatch |
| 7 canonical bytes | Parse -> ParseJSON/Canonical | TestCanonicalRules |
| 8 no retry/fallback | run -> Resolve | TestNonZeroExitNeverYieldsFragmentEvenWithValidStdout |
| 9 scope/E6 | SPEC/import inspection | stated inspection bound; unchanged from CR1 |

Prior exact-tree full-suite and installed A0 evidence is retained, not independently rerun here. No hosted CI. No new launcher stages/imports, managed-home writes, model launches, installs, daemon changes, commits, tags, releases, or control-root/LOGBOOK edits. Reviewer acceptance is still pending; parent owns signed PR and landing. This artifact records F1/F2 findings for parent logbook propagation.

An initial rg used the wrong schema directory and failed with exit 2; recovered with rg --files and inspected schemas/v1/launch-env-fragment-v1.schema.json ($defs.absolutePath minLength=2/maxLength=4096). No failed read was treated as absence.
