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

## Skillfile source and transport diagnostics

These stable classes describe schema-2 project source resolution and
installation. Each class names the selector or member and the reason
without secrets, and the CLI appends a sanitized
remediation. Fetch failures report one closed-vocabulary clause per
attempted endpoint (`availability`, `auth`, `tls`, `host-key`,
`ref-moved`, `identity`, `integrity`, `audit`, `http-404`,
`policy-unreadable`, `unknown`); raw tool output and full URLs never
appear in user-facing text.

### source_alias_unknown

Symptom: `source_alias_unknown: <alias>`.

Cause: a `from` selector names an alias with no `sources` entry.

Remedy: declare the alias under `sources` in Skillfile.json, or fix
the `from` spelling, then run `curator project resolve`.

### source_selection_invalid

Symptom: `source_selection_invalid` with the offending selector.

Cause: the selector breaks a structural rule: unknown fields, a mixed
form, a non-contained directory, a bad collection (`**`, partial
globs, files as members), a missing or doubled ref, or a transitive
branch.

Remedy: fix the named selector (directory, include/exclude, and ref
rules in docs/cli.md), then retry the explicit attempt.

### source_member_missing

Symptom: `source_member_missing` with the member name.

Cause: an explicit `include` literal has no such directory, or a
requirement names a skill outside the lock.

Remedy: add the named member directory with valid SKILL.md, or drop it
from `include`, then run `curator project resolve`.

### source_member_invalid

Symptom: `source_member_invalid` with the member name.

Cause: the package fails validation: SKILL.md frontmatter without the
required name or description, a manifest identity mismatch, a
non-directory member, or an identity the lane cannot prove.

Remedy: fix the named package (valid SKILL.md frontmatter and manifest
identity), then run `curator project resolve`.

### source_name_conflict

Symptom: `source_name_conflict` with the colliding names.

Cause: two selections install one skill name, or two destination
names are filesystem-equivalent.

Remedy: give each installed skill exactly one selection (rename or
drop a duplicate), then run `curator project resolve`.

### source_output_overlap

Symptom: `source_output_overlap` with the overlapping path.

Cause: a selected package sits inside managed output (`.agents`,
adapter directories, the manager home, caches, staging, or the
snapshot store), or a root package lacks `root_inputs` admission.

Remedy: move the authored package out of managed output, or admit a
root package via `root_inputs` in machine source-policy.json, then
run `curator project resolve`.

### source_snapshot_changed

Symptom: `source_snapshot_changed` with the member name.

Cause: admitted inputs changed during capture, or a source replay
produced a package identity or `content_sha256` that differs from the
committed lock.

Remedy: restore the declared source to the package identity and
`content_sha256` in `Skillfile.lock.json`, then retry install. Run
`curator project refresh` only when intentionally changing the lock.

### source_snapshot_unavailable

Symptom: `source_snapshot_unavailable` with the member name.

Cause: the declared path or Git source cannot be reached to replay the
locked package. A missing local snapshot and machine bindings are
expected on a fresh machine when the source is available.

Remedy: restore access to the declared path or Git source, then retry
install.

### manager_state_unreadable

Symptom: `manager_state_unreadable`, sometimes nested under a higher-level
diagnostic such as `environment_source_invalid` or `source_snapshot_unavailable`.

Cause: a manager-owned state path or one of its parent directories cannot be
inspected or read reliably. This is different from a missing state entry, so
the fresh-machine default or snapshot replay fallback does not apply.

Remedy: restore the expected file and directory structure and access to the
state path, then retry the command. Do not remove or recreate the state until
you have confirmed which manager record is unreadable.

### source_lock_stale

Symptom: `source_lock_stale`.

Cause: the Skillfile changed since the lock was published, or the
machine bindings no longer belong to the lock generation.

Remedy: run `curator project refresh`, then `curator install` to
materialize the refreshed lock.

### repository_endpoint_unavailable

Symptom: `repository_endpoint_unavailable` with the canonical identity
and one clause per attempted endpoint.

Cause: every planned endpoint failed with an availability or
authentication error, or a logical `repository` declaration has no
machine policy entry.

