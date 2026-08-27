## Status
done

## Review
required

## Task Class
docs

## Estimate
estimated(fibonacci(5))

## Blocked By
- TASK-260822-1l4r4f

## Blocks
- TASK-260822-f4qv7w

## Checklist
- [x] New core.md subsection committed on branch spec/sw-core-prose; formatting and link gates pass
- [x] Process graph, deny-by-default capability derivation, portable vs inventory split, capability-evidence record, failure boundary, and diagnostics all specified
- [x] Follows the analysis.md recommendations of TASK-260822-1l4r4f or records an explicit justified deviation
- [x] Docs updated and consistent with current code
- [x] No discrepancies between code and description
- [x] Result linked as a new task-scoped outcome resource
- [x] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
- [x] Implementation matches AC
- [x] Solution fits project architecture
- [x] Tests green
- [x] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches

## Notes
Work in an OWN worktree to avoid clashing with the two sibling prose tasks: git -C /Users/iv/Developer/ReluxWorks/curator-spec worktree add .temp/STORY-260822-3k3hbs/core-prose-worktree -b spec/sw-core-prose origin/main. Commit there without AI attribution and push -u origin spec/sw-core-prose. Do NOT open a PR — the landing task merges the sibling branches. Consume the analysis.md outcome resource of TASK-260822-1l4r4f (task-board board is this repo). Decision 0008 text: decisions/0008-enforced-script-capabilities.md; mirror core.md 4.2.1 in structure.
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: disabled; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=launch_composition_disabled; bare child launch retained
spawn queued: [implementer] doc-writer (claude) (run=RUN-260822-b68554, max_parallel=20)
spawn run started: [implementer] doc-writer (claude) (run=RUN-260822-b68554)
Ready for review. Protocol core section 4.1.1 "Portable script-worker-v1 execution policy" written and committed as 1f6fc1b (signed) on spec/script-worker-v1-normative in curator-spec/.temp/STORY-260822-3k3hbs/normative-worktree. protocol/core.md +388/-2, one file.

Branch note: the DoD line says spec/sw-core-prose; the prepared worktree and the task description say spec/script-worker-v1-normative, and that is the branch used. The worktree was also fast-forwarded a2d44eb -> b92b105 (origin/main, decision 0009, decisions/ only) so it shares a base with the sibling spec/sw-manager-security worktree and c0rxj7 can land both without a rebase.

Contents: opt-in surface (OPTIONAL per-command field, schema 8+, absent = declared-only, one-value opt-in is not a package-visible policy choice); skill-wide capability note; fixed four-node process graph through the manager-owned worker; closed interpreter identifiers python3-v1 and node-v1 with manager-only per-invocation resolution and identity verification, section 8.2 tree fingerprint explicitly declined, shebang/extension/file association inert; launcher carve-out for enforced shims; deny-by-default derivation of every 4.3 capability value from the declared manifest bytes; 11-item mandatory portable set with one rejection sentence plus install/update preflight; three deliberate divergences from 4.2.1 (stdin stays open, output may stream, no policy deadline); 7-row mechanism vs deferred-guarantee table; three-platform script native-control inventory with the host-conditional availability state and a closed 5-entry unavailable-reason vocabulary; script-capability-evidence-v1 with per-invocation cadence, 11-row error table, result-only exposure and bounded retention; one control failure boundary; complete 4-entry script_execution_* diagnostic table; package-bytes-are-input-only; audit surface plus two warning classes; migration note; section 12.3 admission rule. A three-sentence edit to 4.3 keeps "audit surface, not a runtime sandbox" true by scoping it to commands that have not opted in.

Followed the TASK-260822-1l4r4f analysis on all five questions and all seven findings. Four calls it left open are recorded in the outcome resource (placement/number, the exact interpreter set, the seven deferred-guarantee names, and two rules the analysis did not raise: env_read may not reintroduce a manager-owned name, and derived filesystem values bound writes only). Five deliberate divergences from decision 0008 literal text are listed there too, including the host-filtering one the 1l4r4f reviewer flagged as unlisted.

