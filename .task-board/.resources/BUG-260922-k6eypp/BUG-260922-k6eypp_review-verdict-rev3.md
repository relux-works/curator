# BUG-260922-k6eypp review verdict — rev3 ACCEPTED
Reviewer: claude-opus-5-5 (low). Candidate tree ea127a2e, base 86552087; worktree `git diff ea127a2e` empty (reviewed exact tree). 3 paths only; no CHANGELOG/LOGBOOK.

## Verified
1. install-rust-toolchain.sh:147-208 — after `rustup toolchain install "$channel"` (channel still from rust-toolchain.toml, no floating), when rustc/cargo missing: resolves real path of rustup (portable readlink loop, 40-hop cap) and prepends the keg dir to PATH+GITHUB_PATH only if it holds rustc/cargo; else `rustup which --toolchain "$channel" rustc` dirname; then the unchanged fail-closed `command -v rustc/cargo || fail`. CARGO_HOME/Windows lanes: the block is a no-op when tools are already on PATH.
2. gate-selftest.sh (macOS, bash via `sh`): `sh .github/ci/gate-selftest.sh` rc=0, 265 passed/0 failed; Homebrew keg rows + in-selftest keg mutant + rustup-which fallback rows all ok. Windows: Homebrew block wrapped in `case $(uname -s) MINGW*|MSYS*|CYGWIN*) skip …`. Hosted gate green on all lanes (per review note).
3. Independent production mutant (deleted `prepend_lane_path "$rustup_proxy_dir"` in the real script): selftest rc=1, FAIL rows "installer finds rustc and cargo from a Homebrew rustup keg", keg-proxy resolution/GITHUB_PATH rows. Restored; tree re-verified equal to ea127a2e.
4. docs/self-hosted-runner-setup.md describes the Homebrew keg case and the rustup-which fallback.

## Residuals (non-blocking)
- R1: the rustup-which fallback row is not in the POSIX-only block (runs on Windows where `ln -s` copies); it passes there because the copy holds no proxies and the fallback fires — green on hosted Windows.
- R2: the real fix is proven only by a green Test (rose-air) run after landing on main (AC 1).
