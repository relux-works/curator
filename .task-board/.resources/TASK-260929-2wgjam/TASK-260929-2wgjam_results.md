# TASK-260929-2wgjam results — Rust lane never shadows other tools

## Change
- `.github/ci/install-rust-toolchain.sh`: the `dirname(rustup)` prepend is gone. The script adds to PATH/GITHUB_PATH only: CARGO_HOME/bin; the symlink-resolved real dir of rustup (Homebrew keg, k6eypp behaviour kept); or, for a non-symlinked rustup outside CARGO_HOME/bin, a lane-private dir under RUNNER_TEMP with symlinks to exactly rustup/rustc/cargo/rustfmt/cargo-fmt/cargo-clippy/clippy-driver (whichever exist); plus the pinned `rustup which` fallback. rustup is invoked by absolute path everywhere.
- Invariant: `command -v go` / `node` recorded before any PATH change. Every addition goes through `prepend_lane_path` → `refuse_shadowing_dir` (fails "refusing to add <dir> to PATH: it would shadow go ..."); a final `verify_lane_resolution` fails "the Rust lane changed go resolution ... shadowing directory: <dir>".
- `.github/ci/gate-selftest.sh`: rows (a) shared prefix symlinked rustup + fake go, later-step PATH replay keeps original go, prefix never in GITHUB_PATH; (b) CARGO_HOME/bin rustup with go on PATH, GITHUB_PATH exactly CARGO_HOME/bin; (c) CARGO_HOME/bin holding a different go → exit 1 with shadowing message, empty GITHUB_PATH. Old-prepend mutant (awk re-inserts the `dirname(rustup)` prepend) → row (a) probe exit 1 (asserted want 1), caught by the resolution invariant.
- Existing Homebrew/fallback rows updated: GITHUB_PATH is now CARGO_HOME then keg (no prefix bin). Keg-proxy narrowing mutant still kills.
- docs/self-hosted-runner-setup.md PATH paragraph updated.

## Evidence
- `bash .github/ci/gate-selftest.sh` → exit 0, "292 passed, 0 failed" (local macOS; run after the final edits).
- An intermediate run failed 2 pre-existing guards because my new comments mentioned `/opt/homebrew/bin` and a Go release string; comments reworded, rerun green.
- `bash -n` clean on both scripts. shellcheck NOT run: not installed on this host (exit 127).
- Hosted gate (incl. Windows rows skip path, real rose-air runners) not run locally; runner publishes CR and runs the gate.

## Lint (added after first attach)
- shellcheck via docker koalaman/shellcheck:stable: `install-rust-toolchain.sh` with `-e SC2010` → exit 0 (SC2010 `ls | grep` is pre-existing in the diagnostics block, untouched). An SC2154 from my first eval-based draft was fixed (explicit guarded_before_go/node).
- gate-selftest.sh: 5 new SC2016 *notes* (intentional single-quoted literals: probe heredoc/awk mutant/grep patterns), same class the file already carries (17 notes on main); no warnings/errors added.
- Final `bash .github/ci/gate-selftest.sh` after the lint fix → exit 0, 292 passed, 0 failed.
