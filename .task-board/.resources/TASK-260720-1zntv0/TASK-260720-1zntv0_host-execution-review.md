# TASK-260720-1zntv0 — independent host-execution decision review

Date: 2026-07-27  
Role: reviewer  
Reviewed artifact: `TASK-260720-1zntv0_host-execution-decision.md`  
Reviewed SHA-256: `e0fb6218cb0bd573db496bca30346277f8f9623a40d2b42d2b0814ecec291978`

## Verdict

**Stop-the-line boundary confirmed.**

The architecture decision packet is accepted as a sound, decision-ready
analysis artifact. The implementation task is not accepted: its current direct
Go host policy still cannot satisfy the native isolation and resource-control
acceptance criteria. Route `TASK-260720-1zntv0` to `blocked` until the exact
human architecture decision below is recorded.

This is not ordinary implementation rework. The current accepted candidate
requires direct execution of the fingerprinted Go binary and forbids any other
process in the graph, while the missing filesystem, network, descendant, and
aggregate resource gates require a trusted pre-exec containment boundary. The
remaining choice changes the normative process graph, supported-host promise,
cache/receipt identity, conformance vectors, and both manager implementations.

## Independent findings

### Accepted contract and current implementation

- Candidate `protocol/core.md §4.2` and `profiles/manager.md §2.2` require the
  manager to invoke fingerprinted Go directly and allow only Go plus
  fingerprinted GOROOT tools to start.
- Candidate `SECURITY.md` requires rejection rather than approximation when the
  fixed environment, network denial, source/context separation, native target,
  or process graph cannot be enforced, and requires filesystem/network plus
  host-supported resource limits.
- Candidate `protocol/core.md` explicitly permits either a new driver identifier
  or an explicit versioned cache-key policy when sandbox-relevant behavior
  changes. There is no local `1.0.0-rc.4` tag/ref, so retaining `go-v1` with a
  mandatory `execution_boundary: manager-worker-v1` policy revision is
  consistent only if the normative candidate, schemas, vectors, receipts, and
  claims are regenerated before release.
- Current Curator `guardedHostPolicy` validates metadata and delegates directly
  to `OSExecutor`. Unix and Windows platform hooks return `nil`; memory and
  process limits are not enforced; disk/artifact bounds are post-exit checks.
  The earlier implementation rejection remains valid.

### Recommended manager-worker contract

The fixed manager-owned worker is a legitimate security boundary rather than a
package-selected command surface:

```text
manager parent
  -> identity-verified manager worker
      -> fingerprinted Go
          -> fingerprinted GOROOT tools
```

The packet correctly excludes package influence over worker identity/mode,
argv, environment, CWD, roots, limits, permits, publication, and artifact
execution. Its single-session state machine preserves exactly one list and one
build, with parent validation between them. It correctly keeps csk's Python
interpreter and installed package tree in the TCB unless csk distributes a
standalone immutable worker.

The implementation and normative text must bind worker launch to an already
verified executable identity (descriptor/handle-backed execution or an exact
platform equivalent). A path recheck plus a later hash is not sufficient to
close replacement races. The packet already makes race-resistant launch,
identity mismatch, substitution, and replacement-race tests mandatory; this is
an implementation acceptance gate, not a reason to reject the decision packet.

### Platform conclusion

- **Hardened Linux can be positive only by capability, never by OS label.**
  Landlock can restrict filesystem execute/read/write and propagates policy to
  descendants; current ABIs add TCP and UDP rules. Cgroup v2 supplies a hard
  `memory.max`, hard subtree `pids.max`, contained delegation, and tree-wide
  `cgroup.kill`. User/mount/network/PID namespaces and a bounded tmpfs can close
  the remaining filesystem, network, descendant, and aggregate-disk surfaces.
  Any absent ABI, controller, delegation, namespace, mount, or syscall-control
  prerequisite must reject before worker or Go.
