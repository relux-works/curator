# Windows lane ownership map — derived from real CI evidence

Source: `test-evidence-windows-latest` (artifact 8789366268) from run 30620739038,
job `Test (windows-latest)` on Curator PR 10. Downloaded and parsed from
`test/go-test.json`, not read from a job summary.

`windows-latest` is a first-class lane in the Test matrix (`ci.yml:66`) that had
never produced a test result until PR 10, because `go vet` aborted the job first.
Every failure below is pre-existing and newly visible, not introduced by PR 10.

## Failing top-level cases per package

| package | failures | owner |
|---|---|---|
| `internal/install` | 60 | BUG-260731-27h1yc |
| `internal/install/atomicity` | 8 | BUG-260731-27h1yc |
| `cmd/curator` | 8 | **none** |
| `internal/transaction` | 5 | BUG-260731-lepevi |
| `internal/managerlock` | 2 | **none** |
| `internal/buildsource` | 2 | BUG-260731-27h1yc |
| `internal/buildcache` | 2 | **none** |
| `internal/staging` | 1 | **none** |
| `internal/runtimestore` | 1 | BUG-260731-11bpa4 |
| `internal/godriver` | 1 | BUG-260731-lepevi |
| `internal/globalbins` | 1 | **none** |

Total 91 failing top-level cases across 11 packages.

## Gap 1 — five packages have no owner

`cmd/curator`, `internal/buildcache`, `internal/globalbins`, `internal/managerlock`
and `internal/staging` carry 14 failing cases and appear in no board item.

Every existing bug is package-scoped by its own AC:

- BUG-260731-11bpa4 → `internal/runtimestore` only
- BUG-260731-lepevi → `internal/godriver`, `internal/transaction`, `cmd/curator` (Linux lane)
- BUG-260731-27h1yc → `internal/buildsource`, `internal/install`, `internal/install/atomicity`

So by construction no item can close these five, and `Test (windows-latest)` stays
red after all three land.

Note on `cmd/curator`: the orchestrator context assigns it to BUG-260731-lepevi,
but that bug's scope is the **Linux** lane. Its eight Windows failures are a
different platform's behaviour and are not covered by a Linux-lane AC. Confirm the
boundary with the orchestrator rather than assuming either reading.

Not in the platform-case ledger: none of the 14 unowned cases appear in
`platform-cases.tsv`, so the ledger gate does not require them. They still fail the
job, because the Test lane runs `go test` and any failure fails it.

## Gap 2 — BUG-260731-27h1yc's recorded scope understates its own work

Its description names five failing cases and "five packages beyond
internal/runtimestore". The evidence shows **70 failing cases in its own three
packages** — 60 in `internal/install` alone, where the description names only
`TestEndToEndInstall` — and ten packages beyond `internal/runtimestore`.

That bug is already in `development` with an agent on it. Its notes were updated
with these counts rather than its scope rewritten mid-flight, so the owner can
re-estimate deliberately; the ~14x underestimate is an execution risk worth
surfacing early.
