# SNAP-LOCK revision 3 validation

Current run reran the following standalone commands, with output redirected directly to task-local logs; no tee pipeline was used.

| Command | Actual exit | Evidence |
| --- | ---: | --- |
| `go test -race ./internal/envprofile ./internal/managerlock ./internal/transaction -run 'TestMaterializeCapture|TestPendingRecovery|TestRecoveryObserver|TestCaptureLock|TestAcquireExistingHomeOnly' -count=1 -timeout=3m` | 0 | focused-resume-01.log |
| `go build ./...` | 0 | build-resume-01.log (empty output) |
| `go vet ./...` | 0 | vet-resume-01.log (empty output) |
| `go test -overlay=.temp/SNAPLOCK-rev3/reviewer-overlay.json ./internal/envprofile -run '^TestReviewCapture' -count=1 -timeout=2m` | 0 | reviewer-resume-01.log; original reviewer-only tests retained via Go overlay |
| `git diff --exit-code HEAD -- ':(glob)**/*_test.go'` | 0 | retained-resume-01.log; all previously committed test files unchanged |
| `git diff --check` | 0 | diff-check-resume-01.log |
| `sh -n scripts/remote-gate.sh` | 0 | syntax-resume-01.log |
| `python3 .temp/SNAPLOCK-rev3/run-mutants.py` | 0 | mutants-resume-01.log; all ten standalone Go child commands exited 1 at named failures |
| `gh run view 38919350971 --json status,conclusion,headSha,url,jobs` | 0 | exact signed candidate hosted result: 20 successes, two optional skips |
| `gh run download 38919350971 -n test-evidence-ubuntu-latest -n test-evidence-macos-latest -n test-evidence-windows-latest -D .temp/SNAPLOCK-rev3/hosted-evidence` | 0 | downloaded terminal test records |

Full gate command `sh scripts/remote-gate.sh` was launched by the preceding producer run on this exact candidate after loading the prescribed agent Git environment. The complete saved log reports hosted success, and this run independently read the successful hosted conclusion and downloaded test artifacts. **The original shell exit code is unavailable**: its tool session cannot be recovered after operator cancellation. It is reported unknown, not 0. The complete hosted full suite is reused, not rerun, under the exact unchanged signed candidate and workflow identity, per the standing evidence-reuse rule. No missing or truncated test log is treated as passing.

Hosted full-suite test records (unique terminal test/subtest names, parents and children counted separately): Linux 57 pass / 0 skip; macOS 57 pass / 0 skip; Windows 49 pass / 6 skip. Native capture and file/keychain refusal all pass on each platform. Windows skips are the declared POSIX shell, executable-bit, unreadability and held-lock replacement limitations. The two optional workflow jobs are skipped, not runtime qualification.

Toolchain: Go 1.25.5 darwin/arm64 locally. Hosted workflow qualifies its pinned toolchains on Linux, macOS and Windows; no local replacement for required hosted status is claimed. Hosted Lint, interop properties and gate self-tests passed.

The snapshot tree was recomputed with a temporary index (`git read-tree HEAD`, `git add -A .`, `git write-tree` using only GIT_INDEX_FILE). Each command exited 0; exact tree equality, signature G, unchanged Story checkpoint and empty real index assertions passed. A redundant earlier per-file git-show comparison was stopped by its recorded PID after the stronger whole-tree equality check passed; that diagnostic exited 143 and is not passing evidence.

An initial read used incorrect spec paths and received three gh API exit-1 refusals. Corrected protocol paths each returned exit 0. Current curator-spec main is 1059b94b4135089d842bf017070b8ab47d79a536. Snapshot text and the targeted consistency contract are unchanged; newer local-node-source clauses do not alter this native capture admission seam.

## Current focused output

```text
ok  	github.com/relux-works/curator/internal/envprofile	32.904s
ok  	github.com/relux-works/curator/internal/managerlock	1.323s
ok  	github.com/relux-works/curator/internal/transaction	1.833s
ok  	github.com/relux-works/curator/internal/envprofile	3.086s
```

All ten mutation child commands, their exact patterns, actual exit-1 results and named failing tests are in the companion mutation evidence table and JSON. Raw failure dumps containing fixture paths are omitted from public artifacts.
