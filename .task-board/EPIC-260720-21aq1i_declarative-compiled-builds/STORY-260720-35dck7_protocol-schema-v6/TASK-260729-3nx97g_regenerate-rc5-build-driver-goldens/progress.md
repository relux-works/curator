## Status
done

## Review
required

## Task Class
code

## Estimate
estimated(fibonacci(8))

## Blocked By
- TASK-260720-1s1vr6
- TASK-260728-2kp3tv

## Blocks
- TASK-260720-2dnqw2
- TASK-260720-12r55p

## Checklist
- [x] Import exact accepted rc.5 candidate bytes and accepted rc.4 build-driver generator/fixture semantics into one isolated task worktree with recorded provenance
- [x] Regenerate complete manager-worker-v1 build-driver vectors and expected artifacts with independently checked positive and negative identities
- [x] Prove deterministic double regeneration, complete manifest inventory and byte preservation outside the owned golden surface
- [x] Run curator-spec validation/generator/release-candidate gates and Curator candidate metadata tests without skip using an explicit conformance root
- [x] Attach candidate-only evidence and hand off for independent review without landing, publication, tag or pin mutation
- [x] Code written per task description and AC
- [x] Relevant tests written for new or changed behavior and passing
- [x] Lint clean
- [x] Relevant build/validation commands run after changes and build not broken
- [x] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [x] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
- [x] Implementation matches AC
- [x] Solution fits project architecture
- [x] Tests green
- [ ] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches

