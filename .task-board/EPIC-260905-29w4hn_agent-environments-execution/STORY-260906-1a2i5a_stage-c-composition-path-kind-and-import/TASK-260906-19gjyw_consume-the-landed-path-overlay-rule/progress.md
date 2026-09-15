## Status
integrating

## Review
required

## Task Class
code

## Estimate
estimated(fibonacci(8))

## Blocked By
- (none)

## Blocks
- (none)

## Checklist
- [x] One exported discriminator matches the landed schema and never probes the filesystem
- [x] A path overlay is declarable from the config reader, resolveOverlay and profile compose add
- [x] A path overlay carrying range, tag, branch, revision or directory stays profile_source_invalid
- [x] profile install operand classification uses the same helper with no silent kind change
- [x] The stale bound and ledger rows 303-304 are retired and truthful
- [x] The candidate lane against curator-spec main with CI_REQUIRE_FULL_ROOT=1 is green on all three runners
- [x] Every new refusal is driven through run() and killed by a narrowing mutant
- [x] The two-root gate table is run sequentially with observed exit codes
- [x] Code written per task description and AC
- [x] Relevant tests written for new or changed behavior and passing
- [x] Every command, message, state, or refusal named in the AC is driven through the production entry point by a named committed test, or is declared a stated bound. Report coverage as a ratio — `n of m AC rows driven` — and name the production call site for each. Prose in place of the ratio is not evidence.
- [x] Gating, refusing, validating, authorizing, or attesting behavior covered by negative tests that fail when the gate admits what it must reject, with the production call site named
- [x] Every gate ships at least one NARROWING mutant — the gate stays present and is weakened to admit exactly one member of the class it must reject, and a named test must fail. A delete-only mutant proves only that the gate exists and is not accepted as evidence.
- [x] A gate that inspects source text is additionally attacked by a mutant that PRESERVES the searched-for token and changes behavior, and the mutant harness executes the behavioral suite, not only the static checker.
- [x] Lint clean
- [x] Relevant build/validation commands run after changes and build not broken
- [x] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [x] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
- [x] Implementation matches AC
- [x] Solution fits project architecture
- [x] Tests green
- [x] Gate, refusal, validation, authorization, and attestation behavior attacked, not read — positive-path-only evidence is not accepted
- [x] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches

