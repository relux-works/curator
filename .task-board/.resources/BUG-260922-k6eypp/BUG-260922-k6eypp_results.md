# BUG-260922-k6eypp results — Homebrew rustup proxies

## Revision 3: POSIX Homebrew fixture and Windows Git Bash skip

- Updated `.github/ci/install-rust-toolchain.sh` to keep installing the channel from `rust-toolchain.toml`, resolve the selected rustup executable through symlinks, and add its keg bin directory to `PATH` and `GITHUB_PATH` when compiler proxies are there. If they are not, the installer asks `rustup which --toolchain "$channel" rustc` for the pinned toolchain bin directory. The final rustc/cargo checks still fail closed.
- Updated `docs/self-hosted-runner-setup.md` with the Homebrew linked rustup and keg proxy layout.
- Updated `.github/ci/gate-selftest.sh` with a Homebrew fixture where rustup is linked from prefix bin and compiler proxies exist only in the versioned keg. It verifies the invoked rustc/cargo and `GITHUB_PATH`; a narrowing mutant without the keg PATH step fails at the pinned rustc check.
- The Homebrew fixture and its mutant print an explicit skip on Windows Git Bash because that host does not guarantee real symlink support. The separate pinned `rustup which` fallback case remains enabled.
- The channel remains pinned and read from the repository toolchain file.

## Fresh local verification

Run on the x86_64 Darwin work host:

- `bash -n .github/ci/install-rust-toolchain.sh .github/ci/gate-selftest.sh` — exit 0.
- `bash .github/ci/gate-selftest.sh` — exit 0; 265 passed, 0 failed. The Homebrew layout positive case and narrowing mutant both ran.
- `make lint` — exit 0; golangci-lint reported 0 issues.
- `make build` — exit 0.
- `git diff --check` — exit 0.

## Findings and bounds

- The earlier hosted Windows self-test failed because Git Bash did not provide a real symlink for the Homebrew fixture. The fixture and mutant are now skipped together on Windows; this Windows branch was not run locally and remains subject to the hosted candidate gate.
- The rose-air ARM64 lane was not run on this x86_64 host. The main-push lane can prove the runner fix only after the change lands; this handoff addresses the repository-side path-widening acceptance criterion.
- No CHANGELOG or LOGBOOK edit was made. This results resource records the implementation, the earlier Windows fixture failure, and the local evidence.
