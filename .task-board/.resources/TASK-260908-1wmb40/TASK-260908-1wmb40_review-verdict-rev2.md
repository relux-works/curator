# TASK-260908-1wmb40 CR2 independent review

Verdict: ACCEPT revision 2. No blocking findings. This is a new exact-tree acceptance, not a transfer of CR1 acceptance.

Base: 18aeaed9af7dc5ffbe6cc79a4731a852fbb716da.
Candidate: ff61be4a8bd43fa4ffb179d31aa38e41891d4313.
Current signed default: 3ff66a9421ff6ddf675a49fc0c2868309f6e3de3.
Reviewed all 21 changed paths; all 97 candidate tracked files match the immutable tree. No reviewer source edits, commits or branch/index movements. All tested regression files are already committed in the signed current default, while this managed candidate remains uncommitted at HEAD 18aeaed.

## Recovery and upstream provenance

Fresh git ls-remote --symref origin HEAD reports refs/heads/main at 3ff66a9; exact-ref fetch returned the same SHA. git verify-commit reports a good configured human signature. Final repeat observation remained equal. This verifies the named authority ref and signed content; no hosted branch-protection setting change is asserted.

Public activity contains the original CR1 acceptance by RUN-260908-1233da and separate CR2 creation by RUN-260909-e08d63. Read the exact public recovery export, recovery RUN49fa72 outcome, publication evidence and CR2 validation resource. Old acceptance was not modified by this review.

Export SHA256: 08e93a55606f3dcaf6d11e731458d298a474cc063c740abada291f1d2709d72d.
Decoded review patch SHA256: 071f8bd5bc8a9f93d2a811fc4215fc9d3a46f46fda337ad7d387eec1af487b53.
Decoded update patch SHA256: e7d9b0af01ea7937c74356fdd7e268a2cae2edd8d19208f752fb866dac5add01.
CR2 resource patch SHA256: 59871ea0d1c64f404b6c6d0a006713653ceec16d584c638f82c6b006a09f4cd5.
All digests independently recomputed. Each of the three patches was applied only to a separate temporary index from its declared base; all produced exactly ff61be4. All 12 git operations exited 0. No patch was applied to working source or the real index.

Nonblocking evidence correction: RUN49fa72 lists decoded patch sizes 137338/110511, which are character counts. Actual UTF-8 byte counts are 137533/110682. Both recorded hashes match actual bytes; no content discrepancy.

Composition source/tests/probe/harness and go.mod/go.sum are byte-identical to CR1. README carries subsequent API documentation. SPEC is byte-identical to upstream d97e6cb (Pi E5 errata). Systemprompt source/tests/harness are byte-identical to adf6276. Execution/config/tests/helpers/harness are the landed 3ff66a9 implementation. Reviewed these APIs, test entry points, process helpers, all three narrowing harnesses and README/SPEC delta. Upstream Pi discovery correction is consistent with conditional file warnings and same-semantics suppression, and does not alter composition §4.5. Reserved path_prepend remains only the existing parser/hash contract; no managed command roots are invented.

## AC coverage

**10 of 12 expanded AC rows driven; 10 of 10 composition API-scope rows.** The original row denominator is retained; broader executable integration is not silently counted as delivered.

| Row | Production API and named driving tests |
| --- | --- |
| 1 Exact plan/prompt/MCP/native argv | Compose; TestComposeOrderAndChannels |
| 2 Binary/WorkDir and independent inputs | Compose; TestComposeOrderAndChannels |
| 3 Full filtered Env, removed names stay removed | Compose; TestComposeEnvironmentBoundary, TestComposeReleasedChildEnvironment |
| 4 Nil-parent same-request ownership, inherited secrets omitted | Compose -> ChildEnv; TestComposeEnvironmentBoundary, TestComposeReleasedChildEnvironment |
| 5 Disjoint names, inherited lookups retained, names-only warnings | Compose; TestComposeEnvironmentBoundary, TestComposeUnownedOverrideNoWarning |
| 6 Unattached/empty/UTF8/binary stdin | Compose; TestComposeStdin |
| 7 MCP path/companions/name/variable/absence | Compose; TestComposeOrderAndChannels, TestLaunchBoundaryUnengaged |
| 8 Fresh filesystem refusal without dropping argv | Value.CheckLaunchBoundary; TestLaunchBoundaryFilesystem, TestLaunchBoundaryLateReplacement |
| 9 Ownership failure propagation | Compose; TestComposeOwnershipFailure |
| 10 Empty direct environment cannot inherit | Compose; TestComposeEmptyEnvironment |
| 11 Actual main/direct exec/ax, stderr and full tracked document | BOUND: later execution Story |
| 12 Native Pi admission using real new operator tag | BOUND: later plan/execution Story |

