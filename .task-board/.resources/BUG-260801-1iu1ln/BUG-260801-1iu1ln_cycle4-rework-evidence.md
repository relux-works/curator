# BUG-260801-1iu1ln cycle-4 rework evidence

## Provenance
- Dedicated worktree: /Users/iv/Developer/intranet/cocoaskills/.temp/BUG-260801-1iu1ln/worktree
- Branch: task/BUG-260801-1iu1ln-lifecycle-observed-traces
- Exact signed base and merge-base: ba250bfc4dfe104a160eadd5b5f4e340693bf892
- Preserved signed commits: 9362cc8c076a85a49c04c82e76026d6f7473a311, afc385f6cb24e12b9ff3ac83bc6d1036f3ea3eef, and bb2e5801d3f4c31e48018028097b525238126b33
- Signed cycle-4 commit: 963d224c9bb3d6fc274b9dbfbac4bdcafd243c2b, parent bb2e5801d3f4c31e48018028097b525238126b33
- git verify-commit for base and cycle-4 HEAD: exit 0 with the expected ECDSA signer
- Final worktree is clean and HEAD has no tag

## Cycle-4 repairs
- Process observation records every absolute argv element through subprocess run and Popen boundaries across dry-run, GC, status, and currentness paths. Protected GC artifacts invoked through an interpreter are therefore visible.
- GC adoption evidence compares the complete protected-entry tree and records in-place Path.chmod and os.chmod mutation attempts, including repair-and-restore behavior.
- Recovery binds both exact transaction IDs to exact canonical project identities, derives cache identity from recovered state, and derives the triggering project from the exact lock identity. Same-basename wrong owners fail the complete case comparison.
- The systematic adjacent-field audit also removed private-artifact basename inference, repeated persistent-generation literals, project-lock basename projection, and unconditional transaction identifier-order labels.
- Literal lifecycle answers are explicitly fail-closed classified, known lossy proxy forms are rejected, and three new product-seam sabotage probes raise total independent sabotage coverage to ten while preserving the 378-leaf exhaustive vector mutation audit.

## Direct validation evidence
- Cycle-4 helper and three new sabotage probes: exit 0; 7 passed in 122.39s.
- Canonical 32 cases plus exhaustive 378 scalar mutations: exit 0; 410 passed in 41.58s.
- Full authenticated exact-root conformance: exit 0; 845 passed in 495.89s.
- Focused preserved product regressions: exit 0; 3 passed in 3.08s.
- Installer, global install, currentness, and installer transaction suite: exit 0; 131 passed in 124.22s.
- Transaction, GC, and status suite: exit 0; 111 passed and 1 expected platform skip in 13.86s.
- Strict configured mypy: exit 0; no issues in 68 source files.
- compileall src tests: exit 0. No standalone formatter or linter is configured.
- Staged and committed exact-base diff checks: exit 0. Forbidden pyproject, CI, changelog, generated version, schema, release, and claim surface guards: exit 0.
- Isolated signed-tree PEP 517 build: exit 0; produced cocoaskills 0.12.6.dev41+g963d224c9 sdist and wheel.
- Twine check of both distributions: exit 0; PASSED. Sdist lifecycle-observer membership: exit 0.
- Post-build clean tree, signature, exact-base diff, release-surface, merge-base, and no-tag checks: exit 0.

No PR, main landing, tag, release, claim, pin, schema-v7, CI, changelog, or pyproject action or change was made. This signed commit is for later PR19 integration only.