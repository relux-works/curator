# Producer brief: ax PR #1 revision — rework 1

Worktree `/Users/iv/Developer/ReluxWorks/.temp/ax-curator-integration/worktree`, branch
`draft/curator-environment-integration`, head `2c7f642`. Findings:
`TASK-260905-30a1bz_review-findings-ax-1.md` (2 major, 1 minor, 1 note). Apply ALL three.
New signed commits on top of `2c7f642` — never rewrite. Author decisions:

- **F1 (major)** — option (a): add fixture case `LAUNCH-PLAN-DETERMINISM-NEG` carrying a
  planning-answer / step-4-answer argv pair (members named to match how §13.1 describes the
  two `launch` calls); extend `scripts/validate_launch_plan.py` to compare them and refuse
  with `provider_protocol_error`; add the case to `required_negatives`; add an expected-red
  mutation that admits the mismatch. Keep §13.1's binding sentence true: every row of the
  table is now in the fixture.
- **F2 (major)** — size `LAUNCH-PLAN-EXTENSIONS-NEG` so the persisted canonical extensions
  object (caller extensions + the four Curator keys + `ax.launch-plan-request` with its real
  `request_digest` length) is exactly 65,537 bytes under the gate's own JCS `canonical`; add
  a positive sibling (`LAUNCH-PLAN-EXTENSIONS-POS` or similar) at exactly 65,536 bytes; add an
  expected-red mutation that widens the fixture object by one byte below the bound (a +1
  narrowing of `EXTENSIONS_MAX_BYTES` must now go red). Compute the sizes with a script and
  paste the numbers into the report.
- **F3 (minor)** — one boundary negative per gate constant (129 argv elements, 65,537 total
  argv bytes, 65 `env_names`, a 4,097-byte literal, 65 extension keys) and an unknown-`schema`
  negative (`field: "schema"`), each in `required_negatives`, each with a +1 narrowing that
  goes red (prove it in the report by running the narrowing on a scratch copy with
  `PYTHONDONTWRITEBYTECODE=1` — reviewer note N1).
- Re-run and record exit codes: `./scripts/validate_spec.py`, `./scripts/test_expected_red.sh`,
  `./run_validation.sh`, `git diff --check`. Re-mint only the `SPEC.md` digest if SPEC changed.
- Plain push (no force) so PR #1 updates in place; update the PR body (validation numbers,
  case count); PR stays open and unmerged. Attach `TASK-260905-30a1bz_revision-report-2.md`
  (commits + signatures, finding → disposition, fixture sizes, gate exit codes, new PR head);
  then `task-board handoff TASK-260905-30a1bz --role developer`. Never write into the control
  root.
