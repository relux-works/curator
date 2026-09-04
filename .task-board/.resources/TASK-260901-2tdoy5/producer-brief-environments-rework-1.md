# Producer brief: rework 1 for protocol/environments.md

## Subject
- Worktree: `/Users/iv/Developer/ReluxWorks/.worktrees/curator-spec-environments-normative`, branch `draft/environments-protocol`, head `eddd509`. Edit ONLY `protocol/environments.md`.
- Findings: board resource `review-findings-environments-1.md` on TASK-260901-2tdoy5 (2 major, 2 minor, 2 nit). Read in full; apply ALL six per their suggested fixes unless you can justify strictly better, recording deviations.

Resolution guidance:
- **M1 (opencode.json bytes)**: pin exact bytes — the managed file is exactly `CCJ-1({"instructions":[...]})` per registry.md §1 with a stated trailing-LF rule; keep it inside the §5.6 hash surface; make the §13 vector surface producible.
- **M2 (case-insensitive collision)**: extend the core §2 platform-path collision rule to every materialization write path (module tree, managed homes per profile, backups); a detected platform collision fails before writing with a new stable `environment_*` code; adjust the "collision-free by construction" sentence honestly.
- Minors: add the missing §7.2 diagnostic code; add the §7.6-style pre-freeze verification caveat to the §7.3 flag spellings.
- Nits: per the report.

## Deliverables
1. One signed commit on the branch (`git commit -S`, verify) applying all six.
2. Re-run `make validate` (tree health) — must stay exit 0.
3. Board: add resource `environments-rework-report-1.md` (finding→disposition table + commit hash); handoff task to `to-review`.

## Do not
Touch other files; push; tag; mark done.
