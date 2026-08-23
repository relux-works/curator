# TASK-260823-qwr5w9 — composite assembly and local gate evidence

Branch: task/TASK-260823-qwr5w9-credential-scopes-composite
Base:   origin/main 6a9b201da828181f18cf285e75a78e39a0337585
PR:     https://github.com/relux-works/curator/pull/22
Spec pin used for every gate: 00b1688a9b2457ca397a0bb550acf47cad8ee967

## Commit series
```
fdbd4c0 G Read a manifest symlink as present even when its target is gone
7193651 G Pin that vendored third-party text is inert to the audit gate
46ed906 G Name the skill whose manifest broke, not the one that declared it
b28376f G Tell the operator what to do about a toolchain executable mismatch
831023d G Let the manager resolve its own launcher link before judging it
14db4d9 G Give the operator one place to choose the SSH key per build repository
```
(`%G?` = G: good signature on every commit.)

## Assembly integrity

Prior run left 42 files staged and uncommitted; nothing had been committed.
Each accepted patch artifact was proven fully present by a reverse-apply check
against the assembled tree before any commit was made:

```
git apply -R --check TASK-260822-4p3dcq_final.patch            -> exit 0
git apply -R --check TASK-260822-27bvo4_symlink-launcher.patch -> exit 0
git apply -R --check TASK-260822-2v5e80_remedy.patch           -> exit 0
```

The three patches touch disjoint file sets, so the admission.go and
skip-classes.tsv overlaps named in the task description were already
reconciled inside the 4p3dcq_final chain; no hand merge was needed.
The only file touched by two groups was internal/install/install_test.go
(closure provenance + vendor audit), split at the function boundary.

Post-split diff vs origin/main is byte-identical to the pre-split blob:
```
 42 files changed, 6014 insertions(+), 30 deletions(-)
```

## Local gates — real exit codes, each run standalone, no pipes

| Gate | Exit | Result |
| --- | ---: | --- |
| `gofmt -l cmd internal` | 0 | empty |
| `go build ./...` | 0 | |
| `go vet ./...` | 0 | |
| `golangci-lint run` | 0 | 0 issues |
| `no-broad-suppression.sh` | 0 | ok |
| `gate-selftest.sh` | 0 | 75 passed, 0 failed |
| `ledger-consistency.sh` | 0 | 72 rows across linux darwin windows |
| `toolchain-identity.sh` | 0 | under CI env |
| `test-gate.sh` | 0 | 34 served, 7 deferred, 49 classified skips |
| `go test ./internal/interop/` | 0 | |
| naming gate (inline) | 0 | 1 README line, 0 elsewhere |

### toolchain-identity environment note

The gate is exit 1 on a bare local shell here and exit 0 under the CI
environment. Both failures it reported are host facts the gate exists to
catch, not properties of this branch:

```
GOTOOLCHAIN=auto, not local                      -> CI sets GOTOOLCHAIN: "local" (ci.yml:44)
GOENV=.../go/env names a per-user go env file    -> resolved with GOENV=off
```

PATH must also put the real GOROOT/bin ahead of the goenv shim — which is
precisely the operator remedy commit b28376f adds to the diagnostic.

## test-gate suite plan
```
suite-plan: GOOS=darwin
suite-plan: root=/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260823-qwr5w9/protocol-spec/conformance/v1
suite-plan: the root publishes no vectors/conformance-claim-v3-qualification.json; the recorded default_excluded_on applies

defer internal/buildcache
      the supplied root publishes none of: vectors/build-drivers.json
      it runs with CURATOR_CONFORMANCE_ROOT unset, taking the path its own tests implement
defer internal/buildsource
      the supplied root publishes none of: vectors/build-drivers.json
      it runs with CURATOR_CONFORMANCE_ROOT unset, taking the path its own tests implement
...

suite-plan: served=34 deferred=7 excluded=0
suite-plan: ok
```

## test-gate verdict
```
platform-case gate: ok

test-gate: go test exit=0, platform-case gate exit=0
```