Remedy: verify the network path and operator authentication for the
listed endpoints, then retry with machine source-policy.json.
A `credential.helper` from user or system Git configuration is
never consulted on this lane and no prompt occurs, so a private
HTTPS Skillfile source that relied on a helper fails here with an
authentication clause: provide the credential through the invoking
environment instead — a non-interactive `GIT_ASKPASS` program that
answers git's username and password prompts (or embed the username in
the endpoint URL), or an SSH endpoint with an agent via
`SSH_AUTH_SOCK` — then retry the explicit attempt.
SSH on this lane runs a curator-owned command with an empty ssh
config: host aliases, `ProxyCommand`, `IdentityFile` and every other
`~/.ssh/config` entry are not read, so an SSH source that relied on an
alias or a `ProxyCommand` fails here — use an agent (or a
default-named key) and the literal host name instead. Unknown hosts
fail closed with a host-key clause: seed `known_hosts` with
`ssh-keyscan` or a first manual `ssh`, then retry. The proxy
environment is not honoured on this lane either: a source reachable
only through a proxy fails here with an availability clause.

### repository_policy_invalid

Symptom: `repository_policy_invalid` with the offending field.

Cause: machine source-policy.json is missing a required shape, names
an unknown member or version, lists a bad endpoint, or cannot be read.

Remedy: fix machine source-policy.json beside the manager
configuration; an invalid policy is never treated as absent.

### repository_mirror_undeclared

Symptom: `repository_mirror_undeclared`.

Cause: the resolved connection host differs from the entry-key host
without a `mirror_of` attestation equal to the key.

Remedy: attest the mirror with `mirror_of` equal to the entry key in
machine source-policy.json.

### repository_alias_unknown

Symptom: `repository_alias_unknown` with the alias name.

Cause: an endpoint `alias` field names no entry of the policy
`aliases` table.

Remedy: declare the alias in the `aliases` table. The table lives in
machine source-policy.json.

### source_audit_rejected

Symptom: `source_audit_rejected` with the member name.

Cause: the machine binding for the locked package is malformed or no
longer matches the locked package, context, policy, or evidence.

Remedy: re-resolve under trusted machine policy; the persisted audit
report must match the locked package.

### source_audit_unavailable

Symptom: `source_audit_unavailable` with the member name.

Cause: no machine binding or audit report exists for the locked
package yet — planning must not invent trust.

Remedy: run the explicit attempt under trusted machine policy so the
audit report is persisted.

### build_repository_identity_invalid

Symptom: `build_repository_identity_invalid: transport plan endpoint
N carries an explicit port or host alias outside the strict
external-build lane grammar`, on the resolved lane only
(`CURATOR_DRAFT_TRANSPORT_RESOLUTION=1` with a machine policy).

Cause: the planned endpoint entry carries an explicit port or a host
alias, which the strict external-build lane refuses deterministically
(§7) before any fetch. Other `build_repository_identity_invalid`
diagnostics keep their lane behavior and carry no appended guidance.

Remedy: fix the endpoint entry in machine source-policy.json: the
strict external-build lane admits no explicit port and no host alias,
then retry the explicit attempt.

## Environment credential links

### environment_credential_conflict

Symptom: `env resolve --repair` stops with
`environment_credential_conflict` naming a managed link path such as
`.../environments/<profile>/<env>/auth.json`, and emits no fragment.
`env status` (and a bare `env resolve`) report the same state as a
detached passthrough entry with conflict-class wording.

Cause: the link path holds something repair must not displace — a
regular file (the tool severed the link by writing over it, or the
managed home holds its own credential bytes), an unrecorded symlink to
an unexpected target, or a non-empty directory. A stale recorded link
(`shared`→`isolated`, or a store gone ambient) and a mis-targeted
*recorded* link — for example a `pi` home still aimed at the pre-0017
native `~/.pi/auth.json` instead of `~/.pi/agent/auth.json` — are
reported detached with conflict-class wording, but repair never moves
them: the refusal says `migration needed` and names the `curator env
migrate --plan` / `--apply --expect <plan-hash>` invocations that print
and perform the move.

