# TASK-260729-3jmqgl — macOS hardened capability probe outcome

**Task:** prototype-macos-hardened-capability-probes
**Story:** STORY-260728-327soo — fail-closed-cross-platform-build-execution
**Date:** 2026-07-29
**Role:** developer
**Revision:** rework cycle 1, after reviewer verdict `RUN-260729-6e3c8f`
**Status:** measured evidence; prototype only, no production integration

---

## 0. What changed in this revision

The previous revision was returned with one blocking acceptance gap and two
supporting ones. All four required items are addressed by executable
measurement, not by rewording.

| Required rework | What was done |
| --- | --- |
| Executable probes plus matched controls for CPU, memory/address space, process count and wall clock across descendants | New bound matrix (`internal/inside/bounds.go`) and reduction (`internal/probe/bounds.go`, `internal/probe/wallclock.go`). Each bound is installed on a real process which then tries to exceed it, with a matched unbounded control and a descendant that inherits the same bound. Descriptor and disk-byte probes retained unchanged in substance. |
| Derive the supervisor accounting/termination verdict from measured membership and atomic termination | `supervisorAccountingCheck` reduces the verdict from this run's session identifiers, teardown handle, teardown error and survivor observations. The old hard-coded `escapable` / `pass: false` is gone. `TestSupervisorAccountingIsDerivedFromThisRun` feeds it the numbers a host with an unescapable domain would report and requires the verdict to flip. |
| Prove deadline cancellation leaves no detached descendant; make an unbuildable probe binary fail the end-to-end suite instead of skipping it | New wall-clock probe measures the descendant tree after a real deadline fires, after the process-group teardown, and after the harness's own pid-directed cleanup, and reports the three separately. `TestProbeBinaryBuilds` fails outright; `requireAgent` now fails rather than skips on a missing binary. |
| Regenerate source, evidence and outcome on the macOS host with exact commands, versions and exit codes | §2, §3 and §10 below; evidence packet regenerated as `evidence-run-06`. |
| Platform inventory must not conclude wider than the executable observations | Every `setrlimit` resource is now its own inventory entry with its own measurement, and each entry carries `exercised` plus an `observation` naming the checks behind it. A mechanism nobody probed says so in as many words. |

The headline platform result is unchanged: macOS establishes 1 of the 6
guarantees, and the harness rejects fail-closed with exit 1. What changed is
that the resource-bound conclusions are now backed by executed measurements
rather than by a platform table.

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
| Date (UTC) | 2026-07-29T13:00:20Z |
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

All six cases ran as standalone processes. No command was piped through `tee`;
each exit status below is the real status of the named process.

| Case | Command | Exit | Expected |
| --- | --- | --- | --- |
| build | `go build -o hardened-probe ./cmd/hardened-probe` | 0 | 0 |
| list-classes | `hardened-probe --list-classes` | 0 | 0 |
| measure | `hardened-probe --work-dir … --evidence evidence.json --report report.json` | **1** | 0 or 1 |
| fail-closed sweep | `hardened-probe --work-dir … --report report-fail-closed.json --fail-closed-sweep --quiet` | **1** | 0 or 1 |
| assert-rejected | `hardened-probe --force-unavailable network-syscall-denial --expect rejected --quiet` | 0 | 0 |
| assert-established | `hardened-probe --force-unavailable network-syscall-denial --expect established --quiet` | **2** | 2 |
| leftover-processes | `pgrep -f <probe binary>` | count **0** | 0 |

`capture-evidence.sh` itself exited 0, meaning every case produced the exit
status it is documented to produce and no probe process survived the capture.

Exit-code contract:

| Code | Meaning |
| --- | --- |
| 0 | every capability applied, every guarantee established |
| 1 | rejected: at least one capability could not be established (fail-closed) |
| 2 | the harness could not produce a trustworthy record |

1 and 2 are deliberately distinct. An unusable harness is not evidence about the
host, and conflating them would let a broken run read as a platform verdict.

### 3.1 Verification commands run on this change

Each ran as its own process; the exit status recorded is the real one.

