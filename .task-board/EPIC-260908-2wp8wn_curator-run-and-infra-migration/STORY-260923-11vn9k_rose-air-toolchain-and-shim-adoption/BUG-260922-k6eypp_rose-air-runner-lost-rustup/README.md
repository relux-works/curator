# BUG-260922-k6eypp: rose-air-runner-lost-rustup

## Description
Since 2026-09-21T22:30Z every main push fails the Test (rose-air) lane at step 'Install pinned Rust toolchain via rustup' with 'rust-pin: rustup is not installed on this runner; install it once per docs/self-hosted-runner-setup.md, then re-run this lane' (runs 35663049586 head 09b25ef6 and 35725359745 head 48da2690). The two preceding main runs were green (35659762491 ccb0c6c0 21:53Z, 35607737219 5708554d 13:46Z), so the ARM64 runner lost rustup between 21:53Z and 22:30Z on 2026-09-21 — this is a runner-host regression, not a repository one: the probe order (PATH, CARGO_HOME/bin, ~/.cargo/bin, HOMEBREW_PREFIX/bin, /opt/homebrew/bin, /usr/local/bin) was already fixed by BUG-260918-18ny6p and BUG-260918-3u6qqq and is unchanged. The orchestration host e11-1 is x86_64 and is NOT the ARM64 runner, so the campaign cannot repair it: this needs the operator on the rose-air machine (install rustup per docs/self-hosted-runner-setup.md, or unset the repository variable ROSE_AIR_RUNNER while the machine is unavailable). The lane runs on main pushes only and does not gate pull requests, so landings are unaffected; main's own CI badge stays red until it is fixed.

## Scope
Runner host (operator action) plus, if the machine stays unavailable, the ROSE_AIR_RUNNER repository variable. No repository code change is implied: the lane's rustup probe is already correct.

## Acceptance Criteria
1) a main push runs Test (rose-air) green, or the lane is explicitly disabled with ROSE_AIR_RUNNER=false and that decision is recorded; 2) if any probe path needed widening, docs/self-hosted-runner-setup.md and .github/ci/install-rust-toolchain.sh are updated together with a gate-selftest row
