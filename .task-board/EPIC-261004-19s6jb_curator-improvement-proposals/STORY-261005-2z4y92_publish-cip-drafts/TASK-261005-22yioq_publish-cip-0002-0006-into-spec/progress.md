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
- (none)

## Checklist
- [x] Five cips/CIP-000N-<slug>.md files (0002-0006) in TEMPLATE.md shape with Status Draft
- [x] cips/README.md index lists CIP-0002..0006
- [x] Evidence linked to curator .research/ files by repo-relative path and commit, not copied
- [x] No normative protocol/schema, CHANGELOG or LOGBOOK edits; no personal paths, hosts or employer names
- [x] Applicable spec docs checks (links/format) run with exit codes recorded
- [x] Code written per task description and AC
- [x] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [x] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
- [x] Implementation matches AC
- [x] Solution fits project architecture
- [x] Tests green
- [x] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches

## Notes
spawn selection rationale tuple: {"role":"developer","pair":"muse-spark-1.3-contributor/max","text":"tb-R164 developer muse-spark-1.3-contributor max; tb-R136 health-gated queue"}
spawn selection rationale for muse-spark-1.3-contributor/max: tb-R164 developer muse-spark-1.3-contributor max; tb-R136 health-gated queue
spawn agent resolution: Agent selection: muse via explicit_override (preferred_agentic_system: mixed[claude,codex,muse], config: spawn.preferred_agentic_system)
spawn queued: [implementer] developer (muse) (run=RUN-261005-909be2, max_parallel=8)
spawn run started: [implementer] developer (muse) (run=RUN-261005-909be2)
5 Draft CIPs filed in cips/ from curator main @ 3d9aa987 research; README index linked. Gates: validate.py 0, unittest 672/672 0, skillfile 0, whitespace/gofmt clean; go test 1 pre-existing rc.14-pin failure (verified on pristine tree, side effects reverted). Finding recorded in outcome artifact, not repo LOGBOOK (brief forbids LOGBOOK edits).
spawn run child final message (run=RUN-261005-909be2, tools=87 patches=12 failed=2):
Five Draft CIPs are filed in `cips/` and ready for review.

**What changed** (worktree only, uncommitted): five new files — [CIP-0002](cips/CIP-0002-project-context-in-managed-launches.md), [CIP-0003](cips/CIP-0003-claude-managed-home-credential-modes.md), [CIP-0004](cips/CIP-0004-shell-hook-without-sourcing-and-path-append.md), [CIP-0005](cips/CIP-0005-audit-backends-and-cli-secret-transport.md), [CIP-0006](cips/CIP-0006-legacy-provider-settings-and-mcp-opt-outs.md) — plus the [README index](cips/README.md:93) with all five rows converted from "In preparation" to linked Draft. Each file keeps the research substance, decisions-needed lists, alternatives, and open questions; evidence is cited by `relux-works/curator` `.research/` path at `main` @ `3d9aa987` (verified equal to the GitHub remote tip), not copied. Titles for 0004/0006 were aligned to the reserved index titles; no normative, schema, CHANGELOG, or LOGBOOK edits; personal-data grep clean.

**Gates (real exit codes):** `validate.py` 0 (73 schemas, 1294 vectors, local links over the new files); `unittest discover` 0, 672/672; skillfile-independence 0; `git diff --check`, trailing-whitespace, and `gofmt` clean. `go test ./tools/...` exits 1 on one pre-existing failure (`TestRC14ReleaseMetadataPinsCandidateAndPreservesSourceBaseline`) reproduced identically with my changes stashed — the run regenerates `environments-read-failure.json` and breaks its own rc.14 pin; those side-effect edits were reverted. Lychee isn't installed (CI-only); external hosts were curl-checked instead (all 200; rfc-editor 302 redirect).

