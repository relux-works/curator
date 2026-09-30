# TASK-260917-2vapkz review verdict rev3 — ACCEPTED

Candidate tree 49a6c3a6 (base 4ad8042b). Delta review against rev2 (ec64babe); rev1/rev2 verified areas carried forward (framing, carriers, byte-freeze guard, registry.md, NUL-opaque rule, CHANGELOG).

Rev2 blocking finding (RemoveAll wiping hand-authored schema-cases): RESOLVED.

1. 25 restored fixtures: all 25 blob OIDs equal 4ad8042b (scripted rev-parse comparison, 25/25 identical).
   `git diff --diff-filter=D --name-only 4ad8042b 49a6c3a6` -> 0 lines.
2. main.go: only change vs rev2 is removal of `must(os.RemoveAll(caseRoot))`.
   Preservation test: `go test -run 'Preserv|SchemaCase'` exit 0 on candidate.
   Mutant (RemoveAll re-inserted after caseRoot assignment): exit 1, schema_cases_preservation_test.go:59
   "regeneration deleted hand-authored fixture agent-environment-marker-v1/valid-no-composition.json". Mutant reverted; tree == candidate.
3. Third modified file = conformance/v1/manifest.json: adds the restored fixtures' path/sha entries — expected.
   release/1.0.0-rc.13.json delta vs rev2: only candidate_protocol_pin.manifest_sha256 and downstream_consumption.required_manifest_sha256 (convention per #113).
4. `.venv/bin/python tools/validate.py` exit 0 ("validated 72 schemas and 1253 vector files"). (System python3 lacks jsonschema -> exit 1, environment only.)
   `make regenerate-check` exit 0; worktree unchanged vs 49a6c3a6 afterwards.

Follow-up (non-blocking, not in scope): decide whether the 25 hand-authored agent-environment-marker-v1 etc. fixtures should be indexed in schema-cases/index.json or retired.
