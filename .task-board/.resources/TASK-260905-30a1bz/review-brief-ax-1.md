# Review brief: ax PR #1 revision per Decision 0013 (cycle 1)

## Subject
- Worktree `/Users/iv/Developer/ReluxWorks/.temp/ax-curator-integration/worktree`, branch
  `draft/curator-environment-integration`, head `2c7f642` (new commits on top of `d7075e1`),
  base main `28bf96d` (v0.5.0). PR https://github.com/relux-works/agent-session-manager-spec/pull/1.
  Diff to review: `git diff d7075e1..2c7f642`; whole proposal: `git diff main...2c7f642`.
- Read `producer-brief-ax-pr1.md` and the revision report (resources on TASK-260905-30a1bz).
  Authority: curator-spec Decision 0013 Decisions 3.1–3.6, 4, 7 (items 1–8), 8 on curator-spec
  main `83de1a5`; ax `CONTRIBUTING.md` §2–§4 and Validation.

## Review dimensions
1. **Decision 7 items 1–8, each verified in the SPEC text**: §7.5 paragraph replaced; §5.1
   `stdin` row (`utf-8|base64url`, ≤65,536 decoded, non-secret) and the fourth key +
   `profile-pin` re-keyed; §7.3 `capability_names` = nine names in order with the manifest
   consequence stated; §7.5 `SpawnPlan.stdin`, `resume.launch_plan`; §13.1 planning-role
   `launch` for both forms, final argv recorded, `ax.launch-plan-request`
   `{form, base_argv_length, request_digest}` (no suffix copy), residual bound, determinism +
   `provider_protocol_error`; §13.10 refuse-on-drift for `system-modules: true`
   (`policy_refused`, `environment_drift`), warn otherwise, strict mode kept; CCJ-1 digest;
   §14.1 `--launch-plan FILE|-` row with exclusivity rules; §15.3 `launch_plan_invalid` (exit
   class 2) and secrets via `secret_policy_violation`; the profile-flag MUST rule keyed on §7.7
   with `reason: "profile_flag"` and `details.argv_index`; the required negative conformance
   cases ("in `argv` or `argv_suffix`"); §1.5 registry entry and Appendix D fixtures for
   `urn:ax:schema:launch-plan-request` 1.0.0; traceability rows citing Decision 0013.
2. **ax invariants untouched**: lease/fencing/checkpoint/materialization semantics, §1.6
   extension rules, task-board launch path (§13.2) unchanged, the three original keys kept.
3. **Validation, rerun yourself**: `./scripts/validate_spec.py`, `./scripts/test_expected_red.sh`,
   `git diff --check` — record exit codes; `./run_validation.sh` if the toolchain exists,
   otherwise the exact blocker. Confirm only the `SPEC.md` digest moved in the frozen map and
   that `VERSION`/`CHANGELOG.md`/`RELEASE_NOTES.md` are untouched.
4. **Delivery shape**: commits signed with the human identity, on top of `d7075e1` (no rewrite:
   `git log --oneline main..HEAD` starts with `d7075e1`), branch pushed by plain push
   (`gh pr view 1 --json headRefOid` equals local head), PR description updated and citing
   Decision 0013, PR still OPEN and not merged.

## Constraints
Read-only on the worktree and every checkout; scratch only under the worktree's `.temp/`.
Never write into the control root.

## Verdict contract
Attach `TASK-260905-30a1bz_review-findings-ax-1.md` (severity, section, quote, what is wrong,
fix, plus validator exit codes). Blocking/major → `development`; else explicit ACCEPT at
`to-review`. Do not mark done. `task-board handoff TASK-260905-30a1bz --role reviewer`.
