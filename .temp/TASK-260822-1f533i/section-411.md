#### 4.1.1 Portable `script-worker-v1` execution policy

A script command is declared-only unless it selects the enforced execution
policy named `script-worker-v1`. Protocol 1.0 defines exactly one script
execution policy, and every conforming manager MUST implement it on macOS,
Linux, and Windows. Selection is an OPTIONAL per-command field on a script
command of manifest schema 8 or later. Schema 7 and earlier script commands
keep their exact meaning above, and an absent field means declared-only on
every schema.

Selecting the policy is not a package-visible choice of policy. The value space
is closed and holds exactly one identity, so a package chooses only whether the
command is enforced, never how it is enforced. Every control, limit, path,
environment value, and executable stays manager-owned, and the influence of the
package field is monotonically restrictive: an enforced command may do strictly
less than the same command declared-only and never more. Section 4.2.1's rule
that a policy identity is never a package-visible option is unchanged. It
forbids package data selecting between execution contracts, which a one-value
opt-in cannot do.

Capabilities are declared once per skill in section 4.3, so every enforced
command of one skill derives the same containment profile. A manager MUST NOT
widen or narrow that profile per command from any other manifest value.

The fixed process graph is:

```text
manager parent
  -> identity-verified manager-owned script worker
       -> identity-verified interpreter for the declared identifier
            -> manager-resolved executables named by the `exec` capability
```

The worker is an exact re-execution of the installed manager executable in one
fixed hidden mode, the same boundary section 4.2.1 defines for builds. It is an
implementation boundary, not a user-visible command surface and not a
package-selected program. No package file, manifest value, script byte, shebang
line, file association, environment value, `PATH` lookup, shell, or user option
selects the worker executable or its mode. The fourth node exists only when
`exec` names executables; under `exec: "none"` the graph is exactly three
nodes. An implementation that cannot distribute an equivalent identity-verified
worker MUST state the exact executable graph it verifies instead and MUST treat
every mutable component of that graph, including the interpreter and its
installed package tree, as trusted computing base.

One worker session serves exactly one invocation of exactly one enforced
command. The worker applies the complete containment profile before the
interpreter starts and admits no second command, second interpreter, additional
program, shell, or control change afterwards.

An enforced script command names a closed interpreter identifier. Protocol 1.0
admits exactly `python3-v1` and `node-v1`. The manager, and only the manager,
resolves that identifier to an executable, once per invocation:

- resolution is package-independent and completes before the manager enters any
  package-controlled directory. A manager MUST NOT resolve an interpreter from
  the commit-keyed runtime store, a runtime root, a project command directory,
  the caller's `PATH`, a manifest value, a project file, or a script byte;
- the resolved target MUST be a canonical regular executable file. Symlink,
  reparse-point, and hard-link substitution MUST be rejected, strong file
  identity MUST be recorded, and the executable's bytes MUST be hashed;
- that identity MUST be re-checked at the launch boundary so a replacement race
  cannot widen the graph; and
- the manager MUST NOT run the interpreter to obtain a version string, and MUST
  NOT hash the interpreter's standard-library, library, or installed-package
  trees.

The interpreter's library and installed-package trees are trusted computing
base, not verified identity, and an implementation MUST say so where it reports
this policy. The toolchain fingerprint of section 8.2 MUST NOT be applied to an
interpreter. It hashes a manager-pinned tree that stays unchanged through the
last child exit, while an interpreter installation is host-owned and
legitimately mutated between two invocations of the same command.

The shebang line, the file extension, and the Windows file association are
inert under this policy and MUST NOT select the executed program. No further
interpreter identifier is admitted in protocol 1.0. `bash-v1` and
`powershell-v1` are the obvious candidates and are deferred because neither a
POSIX shell nor PowerShell resolves on all three supported platforms;
admitting either, or any other interpreter, is a specification revision under
section 12.3 and never a manager configuration option.

The script command launcher rules of the manager profile govern declared-only
commands. An enforced command's shim MUST NOT be a symlink, a POSIX-shell
wrapper, a `.cmd` wrapper, or any other program that resolves the command
through a shell or through the inherited `PATH`. It MUST be a manager-owned
launcher that starts the worker, and the worker MUST start the interpreter by
its resolved path, never through `cmd.exe`, PowerShell, `sh -c`, `PATHEXT`, or
a file association.

