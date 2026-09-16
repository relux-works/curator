# TASK-260916-5aqozl — rev2 gate failure detail (Windows) and scope ruling

Rev2 fixed the output-root problem (TestRealPinnedPNPMTargetPrunedUnreachableRejectsBeforeInstall now passes on Windows). Two cases still fail on windows-latest (run 35098955988, artifact test-evidence-windows-latest):

- TestRealPinnedPNPMLockSupersetSnapshotDependencies — conformance_test.go:883: closure_input_undeclared: pnpm writable store registry contains an undeclared member fields=map[entry:c8805e97311ac3a64fd7c507d26b47dc]
- TestRealPinnedPNPMPrivateStoreAndOfflineMaterialization — conformance_test.go:1136: same shape (entry:4990f83e93c63b5b58405c5ac8c8ca6a); the launch is node.exe … pnpm.cjs install --frozen-lockfile --offline --ignore-scripts --store-dir ../pnpm-install-store …

This is a real Windows gap in the product's writable-store closure model (pnpm writes a store member on Windows that the registry does not declare), not a CI problem. Orchestrator ruling: it is OUT of scope for the pin task and is filed as BUG-260916-2f3xbf. For this task: declare exactly these two cases Windows-deferred in the platform-case ledger (.github/ci/platform-cases.tsv / skip-classes.tsv) with the reason naming BUG-260916-2f3xbf, using the repository's existing deferral mechanism (grep the ledger for an existing windows deferral row and copy its shape); keep the skip narrow (only these two tests, only GOOS=windows), keep the third case running, do NOT weaken the closure check or the product. Record the deferral and the bug id in results.md. Then hand off rev3.
