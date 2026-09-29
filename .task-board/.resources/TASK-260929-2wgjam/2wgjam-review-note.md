# Review note — TASK-260929-2wgjam Rust lane must not shadow Go (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Review rev2 (base 64b12189, tree f12475c5, 3 paths, gate green on every lane; rev1 failed only a Windows selftest ordering row) against
`rustpath-brief.md`. Verify:
1. install-rust-toolchain.sh never adds a shared prefix bin directory (such as the dirname of a symlinked rustup under /opt/homebrew/bin)
   to PATH or GITHUB_PATH. It adds only CARGO_HOME/bin, the symlink-resolved rustup keg dir, or a lane-private shim dir. rustup is
   invoked by absolute path.
2. The invariant is enforced: the resolution of `command -v go` (and node, if present) is captured before and checked after. If it
   changes, the script fails and names the shadowing directory.
3. Selftest rows a–c exist, and the existing rows still pass. Re-run the mutant yourself (restore the old `dirname(rustup)` prepend): row
   (a) must fail. Give real exit codes. Check what rev1→rev2 changed for Windows, and confirm the fix is not a skip that hides the row on
   every platform.
4. The k6eypp keg-proxy behaviour is kept. docs/self-hosted-runner-setup.md is consistent, if it was touched.
accept_cr, or changes requested with file:line. No LOGBOOK.md. Never spell any employer name.
