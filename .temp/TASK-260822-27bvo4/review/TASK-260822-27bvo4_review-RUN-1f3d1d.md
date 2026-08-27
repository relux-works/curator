# TASK-260822-27bvo4 — review verdict: changes requested

Reviewer run RUN-260822-1f3d1d. Read-only: no product or test file in the
reviewed tree was modified. Every mutation below was done in a throwaway
worktree and the reviewed tree re-verified byte-identical afterwards.

Reviewed tree: `.temp/TASK-260822-27bvo4/RUN-260822-507162`,
branch `task/TASK-260822-27bvo4-symlink-launcher-507162`, base `6a9b201`,
uncommitted. Diff is 3 modified + 2 new files, 70 insertions.

## The fix is correct and the evidence in the task notes is accurate

`internal/godriver/identity.go:56` now canonicalizes the running manager with
`physicalPath(absolute)` instead of `filepath.EvalSymlinks(filepath.Clean(absolute))`.
That is the same resolver `selectToolchain`, `verifySelectedRoot` and
`mustPhysical` already use, so the identity path stops being the one place in
the driver with its own rule. The substitution battery in
`readExecutableIdentity` is untouched and now applies to the physical file;
on Windows `artifactHasMultipleLinks` opens the resolved path with
`FILE_FLAG_OPEN_REPARSE_POINT` and still rejects reparse points and
`NumberOfLinks != 1`, so resolution-first does not widen what is trusted.
The worker proves its identity through the same `resolveExecutableIdentity`
(`workerserver.go:88`), so both sides canonicalize identically and `matches`
compares like with like.

I reproduced every claim in the implementer notes independently rather than
reading them back.

### Gates re-run by me, in the reviewed tree, real exit codes

| gate | exit |
|---|---|
| `gofmt -l ./cmd ./internal` | 0 (no output) |
| `go build ./...` | 0 |
| `go vet ./...` | 0 |
| `GOOS=windows go vet ./internal/godriver/` | 0 |
| `GOOS=linux go vet ./internal/godriver/` | 0 |
| `golangci-lint run` | 0 — `0 issues.` |
| `go test ./internal/godriver/ -count=1` | 0 (77.1s) |
| 5 new darwin-visible identity tests `-v` | 0, **0 skips** |
| `make ledger-check` | 0 (63 rows across linux darwin windows) |

### Independent red-to-green, native Windows (`ssh win`, go1.25.5 windows/amd64)

Reviewed tree shipped to `C:\curator-ci\review27` and run there.

* post-fix, 7 identity tests including both junction tests: **exit 0, 0 skips**
* pre-fix baseline (`physicalPath` reverted to `EvalSymlinks` on the host copy,
  restored afterwards, `go build` 0): both junction tests **exit 1** with
  `go-v1 build_execution_worker_identity_invalid: cannot canonicalize the manager executable`,
  and `EvalSymlinks(...\current\curator.exe) = "", The system cannot find the path specified.`
  That is exactly the inversion the task describes — the operator's launch shape
  refused as a fault before any substitution check ran.
* full `go test ./internal/godriver/ -count=1` post-fix: **exit 0, 93.781s**

### Mutation evidence on darwin (throwaway worktree, restored)

* Reverting only `physicalPath` → `EvalSymlinks`: all six identity tests still
  **PASS**. On unix the two are literally the same function, so **the darwin lane
  proves nothing about this change** — the load-bearing proof is Windows-only.
  The implementer said this plainly in the notes; confirmed.
* Stripping canonicalization entirely (`canonical := filepath.Clean(absolute)`):
  all six identity tests **FAIL**. Canonicalization as such is load-bearing on
  darwin too, independently of any package manager.

### Platform evidence recorded by the tests themselves

`os.Executable` reports the **launch path**, not the physical file, on both hosts:

* darwin: `.../001/curator` rather than `.../b001/godriver.test`
* windows: `...\001\curator.exe` rather than `...\b001\godriver.test.exe`

So resolution is observed to be load-bearing on both, not assumed.

## Why this is not accepted yet

Both findings are about wiring the new tests into this repo's own CI gate
contract. Neither touches the product fix, which is right.

### 1. New skip reason is unclassified — CI-fatal on a legitimate host condition

`internal/godriver/identity_test.go:24`

```go
t.Skipf("this account cannot create a file symbolic link: %v", err)
```

