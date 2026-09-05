# Producer brief: Decision 0013 rework 1

Worktree `/Users/iv/Developer/ReluxWorks/.worktrees/curator-spec-decision-0013`, branch
`draft/decision-0013-execution-ownership`, head `71ac9d1`. Edit ONLY
`decisions/0013-execution-ownership-and-launch-plans.md`. Findings:
`TASK-260905-2ft7ts_review-findings-0013-1.md` (outcome resource on this task): 2 major,
3 minor, 2 nit. Apply ALL seven. The orchestrator has taken the author decisions below —
implement exactly these; deviate only with a recorded rationale in your rework report.

## Author decisions per finding
- **F1 (major)** — drop `argv_suffix` from the `ax.launch-plan-request` extension value.
  Keep `form`, `base_argv_length`, `request_digest`. Define the suffix as
  `launch_plan.argv[base_argv_length:]` and make Decision 3.5 replay from there. State the
  residual bound: caller `extensions` plus the ax key plus the Curator keys MUST fit the ax
  §1.6 extensions bound; a document that would not is refused at §3.3 time with
  `launch_plan_invalid`, `field: "extensions"`.
- **F2 (major)** — make it MUST: a `caller_launch_plan` plugin MUST refuse, before process
  creation, any caller-supplied element (in `argv` or `argv_suffix`) equal to a flag of its
  own ax §7.7 `yolo` profile mapping (long form or documented alias) with
  `launch_plan_invalid`, `reason: "profile_flag"`, `details.argv_index`. 0013 lists no
  spellings; the rule keys on §7.7. Add to Decision 7 the matching required negative
  conformance case for the ax PR: a suffix carrying the provider's yolo flag under
  `--profile standard` is refused. Keep "the composer never emits one".
- **F3 (minor)** — secret-rule violations in a caller document refuse with the existing ax
  `secret_policy_violation` (exit 16) with `details.field`; `launch_plan_invalid` covers
  shape, limits, exclusivity, and extension collisions only. Cite ax §15.3.
- **F4 (minor)** — `encoding` is exactly `utf-8` or `base64url` (unpadded, ax §1.6).
- **F5 (minor)** — add one composer rule in Decision 6.3 (and reference it from 6.4): when a
  name appears both in the composed `env_literals` (plan `Env` ⊕ fragment `env` ⊕
  variable-kind channel) and in the fragment `mcp.env_names`, the composer drops it from
  `env_names` — a literal the composer set is an explicit intent and wins over a
  destination-local lookup — and prints a warning naming the variable; the document is
  disjoint by construction and ax §5.1 disjointness never fires for a composed document.
- **F6 (nit)** — default session name becomes `<env-id>-<utc-stamp>`; the profile name is
  already in `works.relux.curator.profile-name`; `--name` overrides; note the ax §2.1
  64-char grammar and that the default always fits.
- **F7 (nit)** — re-verify against skill-agents-management main `91bf945` (only a model-alias
  resolution change since `944c7b4`, no LaunchMode/Lineup change) and move the pin in the
  document to `91bf945`, stating that in the report.

## Deliverables
One additional signed commit (`git commit -S`; paste the `git log --show-signature -1`
line); board resource `TASK-260905-2ft7ts_rework-report-0013-1.md` with a finding →
disposition table and the commit hash; then `task-board handoff TASK-260905-2ft7ts --role developer`.
Do not push, tag, open a PR, or mark done. Never write LOGBOOK.md or anything into the
control root.
