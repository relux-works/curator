# Rework brief after RUN-260819-5a67e0

Implement every reviewer-required change before another handoff.

1. Replace the direct osCargoBuildRunner exec.CommandContext path for both fresh metadata and build. Both Cargo invocations must execute through the manager-owned closureexec causal chain under a committed build-scoped DerivationPermit. Publication and execution evidence must derive from issued Executor receipts, not constructed claims. Add negative regressions for attempted network, child process, undeclared read/write, input mutation, missing or widened permits, and audit mismatch before publication.

2. Replace or supplement the static source-string test with instrumentation that counts every process-start attempt across NewManager, Cargo registration, Rust build-tool registration, and equivalent discovery paths. Prove zero process starts before assurance preflight, committed C0, and the explicit permit boundary.

3. toolchain_registry.go must not infer Cargo 1.91.0 or commit ea2d97820c16195b0ca3fadb4319fe512c199a43 from a user-writable directory name plus arbitrary content. Bind the closed accepted descriptor to exact operator-approved executable/root byte identities or admitted filesystem metadata proving those facts, and reject other content before use. Retain filesystem-only discovery and no vendored compiled binaries.

4. Portable mode remains the functional default. Do not make Rust Build verified-only. Both portable and verified modes route through Executor. Portable receipts claim only established portable capabilities and actually observed declared outputs, immutable replay, and toolchain rechecks; they must not fabricate lossless process/read/write/network observation. Lossless rejection regressions use an injected verified test provider or boundary. Production verified mode without a compatible provider fails closed before session creation or process start.

Reviewer verdict: TASK-260811-3kbf3l_reviewer-verdict_RUN-260819-5a67e0.md. Preserve existing accepted functionality and rerun focused negative tests, race where relevant, make lint with CI-pinned v2.12.2, go vet ./..., go build ./..., go test -count=1 ./..., and board validation.