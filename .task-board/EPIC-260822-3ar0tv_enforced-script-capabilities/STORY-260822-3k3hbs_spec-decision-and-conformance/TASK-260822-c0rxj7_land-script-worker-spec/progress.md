## Status
blocked

## Review
light

## Task Class
metadata

## Estimate
estimated(fibonacci(2))

## Blocked By
- (none)

## Blocks
- TASK-260822-f4qv7w
- TASK-260822-1so0ym

## Checklist
- [x] Lockstep: profiles/manager.md active-process-count-limit Linux cell reads host-conditional: delegated cgroup v2 pids.max, matching core.md 78d544d — if TASK-260822-3fkfmf did not fix it, fix it at landing
- [x] core.md schema-8 deltas (a)-(d) from the TASK-260822-1mwy10 landing input: section 4 preamble v1-v8 incl. csk-skill-v8, Added-behavior row 8, version gates 2-8 plus schemas 1-7 MUST reject execution_policy/interpreter, section 10 marker schemas 1-4 and marker v4 paragraph - on no branch, confirmed still owed at spec/sw-core-prose 78d544d
- [x] Candidate branch combines spec/script-worker-v1-normative (dd9c9fc) and spec/module-roots-prose (bac193c); release/1.0.0-rc.8.json byte-identical to main
- [x] rc.9 candidate release metadata added and live-pin validation moved to it; both vector families regenerate twice byte-identical
- [x] Candidate identity recorded: full SHA, suite-manifest SHA-256, tree SHA-256, file count
- [ ] Specification CI green on the candidate head; candidate-conformance workflow_dispatch run with candidate_ref and manifest digest, evidence attached
- [ ] TASK-260822-f4qv7w and TASK-260822-1so0ym unblocked with the green candidate evidence routed to them
- [x] Code written per task description and AC
- [x] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [x] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant

## Notes
LANDING INPUT from TASK-260822-1mwy10 (schema branch spec/sw-schema, commit ebfed81). Two things the landing PR must carry beyond merging the three sibling branches:

1. CORE.MD DELTAS NOT ON ANY BRANCH. protocol/core.md on origin/spec/sw-core-prose (41cf556) adds section 4.1.1 but does not renumber the manifest schema series, so main would contradict schema 8. Required: (a) section 4 preamble — a manifest conforms to exactly one of agent-skill-v1 through agent-skill-v8 (and csk-skill-v8), currently says v7; (b) section 4 schema table — add a row for 8, opt-in script-worker-v1 execution policy with a closed interpreter identity; (c) section 4 version gates — schemas 2 through 8 reject unknown fields, plus a sentence that schemas 1 through 7 MUST reject execution_policy and interpreter at the top level and on every command; (d) section 10 install markers — managers supporting schema 8 MUST read marker schemas 1, 2, 3, and 4 and MUST write marker schema 4 for schema-8 installation mutations, plus a paragraph that marker v4 permits skill_schema_version 8 and otherwise carries marker v3 meaning unchanged.

2. COMPATIBILITY.md paragraph (already in this task AC): name manifest schema 8, install marker v4, and record that schema 8 is the single manifest bump for this revision, shared with decision 0009 first-party module roots, so module roots takes no sequential version.

Schema branch gates at ebfed81, real exit codes, each standalone: python3 tools/validate.py=0, unittest discover=0 (94 tests), go test ./tools/...=0, gofmt -l tools=0, generate-vectors + git diff --exit-code run twice=0 each. rc.5/rc.6/rc.7 release metadata unchanged; only the rc.8 candidate pin moves. Full evidence: TASK-260822-1mwy10_results.md sections 7 and 8.
Landing input from TASK-260822-3fkfmf (manager profile + SECURITY.md prose). Branch spec/sw-manager-security, head c2371d3 (signed, pushed), base origin/main=b92b105. Two commits: b9ca2ad adds the prose, c2371d3 reconciles it against protocol/core.md section 4.1.1 after TASK-260822-1f533i landed 41cf556. Files: profiles/manager.md (section 3 carve-out, new section 3.1, section 7 audit warning classes) and SECURITY.md (security-model qualifier, new Enforced script execution boundary section). No PR by design.

