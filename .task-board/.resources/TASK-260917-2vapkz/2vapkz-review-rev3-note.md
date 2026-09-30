# Review note — TASK-260917-2vapkz rev3: delta review (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Rev3 (tree 49a6c3a6, green) against your rev2 verdict. The delta from rev2 (ec64babe) is 29 paths: 26 added (the 25 restored fixtures
plus schema_cases_preservation_test.go), and 3 modified (generate-vectors/main.go removes the RemoveAll line; release/1.0.0-rc.13.json
manifest sha; one more file). The orchestrator found `git diff --diff-filter=D --name-only 4ad8042b 49a6c3a6` empty.
Verify:
1. The 25 restored fixtures are byte-identical to 4ad8042b.
2. The generator no longer deletes anything it did not write. The new preservation test fails if a RemoveAll-style wipe comes back:
   re-add the RemoveAll as a mutant and give the real exit code.
3. The third modified file is expected. The release json change is again only the manifest sha fields.
4. `tools/validate.py` and `make regenerate-check` pass (real exit codes).
Everything else was verified at rev2: carry it forward. accept_cr, or changes requested with file:line. Never spell any employer name.
