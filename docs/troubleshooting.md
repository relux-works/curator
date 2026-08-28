# Curator Troubleshooting Guide

This guide provides symptom, cause, and remedy entries for common Curator diagnostic codes and failures. Every error string and code is verified against `internal/` and `cmd/curator/` source files.

## Compiled-command diagnostics and status codes

`curator status` reports machine-readable state codes and cause subcodes.

### unusable-build-toolchain

Symptom: `curator status` reports status code `unusable-build-toolchain`.

Cause: Curator cannot resolve or verify a trusted Go toolchain environment. Source location: `cmd/curator/builds.go:61`.

Remedy: set `CURATOR_GO` to an absolute path pointing to `<GOROOT>/bin/go` (or `bin/go.exe` on Windows), or export `GOROOT`.

Verify the Go executable path:

```bash
export CURATOR_GO="/usr/local/go/bin/go"
curator status
```

The command re-reads toolchain configuration and checks compiled states.

### build-input-drift

Symptom: `curator status` reports status code `build-input-drift` with cause `build-root`, `target`, or `unattributed`.

Cause: recorded logical key differs from the key derived from current build inputs. Source location: `cmd/curator/builds.go:55`.

Remedy: run `curator install` or `curator upgrade` to rebuild artifacts for current inputs.

Reconcile build input drift:

```bash
curator install
```

The command rebuilds affected binaries and updates install markers.

### build-command-drift

Symptom: `curator status` reports status code `build-command-drift`.

Cause: recorded compiled command set differs from the command set activated by current closure. Source location: `cmd/curator/builds.go:45`.

Remedy: run `curator install` to synchronize installed shims with current closure definitions.

Reconcile command drift:

```bash
curator install
```

The command generates required executable shims in `.agents/bin/`.

### build-source-drift

Symptom: `curator status` reports status code `build-source-drift`.

Cause: recorded build-source identity differs from the raw source snapshot. Source location: `cmd/curator/builds.go:50`.

Remedy: run `curator upgrade` to fetch updated source snapshots and update dependency hashes.

Upgrade project dependencies:

```bash
curator upgrade .
```

The command fetches updated source state and updates `Skillfile.lock`.

### missing-build-artifact

Symptom: `curator status` reports status code `missing-build-artifact`.

Cause: protected build cache holds no entry corresponding to the recorded cache key. Source location: `cmd/curator/builds.go:64`.

Remedy: run `curator install` to recompile missing artifacts into protected cache storage.

Recompile missing artifacts:

```bash
curator install
```

The command builds binaries and populates protected cache entries.

### corrupt-build-receipt

Symptom: `curator status` reports status code `corrupt-build-receipt`.

Cause: canonical receipt identity in protected cache differs from recorded receipt state. Source location: `cmd/curator/builds.go:67`.

Remedy: run `curator install` to quarantine corrupt cache entries and generate valid receipts.

Rebuild corrupt entry:

```bash
curator install
```

The command replaces corrupt receipt data with newly generated build receipts.

### build-artifact-drift

Symptom: `curator status` reports status code `build-artifact-drift`.

Cause: stored artifact binary path or SHA-256 digest differs from recorded marker values. Source location: `cmd/curator/builds.go:70`.

Remedy: run `curator install` to recompile drifted binaries into protected cache.

Rebuild drifted artifact:

```bash
curator install
```

The command replaces drifted binary files with verified build output.

### corrupt-build-cache

Symptom: `curator status` reports status code `corrupt-build-cache`.

Cause: protected build cache entry is unreadable or malformed. Source location: `cmd/curator/builds.go:72`.

Remedy: run `curator install` to quarantine corrupt directories and rebuild artifacts.

Recover corrupt build cache:

```bash
curator install
```

The command quarantines damaged cache folders and creates fresh build outputs.

### untrusted-build-cache

Symptom: `curator status` reports status code `untrusted-build-cache`.

Cause: candidate cache files reside outside manager-protected boundary permissions. Source location: `cmd/curator/builds.go:75`.

Remedy: run `curator install` to recompile binaries into manager-owned protected paths.

Rebuild untrusted cache storage:

```bash
curator install
```

The command moves artifacts into manager-protected directories.

### unsupported-build-platform

Symptom: `curator status` reports status code `unsupported-build-platform`.

Cause: current operating system or file system cannot enforce protected cache state. Source location: `cmd/curator/builds.go:77`.

Remedy: execute Curator on a platform supporting POSIX permissions or Windows ACL protections.

Inspect platform capabilities:

```bash
curator status --json
```

The output indicates whether protected store permissions are enforceable.

### build-context-exposed

Symptom: `curator status` reports status code `build-context-exposed`.

Cause: build root directory reached agent-facing context directory. Source location: `cmd/curator/builds.go:47`.

Remedy: inspect skill manifest layout and ensure build roots are excluded from runtime context paths.

Recheck skill manifest layout:

```bash
curator skill check ./my-skill
```

The command validates package structure and highlights exposed build roots.

### unsupported-build-driver

Symptom: `curator status` reports status code `unsupported-build-driver`.

Cause: skill manifest specifies a build driver outside supported `go-v1` specifications. Source location: `cmd/curator/builds.go:57`.

Remedy: update skill manifest to specify supported `go-v1` driver definitions.

Validate skill manifest specification:

