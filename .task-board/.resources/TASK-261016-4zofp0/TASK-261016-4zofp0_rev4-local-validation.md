# Revision 4 local execution evidence

Candidate snapshot: 3bdf4c4f77a89afc12c06c242f69cd3d16deda9e.
Local environment: darwin/arm64, Go 1.25.5. No vendor runtime qualification.
All commands run directly without tee or a pipeline. Logs are retained in the run's ignored scratch directory; sanitized evidence is attached here.

| Command | Exit | Observation |
|---|---:|---|
| `go test ./internal/envprofile ./internal/transaction ./internal/managerlock -run 'Test(MaterializeCapture\|CredentialProjection\|CredentialStorePolicy\|PendingRecovery\|AcquireExistingHomeOnly\|CaptureLock)' -count=1 -timeout=4m` | 0 | Final focused identity, includes all new cache tests |
| `go test -overlay .temp/SNAPLOCK-rev4/reviewer-overlay.json ./internal/envprofile -run '^TestReviewCapture' -count=1 -timeout=2m` | 0 | Unchanged reviewer tests, original names retained |
| `go build ./...` | 0 | Current candidate compiles |
| `make lint` | 2 | Shared cache referenced absent neighboring checkouts; not a pass |
| `GOLANGCI_LINT_CACHE="$PWD/.temp/SNAPLOCK-rev4/lint-cache" make lint` | 0 | Isolated cache, zero issues |
| `git diff --check` | 0 | No whitespace errors |
| `python3 .temp/SNAPLOCK-rev4/run-mutants.py` (initial copy) | 143 | Stopped before any mutant while copying unrelated artifacts |
| `python3 .temp/SNAPLOCK-rev4/run-mutants.py` (bounded copy) | 0 | Twelve real expected-red test commands, each exit 1 |

The preliminary focused command `go test ./internal/envprofile -run 'Test(MaterializeCapture\|CredentialProjection\|CredentialStorePolicy)' -count=1 -timeout=3m` also exited 0, but predates the two new release-cache tests and is not substituted for the final focused run.

Reviewer tests `TestReviewCaptureLateJournal` and `TestReviewCaptureRejectsMutationFlags` are retained byte-for-byte from revision 2 and replayed through the overlay; no previously committed test was removed or renamed. The overlay is supplemental evidence, not a claim that the attachment was committed by this producer.

Full hosted gate subsequently exited 0. Exact command `sh scripts/remote-gate.sh`; signed snapshot and the 20 successful / 2 conditional skipped jobs are recorded in TASK-261016-4zofp0_rev4-full-gate.md.

## focused-01.log

```text
ok  	github.com/relux-works/curator/internal/envprofile	164.365s
ok  	github.com/relux-works/curator/internal/transaction	2.388s
ok  	github.com/relux-works/curator/internal/managerlock	0.411s
```

## reviewer-tests-01.log

```text
ok  	github.com/relux-works/curator/internal/envprofile	15.744s
```

## build-01.log

```text
```

## lint-02.log

```text
golangci-lint run
0 issues.
```
