# TASK-260720-1zntv0 host-policy rework blocker

## Constraint and evidence

The requested native rework cannot satisfy all three simultaneous invariants on the supported Ubuntu, macOS, and Windows CI platforms: start only the fingerprinted Go executable with no launcher or helper; install per-child read-only source and toolchain, private-write-only filesystem, network denial, and executable allowlist before package-controlled compiler input is processed; and keep valid real-toolchain builds working on every supported platform.

Go 1.25 os/exec exposes only syscall.SysProcAttr. On Darwin that surface has chroot, credentials, session, and process-group fields, but no per-child sandbox, network, filesystem-rule, or resource-limit hook. The public macOS sandbox_init API confines the current process, while sandbox-exec is a separate deprecated launcher and therefore widens the exact executable graph. App Sandbox requires signing and entitlements for the Curator process and changes CLI packaging and file-access semantics. On Linux, process groups and namespaces are exposed, but the required mount, Landlock or seccomp, cgroup, and rlimit setup must run in the child before exec or through a privileged broker; os/exec has no supported arbitrary pre-exec hook and namespace availability is host-policy dependent. On Windows, the standard Go launch surface cannot atomically create the child suspended, assign a Job Object and AppContainer or restricted token, install filesystem grants, then resume while preserving os/exec semantics; implementing this needs a custom broker or a materially new native process subsystem. Post-start monitoring has races and cannot undo a network connection or source write.

The existing declarative adapter therefore must not be papered over with mocks, permission-bit checks, environment flags, or post-exit disk scans. Those mechanisms cannot prove native denial and reproduce the reviewer finding in TASK-260720-1zntv0_review-verdict.md.

## Failed assumptions and attempts

The assumption that OSExecutor plus build-tagged validation could enforce the complete HostExecution contract is false. Process groups can improve tree-wide termination, and Windows Job Objects or Linux cgroups can improve resource accounting, but none closes the required cross-platform filesystem and network boundary without child-side setup, a broker or launcher, special privileges, or packaging changes. Failing closed on macOS and Windows would honor security but contradict the acceptance requirement that valid real fixtures build on supported CI platforms.

## Viable options

1. Amend the protocol and task to permit one operator-trusted manager-owned sandbox launcher or broker in the fixed process graph. This enables platform adapters to install sandbox and resource controls before exec, but changes the normative executable graph and requires new conformance vectors and security review.
2. Narrow supported go-v1 build hosts to a platform and deployment profile with specified sandbox primitives, and fail closed elsewhere. This preserves the direct Go graph but removes valid-build support from current macOS and Windows CI and requires explicit product compatibility changes.
3. Accept environment, static graph validation, read-only modes, and post-exit checks as sufficient. This keeps portability but contradicts the security profile and the independent review; it is not recommended.
4. Make Curator itself run inside an externally provisioned sandbox or signed App Sandbox container. This moves enforcement outside godriver and changes installation, signing, privileges, and source-access ownership.

## Recommendation and decision required

Recommend option 1: explicitly add one operator-trusted, manager-owned sandbox boundary to the protocol process graph, then design and test platform adapters against that revised contract. The exact human or architecture decision needed is whether to relax the no-extra-executable rule for that trusted boundary, or instead to narrow supported build hosts and define which platform profile remains supported. No product code was changed during this rework because either choice changes the architecture and acceptance criteria.