**Notes:** an early full-suite run showed 2 transient failures caused by this session's own concurrent tree mutation; the undisturbed rerun is fully green. The Go finding is recorded in the board outcome artifact `TASK-261005-22yioq_results.md`, not the repo LOGBOOK, per the brief's no-LOGBOOK rule. Checklist 8/8, handed off to review.
agent completed: [implementer] developer (muse) (exit=0)
spawn run completed: muse (run=RUN-261005-909be2, pid=28478, exit=0)
spawn selection rationale tuple: {"role":"reviewer","pair":"gpt-6.1-sol/high","text":"R187 cross-provider reviewer; sol high (only sol pair admitted on spec)"}
spawn selection rationale for gpt-6.1-sol/high: R187 cross-provider reviewer; sol high (only sol pair admitted on spec)
spawn agent resolution: Agent selection: codex via explicit_override (preferred_agentic_system: mixed[claude,codex,muse], config: spawn.preferred_agentic_system)
spawn queued: [reviewer] reviewer (codex) (run=RUN-261005-5cea72, max_parallel=8)
spawn run started: [reviewer] reviewer (codex) (run=RUN-261005-5cea72)
Reviewer revision 1: changes requested. (1) All five CIPs replace research evidence Markdown links with code spans: restore commit-pinned source and companion links (0/10 currently linked; 10/10 proposed full-commit blob URLs verified HTTP 200). (2) Escape four inline pipes in CIP-0002 lines 77, 80, 81, 82: GFM rendering drops the defaults/approval/memory text. Source decisions and 31/31 questions preserved; scope 6/6 files correct. Reviewer gates: validate.py 0, 672/672 Python tests 0, Skillfile independence 0, diff whitespace 0, existing HTTP URLs 14/14. No Go tests or repository edits. Full swept-surface evidence: TASK-261005-22yioq_review-verdict-rev1.md. Findings recorded on board rather than LOGBOOK per task-specific prohibition. Route to-dev for bounded docs rework and another reviewer cycle.
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-261005-5cea72, pid=51045, exit=0)
loop-detector rev1: S2/S3/S5 not evaluable — runtime-recorded verdict carries no stamped findings array
spawn selection rationale for muse-spark-1.3-contributor/max: tb-R164 developer muse-spark-1.3-contributor max; tb-R136 health-gated queue
spawn agent resolution: Agent selection: muse via explicit_override (preferred_agentic_system: mixed[claude,codex,muse], config: spawn.preferred_agentic_system)
spawn queued: [implementer] developer (muse) (run=RUN-261005-05ea24, max_parallel=8)
spawn run started: [implementer] developer (muse) (run=RUN-261005-05ea24)
spawn run child final message (run=RUN-261005-05ea24, tools=56 patches=11 failed=1):
Rework 1 for TASK-261005-22yioq (publish CIP drafts 0002–0006) is ready for review.

Both reviewer findings answered, everything else verified byte-identical against the rev1 candidate tree:
- Research links: all five CIPs now carry two commit-pinned Markdown links each (source draft + `_evidence.md` companion) to curator @ `fae2ff9c` — 10/10 links present, 10/10 HTTP 200, files confirmed byte-identical to the prior pin.
- Table pipes: the four CIP-0002 operator-control enums escaped (`off \| admitted`, etc.); GFM render confirms 7/7 two-cell rows with all four semantics intact.

Evidence attached: `TASK-261005-22yioq_rework1-validation.md` with all exit codes (`validate.py` 0, regression probe 0, render check 0, mutants 0, `git diff --check` 0, privacy grep 0 matches), plus the named regression test (`TASK-261005-22yioq_table-regression.py`, 29 tables/193 rows clean) and render probe. Narrowing mutant flags exactly the single unescaped row; delete-variant flags exactly the four original rows. Go tests not run per reviewer instruction; the Python unit suite not rerun since no Python code changed.
agent completed: [implementer] developer (muse) (exit=0)
spawn run completed: muse (run=RUN-261005-05ea24, pid=56974, exit=0)
spawn selection rationale for gpt-6.1-sol/high: R187 cross-provider reviewer; sol high (only sol pair admitted on spec)
spawn agent resolution: Agent selection: codex via explicit_override (preferred_agentic_system: mixed[claude,codex,muse], config: spawn.preferred_agentic_system)
spawn queued: [reviewer] reviewer (codex) (run=RUN-261005-c5b436, max_parallel=8)
spawn run started: [reviewer] reviewer (codex) (run=RUN-261005-c5b436)
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-261005-c5b436, pid=63589, exit=0)
spawn selection rationale tuple: {"role":"developer","pair":"gpt-6-astra/low","text":"tb-R164 developer gpt-6-astra low; tb-R136 health-gated queue"}
spawn selection rationale for gpt-6-astra/low: tb-R164 developer gpt-6-astra low; tb-R136 health-gated queue
spawn agent resolution: Agent selection: codex via explicit_override (preferred_agentic_system: mixed[claude,codex,muse], config: spawn.preferred_agentic_system)
spawn queued: [implementer] developer (codex) (run=RUN-261005-6bf2d3, max_parallel=8)
spawn run started: [implementer] developer (codex) (run=RUN-261005-6bf2d3)
agent completed: [implementer] developer (codex) (exit=0)
spawn run completed: codex (run=RUN-261005-6bf2d3, pid=66352, exit=0)
spawn run RUN-261005-6bf2d3 failed; operator action required; failure: integration_binding_state_invalid: Change Request CR-TASK-261005-22yioq-2 revision 2 is integrated, want accepted or checkpointed

