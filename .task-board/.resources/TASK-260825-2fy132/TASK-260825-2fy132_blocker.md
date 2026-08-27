## Stop-the-line evidence

The requested operator documentation describes `build_https`, `curator config build-https`, `CURATOR_BUILD_HTTPS_TOKEN`, `CURATOR_BUILD_HTTPS_HOST`, terminal candidate selection, and a host-pinned HTTPS askpass broker.

### Direct production-entry-point evidence

`go run ./cmd/curator config --help` exited 0 and listed only `show` and `build-ssh`.

`go run ./cmd/curator config build-https --help` printed `curator: unknown config subcommand "build-https"`; the program returned exit 2 and the Go launcher returned exit 1. This is an expected-red result proving the command is absent, not a passing validation.

A repository-wide source search found no production `build_https`, `BuildHTTPS`, `build-https`, `CURATOR_BUILD_HTTPS`, or HTTPS credential-resolution implementation. The existing configuration and install entry points only wire `BuildSSH`.

### Consequence

The untracked `docs/build-https.md`, matching README addition, CHANGELOG entry, and existing logbook claims do not match the current shipped command output. Publishing them would document a non-existent interface.

### Validation

`make lint` exited 0 (`golangci-lint run`, 0 issues). No documentation-link checker is defined in the Makefile or CI scripts, so there is no repository link-lint command to run.

### Required resolution

Provide the HTTPS implementation in this worktree (including its command output) and rerun this documentation task against it; or explicitly change the acceptance criterion to allow planned-interface documentation. The recommended path is to land the implementation first, then preserve this page only after its examples are regenerated from the real CLI.