| Command | Exit |
| --- | --- |
| `go build ./...` | 0 |
| `go vet ./...` | 0 |
| `gofmt -l .` (empty output) | 0 |
| `golangci-lint run --config ../../.golangci.yml ./...` → `0 issues.` | 0 |
| `go test -count=1 ./...` | 0 |
| `go test -count=1 -cover ./...` | 0 |
| `./capture-evidence.sh <out>` | 0 |

Coverage, per package:

| Package | Coverage | Previous revision |
| --- | --- | --- |
| `cmd/hardened-probe` | 81.2% | 81.2% |
| `internal/evidence` | 99.3% | 99.3% |
| `internal/inside` | 87.4% | 92.5% |
| `internal/probe` | 85.2% | 84.4% |
| `internal/seatbelt` | 97.7% | 97.7% |
| `internal/spec` | 100.0% | 100.0% |

`internal/inside` fell because the new bound code runs mostly in child
processes, which coverage does not credit to the parent. It was brought back
above the target by driving the orchestration in-process (the runs that install
no limit are safe there) and by measuring the soft-limit escape against
`RLIMIT_CORE`, which has the same soft/hard semantics and bounds only a core
dump the test process is never going to write. The runs that install a real
limit stay in subprocesses on purpose: three of the four bounds can end the
process that hits them, and lowering a hard limit is irreversible.

Repository-root `golangci-lint run` remains broken for a pre-existing reason
unrelated to this task: it cannot load export data for the
`skill-go-testing-tools/tuitestkit` submodule dependency. The prototype module
is therefore linted directly, with the repository's own `.golangci.yml`.

The end-to-end suite takes about 4½ minutes, dominated by the two fail-closed
sweeps: each forces all eleven capability classes unavailable in turn and
performs a full host measurement for each. Every one of those measurements
includes a real CPU burn against a real limit and a real 3-second deadline. The
unforced measurement is taken once per package and shared by the tests that
assert different properties of it; repeating it per test would have multiplied
the wall time without observing anything new.

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
| `filesystem-view-restriction` | unavailable | seatbelt path-based denial | `/etc/passwd`, `/Users`, home and `.ssh` are denied, but the root namespace can still be enumerated — see §5.2 |
| `exec-path-allowlist` | **available** | `process-exec*` with literal allowances plus bounded `file-map-executable` | shell, bash, zsh, python3, `/bin/ls`, perl, dyld-as-program and a byte-identical self-copy in the build root all refused; the allowlisted path starts |
| `aggregate-resource-bounds` | unavailable | POSIX `RLIMIT_*` plus a supervisor deadline | no bound is accounted over the descendant tree; two of them cannot be installed at a build-sized value at all; the wall-clock deadline does not reach the tree — see §4.4 and §5.3 |
| `active-capability-probe` | **available** | probe domains via `/usr/bin/sandbox-exec` | every class produced a real observation in this operation; the missing-observation detector fires; an unprobed class cannot be claimed |

### 4.3 Fail-closed sweep

Every one of the eleven classes was forced unavailable in turn. All eleven
rejected before domain entry with `rejected_before: capability-probe`,
`diagnostic: hardened_capability_unavailable`, exit 1, and the forced class
reported `unavailable` / `not-applied`. `pass = true` for all eleven.

Injected verdicts are marked as injected in the report (`"forced unavailable by
--force-unavailable; this value was injected, not measured"`), so a forced value
can never be mistaken for a measurement.

### 4.4 The aggregate-bound matrix

This is the section the previous revision did not have. Every row is executed:
a real limit is installed on a real process, which then tries to pass it, with
a matched control that makes the identical attempt with no limit installed.

