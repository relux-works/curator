# THE ONLY CURRENT INSTRUCTION — TASK-261016-w0yx2w: finish snapshot publication (developer, round 2)

The owner chose option A for the staging boundary, and curator-spec main 15d2fc2 now specifies it (`protocol/managed-home-snapshots.md`, CIP-0014):
- publication stages beside the absent target, in a private staging directory that this invocation creates with an unpredictable name, and publishes by no-replace rename;
- publication MUST NOT overwrite the target, including a competing entry at the destination name;
- a same-user process that mutates entries inside this invocation's private staging directory is **outside the guarantee**;
- destination no-replace/no-overwrite, no-follow ancestor and target checks, prior-snapshot verification, and cleanup limited to this invocation's staging entries all stay;
- success is emitted only after the published tree verifies against the captured manifest (bytes, sizes, modes, complete entry set, no extras); a mismatch emits no success, leaves the published entry for the operator, and never attempts destructive rollback.

The absent-only destination decision (curator-spec 9dfc557) still stands; do not reopen it.

**Do this:** resume from your retained candidate (checkpoint fcd785b; see `TASK-261016-w0yx2w_staging-constraint.md`) on the converged workspace and implement:
- publication: on macOS `renamex_np` with `RENAME_EXCL`, on Linux `renameat2` with `RENAME_NOREPLACE`; Windows only where the platform gives an equivalent no-replace primitive, otherwise the documented refusal;
- post-publication verification;
- failure and interruption cleanup limited to this invocation's staging entries.

Re-scope the three cleanup-invariant tests to the specified boundary. Assert what the contract promises, and do not try to defend against same-user substitution inside private staging. Add publication/cleanup mutant controls that go red for the intended reason. Run the full gate and record what ran on which OS.

**Hand-off:** signed commit; update the validation outcome with a "Round 2" section (each AC row, the command, its exit code and its bounds); `task-board handoff TASK-261016-w0yx2w --role developer`; end your turn.
