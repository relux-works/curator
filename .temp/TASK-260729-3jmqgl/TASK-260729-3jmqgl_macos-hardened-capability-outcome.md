# TASK-260729-3jmqgl — macOS hardened capability probe outcome

**Task:** prototype-macos-hardened-capability-probes
**Story:** STORY-260728-327soo — fail-closed-cross-platform-build-execution
**Date:** 2026-07-29
**Role:** developer
**Status:** measured evidence; prototype only, no production integration

---

## 1. What this is, and what it is not

This is a **capability observation** produced by an executable harness. It is
**not** an enforcement claim, not a qualification, and not an implementation.

The harness always emits `qualification_status: "unqualified"` for macOS and
cannot be made to emit anything else. Its probe domains contain no package byte,
run no Go compiler, and produce no artifact. Nothing in Curator imports it; it
is a separate Go module under `prototypes/macos-hardened-probes/` in the
task-owned worktree and has not been staged, committed, or published.

**Headline result: on the measured host, macOS establishes 1 of the 6 hardened
guarantees. The harness rejects, fail-closed, with exit status 1.**

---

## 2. Host and tooling

Captured by `./capture-evidence.sh`, recorded verbatim in `host.txt`:

| Fact | Value |
| --- | --- |
| Date (UTC) | 2026-07-29T11:22:36Z |
| Product | macOS 26.5 (build 25F71) |
| Kernel | Darwin 25.5.0, `xnu-12377.121.6~2/RELEASE_ARM64_T6031` |
| Architecture | arm64 (Apple Silicon) |
| UID | 502 (unprivileged, not root) |
| System Integrity Protection | enabled |
| Go | go1.25.5 darwin/arm64 |
| `/usr/bin/sandbox-exec` | present, `-rwxr-xr-x 1 root wheel 102560 Apr 30 23:33` |
| golangci-lint | 2.4.0 (built with go1.25.5) |

The host is the local macOS primary host, not `ssh relux`. The acceptance
criterion permits either.

---

## 3. Exact commands and exit codes

All five cases ran as standalone processes. No command was piped through `tee`;
each exit status below is the real status of the named process.

| Case | Command | Exit | Expected |
| --- | --- | --- | --- |
| build | `go build -o hardened-probe ./cmd/hardened-probe` | 0 | 0 |
| list-classes | `hardened-probe --list-classes` | 0 | 0 |
| measure | `hardened-probe --work-dir … --evidence evidence.json --report report.json` | **1** | 0 or 1 |
| fail-closed sweep | `hardened-probe --work-dir … --report report-fail-closed.json --fail-closed-sweep --quiet` | **1** | 0 or 1 |
| assert-rejected | `hardened-probe --force-unavailable network-syscall-denial --expect rejected --quiet` | 0 | 0 |
| assert-established | `hardened-probe --force-unavailable network-syscall-denial --expect established --quiet` | **2** | 2 |

`capture-evidence.sh` itself exited 0, meaning every case produced the exit
status it is documented to produce.

Exit-code contract:

| Code | Meaning |
| --- | --- |
| 0 | every capability applied, every guarantee established |
| 1 | rejected: at least one capability could not be established (fail-closed) |
| 2 | the harness could not produce a trustworthy record |

1 and 2 are deliberately distinct. An unusable harness is not evidence about the
host, and conflating them would let a broken run read as a platform verdict.

### Verification commands run on this change

| Command | Exit |
| --- | --- |
| `go build ./...` | 0 |
| `go vet ./...` | 0 |
| `gofmt -l .` (empty output) | 0 |
| `go test ./...` | 0 |
| `go test -cover ./...` | 0 |
| `golangci-lint run --config ../../.golangci.yml ./...` → `0 issues.` | 0 |

Coverage, per package:

| Package | Coverage |
| --- | --- |
| `cmd/hardened-probe` | 81.2% |
| `internal/evidence` | 99.3% |
| `internal/inside` | 92.5% |
| `internal/probe` | 84.4% |
| `internal/seatbelt` | 97.7% |
| `internal/spec` | 100.0% |

Repository-root `golangci-lint run` remains broken for a pre-existing reason
unrelated to this task: it cannot load export data for the
`skill-go-testing-tools/tuitestkit` submodule dependency. The prototype module
is therefore linted directly, with the repository's own `.golangci.yml`.

---

## 4. Measured result

### 4.1 Guarantees

| Guarantee | Established | Blocked by |
| --- | --- | --- |
| `total-network-denial` | **no** | `preexisting-endpoint-revocation`, `domain-membership-enforcement` |
| `read-only-source-and-toolchain` | **no** | `filesystem-view-restriction` |
| `private-build-root-only-writes` | **no** | `filesystem-view-restriction`, `domain-membership-enforcement` |
| `hard-aggregate-descendant-resource-bounds` | **no** | `aggregate-resource-bounds`, `domain-membership-enforcement`, `domain-atomic-termination` |
| `exact-executable-allowlisting` | **no** | `domain-membership-enforcement` |
| `fail-closed-capability-preflight` | **yes** | — |

