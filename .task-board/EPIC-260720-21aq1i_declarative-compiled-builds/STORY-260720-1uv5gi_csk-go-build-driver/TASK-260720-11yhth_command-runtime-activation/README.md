# Harden runtime reuse and activate built commands

## Description
Make command materialization validate existing script runtime entries and expose immutable build-cache artifacts through self-contained project and global shims on Unix and Windows.

## Scope
Own src/csk/shims.py, narrowly scoped global-bin helpers, and focused shim and runtime tests. Existing script runtime reuse must verify every required active command path and rebuild incomplete or corrupt commit-keyed runtime state. Add a typed activation path for compiled artifacts selected by validated marker and receipt data. Preserve optional shell-profile behavior and do not run built outputs during install.

## Acceptance Criteria
An existing runtime directory is reused only when all required active script paths are regular, contained, and present; missing or wrong-type paths trigger staged replacement. Unix project and global activation targets the immutable cache artifact and forwards argv and exit status without profile setup. Windows .cmd activation quotes the executable safely, forwards %*, preserves ERRORLEVEL, and rejects command-name or path injection. Mixed script and build commands participate in the same collision and stale-shim rules. Install-time tests prove no built artifact is launched, while explicit post-install tests prove arguments and nonzero exit codes propagate. POSIX and Windows-focused pytest plus strict mypy pass.
