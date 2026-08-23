# TASK-260729-3nx97g: regenerate-rc5-build-driver-goldens

## Description
Carry the independently accepted schema-6 build-driver golden suite forward into the exact accepted rc.5 TASK-260728-2kp3tv candidate under execution_policy=manager-worker-v1.

## Scope
Work in a new isolated curator-spec task worktree. Reuse the accepted TASK-260720-1s1vr6 generator, fixture and expected-artifact semantics while taking the exact TASK-260728-2kp3tv rc.5 snapshot as the candidate base. Regenerate conformance/v1/vectors/build-drivers.json and conformance/v1/expected/build-driver/, include them in the deterministic manifest, update implementation-neutral generator/tests only as required by rc.5 canonical identity, and make Curator candidate metadata-artifact tests execute rather than skip. Preserve schema-1 through schema-6 declaration bytes and all unrelated rc.5 bytes. No Curator/CocoaSkills product edits, stage, commit, tag, publication, pin advancement or release claim.

## Acceptance Criteria
The rc.5 candidate publishes a complete deterministic build-driver vector and expected-artifact suite; the positive portable input requires execution_policy=manager-worker-v1 and independently recomputes cache key sha256:529370122ae11e2e961d5265b1a020e046bcd43165b2eb96b05e73a51187ac9b and receipt sha256:919fbbad8e6ce95532219fd952c2309d0d7026f85209650508fd6834af4020cd; legacy rc.4 and reserved-hardened inputs are explicit schema-invalid non-alias negatives; every prior positive, rejection and byte-edge cluster remains represented or is explicitly superseded; two clean regenerations are byte-identical; validate, Python/Go generator tests, release-candidate checks and Curator TestCandidateBuildMetadataArtifacts pass without skip against the explicit candidate root; all unrelated accepted rc.5 and frozen legacy bytes remain unchanged; evidence is candidate-only and authorizes no landing or publication.
