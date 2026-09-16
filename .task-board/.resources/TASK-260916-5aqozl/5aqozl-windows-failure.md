# TASK-260916-5aqozl — rev1 gate failure detail (Windows)

Hosted run 35091265197: only Test (windows-latest) failed. From the artifact test-evidence-windows-latest/test/go-test-served.json, exactly three tests fail, all in internal/pnpmsource, all with the same message:

- TestRealPinnedPNPMTargetPrunedUnreachableRejectsBeforeInstall — conformance_test.go:854: closure_input_undeclared: portable output root is not a private real directory
- TestRealPinnedPNPMLockSupersetSnapshotDependencies — conformance_test.go:876: same
- TestRealPinnedPNPMPrivateStoreAndOfflineMaterialization — conformance_test.go:1129: same

They ran for the first time on Windows because the pin made real pnpm available (before, they skipped). Every other package and the platform-case gate are green; ubuntu and macOS green.

Direction: this is a Windows path-identity problem in the fixture's output root (the hosted runner's temp dir is reached through a junction / short 8.3 name / different case, so the "private real directory" check does not recognise it), not a pnpm problem. Fix in the TEST fixture first (create the output root under a resolved real path — filepath.EvalSymlinks / os.MkdirTemp on a resolved base — the same way other Windows-green fixtures in this repo do; grep for EvalSymlinks in *_test.go). Change product code only if you prove the check itself is wrong on Windows (state the proof in results.md). Do NOT add a skip or widen a skip class. Re-run the three tests locally only if you have a Windows host; otherwise rely on the gate at handoff. Keep the rest of rev1 unchanged.
