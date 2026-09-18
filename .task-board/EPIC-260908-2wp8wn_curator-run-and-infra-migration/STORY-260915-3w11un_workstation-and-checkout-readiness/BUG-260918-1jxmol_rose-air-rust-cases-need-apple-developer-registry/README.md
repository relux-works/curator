# BUG-260918-1jxmol: rose-air-rust-cases-need-apple-developer-registry

## Description
After the rustup fixes (PR #77, PR #78) the rose-air lane installs Rust 1.91.0 but main run 35350749331 (1c464c5) fails internal/rustsource and internal/crossconformance with rust_target_unsupported: closed native Apple developer registry is unavailable. internal/rustsource/build_toolchain.go registeredDarwinSDK resolves /var/db/xcode_select_link and requires <developer>/Platforms/MacOSX.platform/Developer/SDKs/MacOSX.sdk — a full Xcode.app developer directory; Command Line Tools ship SDKs under /Library/Developer/CommandLineTools/SDKs without the Platforms tree. The self-hosted runner macbook-iv has no such registry. Human-only decision.

## Scope
(define bug scope / affected area)

## Acceptance Criteria
(define fix acceptance criteria)
