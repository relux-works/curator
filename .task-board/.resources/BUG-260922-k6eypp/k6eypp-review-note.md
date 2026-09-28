# Review note — BUG-260922-k6eypp rose-air rustc PATH with Homebrew rustup (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Review rev3 (base 86552087, tree ea127a2e, 3 paths: .github/ci/install-rust-toolchain.sh, .github/ci/gate-selftest.sh,
docs/self-hosted-runner-setup.md; hosted gate green on every lane) against `k6eypp-fix-1.md`, `k6eypp-gatefix-1.md`, `k6eypp-gatefix-2.md`
and the bug's AC 2 (probe widening allowed with docs + gate-selftest row). Main has been red only on Test (rose-air) since ~09-24 with
`rust-pin: rustc is not on PATH after installing Rust 1.91.0` while rustup is Homebrew's /opt/homebrew/bin/rustup (proxies in the keg).
Verify: (1) after `rustup toolchain install`, when rustc/cargo are missing the script adds the rustup PROXY directory (keg bin resolved from
the real path of rustup, or the pinned toolchain bin via `rustup which`) to PATH and $GITHUB_PATH, re-checks, and still fails closed; the
pinned channel is unchanged (no floating); other lanes (CARGO_HOME rustup, Windows) behave exactly as before. (2) gate-selftest: the
Homebrew rows run on linux/macOS and are an explicit skip on Windows Git Bash; a mutant (keg step removed) fails the selftest on
linux/macOS — run `sh .github/ci/gate-selftest.sh` locally (macOS) with a real exit code, and the mutant. (3) docs describe the Homebrew
case. No other paths, no CHANGELOG/LOGBOOK. accept_cr or changes requested with file:line. No LOGBOOK.md.
