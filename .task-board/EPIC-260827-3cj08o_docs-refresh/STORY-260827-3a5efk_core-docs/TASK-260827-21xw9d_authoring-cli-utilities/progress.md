## Status
done

## Review
required

## Task Class
docs

## Estimate
estimated(fibonacci(5))

## Blocked By
- (none)

## Blocks
- TASK-260827-xdbobc

## Checklist
- [x] Admission matrix from code with grep evidence; three worked examples validated against tree binary; planned-language error named
- [x] Docs updated and consistent with current code
- [x] No discrepancies between code and description
- [x] Result linked as a new task-scoped outcome resource
- [x] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
- [x] Implementation matches AC
- [x] Solution fits project architecture
- [x] Tests green
- [x] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches

## Notes
Coordination corrected after inspecting the adapter branch: docs/authoring-language-adapters.md there is CONTRIBUTOR-facing (how to add a source-closure adapter to Curator) and does not overlap this task, which stays a standalone AUTHOR-facing document: how to ship CLI utilities in a skill per supported language. After the branch lands, the supported set grows beyond go: the branch admits go-v1, go-repository-v1, node-v1, bash-v1, script-worker-v1, custom-v1 (verify against landed main at execution time). Primary sources then: .spec/skill-facing-cli-source-closure.md (skill-facing contract), docs/source-closure-adapter-conformance.md (supported profiles, unsupported cases, diagnostics: link, do not duplicate), and the admission code. The task runs after the adapter branch reaches origin/main.
spawn agent resolution: Agent selection: agy via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=agy; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [implementer] doc-writer (agy) (run=RUN-260828-7c08ce, max_parallel=20)
spawn run started: [implementer] doc-writer (agy) (run=RUN-260828-7c08ce)
agent completed: [implementer] doc-writer (agy) (exit=0)
spawn run completed: agy (run=RUN-260828-7c08ce, pid=6079, exit=0)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [reviewer] reviewer (claude) (run=RUN-260828-c4e4c4, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260828-c4e4c4)
Review rev1 (RUN-260828-c4e4c4): CHANGES REQUESTED. Evidence: TASK-260827-21xw9d_review-verdict.md. Four blocking factual defects, all reproduced against a binary built from the candidate tree. D1 docs/authoring-cli-commands.md:40 names package_build_command_influence_forbidden, which exists nowhere in the repo; real identifier is build_execution_package_influence_forbidden (internal/godriver/controls.go:18). The task results resource cites the correct constant, so the doc contradicts its own evidence. D2 all three resulting-shim blocks are invented output: installed the doc Example 3 skill and read .agents/bin/run-script, which carries a PATH preamble the doc omits, targets $HOME/.curator/runtime/<skill>/<commit>/scripts/run.sh rather than .agents/skills/..., and single-quotes the path; .curator/cache/builds/mytool is not a path Curator creates (real: <home>/cache/build/<driver>/<sha256> per internal/buildcache/cache.go:262). D3 line 162 states the .git suffix is required; removing it from the doc own Example 2 still gives skill check exit 0, buildrepo.go:127 trims it. Real rule is the HTTPS 301 guidance at docs/build-https.md:140. D4 line 250 names kotlin-v1; LOGBOOK.md:2849 reserves kotlin-native-v1. WHAT HOLDS: every other diagnostic identifier resolves in internal/ with a non-test call site; build vectors match build.go:32-36 literally; script-worker-v1 refusal reproduced; all three worked-example manifests validate (curator skill check, exit 0 each); planned-driver error text exact for kotlin/swift/rust. Docs-only delta, zero .go files changed, no suite rerun.
agent completed: [reviewer] reviewer (claude) (exit=0)
spawn run completed: claude (run=RUN-260828-c4e4c4, pid=9087, exit=0)
spawn agent resolution: Agent selection: agy via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=agy; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [implementer] doc-writer (agy) (run=RUN-260828-df7ee9, max_parallel=20)
spawn run started: [implementer] doc-writer (agy) (run=RUN-260828-df7ee9)
agent completed: [implementer] doc-writer (agy) (exit=0)
spawn run completed: agy (run=RUN-260828-df7ee9, pid=14216, exit=0)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [reviewer] reviewer (claude) (run=RUN-260828-a70495, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260828-a70495)
Rev2 review (RUN-260828-a70495): changes requested. D1 (build_execution_package_influence_forbidden), D3 (.git suffix restated as HTTPS 301 advice), D4 (kotlin-native-v1 plus the -repository-v1 halves) are all fixed and reverified. D2 is fixed for the shim contract and both local paths - the real shim was installed and read, and the build cache path $HOME/.curator/cache/build/go-v1/<hash> checks out against buildcache/cache.go:262 - but one invented path survives at docs/authoring-cli-commands.md:176. D5 (blocking): $HOME/.curator/cache/build_repositories/.../artifacts/<hash>/artifact does not exist; build_repositories is only a manifest key, never a path segment. The real root is install/external.go:105, filepath.Join(home, "external-build-cache"), so the path is $HOME/.curator/external-build-cache/artifacts/<sha256-hex>/artifact (targets.go:185-186, gc.go:125). One-line fix, plus naming the same store at line 68. Everything else in the document was reverified and holds: all 14 diagnostic identifiers resolve at non-test sites, all three worked examples validate exit=0 against the tree binary, all six reserved driver identities produce the documented error verbatim, the mutual execution_policy/interpreter errors and the script-worker refusal were driven not assumed. No .go file in the delta, so no suite rerun. Evidence: TASK-260827-21xw9d_review-verdict-rev2.md
agent completed: [reviewer] reviewer (claude) (exit=0)
spawn run completed: claude (run=RUN-260828-a70495, pid=15063, exit=0)
spawn agent resolution: Agent selection: agy via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=agy; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [implementer] doc-writer (agy) (run=RUN-260828-dd6ef1, max_parallel=20)
spawn run started: [implementer] doc-writer (agy) (run=RUN-260828-dd6ef1)
Rework rev2 completed. Fixed defect D5 by correcting the external build repository artifact store path in docs/authoring-cli-commands.md to $HOME/.curator/external-build-cache/artifacts/<sha256-hex>/artifact. Verified with grep search and curator skill check against tree binary. Ready for review.
agent completed: [implementer] doc-writer (agy) (exit=0)
spawn run completed: agy (run=RUN-260828-dd6ef1, pid=42192, exit=0)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [reviewer] reviewer (claude) (run=RUN-260828-a00bc7, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260828-a00bc7)
Rev3 review (CR revision 3, candidate f88a627d): changes requested. D5 fixed and independently reverified (external-build-cache path at lines 68 and 176 matches internal/install/external.go:105 + targets.go:185-186; cache/build_repositories gone from the repo). New blocking D6: docs/authoring-cli-commands.md:40 attributes a manifest-declared ldflags/gcflags/env/hooks rejection to build_execution_package_influence_forbidden, but the parser refuses it first with skill.manifest_invalid "field is not supported for build commands" (rejectUnknownBuildFields, parse.go:839-855); the godriver diagnostic is unreachable from a manifest because commandObject is rebuilt from the parsed command (plan.go:561-571, comment states it). Driven, not grepped. Same class as rev1 D1 and rev2 D5. Everything else in the document reverified and holding: 14 identifiers, build vectors, network-off env, Linux refusal, interpreter set, script-worker refusal, both shim branches, all three worked examples exit 0 against a tree binary, six reserved driver identities verbatim, README link, prose style. Fix text and reproduction in TASK-260827-21xw9d_rework-instructions-rev3.md; full evidence in TASK-260827-21xw9d_review-verdict-rev3.md.
agent completed: [reviewer] reviewer (claude) (exit=0)
spawn run completed: claude (run=RUN-260828-a00bc7, pid=42970, exit=0)
spawn agent resolution: Agent selection: agy via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=agy; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [implementer] doc-writer (agy) (run=RUN-260828-13f0b4, max_parallel=20)
spawn run started: [implementer] doc-writer (agy) (run=RUN-260828-13f0b4)
agent completed: [implementer] doc-writer (agy) (exit=0)
spawn run completed: agy (run=RUN-260828-13f0b4, pid=47137, exit=0)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [reviewer] reviewer (claude) (run=RUN-260828-10b979, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260828-10b979)
Rev4 ACCEPTED (reviewer run RUN-260828-10b979). CR-TASK-260827-21xw9d-4, base 41ab53cd, candidate 7e43aa4e. Delta from rev3 is one line: the prescribed D6 fix at docs/authoring-cli-commands.md:40. Verification binary built from the candidate tree; worktree byte-compared against candidate for both scope files (MATCH). D6 driven, not grepped: Example 1 manifest + ldflags now yields exactly the documented error: skill.manifest_invalid (agent-skill.json): commands.mytool.ldflags: field is not supported for build commands. Same for gcflags/env/hooks/pre_build/post_build; the schema-6 vs schema-8 boundary for modules was driven too. Whole document reverified this round, not accepted from prior verdicts: 14 diagnostic identifiers at non-test call sites, build vectors, network-off env, Linux refusal, interpreter set and rationale, mutual execution_policy/interpreter errors, script-worker-v1 refusal, both WriteBinShim branches, all three filesystem paths re-derived from code, all three worked examples curator skill check exit 0, all six reserved driver identities verbatim, README link at README.md:60, prose style clean. Zero .go files in the delta so no suite rerun. No commit_ack supplied; orchestrator makes the done transition. Cross-task note for the orchestrator: four revisions of this task all failed on the same class, a code fact stated without being driven (rev1 D1 invented error identifier, rev2 D5 invented filesystem path, rev3 D6 correct identifier attributed to a cause that cannot fire). Worth a LOGBOOK entry by the commit-owning mover, since a reviewer edit would diverge the worktree from the accepted candidate OID.
agent completed: [reviewer] reviewer (claude) (exit=0)
spawn run completed: claude (run=RUN-260828-10b979, pid=49240, exit=0)
Accepted per rev4 verdict; done set by orchestrator per CR contract.