## Notes
Created from independently accepted TASK-260729-1t1z2l cycle-2 finding. This task closes the rc.5 golden coverage hole but does not itself authorize rc.5 substitution, protected-branch landing, tag/sign/publish or downstream pin changes. After board/product authority approves rc.5 supersession, retarget TASK-260720-3ag6pi, TASK-260720-jrrgw9 and the seven documented CocoaSkills briefs and add exact dependency links; do not mutate those literal rc.4 gates before that approval.
EXECUTION DIRECTIVE (queued behind Kotlin Opus cooldown until 2026-07-29 06:26:49 +04): work only in an isolated task-owned curator-spec worktree. Import the exact accepted TASK-260728-2kp3tv rc.5 candidate tree (recorded SHA-256 3e4fd26acd9cafd1a76b2b5312da49ee35d234738263beb17a42be971d9dc582) and the independently accepted rc.4 schema-6 build-driver generator/fixture semantics; preserve both provenance chains and do not reinterpret accepted bytes. Regenerate the complete execution_policy=manager-worker-v1 build-driver golden surface, including positive/negative identities, manifest inventory, generator inputs, expected artifacts, and candidate metadata needed by Curator. Run two clean independent regenerations and require byte identity; prove every byte outside the owned golden surface is unchanged. Run curator-spec validation/generator/release-candidate gates and the Curator candidate metadata/conformance test without skip using explicit CURATOR_CONFORMANCE_ROOT. The existing rc.5 publication snapshot has zero build-driver golden entries, so absence, stale rc.4 naming, partial inventory, nondeterminism, or unrelated-byte drift is a hard failure. Produce candidate-only evidence and hand off to independent review. Do not stage, commit, land, publish, tag, sign, mutate pins, retarget any downstream task, or claim rc.5 supersession; those remain explicit human-authority steps. Do not spawn this Opus before the exact eligible time.
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [implementer] developer (claude) (run=RUN-260729-5f494f, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260729-5f494f)
Ready for review. Isolated worktree .temp/TASK-260729-3nx97g/worktree, detached from curator-spec 57c1f568, imported byte-exact from accepted rc.5 TASK-260728-2kp3tv (manifest 9ba9b8ec, diff -r exit 0). Build-driver golden suite carried forward from accepted TASK-260720-1s1vr6 (fixture imported byte-exact, generator/expected semantics from the exact 37ei85->1s1vr6 delta). Positive portable input requires execution_policy=manager-worker-v1 and independently recomputes cache key sha256:529370122ae11e2e961d5265b1a020e046bcd43165b2eb96b05e73a51187ac9b and receipt sha256:919fbbad8e6ce95532219fd952c2309d0d7026f85209650508fd6834af4020cd from the published bytes. Legacy rc.4 (3fcd714a, no execution_policy) and reserved-hardened (13736230, hardened-worker-v1) are explicit schema-invalid non-alias negatives, both as a cache_identity block and as named rejection cases; verdicts proved against the real compiled build-receipt-v1 schema, aliases=false, three distinct self-derived keys. Coverage: positives 7->8, rejections 75->77, build-source 10, toolchain 12, zero dropped; only supersession is the self-consistent forged receipt 9a23f5b7->e15a8b19, which must move because it embeds the portable input, and both Go and Python assert it stays self-consistent and no longer reproduces the rc.4 digest. Manifest 422->447 files, 25 added, 0 removed, 0 changed. New candidate pin sha256:b6f56aacc0e37dcc6692f73f641bff761e89b645adfe20a47a06d81c6fda204c; release/1.0.0-rc.5.json differs in exactly the two pin lines. All nine rc.4 byte-frozen artifacts and all schemas 1-7 byte-identical. Gates exit 0: generate, validate.py (42 schemas / 447 files), unittest 41 tests (was 29), go test/vet/gofmt/compileall, git diff --check, nothing staged, 0 commits after pin. Two further clean regenerations byte-identical at aggregate 6c989066. Disposable probe repo: make regenerate-check twice and make release-check VERSION=1.0.0-rc.5 all exit 0, porcelain 0. Curator TestCandidateBuildMetadataArtifacts now PASSES against the explicit candidate root (baseline accepted rc.5 root still SKIPs); full internal/buildmeta run exit 0 with 12 PASS and 0 SKIP, executed from a read-only copy with no Curator edits. Seven negative probes each exit 1 for the intended reason. Nothing staged, committed, tagged, pinned, published or claimed; source worktrees unchanged. Evidence: TASK-260729-3nx97g_results.md.
agent completed: [implementer] developer (claude) (exit=0)
spawn run completed: claude (run=RUN-260729-5f494f, pid=67931, exit=0)
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=codex; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [reviewer] reviewer (codex) (run=RUN-260729-04c6dd, max_parallel=20)
spawn run started: [reviewer] reviewer (codex) (run=RUN-260729-04c6dd)
Reviewer verdict: ACCEPTED. Independent evidence is attached as TASK-260729-3nx97g_review-verdict.md. Required identities, non-alias negatives, manifest preservation, deterministic regeneration, validation/test/release gates, and Curator no-skip candidate metadata integration all passed. Reviewer made no code changes and authorizes no landing or publication.
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-260729-04c6dd, pid=88819, exit=0)

## Precondition Resources
(none)

## Outcome Resources
- [TASK-260729-3nx97g_spawn-log_-implementer--developer--claude-_RUN-260729-5f494f.log](file://TASK-260729-3nx97g/TASK-260729-3nx97g_spawn-log_-implementer--developer--claude-_RUN-260729-5f494f.log) — System spawn log captured by task-board
- [TASK-260729-3nx97g_results.md](file://TASK-260729-3nx97g/TASK-260729-3nx97g_results.md) — rc.5 build-driver golden regeneration: provenance, exact identities, cluster coverage, gates and exit codes (rev 2)
- [TASK-260729-3nx97g_spawn-log_-reviewer--reviewer--codex-_RUN-260729-04c6dd.log](file://TASK-260729-3nx97g/TASK-260729-3nx97g_spawn-log_-reviewer--reviewer--codex-_RUN-260729-04c6dd.log) — System spawn log captured by task-board
- [TASK-260729-3nx97g_review-verdict.md](file://TASK-260729-3nx97g/TASK-260729-3nx97g_review-verdict.md) — Independent reviewer verdict and acceptance evidence

## Created
2026-07-29T01:14:39Z

## Last Update
2026-07-29T11:51:44Z

## Assigned To
[reviewer] reviewer (codex)
