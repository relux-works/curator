# TASK-260825-3n4bjj — HTTPS broker and fetch wiring

## Delivered behavior

- Materializes a manager-binary askpass wrapper and secret-free JSON state for one authenticated HTTPS fetch.
- Binds the resolved host, username, and secret to the selected repository only.
- Passes state and secret only to the `git fetch` process tree; other Git children receive neither.
- Points both `GIT_ASKPASS` and command-line `core.askPass` at the same wrapper.
- Answers only Git's exact username/password prompts for the pinned host and otherwise exits silently.
- Preserves pinned source URL, TLS verification, disabled redirects, and the anonymous HTTPS argv/environment path.
- Keeps all secret-bearing diagnostic representations redacted.

## Tests and validation

| Command | Exit | Evidence |
| --- | ---: | --- |
| Focused broker, real TLS Git repository, fetch-environment, anonymous tests | 0 | `go test ./internal/buildrepo -run 'TestHTTPS|TestPrivateHTTPS|TestSelectedHTTPS|TestAnonymousHTTPS' -count=1` |
| Focused resolver/install wiring tests | 0 | `go test ./internal/install -run 'BuildHTTPS|ResolvedHTTPS' -count=1` |
| Full Go suite | 0 | `.temp/TASK-260825-3n4bjj/go-test-full-01.log` |
| Full lint | 0 | `golangci-lint run ./...` (`0 issues`) |
| Native curator build | 0 | `.temp/TASK-260825-3n4bjj/curator-build` |
| Windows amd64 cross-build | 0 | `.temp/TASK-260825-3n4bjj/curator-windows-amd64.exe` |
| Scoped diff check | 0 | Changed task files only |

## Non-authoritative red attempts

- The first broad three-package test checkpoint was interrupted after prolonged contention with several unrelated full-suite processes; its real exit code was 1. The later standalone full suite passed at exit 0.
- The first two scoped lint attempts exited 1 while the new executable wrapper needed explicit `gosec` justification; the corrected full lint passed at exit 0.
- An unscoped `git diff --check` exited 2 on pre-existing whitespace in unrelated board spawn logs. The task-scoped diff check passed at exit 0.
