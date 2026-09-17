# TASK-260910-1952mz — hosted gate failure on Change Request revision 1 (run 35174707552)

Extracted by the orchestrator from the CI evidence artifacts of
https://github.com/relux-works/curator/actions/runs/35174707552 (macOS lanes
green; Test (ubuntu), Race (ubuntu) and Test (windows) red). The only failing
test is `internal/shell` `TestShellHookRefusesHostileCheckout`:

## Linux (`/usr/bin/sh` is dash), subtest `posix`
```
shell_hook_trust_test.go:387: sh activation: exit status 2
stderr: /usr/bin/sh: 183: /tmp/.../hook: Syntax error: "(" unexpected (expecting "fi")
```
The emitted POSIX hook is not POSIX: line 183 of the generated hook uses a
bash-only construct (arrays, `function name()`, `[[ ]]`, `<(…)`, `${var,,}`,
… — inspect the generated text at that line). macOS `sh` is bash, so it passed
there. The manager profile requires the POSIX hook to run under `sh`
(dash-compatible) and under Git Bash; keep it strictly POSIX (`[ ]`,
`case`, no arrays; `sha256sum` / `shasum -a 256` / `openssl dgst -sha256`
fallbacks via `command -v`).

## Windows (Git Bash `sh`, PowerShell), subtests `posix` and `powershell`
```
posix:      warning count = 0, want 1  (the stderr does contain the
            shell_hook_env_unapproved line and PROBE-1)
powershell: stderr lacks the activation probe (the stderr does contain the
            shell_hook_env_unapproved line and PROBE-1)
```
The hook emitted the warning and the probe, but the test's counting/matching
does not see them on Windows — most likely CRLF line endings (`\r\n`) in the
captured stderr, or the MSYS path spelling (`/c/Users/...` vs `C:\Users\...`)
in the expected warning text. Normalize the captured output (`\r\n` → `\n`)
and compare the warning by its diagnostic id + a path-agnostic pattern, or
compare against the path spelling the hook itself prints on that platform.

## What to do in this rework
Fix the generated POSIX hook so it runs under dash and Git Bash, fix the
Windows assertions as above, re-run `go test ./internal/shell/...` locally
(macOS) and hand off again; the runtime re-runs the hosted gate. The board
Change Request stays revision-based: this is rework of revision 1.