Under this policy the section 4.3 declaration is the policy input. Derivation
reads the declared manifest bytes: a capability field the manifest does not
contain derives the deny-by-default meaning below, and the schema default for
that field MUST NOT widen a derived control. The declared-only audit reading of
section 4.3 is unchanged for every command that has not opted in.

- `network: "none"`, and an absent `network`, derive no manager-configured
  network access. The manager sets offline network configuration for the
  interpreter and its descendants, scrubs proxy and resolver configuration out
  of the environment, and applies `network-isolation-domain` when the inventory
  probe reports it present. This is a manager mechanism and is not denial.
- `network` as a host-glob list derives no filtering. The manager MUST record
  the declared hosts, MUST NOT represent them as an applied control, MUST NOT
  place a host-filtering entry in the inventory or in an evidence record, and
  MUST NOT claim filtering on any surface. The command is admitted, not
  rejected, and carries the `script-command-unfiltered-declared-network` audit
  warning class.
- `exec: "none"`, and an absent `exec`, derive a `PATH` that resolves exactly
  the interpreter. A name set additionally derives exactly those names, each
  resolved by the manager to a fixed path under the interpreter resolution
  rules above. The manager MUST discard the inherited `PATH` and MUST build a
  `PATH` whose entries are manager-owned directories exposing exactly the
  resolved interpreter and the resolved declared names. A declared name the
  manager cannot resolve is absent from the built `PATH` and MUST be reported;
  it MUST NOT be resolved from the caller's `PATH` at launch or by the
  interpreter at run time.
- `filesystem` always derives an operation-private runtime area: a private
  temporary root, a private configuration root, a private cache root, and a
  manager-selected working directory, all resolved independently of package
  data. An absent `filesystem` derives that area and nothing else; the section
  4.3 schema default `"repo"` is the declared-only audit reading and MUST NOT
  widen the derived control. `"repo"` additionally derives the canonical
  project root of the invocation. `"home-config"` derives nothing beyond the
  private configuration root, because the manager redirects platform
  configuration, cache, and temporary environment into the private roots; an
  unmodified command that writes to a home configuration path writes into its
  private configuration root. A portable relative path set derives exactly
  those paths beneath the canonical project root.
- `secrets` derives no secret material in every form. The manager passes no
  secret value into the worker or the interpreter, whether the field is absent,
  `"none"`, or a non-empty identifier set. The identifiers remain an audit
  declaration, MUST NOT cause a manager to inject a value, and MUST NOT widen
  the environment. Protocol 1.0 defines no secret-provider contract for this
  policy.
- `env_read` derives exactly the named host variables that are present in the
  manager's own environment; every other inherited variable is absent. The
  manager owns every name it sets under this policy and every name that selects
  a program, a library or module search path, an interpreter startup file or
  option, a temporary or configuration root, or proxy or resolver
  configuration. An `env_read` entry naming a manager-owned name MUST NOT pass
  the inherited value through; the manager-set value stands. The exact reserved
  set per platform and per interpreter identifier is defined by the manager
  profile.
- `prompt_scope` derives no control.

The following controls are mandatory on every supported host, and they are the
only controls whose absence rejects an invocation:

- the fixed process graph above: within this execution boundary the manager and
  the worker start only the nodes above and no other program;
- worker executable identity verification before launch, and re-verification at
  the launch boundary;
- interpreter resolution and per-invocation executable identity verification as
  defined above;
- a manager-built environment: an empty bootstrap, the manager-set values, and
  exactly the `env_read` names passed through, with the inherited environment
  otherwise discarded;
- the manager-built `PATH` derived above, with the inherited `PATH` discarded;
- offline network configuration plus proxy and resolver scrubbing whenever the
  derived network capability is `none`;
- the operation-private temporary, configuration, and cache roots and the
  manager-selected working directory derived above;
- explicit manager-controlled standard-stream binding and release of unrelated
  descriptors or handles before the interpreter starts;
- application of every native control the inventory below marks available for
  the host platform, and of every `host-conditional` control the probe finds
  present;
- exactly one closed `script-capability-evidence-v1` record per invocation; and
- termination and joining of the complete worker domain before the invocation
  returns.