## Precondition Resources
- [TASK-260827-21xw9d_docs-refresh.md](file://TASK-260827-21xw9d/TASK-260827-21xw9d_docs-refresh.md) — precondition
- [TASK-260827-21xw9d_rework-instructions.md](file://TASK-260827-21xw9d/TASK-260827-21xw9d_rework-instructions.md) — Rev1 verdict: four blocking factual defects D1-D4; apply each prescribed fix, verify with grep and the tree binary, literal outputs
- [TASK-260827-21xw9d_rework-instructions-rev2.md](file://TASK-260827-21xw9d/TASK-260827-21xw9d_rework-instructions-rev2.md) — Rev2 verdict: one blocking defect D5; fix the external artifact store path at docs/authoring-cli-commands.md:176 and name the same store at line 68; change nothing else
- [TASK-260827-21xw9d_rework-instructions-2.md](file://TASK-260827-21xw9d/TASK-260827-21xw9d_rework-instructions-2.md) — Rev2 verdict: apply every remaining finding exactly, verify each with grep and the tree binary
- [TASK-260827-21xw9d_rework-instructions-rev3.md](file://TASK-260827-21xw9d/TASK-260827-21xw9d_rework-instructions-rev3.md) — Rev3 verdict: one blocking defect D6; fix the closed-build-parameters bullet at docs/authoring-cli-commands.md:40 (wrong error identifier for a manifest-declared influence field), verify with the tree binary; change nothing else
- [TASK-260827-21xw9d_rework-instructions-3.md](file://TASK-260827-21xw9d/TASK-260827-21xw9d_rework-instructions-3.md) — Rev3 verdict: apply the prescribed replacement text exactly where given; verify each with grep; nothing else

## Outcome Resources
- [TASK-260827-21xw9d_spawn-log_-implementer--doc-writer--agy-_RUN-260828-7c08ce.log](file://TASK-260827-21xw9d/TASK-260827-21xw9d_spawn-log_-implementer--doc-writer--agy-_RUN-260828-7c08ce.log) — System spawn log captured by task-board
- [TASK-260827-21xw9d_results.md](file://TASK-260827-21xw9d/TASK-260827-21xw9d_results.md) — Results summary for authoring CLI commands documentation task including rev3 D6 fix
- [TASK-260827-21xw9d_change-request_rev1.patch](file://TASK-260827-21xw9d/TASK-260827-21xw9d_change-request_rev1.patch) — Change Request CR-TASK-260827-21xw9d-1 revision 1 candidate patch (repository_delta=present, 8 changed paths)
- [TASK-260827-21xw9d_spawn-log_-reviewer--reviewer--claude-_RUN-260828-c4e4c4.log](file://TASK-260827-21xw9d/TASK-260827-21xw9d_spawn-log_-reviewer--reviewer--claude-_RUN-260828-c4e4c4.log) — System spawn log captured by task-board
- [TASK-260827-21xw9d_review-verdict.md](file://TASK-260827-21xw9d/TASK-260827-21xw9d_review-verdict.md) — Rev4 review verdict: accepted; D6 fixed and driven, full document reverified against the tree binary
- [TASK-260827-21xw9d_spawn-log_-implementer--doc-writer--agy-_RUN-260828-df7ee9.log](file://TASK-260827-21xw9d/TASK-260827-21xw9d_spawn-log_-implementer--doc-writer--agy-_RUN-260828-df7ee9.log) — System spawn log captured by task-board
- [TASK-260827-21xw9d_change-request_rev2.patch](file://TASK-260827-21xw9d/TASK-260827-21xw9d_change-request_rev2.patch) — Change Request CR-TASK-260827-21xw9d-2 revision 2 candidate patch (repository_delta=present, 8 changed paths)
- [TASK-260827-21xw9d_spawn-log_-reviewer--reviewer--claude-_RUN-260828-a70495.log](file://TASK-260827-21xw9d/TASK-260827-21xw9d_spawn-log_-reviewer--reviewer--claude-_RUN-260828-a70495.log) — System spawn log captured by task-board
- [TASK-260827-21xw9d_review-verdict-rev2.md](file://TASK-260827-21xw9d/TASK-260827-21xw9d_review-verdict-rev2.md) — Rev2 reviewer verdict: D1/D3/D4 fixed, D2 partially fixed; blocking D5 - external artifact store path cache/build_repositories is invented, real root is external-build-cache
- [TASK-260827-21xw9d_spawn-log_-implementer--doc-writer--agy-_RUN-260828-dd6ef1.log](file://TASK-260827-21xw9d/TASK-260827-21xw9d_spawn-log_-implementer--doc-writer--agy-_RUN-260828-dd6ef1.log) — System spawn log captured by task-board
- [TASK-260827-21xw9d_change-request_rev3.patch](file://TASK-260827-21xw9d/TASK-260827-21xw9d_change-request_rev3.patch) — Change Request CR-TASK-260827-21xw9d-3 revision 3 candidate patch (repository_delta=present, 8 changed paths)
- [TASK-260827-21xw9d_spawn-log_-reviewer--reviewer--claude-_RUN-260828-a00bc7.log](file://TASK-260827-21xw9d/TASK-260827-21xw9d_spawn-log_-reviewer--reviewer--claude-_RUN-260828-a00bc7.log) — System spawn log captured by task-board
- [TASK-260827-21xw9d_review-verdict-rev3.md](file://TASK-260827-21xw9d/TASK-260827-21xw9d_review-verdict-rev3.md) — Rev3 verdict: D5 fixed and reverified; one new blocking defect D6 (closed-build-parameters bullet attributes the wrong error identifier, driven not grepped); full reverification of sections 2-5
- [TASK-260827-21xw9d_spawn-log_-implementer--doc-writer--agy-_RUN-260828-13f0b4.log](file://TASK-260827-21xw9d/TASK-260827-21xw9d_spawn-log_-implementer--doc-writer--agy-_RUN-260828-13f0b4.log) — System spawn log captured by task-board
- [TASK-260827-21xw9d_change-request_rev4.patch](file://TASK-260827-21xw9d/TASK-260827-21xw9d_change-request_rev4.patch) — Change Request CR-TASK-260827-21xw9d-4 revision 4 candidate patch (repository_delta=present, 8 changed paths)
- [TASK-260827-21xw9d_spawn-log_-reviewer--reviewer--claude-_RUN-260828-10b979.log](file://TASK-260827-21xw9d/TASK-260827-21xw9d_spawn-log_-reviewer--reviewer--claude-_RUN-260828-10b979.log) — System spawn log captured by task-board
- [TASK-260827-21xw9d_review-verdict-rev4.md](file://TASK-260827-21xw9d/TASK-260827-21xw9d_review-verdict-rev4.md) — Rev4 accepted verdict: D6 reproduction, full document reverification, driven outputs

## Created
2026-08-27T02:49:18Z

## Last Update
2026-08-28T03:09:34Z

## Assigned To
[reviewer] reviewer (claude)
