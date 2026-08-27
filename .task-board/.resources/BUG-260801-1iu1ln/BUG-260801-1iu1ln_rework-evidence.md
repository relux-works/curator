# BUG-260801-1iu1ln rework evidence

## Provenance
- Worktree: /Users/iv/Developer/intranet/cocoaskills/.temp/BUG-260801-1iu1ln/worktree
- Branch: task/BUG-260801-1iu1ln-lifecycle-observed-traces
- Exact signed base and merge-base: ba250bfc4dfe104a160eadd5b5f4e340693bf892
- Preserved producer commit: 9362cc8c076a85a49c04c82e76026d6f7473a311
- Signed rework commit: afc385f6cb24e12b9ff3ac83bc6d1036f3ea3eef
- git verify-commit for base and rework commit: exit 0, good ECDSA signature for oparin@me.com
- Final worktree status: clean

## Reviewer false-negative repairs
- Planning now fails each of 11 real CocoaSkills gate seams independently; omitting _validate_skills changes the observed case.
- Private-build failure records ManagerHomeLock acquisition across the whole operation and classifies all eight forbidden effects by traps plus persistent-surface witnesses.
- Repair records ten distinct pipeline seams for every one of five conditions; audit_plans must run, so omitting gate_plans changes the observed case.
- Recovery injects a real planning-generation change and requires a second full private-build/recovery pass; omitting _assert_generation_current changes the result.
- Dry-run, GC, status, cross-project, recovery, private-build, and repair taxonomies now use per-field state, trace, negative-control, or concurrent-lock evidence. LOGBOOK wording was corrected.

## Direct gates
- Authenticated exact-root full conformance: exit 0; 835 passed in 247.85s. Includes 32 canonical lifecycle cases, 378 scalar mutations, and 4 product-seam sabotage tests.
- Canonical lifecycle subset: exit 0; 32 passed.
- Scalar mutation subset: exit 0; 378 passed.
- Explicit sabotage subset: exit 0; 4 passed.
- Preserved focused product regressions: exit 0; 3 passed.
- Installer/global/currentness suites: exit 0; 131 passed.
- Transaction/GC/status suites: exit 0; 111 passed, 1 platform skip.
- Strict project mypy: exit 0; no issues in 68 source files.
- compileall src tests: exit 0. No standalone formatter or linter is configured.
- git diff --check exact base..HEAD: exit 0.
- Pin/CI/pyproject/release-surface guards: exit 0. No pin, schema-v7, tag, release, claim, CI, or pyproject change.
- Clean signed-tree python -m build: exit 0; sdist and wheel 0.12.6.dev39+gafc385f6c.
- twine check for both distributions: exit 0; PASSED.
- Sdist lifecycle-helper membership check: exit 0.

This commit is for later PR19 integration only; no PR, main landing, tag, claim, or release action was performed.
