## Status
done

## Review
required

## Task Class
code

## Estimate
estimated(fibonacci(5))

## Blocked By
- (none)

## Blocks
- TASK-260823-1wvgw8

## Checklist
- [x] Schema-case families for agent-skill-v8 consumed by unit tests
- [x] Code written per task description and AC
- [x] Relevant tests written for new or changed behavior and passing
- [x] Lint clean
- [x] Relevant build/validation commands run after changes and build not broken
- [x] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [x] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
- [x] Implementation matches AC
- [x] Solution fits project architecture
- [x] Tests green
- [x] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches

## Notes
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [implementer] developer (claude) (run=RUN-260823-2269a2, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260823-2269a2)
Implementation ready for review. PR https://github.com/relux-works/curator/pull/33 from origin/main@e17b0f1.

Scope delivered:
- internal/moduleroots (new): both halves of Protocol Core 4.2.3, split on the fixed go list per its failure boundary. ValidateDeclaration (portability, real link-free directory strictly inside the snapshot, go.mod directly inside, uniqueness, pairwise disjointness from other declarations/build roots/runtime roots under exact AND platform-folded comparison) runs before go list; EffectiveReplaceSet + ValidateBijection (vendor/modules.txt bytes only, selection-annotation reconciliation, no versioned sides, no module-to-module redirects, one-to-one in both directions) run after it. All five closed phase:preflight diagnostics from profiles/manager.md.
- internal/skillspec: schema 8 admitted. modules only on local go-v1 build commands; execution_policy + interpreter only on script commands, bound both ways, closed constant and closed interpreter set, absence the single spelling of declared-only. Declared-only behaviour: parsed and carried, no containment claimed (decision-0008 worker containment is out of scope). Schemas 2-7 reject all three as unknown; schema 1 keeps extension tolerance and ignores them rather than reading an enforcement claim.
- internal/marker: install marker v4 (= v3 with schema_version 4 / skill_schema_version 8). Not optional - without it marker.Write fails closed for every schema-8 install, so admitting schema 8 alone would land an unusable half-feature. internal/install now bands the explicit receipt-version/execution-policy build record at schema >= 7 instead of == 7.
- .github/ci/root-artifacts.tsv declares the three newly-read artefacts.

Suite consumption: agent-skill-v8 and csk-skill-v8 (67 cases each), install-marker-v4 (27), and all 10 vectors/module-roots.json cases. TestModuleRootVectors asserts each vector against the correct half of the boundary, so it proves the ordering, not just the diagnostic.

Local evidence, each command standalone with its real exit code: gofmt -l cmd internal empty; go build ./... 0; go vet ./... 0; golangci-lint run (v2.12.2) 0 / 0 issues; go test ./internal/... 0; go test ./cmd/... 0 (cmd/curator 276.287s); gate-selftest.sh 0 (81 passed, 0 failed); CURATOR_CONFORMANCE_ROOT=<6001dc3>/conformance/v1 CI_REQUIRE_FULL_ROOT=1 test-gate.sh 0 (served=42 deferred=0 excluded=0, platform-case gate ok, 10 pre-existing skips, no new skip class).

Two pre-existing findings recorded in LOGBOOK and the results artifact, deliberately NOT fixed here: (1) marker v3/v4 external cross-field rules (declared-vs-effective identity, identity kind vs substitution type, revision width vs object format) are unimplemented - five install-marker-v3 AND v4 cases are accepted; encoded as markerV4CasesThisReaderDoesNotModel and asserted in both directions so the allowance cannot go stale. (2) cmd/curator/builds.go:365 reports every marker above v2 as unsupported, contradicting its own comment; marker v4 inherits it, schema 8 adds no new class of breakage.

Not in scope, handed to TASK-260823-1wvgw8: the go-v1 driver still rejects every Module.Replace, so a declared module root does not build yet. EffectiveReplaceSet/ValidateBijection are ready to wire in at internal/godriver/graph.go:306.
agent completed: [implementer] developer (claude) (exit=124)
spawn run completed: claude (run=RUN-260823-2269a2, pid=60942, exit=124)
Landed: PR 33 (Admit schema-8 module roots and the script execution surface) merged to main — every lane verified green pre-merge. Handoff map for TASK-260823-1wvgw8 is in the results resource: EffectiveReplaceSet and ValidateBijection ready to wire; driver must read build-root vendor/modules.txt, replace the unconditional Module.Replace rejection at graph.go:306 with admission of exactly the bijected set, extend the scan surface to declared directories and their vendor copies, and withhold the audited-vendor allowance from any module carrying a replacement.
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [reviewer] reviewer (claude) (run=RUN-260823-8f0f4e, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260823-8f0f4e)
Review verdict RUN-260823-8f0f4e: CHANGES REQUESTED -> to-dev. Full packet: TASK-260823-1vleh5_review-verdict.md.