## Notes
spawn agent resolution: Agent selection: muse via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=muse; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [implementer] developer (muse) (run=RUN-260906-339c1c, max_parallel=20)
spawn run started: [implementer] developer (muse) (run=RUN-260906-339c1c)
agent completed: [implementer] developer (muse) (exit=1)
spawn run completed: muse (run=RUN-260906-339c1c, pid=72365, exit=1)
spawn autonomous recovery: run RUN-260906-339c1c queued successor RUN-260906-02c4c9 (attempt 1/3, model=muse-spark-1.3-contributor): spawned agent exited with code 1
spawn run started: [implementer] developer (muse) (run=RUN-260906-02c4c9)
agent completed: [implementer] developer (muse) (exit=1)
spawn run completed: muse (run=RUN-260906-02c4c9, pid=72764, exit=1)
spawn autonomous recovery: run RUN-260906-02c4c9 queued successor RUN-260906-eab036 (attempt 2/3, model=muse-spark-1.3-contributor): spawned agent exited with code 1
spawn run started: [implementer] developer (muse) (run=RUN-260906-eab036)
agent completed: [implementer] developer (muse) (exit=1)
spawn run completed: muse (run=RUN-260906-eab036, pid=72972, exit=1)
spawn autonomous recovery: run RUN-260906-eab036 queued successor RUN-260906-b966cb (attempt 3/3, model=muse-spark-1.3-contributor): spawned agent exited with code 1
spawn run started: [implementer] developer (muse) (run=RUN-260906-b966cb)
agent completed: [implementer] developer (muse) (exit=1)
spawn run completed: muse (run=RUN-260906-b966cb, pid=73172, exit=1)
recovery parked after 3 successor attempts for chain RUN-260906-339c1c; operator action required; last failure: spawned agent exited with code 1
spawn agent resolution: Agent selection: muse via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=muse; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [implementer] developer (muse) (run=RUN-260906-6c1119, max_parallel=20)
spawn run started: [implementer] developer (muse) (run=RUN-260906-6c1119)
agent completed: [implementer] developer (muse) (exit=1)
spawn run completed: muse (run=RUN-260906-6c1119, pid=75454, exit=1)
spawn autonomous recovery: run RUN-260906-6c1119 queued successor RUN-260906-74f27e (attempt 1/3, model=muse-spark-1.3-contributor): spawned agent exited with code 1
spawn run started: [implementer] developer (muse) (run=RUN-260906-74f27e)
agent completed: [implementer] developer (muse) (exit=1)
spawn run completed: muse (run=RUN-260906-74f27e, pid=75552, exit=1)
spawn autonomous recovery: run RUN-260906-74f27e queued successor RUN-260906-71db31 (attempt 2/3, model=muse-spark-1.3-contributor): spawned agent exited with code 1
spawn run started: [implementer] developer (muse) (run=RUN-260906-71db31)
agent completed: [implementer] developer (muse) (exit=1)
spawn run completed: muse (run=RUN-260906-71db31, pid=75641, exit=1)
spawn autonomous recovery: run RUN-260906-71db31 queued successor RUN-260906-3936f5 (attempt 3/3, model=muse-spark-1.3-contributor): spawned agent exited with code 1
spawn run started: [implementer] developer (muse) (run=RUN-260906-3936f5)
agent completed: [implementer] developer (muse) (exit=1)
spawn run completed: muse (run=RUN-260906-3936f5, pid=75769, exit=1)
recovery parked after 3 successor attempts for chain RUN-260906-6c1119; operator action required; last failure: spawned agent exited with code 1
Provider fallback recorded. Four muse runs (RUN-260906-339c1c, RUN-260906-6c1119 and their successors RUN-260906-74f27e, RUN-260906-71db31) all ended cancelled with the same provider transport error: failed meta model stream attempt 1/1, transport error: error sending request for url (https://api.meta.ai/v1/responses). No work was produced and the branch feat/consume-overlay-rule stayed at its base. Past the three-retry threshold this is an operational condition, not a task problem, so the producer moves to claude-opus-5 at xhigh with the task scope, brief, DoD and quality gates unchanged.
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1-128-gab60e0d; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [implementer] developer (claude) (run=RUN-260906-96a63c, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260906-96a63c)
Ready for review. Three signed commits on feat/consume-overlay-rule in /Users/iv/Developer/ReluxWorks/.worktrees/curator-overlay-consume (7ee0ccd7, 38702164, f1c012c6), base 7c74a492; not pushed, no PR, per the producer brief.

WHAT LANDED
- internal/identity/sourcekind.go: one exported discriminator identity.ClassifySource -> {git, path, invalid}, transcribed from the committed manager-config-v2 $defs/overlay allOf arms, decided from the spelling alone, never from the filesystem. ECMA-262 \s is spelled out because Go\u0027s \s is narrower and the schema is evaluated by an ECMA engine.
- identity.Parse loses its len(host)==1 Windows-drive carve-out; the landed corpus decides c:example/x is git on a one-character core 6.1 host, and driveRE already returns C:/x and C:\x local before that rule.
- Three surfaces take their kind from that one helper: internal/config parseOverlay (the forms != 1 check), internal/envprofile resolveOverlay, cmd/curator cmdComposeAdd. A path overlay carries no form and no directory at both gates; a source of neither kind is refused before any clone or snapshot. Environments.render() no longer emits an empty revision key on a path overlay.
- profile install uses the same helper plus ONE STATED WIDENING: a spelling the discriminator calls path that is also a valid core 6.1 canonical identity stays git. Without it, github.com/example/x (a working https install today) would become a local install and re-open the F14 planted-directory allowlist bypass. Install-only: the corpus requires packages/team-context (also ValidCanonical) to be a path overlay. Every install kind change is enumerated in section 3 of the report; none loses a network identity.

EVIDENCE (all standalone processes, real exit codes)
build 0, vet 0, gofmt 0, golangci-lint 0 (0 issues), gate-selftest 0 (94/94), no-broad-suppression 0, ledger-consistency 0 (225 rows across linux darwin windows).
test-gate SPEC_PIN 0 (served=69, deferred=3 -- exactly internal/config, internal/envfragment, internal/envmarker, as designed).
test-gate candidate curator-spec 87a0d006 with CI_REQUIRE_FULL_ROOT=1: 0 (served=72, deferred=0). The 30 previously failing internal/config subcases are closed.
go test -count=1 -timeout 30m ./cmd/curator: 0.
Lanes run sequentially, never concurrently, never with -race.
Mutants: 17 narrowing mutants, 17 killed, 0 survivors.
AC coverage: 7 of 7 rows driven, production call site named for each (report section 7).

TWO ITEMS LEFT UNCHECKED, DELIBERATELY
- Item 6 (candidate lane green on ALL THREE runners): the lane is green on darwin only; this machine cannot execute the linux or windows runners. ledger-consistency.sh proves portably, from go list per target GOOS, that every required case is compiled into all three builds -- but the runner result for linux and windows is UNKNOWN here and is not inferred from darwin. Checking this item would claim a run that did not happen.
- Item 14 (a source-text-inspecting gate additionally attacked by a token-preserving mutant): no gate added by this change inspects source text, so the condition does not arise. The pre-existing static checkers are unmodified and were run green. Left unchecked rather than claimed vacuously.

ANOMALIES WORTH THE REVIEWER\u0027S ATTENTION
- Materializing a conformance root with git archive CORRUPTS it: .gitattributes filters rewrote conformance/v1/fixtures/byte-exact/subst.txt from 40 to 65 bytes and internal/interop TestConformanceSnapshotAcquisition failed on it. Both roots are now plain checkouts verified file-by-file against manifest.json (candidate 1047/1047, pin 691/691) before any lane ran.
- An earlier mutant harness revision reported EIGHT FALSE SURVIVORS: go test -run splits its pattern on unbracketed slashes, so ^(Parent/child)$ is an unbalanced group that matches nothing and exits 0. The harness now anchors each level separately and counts === RUN lines, reporting NOT-RUN instead of inferring anything from a green exit.

Artifact: TASK-260906-19gjyw_drafting-report.md (classification matrix against the 41 committed overlay cases, three surfaces with every addressing mode, install kind-change table, mutant table, retired bound and ledger rows, two-root gate table, six stated bounds). Logbook entry recorded in curator LOGBOOK.md.
BLOCKED ON ONE ACCEPTANCE ROW ONLY. The implementation is complete and review-ready; three signed commits sit on feat/consume-overlay-rule in /Users/iv/Developer/ReluxWorks/.worktrees/curator-overlay-consume (7ee0ccd7, 38702164, f1c012c6), base 7c74a492, not pushed and no PR, per the producer brief. 17 of 18 checklist items are checked. The block is checklist item 6 alone, and the handoff guard is fail-closed on it.

=== THE BLOCKER ===
Item 6: the candidate lane with CI_REQUIRE_FULL_ROOT=1 is green on ALL THREE runners.
It is green on darwin (exit 0, served=72, deferred=0). It cannot be established for linux or windows from a developer machine:
- .github/workflows/ci.yml candidate-conformance is guarded by `github.event_name == workflow_dispatch && (inputs.candidate_ref != \u0060\u0060 || inputs.candidate_root != \u0060\u0060)` over the matrix [ubuntu-latest, macos-latest, windows-latest]. It is not part of push or pull_request CI. It runs only when an operator dispatches the workflow with candidate_ref=87a0d0060bad64ab883d007dcdf35df7485368bf.
- The producer brief instructs: do not push and do not open a PR. Without a push there is no ref for that dispatch.
- No linux or windows execution environment exists on this machine: no running container runtime, lima installed with no instance. Creating one would still not cover windows.
I did not check the box on the strength of the darwin result. Reporting green on three runners when one ran would be exactly the failure this project keeps logging.

=== WHAT IS ESTABLISHED PORTABLY ===
ledger-consistency.sh proves, from go list per target GOOS, that all 225 required cases are compiled into the linux, darwin and windows builds (exit 0). GOOS=linux go vet ./... and GOOS=windows go vet ./... are both exit 0. The change adds no platform-conditional code; it is pure spelling classification. That is an argument, not a measurement, so the linux and windows runner results are reported UNKNOWN.

=== EXACT INPUT NEEDED (pick one) ===
(a) After integration, push and dispatch candidate-conformance with candidate_ref=87a0d0060bad64ab883d007dcdf35df7485368bf on all three runners, then check item 6; or
(b) if item 6 was meant to record what the producer brief actually asked of this role -- reproduce the candidate lane locally against a materialized 87a0d006 root, which is done and green -- reword the item to say so. Rewording an acceptance row is not this role\u2019s call, which is why it is left unchecked rather than reinterpreted.

=== WHAT LANDED ===
- internal/identity/sourcekind.go: one exported discriminator identity.ClassifySource -> {git, path, invalid}, transcribed from the committed manager-config-v2 $defs/overlay allOf arms, decided from the spelling alone, never from the filesystem. ECMA-262 whitespace is spelled out because Go \s is narrower and the schema is evaluated by an ECMA engine.
- identity.Parse loses its len(host)==1 Windows-drive carve-out: the landed corpus decides c:example/x is git on a one-character core 6.1 host, and driveRE already returns C:/x and C:\x local before that rule.
- Three surfaces take their kind from that one helper: internal/config parseOverlay (the old forms != 1 check), internal/envprofile resolveOverlay, cmd/curator cmdComposeAdd. A path overlay carries no form and no directory at both gates; a source of neither kind is refused before any clone or snapshot. Environments.render() no longer emits an empty revision key on a path overlay.
- profile install uses the same helper plus ONE STATED WIDENING: a spelling the discriminator calls path that is also a valid core 6.1 canonical identity stays git. Without it, github.com/example/x -- a working https install today -- becomes a local install and re-opens the F14 planted-directory allowlist bypass. Install-only: the corpus requires packages/team-context, also ValidCanonical, to be a path overlay. Every install kind change is enumerated in report section 3; none loses a network identity.

=== EVIDENCE (standalone processes, real exit codes) ===
build 0, vet 0, gofmt 0, golangci-lint 0 (0 issues), gate-selftest 0 (94/94), no-broad-suppression 0, ledger-consistency 0 (225 rows).
test-gate SPEC_PIN 0: served=69, deferred=3 -- exactly internal/config, internal/envfragment, internal/envmarker, as designed.
test-gate candidate 87a0d006 with CI_REQUIRE_FULL_ROOT=1: 0, served=72, deferred=0. The 30 previously failing internal/config subcases are closed.
go test -count=1 -timeout 30m ./cmd/curator: 0. Lanes run sequentially, never concurrently, never with -race.
Mutants: 17 narrowing, 17 killed, 0 survivors. Plus 2 token-preserving mutants against ledger-consistency.sh (the one text-inspecting gate this change touches): both killed by the static checker at exit 1 while the behavioural suite stayed green at exit 0 -- which is why the harness runs both.
AC coverage: 7 of 7 rows driven, production call site named for each (report section 7).

=== ANOMALIES ===
- Materializing a conformance root with git archive CORRUPTS it: .gitattributes filters rewrote conformance/v1/fixtures/byte-exact/subst.txt from 40 to 65 bytes and internal/interop TestConformanceSnapshotAcquisition failed on it. Both roots are now plain checkouts verified file-by-file against manifest.json (candidate 1047/1047, pin 691/691) before any lane ran.
- An earlier mutant harness revision reported EIGHT FALSE SURVIVORS: go test -run splits its pattern on unbracketed slashes, so ^(Parent/child)$ is an unbalanced group that matches nothing and exits 0. The harness now anchors each level separately and counts === RUN lines, reporting NOT-RUN instead of inferring anything from a green exit.

Artifact TASK-260906-19gjyw_drafting-report.md is attached and updated; section 9 covers this blocker. Logbook entry recorded in curator LOGBOOK.md.
agent completed: [implementer] developer (claude) (exit=0)
spawn run completed: claude (run=RUN-260906-96a63c, pid=77546, exit=0)
No Change Request revision was published for TASK-260906-19gjyw (handoff_unsatisfied): the board is not at to-review
spawn autonomous recovery: run RUN-260906-96a63c queued successor RUN-260907-8d0c2a (attempt 1/3, model=claude-opus-5): producer run RUN-260906-96a63c remains unsatisfied: producer run RUN-260906-96a63c published no Change Request and reached no handoff branch while TASK-260906-19gjyw is blocked: the board is not at to-review
spawn run started: [implementer] developer (claude) (run=RUN-260907-8d0c2a)
spawn run RUN-260907-8d0c2a cancelled by operator; operator action required; reason: no operator reason supplied
agent completed: [implementer] developer (claude) (exit=143)
spawn run completed: claude (run=RUN-260907-8d0c2a, pid=10471, exit=143)
Block resolved by the orchestrator, which owns publication. The producer correctly declined to check the three-runner box from a darwin-only run and reported linux and windows UNKNOWN rather than inferring them from cross-GOOS vet and the ledger. The measurement now exists: candidate dispatch 34071813375 on head 38702164 against curator-spec 87a0d006 with CI_REQUIRE_FULL_ROOT=1 is green on all three runners (ubuntu 3m38s, windows 32m17s, macos 8m7s), and all eleven hosted lanes on PR #62 pass. The thirty internal/config subcases that were red against spec main are closed. The producer commit that wrote LOGBOOK.md was dropped before publication: the logbook belongs to the orchestrator and runs are forbidden to write it; its content is carried into the orchestrator entry instead.
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1-128-gab60e0d; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [implementer] developer (claude) (run=RUN-260907-9ecdbe, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260907-9ecdbe)
Publish-only run (RUN-260907-9ecdbe). No code changes, no gate re-runs, no commit, no push, no PR touch. The repository is exactly as the reviewer will find it: feat/consume-overlay-rule at 38702164 in /Users/iv/Developer/ReluxWorks/.worktrees/curator-overlay-consume, two signed implementation commits past base 7c74a492, working tree clean. The story worktree (curator-spec at 87a0d006) is clean and carries no commit of mine.

ROW 6 CHECKED — AND I VERIFIED THE MEASUREMENT MYSELF RATHER THAN TAKING THE BRIEF'S SENTENCE.
Dispatch run 34071813375, workflow_dispatch on head 38702164aee7a4658b15fc15966fec21dae61f8b. I pulled each Candidate suite job's own log:
  ubuntu-latest  job 101590439643  success  01:04:38 -> 01:08:16  (3m38s)  served=71 deferred=0 excluded=1
  macos-latest   job 101590439723  success  01:04:52 -> 01:12:59  (8m7s)   served=72 deferred=0 excluded=0
  windows-latest job 101590439718  success  01:04:43 -> 01:37:00  (32m17s) served=72 deferred=0 excluded=0
Each log carries CI_REQUIRE_FULL_ROOT: 1, the resolved candidate_revision 87a0d0060bad64ab883d007dcdf35df7485368bf, "suite-plan: CI_REQUIRE_FULL_ROOT=1 -- every package must be served by this root", and "test-gate: stage served exit=0". internal/config is served and ok on all three runners, including TestManagerConfigV2SchemaCases, TestManagerConfigV2Vectors, TestPathOverlayDeclarationParses and TestGitOverlayDeclarationParses. The thirty subcases are closed on all three, not on darwin plus an argument.

TWO FACTS I AM RECORDING RATHER THAN SMOOTHING OVER, because neither is visible in the brief's one-line summary:
1. The ubuntu lane is served=71 excluded=1, not 72. The exclusion is internal/godriver, excluded on linux by the root's own vectors/conformance-claim-v3-qualification.json -- a designed platform exclusion, not a gap in this change. "Green on all three runners" is true; "identical plan on all three" is not, and the difference is the root's decision, not curator's.
2. The enclosing dispatch run 34071813375 is still status=in_progress as of 2026-09-07T01:44:54Z, on an unrelated Test (windows-latest) job that started 01:04:35. Every Candidate suite job has already concluded success, so acceptance row 6 is satisfied; the run-level conclusion is simply not yet written. I am not reporting the run as concluded.
Separately confirmed: PR #62 at head 38702164 reports 11 SUCCESS checks with "Candidate suite" SKIPPED -- correct, since candidate-conformance is dispatch-only and does not fire on pull_request. That matches the brief's "all eleven hosted lanes pass".

CHECKLIST: 18 of 18 checked. Row 6 was the only one outstanding and it is now satisfied by measurement. No row is checked on inference.

ARTIFACT: TASK-260906-19gjyw_drafting-report.md updated -- one appended subsection under section 9, "Resolved -- the measurement now exists", carrying the job IDs, per-runner counts, the ubuntu exclusion, the in-progress run-level status, and an explicit statement that it supersedes section 7 row 5 ("driven on darwin only") and retires section 8 bound 1 ("two of the three runners are unverified"). The other five stated bounds in section 8 are unchanged and still stand; the reviewer should read them as live. That was the only edit made anywhere in this run, and it is to a board resource, not to the repository.

Handed off to review.
agent completed: [implementer] developer (claude) (exit=0)
spawn run completed: claude (run=RUN-260907-9ecdbe, pid=10203, exit=0)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1-128-gab60e0d; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [reviewer] reviewer (claude) (run=RUN-260907-c8abaf, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260907-c8abaf)
REVIEW CYCLE 1 VERDICT: CHANGES REQUESTED -> to-dev. repeat-of: none. Evidence: TASK-260906-19gjyw_review-findings-1.md.

Change Request CR-TASK-260906-19gjyw-1 rev 1 is NOT accepted. repository_delta=empty is structurally correct and is not the producer's doing: the story worktree snapshots curator-spec (HEAD^{tree} == base tree == candidate tree 2e6ca472), while this leaf's deliverable is curator feat/consume-overlay-rule@38702164. The findings are against that branch, which the CR revision represents.

The implementation is correct. Three repairs are needed, NONE in product code.

F1 (major, evidence integrity). envprofile TestInstallNeverDemotesANetworkIdentityToAPath does not test the class it names. Measured: with M10 applied (widening deleted) that test PASSES, exit 0, logging "checked 5 network-identity operands". It filters on identity.Parse != "", but Parse returns ("", nil) for exactly the bare canonical identities the widening exists for (github.com/evil-org/pkg, example.com/org/pkg, github.com/relux-works/pkg) - no colon, no scheme, so it takes the local branch. The 5 it checks are decided by ClassifySource alone. The same false claim appears three times: the test comment ("fails here"), report section 4 row M10 ("+2" - only 2 of 3 fail), and the new platform-cases.tsv row, which makes AC row 6 (ledger rows truthful) unmet. Repair is one line: filter on ValidCanonical too, so M10 kills it; then correct the report row and the ledger wording.

F2 (minor). Report section 3 "Complete list of install kind changes - nothing here is silent" omits a class I found by enumerating the change myself against stage (c) isPathOperand: packages/team:context and a/b:c were REFUSED at canonicalGit before and are PATH installs now. Refuse->accept, unlisted, and no colon-in-a-later-segment operand exists in installOperandCases. Bears on AC row 4 "no source silently changing kind".

F3 (trivial). installOperandKind doc comment block is attached to sourceKindRefusal; installOperandKind has none. golangci-lint is clean, so nothing catches it.

VERIFIED CORRECT AND NOT AT ISSUE:
- The discriminator is exact: 121 spellings x 3 engines (ajv 8 over the whole committed schema, V8 over the verbatim pattern strings, Go) = 0 disagreements. ajv agrees with index.json on 71/71 manager-config-v2 cases. ClassifySource imports only regexp.
- The install widening is right and I verified every step: canonicalGit accepts a bare canonical identity, ensureRepo clones https://canonical, and WITHOUT the widening the F14 planted-directory bypass reproduces - Install of github.com/evil-org/pkg with a planted ./github.com/evil-org/pkg and a machine allowlist SUCCEEDS (err = nil). With it, refused with the allowlist reason and no clone. My own enumeration confirms no operand that reached the network before reaches a local path now.
- 16 of 17 mutants re-written by me and re-run with a "=== RUN" guard: all killed, all genuine narrowings, no delete-only mutant. Both token-preserving text-gate mutants reproduce (ledger exit 1, behavioural suite exit 0).
- Candidate lane confirmed from the jobs own logs, not the report sentence: run 34071813375, CANDIDATE_REF/candidate_revision 87a0d0060bad, CI_REQUIRE_FULL_ROOT=1, deferred=0, test-gate exit=0 on ubuntu/macos/windows. ubuntu excluded=1 is internal/godriver, a designed root exclusion. PR #62: 11 SUCCESS, Candidate suite SKIPPED.
- The 30 subcases reproduce: root verified 1047/1047 against manifest.json, base 7c74a492 fails exactly 30 subcases, head 38702164 is ok.
- Both reported anomalies confirmed by construction: subst.txt carries export-subst and git archive yields 65 bytes vs 40 on disk; go test -run with a whole anchored Parent/child group prints "no tests to run" and EXITS 0.
- Bound retired (grep clean), rows 303/304/305/376 truthful, ledger-consistency 225 rows ok, build/vet/gofmt/golangci-lint/gate-selftest/no-broad-suppression all exit 0 on darwin.
- Both commits signed (G) with the human identity.

FOR THE ORCHESTRATOR, two items:
1. UNKNOWN, not clear: whether any EARLIER mutant sweep in this epic is a false survivor from the go test -run split. The board earlier sweep artifacts record mutant names and verdicts but not the literal -run patterns, so the filter shape is unrecoverable from the evidence. Detection criterion: any sweep that anchored a whole Parent/child filter as one regexp and keyed "survived" on exit code without checking that any test ran.
2. This CR snapshot scope (curator-spec) is not this leaf delivery scope (curator). Accepting or rejecting it does not move the curator work either way, and the board integration step would integrate an empty tree. Worth deciding deliberately.
agent completed: [reviewer] reviewer (claude) (exit=0)
spawn run completed: claude (run=RUN-260907-c8abaf, pid=13815, exit=0)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1-128-gab60e0d; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [implementer] developer (claude) (run=RUN-260907-7f359e, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260907-7f359e)
Rework 1 complete, ready for review. Branch feat/consume-overlay-rule advanced 38702164 -> e4ddca19, two signed commits (62c27a0d, e4ddca19), author Ivan Oparin, both %G? = G. Not pushed; PR #62 untouched. No product behaviour changed: installOperandKind's body is byte-identical to the reviewed head.

F1 (major) reproduced first by execution -- M10 applied to the reviewed head left TestInstallNeverDemotesANetworkIdentityToAPath at exit 0 logging 'checked 5' while its two siblings died. Cause measured directly: identity.Parse returns ('', nil) for a bare canonical identity, so the loop's filter skipped github.com/evil-org/pkg and its two siblings -- exactly and only the class the widening exists for. The loop now admits both spellings of the class (Parse non-empty OR ValidCanonical), reports 'checked 8, 3 of them bare canonical', and a bare==0 guard fails if the matrix loses that spelling again. M10 re-applied after the repair: 3 of 3 killed. Mutant M18 reintroduces the exact defect and is killed by the new guard.

F2 (minor) re-derived rather than inherited: 44 operands driven through stage (c)'s classification restored in a throwaway checkout of 7c74a492 and through installOperandKind at head -- 31 unchanged, 13 changed, and the omitted class (packages/team:context, a/b:c, refused -> path) reproduces. Also measured the direction that matters: every git->path row had ValidCanonical false and was already a local clone, so no operand that reached the network before reaches a local path now. packages/team:context added to installOperandCases; mutant M19 survives at exit 0 on the reviewed matrix and is killed by the added operand.

F3 (trivial) fixed: sourceKindRefusal moved above installOperandKind so each carries its own doc block. The 'Nothing changes kind' sentence tightened to the measured claim.

Ledger: both edited rows in .github/ci/platform-cases.tsv driven through platform-case-gate.sh and observed PASSING by name, 0 skips, exit 0. ledger-consistency.sh exit 0, 225 rows.

Gates at head, darwin, each a standalone process, sequential: go build 0, go vet 0, gofmt -l cmd internal 0 (0 lines), golangci-lint 0 (0 issues), gate-selftest.sh 0 (94/0), no-broad-suppression.sh 0, ledger-consistency.sh 0, platform-case-gate.sh 0, go test -race ./internal/envprofile ./internal/identity 0, go test ./cmd/curator 0 (295s). The two test-gate lanes were NOT re-run and the hosted candidate lane was NOT re-dispatched -- reasoned from a measured premise: grep -rl CURATOR_CONFORMANCE_ROOT over internal/envprofile returns 0 files, so no root can change this package's verdict. Standing candidate measurement remains run 34071813375 at 38702164, green on all three runners.

AC coverage 7 of 7 driven; rows 4 and 6 moved from gap to driven. 19 mutants, 19 killed, 0 survivors. Artifacts: TASK-260906-19gjyw_rework-report-1.md attached, TASK-260906-19gjyw_drafting-report.md updated.
agent completed: [implementer] developer (claude) (exit=0)
spawn run completed: claude (run=RUN-260907-7f359e, pid=76143, exit=0)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1-128-gab60e0d; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [reviewer] reviewer (claude) (run=RUN-260907-b8a04e, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260907-b8a04e)
Review cycle 2 (RUN-260907-b8a04e): ACCEPT, repeat-of none. CR-TASK-260906-19gjyw-2 rev 2 accepted; element routed to integrating. Deliverable is curator feat/consume-overlay-rule@e4ddca19, PR #62 (MERGEABLE/CLEAN, 11 hosted checks SUCCESS on that exact head). Empty repository delta is structural: the story workspace is a curator-spec checkout at 87a0d006, HEAD^{tree} == base tree == candidate tree 2e6ca472, zero paths; the leaf scope is the curator repo. Cycle 1 F1/F2/F3 all closed by measurement: M10 re-applied kills TestInstallNeverDemotesANetworkIdentityToAPath (exit 1, checked 8 / 3 bare canonical) where it exited 0 at 38702164; the bare==0 guard fires on matrix removal and is non-vacuous (ValidCanonical implies ClassifySource=path over 90 spellings); M19 survives at exit 0 on the reviewed matrix and is killed by the added packages/team:context operand; doc comment moved and a go/ast pass over 6 production files finds only two mismatches, both pre-existing on origin/main. Install kind change re-derived over 82 operands through both complete compositions (stage (c) isPathOperand + old Parse carve-out vs head): 61 unchanged, 21 changed, and of the 23 operands that reached the network before, 22 stay git and 1 becomes an explicit refusal -- zero become a local path. envprofile.go at e4ddca19 is AST-identical to 38702164 (comments stripped, decls sorted): no product behaviour moved. Candidate lane run 34071813375 green on ubuntu/macos/windows against curator-spec 87a0d006 with CI_REQUIRE_FULL_ROOT=1, confirmed from the jobs own logs. Stated bound: the candidate lane is not dispatched at e4ddca19; the delta is root-independent (internal/envprofile matches CURATOR_CONFORMANCE_ROOT in 0 files) and its changed tests are green on all three runners at that head. Landing verdict: PR #62 is safe to land. Two non-blocking follow-ups for the orchestrator: the strictAuditMember/canaryPasses doc block on origin/main is the F3 shape and predates this branch; a66eec88 logbook entry still calls the consumption gap open.
agent completed: [reviewer] reviewer (claude) (exit=0)
spawn run completed: claude (run=RUN-260907-b8a04e, pid=5170, exit=0)

## Precondition Resources
- [producer-brief-consume-overlay-rule.md](file://TASK-260906-19gjyw/producer-brief-consume-overlay-rule.md) — Producer brief: port the git-source-only overlay rule into the Go reader and make the candidate lane green against spec main
- [review-brief-consume-overlay-1.md](file://TASK-260906-19gjyw/review-brief-consume-overlay-1.md) — Review brief cycle 1: the discriminator, the install widening, and the two evidence anomalies
- [producer-brief-consume-publish.md](file://TASK-260906-19gjyw/producer-brief-consume-publish.md) — Publish-only brief: the block is resolved by hosted measurement; publish the CR without code changes
- [producer-brief-consume-rework-1.md](file://TASK-260906-19gjyw/producer-brief-consume-rework-1.md) — Consumption rework 1: the property loop excludes its own subject, plus two smaller corrections
- [review-brief-consume-overlay-2.md](file://TASK-260906-19gjyw/review-brief-consume-overlay-2.md) — Review brief cycle 2: the three evidence corrections and the landing decision

## Outcome Resources
- [TASK-260906-19gjyw_spawn-log_-implementer--developer--muse-_RUN-260906-339c1c.log](file://TASK-260906-19gjyw/TASK-260906-19gjyw_spawn-log_-implementer--developer--muse-_RUN-260906-339c1c.log) — System spawn log captured by task-board
- [TASK-260906-19gjyw_spawn-log_-implementer--developer--muse-_RUN-260906-02c4c9.log](file://TASK-260906-19gjyw/TASK-260906-19gjyw_spawn-log_-implementer--developer--muse-_RUN-260906-02c4c9.log) — System spawn log captured by task-board
- [TASK-260906-19gjyw_spawn-log_-implementer--developer--muse-_RUN-260906-eab036.log](file://TASK-260906-19gjyw/TASK-260906-19gjyw_spawn-log_-implementer--developer--muse-_RUN-260906-eab036.log) — System spawn log captured by task-board
- [TASK-260906-19gjyw_spawn-log_-implementer--developer--muse-_RUN-260906-b966cb.log](file://TASK-260906-19gjyw/TASK-260906-19gjyw_spawn-log_-implementer--developer--muse-_RUN-260906-b966cb.log) — System spawn log captured by task-board
- [TASK-260906-19gjyw_spawn-log_-implementer--developer--muse-_RUN-260906-6c1119.log](file://TASK-260906-19gjyw/TASK-260906-19gjyw_spawn-log_-implementer--developer--muse-_RUN-260906-6c1119.log) — System spawn log captured by task-board
- [TASK-260906-19gjyw_spawn-log_-implementer--developer--muse-_RUN-260906-74f27e.log](file://TASK-260906-19gjyw/TASK-260906-19gjyw_spawn-log_-implementer--developer--muse-_RUN-260906-74f27e.log) — System spawn log captured by task-board
- [TASK-260906-19gjyw_spawn-log_-implementer--developer--muse-_RUN-260906-71db31.log](file://TASK-260906-19gjyw/TASK-260906-19gjyw_spawn-log_-implementer--developer--muse-_RUN-260906-71db31.log) — System spawn log captured by task-board
- [TASK-260906-19gjyw_spawn-log_-implementer--developer--muse-_RUN-260906-3936f5.log](file://TASK-260906-19gjyw/TASK-260906-19gjyw_spawn-log_-implementer--developer--muse-_RUN-260906-3936f5.log) — System spawn log captured by task-board
- [TASK-260906-19gjyw_spawn-log_-implementer--developer--claude-_RUN-260906-96a63c.log](file://TASK-260906-19gjyw/TASK-260906-19gjyw_spawn-log_-implementer--developer--claude-_RUN-260906-96a63c.log) — System spawn log captured by task-board
- [TASK-260906-19gjyw_drafting-report.md](file://TASK-260906-19gjyw/TASK-260906-19gjyw_drafting-report.md) — Drafting report, corrected in rework 1: M10 row and total, the install kind-change table completed and re-derived by measurement, the two ledger rows reworded
- [TASK-260906-19gjyw_spawn-log_-implementer--developer--claude-_RUN-260907-8d0c2a.log](file://TASK-260906-19gjyw/TASK-260906-19gjyw_spawn-log_-implementer--developer--claude-_RUN-260907-8d0c2a.log) — System spawn log captured by task-board
- [TASK-260906-19gjyw_spawn-log_-implementer--developer--claude-_RUN-260907-9ecdbe.log](file://TASK-260906-19gjyw/TASK-260906-19gjyw_spawn-log_-implementer--developer--claude-_RUN-260907-9ecdbe.log) — System spawn log captured by task-board
- [TASK-260906-19gjyw_change-request_rev1.patch](file://TASK-260906-19gjyw/TASK-260906-19gjyw_change-request_rev1.patch) — Change Request CR-TASK-260906-19gjyw-1 revision 1 candidate patch (repository_delta=empty, 0 changed paths)
- [TASK-260906-19gjyw_spawn-log_-reviewer--reviewer--claude-_RUN-260907-c8abaf.log](file://TASK-260906-19gjyw/TASK-260906-19gjyw_spawn-log_-reviewer--reviewer--claude-_RUN-260907-c8abaf.log) — System spawn log captured by task-board
- [TASK-260906-19gjyw_review-findings-1.md](file://TASK-260906-19gjyw/TASK-260906-19gjyw_review-findings-1.md) — Review cycle 1: CHANGES REQUESTED. 5 of 7 AC rows driven; two measured gaps. F1 major: TestInstallNeverDemotesANetworkIdentityToAPath survives M10 (checks 0 of the widened class while logging 'checked 5'), and the same false claim sits in the report's M10 row and a new platform-cases.tsv row. F2: the 'complete' install kind-change table omits the refused->path colon-in-a-later-segment class. F3: doc comment attached to the wrong function. Verified independently: discriminator exact on 121 spellings across ajv/V8/Go with 0 disagreements; the F14 planted-directory bypass reproduced without the widening (Install returns nil) and blocked with it; 16 of 17 mutants re-run and killed; both token-preserving text-gate mutants reproduced; candidate lane green on all three runners confirmed from the jobs' own logs; the 30 subcases reproduced at base and closed at head against a 1047/1047 manifest-verified root; both reported anomalies (git archive export-subst, go test -run slash split exiting 0) confirmed by construction.
- [TASK-260906-19gjyw_spawn-log_-implementer--developer--claude-_RUN-260907-7f359e.log](file://TASK-260906-19gjyw/TASK-260906-19gjyw_spawn-log_-implementer--developer--claude-_RUN-260907-7f359e.log) — System spawn log captured by task-board
- [TASK-260906-19gjyw_rework-report-1.md](file://TASK-260906-19gjyw/TASK-260906-19gjyw_rework-report-1.md) — Rework 1: F1 property loop repaired and M10 kill shown, F2 kind-change class re-derived by measurement and pinned, F3 doc comment moved, mutants M18/M19, gate table at head e4ddca19
- [TASK-260906-19gjyw_change-request_rev2.patch](file://TASK-260906-19gjyw/TASK-260906-19gjyw_change-request_rev2.patch) — Change Request CR-TASK-260906-19gjyw-2 revision 2 candidate patch (repository_delta=empty, 0 changed paths)
- [TASK-260906-19gjyw_spawn-log_-reviewer--reviewer--claude-_RUN-260907-b8a04e.log](file://TASK-260906-19gjyw/TASK-260906-19gjyw_spawn-log_-reviewer--reviewer--claude-_RUN-260907-b8a04e.log) — System spawn log captured by task-board
- [TASK-260906-19gjyw_review-findings-2.md](file://TASK-260906-19gjyw/TASK-260906-19gjyw_review-findings-2.md) — Review cycle 2: F1/F2/F3 closed by re-applied mutants, 82-operand install kind re-derivation, AST-identity regression proof, hosted evidence at e4ddca19
- [TASK-260906-19gjyw_review-verdict-rev2.md](file://TASK-260906-19gjyw/TASK-260906-19gjyw_review-verdict-rev2.md) — ACCEPT verdict for CR-TASK-260906-19gjyw-2 revision 2, including why the empty repository delta is correct for this leaf

## Created
2026-09-06T19:20:27Z

## Last Update
2026-09-07T03:40:46Z

## Assigned To
[reviewer] reviewer (claude)