| Bound | Resource | Declared | Installs? | Binds the process that set it? | Domain gets the budget? | Descendant shares it? |
| --- | --- | --- | --- | --- | --- | --- |
| descriptors | `RLIMIT_NOFILE` | 64 | yes | yes — 61 opened, then refused | yes | **no** — child opened a fresh 61, aggregate 122 |
| bytes on disk | none exists | 4 MiB | n/a | **no** — wrote 8 MiB unrefused | n/a | n/a |
| CPU time | `RLIMIT_CPU` | 1000 ms | yes | yes — `SIGXCPU` at 1002 ms | yes | **no** — descendant burned a fresh 1003 ms, aggregate 2005 ms |
| address space | `RLIMIT_AS` | 256 MiB | **no** — `EINVAL` | not reachable | not reachable | not reachable |
| data segment | `RLIMIT_DATA` | 256 MiB | **no** — `EINVAL` | not reachable | not reachable | not reachable |
| process count | `RLIMIT_NPROC` | 4 | yes | yes — `EAGAIN` | **no** — refused at 0 of 4 | not evaluable: the bound refused the descendant's creation |
| wall-clock time | supervisor deadline | 3 s | yes | yes — root ended at 3.002 s | n/a | **no** — see §5.5 |

Every row above carries its matched control, and all of them passed:

- the unbounded control passed the declared budget in all four matrix bounds
  (CPU 1600 ms, memory 512 MiB twice, 6 processes), so a refusal measured
  against the budget is attributable to the bound and not to a stress that never
  reached anything;
- for `RLIMIT_AS` and `RLIMIT_DATA`, where the declared value was refused, the
  harness searched for the lowest value the kernel *does* accept and found
  **421 707 907 058 bytes — 1570 times the declared build budget**. The call
  works; it is the value that is refused. Without that control, `EINVAL` would
  be indistinguishable from a broken instrument;
- the descriptor bound's negative control shows the limit binds the process that
  set it, so the aggregate finding is not an artifact of a limit that never
  applied.

A second escape was measured for every installable bound: **a domain member can
raise its own soft limit back to the hard limit it inherited** (`raised`, for
both `RLIMIT_CPU` and `RLIMIT_NPROC`). The matched control lowers the hard limit
too and the same raise is then refused with `EPERM`, so the difference is
attributable to the hard limit. This means a supervisor that lowers only soft
limits has bounded nothing the bounded process does not consent to.

---

## 5. The findings that block macOS

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

### 5.3 Not one resource bound is accounted over the descendant tree

Measured for six quantities. There are three distinct failure modes, and they
matter separately because they call for different answers.

**A bound that binds one process and hands each descendant a fresh copy.**
`RLIMIT_NOFILE` and `RLIMIT_CPU` both do exactly what POSIX says: the limit
binds the process that set it, and a child inherits the same *number* as a new
budget. The domain root opened 61 descriptors and its child opened 61 more under
a single declared 64; the domain root burned 1002 CPU-milliseconds and its
descendant burned a fresh 1003 under a single declared 1000. The bound is real
and per-process; the guarantee needs it per-domain, and no arrangement of
per-process limits adds up to one.

**A bound that cannot be expressed at a build-sized value at all.** This kernel
refuses `setrlimit` for both `RLIMIT_AS` and `RLIMIT_DATA` at 256 MiB with
`EINVAL`, and the measured floor is roughly 392 GiB — 1570× the declared budget,
and above the address space the calling process has already reserved. There is
no memory bound to escape from, because there is no memory bound. (On Darwin
`RLIMIT_AS` is `RLIMIT_RSS`; `RLIMIT_DATA` bounds the data segment and would not
bound an `mmap`-based allocator even if it installed.)

**A bound accounted at the wrong scope.** `RLIMIT_NPROC` installs cleanly and
does refuse — but it refused the domain's *first* descendant under a declared
budget of 4, because it is accounted per real user id and the budget was already
spent by processes that are not domain members. A limit whose denominator is the
whole login session cannot bound one build operation.

**And storage is not bounded at all.** Nothing refused a write past the declared
4 MiB budget below the private build root; the domain wrote 8 MiB. macOS has no
per-directory byte cap.

Supervisor-side accounting could substitute for host accounting, but only if no
domain member could evade or survive it. That verdict is now **derived from the
same run** rather than declared: this run measured domain session 89776 against
detached-descendant session 95921 (membership renounced) and a survivor after a
successfully issued teardown, so supervisor-side accounting is reduced to
`escapable`. On a host where membership held and the teardown left no survivor,
the same reduction would report `unescapable`.

### 5.4 Also measured: inherited descriptors are not revoked

