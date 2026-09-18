# Decision packet — rose-air Rust cases need the closed Apple developer registry (macbook-iv)

## Constraint
internal/rustsource (closed Cargo target registry, aarch64-apple-darwin only) proves the darwin toolchain from the native Apple developer registry: `registeredDarwinSDK()` resolves `/var/db/xcode_select_link` and requires `<developer>/Platforms/MacOSX.platform/Developer/SDKs/MacOSX.sdk` (plus `usr/bin/cc` as the linker). That layout exists only in a full Xcode.app developer directory; Command Line Tools (`/Library/Developer/CommandLineTools`) has `SDKs/` but no `Platforms/` tree, so the registry is reported unavailable and every rust build case refuses with `rust_target_unsupported`.

## Evidence
- main run 35350749331 (commit 1c464c5, 2026-09-18 13:4xZ), job "Test (rose-air)": step "Install pinned Rust toolchain via rustup" passed (Rust 1.91.0 via the Homebrew rustup, PR #78); step "go test + platform-case gate" failed: 15 test failures, all `rust_target_unsupported: closed native Apple developer registry is unavailable` (internal/rustsource TestRustConformance*, TestBuildToolchainRegistrationStartsNoProcessBeforeC0; internal/crossconformance TestCrossAdapterConformance/project-every-path/rust).
- Hosted macos-latest passes the same cases because the image ships Xcode with xcode-select pointing at it.
- platform-cases.tsv declares the production rust cases as running on darwin (host absence tolerated elsewhere only), so the rose-air lane (darwin/arm64) must run them or declare a host-capability skip.

## Failed assumptions / attempts
- Assumed rustup alone completes the runner setup (TASK-260916-vygif0 docs): true for the toolchain, not for the Apple SDK registry.
- No orchestrator-side workaround: installing Xcode is a human action, and the campaign boundary "Xcode only with explicit consent" applies.

## Alternatives
1. Install full Xcode on macbook-iv, run `sudo xcode-select -s /Applications/Xcode.app/Contents/Developer` and accept the license; record it in docs/self-hosted-runner-setup.md (one-time manual step like rustup). Lane then exercises the rust cases on real Apple Silicon — the intended purpose of rose-air.
2. Declare the rust cases host-capability-skipped on rose-air (platform-cases.tsv / a declared skip reason when the registry is absent): keeps the lane green without Xcode, but the ARM64 rust evidence stays "unverified" and the closed-registry design loses its only self-hosted proof.
3. Relax `registeredDarwinSDK` to accept the Command Line Tools SDK layout: product change to the closed registry decision (previously accepted as Xcode-only); needs a spec/product ruling, not a lane fix.

## Recommendation
Alternative 1 (install Xcode on macbook-iv, one-time, documented), because the rose-air lane exists to prove the darwin/arm64 rust path; alternative 2 only if the operator decides rose-air should not carry rust evidence.

## Decision needed from the operator
Which alternative; if 1, the operator installs Xcode + xcode-select on macbook-iv (and tells the orchestrator), after which the next main push proves the lane and this bug closes with that run as evidence; if 2 or 3, the orchestrator files the corresponding code task.
