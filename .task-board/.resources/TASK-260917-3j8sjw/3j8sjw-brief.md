# TASK-260917-3j8sjw — registry service: content-hash framing version (THE ONLY CURRENT INSTRUCTION)

curator-skill-registry repository. Read this task's README and STORY-260917-hbuawd's README. The normative text is curator-spec main
b1a2efb (PR #116): registry.md equal-version matching, and registry-log-entry-v2 / registry-bundle-v2 / log-response-v3 carrying
`hash_version`.
1. Records and snapshots carry the content-hash framing version, using the spec's v2 shapes. Existing records stay v1 (frozen shapes
   mean version 1).
2. Matching and publication refuse a version mismatch: a v1 record never matches a v2 identity, and the reverse. Publication of a
   record whose declared version disagrees with its hash form is refused with a stable error.
3. Log/bundle/log-response serialization emits the v2 shapes when a v2 record is present, and bundle import validates the version.
   Keep the P4 high-water and R-series hardening intact.
4. Service tests for v1 and v2 records: match, mismatch refusal, publication refusal, and import of mixed bundles. Where the spec's
   registry vectors apply to the service, drive them from a disposable clone of curator-spec at b1a2efb.
5. Mutant: version not compared in matching → a test fails. Give real exit codes.
6. Deployment note: a docs/ entry or README section on how existing deployments behave (v1 records keep working; v2 records appear
   when curator publishes v2). Add a CHANGELOG entry if this repo keeps one under Unreleased. Never spell any employer name.

Run the repo's test suite (pytest or its Makefile target) and record the real exit codes. Update the results, then run
`task-board handoff TASK-260917-3j8sjw --role developer`, then END YOUR TURN.
