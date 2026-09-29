# TASK-260929-2wgjam review verdict — rev2: ACCEPTED

CR-TASK-260929-2wgjam-2, base 64b12189, tree f12475c5 (worktree == candidate tree, `git diff f12475c5` empty).

1. No shared prefix bin: the old `dirname(rustup)` prepend was removed. The additions are now CARGO_HOME/bin (before, only when rustup_dir==cargo_home/bin), the symlink-resolved keg (real_binary_dir != canonical dirname), or a lane-private `mktemp -d $RUNNER_TEMP/rust-lane-bin.*` that links exactly the rustup/rustc/cargo/rustfmt/cargo-fmt/cargo-clippy/clippy-driver proxies. Each addition goes through prepend_lane_path → refuse_shadowing_dir. rustup is invoked by absolute path in all 3 places: install, which, and show.
2. The invariant: go/node `command -v` is captured before any PATH change. refuse_shadowing_dir runs for every added dir, including CARGO_HOME/bin, and verify_lane_resolution runs before the final rustc/cargo checks. Both failure messages name the directory.
3. Selftest, full local run: `bash .github/ci/gate-selftest.sh` gave exit 0 and "292 passed, 0 failed" (6m13s). Rows (a) 240-243, (b) 221-222 and (c) 223-225 are ok. The existing Homebrew keg rows (k6eypp, including the keg mutant) and the fallback rows are ok.
4. Independent mutant, my own reproduction outside the selftest (/tmp fixture: setup-go dir first on PATH, Homebrew-style prefix/bin holding a `go` and a rustup symlink into Cellar keg). I re-inserted the old `dirname(rustup)` prepend after the "using rustup" line:
   - candidate: exit 0. GITHUB_PATH = cargo/bin and the keg real dir; the prefix bin is not listed.
   - mutant: exit 1, "the Rust lane changed go resolution: …/sg/go before, …/pfx/bin/go after; shadowing directory: …/pfx/bin".
   The in-suite mutant row also gives assert exit 1, and the invariant message is asserted too.
5. Windows rev1→rev2: row (a) and the mutant sit inside the pre-existing `MINGW*|MSYS*|CYGWIN*` skip of the Homebrew keg block. That skip only applies on Windows; the rows run on macOS/Linux and I executed them here. Rows (b) and (c) are outside the case and run on every platform. The fallback-order row got a platform-conditional expected middle entry: the keg when `ln -s` made a real symlink, otherwise the lane-private rust-lane-bin dir, and never the prefix. That is an expectation fix, not a skip.
6. docs/self-hosted-runner-setup.md matches the new PATH handling.

Residuals (non-blocking):
- R1: guarded set is go/node only; other non-Rust tools in a keg/lane-private dir aren't checked. By construction, those dirs hold only Rust proxies.
- R2: a failing run may already have appended entries to GITHUB_PATH before verify fails. This is harmless because the lane fails.
