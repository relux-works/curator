# How to produce test evidence without running tests on the mini (applies to N6–N9, read before handoff)
Checklist items that need test results ("Relevant tests written … and passing", the regression red/green item, the negative-controls item) are satisfied by a HOSTED run, never a local one (R223):
1. From a disposable clone (`git clone --shared` of the control root into $TMPDIR is fine), commit your exact candidate on a branch `scratch/<task-id>-green` and push it to origin; GitHub CI runs the full test matrix. Wait for it (`gh run watch <id> --exit-status`).
2. Commit the same candidate with ONLY the production fix reverted on `scratch/<task-id>-red` and push it; the regression test must fail there (record the failing test name from the run log / test-evidence artifact).
3. Put both run URLs and the outcome in your results resource, then tick the items with that evidence. Delete both scratch branches after recording (`git push origin --delete …`).
This is the sanctioned way; do not leave items unticked and do not tick them without the hosted evidence.

Note (orchestrator, 00:04Z): the workspace was converged onto main after N6 landed; your CHANGELOG.md line was reverted to avoid a conflict and is attached as N7-changelog.patch — re-add it under Unreleased.
