# TASK-260929-2wgjam: rust-lane-never-shadow-other-tools

## Description
install-rust-toolchain.sh must expose rustup/rustc/cargo without changing which go (or any non-Rust tool) resolves on the lane.

## Scope
(define task scope)

## Acceptance Criteria
1. Only Rust proxy dirs are added to PATH/GITHUB_PATH, never a shared prefix bin. 2. Script fails if go/node resolution changes. 3. Selftest rows a-c. 4. Old-prepend mutant killed. 5. Existing rows pass.