Not this error: a correctly targeted link whose native target does not
exist yet is the detached-pending warning (`link target ... does not
exist yet — log in to <tool> to populate it`), not a conflict —
provisioning, repair, and bare resolve all succeed loudly. Establish the
native credential — log in natively, or log in inside the managed home —
and the finding clears on the next run.

Remedy: when the refusal says `migration needed`, run `curator env
migrate --plan` to print the exact operations, then `curator env
migrate --apply --expect <plan-hash>` to execute them; do not re-point
the link by hand to silence the finding. Otherwise resolve the conflict
out of band, then re-run: inspect both sides — the managed path the
diagnostic names and the declared native target — decide which
credential bytes win, move the loser aside yourself (the manager never
moves credential bytes), and re-run with `--repair`.

### Migration conflicts

Symptom: `curator env migrate --apply --expect <hash>` refuses with
`environment_credential_conflict`, and the printed plan ends with
`blocked:`.

Cause: a state the migration must not touch without an operator
decision. The plan names the exact choice: an isolated→shared account
choice (the managed home holds its own credential bytes at the link
path — decide which bytes win and move the loser aside yourself); a
regular file at a link path (same choice); two live Pi credentials
(`~/.pi/auth.json` and `~/.pi/agent/auth.json` both hold bytes — decide
which holds the live credential and reconcile them yourself); a foreign
or uninspectable link (remove it or restore access yourself). The
manager never copies, moves, or deletes credential bytes at any step.

Remedy: carry out the named choice out of band, then re-run
`--plan` and `--apply --expect <plan-hash>`. Bytes at the old Pi root
alone (with the agent root empty) are not a conflict: the migration
relinks and warns, leaving the old bytes untouched — move them yourself
if they are live.

### Migration plan drift

Symptom: `curator env migrate --apply --expect <hash>` refuses with
`environment_credential_conflict: migration plan drift`, or `--apply`
without `--expect` is rejected outright.

Cause: every apply requires the hash of a prior complete plan. A bare
`--apply` is a usage error (the library refuses it the same way before
any mutation). A drift refusal means the inventory changed after `--plan`
printed — another operation moved a link, a marker was edited (the hash
covers the marker identities), a native file appeared or vanished, or
the hash was copied wrong. The refusal writes nothing.

Remedy: re-run `curator env migrate --plan` (with the same scope flags)
and apply with the new hash it prints.

### Interrupted migration apply

Symptom: `curator env migrate --inspect` or `--plan` banners an
`interrupted migration apply` and the plan ends with `recover:` instead
of `ready:`.

Cause: a previous `--apply` was killed between mutations (or failed
where the rollback could not complete). The durable journal below
manager state records the intent and the rollback data; the
read-only commands report it and recover nothing.

Remedy: run `curator env migrate --apply --expect <plan-hash>` with the
hash the banner names. The apply recovers to the prior state under the
manager lock, announcing the recovery before mutating, then executes
the plan. If the banner instead says the journal is unusable, back the
journal file up out of band, verify every managed link by hand, and
only then remove the journal and re-run `--plan`.

### environment_credential_unsupported

Symptom: provisioning or repairing a `codex_cli` home fails with
`environment_credential_unsupported` naming the native
`cli_auth_credentials_store` selector.

Cause: the operator's native `config.toml` selects a credential store
outside the verified `file`/`keyring`/`auto` set, so the manager cannot
establish the credential strategy and fails closed. Only the top-level
`cli_auth_credentials_store` key counts, in any valid TOML spelling; a
same-named key nested inside a table is not the selector. An absent
file or an absent key is not this error: both resolve to the platform
default `file` store. A `config.toml` that does not parse, or a
non-string value for the key, fails closed the same way instead of
reading as absent.

Remedy: set `cli_auth_credentials_store` in the **native**
`config.toml` to one of `file`, `keyring`, or `auto` (or remove the key
for the `file` default) and retry. Sharing is defined by the native
effective storage only: editing the managed copy changes nothing. Note
that `isolated` is admitted under `file` storage only; under `keyring`
or `auto` it is refused with `environment_isolated_unsupported`.