`.github/ci/platform-case-gate.sh` TIER 2 classifies **every** skip anywhere in
the run against `.github/ci/skip-classes.tsv`, and its own text says an
unrecognised reason is fatal: *"add it to .github/ci/skip-classes.tsv with a
class, or fix the case."* The change did neither.

Matched each of the three new skip reasons against every regex in the table,
unanchored extended-regex, the same way the gate does:

| reason | result |
|---|---|
| `this account cannot create a file symbolic link: …` | **UNCLASSIFIED → fatal** |
| `this host cannot create a junction os.Lstat reports as a non-directory` | ok — `host-capability :: this host cannot create` |
| `this host cannot create a hard link: …` | ok — `host-capability :: this host cannot create` |

`internal/godriver` runs on the darwin and windows runners, so on any Windows
host without `SeCreateSymbolicLinkPrivilege` this skips four subtests of
`TestExecutableIdentityResolvesALauncherLink` plus five of
`TestExecutableIdentityRejectsSubstitutionThroughALauncherLink` and turns CI red
for the wrong reason, instead of the graceful degradation the branch was written
to provide. It did not fire on the Windows host I used, which is exactly why it
would land unnoticed.

The repo already has the classified phrasing for this identical situation:
`internal/buildcache/protection_windows_test.go:272` uses
`t.Skipf("creating Windows symlink requires host support: %v", err)`, and
`skip-classes.tsv` carries a `host-capability` row for it. Reuse that wording,
or add a row for the new one.

### 2. No `platform-cases.tsv` rows for the new identity cases

The fix is a Windows-only behaviour change — I verified the pre-fix resolver
passes every darwin test — yet no ledger row requires any runner to execute the
cases that prove it. Consequences, both while green:

* a rename or deletion of `TestExecutableIdentityResolvesALauncherReachedThroughADirectoryJunction`
  is invisible to CI;
* its `t.Skip("this host cannot create a junction os.Lstat reports as a non-directory")`
  classifies as `host-capability`, policy `allow`, so a runner that stops
  producing junctions silently drops the only case that catches a regression of
  this fix.

This is in-convention, not a new ask: `platform-cases.tsv:107` already carries
the sibling `internal/godriver TestExecutableIdentityRejectsSubstitutionAndMutation
darwin,windows -`. Suggested rows — `ledger-consistency.sh` accepts them, since
the junction cases live in a `//go:build windows` file and the rest compile on
both served platforms:

```
internal/godriver	TestExecutableIdentityResolvesALauncherLink	darwin,windows	-	-	the manager resolves its own executable through a package-manager launcher link before identity checks
internal/godriver	TestExecutableIdentityRejectsSubstitutionThroughALauncherLink	darwin,windows	-	-	resolving the launch shape does not trust what it reaches; substitution of the installed file still fails closed
internal/godriver	TestExecutableIdentityResolvesALauncherReachedThroughADirectoryJunction	windows	-	-	a manager installed under a directory junction resolves rather than being refused as a non-canonical file
internal/godriver	TestExecutableIdentityStillRejectsSubstitutionBehindAJunction	windows	-	-	a retargeted or dangling junction cannot move a recorded manager identity
```

Widening row 107 to cover the new cases instead is fine; the point is that some
runner must be *required* to execute the Windows case.

## What I am not asking for

The product change, the test set and the evidence are good. The tests cover four
launch shapes resolving to one identity, nine substitution shapes still failing
closed, a real re-exec through a link, the full `Build` handshake with
`managerExecutable` pointed at a link, and the worker process itself started
from a link. `identityProbeMode` lives in `main_test.go`, so it is not in the
shipped binary. The doc comment on `resolveExecutableIdentity` states the
ordering rule and the Windows junction reason accurately.

## Verdict

`to-dev` — two mechanical fixes against the repo's gate contract, then another
reviewer cycle. Both are small; no re-litigation of the fix or the test set.

## Artifacts from this review

Under `.temp/TASK-260822-27bvo4/review/`: `godriver-suite.log`,
`identity-tests-verbose.log`, `darwin-prefix-evalsymlinks.log`,
`darwin-mutant-no-canonical.log`, `windows-identity-postfix.log`,
`windows-prefix-red.log`, `windows-godriver-full.log`, `ledger-check.log`,
`lint-darwin.log`.
