# BUG-260922-k6eypp — rose-air: rustc not on PATH with Homebrew rustup (THE ONLY CURRENT INSTRUCTION)

Read `campaign-producer-rules.md`. Every main push since ≥ 2026-09-24 is red ONLY on Test (rose-air) (latest run 36346053973 head 86552087):
  rust-pin: using rustup at /opt/homebrew/bin/rustup
  info: syncing channel updates for 1.91.0-aarch64-apple-darwin … unchanged
  info: self-update is disabled for this build of rustup
  rust-pin: rustc is not on PATH after installing Rust 1.91.0
Diagnosis: the runner now has the Homebrew `rustup` formula. Homebrew links only `rustup` into /opt/homebrew/bin; the rustc/cargo proxies
live in the keg (`$(brew --prefix rustup)/bin`, e.g. /opt/homebrew/opt/rustup/bin) and are not linked, so `.github/ci/install-rust-toolchain.sh`
(which prepends only dirname(rustup_bin)) finds no rustc. AC 2 of this bug allows widening the probe path: update
`.github/ci/install-rust-toolchain.sh` so that after `rustup toolchain install`, when rustc/cargo are not on PATH, it adds the directory that
holds the rustup PROXIES (the keg bin next to a Homebrew rustup — resolve via the real path of rustup_bin, or `rustup which --toolchain
"$channel" rustc` dirname as the pinned-toolchain fallback) to PATH and $GITHUB_PATH, then re-checks and still fails closed if absent; keep
the pinned channel (no floating). Update docs/self-hosted-runner-setup.md (Homebrew rustup note) and add a gate-selftest row
(.github/ci/gate-selftest.sh) that simulates a Homebrew layout (rustup linked in bin, proxies only in the keg) and proves the lane finds
rustc; a mutant (keg-dir step removed) must fail the selftest. Other lanes unchanged.
Set status development (the bug was blocked on an operator; this is now a repository fix under AC 2). Update the results resource, handoff,
END YOUR TURN (the runner publishes and gates). The rose-air lane itself proves the fix only after landing on main. No CHANGELOG/LOGBOOK edit.