A seatbelt profile is evaluated at `exec`. A socket that was already connected
before the domain existed is not re-evaluated, and stayed writable from inside.
Network *denial* is real for new endpoints; *revocation* of a pre-existing one is
not available. The negative control (an unpassed descriptor failing `EBADF`)
confirms the agent does observe write failures, so "allowed" here is a real
observation.

### 5.5 A deadline ends one process, not a build

The wall-clock bound is the one aggregate bound no resource limit expresses, so
it has to be implemented as a supervisor deadline. Measured with a real 3-second
deadline against a domain that was still working when it arrived:

1. **The deadline ends the process the supervisor started.** The domain root ran
   3.002 s and was cancelled. The control — a domain whose work finishes inside
   the deadline — exited on its own with status 0 after 36 ms, so the
   termination is attributable to the deadline and not to a process that was
   going to exit anyway.
2. **It does not reach the descendants.** Both the attached and the detached
   descendant were still running after cancellation. Cancelling a supervised
   process signals that process alone, never the tree below it. This is a
   property of process cancellation, not of Go: the supervisor holds a handle to
   one process.
3. **A process-group signal reaches the descendant that stayed.** The attached
   descendant was gone after a group-directed `SIGKILL`, which is the negative
   control proving the teardown lands. The detached one survived it — the same
   escape as §5.1.
4. **So a deadline leaves a survivor.** After the deadline and after everything
   a plain macOS supervisor can issue, one domain member was still running.

The harness then kills that survivor by pid and records, separately, that
nothing remained afterwards. The two are kept as distinct checks on purpose: a
production implementation **cannot** rely on the second, because it only reaches
descendants the supervisor already recorded, and a descendant that forks after
the deadline fires is not one of them. Reporting them as a single "cleaned up"
line would have hidden exactly the gap that matters.

The evidence capture re-checks the same property from outside, with
`pgrep -f <probe binary>` after every case: count 0.

---

## 6. Platform mechanism inventory

Every mechanism considered, with the support status that decides whether a
production hardened profile could depend on it. Mechanisms rejected without
being used are listed too: a considered-and-rejected mechanism is as much a
result as one that was used.

**Every entry in the emitted report now carries `exercised` and an
`observation`.** A mechanism this run probed names the checks that measured it
and what they saw; one that was only considered says, in the artifact itself,
that its status is a reading of the published interface and not a measurement.
That split is what keeps the inventory from concluding wider than the probes.

Exercised in this run (9 of 20): `/usr/bin/sandbox-exec`, the seatbelt profile
language, POSIX process group and session, the supervisor deadline, and each of
`RLIMIT_NOFILE`, `RLIMIT_CPU`, `RLIMIT_AS`, `RLIMIT_DATA`, `RLIMIT_NPROC`
separately.

### 6.1 Deprecated — shipped, functional, withdrawn interface

| Mechanism | Note |
| --- | --- |
| `/usr/bin/sandbox-exec` | present and working on this host, but it is a thin wrapper over `sandbox_init`, which `<sandbox.h>` declares deprecated. Apple publishes no replacement for applying a dynamic profile to an arbitrary already-built binary. **Exercised.** |
| `sandbox_init` / `sandbox_init_with_parameters` | declared deprecated in the public SDK header. The named built-in profiles it accepts are coarse. Not exercised directly. |
| `posix_spawn` `POSIX_SPAWN_START_SUSPENDED` + pre-exec sandbox call | the usual way to apply a profile without `sandbox-exec` still calls the deprecated `sandbox_init` in the child. It changes who calls the interface, not whether it is deprecated. Not exercised. |

### 6.2 Private — SPI, unpublished, or entitlement-gated

| Mechanism | Note |
| --- | --- |
| seatbelt profile language (version 1 S-expressions) | not a published, versioned interface. Operation names and filter semantics are discovered from the shipped profiles under `/usr/share/sandbox` and can change between releases without notice. **This is what every available verdict above rests on. Exercised.** |
| Endpoint Security framework | could veto `exec` system-wide, but requires the `com.apple.developer.endpoint-security.client` entitlement, granted case by case, plus a system extension and full disk access. It is also a global authorizer, not a per-operation domain. Not exercised. |
| Mach task ports, `task_policy_set` | `task_for_pid` on another process requires elevated privilege and is blocked by SIP for protected processes. The policy interfaces set scheduling and QoS bands, not hard aggregate bounds. Not exercised. |