Row 11 refinement for the new tree: execution.Prepare -> Launch.Run -> run -> exec.Cmd.Start now drives real synthetic subprocesses and the MCP late probe in both modes. TestRealProcessMatrix verifies argv, full direct Env, tracked literals, closed document, schema/extensions, stdin and WorkDir. TestLateChecksBothModes verifies no subprocess after nil/refusing callback, missing/nonexecutable provider and missing/dangling/directory/unreadable MCP. TestLookupCollisionWarningsBothModes verifies names-only stderr. TestAxFailureNoFallback verifies exact error bytes and no direct fallback. Main still does not bind these APIs or the real systemprompt.PrepareLaunch callback; full pipeline remains TASK-260908-1o7i8y. Row 12 remains excluded: module v0.5.10 does not attest native Pi admission.

Additional upstream behavior independently checked through Select/PrepareLaunch (TestSelectExactArgv, TestSelectRefusalsAndNoOptIn, TestPrepareLaunchLateFilesAndWarnings, TestPrepareLaunchFileFailures, TestPrepareLaunchSelectedPathChangesLate) and axconfig.Load (TestLoadClosedSchema, TestLoadPrecedenceBeforeParse, TestLoadFilesystemFailures). These do not substitute for a main call site.

## Validation and negative evidence

Independently ran:
`go test ./internal/composition ./internal/systemprompt ./internal/axconfig ./internal/execution -count=1 -v -run 'TestCompose|TestLaunchBoundary|TestSelect|TestPrepareLaunch|TestLoad|TestLateChecksBothModes|TestRealProcessMatrix|TestLookupCollisionWarningsBothModes|TestAxFailureNoFallback'`
Exit 0, focused-01.log. No skipped permission tests. This targets composition plus new-tree compatibility and late refusal behavior on Darwin arm64, Go 1.25.5.

Accepted actual CR2 runtime make check evidence ending [exit 0]: build, fmt, vet, full tests and race. Did not duplicate broad validation or rerun old mutants. Reviewed previous composition reviewer archive: all nine actual expected-red logs have failing behavioral tests and exit 1, matching the 9/9 summary. M1/M2/M3 attack inherited SECRET, OWN overlap and REMOVED re-admission at Compose; M4 attacks selective ChildEnv failure; M5-M8 attack missing/dangling/directory/permission classes at CheckLaunchBoundary; M9 attacks attached-empty stdin. All those implementation/test bytes remain identical.

Read authenticated TASK-260908-3ued5d_review-verdict-rev2.md for the exact same ff61be4 tree. Reuse its independent real PTY/cleanup verification and F1/F2/E6 narrowing failures, plus its explicitly attributed unchanged C1-C9/E1-E5/E7 evidence. Inspected the shipped behavioral narrowing harnesses; no new source-text gate exists, so token-preserving static-gate attack is not applicable. Reuse is evidence for unchanged behavior, not acceptance transfer.

Actual go list -m reports skill-agents-management v0.5.10 with no Replace; GOWORK is empty and GOMOD is this worktree. No pseudo-version or workspace replacement. TestComposeReleasedChildEnvironment exercises the actual released Codex ownership interface. git diff --check exits 0.

## Explicit limits and disposition

Same system/request as the admitted plan and validated fragment are typed API preconditions. Compose does not build plans. Final main must load axconfig before usage validation, admit exactly one plan, select/apply prompt policy, compose once, prepare with composition time, bind fresh systemprompt.PrepareLaunch with its warnings, Run and propagate status. Native provider profile conflict remains native uninspected argv behavior. Tracked destination filtering, post-probe TOCTOU, native discovery, Linux runtime and full shell job-control bounds remain as documented; no actual ax/provider launch or native Pi admission was tested.

All merged checklist items assessed within these explicit bounds; no reviewer commit or done transition. Outcomes replace forbidden LOGBOOK writes. No installs, hosted CI, tags, releases, real homes, private records or control-root source writes. Acceptance must use public accept_cr revision=2, routing to integrating. Parent/producer owns public completion of already-landed content; no new code PR is required while exact-tree equality holds.

Operational non-gate failures: an initial query used unsupported resources projection (exit 1; repaired to outcomeResources), an exploratory task-board cr --help was unsupported (exit 1; used public activity/resource commands), and optional skill-directory search returned 2 for absent directories. These were discovery failures, not successful validation claims.
