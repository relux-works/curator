# BUG-260918-1jxmol — closure evidence (orchestrator, 2026-09-21)

Decision packet option (1) was taken by the operator: full Xcode installed on macbook-iv (rose-air),
`sudo xcode-select -s /Applications/Xcode.app/Contents/Developer` and `xcodebuild -license accept`
done (operator, 2026-09-21). Verification: main run 35569760474 (head 6cb3f738), job `Test (rose-air)`
106296620784 re-run after the switch → conclusion **success**; the previously failing required
cases `internal/rustsource :: TestProductionManagerCapturesRegistryFromRawPaths` and
`TestProductionManagerCapturesGitWithoutCallerProjection` (and the 12 rust_target_unsupported
"closed native Apple developer registry is unavailable" failures + crossconformance/rust) now pass:
4 rustsource cases ok; platform-case gate ok.
Previous evidence of the failure: runs 35495615639 (before and after the Xcode install but before
xcode-select) with the same 12 failures. No code change was needed (registeredDarwinSDK requires the
full Xcode developer dir through /var/db/xcode_select_link).
