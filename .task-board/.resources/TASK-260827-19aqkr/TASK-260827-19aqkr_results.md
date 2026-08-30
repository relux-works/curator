# Outcome Evidence: TASK-260827-19aqkr (build-https-contract)

## Summary of Deliverable

1. Created `docs/build-https.md`: The operator HTTPS credential contract for external build repositories in Curator.
   - Shape, structure, tone, and register strictly mirror `docs/build-ssh.md`.
   - Adheres to the CocoaSkills / Curator prose style guide (`TASK-260827-19aqkr_cocoaskills-prose-style.md`): active voice, load-bearing short sentences, definition before consequences, problem before solution, no em-dashes or en-dashes, no guillemets, code formatting for all identifiers, no marketing register, no blacklist phrases.
   - Documents the three token sources (`git-credentials`, `keyring`, `token_env`), `build_https` scope grammar and longest-prefix matching, `curator config build-https` commands, environment overrides (`CURATOR_BUILD_HTTPS_TOKEN`, `CURATOR_BUILD_HTTPS_USERNAME`, `CURATOR_BUILD_HTTPS_HOST`), TTY vs off-TTY install-time precheck behavior, and credential broker fail-closed rules.

2. Updated `docs/build-ssh.md`: Added reciprocal cross-link pointing to `docs/build-https.md`.

## Empirical Grep Evidence from `internal/` Codebase

- **Stable Error Code**: `CodeIdentityInvalid` = `"build_repository_identity_invalid"`
  - `internal/buildrepo/admission.go:31`: `CodeIdentityInvalid = "build_repository_identity_invalid"`
  - `internal/buildrepo/admission.go:256`: `if request.Source.Transport == "https" && request.Tool.AskPass == "" { return nil, admissionError(CodeIdentityInvalid, "HTTPS requires a manager credential broker") }`
- **Credential Broker Flags & Env**:
  - `internal/buildrepo/admission.go:373`: `-c credential.helper=`, `-c core.askPass=` + askPass
  - `internal/buildrepo/admission.go:401`: `GIT_ASKPASS=` + tool.AskPass
  - `internal/buildrepo/admission.go:397`: `GIT_TERMINAL_PROMPT=0`
  - `internal/buildrepo/admission.go:376-377`: `-c http.followRedirects=false`, `-c http.sslVerify=true`, `-c http.proxy=`, `-c https.proxy=`
- **Git Executable & AskPass Resolution**:
  - `cmd/curator/main.go:1398-1401`: `tool.AskPass` resolved from `os.Getenv("GIT_ASKPASS")` or `os.Executable()`
- **Scope Grammar & Scope Matching**:
  - `internal/config/config.go`: `build_ssh` and `build_https` scope matching on lowercase host + `/`-separated segments, longest match wins.

## File Verification

- `head -n 25 docs/build-https.md` verified file creation and initial content.
- `grep -n -C 3 "build-https.md" docs/build-ssh.md` verified reciprocal cross-link at line 18 of `docs/build-ssh.md`.

## Definition of Done Verification

- [x] Every build-https claim verified against internal/ with grep evidence
- [x] Docs updated and consistent with current code
- [x] No discrepancies between code and description
- [x] Result linked as a new task-scoped outcome resource
- [x] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
