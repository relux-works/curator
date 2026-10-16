# SNAP-LOCK validation evidence

Candidate: `1e6be490fee5768d0aa6808b825a20437058ef1c`.
Tree: `b91a63aa36e9e266a7514d9a4d92356181bbfc6f`.
Go 1.25.5, darwin/arm64; golangci-lint 2.12.2.

Each command is run as a standalone process. Empty output is not used to infer success: its recorded process exit is stated explicitly.

## Focused race suite

```sh
go test -race ./internal/envprofile ./internal/managerlock ./internal/transaction -run 'TestMaterializeCapture|TestPendingRecovery|TestRecoveryObserver|TestCaptureLock|TestAcquireExistingHomeOnly' -count=1
```

Exit: **0**. Log: `race-07.log`.

```text
ok  	github.com/relux-works/curator/internal/envprofile	24.405s
ok  	github.com/relux-works/curator/internal/managerlock	1.414s
ok  	github.com/relux-works/curator/internal/transaction	1.863s
```

## Lint

```sh
GOLANGCI_LINT_CACHE="$PWD/.temp/TASK-261016-4zofp0/lint-cache" make lint
```

Exit: **0**. Log: `lint-09.log`.

```text
golangci-lint run
0 issues.
```

## Build

```sh
go build ./...
```

Exit: **0**. Log: `build-07.log`.

```text
(empty output)
```

## Vet

```sh
go vet ./...
```

Exit: **0**. Log: `vet-05.log`.

```text
(empty output)
```

## Windows compilation only

```sh
GOOS=windows GOARCH=amd64 go test -c ./internal/envprofile -o .temp/TASK-261016-4zofp0/envprofile-windows.test.exe
```

Exit: **0**. Log: `windows-compile-02.log`.

```text
(empty output)
```

## Whitespace

```sh
git diff --check
```

Exit: **0**. Log: `diff-check-05.log`.

```text
(empty output)
```

## Shell syntax, unchanged script

```sh
sh -n scripts/remote-gate.sh
```

Exit: **0**. Log: `gate-syntax-03.log`.

```text
(empty output)
```

## Candidate identity

```text
HEAD 0d00a109ad1da77a5506747c89f24af9442eb341
Candidate tree b91a63aa36e9e266a7514d9a4d92356181bbfc6f
Working tree matches signed gate candidate; Story HEAD unchanged.
```

The detached snapshot signature verified as `G`; Story HEAD remains at its checkpoint.

## Previous full gates: regression evidence

[First frozen run](https://github.com/relux-works/curator/actions/runs/38835465876), candidate `90685b712e500a752f23430a9b5f9385740cd8b5`, returned **exit 1**. Nineteen enabled jobs passed; Windows Test failed: held bookkeeping-lock reads, held-object replacement plants, and unknown skip descriptions.

[Second frozen run](https://github.com/relux-works/curator/actions/runs/38840074132), candidate `37f08891e6d0325cc6cd2d3c9a9d604de1860719`, returned **exit 1**. Nineteen enabled jobs passed; Windows Test failed only four replacement plants with AccessDenied. Its platform-case checker exited 0. This disproves the readonly-only explanation: clearing that attribute did not make MoveFileEx a suitable held-object plant.

The current fixture uses FileRenameInformationEx with POSIX replacement semantics through the existing Windows dependency. No product gate is weakened. The replacement full-gate process has now exited 0; the current result is recorded below.

The attempted in-progress Windows job-log download returned exit 1 (HTTP 404). It was not used to infer absence, failure cause, or a CI fallback. Completed-job evidence was downloaded after each run ended.

## Current full gate

```sh
sh scripts/remote-gate.sh
```

Exit: **0**, observed from the actual gate process after it returned. Log: `remote-gate-09.log`. [Hosted run](https://github.com/relux-works/curator/actions/runs/38845304296). The completed run head equals the signed candidate. Twenty enabled jobs succeeded; two workflow-optional jobs skipped for their declared conditions. The sanitized matrix is `TASK-261016-4zofp0_gate.json`. Windows evidence download also exited 0; its platform-case gate reports `ok`. Linux and macOS evidence downloads each exited 0.

```sh
git diff --exit-code HEAD -- ':(glob)**/*_test.go'
```

Exit: **0**. Log: `retained-tests-01.log`; all previously tracked test files are unchanged. Fresh temporary-index identity: `b91a63aa36e9e266a7514d9a4d92356181bbfc6f`, equal to the candidate tree. Real index empty, Story HEAD unchanged, signature verified `G`.

## Hosted leaf runtime evidence

| Platform | Pass records | Skip records | Bound |
|---|---:|---:|---|
| Linux | 54 | 0 | All supplied leaf cases ran |
| macOS | 54 | 0 | All supplied leaf cases ran |
| Windows | 46 | 6 | Two held-lock identity plants are prevented by OS sharing; two executable-mode cases, the shell vendor sentinel and POSIX unreadability are unavailable |

Counts are named terminal Go test/subtest records; parents and children are counted separately. Windows successfully executed the same-fragment ABA, store-object identity and marker-identity witnesses using the native replacement fixture. The six skipped cases are declared bounds, not passing qualification; the corresponding Unix cases ran. Exact names and verdicts are in `TASK-261016-4zofp0_platform-cases.json`. Narrowing mutants were executed on darwin/arm64, not replayed on Windows.
