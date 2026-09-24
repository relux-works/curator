# TASK-260916-1h82gq rework 1 (orchestrator, binding)

Revision 1 gate FAILED only on `Test (windows-latest)` (run 35627712332; Linux/macOS lanes green).
30 failures, four causes — all Windows-specific:

1. `internal/scriptworker` (26 rows): every session refuses with
   `script_execution_worker_identity_invalid: the resolved interpreter python3-v1 has multiple
   filesystem links or is a reparse point` on the stub interpreter fixture. R1 rev4 passed these
   identity rows on windows-latest, so R2 introduced it: find what now makes the stub `.exe` look
   multi-linked on NTFS (a staging/launcher step that hard-links or re-opens it? the link-count
   read on Windows — `GetFileInformationByHandle.NumberOfLinks` — taken from a handle opened
   without `FILE_SHARE_*`/on a copy? a fixture that now `os.Link`s the interpreter?) and fix the
   real cause; do not relax the check. Rows like `TestBuiltCuratorWorkerHandshake`,
   `TestWorkerWaitsForPermit`, `TestScriptWorkerHappyPath` ("worker sent failure, want ready") and
   `TestParentWithholdsPermitOnProofMismatch` (forge worker never recorded shutdown, 10 s) are
   the same root cause downstream — re-check them after the fix, and make the forge-worker row
   fail fast with the worker's failure detail instead of a 10 s wait.
2. `internal/runtimestore :: TestStageEnforcedShimTransitionStagesManagerCopies` — "staged launcher
   mode = -rw-rw-rw-, want owner-only executable": Windows has no POSIX exec bits; assert the
   platform's notion (on Windows: the staged launcher is a `.exe` regular file, ACL/attributes as
   the go-v1 worker staging rows do), keep the POSIX assertion on POSIX.
3. `internal/install :: TestEnforcedToDeclaredOnlyFlipRemovesNativeLauncher` — `duplicate staged
   target 80-removal/shim/flip-skill-tool`: on Windows the shim carries the `.cmd` suffix while the
   native launcher does not, so the removal set stages two targets under one name (or the same
   target twice). Fix the staging (distinct target identities for shim and launcher, or a single
   removal per path) and add a Windows row that pins it; check the Windows shim/launcher naming
   everywhere the launcher is staged.
4. `internal/snapshot :: TestConcurrentGetAcceptsOneImmutablePublisher` — Windows "file is being
   used by another process" under concurrent publishers. Determine whether R2 touched
   `internal/snapshot` or its callers; if not (likely a pre-existing Windows flake), say so in
   results.md with the evidence (main-run history) and do not change it — the orchestrator files
   it separately; if R2 changed the publisher path, fix it.
Also `TestEnforcedScriptCommandInstallsNativeLauncherBehindSeam` (cause 1 through the launcher).
Continue from the revision-1 tree (no checkout/clean/stash); append "Revision 2" to results.md
(causes, fixes, Windows proof status); republish only when the gate is green. Budget: this is a
bounded rework — no new scope; keep the brief's honest-partial list as it is.
