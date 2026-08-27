# TASK-260728-wy3dsw manager-profile review cycle 3

Verdict: ACCEPTED.

## Scope and integrity

Reviewed the isolated worktree read-only at pinned HEAD `57c1f56846d221ecc55786bd3c2467ec32f11730`. The Git index is clean. An accepted-predecessor `rsync -anic --delete` comparison, excluding Git metadata, temporary state, and caches, reports only `profiles/manager.md` and `cli/curator.md`. Final SHA-256 values match the rework-2 handoff: manager `25fe0c397f7cc4770279a2fe465ba7f60c3b3cd995644464b39c35873084b671`; CLI `6160f08ad4b433aacb772085949c8bb1f3361eddd6ca5cc52179bd5dbb7c1ba6`.

## Signer boundary

The cycle-2 finding is fully closed. Manager section 11.8 states that first-release `go-repository-v1` MUST NOT perform manager post-build signing, timestamping, or notarization. A platform or operator local-signing requirement fails before publication with exact tuple `signer-policy` / `unsupported` / `error` / `build_repository_signer_policy_unsupported` until a separately versioned and reviewed signer profile exists. Apple Developer ID signing/notarization and Windows Authenticode signing/timestamping remain operator or release-pipeline work outside install-time receipts, cache identity, manager publication, and this profile. Package-controlled signer identity, argv, entitlement, service, generic hook, or prebuilt provenance remains rejected. This is consistent with Decision 0005 and Protocol Core section 12.2.

## Regression review

All five cycle-1 findings remain closed: self-contained config/files-ref byte grammar and precedence; complete commit/tag/continuation grammar plus exact outer tag-name equality; 33 unique stable diagnostic mappings and specific package-controlled argv/environment/output/credential/signing rejections; disjoint syntax-only offline warning versus install/update/repair/coverage-audit hard failure; and the complete SSH wrapper tuple including absolute `argv[0]`. Exact-tag-only acquisition, raw-object proof, all-blob LFS rejection, audit-before-cache/compiler ordering, receipt-v2/marker-v3 mixed planning, protected identities, transaction/rollback, structural shim/PATH checks, read-only status, repair, deduplication, and GC roots agree with architecture v6 and current Core/Decision text. The CLI examples and exit/output contract do not claim rc.4 implementation support.

## Independent evidence

- `make validate` under the task-local pinned venv passed: 42 schemas, 400 vector files, 15 Python tests, and Go tool tests.
- `git diff --check` and `git diff --cached --exit-code` passed.
- Focused signer, prior-five-finding, lifecycle, offline, audit/cache, transaction, status, repair, and GC assertions passed.
- The stable diagnostic matrix contains 33 rows and 33 unique codes.
- Git 2.50.1 exact private init, exact-tag-only fetch, fixed destination, absent `FETCH_HEAD`, sole ref, and full-OID `cat-file --batch=%(objectname) %(objecttype) %(objectsize)` smoke passed.
- A Git 2.50.1 Trace2 observer probe with `GIT_SSH_VARIANT=ssh` showed the transport child tuple as absolute observer-wrapper `argv[0]`, `git@example.test`, exact `git-upload-pack /example/repo.git`, with no capability probe or extra SSH option; the observer exited 128 intentionally before network.

An initial scratch smoke was rejected before execution by the shell safety layer because its cleanup trap used recursive deletion; the same smoke without deletion passed. An initial multiline text assertion pattern exited 1 because the normative sentence wraps across Markdown lines; the corrected multiline assertion passed. Neither anomaly indicates a product or documentation defect. No project file was modified during review.