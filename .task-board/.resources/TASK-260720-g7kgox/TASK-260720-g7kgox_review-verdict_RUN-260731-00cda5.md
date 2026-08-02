# TASK-260720-g7kgox reviewer verdict

## Verdict

ACCEPTED. No acceptance-blocking findings were found at the exact reviewed head.

## Provenance and review boundary

- Canonical base: c7dbd6daf6562a264275fca06b50a527bce236d4. The canonical clone was clean, main matched origin/main, and both accepted dependency handoffs were present before the task worktree was created.
- Reviewed head: ea64669df0fa58b776ffd67842d40d85c32f4857 on task/TASK-260720-g7kgox-atomic-global-builds. The worktree was clean and the local head matched the remote task branch.
- git merge-base returned the exact base above. The review covers the single signed task commit, 10 changed files, and the reported 1,877 insertions plus 304 deletions.
- git verify-commit reported a good ECDSA signature for oparin@me.com with fingerprint SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM.
- PR 17 remains open, non-draft, mergeable, and points at the exact reviewed head: https://github.com/ivanopcode/cocoaskills/pull/17
- The requested marker-v2 specification head is current remote main at 432eb2ee1fe2d6b271e37269f867c8851c325539: https://github.com/relux-works/curator-spec/commit/432eb2ee1fe2d6b271e37269f867c8851c325539
- The CocoaSkills CI pin and the requested specification head expose the same marker-v2 golden blob, e1df535f8489ce139aa8b84ecae6a246a9491297.

## Acceptance-criteria audit

1. Planning, lock boundary, publication, and marker v2: global install captures a generation, runs shared closure validation and shared build planning, creates private cache-miss builds before the manager-home lock, then recovers, revalidates generation, target preimages, and closure before protected cache publication and materialization under the manager-home lock. See https://github.com/ivanopcode/cocoaskills/blob/ea64669df0fa58b776ffd67842d40d85c32f4857/src/csk/global_install.py#L154-L207 and https://github.com/ivanopcode/cocoaskills/blob/ea64669df0fa58b776ffd67842d40d85c32f4857/src/csk/global_install.py#L311-L406.

2. Shared architecture: global installation calls the existing shared planner, build cache, marker writer, shim activation, private builder, publication helpers, and generic durable transaction commit. The private builder was generalized with explicit operation roots, and both project and global paths call the same transaction-target commit routine. See https://github.com/ivanopcode/cocoaskills/blob/ea64669df0fa58b776ffd67842d40d85c32f4857/src/csk/installer.py#L988-L1075 and https://github.com/ivanopcode/cocoaskills/blob/ea64669df0fa58b776ffd67842d40d85c32f4857/src/csk/installer.py#L1299-L1348. No second global build policy was introduced.

3. Complete transaction boundary: the global target plan includes contexts and marker-only nodes, per-commit runtimes, canonical shims, PATH-visible user-bin forwarders and ledger, env.sh and env.ps1, adapter mirrors and ledgers, plus stale context, runtime, shim, user-bin, and adapter removal. Duplicate target keys fail before mutation. See https://github.com/ivanopcode/cocoaskills/blob/ea64669df0fa58b776ffd67842d40d85c32f4857/src/csk/global_install.py#L419-L565. Staging writes live activation paths while keeping staged bytes private, then commits all desired targets through one durable journal at lines 603-875.

4. Rollback and prior-install preservation: independent source review confirmed reverse rollback is delegated to the shared transaction engine with captured preimages. Focused fault-injection tests cover build failure, cache publication failure, and committed-target failure at target classes 10-context, 20-runtime, 30-shim-canonical, 40-user-bin, 50-env-file, 60-adapter-ledger, and 80-removal. Each test compares the complete watched prior tree after failure. See https://github.com/ivanopcode/cocoaskills/blob/ea64669df0fa58b776ffd67842d40d85c32f4857/tests/test_global_install_transactions.py#L351-L524.

5. Dry-run purity: dry-run substitutes an ExitStack for the operation lock and returns after pure planning. Its focused test replaces ProjectLock, ManagerHomeLock, and BuildLock with constructors that fail if called, then proves manager-home and user-home trees are byte-identical. See https://github.com/ivanopcode/cocoaskills/blob/ea64669df0fa58b776ffd67842d40d85c32f4857/tests/test_global_install_transactions.py#L527-L567.

6. Activation and compatibility: the build integration test verifies provider-first build order while the home lock is absent, marker schema_version 2, excluded build roots, and executable canonical plus user-bin surfaces. Shared Unix tests execute launchers with arguments and require the child nonzero exit code; native Windows tests do the same for cmd wrappers, including ERRORLEVEL handling. Script-only, system-only, dependency, partial-source all-or-nothing, global upgrade option/exit, and stale-removal cases remain in the focused suite. See https://github.com/ivanopcode/cocoaskills/blob/ea64669df0fa58b776ffd67842d40d85c32f4857/tests/test_global_install_transactions.py#L245-L349, https://github.com/ivanopcode/cocoaskills/blob/ea64669df0fa58b776ffd67842d40d85c32f4857/tests/test_build_activation.py#L492-L561, and https://github.com/ivanopcode/cocoaskills/blob/ea64669df0fa58b776ffd67842d40d85c32f4857/tests/test_shims.py#L502-L555.

7. CLI wiring: global install delegates locking to the new global operation; upgrade retains the repository update under GlobalLock, releases it before install, forwards install options, and returns the install failure when present or the update failure otherwise. See https://github.com/ivanopcode/cocoaskills/blob/ea64669df0fa58b776ffd67842d40d85c32f4857/src/csk/cli.py#L603-L649 and https://github.com/ivanopcode/cocoaskills/blob/ea64669df0fa58b776ffd67842d40d85c32f4857/src/csk/cli.py#L867-L888.

## Independent validation

All commands below ran in the clean task worktree at ea64669df0fa58b776ffd67842d40d85c32f4857 with Python bytecode and pytest cache writes disabled where applicable.

- Focused global, transaction, adapter, env, activation, shim, and CLI suite: exit 0; 226 passed and 6 skipped in 82.49 seconds.
- python -m mypy: exit 0; strict type checking passed for 67 source files.
- Ruff over every changed Python source and test file: exit 0; all checks passed.
- git diff --check against c7dbd6daf6562a264275fca06b50a527bce236d4: exit 0.
- Post-gate git status remained clean.

Remote CI run 30671637640 is complete and successful with all 14 jobs green: artifact build, strict mypy, and the full test matrix on Python 3.11 through 3.14 for Ubuntu, macOS, and Windows: https://github.com/ivanopcode/cocoaskills/actions/runs/30671637640

## Conclusion

The implementation matches the task acceptance criteria, fits the established project architecture, and has passing focused, strict typing, lint, artifact, full-suite, and cross-platform evidence. Reviewer verdict: accepted; route TASK-260720-g7kgox to done. The reviewer made no product-code changes.