### 6.3 Resource bounds — one entry per resource, each with its own measurement

The previous revision grouped `RLIMIT_AS`, `RLIMIT_NPROC` and `RLIMIT_CPU` with
the measured `RLIMIT_NOFILE` under a single "supported" heading. They do not
behave the same way, and only one of them had been measured. Each now stands on
its own observation.

| Mechanism | Status | Measured on this host |
| --- | --- | --- |
| `setrlimit RLIMIT_NOFILE` | supported | binds the process that set it (61 of a declared 64); a child gets a fresh full budget, aggregate 122. Soft limit raisable back to the hard limit by the bounded process. |
| `setrlimit RLIMIT_CPU` | supported | binds via `SIGXCPU` at 1002 ms of a declared 1000 ms; a descendant gets a fresh 1003 ms. Whole seconds only, so 1 s is the finest bound expressible. Soft limit raisable back to hard. |
| `setrlimit RLIMIT_AS` (`RLIMIT_RSS` on Darwin) | **conditional** | **no build-sized bound can be installed.** `EINVAL` at 256 MiB; lowest accepted value ≈ 392 GiB. |
| `setrlimit RLIMIT_DATA` | **conditional** | same refusal, same floor. Bounds the data segment, not mapped memory. |
| `setrlimit RLIMIT_NPROC` | **conditional** | installs, but is accounted per real user id: refused the domain's first descendant under a declared budget of 4. |
| Supervisor deadline (cancellation + process-group signal) | supported | the only way to bound wall-clock time. Ends the supervised process; does not reach its descendants; a group signal reaches those that stayed; a detached one survives everything. **Exercised.** |

### 6.4 Other supported mechanisms (considered, not exercised)

| Mechanism | Note |
| --- | --- |
| POSIX process group and session | the only grouping a plain macOS supervisor can create. A descendant leaves it with `setsid`. **Exercised** — this one is measured. |
| Disk image (`hdiutil`) as a size-bounded private volume | a fixed-size attached image does bound bytes written below the build root — the one aggregate quantity macOS can enforce without private interfaces. Costs an attach and detach per operation, and bounds nothing else. Not exercised. |
| Virtualization.framework guest | an unescapable domain with aggregate memory and CPU bounds and atomic destruction. The only public macOS mechanism that satisfies the three blocking classes, at the cost of a guest image in the TCB and a much larger per-operation setup. Not exercised. |

### 6.5 Conditional or unavailable

| Mechanism | Status | Note |
| --- | --- | --- |
| App Sandbox (`com.apple.security.app-sandbox`) | conditional | supported and non-deprecated, but applies to a signed, entitled bundle. It cannot be imposed on an arbitrary toolchain binary the manager did not build and sign, and its container model does not express an exact executable allowlist. |
| Filesystem quotas (`edquota`) | conditional | per user or group on a whole volume, requires root, not enabled on a default APFS install. Cannot bound what one operation writes below one directory. |
| `chroot` | conditional | requires root and is not a security boundary on macOS. |
| Linux cgroup v2 equivalent | unavailable | no `pids.max`, `memory.max` or `cgroup.kill` analogue exposed to third-party code. |
| Windows Job Object equivalent | unavailable | nothing with kill-on-close semantics that a descendant cannot leave. |

None of these five was exercised; all five say so in the emitted report.

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
  the agent entirely is indistinguishable from perfect enforcement.
- **The bound-probe shape (new).** For a resource bound the three-process form is
  the reusable part: a *bounded* process that installs the declared limit and
  tries to pass it, an *unbounded control* that makes the identical attempt, and
  a *nested descendant* that inherits the limit and tries again. The aggregate
  question is answered by arithmetic on what the first and third reached. Every
  bounded attempt must run in its own process — three of the four bounds can end
  the process that hits them, and a measurement that dies with its measurer is
  not a measurement.