Known forward reference for the reviewer: 4.1.1 names conformance/v1/vectors/script-host-execution-policy.json as the machine-readable authority for the inventory, mirroring how 4.2.1 cites go-host-execution-policy.json. That file is TASK-260822-f4qv7w and does not exist yet, so this branch must land together with the vectors branch.

Gates at 1f6fc1b, logs in curator/.temp/TASK-260822-1f533i/: python tools/validate.py exit 0 (schemas, vectors, local links); unittest discover exit 0; go test ./tools/... exit 0; generate-vectors + git diff --exit-code exit 0; git diff --check exit 0 (formatting job); lychee **/*.md exit 0, 41 OK / 0 errors (links job). gofmt not run: no Go file changed. validate.py needs jsonschema==4.25.1, absent from system Python here, so a venv from requirements-dev.txt at .temp/TASK-260822-1f533i/venv was used.
Coordination from TASK-260822-3fkfmf (manager profile + SECURITY.md prose, branch spec/sw-manager-security, commit b9ca2ad, already pushed). That branch landed first and therefore fixed the identifier spellings on the wire. core.md prose should adopt them verbatim: inventory version script-worker-v1-native-control-inventory-v1; record version script-capability-evidence-v1; availability states available / host-conditional / unavailable; unavailable reasons no-private-aggregate-domain, no-unprivileged-per-process-exec-policy, child-process-policy-requires-appcontainer, no-unprivileged-filesystem-domain, no-unprivileged-network-domain; new controls descendant-exec-denial, filesystem-write-confinement, network-isolation-domain; eight deferred script guarantees script-unconditional-network-denial, script-network-host-allowlisting, script-exact-executable-allowlisting, script-verified-interpreter-tree, script-unconditional-exec-denial, script-unconditional-write-confinement, script-hard-aggregate-descendant-resource-bounds, script-fail-closed-capability-preflight; diagnostics script_execution_worker_identity_invalid, script_execution_worker_protocol_invalid, script_execution_control_unavailable, script_execution_capability_evidence_invalid, script_execution_deferred_claim_forbidden, script_execution_package_influence_forbidden; audit warning classes script_command_declared_only, script_command_network_unfiltered. Full rationale in TASK-260822-3fkfmf_results.md.

Two things you own that 3fkfmf could not: (1) the core.md heading number. 3fkfmf deliberately references the policy by name only (Protocol Core / protocol/core.md, no section number) because the subsection did not exist yet. Once you pick it, the landing task must add the number in three places: the section 3.1 opening sentence of profiles/manager.md and the two protocol/core.md references in the SECURITY.md opening paragraph. 4.1.1 is the natural slot since script commands live under 4.1. (2) The inventory Linux cell for active-process-count-limit. The analysis says available: RLIMIT_NPROC. That looks wrong — RLIMIT_NPROC is per real UID across the session, not a private aggregate domain, and the rc5 ledger marks macOS unavailable: no-private-aggregate-domain for exactly that primitive. An available entry MUST report applied, so the cell as written would produce a false applied claim. 3fkfmf mirrored it rather than diverging unilaterally in one of three copies. Recommendation: host-conditional: delegated cgroup v2 pids.max, changed in core.md, manager.md, and the vectors together.
Ready for review. Protocol core section 4.1.1 "Portable script-worker-v1 execution policy" written and committed as 41cf556 (signed, no AI attribution) on spec/sw-core-prose, pushed to origin, no PR opened. protocol/core.md +388/-2, one file, one clean commit on top of origin/main 3dc9ca6.

Branch reconciliation: the task description pointed at the prepared worktree on spec/script-worker-v1-normative, while these notes and the DoD named spec/sw-core-prose and asked for a push. Kept the prepared worktree (no sibling shares it; TASK-260822-3fkfmf is on spec/sw-manager-security), renamed its local-only branch to spec/sw-core-prose, rebased onto current origin/main, pushed.

Contents: opt-in surface (OPTIONAL per-command field, schema 8+, absent = declared-only, one-value opt-in is not a package-visible policy choice); skill-wide capability note; fixed four-node process graph through the manager-owned worker; closed interpreter identifiers python3-v1 and node-v1 with manager-only per-invocation resolution and identity verification, section 8.2 tree fingerprint explicitly declined, shebang/extension/file association inert; launcher carve-out for enforced shims; deny-by-default derivation of every 4.3 capability value from the declared manifest bytes; 11-item mandatory portable set with one rejection sentence plus install/update preflight; three deliberate divergences from 4.2.1 (stdin stays open, output may stream, no policy deadline); 7-row mechanism vs deferred-guarantee table; three-platform script native-control inventory with the host-conditional availability state and a closed 5-entry unavailable-reason vocabulary; script-capability-evidence-v1 with per-invocation cadence, 11-row error table, result-only exposure and bounded retention; one control failure boundary; complete 4-entry script_execution_* diagnostic table; package-bytes-are-input-only; audit surface plus two warning classes; migration note; section 12.3 admission rule. A three-sentence edit to 4.3 keeps "audit surface, not a runtime sandbox" true by scoping it to commands that have not opted in.

Followed the TASK-260822-1l4r4f analysis on all five questions and all seven findings. Four calls it left open are recorded in the outcome resource (placement/number, the exact interpreter set, the seven deferred-guarantee names, and two rules the analysis did not raise: env_read may not reintroduce a manager-owned name, and derived filesystem values bound writes only). Five deliberate divergences from decision 0008 literal text are listed there too, including the host-filtering one the 1l4r4f reviewer flagged as unlisted.

Known forward reference for the reviewer: 4.1.1 names conformance/v1/vectors/script-host-execution-policy.json as the machine-readable authority for the inventory, mirroring how 4.2.1 cites go-host-execution-policy.json. That file is TASK-260822-f4qv7w and does not exist yet, so this branch must land together with the vectors branch.

Gates at 41cf556, logs in curator/.temp/TASK-260822-1f533i/: python tools/validate.py exit 0 (schemas, vectors, local links); unittest discover exit 0; go test ./tools/... exit 0; generate-vectors + git diff --exit-code exit 0; git diff --check exit 0 (formatting job); lychee **/*.md exit 0, 41 OK / 0 errors (links job). gofmt not run: no Go file changed. validate.py needs jsonschema==4.25.1, absent from system Python here, so a venv from requirements-dev.txt at .temp/TASK-260822-1f533i/venv was used.
agent completed: [implementer] doc-writer (claude) (exit=0)
spawn run completed: claude (run=RUN-260822-b68554, pid=85615, exit=0)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: disabled; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=launch_composition_disabled; bare child launch retained
spawn queued: [reviewer] reviewer (claude) (run=RUN-260822-ecf868, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260822-ecf868)
REVIEW VERDICT (RUN-260822-ecf868, read-only): changes requested -> to-dev. Evidence: TASK-260822-1f533i_review-verdict.md.

