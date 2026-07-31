# TASK-260720-3ag6pi reviewer verdict

Verdict: BLOCKED — genuine external release-candidate precondition.

## Accepted evidence

- Independent pinned-environment `make validate` passed: 35 schemas, 189 vector files, 27 Python tests, and Go tests; no skip additions were found.
- Both retained independent regenerations contain 190 files, have aggregate SHA-256 `41aa37774478c26377455877ee79ef74f8cb5cf5562ea5b1501e5c94fe9c1fa0`, and are recursively byte-identical.
- Independent manifest audit found 189 filesystem entries, 189 manifest entries, zero missing/extra paths, and zero hash mismatches; all new schema cases, 11 build-driver expected files, 13 fixture files, and both vector files are present.
- Legacy agent-skill/csk-skill v1-v5, install-marker-v1, conformance-claim-v1, and their filtered index semantics match origin/main; claim v1 remains schema 1 / rc.3.
- The coverage matrix maps all seven story AC clauses and its 149 case-like references resolve to executable schema/vector names. All 75 rejection cases require artifact_executed=false and reuse=false.

## Stop-the-line evidence

The required landed candidate does not exist. `origin/main`, local main, every protocol feature branch, and the verification worktree still point to rc.3 commit `57c1f56846d221ecc55786bd3c2467ec32f11730`; the complete rc.4 product remains modified/untracked worktree content. An unwrapped `python3 tools/release_gate.py --version 1.0.0-rc.4 --commit HEAD` fails with `release gate requires a clean candidate checkout`. Producer attempt 4 records the same failure.

The reported passing release log used a disposable alternate index plus a PATH-prepended Git wrapper that special-cases `git status --porcelain` to return clean. `release_gate.py` nevertheless resolves and reports release commit HEAD as `57c1f568...`, whose tree does not contain the rc.4 artifacts. This bypasses the candidate-commit invariant and makes the statement `release gate passed ... at 57c1f568...` invalid release evidence. It therefore cannot satisfy either the release-check AC or the prohibition on fabricated release evidence.

## Failed assumptions and alternatives

- Failed assumption: predecessor tasks had landed before integrated verification. They are board-done but their accepted product exists only as uncommitted composite worktrees.
- Failed attempt: alternate-index execution preserves generated-diff checks but cannot turn uncommitted bytes into the candidate commit required by release_gate.py.
- Recommended: an authorized curator-spec integrator must land the accepted rc.4 composite as a real commit/ref, provide that exact commit SHA, and expose a clean checkout at that SHA. Then rerun all four gates without a Git wrapper or alternate clean-status semantics and revise the outcome/matrix claims.
- Alternative: redefine this task as pre-publication conformance only and remove the release-check-pass claim; tradeoff: it cannot certify rc.4 publication readiness.
- Not recommended: authorize a disposable/fake commit solely to satisfy the check, because that would still not be the landed release candidate and would weaken provenance.

Exact external input required: a clean curator-spec commit/ref containing the accepted rc.4 composite, plus its commit SHA and authorization to verify that real candidate. No code was modified by this review.