Record outcome: `rejected`, `rejected_before: capability-probe`,
`diagnostic: hardened_capability_unavailable`.

### 4.2 Capability classes

| Class | Verdict | Mechanism | What decided it |
| --- | --- | --- | --- |
| `domain-membership-enforcement` | unavailable | process group and session; seatbelt policy inheritance | a descendant that calls `setsid` leaves the group and session — the only grouping a plain macOS supervisor can create |
| `domain-atomic-termination` | unavailable | process-group signal | the detached descendant survived a process-group `SIGKILL`; the attached control was destroyed, so the teardown itself works |
| `network-syscall-denial` | **available** | seatbelt `(deny default)` with no network allowance | loopback/off-host TCP, UDP, unix connect and inbound listen all denied; the network-allowed control connects |
| `preexisting-endpoint-revocation` | unavailable | seatbelt profile applied at exec | a connected socket inherited across `exec` stayed writable inside the domain |
| `read-only-source-view` | **available** | seatbelt `file-read*` with no `file-write*` | eleven mutation classes denied, including `setxattr`, hard link and symlink; the writable control creates files |
| `read-only-toolchain-view` | **available** | same | same |
| `write-path-confinement` | **available** | `file-write*` restricted to the private build root | absolute, home, `/tmp`, `/private/tmp`, parent, `..`-traversal, `TMPDIR`, symlink and hard-link escapes all denied; no escape artifact exists on disk afterwards |
| `filesystem-view-restriction` | unavailable | seatbelt path-based denial | `/etc/passwd`, `/Users`, home and `.ssh` are denied, but the root namespace can still be enumerated — see §5 |
| `exec-path-allowlist` | **available** | `process-exec*` with literal allowances plus bounded `file-map-executable` | shell, bash, zsh, python3, `/bin/ls`, perl, dyld-as-program and a byte-identical self-copy in the build root all refused; the allowlisted path starts |
| `aggregate-resource-bounds` | unavailable | POSIX `RLIMIT_*` | the limit binds each process separately: parent and child budgets add past the declared cap; the declared byte budget is not enforced at all |
| `active-capability-probe` | **available** | probe domains via `/usr/bin/sandbox-exec` | every class produced a real observation in this operation; the missing-observation detector fires; an unprobed class cannot be claimed |

### 4.3 Fail-closed sweep

Every one of the eleven classes was forced unavailable in turn. All eleven
rejected before domain entry with `rejected_before: capability-probe`,
`diagnostic: hardened_capability_unavailable`, exit 1, and the forced class
reported `unavailable` / `not-applied`. `pass = true` for all eleven.

Injected verdicts are marked as injected in the report (`"forced unavailable by
--force-unavailable; this value was injected, not measured"`), so a forced value
can never be mistaken for a measurement.

---

## 5. The three findings that block macOS

### 5.1 There is no unescapable domain

A descendant that calls `setsid` leaves the process group and session the
supervisor created, survives the group-directed `SIGKILL`, and is no longer
reachable by any handle the supervisor holds. Measured directly: the detached
descendant survived, the attached descendant did not.

The survivor **does** keep its seatbelt policy — policy inheritance is real on
this backend and cannot be renounced. So the survivor still cannot touch the
filesystem or the network. But policy inheritance binds *access rules* only, not
accounting and not termination membership. Because `domain-membership-enforcement`
is a required class for four of the six guarantees, this single fact blocks
network denial, private writes, resource bounds and executable allowlisting even
where the underlying control itself measured available.

macOS exposes nothing with `JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE` or `cgroup.kill`
semantics to third-party code.

### 5.2 The dynamic loader forces the root namespace open

A `(deny default)` profile that omits read access to the literal `/` aborts every
dynamically linked program during `dyld` startup with `SIGABRT`, even when every
other needed subpath is allowed. Metadata-only access is not enough; the loader
needs `file-read-data` on `/`.

Granting it leaves the root namespace enumerable from inside the domain. Reading
the individual undeclared paths is still denied — `/etc/passwd`, `/etc/hosts`,
`/Users`, the home directory, `/Applications` and `.ssh` all refuse — and the
uncontained negative control shows those same reads succeed outside a domain, so
the denials are attributable to the profile. But listing the root by name is
reaching outside the declared views, so `filesystem-view-restriction` is recorded
as unavailable rather than quietly excused.

This is a property of the backend, not of the profile: it cannot be closed with
a better seatbelt profile.

### 5.3 Resource bounds are per process, and storage is not bounded at all

