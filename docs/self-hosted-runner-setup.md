# Self-hosted runner setup (rose-air)

The `test-self-hosted` lane (`Test (rose-air)` in
[ci.yml](../.github/workflows/ci.yml)) runs the default gate on the
organisation's self-hosted macOS ARM64 runner (labels `self-hosted, macOS,
ARM64`). It runs on `main` pushes only, and only while the repository
variable `ROSE_AIR_RUNNER` is `true`.

## One-time prerequisites on the runner

- **rustup.** Install once for the runner user (the installer at
  https://rustup.rs), then verify `rustup --version` runs. The lane installs
  the pinned toolchain itself from [rust-toolchain.toml](../rust-toolchain.toml)
  on every run and never relies on a toolchain the runner happens to ship,
  so no toolchain needs installing or pinning by hand. A lane on a runner
  without rustup fails in the `Install pinned Rust toolchain via rustup`
  step with this note named.

Nothing else is installed by hand: Go and Node come from the
`actions/setup-go` / `actions/setup-node` steps, pnpm is installed per lane
into a lane-local prefix, and the Rust toolchain comes from rustup as above.

## After setup

No maintenance: toolchain versions are pinned in the repository
(`rust-toolchain.toml`, `PNPM_PIN`) and each lane verifies its pin against
the Go constant (`rust-pin-guard.sh`, `pnpm-pin-guard.sh`) before anything
installs.
