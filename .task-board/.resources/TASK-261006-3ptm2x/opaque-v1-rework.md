# THE ONLY CURRENT INSTRUCTION — TASK-261006-3ptm2x rework 1 (developer, code)
The reviewer (sol high, RUN-261006-a8c933) requested changes. Read the attached `TASK-261006-3ptm2x_review-verdict-rev1.md` IN FULL, including its surface sweep and its attack probes. Fix all three findings:
1. **P1.** Legacy installed-tree currentness still trusts a v1 NUL collision. `internal/marker/marker.go:1347` and the matching `cmd/curator` site rehash at the recorded version with no NUL guard. Every v1 compute or trust site must run the opaque guard BEFORE hashing, and refuse.
2. **P1.** V1 identities are computed before the opaque refusal. In `internal/audit/audit.go`, the hash at :328 precedes the opaqueReport block at :340. Reorder so that NO v1 digest is ever computed over a NUL tree. Do not just refuse after computing.
3. **P2.** V2 audit verdicts use an unversioned carrier and a digest-only cache identity. `storeCachedFindings` (:511) writes a v2 digest into the schema-1 JSON without `hash_version`. Per core.md §8, version a new carrier with `hash_version: 2`, compare (version, digest), and never put a v2 digest in a frozen v1 shape. Keep reading old schema-1 cache entries as v1.
Make the reviewer's attack probes into production-entry regression tests, and name the mutants that kill each one.
Host rules (attached): targeted tests only through `mini-build-lock`; no cmd/curator or install suites run locally.
Add or keep the CHANGELOG entry. No LOGBOOK or remote-gate.sh edits.
Then `task-board handoff TASK-261006-3ptm2x --role developer` and END YOUR TURN.
