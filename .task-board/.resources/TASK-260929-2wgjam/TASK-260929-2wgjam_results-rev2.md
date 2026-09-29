# TASK-260929-2wgjam results (rev2)

## Change
- install-rust-toolchain.sh: never prepends dirname(rustup). Only CARGO_HOME/bin, the symlink-resolved rustup keg, or a lane-private RUNNER_TEMP dir of links to exactly the Rust proxies reach PATH/GITHUB_PATH. rustup is invoked by absolute path. go/node resolution is recorded before any change; each added directory is refused if it holds a different go/node (message names the directory), and a final check re-verifies resolution.
- rev2 fix (rev1 validation failure on windows-latest): the keg-vs-plain decision compared the resolved dir `/c/Users/.../Temp/...` with the textual `/tmp/...` form, so on Git Bash (where ln -s copies) the shared prefix bin was mistaken for a keg and added. Now compares against the canonical (cd -P) rustup dir. The fallback selftest row expects the keg where ln -s makes a real symlink and the lane-private dir otherwise, and never the prefix bin.
- docs/self-hosted-runner-setup.md PATH paragraph updated.

## Evidence
- `bash .github/ci/gate-selftest.sh` (macOS local): exit 0, 292 passed, 0 failed.
- Rows (a), (b), (c) pass. The old dirname(rustup) prepend mutant is killed by row (a): probe exit 1 (asserted), message "the Rust lane changed go resolution".
- Not run locally: the windows-latest gate. The hosted gate is the check there.