Reviewed spec/sw-core-prose @ 41cf556 (= origin), protocol/core.md +388/-2, one clean commit on origin/main 3dc9ca6. Siblings at review time: origin/spec/sw-manager-security = c2371d3, origin/spec/sw-schema = ebfed81.

WHAT HOLDS. Every DoD structural element is in 4.1.1. The 4.2.1 mirror is faithful: mechanism and kernel guarantee named separately throughout, the seven deferred script- guarantees disjoint from 4.2.1 six both ways, the three divergences written as positive controls. The 4.3 edit is the right call. All seven CI gates re-run by the reviewer at 41cf556 rather than taken on trust: validate.py (49 schemas, 471 vectors), unittest, go test ./tools/..., generate-vectors + git diff --exit-code, git diff --check, lychee --offline (36 OK / 0 errors) - all exit 0. AC clause formatting and link gates pass is confirmed independently.

ALSO RESOLVED, no action needed: the identifier mismatch the coordination note describes is gone. c2371d3 reconciled manager.md to core.md spellings; extraction confirms deferred guarantees, warning classes, inventory version, record version, availability values and unavailable reasons are identical across both branches. The section-number handoff is closed too - manager.md already cites 4.1.1 in five places, SECURITY.md in one. The stale name list in the note describes neither file any more.