A manager that cannot apply all of them MUST reject the invocation with
`script_execution_control_unavailable` before starting the worker or the
interpreter. It MUST apply the same check when it publishes or updates the shim
of an enforced command and MUST reject that install or update with the same
diagnostic, so a shim that can never run is never published.

Three controls of section 4.2.1 are deliberately not carried across, because a
script command is a user-facing program and a build is not. This specification
states each divergence so that neither a reader nor an implementation inherits
the build rule by reflex:

1. Standard input stays open. Section 4.2.1 requires closed standard input.
   Under this policy the manager binds standard input explicitly, to the
   intended stream or to the host's null device, MUST NOT leave it bound to an
   inherited descriptor it did not intend, and MUST NOT close it as a matter of
   policy.
2. Output MAY stream. The bounded and redacted combined-output rule of section
   4.2.1 applies to output a manager captures. Direct pass-through streaming is
   permitted when the manager binds the streams itself, because a command
   launcher forwards the child's output and exit status transparently.
3. There is no policy deadline. An enforced command MAY run indefinitely. A
   wall-clock bound is an operator or invocation bound, not a control of this
   policy, and its absence MUST NOT be reported as an unavailable control.

Each mandatory control is a manager-enforced mechanism, not a kernel-enforced
guarantee. This specification states both sides so that neither a reader nor an
implementation can mistake one for the other:

| Portable mechanism | What it means | What it does not mean | Deferred guarantee |
|---|---|---|---|
| derived `network: "none"` | offline network configuration for the interpreter and its descendants, and proxy and resolver configuration scrubbed out of the manager-built environment | kernel-enforced network denial for the worker domain or its descendants | `script-total-network-denial` |
| declared host globs | the declared hosts are recorded and reported on the audit and result surfaces | any filtering, allowlisting, resolution failure, or denial of the named or unnamed hosts | `script-network-host-allowlisting` |
| manager-built `PATH` over the resolved interpreter and the resolved `exec` names | the inherited `PATH` is discarded and bare-name resolution reaches exactly the manager-resolved executables | kernel-enforced allowlisting of the executables a descendant may run; a script that names an absolute path, or that re-executes through an interpreter-specific mechanism, is not prevented from doing so | `script-exact-executable-allowlisting` |
| operation-private runtime area plus redirected configuration, cache, and temporary roots | every manager-directed write target is private to the invocation and the platform environment points at those private roots | kernel-enforced confinement of every descendant write to the private area and the derived paths | `script-private-runtime-area-only-writes` |
| worker and interpreter identity verification | the worker and the interpreter are canonical regular files whose identity is verified before launch and re-checked at the launch boundary | kernel-enforced read-only presentation of the runtime tree, the interpreter, or its installed package tree to descendants | `script-read-only-runtime-tree` |
| probe plus applied inventory controls | every control the inventory marks available for the platform, and every `host-conditional` control the probe finds present, is applied and recorded | hard aggregate process, memory, disk, time, and output bounds over every descendant | `script-hard-aggregate-descendant-resource-bounds` |
| mandatory-control preflight | the invocation, install, and update reject before the worker when a mandatory portable control cannot be applied | terminal rejection when an inventory or kernel-grade capability is absent | `script-fail-closed-capability-preflight` |

The script native-control inventory is exhaustive and normative per platform.
Its authority is the `native_control_inventory` section of
`conformance/v1/vectors/script-host-execution-policy.json`, version
`script-worker-v1-native-control-inventory-v1`:

| Control | macOS | Linux | Windows |
|---|---|---|---|
| `descendant-domain-termination` | available: process group and session teardown | available: process group and session teardown | available: Job Object kill-on-close |
| `active-process-count-limit` | unavailable: `no-private-aggregate-domain` | available: `RLIMIT_NPROC` | available: Job Object active-process limit |
| `aggregate-memory-limit` | unavailable: `no-private-aggregate-domain` | host-conditional: delegated cgroup v2 memory maximum | available: Job Object process and job memory limit |
| `per-file-size-limit` | available: `RLIMIT_FSIZE` | available: `RLIMIT_FSIZE` | unavailable: `no-private-aggregate-domain` |
| `inherited-handle-restriction` | available: close-on-exec plus explicit descriptor release | available: close-on-exec plus explicit descriptor release | available: explicit handle inheritance list |
| `descendant-exec-denial` | unavailable: `no-unprivileged-per-process-exec-policy` | host-conditional: kernel execute-right restriction | unavailable: `child-process-policy-requires-appcontainer` |
| `filesystem-write-confinement` | unavailable: `no-unprivileged-filesystem-domain` | host-conditional: kernel write-right restriction | unavailable: `no-unprivileged-filesystem-domain` |
| `network-isolation-domain` | unavailable: `no-unprivileged-network-domain` | host-conditional: private network namespace without interfaces | unavailable: `no-unprivileged-network-domain` |

