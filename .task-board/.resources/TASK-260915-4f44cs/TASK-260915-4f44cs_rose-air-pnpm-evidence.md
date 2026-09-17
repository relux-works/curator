# Rose-air evidence after the pnpm pin (TASK-260916-5aqozl landed on main 7c3ce2f)

Main push run 35121791685 (https://github.com/relux-works/curator/actions/runs/35121791685), job Test (rose-air) on macbook-iv, artifact test-evidence-rose-air/test/go-test-served.json:

- internal/pnpmsource TestRealPinnedPNPMTargetPrunedUnreachableRejectsBeforeInstall: ok
- internal/pnpmsource TestRealPinnedPNPMLockSupersetSnapshotDependencies: ok
- internal/pnpmsource TestRealPinnedPNPMPrivateStoreAndOfflineMaterialization: ok

The broken PATH pnpm shim no longer affects the lane (the pinned pnpm is used). The lane still fails on exactly two internal/rustsource cases because rustc is absent on the runner (authority_external_test.go:229 → t.Fatal); tracked as the sibling task rose-air-lane-rust-toolchain. platform-case gate: ok. The blocking item recorded on this task (rose-air red because of pnpm) is resolved; the lane itself stays red until the Rust toolchain is provided.