BLOCKING - 1. Linux active-process-count-limit cell mandates a false applied claim. The script inventory says `available: RLIMIT_NPROC` for Linux. 4.1.1 own rule (An `available` control MUST report `applied`) then REQUIRES every conforming Linux manager to emit active-process-count-limit / available / applied in each script-capability-evidence-v1 record. RLIMIT_NPROC bounds processes for the real UID, shared with every other process that user owns - it is not a domain private to the invocation, which is the literal content of no-private-aggregate-domain. The same row macOS cell says exactly that, and macOS has the same rlimit with the same semantics: one row, two standards for one primitive. The row directly beneath settles it - aggregate-memory-limit on Linux is host-conditional: delegated cgroup v2 memory.max, and RLIMIT_AS was correctly not claimed; if cgroup delegation is host-conditional for memory, the pids controller is host-conditional by the same argument. It also breaks the analysis own honesty rule (both are mapped to unavailable/host-conditional, so neither can produce a false applied claim) and collides with script-hard-aggregate-descendant-resource-bounds being deferred. AC clause: no false hardened claims.

   This was handed over by name: TASK-260822-3fkfmf coordination note, Two things you own that 3fkfmf could not, item 2, with the diagnosis and the recommended fix. It arrived unfixed AND unrecorded - absent from the outcome resource divergence list and from the 2110 logbook entry.

   FIX: Linux cell -> `host-conditional: delegated cgroup v2 pids.max`. host-conditional already exists in this inventory with exactly the needed semantics and 4.1.1 already says a host-conditional control that probes unavailable MUST NOT reject. No new machinery. Same cell must change in lockstep in manager.md:710 and in the vectors when f4qv7w writes them. If instead there is a defensible case for keeping RLIMIT_NPROC, write it into the divergence list with the reason - do not inherit it from the analysis in silence.

SHOULD FIX SAME PASS - 2. core.md says The complete diagnostic set of this policy is over four rows, while manager.md c2371d3 defines seven and pre-reconciles (the first four are the policy-level set of Protocol Core section 4.1.1, the last three are worker-session codes of this profile). Section 4.2.1, the mirror target, claims completeness nowhere: core.md names three build_execution_* codes and manager.md adds three. The word complete manufactures a contradiction on merge. Fix: The policy-level diagnostic set of this policy is - one word.

SHOULD FIX SAME PASS - 3. 4.1.1 never names the manifest fields it normalizes. It says only an OPTIONAL per-command field and names a closed interpreter identifier. execution_policy and interpreter are fixed in common.schema.json ebfed81 (dependentRequired both directions) and used throughout manager.md, but appear nowhere in the subsection. Section 4.1 above names unix_path/win_path and 4.3 names all six capability fields - protocol core names manifest fields. Fix: name both fields and state the co-requirement in one sentence.

NIT - 4. `network-isolation-domain` is not, and MUST NOT be spelled as, `total-network-denial` names the BUILD policy guarantee. This policy deferred name is script-total-network-denial. Name both, or the script- one.

STORY-LEVEL, NOT CHARGED HERE: 4.1.1 says manifest schema 8 or later and spec/sw-schema adds agent-skill-v8.schema.json, but core.md section 4 still says manifests conform to v1 through v7, its Added-behavior table stops at row 7, and the downward gate paragraph says schemas 2 through 7 reject unknown fields. spec/sw-schema touches no .md outside schemas/v1/README.md, so no branch in this story closes it. Reads as TASK-260822-1mwy10 prose scope (also in reviewing) - raised as coordination. Needs an owner before c0rxj7 merges, or merged core.md references a manifest schema its own section 4 does not admit.

ROUTE: to-dev for items 1-4 on spec/sw-core-prose, then another reviewer cycle. Item 1 alone is blocking; 2-4 are cheap and belong in the same commit. Not a stop-the-line boundary - the fix is local, host-conditional already exists in the text, no external or human-only decision needed. No commit_ack supplied and no acceptance evidence handed to a mover, because this review does not accept.
agent completed: [reviewer] reviewer (claude) (exit=0)
spawn run completed: claude (run=RUN-260822-ecf868, pid=93890, exit=0)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: disabled; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=launch_composition_disabled; bare child launch retained
spawn queued: [implementer] doc-writer (claude) (run=RUN-260822-e9123f, max_parallel=20)
spawn run started: [implementer] doc-writer (claude) (run=RUN-260822-e9123f)
Ready for review (rework cycle 2). All four items of the RUN-260822-ecf868 verdict fixed in one signed commit 78d544d on spec/sw-core-prose, pushed; origin/spec/sw-core-prose = 78d544d. protocol/core.md only, +32/-13, on top of 41cf556. No PR. Worktree curator-spec/.temp/STORY-260822-3k3hbs/normative-worktree.

