# TASK-260728-wy3dsw manager-profile review cycle 2

Verdict: CHANGES REQUESTED. Route to documentation rework and require another independent reviewer cycle.

## Independent evidence

- Reviewed the isolated worktree read-only at pinned HEAD 57c1f56846d221ecc55786bd3c2467ec32f11730.
- Accepted-predecessor rsync comparison is empty outside profiles/manager.md and cli/curator.md; the Git index is clean.
- Final hashes match rework handoff: manager 4db786194415ac86df0b912b15201bcb24109a50ceb4d8dd0234740ce04b32ba; CLI 6160f08ad4b433aacb772085949c8bb1f3361eddd6ca5cc52179bd5dbb7c1ba6.
- Independent make validate passed: 42 schemas, 400 vectors, 15 Python tests, and Go tool tests. git diff --check passed.
- Git 2.50.1 exact private init and full-OID cat-file batch smoke passed under the documented clean environment. Exact-tag-only fetch created only refs/curator/tag, matched the lock, and left FETCH_HEAD absent. Trace2 showed the SSH transport child tuple exactly as absolute wrapper argv[0], git@example.test, and git-upload-pack /example/repo.git with no probe option.
- All five cycle-1 findings are closed: self-contained config/ref grammar, complete commit/tag grammar and outer tag equality, 33 unique stable diagnostic mappings, disjoint syntax-only offline behavior, and full SSH wrapper tuple.

## Required change

Manager section 11.8 does not carry the accepted signer boundary. It says repository data cannot request signing and section 11.10 names build_repository_signer_policy_unsupported, but nowhere states that the first go-repository-v1 profile performs no manager post-build signing, timestamping, or notarization, or that a platform requiring local signing must reject until a separately reviewed signer profile exists. Architecture v6 section 11 and Decision 0005 require both. As written, an implementation could invoke an operator-selected signer during a real install, or continue despite a platform signing requirement, without contradicting the profile. Add explicit MUST NOT post-sign language and the fail-closed platform transition, mapped to the existing stable diagnostic. Keep release-pipeline signing/notarization outside install-time receipts and publication.

This is bounded documentation rework, not a stop-the-line boundary. No project file was modified during review.