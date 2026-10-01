# TASK-260917-2tx81l — gate fix 1: v2 must be additive under SPEC_PIN rc.13 (THE ONLY CURRENT INSTRUCTION, with 2tx81l-brief.md and 2tx81l-decision-1.md)

Rev1 (tree a28fa9f7) broke 35 existing top-level tests across every lane (run 36735846431). By package: cmd/curator 15, internal/scopes
9, envprofile 3, config 2, environments 2, install 2, contextmaterialize 1, crossconformance 1. Examples:
- TestClassifyDraftMemberPresenceRows/legacy-marker now says content-drift instead of needs-install;
- config TestParseRejections/schema and TestUnknownSchemaVersionsRejected fail;
- contextmaterialize TestSystemModuleAdmissionVectors fails.

Root rule you must follow: curator main pins curator-spec rc.13 (SPEC_PIN), and rc.13 has NO v2 carrier schemas. Until SPEC_PIN
moves to a release containing b1a2efb's v2 shapes:
1. WRITERS keep emitting exactly the rc.13 shapes: v1 hashes and the existing marker/lock/config/audit versions. Existing on-disk
   artifacts must classify exactly as before (legacy marker → needs-install, and so on).
2. READERS accept and validate the v2 shapes, and hashing v2 is available behind the explicit version. Registry matching compares
   versions: a v1 identity never equals a v2 identity. The v2 conformance families you drove stay driven under the b1a2efb suite
   digest.
3. Any write cut-over to v2 goes behind one explicit internal switch that is OFF by default and has a TODO naming the future SPEC_PIN
   bump (rc.14). No test depends on it being on, except rows that set it explicitly.
4. Every one of the 35 failing tests passes again unchanged: do not edit an existing assertion to fit new behaviour. Where a test
   genuinely needs a v2-aware update because reading changed, explain it per test in the results.

Get the full list first: download the run's test-evidence artifacts (`gh run download 36735846431 -p 'test-evidence-*'` from the
worktree). Run the affected packages locally in split -run groups with real exit codes, against rc.13 AND the b1a2efb suite. Keep the
three mutants killed. Set status development, update the results, run `task-board handoff TASK-260917-2tx81l --role developer`, then
END YOUR TURN. No CHANGELOG/LOGBOOK edit. Never spell any employer name.