- **macOS is not a current positive profile.** App Sandbox is a signed,
  entitlement-based packaging contract. Child inheritance carries static
  rights rather than post-launch PowerBox rights, and current Apple
  documentation says user-selected access cannot be used to execute programs
  outside the app bundle/container/group. The external fingerprinted GOROOT
  model therefore does not fit without bundling/packaging changes. `setrlimit`
  is per-process (and `RLIMIT_NPROC` is UID-scoped), not an aggregate private
  descendant domain; it does not close the required memory/process/disk model.
- **Windows is not a current stable positive profile.** Job Objects provide
  descendant containment, kill-on-close, accounting, memory/time and active
  process controls, while AppContainer provides default-deny file/network
  isolation. Stable child-process mitigation is all-or-none, not an exact
  executable allowlist, and Job Objects do not impose a hard aggregate private
  storage quota. Microsoft's newer CreateProcessInSandbox API is explicitly
  experimental, Windows-11-only, dynamically resolved with no public header,
  and cannot be the rc.4 production contract.

The packet is therefore right to specify deterministic
`host_profile_unsupported` before worker/Go on macOS and Windows, while keeping
their parser/protocol/vector/compile lanes.

## Alternatives and tradeoffs

1. **Authorize the manager worker and initially support hardened Linux.**
   This is the recommended option. It adds a small auditable TCB component,
   enables pre-exec enforcement, and preserves package non-influence, at the
   cost of a normative graph/policy revision and Linux-only positive native
   builds.
2. **Preserve direct Go and narrow deployment to externally contained or custom
   native-launch profiles.** This keeps the existing graph but shifts
   containment ownership outside the manager or requires a higher-risk native
   process subsystem. It remains difficult to port to csk and still provides no
   reviewed positive macOS/Windows profile.

Environment-only denial, permission-bit checks, mock-only evidence, and
post-exit scans are not viable alternatives because they cannot undo a source
write, network connection, escaped process, or resource exhaustion.

## Board and provenance integrity

- The reviewed resource is attached to `TASK-260720-1zntv0`, has a task-scoped
  name, and its content hash is recorded above.
- The canonical Curator checkout has no tracked or staged changes.
- The task implementation worktree files were last changed on 2026-07-21; the
  candidate specification worktree files were last changed on 2026-07-20/21;
  the decision artifact was created on 2026-07-27. No product, candidate,
  release, claim, pin, ref, or provenance file changed during this analysis.
- Existing owner `TASK-260720-2g21eg` covers the csk driver; the cited
  curator-spec owners and downstream staging dependency exist. No speculative
  duplicate task is needed before authorization.
- Global `task-board validate` still reports 45 pre-existing board issues
  (legacy `EPIC-260712` broken links/status mismatches, one unsupported
  container link, and two orphan resources). None is introduced by or local to
  this task-scoped decision resource, but the board is not globally clean.

## Exact human decision required

> Authorize protocol 1.0.0-rc.4 to permit one identity-verified manager-owned
> sandbox worker in the fixed go-v1 process graph, retain the `go-v1` identifier
> with a new `manager-worker-v1` execution-policy cache revision, and ship real
> builds only on the fail-closed hardened Linux profile; macOS and Windows must
> reject go-v1 until independently reviewed native profiles satisfy the same
> executable, filesystem, network, and resource gates.

## Primary evidence

- Linux Landlock: https://cdn.kernel.org/doc/html/latest/userspace-api/landlock.html
- Linux cgroup v2: https://docs.kernel.org/admin-guide/cgroup-v2.html
- Apple App Sandbox: https://developer.apple.com/documentation/security/protecting-user-data-with-app-sandbox
- Apple sandbox file access: https://developer.apple.com/documentation/security/accessing-files-from-the-macos-app-sandbox
- Apple `setrlimit(2)`: https://developer.apple.com/library/archive/documentation/System/Conceptual/ManPages_iPhoneOS/man2/setrlimit.2.html
- Windows Job Objects: https://learn.microsoft.com/en-us/windows/win32/procthread/job-objects
- Windows AppContainer isolation: https://learn.microsoft.com/en-us/windows/win32/secauthz/appcontainer-isolation
- Experimental Windows sandbox creation: https://learn.microsoft.com/en-us/windows/win32/secauthz/createprocessinsandbox
