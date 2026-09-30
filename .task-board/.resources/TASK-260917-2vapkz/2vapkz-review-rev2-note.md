# Review note — TASK-260917-2vapkz rev2 re-review (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Re-review rev2 (curator-spec, base 4ad8042b, tree ec64babe; validators green) against your rev1 verdict and `2vapkz-rework-1.md`.
The orchestrator has already checked two things:
- log-response-v2.schema.json and manager-config-v2.schema.json are byte-identical to base, and log-response-v3 and manager-config-v3
  were added;
- release/1.0.0-rc.13.json still changes, but only `candidate_protocol_pin.manifest_sha256` and
  `downstream_consumption.required_manifest_sha256`. The previous post-release PR #113 (4ad8042, TASK-260928-q5100t) changed the same
  two fields, so this looks like the repo's convention, where the candidate pin tracks the conformance manifest. Confirm it from
  RELEASE.md / tools/validate.py; if confirmed, it is not a defect.

Verify:
1. The rev1 blocking findings are resolved. References in registry.md, profiles, registry-service.md and the normative text now name
   log-response-v3 and manager-config-v3 wherever v2 entries or the waiver hash_version are needed. The moved schema-cases are in the
   right directories.
2. The new validate.py byte-freeze guard compares every schema present at the latest release tag and fails on any byte difference.
   Its negative and positive tests exist, and the guard does NOT block the legitimate manifest_sha256 update above. Run
   `python3 tools/test_validate.py -k <guard>` or the equivalent, with the real exit code.
3. Now finish the areas you left at "first look":
   - registry.md: equal-version match, and a version mismatch is a non-match with a vector;
   - the interim NUL-opaque rule for v1 readers;
   - vectors are generator-driven (the regeneration check is green);
   - hash_version carrier completeness (markers, context locks, verdict/pin state, registry records/bundles/log entries,
     skillfile-sources files);
   - CHANGELOG lists the new schema versions.
accept_cr, or changes requested with file:line. Never spell any employer name.
