## Status
done

## Review
required

## Task Class
docs

## Estimate
estimated(fibonacci(5))

## Blocked By
- TASK-260827-2232c0

## Blocks
- TASK-260827-xdbobc

## Checklist
- [x] cli.md verbatim vs tree binary; troubleshooting error strings grep in internal/; README Commands section landed
- [x] Docs updated and consistent with current code
- [x] No discrepancies between code and description
- [x] Result linked as a new task-scoped outcome resource
- [x] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
- [x] Implementation matches AC
- [x] Solution fits project architecture
- [x] Tests green
- [x] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches

## Notes
spawn agent resolution: Agent selection: agy via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=agy; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [implementer] doc-writer (agy) (run=RUN-260828-d13a5f, max_parallel=20)
spawn run started: [implementer] doc-writer (agy) (run=RUN-260828-d13a5f)
agent completed: [implementer] doc-writer (agy) (exit=124)
spawn run completed: agy (run=RUN-260828-d13a5f, pid=33706, exit=124)
spawn run RUN-260828-d13a5f failed; operator action required; failure: run exceeded --timeout 30m0s and was terminated by the launcher; a child process could not be proven terminated and may still be running
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [reviewer] reviewer (claude) (run=RUN-260828-73644d, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260828-73644d)
agent completed: [reviewer] reviewer (claude) (exit=143)
spawn run completed: claude (run=RUN-260828-73644d, pid=81363, exit=143)
spawn agent resolution: Agent selection: agy via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=agy; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [implementer] doc-writer (agy) (run=RUN-260828-8a8681, max_parallel=20)
spawn run started: [implementer] doc-writer (agy) (run=RUN-260828-8a8681)
spawn run RUN-260828-73644d cancelled by operator; operator action required; reason: no operator reason supplied
agent completed: [implementer] doc-writer (agy) (exit=0)
spawn run completed: agy (run=RUN-260828-8a8681, pid=82026, exit=0)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [reviewer] reviewer (claude) (run=RUN-260828-6e9c4a, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260828-6e9c4a)
Review rev1: CHANGES REQUESTED. docs/cli.md claims verbatim verification but six entries do not match cmd/curator/main.go: project add (documented flags all rejected, real --agents missing, arity is <alias> <path>), global add --project (undefined), hybrid status --check/--json (no flagset at all, always exits 0 - fail-open CI gate), global install/upgrade --all (explicit usage error), skill check --locale code default en (real: -locale string, empty default), list/ui [path] (cmdList/cmdUI take no args). docs/troubleshooting.md maps two Git admission errors (ValidateGitTool, admission.go:203/211) to a Go toolchain cause with a CURATOR_GO remedy that cannot fix either, and omits the real toolchain preflight diagnostics (untrusted_go_executable, toolchain_executable_mismatch). Correct: all builds.go line refs, the six literal error identifiers, the gc grace period, README Commands covers all 18 groups, links resolve, style guide clean, README doc guard tests green. Evidence: TASK-260827-1nauj3_review-verdict.md
agent completed: [reviewer] reviewer (claude) (exit=0)
spawn run completed: claude (run=RUN-260828-6e9c4a, pid=86692, exit=0)
spawn agent resolution: Agent selection: agy via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=agy; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [implementer] doc-writer (agy) (run=RUN-260828-acf233, max_parallel=20)
spawn run started: [implementer] doc-writer (agy) (run=RUN-260828-acf233)
agent completed: [implementer] doc-writer (agy) (exit=0)
spawn run completed: agy (run=RUN-260828-acf233, pid=94610, exit=0)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [reviewer] reviewer (claude) (run=RUN-260828-a12dbd, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260828-a12dbd)
Rev2 review: CHANGES REQUESTED. Evidence: TASK-260827-1nauj3_review-verdict-rev2.md. Rev1 F1/F3/F4/F5/F6/F7 fixed. BLOCKING: (G1) rev1 F2 untouched - docs/cli.md:331 still documents global add --project, binary rejects it; (G2) curator add omits real --project string; (G3) curator install omits real --all; (G4) curator upgrade omits real --all; (G5) curator status omits real --all; (G6) README.md:137 describes curator project add as adding a skill declaration - it registers a project alias and path. Six targeted edits listed in the verdict. Guard tests green: go test ./cmd/curator -run TestEveryCurrentnessCodeIsDocumented|TestInputCausesAreDistinctAndDocumented -> ok 0.618s.
agent completed: [reviewer] reviewer (claude) (exit=0)
spawn run completed: claude (run=RUN-260828-a12dbd, pid=95735, exit=0)
spawn agent resolution: Agent selection: agy via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=agy; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [implementer] doc-writer (agy) (run=RUN-260828-7e276b, max_parallel=20)
spawn run started: [implementer] doc-writer (agy) (run=RUN-260828-7e276b)
agent completed: [implementer] doc-writer (agy) (exit=0)
spawn run completed: agy (run=RUN-260828-7e276b, pid=98675, exit=0)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [reviewer] reviewer (claude) (run=RUN-260828-e0e8b1, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260828-e0e8b1)
Reviewer rev3 (RUN-260828-e0e8b1): ACCEPTED. All six rev2 rework items (G1-G6) plus notes 4 and 6 verified fixed against bin/curator rebuilt with make build; per-command -h probes run from /tmp/curtest3 outside the tree, so no curator init side effect. Every documented flag name, value type, default, and positional arity matches cmd/curator/main.go. All 13 builds.go and all 8 admission/credentials citations land on the named constant or string; session.go:489 and :521 exact. Links resolve, style guide holds, doc guards green (TestEveryCurrentnessCodeIsDocumented, TestInputCausesAreDistinctAndDocumented, 0.583s). Open for the orchestrator: the .gitignore # Curator block is a live repo decision riding in a docs delta (raised rev1, rev2, still unresolved). Evidence: TASK-260827-1nauj3_review-verdict-rev3.md
agent completed: [reviewer] reviewer (claude) (exit=0)
spawn run completed: claude (run=RUN-260828-e0e8b1, pid=99451, exit=0)
Accepted per rev3 verdict; done set by orchestrator per CR contract.

