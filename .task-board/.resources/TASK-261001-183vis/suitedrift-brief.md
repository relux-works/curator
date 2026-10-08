# THE ONLY CURRENT INSTRUCTION — TASK-261001-183vis (read-only diagnosis, researcher)

## Symptom
Every curator landing spawn since 2026-09-30 20:22Z ends with:
`validation_suite_changed: runner integrate refused: validation suite or environment drift is not an exact reviewed change of its configured source`.

| run | task | outcome |
|---|---|---|
| RUN-260930-a223cf | TASK-260930-1kylpg | last SUCCESS (18:05Z) |
| RUN-260930-dec728 | TASK-260930-o5uu7a | integration_runner_failed (20:22Z) |
| RUN-260930-96f5cd, -bf7687, -b10079 | TASK-260930-o5uu7a | suite_changed |
| RUN-261001-61ec44 | TASK-260728-rjxrgs | suite_changed |
| RUN-261001-7fed72 | TASK-261001-1klixs | suite_changed (CR validated 03:04Z, landed 03:46Z) |

RUN-261001-e841c2 (TASK-260917-2tx81l) is landing now; record its outcome too.

## Facts already checked
- The orchestration config `.temp/orchestration/task-board.config.json` is byte-identical to the copy saved at 02:47Z, which is before 1klixs was validated. Only its mtime changed (03:12Z).
- Every landing run, the successful one included, ran on muse muse-spark-1.3-contributor/xhigh.
- All spawn-runner processes use `~/.curator/cache/build/go-v1/f32104b4…/bin/task-board`. The wrapper `~/.local/bin/task-board` execs `~/.curator/global/bin/task-board`; it was installed ~2026-09-30 21:22Z local.
- tb-sessiond for this board has been running since 2026-09-29 from build 59b546aa. Do NOT restart it.

## Find
1. In the task-board source, find exactly what makes up the "validation suite / environment" identity, and how integrate compares it. The source is the skill-project-management repo under ~/Developer/ReluxWorks, or the build source behind the f32104b4 cache entry.
2. Compute both identities for TASK-261001-1klixs: the one recorded with the CR validation, and the one integrate sees now. Name the component that differs.
3. Explain why 1kylpg passed and the later runs do not. Candidates:
   - the binary upgrade;
   - daemon-vs-runner version skew;
   - env vars (e.g. a `.zshenv` change);
   - trunk-side workflow files (`.github/workflows/ci.yml` changed in 38fjt0);
   - something else.
4. Give the smallest legitimate fix:
   - a host step for the human (e.g. a session restart);
   - a reviewed config/source change;
   - or the correct task-board command.
   Never suggest bypassing validation.

## Rules
- Read-only:
  - no landings, no `worktree` mutations, no config edits;
  - no `make install`, `setup.sh`, `self-update` or daemon restarts.
- Never put secrets in the results.
- Never spell any employer name.

## Handoff
1. Write the findings (evidence with file:line and the identity values) into the task results resource.
2. Run `task-board handoff TASK-261001-183vis --role researcher`.
3. END TURN.
