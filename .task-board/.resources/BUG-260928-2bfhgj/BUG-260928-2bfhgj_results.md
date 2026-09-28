# BUG-260928-2bfhgj results

## Root cause
- Evidence: run 36387486082, test-evidence-rose-air go-test.json: failures sit at worker_test.go:546 and :601. Both are the `expectFailure` calls on the **byte-flipped** binaries (`payload[len-1]^=0xff`: "curator-modified" and "curator-other"). The plain copy row (line 524) passes on rose-air, so the identity check itself works.
- rose-air: go1.25.5 darwin/arm64. On arm64 macOS the kernel enforces code signatures: a Mach-O whose signature no longer matches its bytes gets SIGKILL at exec or on page-in, before `main` runs. The worker dies, stdout closes, and the test sees EOF ("worker session channel closed"). So the test was really proving a kernel kill, not the identity refusal. Hosted macos-latest is arm64 too but stays green. That points to different trailing bytes in the image (linker/signature padding), so the flip fell outside the validated blob there. Stated bound: I could not reproduce this locally (the host is x86_64 and does not enforce signatures). I have no captured exit status from rose-air, because the old harness sent stderr to os.Stderr and never reported the exit state. The new diagnostic below will name it if this happens again.
- This is not a product defect and not a missing host control, so manager.md §3.1 does not apply. The fixture was not launchable on that platform.

## Fix (test only, internal/scriptworker/worker_test.go)
- New `tamperExecutableFixture`. On darwin it re-signs the copy ad hoc under a different identifier (`/usr/bin/codesign --force --sign - --identifier curator.test.tampered`). That gives a runnable substitute with different bytes, which is what a real attacker would build. Other OSes keep the last-byte flip. The helper fails the test if the bytes did not change. Both refusals still assert `CodeWorkerIdentityInvalid` and the interpreter-never-ran marker.
- `rawWorker` now captures stderr. `receive`/`send` failures append `exitReport()` (the Wait error, the ProcessState such as `signal: killed`, and the stderr text).

## Local evidence (zsh, darwin/amd64 15.7.4)
- `go test -count=1 -run 'TestScriptWorkerRejects(ForgedWorkerIdentity|SubstitutedManager)' -v ./internal/scriptworker`: both PASS
- `go test -count=1 ./internal/scriptworker`: exit 0
- `golangci-lint run ./internal/scriptworker/`: 0 issues, exit 0. gofmt clean.
- Diagnostic mutant (kill the worker before send at line 560) gives: `cannot send request: write |1: file already closed; worker exit: exec: Wait was already called (state signal: killed); worker stderr: ""`, so the kill is now named.
- No tamper mutant was run separately. The helper's bytes-equal guard fails the test if the tamper is a no-op.

## Unverified
- rose-air arm64 and hosted arm64 were not run by me. Test (rose-air) runs only on main pushes, so rose-air evidence before landing needs an orchestrator decision.

## CHANGELOG entry (for release prep)
- Tests: the scriptworker forged-identity and substituted-manager rows now build a runnable re-signed substitute on macOS, instead of a byte-flipped image the arm64 kernel kills at exec. Worker channel-closed failures now report the worker's exit status and stderr.