## Precondition Resources
- [cip-publish-brief.md](file://TASK-261005-22yioq/cip-publish-brief.md)
- [cip-publish-review-note.md](file://TASK-261005-22yioq/cip-publish-review-note.md)
- [cip-publish-rework.md](file://TASK-261005-22yioq/cip-publish-rework.md)
- [22yioq-complete.md](file://TASK-261005-22yioq/22yioq-complete.md)

## Outcome Resources
- [TASK-261005-22yioq_spawn-log_-implementer--developer--muse-_RUN-261005-909be2.log](file://TASK-261005-22yioq/TASK-261005-22yioq_spawn-log_-implementer--developer--muse-_RUN-261005-909be2.log) — System spawn log captured by task-board
- [TASK-261005-22yioq_results.md](file://TASK-261005-22yioq/TASK-261005-22yioq_results.md) — CIP 0002-0006 publish: file inventory, source pins, gate exit codes, findings
- [TASK-261005-22yioq_change-request_rev1.patch](file://TASK-261005-22yioq/TASK-261005-22yioq_change-request_rev1.patch) — Change Request CR-TASK-261005-22yioq-1 revision 1 candidate patch (repository_delta=present, 6 changed paths)
- [TASK-261005-22yioq_change-request_rev1-validation.log](file://TASK-261005-22yioq/TASK-261005-22yioq_change-request_rev1-validation.log) — Change Request CR-TASK-261005-22yioq-1 revision 1 bounded validation log
- [TASK-261005-22yioq_spawn-log_-reviewer--reviewer--codex-_RUN-261005-5cea72.log](file://TASK-261005-22yioq/TASK-261005-22yioq_spawn-log_-reviewer--reviewer--codex-_RUN-261005-5cea72.log) — System spawn log captured by task-board
- [TASK-261005-22yioq_review-verdict-rev1.md](file://TASK-261005-22yioq/TASK-261005-22yioq_review-verdict-rev1.md) — Revision 1 changes requested: pinned research links and four GFM table rows; full review coverage and measured gates
- [TASK-261005-22yioq_spawn-log_-implementer--developer--muse-_RUN-261005-05ea24.log](file://TASK-261005-22yioq/TASK-261005-22yioq_spawn-log_-implementer--developer--muse-_RUN-261005-05ea24.log) — System spawn log captured by task-board
- [TASK-261005-22yioq_rework1-validation.md](file://TASK-261005-22yioq/TASK-261005-22yioq_rework1-validation.md) — Rework 1 validation: pinned research links, escaped table pipes, exit codes
- [TASK-261005-22yioq_table-regression.py](file://TASK-261005-22yioq/TASK-261005-22yioq_table-regression.py) — Named regression test: GFM table-width scan for CIP-0002..0006
- [TASK-261005-22yioq_table-render.py](file://TASK-261005-22yioq/TASK-261005-22yioq_table-render.py) — GFM render check for CIP-0002 operator-control table semantics
- [TASK-261005-22yioq_change-request_rev2.patch](file://TASK-261005-22yioq/TASK-261005-22yioq_change-request_rev2.patch) — Change Request CR-TASK-261005-22yioq-2 revision 2 candidate patch (repository_delta=present, 6 changed paths)
- [TASK-261005-22yioq_change-request_rev2-validation.log](file://TASK-261005-22yioq/TASK-261005-22yioq_change-request_rev2-validation.log) — Change Request CR-TASK-261005-22yioq-2 revision 2 bounded validation log
- [TASK-261005-22yioq_spawn-log_-reviewer--reviewer--codex-_RUN-261005-c5b436.log](file://TASK-261005-22yioq/TASK-261005-22yioq_spawn-log_-reviewer--reviewer--codex-_RUN-261005-c5b436.log) — System spawn log captured by task-board
- [TASK-261005-22yioq_review-verdict-rev2.md](file://TASK-261005-22yioq/TASK-261005-22yioq_review-verdict-rev2.md) — Revision 2 acceptance: both prior findings resolved, merged checklist, full surface sweep, independent docs/link/render/privacy gates and evidence limits
- [TASK-261005-22yioq_spawn-log_-implementer--developer--codex-_RUN-261005-6bf2d3.log](file://TASK-261005-22yioq/TASK-261005-22yioq_spawn-log_-implementer--developer--codex-_RUN-261005-6bf2d3.log) — System spawn log captured by task-board
- [TASK-261005-22yioq_complete-integration.log](file://TASK-261005-22yioq/TASK-261005-22yioq_complete-integration.log) — Bound worktree complete output; exit code 0. Landing proven, board commit published; cleanup_pending and shared_plane_deferred reported.

## Created
2026-10-05T12:28:54Z

## Last Update
2026-10-05T16:48:06Z

## Assigned To
[implementer] developer (codex)
