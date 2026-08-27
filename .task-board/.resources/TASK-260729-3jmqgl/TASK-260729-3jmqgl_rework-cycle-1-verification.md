# TASK-260729-3jmqgl — rework cycle 1 verification log

Run after reviewer verdict `RUN-260729-6e3c8f`, on the macOS primary host.
Every command below ran as a standalone process. Nothing was piped through
`tee`; each exit status is the real status of the named process.

## Host

```
date-utc: 2026-07-29T13:00:20Z
uname: Darwin MacBook-Pro-2.local 25.5.0 Darwin Kernel Version 25.5.0: Mon Apr 27 20:39:42 PDT 2026; root:xnu-12377.121.6~2/RELEASE_ARM64_T6031 arm64
sw_vers-productName: macOS
sw_vers-productVersion: 26.5
sw_vers-buildVersion: 25F71
arch: arm64
uid: 502
csrutil: System Integrity Protection status: enabled.
go: go version go1.25.5 darwin/arm64
golangci-lint: golangci-lint has version 2.4.0 built with go1.25.5 from (unknown, modified: ?, mod sum: "h1:qz6O6vr7kVzXJqyvHjHSz5fA3D+PM8v96QU5gxZCNWM=") on (unknown)
```

Working directory: `.temp/TASK-260729-3jmqgl/worktree/prototypes/macos-hardened-probes`

## Verification gates

| Command | Exit |
| --- | --- |
| `go build ./...` | 0 |
| `go vet ./...` | 0 |
| `gofmt -l .` | 0 |
| `golangci-lint run --config ../../.golangci.yml ./...` | 0 |
| `go test -count=1 ./...` | 0 |
| `go test -count=1 -cover ./...` | 0 |
| `./capture-evidence.sh <out>` | 0 |

`gofmt -l .` produced no output. `golangci-lint` reported `0 issues.`

## Test results

```
ok  	github.com/relux-works/curator/prototypes/macos-hardened-probes/cmd/hardened-probe	145.586s
ok  	github.com/relux-works/curator/prototypes/macos-hardened-probes/internal/evidence	0.765s
ok  	github.com/relux-works/curator/prototypes/macos-hardened-probes/internal/inside	24.415s
ok  	github.com/relux-works/curator/prototypes/macos-hardened-probes/internal/probe	117.639s
ok  	github.com/relux-works/curator/prototypes/macos-hardened-probes/internal/seatbelt	2.054s
ok  	github.com/relux-works/curator/prototypes/macos-hardened-probes/internal/spec	1.649s
```

## Coverage

```
ok  	github.com/relux-works/curator/prototypes/macos-hardened-probes/cmd/hardened-probe	146.751s	coverage: 81.2% of statements
ok  	github.com/relux-works/curator/prototypes/macos-hardened-probes/internal/evidence	0.345s	coverage: 99.3% of statements
ok  	github.com/relux-works/curator/prototypes/macos-hardened-probes/internal/inside	24.528s	coverage: 87.4% of statements
ok  	github.com/relux-works/curator/prototypes/macos-hardened-probes/internal/probe	118.178s	coverage: 85.2% of statements
ok  	github.com/relux-works/curator/prototypes/macos-hardened-probes/internal/seatbelt	1.857s	coverage: 97.7% of statements
ok  	github.com/relux-works/curator/prototypes/macos-hardened-probes/internal/spec	1.095s	coverage: 100.0% of statements
```

`internal/inside` is 87.4% against a ~80% target. The new bound code runs
mostly in child processes, which coverage does not credit to the parent; the
safe orchestration paths are driven in-process and the soft-limit escape is
measured against `RLIMIT_CORE`, whose hard limit is harmless to lower. The
runs that install a real limit stay in subprocesses: three of the four bounds
can end the process that hits them.

## Evidence capture

```
list-classes	exit=0	expected=0	cmd=hardened-probe --list-classes
measure	exit=1	expected=0-or-1	cmd=hardened-probe --work-dir ... --evidence ... --report ...
fail-closed-sweep	exit=1	expected=0-or-1	cmd=hardened-probe --fail-closed-sweep --quiet ...
assert-rejected	exit=0	expected=0	cmd=hardened-probe --work-dir /Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3jmqgl/evidence-run-06/work-assert-rejected --force-unavailable network-syscall-denial --expect rejected --quiet
assert-established	exit=2	expected=2	cmd=hardened-probe --work-dir /Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3jmqgl/evidence-run-06/work-assert-established --force-unavailable network-syscall-denial --expect established --quiet
leftover-processes	count=0	expected=0	cmd=pgrep -f /Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3jmqgl/evidence-run-06/hardened-probe
```

`capture-evidence.sh` exited 0, so every case produced the exit status it is
documented to produce, and `pgrep -f <probe binary>` found no surviving
process after the capture.

## Rework items and where they were addressed

| Reviewer requirement | Where |
| --- | --- |
| Executable probes plus matched controls for CPU, memory/address space, process count and wall clock across descendants | `internal/inside/bounds.go` (bound matrix, stress and escape processes), `internal/probe/bounds.go` (reduction), `internal/probe/wallclock.go` (deadline probe) |
| Retain the descriptor and disk-byte probes | `descriptorAndDiskChecks` in `internal/probe/bounds.go`, unchanged in substance |
| Derive the supervisor accounting verdict from measured membership and atomic termination | `supervisorAccountingCheck` in `internal/probe/bounds.go`; `TestSupervisorAccountingIsDerivedFromThisRun` requires the verdict to change when the measurements change |
| Prove deadline cancellation leaves no detached descendant | `probeWallClock`; the three states are reported separately as `deadline-cancellation-leaves-no-descendant` (platform) and `harness-leaves-no-descendant-behind` (harness hygiene) |
| Make an unbuildable probe binary fail the end-to-end tests instead of skipping | `TestProbeBinaryBuilds` and `requireAgent` in `internal/probe/run_test.go`; `buildProbeBinary` in `internal/inside/agent_subprocess_test.go` |
| Platform inventory must not conclude wider than the executable observations | `internal/probe/mechanisms.go`: one entry per `RLIMIT_*` resource, plus `exercised` and `observation` fields filled by `annotateMechanisms` from this run's checks |
| Regenerate source, evidence and outcome archives on the macOS host | `evidence-run-06`; all three board artifacts replaced |

## Scope

No production Curator code, shared caches, specs or unrelated files were
modified. Nothing was committed, staged, published, installed or downloaded, and
no host configuration was changed. Only task-local prototype tests and evidence
capture were run; no repository-wide Curator suite was executed.
