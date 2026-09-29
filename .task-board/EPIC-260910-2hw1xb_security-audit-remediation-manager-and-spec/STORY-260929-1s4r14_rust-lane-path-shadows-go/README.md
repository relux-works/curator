# STORY-260929-1s4r14: rust-lane-path-shadows-go

## Description
Main Test (rose-air) red on the second rose-air runner (macbook-iv): install-rust-toolchain.sh prepends dirname(rustup)=/opt/homebrew/bin to GITHUB_PATH, which also exposes Homebrew go1.27.1 and shadows setup-go go1.25.5; toolchain-identity.sh then fails (runs 36566058066, 36558891906).

## Scope
(define story scope)

## Acceptance Criteria
(define acceptance criteria)
