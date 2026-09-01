# macOS hardened capability probes

A prototype that measures whether a macOS host can actually provide the six
guarantees of the Curator Hardened Execution Profile 1.0, before any of it is
wired into Curator or csk.

It is a **capability observation harness, not an enforcement implementation.**
A run that reported every capability available would still not be a claim that
this host enforces hardened builds — that claim belongs to a qualified
implementation with independent review. The harness always emits
`qualification_status: "unqualified"` for macOS and cannot be made to emit
anything else.

The probes create *probe domains*. A probe domain contains no package byte, runs
no Go compiler, and produces no artifact.

## What it measures

The six guarantees, and the eleven capability classes they are built from:

| Guarantee | Classes it requires |
| --- | --- |
| `total-network-denial` | `network-syscall-denial`, `preexisting-endpoint-revocation`, `domain-membership-enforcement` |
| `read-only-source-and-toolchain` | `read-only-source-view`, `read-only-toolchain-view`, `filesystem-view-restriction` |
| `private-build-root-only-writes` | `write-path-confinement`, `filesystem-view-restriction`, `domain-membership-enforcement` |
| `hard-aggregate-descendant-resource-bounds` | `aggregate-resource-bounds`, `domain-membership-enforcement`, `domain-atomic-termination` |
| `exact-executable-allowlisting` | `exec-path-allowlist`, `domain-membership-enforcement` |
| `fail-closed-capability-preflight` | `active-capability-probe` |

Each class probe has the same shape:

- a **positive** test, which applies the candidate control in a probe domain and
  observes whether the kernel refuses the operation;
- a **negative control**, which repeats the same operation with the control
  removed and must observe success — without it, a profile that simply broke the
  agent would look like perfect enforcement;
- one or more **adversarial escapes**, which try to reach the same effect by a
  different route (a second name for the same bytes, a detached descendant, an
  executable mapping instead of an exec).

A class is available only when the positive test denies, the negative control
succeeds, and no adversarial escape succeeds. Anything else — including a probe
that could not run — is unavailable, because an inconclusive probe must reject.

### The aggregate bounds

`hard-aggregate-descendant-resource-bounds` names more than one quantity, so
`aggregate-resource-bounds` measures each of them separately. Every bound below
is executed — a real limit is installed on a real process, which then tries to
pass it — and every one carries a matched control.

| Bound | Mechanism | Positive test | Controls |
| --- | --- | --- | --- |
| descriptors | `RLIMIT_NOFILE` | the declared cap binds the process that set it | a child gets a second full budget under the same cap |
| bytes on disk | none | a write past the declared byte budget is refused | the write reaches twice the budget |
| CPU time | `RLIMIT_CPU` | `SIGXCPU` arrives at the declared CPU-second | an unbounded run passes the budget; a descendant gets a fresh budget; the member tries to raise its own soft limit back |
| address space | `RLIMIT_AS` | a bound at the declared budget can be installed | the kernel accepts *some* value, so a refusal is about the value and not the call |
| data segment | `RLIMIT_DATA` | as above | as above |
| process count | `RLIMIT_NPROC` | the declared bound refuses a descendant | the domain is asked whether it received the budget declared for it |
| wall-clock time | a supervisor deadline | the deadline ends the domain root while it still has work | a domain that finishes in time exits on its own; the descendant tree is inspected after cancellation |

The wall-clock probe is also the cleanup probe: after the deadline has fired and
the supervisor has issued everything it can, the harness records which
descendants remain, then signals them by pid and records whether *that* left
anything. A production implementation cannot rely on the second step — it only
reaches descendants the supervisor already knew about — so the two are reported
as separate checks rather than one.

## Running it

```sh
go build -o ./hardened-probe ./cmd/hardened-probe

# The measurement run. Exit 1 is the fail-closed outcome and the expected
# result on an unqualified platform.
./hardened-probe --evidence evidence.json --report report.json

# Force each capability class unavailable in turn and check that the run
# rejects before domain entry every time.
./hardened-probe --fail-closed-sweep --quiet --report report-fail-closed.json

# Assert an outcome from a script. Exit 0 when the assertion holds.
./hardened-probe --expect rejected --quiet
```

Or capture a whole evidence packet, including host and tool versions and the
real exit status of every case:

```sh
./capture-evidence.sh ./out
```

### Exit status

| Code | Meaning |
| --- | --- |
| 0 | every capability applied and every guarantee established |
| 1 | rejected: at least one capability could not be established (fail-closed) |
| 2 | the harness itself could not produce a trustworthy record |

Exit 1 and exit 2 are deliberately distinct: an unusable harness is not evidence
about the host.

### Output

- **stdout** is the closed `hardened-capability-evidence-v1` record: eleven
  fields, exactly one capability entry per class, exactly one guarantee entry
  per guarantee, and nothing else. It is what a caller parses.
- **stderr** is the operator summary.
- `--report` writes the detailed prototype-only artifact: every check, its
  expectation in words, what was observed, the mechanism inventory, and the host
  the run happened on. The closed record forbids extra fields, and an operator
  still needs the observations behind the verdicts, so they are separate files.

Every entry in the mechanism inventory carries an `exercised` flag and an
`observation`. A mechanism this run probed says which checks measured it and
what they saw; one that was only considered says so in as many words. Without
that split, a status written by hand next to a mechanism nobody probed reads
exactly like one the run established.

## Layout

| Path | What it is |
| --- | --- |
| `cmd/hardened-probe` | the command, and the in-domain agent under `__inside` |
| `internal/spec` | transcription of the normative constants the probes are measured against |
| `internal/evidence` | the closed evidence record: build, validate, decode |
| `internal/seatbelt` | seatbelt profile rendering and probe-domain launch |
| `internal/inside` | the in-domain agent: attempts operations, reports what it observed |
| `internal/probe` | the class probes, and the reduction from checks to a record |
| `capture-evidence.sh` | one-command evidence packet with host facts and exit codes |

The probe binary is its own in-domain agent. That is deliberate: an exact
executable allowlist is sharpest when it has exactly one entry.

## Testing

```sh
go test -count=1 ./...          # includes end-to-end runs against the real host
go test -count=1 -short ./...   # skips everything that creates probe domains
```

The class probes are driven end to end rather than against a fake. What they
measure is whether this kernel refuses an operation, and a stub that answered
"denied" would only make the harness agree with itself. The end-to-end
assertions are therefore about the shape and internal consistency of the result,
never about which verdict a particular host produces — a host that gained or
lost a capability should change the observation, not turn a test red.

Two things are asserted absolutely, because they are properties of the harness
rather than of the host:

- **the probe binary must build.** If it does not, the end-to-end tests fail
  rather than skip. A suite that skips its subprocess tests because the thing
  under test would not compile reports green while measuring nothing.
- **the run must leave no descendant behind.** The whole prototype is a
  statement about descendant control; a harness that leaked its own would be
  arguing against itself.

The reduction from measurements to verdicts *is* tested against constructed
input, in `internal/probe/bounds_test.go`. That is the one part that can be:
it is arithmetic over numbers the host produced. Those tests feed it the numbers
a host with an unescapable domain or an aggregate bound would report, and
require the verdict to change — which is what makes the conclusions
measurements rather than constants.

## Status of this prototype

This directory is exploratory. It is not built by the repository's `Makefile`,
is a separate Go module, and nothing in Curator imports it.
