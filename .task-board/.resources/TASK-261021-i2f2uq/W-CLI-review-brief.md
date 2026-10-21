# Review brief: TASK-261021-i2f2uq, CLI grammar and UX proposal (read-only review)

There is no branch: the deliverable is the attached `TASK-261021-i2f2uq_cli-review.md`, with `_help-evidence.md` and `_results-r1.md`. The producer reviewed curator `cdcba3bd` and launcher `dd70beb4`. Check out both, detached, in your `.val/` (the launcher is your `curator-agent-launcher` clone).

| Surface | Attack |
| --- | --- |
| Evidence | Pick at least eight findings, including F2, F3, F9, F10, F17 and F18. Reproduce each from the cited source or a real help/probe run. Report any finding whose evidence does not hold. |
| Inventory | Spot-check completeness: compare `curator help` and every nested `--help`, and the launcher's parser, against section 1. List missing commands or flags. |
| Grammar table | For each "proposed" row, check that it does not change semantics, that the alias plan is stated, and that "keep" rows are justified. Flag any rename that would break a wire, config or scripted contract. |
| Registration | Section 4 respects "Curator runs no package code at install". Lazy registration does not borrow the permissive PATH trust (F17). Selection and ownership semantics are preserved. |
| Ownership | The top-five slices name the right owning repository and spec. |
| Hygiene | No personal paths or names in the proposal. It is a board resource on a public board. |

Note: the launcher lane has already fixed the `curator run` against `curator-run` spelling in notices (branch `feat/q1-signoff-activation`), so F4 may be partly resolved upstream. Say so if you see it.
