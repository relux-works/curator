# TASK-260917-2vapkz — rework 1 (THE ONLY CURRENT INSTRUCTION, with 2vapkz-brief.md)

Review of rev1 (`TASK-260917-2vapkz_review-verdict-rev1.md`) = CHANGES REQUESTED. The framing itself was independently verified: the
colliding pair, the empty tree and the hash_version mechanism are all fine. Fix exactly these:
1. Released schemas were modified in place. Keep every file that exists at tag v1.0.0-rc.13 byte-identical to base 4ad8042b.
   - log-response-v2.schema.json: revert it. Add `log-response-v3.schema.json`, which refs registry-log-entry-v2. Move the schema-cases
     that need v2 entries to v3, and update the references in registry.md, profiles and registry-service.md.
   - manager-config-v2.schema.json: revert it byte-identically, which also removes the whitespace churn. Add `manager-config-v3` with
     the waiver-state `hash_version`. Move valid-v2-state-hash-waiver, invalid-waiver-hash-version and
     invalid-v2-state-waiver-commit-length (renamed for v3) to the v3 cases. Update the normative text that names the config schema
     version.
2. release/1.0.0-rc.13.json is a published release record. Revert it. If tooling truly requires a change, explain it in the results
   instead.
3. Add a guard to tools/validate.py: fail when a schema file that exists at the latest release tag differs byte-wise from that tag.
   Add a negative test in tools/test_validate.py (a modified released schema → failure) and a positive one.
4. Regenerate the vectors (generator only, no hand edits). Run the validators and the regeneration check, and record the real exit
   codes. Update the CHANGELOG entry to list the new schema versions.

No LOGBOOK.md. Never spell any employer name. Set status development, update the results, run
`task-board handoff TASK-260917-2vapkz --role developer`, then END YOUR TURN.