### environment_repair_failed on an uninspectable credential target

Symptom: `env resolve --repair` stops with
`environment_repair_failed` carrying `cannot be inspected` and
`environment_credential_conflict` for a managed link path, and emits no
fragment. A bare `env resolve` (and `env status`) report the same home
stale with the inspection diagnostic.

Cause: the managed link correctly targets the declared native store,
but the native target itself cannot be statted — permissions, I/O, or
a broken parent — so there is no re-link that heals it and repair
fails instead of guessing. This is distinct from absence: a target
that does not exist yet is the detached-pending warning, never this
error. Nothing is moved, removed, or re-pointed.

Remedy: restore access to the native target out of band — fix the
permissions or parent, or restore the file — then re-run with
`--repair`: repair converges and the link reads through once the
target stats again.
## Enforced script execution diagnostics

Enforced script commands (`execution_policy: "script-worker-v1"`) launch
through the manager-owned script worker. Source locations:
`internal/scriptpolicy/scriptpolicy.go` (admission and preflight),
`internal/scriptworker/` (interpreter resolution and worker session).

### script_execution_worker_protocol_invalid for stream bounds

Symptom: an enforced launcher reports script_execution_worker_protocol_invalid
with "launcher standard input exceeds the session bound" or "interpreter
output exceeded the capture bound".

Cause: piped or file standard input is read completely before the worker
starts and is limited to 64 MiB. A terminal is bound to the null device,
not forwarded interactively. Standard output and standard error are
captured together under a 16 MiB budget; crossing it refuses the invocation
and the partial capture is not forwarded.

Remedy: keep piped input within the limit and reduce combined command
output. Commands that need interactive input, live output, or larger output
are outside the bounded script-worker stream model. If the input and output
are within those limits and the refusal repeats, report it as a manager
protocol failure.

### script_execution_worker_protocol_invalid for worker setup

Symptom: an enforced launcher reports script_execution_worker_protocol_invalid
for a worker channel, session frame, manager-derived path, private runtime
directory, or interpreter start failure.

Cause: the manager and worker could not complete their fixed request/result
session. This includes a malformed or out-of-order session message, an
unknown session nonce, an invalid manager-owned working or private path, a
failure to create the operation-private roots, or a worker channel/start
failure. The launcher fails closed and does not forward child output when
the session cannot return a valid result.

Remedy: verify the manager executable is intact and the operator temporary
directory is available and writable. If the paths and host are valid but
the refusal repeats, report the diagnostic as a manager/worker protocol
failure. See the stream and Linux Landlock sections below for those
specific protocol-invalid diagnostics.

### script_execution_control_unavailable

Symptom: install of an enforced command, or an executed native launcher,
reports `script_execution_control_unavailable` naming a native control
such as `descendant-domain-termination`, and no worker starts.

Cause: the control table is complete, but this host cannot provide the
named mandatory control: the per-invocation native-control probe did not
find its mechanism. Install probes before staging anything, and every
invocation probes again before the worker starts, so the refusal names
exactly the controls the host lacks. Host-conditional controls the probe
does not find (Linux cgroup delegation, Landlock, network namespaces)
never refuse: they are reported unavailable in the invocation's evidence
record and the run proceeds. An unbound interpreter refuses with the same
code because interpreter resolution cannot be applied on this host (see
`docs/script-interpreters.md`).

Remedy: provide the missing host capability (for example, run where the
platform mechanism exists), or bind the interpreter when the detail names
interpreter resolution. Do not work around the refusal by removing the
`execution_policy` field: that silently downgrades the command to
uncontained execution.

Verify which controls this host provides:

```bash
curator skill check ./my-skill
```

Admission succeeds; only a host that cannot provide a mandatory control
refuses, at install and at invocation.

### script_execution_worker_protocol_invalid naming a Landlock control

Symptom: on Linux, an enforced invocation reports
`script_execution_worker_protocol_invalid` with `cannot install inventory
control "filesystem-write-confinement"` (or `"descendant-exec-denial"`)
followed by the cause naming the offending path, and the interpreter
never runs.