BLOCKING: the schema-8 script surface that shipped with this task puts curator in the exact state profiles/manager.md forbids by name (candidate 6001dc3, below the script diagnostics table): "A manager that does not implement this policy MUST reject such a command with script_execution_policy_unsupported. It MUST NOT install the command declared-only, downgrade it, or ignore the field, because the resulting shim would run package code the manifest says is contained." parse.go:305 parses execution_policy/interpreter and stores them on Command; NOTHING reads them. grep for .Interpreter in non-test cmd+internal returns nothing; grep for script_execution returns nothing, so the closed diagnostic does not exist in this codebase; install/targets.go:266 selects script shims on command.Type alone. Before this commit SupportedSchemaVersions stopped at 7 so an enforced command could never be installed; after it, valid-script-worker-enforced.json installs an ordinary shim and runs the script with the caller full ambient authority. Fail-OPEN regression introduced by this change. Contrast the module-roots half, which fails CLOSED: godriver/graph.go:306 still rejects every Module.Replace, so the TASK-260823-1wvgw8 handoff is safe. STORY-260822-2h0v9j (the worker) is in backlog and no board item owns the interim behaviour.

REQUIRED FIX: keep skillspec.Load accepting the manifest (the suite marks valid-script-worker-enforced.json valid and the task says parsing must not reject valid schema-8 manifests). Reject one layer down, at command admission / shim installation, any command carrying ExecutionPolicy == ScriptExecutionPolicy with the closed diagnostic script_execution_policy_unsupported (state unsupported, severity error), until the worker lands. Tests: pin the refusal and its code, and pin that a schema-8 DECLARED-ONLY script command still installs unchanged.

NON-BLOCKING: schema-cases/agent-skill-v8/invalid-script-worker-missing-path.json is rejected by "runtime_roots[0]: runtime root does not exist: scripts", not by the missing-path rule it exists to test — materialize declared runtime_roots in materializeManifestFixture. The rule itself is implemented and unit-covered. All other 32 module-roots/script-worker cases verified rejected for their own reason.

ACCEPTED AND INDEPENDENTLY VERIFIED: the whole module-roots half. Failure boundary asserted per-half by TestModuleRootVectors, not just the diagnostic. The load-bearing Go premise was reproduced empirically, not taken from prose: a real go mod vendor (go1.25.5) over replace example.com/board => ../../pkg/board writes BOTH "# example.com/board v0.0.0 => ../../pkg/board" and "# example.com/board => ../../pkg/board", so the selection-annotation reconciliation is sound, and matching on the whole annotation rather than the left token is the right trap-avoidance. Containment (Lstat per component, real regular go.mod, uniqueness, disjointness under exact AND NFD-fold-NFD, component-wise contains), scoping (modules only on local go-v1, unknown on 2-7, ignored under schema-1 extension tolerance), marker v4 (not optional — Write fails closed for every schema-8 install without it), and root-artifacts.tsv are all correct.

Pre-existing findings self-reported by the implementer are real and correctly scoped out. I confirmed the marker one independently: the five install-marker-v4 cases in markerV4CasesThisReaderDoesNotModel exist in install-marker-v3 with byte-identical bodies once the two version fields are removed, so v4 inherits the v3 gap rather than widening it. NEITHER is covered by TASK-260823-2u5xov (suite-consumption + qualification) — the marker cross-field gap needs its own board item.