This inventory is independent of `rc5-native-control-inventory-v1`. The first
five rows carry the same macOS and Windows verdicts because the underlying host
facts do not depend on what the child is, but they are copied, not referenced:
each inventory is versioned on its own, and a revision motivated by the build
policy MUST NOT re-scope script conformance, or the reverse.

`network-isolation-domain` is not, and MUST NOT be spelled as,
`total-network-denial`. Applying the control on a host that provides it does
not license claiming any policy-level guarantee: the mechanism is
host-conditional, so the guarantee stays deferred.

Availability has exactly three values in this inventory. `available` and
`unavailable` keep their section 4.2.1 meaning: a fixed normative per-platform
verdict that the probe confirms. `host-conditional` asserts that the platform
MAY provide the control and that the per-invocation probe decides, because the
kernel mechanisms behind the three Linux rows genuinely vary between hosts of
one platform and neither fixed verdict would be true. An `available` control
MUST report `applied`, an `unavailable` control MUST report `unavailable`, and
a `host-conditional` control MUST report `applied` or `unavailable` exactly as
probed. A `host-conditional` control that probes unavailable MUST NOT reject
the invocation.

The unavailable-reason vocabulary of this inventory is closed and contains
exactly `no-private-aggregate-domain`,
`no-unprivileged-per-process-exec-policy`,
`child-process-policy-requires-appcontainer`,
`no-unprivileged-filesystem-domain`, and `no-unprivileged-network-domain`.

A manager MUST apply exactly the controls this inventory marks available for
its platform, plus the `host-conditional` controls its probe finds present,
MUST NOT apply or report a control outside the inventory, and MUST NOT
substitute a host label for the availability probe. Availability MUST be probed
once per invocation before worker launch; a cached, inherited, or configured
result is not a probe, and an install-generation result replayed at invocation
time is a cached result. Adding, removing, or re-scoping an entry requires a new
inventory version. That is a specification revision, not an execution-policy
revision, because inventory membership never enters an installation input or
any hashed identity.

Host capability evidence is exactly one closed `script-capability-evidence-v1`
record per enforced-command invocation. The record contains exactly
`record_version`, `execution_policy`, `platform`, and `controls`.
`execution_policy` is `script-worker-v1`. `platform` is `linux`, `macos`, or
`windows`. `controls` contains exactly one entry per inventory control, and
each entry contains exactly `name`, `availability`, `status`, and `probed_at`.
`availability` is `available`, `host-conditional`, or `unavailable`, `status`
is `applied` or `unavailable`, and `probed_at` is `pre-worker-launch`.

The record describes the controls installed at launch, and that is a complete
statement for the whole session. A manager MUST NOT re-probe or emit a second
record while an enforced command runs, however long it runs, because controls
are installed before the interpreter starts and cannot change afterwards.

Each condition below is an error, not a permitted variation:

| Condition | Diagnostic |
|---|---|
| an `available` control reported with a status other than `applied` | `script_execution_capability_evidence_invalid` |
| an `unavailable` control reported with a status other than `unavailable` | `script_execution_capability_evidence_invalid` |
| a `host-conditional` control reported with a status the probe did not produce | `script_execution_capability_evidence_invalid` |
| a missing, duplicated, or extra control entry | `script_execution_capability_evidence_invalid` |
| an unknown `record_version` | `script_execution_capability_evidence_invalid` |
| availability not probed once per invocation before worker launch | `script_execution_capability_evidence_invalid` |
| more than one record for one invocation | `script_execution_capability_evidence_invalid` |
| a deferred guarantee of this policy named as a control entry | `script_execution_hardened_claim_forbidden` |
| a guarantee deferred by section 4.2.1 named as a control entry | `script_execution_hardened_claim_forbidden` |
| an `execution_policy` other than `script-worker-v1` | `script_execution_hardened_claim_forbidden` |