## Precondition Resources
- [TASK-260827-1nauj3_cocoaskills-prose-style.md](file://TASK-260827-1nauj3/TASK-260827-1nauj3_cocoaskills-prose-style.md) — Source style guide to port/apply (English rules and blacklist)
- [TASK-260827-1nauj3_docs-refresh-spec.md](file://TASK-260827-1nauj3/TASK-260827-1nauj3_docs-refresh-spec.md) — Curator docs refresh spec
- [TASK-260827-1nauj3_tooling-note.md](file://TASK-260827-1nauj3/TASK-260827-1nauj3_tooling-note.md) — Shell-only edits, quoted heredocs, grep verification, literal outputs
- [TASK-260827-1nauj3_reviewer-note.md](file://TASK-260827-1nauj3/TASK-260827-1nauj3_reviewer-note.md) — Single-pass review, immediate verdict, no monitors
- [TASK-260827-1nauj3_finalize-note.md](file://TASK-260827-1nauj3/TASK-260827-1nauj3_finalize-note.md) — Docs written; verify, CR, handoff; no rewrites
- [TASK-260827-1nauj3_rework-instructions.md](file://TASK-260827-1nauj3/TASK-260827-1nauj3_rework-instructions.md) — Rev1 verdict: fix the six mismatched CLI entries against cmd/curator/main.go, apply remaining findings, literal outputs
- [TASK-260827-1nauj3_rework-instructions-2.md](file://TASK-260827-1nauj3/TASK-260827-1nauj3_rework-instructions-2.md) — Rev2 verdict: apply EVERY G-item exactly (G1 remove global add --project line 331, G2 add --project to curator add, G3/G4 add --all to install/upgrade, plus remaining G items); after each edit grep-verify; before handoff run ./bin/curator <cmd> -h for each touched command and paste outputs

## Outcome Resources
- [TASK-260827-1nauj3_spawn-log_-implementer--doc-writer--agy-_RUN-260828-d13a5f.log](file://TASK-260827-1nauj3/TASK-260827-1nauj3_spawn-log_-implementer--doc-writer--agy-_RUN-260828-d13a5f.log) — System spawn log captured by task-board
- [TASK-260827-1nauj3_results.md](file://TASK-260827-1nauj3/TASK-260827-1nauj3_results.md) — Task outcome summary for cli-and-troubleshooting documentation refresh
- [TASK-260827-1nauj3_spawn-log_-reviewer--reviewer--claude-_RUN-260828-73644d.log](file://TASK-260827-1nauj3/TASK-260827-1nauj3_spawn-log_-reviewer--reviewer--claude-_RUN-260828-73644d.log) — System spawn log captured by task-board
- [TASK-260827-1nauj3_spawn-log_-implementer--doc-writer--agy-_RUN-260828-8a8681.log](file://TASK-260827-1nauj3/TASK-260827-1nauj3_spawn-log_-implementer--doc-writer--agy-_RUN-260828-8a8681.log) — System spawn log captured by task-board
- [TASK-260827-1nauj3_change-request_rev1.patch](file://TASK-260827-1nauj3/TASK-260827-1nauj3_change-request_rev1.patch) — Change Request CR-TASK-260827-1nauj3-1 revision 1 candidate patch (repository_delta=present, 8 changed paths)
- [TASK-260827-1nauj3_spawn-log_-reviewer--reviewer--claude-_RUN-260828-6e9c4a.log](file://TASK-260827-1nauj3/TASK-260827-1nauj3_spawn-log_-reviewer--reviewer--claude-_RUN-260828-6e9c4a.log) — System spawn log captured by task-board
- [TASK-260827-1nauj3_review-verdict.md](file://TASK-260827-1nauj3/TASK-260827-1nauj3_review-verdict.md) — Reviewer verdict rev3: accepted; binary-probe evidence for all rev2 findings
- [TASK-260827-1nauj3_binary-help-dump.txt](file://TASK-260827-1nauj3/TASK-260827-1nauj3_binary-help-dump.txt) — Raw -h output of every curator command group from bin/curator rebuilt with make build in the story worktree
- [TASK-260827-1nauj3_spawn-log_-implementer--doc-writer--agy-_RUN-260828-acf233.log](file://TASK-260827-1nauj3/TASK-260827-1nauj3_spawn-log_-implementer--doc-writer--agy-_RUN-260828-acf233.log) — System spawn log captured by task-board
- [TASK-260827-1nauj3_rework_results.md](file://TASK-260827-1nauj3/TASK-260827-1nauj3_rework_results.md) — Rework verification and test evidence for TASK-260827-1nauj3
- [TASK-260827-1nauj3_change-request_rev2.patch](file://TASK-260827-1nauj3/TASK-260827-1nauj3_change-request_rev2.patch) — Change Request CR-TASK-260827-1nauj3-2 revision 2 candidate patch (repository_delta=present, 7 changed paths)
- [TASK-260827-1nauj3_spawn-log_-reviewer--reviewer--claude-_RUN-260828-a12dbd.log](file://TASK-260827-1nauj3/TASK-260827-1nauj3_spawn-log_-reviewer--reviewer--claude-_RUN-260828-a12dbd.log) — System spawn log captured by task-board
- [TASK-260827-1nauj3_review-verdict-rev2.md](file://TASK-260827-1nauj3/TASK-260827-1nauj3_review-verdict-rev2.md) — Rev2 reviewer verdict: changes requested. F2 unfixed plus four omitted real flags and a wrong README one-liner; F1/F3-F7 confirmed fixed.
- [TASK-260827-1nauj3_spawn-log_-implementer--doc-writer--agy-_RUN-260828-7e276b.log](file://TASK-260827-1nauj3/TASK-260827-1nauj3_spawn-log_-implementer--doc-writer--agy-_RUN-260828-7e276b.log) — System spawn log captured by task-board
- [TASK-260827-1nauj3_rework_rev2_results.md](file://TASK-260827-1nauj3/TASK-260827-1nauj3_rework_rev2_results.md) — Rev2 Rework Verification Report and Binary Help Dumps
- [TASK-260827-1nauj3_change-request_rev3.patch](file://TASK-260827-1nauj3/TASK-260827-1nauj3_change-request_rev3.patch) — Change Request CR-TASK-260827-1nauj3-3 revision 3 candidate patch (repository_delta=present, 7 changed paths)
- [TASK-260827-1nauj3_spawn-log_-reviewer--reviewer--claude-_RUN-260828-e0e8b1.log](file://TASK-260827-1nauj3/TASK-260827-1nauj3_spawn-log_-reviewer--reviewer--claude-_RUN-260828-e0e8b1.log) — System spawn log captured by task-board
- [TASK-260827-1nauj3_review-verdict-rev3.md](file://TASK-260827-1nauj3/TASK-260827-1nauj3_review-verdict-rev3.md) — Reviewer verdict rev3: accepted; binary-probe evidence closing all rev2 findings

## Created
2026-08-27T01:40:47Z

## Last Update
2026-08-28T02:36:08Z

## Assigned To
[reviewer] reviewer (claude)