```bash
curator skill check ./my-skill
```

The command flags invalid driver selections.

### build-state-changed

Symptom: `curator status` reports status code `build-state-changed`.

Cause: install marker or protected cache state moved while status classification was executing. Source location: `cmd/curator/builds.go:80`.

Remedy: re-run `curator status` without concurrent modification processes.

Rerun status check:

```bash
curator status
```

The command evaluates static state and returns consistent verdicts.

## Toolchain preflight mismatches

Toolchain checks verify Go compilers and language closure adapters before compilation starts.

### untrusted_go_executable

Symptom: error `CURATOR_GO must name an absolute GOROOT/bin/go` (or `GOROOT/bin/go.exe` on Windows; formatted via `GOROOT/bin/%s`) (code: `unusable-build-toolchain`).

Cause: `CURATOR_GO` environment variable or derived `GOROOT` points to a non-absolute or unadmitted binary path. Source location: `internal/godriver/session.go:489`.

Remedy: set `CURATOR_GO` to an absolute path pointing to a regular `GOROOT/bin/go` binary.

Set path to verified Go toolchain:

```bash
export CURATOR_GO="/usr/local/go/bin/go"
curator status
```

The command validates the toolchain executable and proceeds with planning.

### toolchain_executable_mismatch

Symptom: error `go-v1 toolchain_executable_mismatch: selected Go executable is not the regular executable under the derived GOROOT; put the real GOROOT/bin first on PATH, e.g. PATH="$(go env GOROOT)/bin:$PATH"`.

Cause: selected Go executable on `PATH` is a shim or environment wrapper (such as goenv, asdf, or mise) rather than the standard binary under `GOROOT/bin`. Source location: `internal/godriver/session.go:521` and `cmd/curator/toolchain_remedy_test.go:51`.

Remedy: prepend the actual `GOROOT/bin` directory to `PATH`.

Set PATH to standard Go toolchain binary:

```bash
export PATH="$(go env GOROOT)/bin:$PATH"
curator status
```

The command verifies the regular Go binary location and resumes execution.

### Untrusted or non-operator-pinned Git executable

Symptom: error `trusted Git version probe failed` or `Git release family is not operator-pinned`.

Cause: system Git binary fails execution probes (`git --version` error or output > 256 bytes) or its release family is not operator-pinned. Source location: `internal/buildrepo/admission.go:203` and `internal/buildrepo/admission.go:211`.

Remedy: install an operator-pinned Git release family and ensure `git` on `PATH` passes admission.

Verify system Git version and path:

```bash
git --version
```

The shell prints the active Git binary version string.

### Language source-closure adapter preflight checks

Symptom: source acquisition or resolution fails for Rust, SwiftPM, npm, pnpm, Yarn Classic, or Yarn Modern skill projects.

Cause: missing system toolchain binary or missing source closure lockfiles. Source location: `internal/closureexec/acquisition.go:549`.

Remedy: install required ecosystem package managers (`cargo`, `swift`, `npm`, `pnpm`, `yarn`) on host `PATH`.

Verify ecosystem toolchain availability:

```bash
swift --version
```

The shell prints the version of the installed language toolchain.

## External repository fetch and credential failures

Fetch operations enforce strict operator-owned credential scope rules.

### build_repository_ssh_credential_missing

Symptom: error code `build_repository_ssh_credential_missing`.

Cause: SSH external build repository has no configured operator SSH identity or agent socket. Source location: `internal/buildrepo/credentials.go:13`.

Remedy: configure SSH credential scope using `curator config build-ssh add`.

Configure SSH credentials for host scope:

```bash
curator config build-ssh add git.example.com/portals --identity ~/.ssh/id_ed25519
```

The command records identity path mapping for the specified repository scope.

### SSH identity or host key verification failure

Symptom: errors `SSH identity is unavailable`, `SSH agent socket is unavailable`, or `SSH known hosts is unavailable`.

Cause: specified SSH identity file, agent socket, or known_hosts file is missing, unreadable, or not an admitted file mode. Source location: `internal/buildrepo/credentials.go:63` and `internal/buildrepo/credentials.go:73`.

Remedy: verify SSH identity file path, existence, and file permissions.

Check SSH key file permissions:

```bash
chmod 600 ~/.ssh/id_ed25519
```

The shell updates file permissions to ensure private key security.

### HTTPS credential host mismatch

Symptom: error `HTTPS credential host does not match protected source`.

Cause: configured HTTPS credentials host does not match target repository canonical host identity. Source location: `internal/buildrepo/admission.go:332`.

Remedy: update HTTPS credential scope to match repository host name exactly.

Configure HTTPS token scope:

```bash
curator config build-https add git.example.com/portals --token-env GITHUB_TOKEN
```

The command registers token resolution for matching repository URLs.

### HTTPS credential broker materialization failure

Symptom: error `HTTPS requires a manager credential broker` or `cannot materialize HTTPS credential broker`.

Cause: Curator cannot write or execute host-pinned askpass broker binary in temporary execution root. Source location: `internal/buildrepo/admission.go:259` and `internal/buildrepo/admission.go:337`.

Remedy: ensure temporary directory permissions allow executable file creation.

Check temporary execution environment:

```bash
curator install --verbose
```

The command prints detailed execution diagnostic messages.
