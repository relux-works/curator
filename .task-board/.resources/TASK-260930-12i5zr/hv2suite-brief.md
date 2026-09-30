# TASK-260930-12i5zr — accept the content-hash-v2 candidate suite (lockstep for curator-spec PR #116) (THE ONLY CURRENT INSTRUCTION)

Background: curator-spec PR #116 (TASK-260917-2vapkz, content-hash framing v2, accepted) has a failing "Implementations" job. The job
runs pinned curator 80fd617f against the PR's conformance suite:
`marker/install-marker-v4/schema-cases publishes 28 cases, want pinned count 27` (TestReadAuthoritativeMarkerV4SchemaCases). The PR
adds these to already-released families:
- install-marker-v4/invalid-hash-version-on-frozen-marker.json;
- context-lock-v1/invalid-hash-version-on-frozen-lock.json;
- agent-environment-marker-v2/invalid-hash-version-on-frozen-marker.json.

It also adds new families and vectors: install-marker-v5, context-lock-v2, agent-environment-marker-v3, audit-record-v2,
registry-log-entry-v2, registry-bundle-v2, log-response-v3, manager-config-v3, the skillfile-sources v2/v6 files and
content-hashes-v2.json.

The repo practice (the "lockstep" used for PRs #88 and #97): a curator commit on MAIN passes against BOTH the current SPEC_PIN (rc.13,
which curator CI uses) and the candidate suite. The spec PR then pins that commit in .github/workflows/implementations.yml.

Do exactly this, and do NOT implement v2 hashing (that is TASK-260917-2tx81l, later):
1. Check out the candidate suite: a disposable clone of relux-works/curator-spec at 526a9aa0 (branch spec-2vapkz-content-hash-v2).
2. Run curator's conformance consumers with CURATOR_CONFORMANCE_ROOT pointed at the candidate's conformance/v1:
   - `go test ./internal/marker ./internal/contextlock ./internal/envmarker ./internal/registry ./internal/manifest ./internal/skillspec ./internal/crossconformance`;
   - any package that reads schema-cases.

   List every failure.
3. Make each one pass under BOTH roots without weakening a ratchet:
   - the three frozen-shape negative cases must be DRIVEN, i.e. rejected. Curator's closed shapes probably reject an unknown
     hash_version already; prove it;
   - the case-count ratchet must accept both suites exactly. Key the pinned counts by suite identity, e.g. the conformance manifest
     digest or the protocol release label, the way the closed set {rc.9, rc.13} handled the release label in TASK-260926-4hd81z. Do
     not replace an exact count with a lower bound;
   - new families that curator does not implement yet become known-gap rows in .github/ci/conformance-gaps.tsv, owned by
     TASK-260917-2tx81l, and only when present in the root. Nothing silently skips.
4. Prove it with real exit codes: the same package set green with CURATOR_CONFORMANCE_ROOT=rc.13 (SPEC_PIN) and with the candidate.
   Mutant: bump one pinned count → the ratchet fails under the matching root.
5. No CHANGELOG/LOGBOOK edit. Never spell any employer name.

Update the results with the per-family table (rc.13 vs candidate: driven / known-gap / count), then run
`task-board handoff TASK-260930-12i5zr --role developer`, then END YOUR TURN.
