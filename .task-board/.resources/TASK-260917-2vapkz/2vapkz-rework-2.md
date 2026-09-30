# TASK-260917-2vapkz — rework 2 (THE ONLY CURRENT INSTRUCTION, with 2vapkz-brief.md)

The rev2 re-review (`TASK-260917-2vapkz_review-verdict-rev2.md`) verified everything else: the framing, the carriers, the guard and
registry.md. One blocking finding remains:

tools/generate-vectors/main.go writeSchemaCases now does `must(os.RemoveAll(caseRoot))` over conformance/v1/schema-cases. That DELETES 25
hand-authored released fixtures the generator never produced, such as agent-environment-marker-v1/* (the full list is in the verdict).
1. Remove the RemoveAll, or limit cleanup to paths the generator itself writes, e.g. by tracking the files it emitted in this run.
2. Restore all 25 files byte-identical to 4ad8042b: `git checkout 4ad8042b -- <paths>`. Then check the whole tree:
   `git diff --diff-filter=D --name-only 4ad8042b -- conformance` must print only deletions that the change intends and discloses in
   the results. Ideally it prints nothing.
3. Regenerate, then run `tools/validate.py` and `make regenerate-check` and record the real exit codes.
4. If you think those fixtures should be wired into index.json or retired, write that in the results as a follow-up. Do not do it here.

Set status development, update the results, run `task-board handoff TASK-260917-2vapkz --role developer`, then END YOUR TURN. No
LOGBOOK.md. Never spell any employer name.