The record is result-only. It is exposed on manager result surfaces: invocation
plan, dry-run, and status results, plus an explicitly operator-selected
diagnostic destination that package data can never choose. It MUST NOT be
written to the command's standard output or standard error,
because those streams belong to the caller's pipeline, and it MUST NOT appear
in a cache key, receipt input, marker record, or claim. A manager retains at
most the most recent record per command by default; retention is machine-local
and operator-configurable, and a retained record MUST NOT become an input to
any hashed identity.

A `script-capability-evidence-v1` record and a `capability-evidence-v1` record
are separate closed objects. A manager MUST NOT emit either one in place of the
other, MUST NOT admit `script-worker-v1` into a `capability-evidence-v1`
record, and MUST NOT admit `manager-worker-v1` into a
`script-capability-evidence-v1` record.

This policy does not provide, and a conforming implementation MUST NOT claim,
these guarantees: `script-total-network-denial`;
`script-network-host-allowlisting`; `script-exact-executable-allowlisting`;
`script-private-runtime-area-only-writes`; `script-read-only-runtime-tree`;
`script-hard-aggregate-descendant-resource-bounds`; and
`script-fail-closed-capability-preflight`. They are reserved for a separately
named script execution policy backed by a verified provider under
[`assurance.md`](assurance.md), and none of the seven names may appear in the
mandatory-control set, the native-control inventory, or an evidence record. The
six guarantees deferred by section 4.2.1 are disjoint from these seven, describe
the build policy only, and may not appear on any script surface either.

There is exactly one portable failure boundary. A mandatory portable control
that cannot be applied rejects the invocation with
`script_execution_control_unavailable` before the worker starts, and rejects
the install or update of the command's shim with the same diagnostic. An
unavailable inventory native control, a `host-conditional` control the probe
does not find, and the absence of any deferred guarantee MUST NOT reject an
enforced invocation, MUST NOT produce a diagnostic, and MUST NOT prevent the
command from running; the invocation MUST NOT record any of them as applied.

The complete diagnostic set of this policy is:

| Diagnostic | Condition |
|---|---|
| `script_execution_control_unavailable` | a mandatory portable control cannot be applied, at shim install or update and again before worker launch |
| `script_execution_capability_evidence_invalid` | an evidence record violates the closure rules above |
| `script_execution_hardened_claim_forbidden` | an evidence record names a deferred guarantee or a foreign execution policy |
| `script_execution_policy_unsupported` | a manager that does not implement `script-worker-v1` reads a command that selects it |

A manager that does not implement this policy MUST reject such a command with
`script_execution_policy_unsupported`. It MUST NOT install the command
declared-only, downgrade it, or ignore the field, because the resulting shim
would run package code the manifest says is contained.

Package-controlled bytes remain interpreter input only. They MUST NOT select or
modify the manager, worker, or interpreter executable, hidden mode, or
identity; any argument vector the manager builds, environment value, `PATH`
entry, working directory, or standard-stream binding; the applied controls,
limits, permitted roots, or probe result; the evidence record or any surface it
appears on; or the audit record, warning classes, or policy identity. A shebang
line, a file association, and a script byte are inert with respect to program
selection.

The execution policy identity is part of the skill's audit surface, so a
reviewer and a registry can tell an enforced command from a declared-only one.
Two audit warning classes are named by this policy and emitted on the surfaces
the manager profile defines: `script-command-declared-only`, for every script
command that has not opted in, including every script command of manifest
schema 7 and earlier; and `script-command-unfiltered-declared-network`, for an
enforced command whose declared `network` is a host-glob list.

Opting in is a breaking change for a skill whose capability declaration was
written as documentation. Under enforcement an undeclared environment variable
is absent, an undeclared executable does not resolve, and a write outside the
derived paths lands in the private runtime area or fails. A manager SHOULD
report that difference when the shim of an enforced command is installed or
updated, rather than leaving it to be discovered at first invocation.

A different execution contract requires a different execution-policy identity.
`script-worker-v1` and `manager-worker-v1` are separate closed identities that
never alias, are never substituted for one another, and are never widened in
place. Admitting a further script execution policy, a further interpreter
identifier, a further inventory control, or a further availability value is a
specification revision under section 12.3 with its own review, conformance
vectors, and record version.
