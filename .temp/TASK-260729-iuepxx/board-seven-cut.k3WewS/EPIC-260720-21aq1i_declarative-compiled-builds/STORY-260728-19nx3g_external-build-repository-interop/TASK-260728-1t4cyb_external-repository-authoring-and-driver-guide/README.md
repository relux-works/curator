# TASK-260728-1t4cyb: external-repository-authoring-and-driver-guide

## Description
Publish rc.5 authoring and operator guidance for repository-backed CLI dependencies plus an evidence-based admission guide for future language drivers. Explain the safe repository descriptor and command ownership model without presenting wrappers, arbitrary build commands, or package-controlled signing as supported escape hatches.

## Scope
Cross-project authoring examples, monorepo target guidance, lock/tag and access behavior, substitutions, audit/cache/offline/PATH/signing explanation, troubleshooting, and future Rust/Swift/Kotlin-JVM/C-C++/.NET driver threat-review checklist.

## Acceptance Criteria
Examples validate against schema 7 and use command keys plus logical repository targets; guidance clearly distinguishes syntax-only warning from installation failure and protected offline reuse; repository markers cannot choose binary names, output paths, argv, environment, credentials, or signing; each future language requires a separately versioned closed driver covering build scripts/plugins/macros/dependency/network/native inputs and deterministic artifacts; unsupported languages are not advertised as generic equivalents to Go; docs tests and links pass.
