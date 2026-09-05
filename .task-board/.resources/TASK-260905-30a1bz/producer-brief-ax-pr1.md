# Producer brief: revise ax PR #1 (Curator integration) per Decision 0013

## Where and what

- Worktree `/Users/iv/Developer/ReluxWorks/.temp/ax-curator-integration/worktree`, branch
  `draft/curator-environment-integration` (PR #1 of `relux-works/agent-session-manager-spec`,
  head `d7075e1`, base main `28bf96d` = v0.5.0). Work on top of `d7075e1` with new commits;
  never rewrite it. The PR stays a proposal: do not merge it, do not touch `VERSION`,
  `CHANGELOG.md`, `RELEASE_NOTES.md`, or `README.md` beyond what the validator forces.
- Authority: curator-spec Decision 0013 (`decisions/0013-execution-ownership-and-launch-plans.md`
  at `83de1a5` on `~/Developer/ReluxWorks/curator-spec` main): Decisions 3 (all of
  3.1–3.6), 4, 7 (items 1–8), 8. Read `CONTRIBUTING.md` §2–§4 and Validation, then the
  sections of `SPEC.md` you touch: §1.5 registry, §1.6, §2.1, §2.4, §5.1, §7.3, §7.5, §7.7,
  §13.1, §13.10, §14.1, §15.3, §16.2, Appendix A/D as the traceability rules require.
- Deliverable: signed commits (`git commit -S`; paste `git log --show-signature` lines),
  `./scripts/validate_spec.py` exit 0, `./scripts/test_expected_red.sh` exit 0,
  `git diff --check` exit 0 (record exact exit codes; `./run_validation.sh` needs
  structurizr/plantuml — run it, and if the toolchain is missing say exactly which step
  failed and why). Push the branch to `origin` with a plain push (fast-forward of the PR
  branch; never `--force`), so PR #1 updates in place. Update the PR description with
  `gh pr edit 1 --body-file …` to summarize the revision and cite Decision 0013. Do not
  merge. Never write LOGBOOK.md or anything into the control root.

## The revision (Decision 0013 Decision 7, item by item)

1. **§7.5 paragraph**: replace "the launcher merges the env variables … into `env_literals`"
   with the `--launch-plan` operation; withdraw "adds no member to `SpawnPlan`".
2. **§5.1**: Launch Plan table gains `stdin` (Decision 4 shape: `{encoding: utf-8|base64url,
   bytes}`, ≤65,536 decoded bytes, non-secret, absent/null = terminal); the extension
   paragraph gains `works.relux.curator.system-modules` (boolean) and re-keys
   `profile-pin` to "the `sha256:`-prefixed lock hash of the profile as resolved at launch"
   (Decision 8 explains the pre/post-revision distinguishability).
3. **§7.3/§7.5**: `capability_names` gains `caller_launch_plan` (8th) and
   `stdin_resume_replay` (9th), appended in order; state the manifest-schema consequence
   for the maintainer; `SpawnPlan` gains `stdin`; `resume` request gains `launch_plan`.
   Plugin contract per Decision 3.5: verbatim translation, no reordering/dedup/rewrite,
   no second spelling; `capability_unavailable` (`details.capability`) before invocation
   for a record with `ax.launch-plan-request` against a non-declaring plugin.
4. **§13.1**: the planning-role `launch` step before persistence for the `argv_suffix` form;
   final argv recorded in `launch_plan.argv`; the `ax.launch-plan-request` extension
   `{form, base_argv_length, request_digest}` (no `argv_suffix` copy — review F1; the suffix
   is `argv[base_argv_length:]`); residual bound: caller extensions + ax key + Curator keys
   fit §1.6 or `launch_plan_invalid` `field: "extensions"`; determinism requirement and
   `provider_protocol_error` on mismatch.
5. **§13.10 drift**: refuse by default (`policy_refused`, exit 16, `details.reason:
   "environment_drift"`) when `system-modules: true` and the resolved lock hash differs;
   warn-and-continue otherwise; strict mode stays; failed resolution stays distinct.
6. **`fragment-digest`**: over CCJ-1 canonical bytes (curator-spec `protocol/registry.md` §1)
   of the parsed fragment, not the pretty-printed output.
7. **§14.1**: `ax start NAME --provider ID --launch-plan FILE|- [--profile standard|yolo]
   [--workspace PATH]`; exclusivity with `--task-board` = `invalid_arguments`; `argv` form
   + `--profile yolo` = `invalid_arguments`; document shape and validation per Decision
   3.2/3.3 (`launch_plan_invalid` new code, exit class 2, `details.field`; secrets →
   existing `secret_policy_violation` exit 16 — review F3); the profile-flag rule (review
   F2): a `caller_launch_plan` plugin MUST refuse a caller element equal to a flag of its
   own §7.7 `yolo` mapping with `launch_plan_invalid`, `reason: "profile_flag"`,
   `details.argv_index`; required negative conformance case for it (Appendix D / §19 as the
   repo's fixture rules require). The `curator session` informative note stands.
8. **Version**: propose v0.6.0 in the PR description only; the maintainer decides. Add the
   new error code to §15.3, the new schema `urn:ax:schema:launch-plan-request` 1.0.0 to the
   §1.5 registry and Appendix D fixture catalog with a valid and an invalid fixture under
   `fixtures/` following the repo's fixture conventions, and traceability rows per
   CONTRIBUTING §3 (cite Decision 0013 as the settled-decision input).

If the validator's frozen-baseline digest map must move for the changed `SPEC.md` (as
`d7075e1` did), update only the `SPEC.md` digest and say so.

## Board

Task `TASK-260905-30a1bz` (curator board; `TASK_BOARD_DIR` set). Attach `TASK-260905-30a1bz_revision-report.md`
(commits + signatures, item → SPEC-section table, validator exit codes, PR URL and new head,
anything unverified labeled) as an outcome resource, then
`task-board handoff TASK-260905-30a1bz --role developer`.