`RLIMIT_NOFILE` was set to 64 and did bind the process that set it — the negative
control confirms the instrument works. But a child inherits the same soft limit
as a *fresh* budget, so parent and child together open past the cap. `RLIMIT_NPROC`
is per user, not per operation. And nothing refused a write past the declared
4 MiB budget below the private build root: macOS has no per-directory byte cap.

Supervisor-side accounting could substitute for host accounting, but only if no
domain member could evade or survive it — which §5.1 shows is false.

### 5.4 Also measured: inherited descriptors are not revoked

A seatbelt profile is evaluated at `exec`. A socket that was already connected
before the domain existed is not re-evaluated, and stayed writable from inside.
Network *denial* is real for new endpoints; *revocation* of a pre-existing one is
not available. The negative control (an unpassed descriptor failing `EBADF`)
confirms the agent does observe write failures, so "allowed" here is a real
observation.

---

## 6. Platform mechanism inventory

Every mechanism considered, with the support status that decides whether a
production hardened profile could depend on it. Mechanisms rejected without being
used are listed too: a considered-and-rejected mechanism is as much a result as
one that was used.

### 6.1 Deprecated — shipped, functional, withdrawn interface

| Mechanism | Note |
| --- | --- |
| `/usr/bin/sandbox-exec` | present and working on this host, but it is a thin wrapper over `sandbox_init`, which `<sandbox.h>` declares deprecated. Apple publishes no replacement for applying a dynamic profile to an arbitrary already-built binary. |
| `sandbox_init` / `sandbox_init_with_parameters` | declared deprecated in the public SDK header. The named built-in profiles it accepts are coarse. |
| `posix_spawn` `POSIX_SPAWN_START_SUSPENDED` + pre-exec sandbox call | the usual way to apply a profile without `sandbox-exec` still calls the deprecated `sandbox_init` in the child. It changes who calls the interface, not whether it is deprecated. |

### 6.2 Private — SPI, unpublished, or entitlement-gated

| Mechanism | Note |
| --- | --- |
| seatbelt profile language (version 1 S-expressions) | not a published, versioned interface. Operation names and filter semantics are discovered from the shipped profiles under `/usr/share/sandbox` and can change between releases without notice. **This is what every available verdict above rests on.** |
| Endpoint Security framework | could veto `exec` system-wide, but requires the `com.apple.developer.endpoint-security.client` entitlement, granted case by case, plus a system extension and full disk access. It is also a global authorizer, not a per-operation domain. |
| Mach task ports, `task_policy_set` | `task_for_pid` on another process requires elevated privilege and is blocked by SIP for protected processes. The policy interfaces set scheduling and QoS bands, not hard aggregate bounds. |

### 6.3 Supported

| Mechanism | Note |
| --- | --- |
| POSIX process group and session | the only grouping a plain macOS supervisor can create. A descendant leaves it with `setsid`, so it is not an unescapable domain. |
| `setrlimit` `RLIMIT_AS` / `NOFILE` / `NPROC` / `CPU` | per process, inherited by children as a fresh budget each. `RLIMIT_NPROC` is per user. None account over a process tree. |
| Disk image (`hdiutil`) as a size-bounded private volume | a fixed-size attached image does bound bytes written below the build root — the one aggregate quantity macOS can enforce without private interfaces. Costs an attach and detach per operation, and bounds nothing else. |
| Virtualization.framework guest | an unescapable domain with aggregate memory and CPU bounds and atomic destruction. The only public macOS mechanism that satisfies the three blocking classes, at the cost of a guest image in the TCB and a much larger per-operation setup. |

### 6.4 Conditional or unavailable

| Mechanism | Status | Note |
| --- | --- | --- |
| App Sandbox (`com.apple.security.app-sandbox`) | conditional | supported and non-deprecated, but applies to a signed, entitled bundle. It cannot be imposed on an arbitrary toolchain binary the manager did not build and sign, and its container model does not express an exact executable allowlist. |
| Filesystem quotas (`edquota`) | conditional | per user or group on a whole volume, requires root, not enabled on a default APFS install. Cannot bound what one operation writes below one directory. |
| `chroot` | conditional | requires root and is not a security boundary on macOS. |
| Linux cgroup v2 equivalent | unavailable | no `pids.max`, `memory.max` or `cgroup.kill` analogue exposed to third-party code. |
| Windows Job Object equivalent | unavailable | nothing with kill-on-close semantics that a descendant cannot leave. |

---

## 7. What Curator and csk can reuse

### 7.1 Reusable as-is, by both

- **The class inventory, guarantee mapping, and closed evidence record.**
  `internal/spec` and `internal/evidence` are a working transcription of the
  normative constants plus a validator that implements the error table. csk needs
  the same record shape; the guarantee→class mapping and the reduction rules are
  language-independent.
