# TASK-260811-27xisf Stop-The-Line evidence

Run: `RUN-260817-d840aa`

Authoritative checkpoint: `GOAL-260817-709e3d` revision 1, resolved scope
`TASK-260811-27xisf`.

## Constraint

Reviewer requirement R1 and directive `nudge:b2e892` require authoritative,
OS-observed attempted and successful process, read, write, network, and output
events. The current Darwin implementation uses `sandbox-exec`. That API can
enforce a sandbox profile, but it returns only child exit status and output; it
does not expose a structured audit/event stream. The current code therefore
copies permit declarations into `Audit`, exactly as the reviewer found.

The public macOS API that supplies lossless process/file/network security
events is Endpoint Security. The installed SDK states at
`/Library/Developer/CommandLineTools/SDKs/MacOSX.sdk/usr/include/EndpointSecurity/ESClient.h:638`
that callers of `es_new_client` must have the Apple-granted
`com.apple.developer.endpoint-security.client` entitlement. The current
developer process and `task-board` launcher have no such entitlement.
`fs_usage` is not a viable substitute: its installed manual states that it
requires root privileges because it uses the kernel tracing facility.

An exploratory `sandbox-exec` process that handled a denied `/etc/passwd` read
and exited zero confirmed the security issue. The denial appeared only later
in the global unified kernel log (`Sandbox: sh(...) deny(1) file-read-data ...`),
not in the child result or a sandbox-exec audit stream. Unified logging is a
global, asynchronous, potentially dropped diagnostic channel; it cannot prove
the complete per-run set of allowed reads, successful child processes, writes,
network operations, and outputs. Treating absence of a log line as absence of
an operation would synthesize evidence and directly violate the accepted
contract and directive.

## Failed assumptions and attempts

1. `sandbox-exec` does not support a report modifier that produces a caller-
   owned audit stream; `(deny default (with report))` is rejected by the
   sandbox compiler.
2. `log stream`/`log show` can expose some asynchronous kernel deny messages,
   but not a complete, scoped audit of allowed and denied descendant activity.
3. `fs_usage` documents a root requirement and therefore cannot be embedded in
   the unprivileged Curator process.
4. Polling process trees, scanning output directories, child-emitted manifests,
   interposition, or copying permit declarations would be bypassable or
   incomplete and are specifically forbidden forced fits.

No product-code workaround was added in this run.

## Viable options

1. **Provide a signed privileged macOS observer (recommended if Darwin support
   is required).** Approve a separately owned Endpoint Security system
   extension/daemon, obtain Apple's entitlement, define its authenticated IPC
   and event-loss/fail-closed protocol, and authorize this task to integrate
   with it. This satisfies authoritative observation but adds deployment,
   signing, privilege, lifecycle, and ownership scope.
2. **Move the protected execution implementation to an approved platform with
   an available enforce-and-observe primitive.** Define Linux as the initial
   supported executor and supply a Linux validation runner plus the accepted
   kernel mechanism/privilege model. Keep Darwin fail-closed. This changes the
   delivery platform promise and cannot be assumed locally.
3. **Relax R1 to enforcement-only.** Continue with sandbox-exec and post-run
   output inspection. This is not recommended because it contradicts the
   accepted checkpoint contract, CGN18, the reviewer verdict, and the explicit
   directive.

## Exact decision or external input needed

Choose option 1 or 2 and provide the corresponding platform authority:

- for option 1, the entitled/signed Endpoint Security observer boundary and
  ownership/IPC contract; or
- for option 2, approval of Linux-only initial support, the intended kernel
  enforcement/observation mechanism, and a Linux validation environment.

Without one of those inputs, R1 cannot be implemented honestly. R2-R6 are
ordinary code rework, but completing them cannot make the task handoff-worthy
while the mandatory protected-execution trust boundary remains unavailable.

## Evidence commands

All commands were standalone and their real exit status is recorded:

- `task-board spawn goal "$TASK_BOARD_RUN_ID"` — exit 0; goal
  `GOAL-260817-709e3d` revision 1, scope `TASK-260811-27xisf`.
- `task-board spawn directives "$TASK_BOARD_RUN_ID"` — exit 0; directive
  `nudge:b2e892` observed and acknowledged.
- `man sandbox-exec | col -b` — exit 0; only profile selection and command
  launch are documented, with no audit/event output API.
- `sandbox-exec` with `(deny default (with report))` — sandbox-exec exit 65;
  expected exploratory failure: the compiler reports that `report` does not
  apply to the deny action.
- zero-exit sandboxed denied-read probe — probe exit 0; the denied read was
  handled by the child, demonstrating exit status is insufficient evidence.
- `log show --last 2m ...` — exit 0; found an asynchronous global kernel deny
  for the probe, but no complete per-run allowed/denied event set.
- `rg` over installed Endpoint Security headers — exit 0; found the mandatory
  entitlement declaration and `ES_NEW_CLIENT_RESULT_ERR_NOT_ENTITLED`.
- `man fs_usage | col -b` — exit 0; documents required root privileges.
- `id -u` — exit 0; current UID is 502 (unprivileged).

No test, lint, or build gate was rerun because this run deliberately made no
product-code change after the forced-fit boundary was established.
