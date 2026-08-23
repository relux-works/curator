# Mandatory pre-C0 security gate

This is a binding acceptance constraint for TASK-260811-3kbf3l.

- No child process may start during manager construction, toolchain registration, or discovery before assurance preflight, committed C0, and an explicit derivation permit.
- Remove the current /usr/bin/xcrun subprocess from registerRustBuildToolchain and every equivalent ambient rustup, PATH, environment, or process-based discovery path.
- SDK, linker, Cargo, rustc, sysroot, and target stdlib selection must come from a closed operator/platform registry using filesystem evidence only, or from the same manager-owned closureexec causal chain after assurance preflight with explicit permit and receipt.
- Do not use hard-coded version-output claims unless exact installed bytes or admitted filesystem metadata prove them. Content fingerprints and proven registry facts are authoritative.
- Add a regression that observes zero process starts before C0 while build-tool registration occurs.
- Reviewer must inspect production source, execute the regression, and reject handoff if any pre-C0 subprocess or unsupported version claim remains.