- **The reduction rules.** Inconclusive collapses to unavailable. `available`
  implies `applied`; a class that could be provided but was not installed in this
  operation is reduced to unavailable rather than claimed. A class missing from
  the observations is `unprobed`, never defaulted. These are the rules that make
  a record fail-closed, and they are the part most easily got wrong twice.
- **The probe shape.** Positive test + negative control + adversarial escape, per
  class. The negative control is not optional: without it, a profile that broke
  the agent entirely is indistinguishable from perfect enforcement. This harness
  found two places where that would have mattered.
- **The fail-closed sweep** as a conformance fixture: force each class
  unavailable in turn, require rejection before domain entry with the mapped
  diagnostic. It is executable, cheap, and platform-independent.
- **The exit-code contract**, including the separation of "the host cannot do
  this" (1) from "the harness broke" (2).
- **The measured platform verdicts in §4 and §5.** Neither implementation needs
  to re-derive them. A macOS implementation of either should reject `go-v1` with
  `hardened_profile_unsupported` at platform qualification, before any probe.

### 7.2 Reusable by Curator only

- `internal/seatbelt` (profile rendering, deterministic ordering, `subpath` vs
  `literal` selection, the loader-path set) and `internal/probe` are Go and would
  port to Curator directly if a macOS profile is ever pursued.
- The self-as-agent pattern: the probe binary is its own in-domain agent, so the
  exact executable allowlist has exactly one entry. Curator can do the same with
  a hidden worker mode. **csk cannot**, without a standalone immutable worker: a
  Python entrypoint makes the interpreter and the installed package tree part of
  the TCB.

### 7.3 Not reusable — must be re-derived per implementation

- Everything that depends on the seatbelt profile language. It is private and
  unversioned; a csk implementation calling the same `sandbox-exec` inherits the
  same deprecation and the same absence of a stability contract.
- The path-resolution discipline. A seatbelt filter matches the absolute path the
  kernel resolved, so a relative root or one reached through a symlink (`/tmp`
  and `/var` are symlinks on macOS) matches nothing and turns every probe into a
  denial — which looks exactly like perfect enforcement. `NewEnvironment`
  rejects unresolved paths for this reason. Any reimplementation must do the same
  or it will produce a confidently wrong "available" sweep.

### 7.4 Two traps worth writing down

- **`os/exec` and `/dev/null`.** A nil `Stdout`/`Stderr` makes `os/exec` open
  `/dev/null`, which a `(deny default)` profile denies. The failure surfaces as
  `EPERM` from `Start`, indistinguishable from the execution denial the probe is
  trying to observe. Every stream must be wired explicitly.
- **`PROT_EXEC` on Apple Silicon.** Mapping any file executable returns `EPERM`
  without a JIT entitlement — a signed system binary included. The `mmap`-exec
  escape therefore fails for a *platform* reason, not because the allowlist
  stopped it. The harness records that attribution rather than crediting the
  profile, and any reimplementation on a host where executable mappings are
  obtainable must expect that escape to succeed against an exec-only allowlist.

---

## 8. Relationship to the existing decision

This is measured confirmation of the macOS finding in
`.research/260727_go-v1-host-execution-decision.md`, which concluded on
documentation review that macOS could not provide a fail-closed profile. The
probes reach the same conclusion by measurement, and add three things that
review did not have:

1. **which** classes are actually available on macOS — six of eleven are (five
   substantive controls plus the probe class), and they are genuinely available,
   not approximations;
2. the **specific reason** each blocking class fails, with the exact escape that
   demonstrates it;
3. the loader/root-namespace finding (§5.2), which is a hard property of the
   backend rather than a profile-authoring gap.

Nothing here contradicts the decision. macOS must reject `go-v1` before starting
any worker or Go process.

## 9. Explicit non-claims

- This does **not** claim macOS enforces anything. Six available classes are an
  observation about `/usr/bin/sandbox-exec` on one host on one date.
- This does **not** qualify macOS. `qualification_status` is `unqualified` in
  every record the harness can emit.
- The available verdicts rest on a **private, unversioned** profile language
  reached through a **deprecated** interface. Neither is a basis for a production
  guarantee, independent of what was measured.
- No curator-spec text was changed. No Curator or csk production code was
  touched. Nothing was staged, committed, or published.

## 10. Artifacts

| Artifact | What it holds |
| --- | --- |
| `TASK-260729-3jmqgl_macos-hardened-capability-outcome.md` | this document |
| `TASK-260729-3jmqgl_evidence-packet.tar.gz` | the captured run: `host.txt`, `exit-codes.txt`, `evidence.json`, `report.json`, `report-fail-closed.json`, and the stdout/stderr of every case |
| `TASK-260729-3jmqgl_macos-hardened-probes-source.tar.gz` | the prototype source, tests and `capture-evidence.sh` |