Cause: the write confinement rules the derived path set — the declared
`filesystem` paths beneath the canonical project root — plus the
operation-private area and the null device (which confined scripts and
their descendants legitimately open for redirected standard streams),
and a derived member that does not exist cannot be ruled. The worker
refuses fail-closed rather than running with a control the probe found
present left unenforced. Outside the ruled set, file writes and
truncation, entry creation and removal (including `mkdir`, `unlink`,
`rmdir`, `mkfifo`, and rename), and reparenting are all denied with
`permission denied`; reads stay unrestricted.

Remedy: declare paths that exist at invocation time. A `repo`
filesystem always grants the project root itself; a path set grants
exactly its members, so create the members before invoking, or narrow
the declaration to the members the command needs.

### script_execution_capability_evidence_invalid

Symptom: an enforced invocation reports
`script_execution_capability_evidence_invalid`.

Cause: the worker's capability-evidence record contradicted this
invocation's own probe — a missing, duplicated, or unknown entry, a
status inconsistent with the probed availability, a foreign record
version, a replayed install-generation timing, or a second record for
one invocation. The parent validated the record before the permit, so
the interpreter never ran.

Remedy: none on the operator side beyond retrying on a stable host; a
repeated refusal names a manager defect, not a package defect. The
invocation's record, when one was produced, is result-only and never
carries command output.

### script_execution_hardened_claim_forbidden

Symptom: an enforced invocation reports
`script_execution_hardened_claim_forbidden`.

Cause: the evidence record claimed a guarantee the policy defers — a
foreign execution policy, or one of the thirteen deferred guarantees
(seven `script-*` plus six build guarantees). The claim is rejected
before the permit.

Remedy: none; the record is rejected by construction. See Protocol Core
§4.1.1 for the deferred list.

### script_execution_policy_unsupported

Symptom: install or `skill check` reports
`script_execution_policy_unsupported` for an enforced command.

Cause: the command selects an execution policy this manager does not
implement. Protocol 1.0 admits exactly `script-worker-v1`, which this
manager implements. A refusal with this code therefore points to a
different selected policy, which must not be installed declared-only,
downgraded, or ignored.

Remedy: use the supported `script-worker-v1` policy with its co-required
interpreter (`node-v1` or `python3-v1`), or remove both fields to keep the
command declared-only. A supported policy can still refuse with
`script_execution_control_unavailable` when a mandatory host control or
operator interpreter binding is unavailable.

### script_execution_worker_identity_invalid

Symptom: an enforced invocation reports
`script_execution_worker_identity_invalid`.

Cause: the worker or interpreter executable failed identity verification:
the file is not a canonical regular executable, carries multiple links
or a reparse point, changed between resolution and the launch boundary,
or does not hash to the recorded identity. The interpreter binding comes
exclusively from the operator-trusted `script_interpreters` mapping in
machine configuration — never from the repository, the runtime store,
`.agents/bin`, the user `PATH`, or a manifest value. A binding that is
not a native interpreter image is refused the same way: POSIX `#!`
wrapper scripts (pyenv/asdf/volta-style shims) and Windows `.cmd`/`.bat`
files would interpose another program between the worker and the
interpreter, so only the native interpreter executable is accepted. On
Windows the binding must name the `.exe` itself — an extensionless path
would execute a different file than the verified one.

Remedy: verify the `script_interpreters` entry for the command's
identifier names the absolute path of the installed native interpreter
executable, that no other process replaced the manager or interpreter
files, and retry. If the binding names a shim or wrapper, replace it
with the interpreter binary it launches. If the binding is absent, add
it:

```json
{
  "script_interpreters": {
    "node-v1": "/opt/node/bin/node",
    "python3-v1": "/usr/bin/python3"
  }
}
```

On Windows the values name the `.exe` files, for example
`"node-v1": "C:\\tools\\node\\node.exe"`.

The bindings are documented in full in [Operator-trusted script
interpreter bindings](script-interpreters.md).

### Enforced derivation reports