REVIEWER GATES, worktree at 62d578c, each standalone: gofmt -l cmd internal empty; go build ./... 0; go vet ./... 0; golangci-lint 2.12.2 0 (0 issues); go test ./internal/... with the candidate root 0; gate-selftest.sh 0 (81 passed, 0 failed); suite-plan.sh vs 6001dc3 0 (served=42 deferred=0 excluded=0); CURATOR_CONFORMANCE_ROOT=6001dc3 CI_REQUIRE_FULL_ROOT=1 test-gate.sh 0 (go test exit=0, platform-case gate exit=0, all 42 served packages pass, 10 pre-existing skips, no new skip class). PR 33 CI: all twelve required lanes green. The evidence is real and the gates are honest — no gate can catch a diagnostic the codebase never implements.
agent completed: [reviewer] reviewer (claude) (exit=0)
spawn run completed: claude (run=RUN-260823-8f0f4e, pid=28244, exit=0)
REWORK per the review verdict blocking finding: parsing execution_policy without acting on it violates profiles/manager.md (candidate 6001dc3 lines 908-911) — a manager that does not implement script-worker-v1 MUST reject such a command with script_execution_policy_unsupported, never install it declared-only or downgraded. Until the containment story (STORY-260822-2h0v9j) lands, curator honestly does not implement the policy: add the fail-closed rejection at install planning for any script command carrying execution_policy=script-worker-v1, with the exact diagnostic, plus tests (schema-8 case with enforced command must fail installation with that diagnostic; plain schema-8 manifests unaffected). Also fix the non-blocking finding: the schema case rejected for the wrong reason — align the diagnostic. Land via PR, every lane green pre-merge, then handoff to-review. Executor: claude only.
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [implementer] developer (claude) (run=RUN-260823-bb1fae, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260823-bb1fae)
agent completed: [implementer] developer (claude) (exit=124)
spawn run completed: claude (run=RUN-260823-bb1fae, pid=69805, exit=124)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [reviewer] reviewer (claude) (run=RUN-260823-f6b96b, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260823-f6b96b)
Rework landed: PR 34 (Refuse an enforced script command instead of installing it uncontained) merged as 77aafa0, all lanes green pre-merge — fail-closed script_execution_policy_unsupported until STORY-260822-2h0v9j implements containment; wrong-reason schema-case diagnostic aligned. Reviewer cycle 2 running as RUN-260823-f6b96b.
agent completed: [reviewer] reviewer (claude) (exit=124)
spawn run completed: claude (run=RUN-260823-f6b96b, pid=71395, exit=124)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [reviewer] reviewer (claude) (run=RUN-260823-2c49ba, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260823-2c49ba)
agent completed: [reviewer] reviewer (claude) (exit=124)
spawn run completed: claude (run=RUN-260823-2c49ba, pid=26635, exit=124)
REVIEWER SCOPE NOTE (two reviewer runs timed out re-verifying too broadly): review ONLY the cycle-2 delta — PR 34 (77aafa0): fail-closed script_execution_policy_unsupported rejection at install planning + the realigned schema-case diagnostic. Evidence base: PR 33 delta was already reviewed in cycle 1 (verdict resource); PR 34 lanes were all green pre-merge (checks tab). Run TARGETED go tests only (internal/skillspec, the install-planning package tests touching the rejection path) — do NOT run the full suite or the full candidate gate locally; cite the PR CI lanes for the rest. Deliver the verdict within the budget.
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [reviewer] reviewer (claude) (run=RUN-260823-047c36, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260823-047c36)
REVIEW VERDICT RUN-260823-047c36: ACCEPTED (done). Reviewed origin/main@77aafa0 = merge of PR #34, impl commit a74f1f5 (diff a74f1f5..77aafa0 empty), against candidate 6001dc3. The RUN-260823-8f0f4e blocking finding is closed at the layer that verdict named: skillspec.Load still accepts the manifest (all 66 agent-skill-v8 cases resolve correctly), and admission moved into internal/scriptpolicy, which owns script_execution_policy_unsupported with the profile-fixed unsupported/error pair (profiles/manager.md:901 — both values checked against the table). Two consumers: skillcheck (reaches project install.go:387, global.go:139 and curator skill check, error severity aborts before any staging) and install.activeScriptCommands (the shim writer itself, both stageRuntimeAndShims call sites propagate). I re-ran the mutation matrix myself: both active=pass, skillcheck disabled=pass, shim writer disabled=pass, BOTH disabled=FAIL with Status:ok and the launcher published — the reviewed regression reproduced exactly, so neither layer is decoration. No remaining path to an uncontained enforced shim: stageRuntimeAndShims is the only shim producer with exactly two callers both behind validateNodes; globalbins only mirrors already-published canonical shims; ExecutionPolicy is written only by parse.go:314 for script commands at schema>=8, so Enforced() cannot catch a go-v1 build command (compiled manager-worker-v1 lives on a different type). Determinism holds on both layers (Admit sorts; ActiveCommandNames is already sorted). Curator emits conformance_claim none, so nothing claims what this refusal denies. Previous non-blocking finding also fixed and independently verified: I instrumented the whole agent-skill-v8 family and printed the real rejection reason for all 66 cases — every invalid case now fails for its own rule, including invalid-script-worker-missing-path. GATES (mine, standalone, worktree at 77aafa0): gofmt clean, go build 0, go vet 0, golangci-lint v2.12.2 0 issues, gate-selftest 81 passed 0 failed, test-gate.sh against candidate 6001dc3 with CI_REQUIRE_FULL_ROOT=1 exit 0 — served=43 deferred=0 excluded=0 (43 not 42: internal/scriptpolicy is served, not deferred), go test exit=0, platform-case gate exit=0, 10 skips all pre-existing classes. REMOTE: PR #34 11/11 default lanes pass (Race is ubuntu+macos by design per ci.yml:158, so 11 is the complete set); dispatched run 32668086905 on a74f1f5 with CANDIDATE_REF=6001dc33... is 14/14 including Candidate suite on ubuntu, macos AND windows — ref read out of the job log, not taken from the delivery note. CARRIED FORWARD, none blocking, none owned: (1) the two REQUIRED audit warning classes (profiles/manager.md:1008) script-command-declared-only and script-command-unfiltered-declared-network appear nowhere in the tree — NOT a regression here (the first applies to schema<=7 too, the second is unreachable while every enforced command is refused) but no board item owns it and TASK-260823-2u5xov does not cover it; (2) the marker v3/v4 cross-field gap still has no board item, as RUN-260823-8f0f4e required; (3) the two guards differ in scope on purpose — skillcheck refuses a DECLARED enforced command, the shim writer only an ACTIVE one, so a declared-but-inactive enforced command fails the whole install; fail-closed and cannot regress anything since schema 8 is new; (4) Admit reports only the first enforced command per skill (nit, buys the tested determinism); (5) pre-existing repo-wide doc staleness — README.md:36 still says schemas 1 through 5 while SupportedSchemaVersions spans 1-8, and CHANGELOG Unreleased is empty after 68 commits. STOP-THE-LINE: correctly judged not applicable — the task text and the profile are reconcilable (parse the field, refuse the install), they were just reconciled wrongly the first time, and the rework says so explicitly. Verdict artifact: TASK-260823-1vleh5_review-verdict-rework.md. Reviewer-archetype run, no commit_ack supplied; a74f1f5 is already on main as 77aafa0, so nothing remains to commit.
agent completed: [reviewer] reviewer (claude) (exit=0)
spawn run completed: claude (run=RUN-260823-047c36, pid=95689, exit=0)

