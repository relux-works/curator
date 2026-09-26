# Self-hosted runner setup (rose-air)

The `test-self-hosted` lane (`Test (rose-air)` in
[ci.yml](../.github/workflows/ci.yml)) runs the default gate on the
organisation's self-hosted macOS ARM64 runner (labels `self-hosted, macOS,
ARM64`). It runs on `main` pushes only, and only while the repository
variable `ROSE_AIR_RUNNER` is `true`.

## One-time prerequisites on the runner

- **rustup.** Install once for the runner user, either with the installer at
  https://rustup.rs using its `--no-modify-path` form (verify
  `~/.cargo/bin/rustup --version` runs) or as the Homebrew formula
  (`brew install rustup`; verify `/opt/homebrew/bin/rustup --version`
  runs). An existing install of either kind is fine as is. The lane installs
  the pinned toolchain itself from [rust-toolchain.toml](../rust-toolchain.toml)
  on every run and never relies on a toolchain the runner happens to ship,
  so no toolchain needs installing or pinning by hand. A lane on a runner
  without rustup fails in the `Install pinned Rust toolchain via rustup`
  step with this note named.

The runner service starts from launchd with a minimal PATH
(`/usr/bin:/bin:/usr/sbin:/sbin`, no `/opt/homebrew/bin`) and reads no
shell profiles, so `rustup` is not expected on the service PATH. The lane
locates it itself: it probes `PATH`, then `~/.cargo/bin` of the runner user
(or `$CARGO_HOME/bin`), then the Homebrew prefixes (`$HOMEBREW_PREFIX/bin`,
`/opt/homebrew/bin`, `/usr/local/bin`), and prepends the directory that
holds `rustup` (plus `~/.cargo/bin` for the toolchain proxies) to `PATH`
before any `rustup` call. A Homebrew `rustup` keeps its binary under
`/opt/homebrew/bin` and only the proxies under `~/.cargo/bin`, which is why
the Homebrew prefixes are probed. Nothing on the runner needs changing for
this; as an alternative, the runner's `.path` file (in the runner
directory, read by the service at start) may carry `/opt/homebrew/bin`,
but the lane does not depend on it.

When no probed location holds `rustup`, the step fails with one
diagnostics block before the remedy note: the runner name (`RUNNER_NAME`),
`hostname`, `whoami`, `HOME`, `CARGO_HOME` (set or defaulted),
`HOMEBREW_PREFIX`, the `PATH` it searched, one
executable/exists-not-executable/absent line per probed candidate, a
bounded rust/cargo listing of the searched bin directories, and
`command -v` / `type -a` for `rustup`. The next red run is therefore
self-diagnosing: the block tells whether the job landed on a different
machine, a different service user, or a `rustup` outside the probed
paths.

Nothing else is installed by hand: Go and Node come from the
`actions/setup-go` / `actions/setup-node` steps, pnpm is installed per lane
into a lane-local prefix, and the Rust toolchain comes from rustup as above.

## After setup

No maintenance: toolchain versions are pinned in the repository
(`rust-toolchain.toml`, `PNPM_PIN`) and each lane verifies its pin against
the Go constant (`rust-pin-guard.sh`, `pnpm-pin-guard.sh`) before anything
installs.