Three things this branch needs from you at merge time.
(1) TRAP: curator-spec has no .gitignore, and the worktree curator-spec/.temp/STORY-260822-3k3hbs/manager-security-worktree contains an untracked .temp with a python venv and gate logs. Never git add -A from that worktree. Same trap already recorded twice in LOGBOOK.md for the sibling branches.
(2) CROSS-DOCUMENT CONFLICT TO RESOLVE, NOT INHERIT: protocol/core.md 4.1.1 says "The complete diagnostic set of this policy is:" and lists four codes; profiles/manager.md 3.1 carries seven — those four plus script_execution_worker_identity_invalid, script_execution_worker_protocol_invalid, script_execution_package_influence_forbidden. That extension is the established shape (core.md 4.2.1 names three build_execution_* codes, manager.md 2.2.1 carries six including exactly these three analogues), but 4.2.1 never claims completeness and 4.1.1 does. Either soften 4.1.1 to "policy-level diagnostic set" or promote the three into it. Dropping them would leave the identity-verified worker failure modes undiagnosable.
(3) LIKELY-FALSE INVENTORY CELL IN ALL THREE COPIES: the script native-control inventory marks Linux active-process-count-limit as "available: RLIMIT_NPROC" in core.md 4.1.1, manager.md 3.1, and whatever f4qv7w generates. RLIMIT_NPROC caps processes per real UID across the whole session, so it is not a private aggregate domain — and the rc5 ledger already marks macOS unavailable: no-private-aggregate-domain for that same primitive. An available entry MUST report applied, so this cell manufactures exactly the false applied-claim decision 0006 forbids. Recommendation: change it to "host-conditional: delegated cgroup v2 pids.max" in all three places at once. 3fkfmf mirrored it verbatim rather than editing one of three copies unilaterally.

Also still open from the analysis review, for f4qv7w rather than for this branch: host-conditional controls route to platform-case class host-capability, but platform-cases.tsv defines that class as a runner-filesystem limitation, not a missing kernel feature. Needs a widened or new class decided in prose.

