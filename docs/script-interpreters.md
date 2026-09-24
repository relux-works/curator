# Operator-trusted script interpreter bindings

Enforced script commands (`execution_policy: script-worker-v1`) run through
the manager-owned script worker. This page documents the
`script_interpreters` machine configuration that binds the closed
interpreter identifiers to executables. It follows the [Curator
Specification](https://github.com/relux-works/curator-spec): interpreter
identity and the reserved execution environment are defined by `Protocol
Core §4.1.1`, and the enforced launch order by `manager profile §3.1`.

Interpreters are selected only by the operator. A manifest value, project
file, repository path, runtime root, command directory, caller `PATH`, or
script byte cannot choose one.

## Native control matrix

The eight-control inventory records fixed availability by operating system
and probes Linux host-conditional controls on every invocation.

| Platform | Available and applied | Host-conditional | Fixed unavailable |
| --- | --- | --- | --- |
| Linux | descendant termination (process group/session), per-file size (RLIMIT_FSIZE), inherited-handle restriction | active-process count and aggregate memory (delegated cgroup v2); descendant exec and filesystem writes (Landlock); network isolation (network namespace without interfaces) | none |
| macOS | descendant termination (process group/session), per-file size (RLIMIT_FSIZE), inherited-handle restriction | none | active-process count, aggregate memory, descendant exec, filesystem writes, network isolation |
| Windows | descendant termination, active-process count, and aggregate memory (Job Object); inherited-handle restriction (explicit inheritance list) | none | per-file size, descendant exec, filesystem writes, network isolation |

Fixed-unavailable controls are recorded as unavailable and do not refuse an
invocation. A mandatory-control probe failure still refuses before the worker
starts.

## Configuration

`script_interpreters` is an optional object in the machine configuration
(`~/.curator/config.json`, or `CURATOR_CONFIG`). Each key is a closed
interpreter identifier and each value is the absolute path of the installed
native interpreter executable. Exactly two identifiers exist:

```json
{
  "script_interpreters": {
    "node-v1": "/opt/node/bin/node",
    "python3-v1": "/usr/bin/python3"
  }
}
```

Unknown identifiers are rejected when the configuration loads. A binding
that is absent refuses at launch with
`script_execution_control_unavailable`: the command stays installed but
never runs until the operator binds its identifier.

## Per-invocation controls and evidence

Every enforced invocation probes the eight-control native inventory before
the worker starts, applies exactly the controls the probe found
(`available` controls must be present; host-conditional controls apply
only when found; fixed-unavailable controls never apply and never
reject), and returns exactly one closed result-only
`script-capability-evidence-v1` record. The parent validates the record
against its own probe before permitting the run: a contradiction refuses
with `script_execution_capability_evidence_invalid`, and a claim of a
deferred guarantee refuses with
`script_execution_hardened_claim_forbidden`. A host that cannot provide
a mandatory control refuses install and invocation with
`script_execution_control_unavailable` before any worker starts. The
applied Linux controls include Landlock filesystem write confinement
over the derived path set, the operation-private area, and the null
device (for redirected standard streams). The confinement handles
every filesystem mutation right the running kernel's Landlock ABI
provides — file write and truncation, entry creation (regular files,
directories, FIFOs, symlinks, and the other node types), entry
removal, and reparenting, plus device ioctl on newer ABIs — granting
them only beneath the writable roots, so outside the grants every one
of those operations is denied and everything else stays denied;
reads stay unrestricted, and the null-device grant is file-typed so
the sink stays fully usable. Landlock descendant exec denial covers
the manager-owned `PATH` farm, with delegated cgroup v2 process and
memory bounds and a network namespace without interfaces; macOS and
Windows carry process-group or Job Object teardown, exact Job Object
or `RLIMIT_FSIZE` bounds where the platform provides them, and handle
hygiene.

The invocation record and the derivation report behind it are available
through the operator-selected `script_diagnostics_dir` machine
configuration: an absolute directory path (absent disables reporting)
that receives at most the most recent result-only document per command,
named `<skill>-<command>.script-capability-evidence-v1.json`. Package
data can never choose the destination. The record carries versions,
names, and statuses only — never command output — and a destination
failure never fails the invocation.

Streams use a bounded request/result protocol. Piped or file standard
input is read into one payload before launch and refused with
script_execution_worker_protocol_invalid above 64 MiB. Terminal input is
bound to the null device, so interactive input is not supported. Standard
output and standard error share a 16 MiB capture budget. If either stream
crosses the remaining budget, the worker reports overflow and launch fails
with script_execution_worker_protocol_invalid; the partial capture is not
forwarded. Live pass-through and interactive commands are outside the
supported behavior, as are commands whose combined output exceeds the cap.

## Binding rules

Each binding is verified before every enforced invocation, and re-verified
at the launch boundary:

- The value must be an absolute path to a canonical regular executable
  file. Directories, devices, and other non-regular files are refused.
- Symbolic links resolve to their physical target, so Homebrew-style
  linked interpreters keep working; hard links and Windows reparse points
  are refused as substitutions.
- The binding must name a native interpreter image, and the worker
  re-checks the image header before the interpreter runs. POSIX `#!`
  wrapper scripts (pyenv/asdf/volta-style shims) and Windows `.cmd`/`.bat`
  files would interpose another program between the worker and the
  interpreter, so they are refused. On Windows the binding must name the
  `.exe` itself — an extensionless path would execute a different file
  than the verified one.
- A binding under a snapshot, the runtime store, or `.agents/bin` is
  refused: interpreters never resolve from package-controlled roots.

When verification fails the invocation refuses with
`script_execution_worker_identity_invalid` (the file changed or is not a
usable interpreter) or `script_execution_control_unavailable` (no usable
binding is configured). See [Troubleshooting](troubleshooting.md) for the
diagnostics.

## What the binding does not do

The binding selects the interpreter executable and nothing else. The
derived execution environment — the manager-built `PATH`, the offline
network configuration, the working directory, the private temporary,
configuration, and cache roots, and which host variables pass through —
derives from the command's declared capabilities at every invocation, and
declared network hosts and secrets are recorded as identifiers without
widening what the interpreter can reach. An enforced install reports each
command's derivation (withheld variables, recorded hosts, manager-resolved
executables, and unresolved names) in its result messages. A declared `exec`
name enters the private PATH farm only when the manager resolves it through
the fixed search directories; unresolved names are reported and remain absent
from the farm. Caller `PATH` is never a fallback.
On Windows, nil `ExecSearchDirs` uses `%SystemRoot%\System32` followed by
`%SystemRoot%`. System32 component-store files may have multiple hard links;
the allowance applies only to the default list, to a candidate physically
below canonical `%SystemRoot%\System32` derived from the manager's captured
`SYSTEMROOT`. The manager hashes and rechecks that source and copies verified
bytes into the private PATH farm before launch.
