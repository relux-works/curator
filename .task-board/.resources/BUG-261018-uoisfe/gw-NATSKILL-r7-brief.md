# THE ONLY CURRENT INSTRUCTION — BUG-261018-uoisfe: rework round 7 of W-NATSKILL (developer)

The deciding review of CR revision 6 requested changes with one finding. Read the attached `BUG-261018-uoisfe_review-verdict-rev6.md`, the probe `BUG-261018-uoisfe_review-probe-rev6.go` and its evidence.

**Finding (row 12, bypass).** `cmd/curator/command_help.go:152-154` admits every `env resolve` takeover argument and leaves activation validation to `cmdEnvResolve`; `cmd/curator/main.go:315` returns successful help before `cmdEnvResolve` (`env.go:141`) can enforce `--repair`. Result: `env resolve --takeover --help`, with `--repair` missing, `--repair=false` or `--repair=0`, exits 0 instead of 2.

**Do this.**
1. Enforce the takeover admission rule before any help short-circuit: `--takeover` is accepted only by `profile install --use`, `profile use`, `profile sync`, `profile update` and `env resolve --repair` (true), whether or not help is requested. Every other combination exits 2 with the existing diagnostic and writes nothing. Prefer one admission check shared by the help path and the command path over a second copy.
2. Keep the revision-1 list-wrapper regression. Add the reviewer's help witness as a permanent test (missing, `=false`, `=0` → 2; no-consent help controls → 0; `env resolve --repair --takeover --help` → 0), plus a narrowing mutant that re-opens the help short-circuit and fails the witness for that reason.
3. Sweep the help cross-product: every command and flag form that reaches help (`-h`, `--help`, `help <cmd>`) combined with `--takeover`, against the five admitted operations. Record the table in the gaps file.
4. Stay inside the original brief `gw-NATSKILL-brief.md`. rc.7 wave rule: do not edit `CHANGELOG.md` or the root `README.md`. Do not add checklist items about the post-turn validation: it runs after your handoff and is the gate itself.

**Hand-off:**
1. Update `BUG-261018-uoisfe_gaps.md` with a "Rework round 7" section: the finding, the fix with file and line, the tests, the sweep table.
2. Run `task-board handoff BUG-261018-uoisfe --role developer`.
3. END YOUR TURN.