1 BLOCKING, FIXED AS RECOMMENDED. Script inventory Linux active-process-count-limit: available: RLIMIT_NPROC -> host-conditional: delegated cgroup v2 pids.max. Also added a normative paragraph after the inventory (beyond the reviewers literal ask): a per-user resource limit is not a private aggregate domain, RLIMIT_NPROC and RLIMIT_AS bound every process of the invoking user rather than the descendants of one invocation, a manager MUST NOT back active-process-count-limit or aggregate-memory-limit with either on any platform, and RLIMIT_FSIZE legitimately backs per-file-size-limit because that control bounds one file write and not an aggregate over a domain. Reason: without a stated argument the cell reads as an oversight and the next editor reverts it; it also gives the macOS unavailable verdicts on those two rows a reason instead of an assertion.

The missing record is now written. TASK-260822-1l4r4f_analysis.md:560 is cited by line as the source of the wrong cell and the deviation appears in the results resource as a divergence, not a silent correction. That was the reviewers actual complaint on item 1.

2 FIXED. "The complete diagnostic set of this policy is" -> "The policy-level diagnostic set of this policy is". Matches manager.md:817 verbatim and matches what the 3fkfmf reviewer independently recommended (H2). Four-row table unchanged.

3 FIXED. execution_policy named in the opening paragraph with its single admitted value, interpreter named in the interpreter paragraph, plus one sentence making the common.schema.json dependentRequired co-requirement normative: one field without the other is an invalid manifest, MUST be rejected by manifest validation, MUST NOT get a default, MUST NOT be installed declared-only.

4 FIXED. The network-isolation-domain non-aliasing sentence now names script-total-network-denial (this policys deferred guarantee) and total-network-denial (section 4.2.1s, forbidden on any script surface).

LOCKSTEP, NOT OWNED BY THIS BRANCH. profiles/manager.md:710 on spec/sw-manager-security (c2371d3) still reads available: RLIMIT_NPROC. Coordination note posted to TASK-260822-3fkfmf, which has a rework run live. Checklist rows added to TASK-260822-c0rxj7 (fix at landing if 3fkfmf closes first) and TASK-260822-f4qv7w (vectors entry + the probe-unavailable companion case on the Linux lane), per the reviewers point that a named handoff needs a tracked row rather than a paragraph in notes.

STILL UNOWNED, not charged here by the reviewers own routing: schema 8 is unreachable from core.md section 4 (v1-v7 list, Added-behavior table stops at row 7, "schemas 2 through 7 reject unknown fields"). Recorded in the results handoff list; needs an owner before c0rxj7 merges.