- **The two escape questions per bound.** Does a descendant get a fresh budget,
  and can a member raise its own soft limit back to the hard limit it inherited?
  Both were real escapes here, and the second is easy to miss: a supervisor that
  lowers only soft limits has bounded nothing.
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
  `literal` selection, the loader-path set), `internal/probe` and
  `internal/inside` are Go and would port to Curator directly if a macOS profile
  is ever pursued.
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
- The `RLIMIT_*` resource numbers. Go's `syscall` package does not export
  `RLIMIT_NPROC` on darwin; it is 7 in `<sys/resource.h>` and is transcribed with
  that source named. A resource number discovered by guessing would silently
  bound something other than what the report claims.

### 7.4 Traps worth writing down

- **`os/exec` and `/dev/null`.** A nil `Stdout`/`Stderr` makes `os/exec` open
  `/dev/null`, which a `(deny default)` profile denies. The failure surfaces as
  `EPERM` from `Start`, indistinguishable from the execution denial the probe is
  trying to observe. Every stream must be wired explicitly.
- **`PROT_EXEC` on Apple Silicon.** Mapping any file executable returns `EPERM`
  without a JIT entitlement — a signed system binary included. The `mmap`-exec
  escape therefore fails for a *platform* reason, not because the allowlist
  stopped it. The harness records that attribution rather than crediting the
  profile.
- **`RLIMIT_CPU` counts whole seconds, `getrusage` reports microseconds.** An
  earlier draft of this revision declared the CPU bound in milliseconds and
  handed the same number straight to `setrlimit`, installing a limit a thousand
  times larger than the declared one. Every CPU measurement came back unrefused
  and looked like a clean platform finding. The unit conversion is now one named
  function with its own test, and every bound kind asserts that its stress
  ceiling sits above its declared budget — without that, a stress stopping at its
  ceiling is indistinguishable from one the bound refused.
- **Lowering a hard limit is irreversible.** A process that lowers `rlim_max`
  cannot raise it again, so a supervisor must set limits in the child before
  `exec`, never in a process it intends to reuse. This is also why the escape
  probe runs last, in a process with nothing left to measure.
- **Coverage and subprocesses.** Code that only ever runs in a child process is
  not credited to the parent's coverage profile. That is a measurement artifact,
  not dead code, and the fix is to drive the safe paths in-process rather than to
  stop using subprocesses — the subprocess is what makes the bound measurable at
  all.

---

## 8. Relationship to the existing decision

This is measured confirmation of the macOS finding in
`.research/260727_go-v1-host-execution-decision.md`, which concluded on
documentation review that macOS could not provide a fail-closed profile. The
probes reach the same conclusion by measurement, and add four things that
review did not have:

1. **which** classes are actually available on macOS — six of eleven are (five
   substantive controls plus the probe class), and they are genuinely available,
   not approximations;
2. the **specific reason** each blocking class fails, with the exact escape that
   demonstrates it;
3. the loader/root-namespace finding (§5.2), which is a hard property of the
   backend rather than a profile-authoring gap;
4. the per-resource bound results (§4.4, §5.3, §5.5) — including that two of the
   memory limits cannot be installed at a build-sized value at all, that
   `RLIMIT_NPROC` is accounted at the wrong scope, and that a wall-clock deadline
   ends one process rather than a build.

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
- The measured soft-limit floor for `RLIMIT_AS` and `RLIMIT_DATA` is a property
  of this kernel *and* of the calling process's existing address space. It is
  reported as what was measured, not as a documented constant.
- No curator-spec text was changed. No Curator or csk production code was
  touched. Nothing was staged, committed, or published.

## 10. Artifacts

| Artifact | What it holds |
| --- | --- |
| `TASK-260729-3jmqgl_macos-hardened-capability-outcome.md` | this document |
| `TASK-260729-3jmqgl_evidence-packet.tar.gz` | the captured run: `host.txt`, `exit-codes.txt`, `leftover-processes.txt`, `evidence.json`, `report.json`, `report-fail-closed.json`, and the stdout/stderr of every case |
| `TASK-260729-3jmqgl_macos-hardened-probes-source.tar.gz` | the prototype source, tests and `capture-evidence.sh` |
