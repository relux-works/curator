# TASK-260929-2wgjam — Rust lane must not shadow Go (THE ONLY CURRENT INSTRUCTION)

## Problem (main red on Test (rose-air))

Two self-hosted runners now carry the rose-air label. On the second one ("macbook-iv"), rustup is at `/opt/homebrew/bin/rustup`.
`.github/ci/install-rust-toolchain.sh` (around line 140) prepends `dirname(rustup)` = `/opt/homebrew/bin` to PATH and GITHUB_PATH.
That directory also holds Homebrew `go` (go1.27.1), which then shadows the go1.25.5 installed by setup-go. `toolchain-identity.sh`
fails with "expected go1.25.5 (from go.mod), got go1.27.1". Evidence: runs 36566058066 and 36558891906, job "Test (rose-air)". A green
run on the first runner (rustup in ~/.cargo/bin) is 36527630366.

## Required change

1. Never prepend a shared prefix bin directory, such as the one holding a symlinked rustup, to PATH or GITHUB_PATH. Expose only
   directories that contain the Rust proxies:
   - CARGO_HOME/bin;
   - the symlink-resolved real directory of rustup (the Homebrew keg, e.g. /opt/homebrew/Cellar/rustup/<v>/bin), which the script
     already computes for the rustc/cargo proxies;
   - or, as a fallback, a lane-private directory holding symlinks to exactly rustup, rustc, cargo (and rustfmt/clippy if used), created
     under RUNNER_TEMP.

   Invoke rustup by its absolute path.
2. Invariant, enforced by the script itself: record `command -v go` (and node, if present) before any PATH change. After the script's
   PATH changes, the same paths must resolve; otherwise fail with a clear message naming the shadowing directory.
3. Add rows to `.github/ci/gate-selftest.sh`, in its existing style, with fake bins in a temp dir and a fake GITHUB_PATH file:
   - (a) rustup symlinked from a shared bin dir that also contains a fake `go` → the script succeeds and `go` still resolves to the
     original; GITHUB_PATH does not contain the shared dir;
   - (b) rustup in CARGO_HOME/bin → unchanged behaviour;
   - (c) a directory the script would add that contains a different `go` → the script refuses with the shadowing message.
4. Mutant: restore the old `dirname(rustup)` prepend → row (a) fails. Give real exit codes.
5. Keep the k6eypp behaviour (keg proxy dir) and the existing rows passing. Update docs/self-hosted-runner-setup.md only if it describes
   the PATH handling.

No CHANGELOG/LOGBOOK edit. Never spell any employer name. Update the results, then run `task-board handoff TASK-260929-2wgjam --role developer`, then
END YOUR TURN. The runner publishes the CR and runs the gate.