Evidence: TASK-260822-3fkfmf_results.md (board outcome resource) and LOGBOOK.md 2026-08-22 2110.
RE-SCOPED per TASK-260823-omp8zt impact-analysis (executive findings 1-2 and the 9-step landing order): do NOT open a vectors-only PR to main — pinned Implementations jobs are falsely green (Go pin consumes only schema-6 cases, Python pin ignores the schema-case index). Instead this task delivers steps 1-3: (1) create one immutable schema-8 candidate commit outside main on a candidate branch merging the reviewed script-worker story branch AND the module-roots prose/vectors work — schema 8 is one shared bump, both families land together — with protocol version 1.0.0-rc.9 metadata, vectors regenerated twice; (2) record candidate identity (full SHA, suite-manifest SHA-256, tree SHA-256, file count); (3) qualify through the existing candidate-conformance workflow_dispatch with candidate_ref and manifest digest, SPEC_PIN unchanged, CI_REQUIRE_FULL_ROOT=1. The actual spec landing PR (pins advanced atomically, steps 6-7) happens ONLY after both implementations qualify against the candidate — that is a later task, do not merge to main here.
Inputs ready when TASK-260822-1so0ym lands: script-worker family committed+pushed on spec/script-worker-v1-normative at dd9c9fc (vectors, minus the excluded rc.8 rewrite; branch CI red ONLY on the rc.8 live-pin invariant — see TASK-260822-f4qv7w delivery-rework packet with the three rejected forced fits), module-roots family on spec/module-roots-prose. Your rc.9 migration resolves the rc.8 pin invariant: new release/1.0.0-rc.9.json candidate identity, validator/generator/tests move to the rc.9 pin, rc.8 file stays byte-immutable. After the candidate is green, unblock TASK-260822-f4qv7w back to review with the green run as evidence.
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=codex; schema=1; producer=v1.6.1; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [implementer] developer (codex) (run=RUN-260823-320616, max_parallel=20)
spawn run started: [implementer] developer (codex) (run=RUN-260823-320616)
BLOCKED after producing immutable Schema 8 rc.9 candidate 859727b103ed175ff214cbb64641f4686d8c6a68 (manifest sha256:782d6868f6d9725f7bf38d3fb1944f307f2d5d9c060b8816b0f55a5c2e97f11f; tree sha256:f88a76263040c90470be8f18c110927267ac5862f7e25b3b161bcd5ef97319f3; 692 files). Spec CI 32633567102 is green on the exact SHA. Candidate jobs in curator run 32633572039 are definitively red: Ubuntu/macOS expose the unlanded multi-project dry-run binding fix (open PR #14, d345420); Windows identity exposes shasum path-escaping that prefixes the digest with backslash. Rejected forced fits: weakening the gate, supplying a false digest, or cherry-picking unqualified implementation code solely to manufacture green. Required external changes: land/qualify the Curator binding implementation, normalize Windows candidate hashing, add explicit Schema 8 consumption coverage in both implementations, then rerun the same candidate SHA/digest. SPEC_PIN remains unchanged; branch/worktree preserved. Unblock evidence recorded on both implementation stories and both vector tasks; board dependency edges prevented the two vector tasks from transitioning out of blocked while this task remains active. Outcome: TASK-260822-c0rxj7_results.md.
agent completed: [implementer] developer (codex) (exit=0)
spawn run completed: codex (run=RUN-260823-320616, pid=77083, exit=0)
RC.9 rerun 32638424105 is terminal RED, not release-unblocking evidence. Exact candidate SHA/digest unchanged. Candidate macOS passed; Ubuntu and Windows failed on newly exposed implementation gaps. Detailed matrix and stop-line packet attached as TASK-260822-c0rxj7_rc9-rerun-32638424105.md. Do not release blocked landing statuses from this run.
SUPERSEDING immutable rc.9 candidate published after Windows environment-vector reconciliation: edd07210d4f3db34fd60238cb14b90f837de03cb; manifest sha256:803918bf8672f76cf990985e51db213b826674cd5bb54fbf47731b8404b44403; tree sha256:9d5a10b6ef1bd867f4d055d830d10a240620d759ff245fed9ccdb40b888ab769; 692 files. Old 859727b/782d686 identity remains immutable in history. Double regeneration inventories updated and cmp exited 0. Spec CI 32642316308 and Curator candidate run 32642340559 dispatched; qualification pending terminal results.

## Precondition Resources
(none)

## Outcome Resources
- [TASK-260822-c0rxj7_spawn-log_-implementer--developer--codex-_RUN-260823-320616.log](file://TASK-260822-c0rxj7/TASK-260822-c0rxj7_spawn-log_-implementer--developer--codex-_RUN-260823-320616.log) — System spawn log captured by task-board
- [TASK-260822-c0rxj7_results.md](file://TASK-260822-c0rxj7/TASK-260822-c0rxj7_results.md) — Immutable Schema 8 rc.9 candidate identity, validation, CI, and blocker packet
- [TASK-260822-c0rxj7_candidate-suite-identity.txt](file://TASK-260822-c0rxj7/TASK-260822-c0rxj7_candidate-suite-identity.txt) — Superseding candidate full SHA, manifest digest, tree digest, and file count
- [TASK-260822-c0rxj7_regeneration-pass1.sha256](file://TASK-260822-c0rxj7/TASK-260822-c0rxj7_regeneration-pass1.sha256) — Superseding candidate first deterministic regeneration checksum inventory
- [TASK-260822-c0rxj7_regeneration-pass2.sha256](file://TASK-260822-c0rxj7/TASK-260822-c0rxj7_regeneration-pass2.sha256) — Superseding candidate second deterministic regeneration checksum inventory
- [TASK-260822-c0rxj7_candidate-ubuntu.log](file://TASK-260822-c0rxj7/TASK-260822-c0rxj7_candidate-ubuntu.log) — Candidate-conformance Ubuntu failed-job log
- [TASK-260822-c0rxj7_candidate-windows.log](file://TASK-260822-c0rxj7/TASK-260822-c0rxj7_candidate-windows.log) — Candidate-conformance Windows failed-job log
- [TASK-260822-c0rxj7_rc9-rerun-32638424105.md](file://TASK-260822-c0rxj7/TASK-260822-c0rxj7_rc9-rerun-32638424105.md) — Red rc.9 rerun matrix and newly exposed Curator conformance gaps; not release-unblocking evidence

## Created
2026-08-22T16:00:35Z

## Last Update
2026-08-23T13:27:06Z

## Assigned To
[implementer] developer (codex)