Symptom: an enforced install prints `is enforced (script-worker-v1)`
messages about withheld variables, recorded hosts, or unresolvable
executables.

Cause: this is the install record, not an invocation result. Every enforced
invocation derives its environment, `PATH`, network configuration, and
working directory from the command's declared capabilities, deny by
default: manager-owned variable names (loaders, interpreter options,
temporary and configuration roots, proxy and resolver configuration)
never pass through, declared network hosts are recorded reporting-only
without filtering, manager-resolved `exec` grants enter the private `PATH`
farm, and secrets remain identifiers. The messages name each decision
so nothing is dropped silently.

Remedy: none required. If a withheld variable breaks a script, remove it
from `env_read` and read the value through a different declared variable.

### Unresolved declared `exec` name

Symptom: an enforced install or invocation report names an unresolved
`exec` grant.

Cause: the manager could not verify that executable in its fixed search
directories, so it is omitted from the private PATH farm and reported. The
worker still starts, but the invocation cannot resolve that name through
caller `PATH`, the repository, or script content. On Windows the default
list is `%SystemRoot%\System32` followed by `%SystemRoot%`; the multiple-link
allowance is limited to manager-resolved files physically below canonical
`%SystemRoot%\System32` derived from the manager's captured `SYSTEMROOT`.

Remedy: install the tool in a manager search directory, or remove its name
from the command's `exec` declaration if the script does not need it. Do
not add the caller's `PATH` to the search list.

### script_execution_package_influence_forbidden

Symptom: an enforced invocation reports
`script_execution_package_influence_forbidden`.

Cause: package-shaped input reached a boundary that refuses it: a
tampered or corrupt launcher sidecar, an interpreter identifier outside
the closed set, or a malformed capability declaration. The invocation
runs nothing.

Remedy: reinstall the skill so the manager republishes the launcher and
its sidecar contract, verify the command's `interpreter` is `node-v1`
or `python3-v1`, and validate the manifest's `capabilities` object
against the schema.

## Script audit labels

`curator audit`, `skill check`, and install report two warning classes for
script commands (manager profile §7). Both are always warnings in every
mode and never block: neither is a finding about source content, neither
is subject to `fail_on`, and neither may be reported as an applied
control. Declared-only skills validate and install exactly as before; the
warnings state the containment posture so reviewers and registries can
gate on the distinction.

### script-command-declared-only

Symptom: audit or `skill check` warns `script-command-declared-only` for
a command.

Cause: the command does not declare `execution_policy:
"script-worker-v1"` — every schema-7 script command and every schema-8
script command without the field. Its capability declaration is
documentation and bounds nothing at run time.

Remedy: to enforce containment, adopt `execution_policy:
"script-worker-v1"` with its co-required `interpreter` (`node-v1` or
`python3-v1`). To keep the command declared-only, no action is required;
the warning records the posture.

### script-command-unfiltered-declared-network

Symptom: audit or `skill check` warns
`script-command-unfiltered-declared-network` for an enforced command.

Cause: the command is enforced (`execution_policy: "script-worker-v1"`)
and declares non-empty `network` hosts. The globs are recorded and
reported; no portable filtering is applied and none is claimed. The
command is still admitted so the `exec`, `filesystem`, and environment
controls it can have are applied.

Remedy: declare only the hosts you need — reporting only. A
declared-only command with network hosts carries only
`script-command-declared-only`: the network label is about enforcement,
not declaration.

### Per-command execution-policy record

Symptom: `curator audit` prints `audit info: <skill>: command '<name>'
execution_policy=<identity>` lines for commands that carry no warning.

Cause: this is the audit record, not a finding. For every script command
the record carries the effective execution-policy identity or its
explicit absence — `script-worker-v1` for enforced commands, `(none)`
for declared-only ones — independent of warning eligibility. The same
entries are carried under `script_policies` in `curator audit --json`
output and in the stored verdict file, so reviewers and registries can
gate on the enforced/declared-only distinction directly.

Remedy: no action is required. An enforced command with no declared
`network` hosts audits clean and still records its identity; a
declared-only entry that should be enforced calls for the same remedy as
`script-command-declared-only` above.
