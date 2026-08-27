# Superseding candidate e66cb72 — fully green (corrected attribution)

> **SUPERSEDED.** `e66cb72` was fully green but was superseded by
> `6001dc33281b94a4ec7442ab15278550dd0f51d9`, which bands
> `profiles/manager.md` §11 on marker v4 (review cycle 2 item 3.1, decided by
> fixing). Current evidence: `TASK-260822-c0rxj7_6001dc3-green-matrix.md` and
> `TASK-260822-c0rxj7_results.md`. Kept unrewritten as history.

- Identity: e66cb72d9988c614c7232af9195bf829c82d328e, protocol 1.0.0-rc.9, manifest sha256 803918bf..., tree 9d5a10b6..., 692 files (conformance-root digests unchanged from edd0721 — the marker-v4 delta is protocol prose outside the conformance root).
- Marker-v4 delta content (CORRECTED per review cycle 2 finding 2.2): core.md section 10 obligations extended — managers supporting schema 8 MUST read marker schemas 1-4 and MUST write marker v4 for schema-8 installation mutations, plus marker-v4 semantics prose, plus conformance/README.md updates. The profiles/manager.md +5/-1 hunk in the edd0721..e66cb72 range is UNRELATED to marker v4: it is the independently merged build_ssh credential-scopes prose (517a130); the earlier claim of a matching manager.md marker surface was wrong — recorded here as corrected.
- Candidate-conformance run 32654422338 for candidate_ref e66cb72: Candidate suite SUCCESS on ubuntu, macos, windows; the single non-candidate Test (windows-latest) failure re-ran green (full workflow conclusion success), consistent with the proven managerlock flake pattern.
- Supersedes edd0721 (own 3-OS green run 32651139699); nothing rewritten.