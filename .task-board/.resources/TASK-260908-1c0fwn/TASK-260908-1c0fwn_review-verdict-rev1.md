# TASK-260908-1c0fwn — review verdict rev1: ACCEPTED

Reviewer run RUN-260907-e60ef7, read-only. CR-TASK-260908-1c0fwn-1, base 484933b, candidate tree d1f46b8 (empty delta).

## Why an empty repository delta is the right outcome
The brief (Env-contract-brief.md) forbids code, spec edits, ax calls and PR mutation; the deliverable is a bounded source analysis plus an erratum/test proposal. The three outcome resources are that deliverable. No repo change was expected or permitted.

## What I verified myself (exit 0 unless noted)
- Tag identity: `v0.5.10` is an annotated tag object `13d167e2c5cabb226eb0563f591ae6c63095855a`, signed by the operator; its peeled commit is `12f443d10bc217ca7a48e2edab19c739f441df9c`. The results cite only the tag-object SHA. Every `file:line` citation (plan.go:38/280, system.go:419/534-541, claude/env.go CLAUDECODE, runcontext.go) was checked against the peeled commit and holds. `go mod download -json` returns `Sum h1:mrpOOLcW+YaNvwdsQldULZ7H5WLTSoZoUVgFvLdinAQ=`, equal to the results and the probe go.sum. Precision remark, non-blocking: the erratum text should carry the peeled commit next to the tag object.
- Producer probe rerun from `.temp/TASK-260908-1c0fwn/probe`: 5/5 PASS, output identical to `probe-01.log`.
- `ChildEnv(nil, req)` extraction: no `os.Environ/Getenv/LookupEnv` call in claude/codex ChildEnv, runcontext, runtimeenv or launchenv at the peeled commit (only comments). Own values therefore cannot depend on the launcher's inherited environment, and no second `BuildPlan` is needed (ErrNoPath refusal confirmed by producer T-probe). Reviewer probe `TestOwnNamesSubsetOfFullEnv`: for both plugins, keys(ChildEnv(nil,req)) with a Run context = exactly the five TASK_BOARD_* names, value-equal to the same keys in ChildEnv(parent,req); inherited `SECRET` absent; claude strips CLAUDECODE; `TASK_BOARD_RUN_ID` rewritten. With the SPEC §4.4 closed request (no Run) the own-name set is empty, as the results state.
- Mutant check (producer T2 narrowing): using Plan.Env in place of ChildEnv(nil) serializes the inherited secret; `TestMutantFullEnvAsLiteralsFails` PASS = the mutant is detectable.
- Pinned vs future: the results consistently scope claims with "at v0.5.10" and propose T5 as a drift guard; no universal uniqueness/parity claim is made.
- Untracked removals (CLAUDECODE, codex runtime family, token pointer + token, run-context keys, PATH rewrite) and inherited-value retention: reproduced. Tracked: inherited names and PATH cannot reach env_literals by construction. Destination-local inherited env stays ax's layer (0013 §6.4, read). No `env_unset` member in 0013 §3.2 (read, closed grammar); residual correctly recorded, and the ax-plugin strip question is honestly reported as unread/unknown.
- Decision 0013 is *Proposed, draft* (@87a0d00, file 83de1a5): amendment, not erratum. Correct.

## Findings for the follow-up stories (non-blocking)
1. Proposed T4 second half (`env_names=[CLAUDE_CONFIG_DIR]` -> dropped) is a composer-unit fixture, not an accepted end-to-end fragment: CLAUDE_CONFIG_DIR is a reserved registry adapter name excluded by Decision 0012 D6 before the launcher. Label it so. At v0.5.10 with the closed §4.4 request the own-name set is empty, so the collision warning has no reachable positive case from an accepted fragment; the load-bearing negative is the inherited-name case (`FIGMA_API_KEY` must stay in env_names). Implementation story must keep that negative test.
2. Erratum item 1 should cite the peeled commit alongside the tag object.

## Definition-of-done mapping
Tagged semantics verified for both modes; correction + deterministic tests + unexpressible constraint attached; no ax mutation; gate behavior attacked via mutant and hidden-dependency grep, not read only. No logbook write (operator rule); this resource is the record. Reviewer probe attached as TASK-260908-1c0fwn_review_probe_test.go; run under `.temp/TASK-260908-1c0fwn-review/`.