## Precondition Resources
(none)

## Outcome Resources
- [TASK-260823-1vleh5_spawn-log_-implementer--developer--claude-_RUN-260823-2269a2.log](file://TASK-260823-1vleh5/TASK-260823-1vleh5_spawn-log_-implementer--developer--claude-_RUN-260823-2269a2.log) — System spawn log captured by task-board
- [TASK-260823-1vleh5_results.md](file://TASK-260823-1vleh5/TASK-260823-1vleh5_results.md) — Schema-8 parse surface, module-root validation, marker v4, gate evidence, two pre-existing findings
- [TASK-260823-1vleh5_spawn-log_-reviewer--reviewer--claude-_RUN-260823-8f0f4e.log](file://TASK-260823-1vleh5/TASK-260823-1vleh5_spawn-log_-reviewer--reviewer--claude-_RUN-260823-8f0f4e.log) — System spawn log captured by task-board
- [TASK-260823-1vleh5_review-verdict.md](file://TASK-260823-1vleh5/TASK-260823-1vleh5_review-verdict.md) — Reviewer verdict RUN-260823-8f0f4e: changes requested — enforced schema-8 script commands install and run uncontained; module-roots half verified and accepted
- [TASK-260823-1vleh5_spawn-log_-implementer--developer--claude-_RUN-260823-bb1fae.log](file://TASK-260823-1vleh5/TASK-260823-1vleh5_spawn-log_-implementer--developer--claude-_RUN-260823-bb1fae.log) — System spawn log captured by task-board
- [TASK-260823-1vleh5_rework-results.md](file://TASK-260823-1vleh5/TASK-260823-1vleh5_rework-results.md) — Rework delivery: enforced script commands refused at admission, not installed uncontained; mutation-checked guards; gate evidence
- [TASK-260823-1vleh5_spawn-log_-reviewer--reviewer--claude-_RUN-260823-f6b96b.log](file://TASK-260823-1vleh5/TASK-260823-1vleh5_spawn-log_-reviewer--reviewer--claude-_RUN-260823-f6b96b.log) — System spawn log captured by task-board
- [TASK-260823-1vleh5_spawn-log_-reviewer--reviewer--claude-_RUN-260823-2c49ba.log](file://TASK-260823-1vleh5/TASK-260823-1vleh5_spawn-log_-reviewer--reviewer--claude-_RUN-260823-2c49ba.log) — System spawn log captured by task-board
- [TASK-260823-1vleh5_spawn-log_-reviewer--reviewer--claude-_RUN-260823-047c36.log](file://TASK-260823-1vleh5/TASK-260823-1vleh5_spawn-log_-reviewer--reviewer--claude-_RUN-260823-047c36.log) — System spawn log captured by task-board
- [TASK-260823-1vleh5_review-verdict-rework.md](file://TASK-260823-1vleh5/TASK-260823-1vleh5_review-verdict-rework.md) — Review verdict RUN-260823-047c36: ACCEPTED — enforced script commands refused at both admission layers, mutation matrix and schema-case reasons independently reproduced, candidate suite green on all three platforms

## Created
2026-08-23T19:42:43Z

## Last Update
2026-08-23T23:35:38Z

## Assigned To
[reviewer] reviewer (claude)