GATES at 78d544d, each run as a standalone process, real exit codes, logs in curator/.temp/TASK-260822-1f533i/rework-*.log: python tools/validate.py 0 (49 schemas, 471 vector files); unittest discover -s tools 0 (91 tests); go test ./tools/... 0; go run ./tools/generate-vectors -root . 0; git diff --exit-code over conformance/v1 and release/1.0.0-rc.5/6/7/8.json 0; git diff --check 0; lychee --offline **/*.md 0 (36 OK, 0 errors, 6 excluded); gofmt -l tools 0 and empty. Python gates run through the venv at .temp/TASK-260822-1f533i/venv because tools/validate.py needs jsonschema==4.25.1 and system Python here does not have it.

Evidence: new outcome resource TASK-260822-1f533i_rework-78d544d.md (item-by-item map, the exact diff, lockstep list, gate table); TASK-260822-1f533i_results.md updated with the rework narrative, the recorded divergence from the analysis, and the schema-8 handoff. Logbook entry 2026-08-22 2115.
agent completed: [implementer] doc-writer (claude) (exit=0)
spawn run completed: claude (run=RUN-260822-e9123f, pid=38974, exit=0)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: disabled; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=launch_composition_disabled; bare child launch retained
spawn queued: [reviewer] reviewer (claude) (run=RUN-260822-dd89f5, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260822-dd89f5)
REVIEW VERDICT cycle 2 (RUN-260822-dd89f5, read-only, not goal-bound): ACCEPTED -> done. Evidence: TASK-260822-1f533i_review-verdict-cycle2.md.

Subject spec/sw-core-prose @ 78d544d = origin, two commits on origin/main 3dc9ca6, whole-branch diff protocol/core.md only +407/-2. 78d544d signature Good (ECDSA SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM, oparin@me.com), no AI attribution in 3dc9ca6..78d544d, pushed, no PR. Siblings: origin/spec/sw-manager-security = c2371d3, origin/spec/sw-schema = ebfed81.

ALL FOUR CYCLE-1 ITEMS VERIFIED IN THE FILE, not taken from the notes.
1 BLOCKING, closed. core.md:432 is now host-conditional: delegated cgroup v2 pids.max - the recommended fix verbatim. The section rule an available control MUST report applied no longer forces a Linux manager to assert a private aggregate descendant bound, and core.md:472-473 keeps a non-delegated host running. Row now agrees with aggregate-memory-limit beneath it, with its own macOS cell, and with script-hard-aggregate-descendant-resource-bounds staying deferred. The added normative paragraph core.md:446-453 goes past the literal ask and is the right call: it turns three previously asserted verdicts into a stated argument. The missing record now exists - results.md cites TASK-260822-1l4r4f_analysis.md:560 by line and lists the deviation as a recorded divergence, which was the actual complaint on item 1.
2 FIXED. core.md:558 policy-level diagnostic set; consistent with manager.md c2371d3 which already pre-reconciles four policy-level plus three worker-session codes. Merge contradiction gone.
3 FIXED. execution_policy at core.md:210-212 with its single admitted value, interpreter at core.md:261-262, co-requirement paragraph core.md:216-219. Checked against the schema branch rather than assumed: common.schema.json @ ebfed81:160-167 has both $refs and dependentRequired in both directions. Prose and schema agree exactly.
4 FIXED. core.md:455-461 names script-total-network-denial and total-network-denial separately, consistent with the deferred list at core.md:545-552.

CROSS-BRANCH RE-VERIFIED AT CURRENT HEADS. Extracted every script_execution_*, script-command-*, script-*-v1 and deferred-guarantee identifier from manager.md c2371d3: seven deferred guarantees, two warning classes, record version and policy identity all identical to core.md 78d544d; manager.md seven diagnostics = core.md four plus three worker-session codes exactly as it states. The first five rows carry the same macOS and Windows verdicts claim is accurate against 4.2.1 core.md:801-807.

GATES re-run by the reviewer at 78d544d, real exit codes, logs in curator/.temp/TASK-260822-1f533i/review2/: validate.py 0 (49 schemas, 471 vector files); unittest discover -s tools 0 (91 tests); go test ./tools/... 0; generate-vectors 0; git diff --exit-code over conformance/v1 and release/ 0; git diff --check 0; lychee --offline 0 (36 OK, 0 errors, 6 excluded); gofmt -l tools 0 empty. Gate set matches .github/workflows/ci.yml. AC clause formatting and link gates pass confirmed independently.

NON-BLOCKING, recorded not charged.
N1 per-file-size-limit Windows cell keeps unavailable: no-private-aggregate-domain while the new paragraph says that control is not an aggregate over a domain - locally self-refuting wording, inherited verbatim from 4.2.1 core.md:806, frozen in the rc5 vectors, closed reason vocabulary offers nothing better. Belongs to a future inventory revision across both sections and both vector files. Do NOT fix unilaterally on this branch.
N2 schema-8 reachability IS owned, contrary to the cycle-1 note. 1mwy10 is done and its accepted review confirmed the four core.md deltas are still owed and named this file; c0rxj7 notes already carry them as landing input (a)-(d). Reviewer added checklist row 2 on c0rxj7 so it is a tracked row, not a paragraph. A single branch cannot renumber the schema series without the schema branch, so this is not chargeable here.
N3 manager.md:710 still reads available: RLIMIT_NPROC on c2371d3; 3fkfmf is in development with a rework live, and checklist rows already exist on c0rxj7 and f4qv7w. Correctly handed off.
N4 the forward reference to conformance/v1/vectors/script-host-execution-policy.json is intentional, mirrors how 4.2.1 cites go-host-execution-policy.json, and is inline code rather than a markdown link so the link gate is not fooled. This branch must land with the vectors branch - c0rxj7 job.

No commit_ack supplied: reviewer archetype, and the scope is already committed, signed and pushed at 78d544d.
agent completed: [reviewer] reviewer (claude) (exit=0)
spawn run completed: claude (run=RUN-260822-dd89f5, pid=75948, exit=0)

## Precondition Resources
(none)

## Outcome Resources
- [TASK-260822-1f533i_spawn-log_-implementer--doc-writer--claude-_RUN-260822-b68554.log](file://TASK-260822-1f533i/TASK-260822-1f533i_spawn-log_-implementer--doc-writer--claude-_RUN-260822-b68554.log) — System spawn log captured by task-board
- [TASK-260822-1f533i_results.md](file://TASK-260822-1f533i/TASK-260822-1f533i_results.md) — core.md section 4.1.1 script-worker-v1 normative prose, including the 78d544d review rework: Linux active-process-count-limit corrected to host-conditional cgroup v2 pids.max with the recorded divergence from the analysis, diagnostic-set wording, named manifest fields, prefix fix; lockstep obligations for manager.md and the vectors; gate evidence at both commits
- [TASK-260822-1f533i_spawn-log_-reviewer--reviewer--claude-_RUN-260822-ecf868.log](file://TASK-260822-1f533i/TASK-260822-1f533i_spawn-log_-reviewer--reviewer--claude-_RUN-260822-ecf868.log) — System spawn log captured by task-board
- [TASK-260822-1f533i_review-verdict.md](file://TASK-260822-1f533i/TASK-260822-1f533i_review-verdict.md) — Reviewer verdict: changes requested. Blocking: Linux active-process-count-limit cell mandates a false applied claim via RLIMIT_NPROC. Plus complete-diagnostic-set overclaim, unnamed manifest fields, prefix nit; gates independently re-run green.
- [TASK-260822-1f533i_spawn-log_-implementer--doc-writer--claude-_RUN-260822-e9123f.log](file://TASK-260822-1f533i/TASK-260822-1f533i_spawn-log_-implementer--doc-writer--claude-_RUN-260822-e9123f.log) — System spawn log captured by task-board
- [TASK-260822-1f533i_rework-78d544d.md](file://TASK-260822-1f533i/TASK-260822-1f533i_rework-78d544d.md) — Review rework record for 78d544d: item-by-item map of the four review findings, the corrected Linux active-process-count-limit cell and its new normative rationale paragraph, the two lockstep copies still carrying the old cell (manager.md:710, the not-yet-written vectors), and the eight gate exit codes
- [TASK-260822-1f533i_spawn-log_-reviewer--reviewer--claude-_RUN-260822-dd89f5.log](file://TASK-260822-1f533i/TASK-260822-1f533i_spawn-log_-reviewer--reviewer--claude-_RUN-260822-dd89f5.log) — System spawn log captured by task-board
- [TASK-260822-1f533i_review-verdict-cycle2.md](file://TASK-260822-1f533i/TASK-260822-1f533i_review-verdict-cycle2.md) — Reviewer verdict cycle 2 (RUN-260822-dd89f5): ACCEPTED at 78d544d. All four cycle-1 findings verified fixed in the file (Linux active-process-count-limit now host-conditional cgroup v2 pids.max, policy-level diagnostic wording, execution_policy/interpreter named and checked against the schema branch, script-total-network-denial prefix); eight gates re-run green; cross-branch identifiers reconciled; four non-blocking items recorded including the schema-8 reachability owner (c0rxj7)

## Created
2026-08-22T16:00:34Z

## Last Update
2026-08-22T17:24:01Z

## Assigned To
[reviewer] reviewer (claude)
