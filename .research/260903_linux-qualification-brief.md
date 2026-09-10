# Qualify Linux for compiled builds

You are adding Linux to the Curator Protocol's portable execution policy, so that
`go-repository-v1` skills can be built on Linux at all. Today they cannot: the driver
refuses before a worker starts, by design, and the refusal is written into the protocol
rather than into an implementation.

Read the whole brief before touching anything. The design decision in §3 is the part
that matters; the code is largely a port.

---

## 1. Ground truth

Establish these for yourself rather than trusting the summary. Every claim below is
checkable in one command.

**The authority is a conformance vector, not code.**
`curator-spec/conformance/v1/vectors/go-host-execution-policy.json#native_control_inventory`
declares five controls. Each carries a closed per-platform record with exactly two keys:
`macos` and `windows`. There is no `linux` key. Each record is either
`{"availability": "available", "mechanism": "<name>", "unavailable_reason": null}` or
`{"availability": "unavailable", "mechanism": null, "unavailable_reason": "<reason>"}`.

**The exclusion is explicit and names this work.**
`curator-spec/conformance/v1/vectors/conformance-claim-v3-qualification.json` carries
`{"name": "linux", "status": "excluded", "until_task": "TASK-260728-1skseh"}`. Note that
`macos` and `windows` are `pending-downstream-native-evidence`, not `claimed` — no
platform is claimed yet; Linux is separately excluded.

**The implementation mirrors the vector and says so.**
`curator/internal/godriver/controls.go` holds `nativeControlInventory` (the five control
names, in order) and `nativeControlPlatforms` (the per-platform records). Its comment is
the constraint you must respect:

> "A probe may only confirm or contradict an entry; it may not add or rename one."

`InventoryPlatform(goos string)` maps `darwin`→`macos`, `windows`→`windows`, everything
else→`""`.

**The refusal path.** `curator/internal/godriver/controls_other.go`, build-tagged
`!darwin && !windows`, returns `CodeControlUnavailable` from `prepareControlDomain` with
"the portable execution policy is specified for macOS and Windows only". This file exists
only to keep the package buildable and states that it must never be reached.

**The CI carve-out.** `curator/README.md` documents it at two granularities: the whole
`internal/godriver` package is not executed where the supplied root's qualification vector
marks the platform excluded, while `TestProbeRejectsAnUncoveredPlatformBeforeTheWorker`
still runs on that runner, so the exclusion is *asserted* rather than obeyed. Individual
`cmd/curator` cases are carved out by `requireNativeControlInventoryPlatform`, which reads
`godriver.InventoryPlatform` rather than a `GOOS` list so it cannot drift.

**The five controls, and what backs them today.**

| Control | macOS | Windows |
| --- | --- | --- |
| `descendant-domain-termination` | available — `process-group-and-session-teardown` | available — `job-object-kill-on-close` |
| `active-process-count-limit` | **unavailable** — `no-private-aggregate-domain` | available — `job-object-active-process-limit` |
| `aggregate-memory-limit` | **unavailable** — `no-private-aggregate-domain` | available — `job-object-process-and-job-memory-limit` |
| `per-file-size-limit` | available — `rlimit-fsize` | **unavailable** — `no-private-aggregate-domain` |
| `inherited-handle-restriction` | available — `close-on-exec-and-explicit-descriptor-release` | available — `explicit-handle-inheritance-list` |

Only the absence of a *mandatory* control rejects an operation. The inventory controls are
`applied_when_available`. Read `mandatoryControls` in `controls.go` and confirm which set
is which before you design anything — getting this backwards would make Linux either
falsely permissive or unusable.

---

## 2. Two board records exist, and one of them is stale

`TASK-260728-1skseh run-linux-native-external-repository-qualification` (backlog) is the
one named inside the protocol vector. It is the real gate: a dedicated Linux host, a clean
non-root account, the accepted external-build-repository suite run against **released**
manager binaries, and only then a proposed Linux claim-v3 addendum.

`TASK-260729-2sxx7k validate-accepted-curator-go-snapshot-on-linux` (blocked) is narrower
and is now obsolete in two independent ways: it pins to `.temp/TASK-260720-1ljev5/worktree`,
which was archived and removed on 2026-08-27 (recoverable at
`refs/archive/wt/TASK-260720-1ljev5-worktree`), and it targets `ssh lev`, a host that has
been offline for over a month. It also states of itself that its evidence "must be rerun on
the final integrated candidate". Close it as superseded by `TASK-260728-1skseh` once you
have confirmed both facts, or say why it should live.

---

## 3. The design decision — try to break this before you accept it

The obvious move is cgroup v2: it would give Linux all five controls, including the two
aggregate ones that macOS cannot provide, making Linux the best-covered platform. **Do not
start there.** Work the argument below and try to falsify it. If you can, say so with
evidence and change course; if you cannot, implement it.

**The claim: the honest Linux record is the macOS record, and that is the correct answer.**

The two aggregate controls require a *private aggregate domain* — a bound over exactly the
build's descendants and nothing else. Without cgroup delegation, Linux has no such thing:

- `RLIMIT_NPROC` is aggregate but not private. It counts every process of the UID, so a
  build would be bounded by, and could be starved by, unrelated work of the same user.
- `RLIMIT_AS` / `RLIMIT_DATA` are private but not aggregate. They bound one process, not
  the descendant set.

Neither satisfies both halves, which is exactly the condition macOS records as
`no-private-aggregate-domain`. Recording either as the mechanism would be a false record,
and the model has no room for a half-true one.

cgroup v2 *does* provide the private aggregate domain — and it is the wrong tool for this
layer for two reasons:

1. **Availability is a property of the host, not the platform.** Unprivileged cgroup v2
   needs a delegated subtree, in practice a systemd user slice. Inside a container — a
   GitHub Actions runner included — that delegation is frequently absent. The inventory is
   static per platform and a probe may only confirm or contradict an entry. Declaring
   cgroup mechanisms `available` would make the probe contradict the vector on a large
   share of real hosts, and the manager would then refuse to build there. Declaring them
   `unavailable` while implementing them anyway would make the record a lie in the other
   direction.
2. **The specification already put them somewhere else.** `deferredHardenedGuarantees` in
   `controls.go` lists `hard-aggregate-descendant-resource-bounds`, and the comment states
   that none of those guarantees may appear in the mandatory set, the native-control
   inventory, or an evidence record, and that their absence never rejects a portable build.
   Aggregate descendant bounds were deliberately carved out of the portable profile. Adding
   them back for one platform would undo a decision the spec made on purpose.

**Therefore the Linux record is:**

| Control | Linux | Mechanism / reason |
| --- | --- | --- |
| `descendant-domain-termination` | available | `process-group-and-session-teardown` |
| `active-process-count-limit` | unavailable | `no-private-aggregate-domain` |
| `aggregate-memory-limit` | unavailable | `no-private-aggregate-domain` |
| `per-file-size-limit` | available | `rlimit-fsize` |
| `inherited-handle-restriction` | available | `close-on-exec-and-explicit-descriptor-release` |

Identical to macOS, in both mechanism strings and reason strings. That identity is the
point: no new vocabulary is introduced, and a reader comparing the two platforms learns
something true.

**Where cgroups do belong:** a follow-up under the hardened profile
(`STORY-260728-327soo fail-closed-cross-platform-build-execution`), where conditional,
host-dependent capability is already the subject. Record that as a finding; do not
implement it here.

**Falsification targets.** Before accepting the above, establish and report:
- whether `PR_SET_CHILD_SUBREAPER` plus process-group teardown is strictly stronger than
  macOS's session teardown for the descendant-termination control, and whether that
  difference deserves a distinct mechanism string rather than reusing the macOS one;
- whether any unprivileged, container-safe, distribution-independent mechanism provides a
  private aggregate bound (if you find one, the table above is wrong and you should say so
  loudly);
- whether `RLIMIT_FSIZE` and close-on-exec semantics on Linux differ from Darwin in any way
  that the darwin implementation's invariants depend on.

---

## 4. The work, in order

Each phase has a bar. Do not start a phase before the previous one meets its bar.

### Phase 1 — specification (`curator-spec`)

Add the `linux` record to all five controls in
`conformance/v1/vectors/go-host-execution-policy.json#native_control_inventory`, and change
the Linux entry in `conformance-claim-v3-qualification.json` from `excluded`. What it
changes *to* depends on Phase 4: it becomes claimable only with native evidence, so the
correct interim value is the same `pending-downstream-native-evidence` that macOS and
Windows carry — not `claimed`. Update the normative prose that describes the inventory as
covering exactly two platforms; search for that phrasing, it appears in more than one place.

**Bar:** local spec validation passes; both vector families regenerate **twice,
byte-identically** (this repository's release-check requires determinism, not just success);
the specification CI run is green on the exact candidate SHA across Linux, macOS and
Windows. A new release candidate is cut. Nothing downstream moves yet.

### Phase 2 — Curator (Go)

- Split `controls_other.go`. Its build tag is `!darwin && !windows`; Linux gets
  `controls_linux.go` and the remaining platforms keep the refusal path unchanged.
- Add `PlatformLinux = "linux"` beside `PlatformMacOS` and `PlatformWindows`, a `case
  "linux"` in `InventoryPlatform`, and the Linux map in `nativeControlPlatforms`.
- Implement the three available controls. This is largely a port of
  `controls_darwin.go` (298 lines): the `RLIMIT_FSIZE` window with its manager-side mutex,
  the process-group and session teardown, and close-on-exec with explicit descriptor
  release. Read that file first and port its *invariants*, not just its calls — in
  particular the ordering guarantee that the control domain is prepared before process
  creation and that capability evidence is derived only from `domain.installedControls()`
  after installation.
- Bump `SPEC_PIN` to the Phase 1 candidate.
- Update `.github/ci/platform-cases.tsv` and `.github/ci/skip-classes.tsv` so
  `internal/godriver` is no longer excluded on Linux, and so the cases that were skipped by
  `requireNativeControlInventoryPlatform` now run there. Do not weaken a gate to make a
  lane green; if a case cannot run on Linux, it needs a classified skip reason and a ledger
  row, not silence.
- Update `README.md`: the paragraph documenting the carve-out is now wrong.

**Bar:** `go build ./...`, `go vet ./...`, `golangci-lint run`, `gofmt` all clean;
`go test ./...` green on macOS, Linux and Windows; the platform-case gate green in the
candidate lane with `CI_REQUIRE_FULL_ROOT=1`; and — the load-bearing one —
`TestProbeRejectsAnUncoveredPlatformBeforeTheWorker` still passes, now asserting refusal on
a platform that is genuinely uncovered rather than on Linux.

### Phase 3 — csk (Python)

The same inventory, implemented independently. The protocol's value is that two
implementations agree without sharing code; do not port the Go code, derive from the vector.
Bump its spec pin to the same candidate. Its own test suite, typing and lint gates pass.

**Bar:** csk passes the shared conformance vectors for the Linux inventory on a Linux host;
cross-implementation interop is green.

### Phase 4 — qualification (`TASK-260728-1skseh`)

This is the gate the protocol names, and it cannot be simulated. On a clean Linux host with
a non-root account: record distribution, kernel, architecture, filesystem and shell, the
pinned Git/Go/Python toolchains, the manager binaries and the spec and fixture pins; then
run project and global user install, activation, protected caches, exact-tag and
inaccessible-source cases, independent audit ordering, compiler containment, cache
hit/corruption/offline reuse, status/repair/GC, shim and PATH behaviour, collisions,
crash and rollback, uninstall, and permissions. Failures must leave no unauthorized
mutation.

**Bar:** a reproducible host manifest and setup commands are attached; a reviewer confirms
the results are native rather than simulated; only then is a Linux claim-v3 addendum
proposed with exact pins. If anything fails, Linux stays unclaimed and the gaps are
recorded as typed findings — that is a legitimate outcome of this phase, not a failure of
it.

**You may not have a Linux host.** As of 2026-08-27 the only Linux peer on the tailnet had
been offline for a month. Phases 1–3 do not need one; Phase 4 cannot start without one. Say
so plainly rather than substituting a container, a VM image nobody can reproduce, or a CI
runner — the acceptance criterion asks a reviewer to confirm the evidence is native, and a
GitHub runner is not a host anyone can inspect afterwards.

---

## 5. Constraints

**Naming.** The employer's name and its two-letter short form must not appear anywhere in
the curator repository — not in source, docs, task records or commit messages. A CI gate
enforces it. Historical records that named a local checkout path or an internal remote were
masked to `intranet`; follow that if you need to refer to one. The peer Python
implementation may be named normally.

**Tree selection.** Cut every branch from `origin/main`. The board's managed worktrees pin
their base to whatever the local checkout's `main` points at, which has been stale before
and silently produced change requests against a base 51 commits behind. Verify
`git rev-parse origin/main` against your branch point rather than assuming.

**Identity.** Commit as `Ivan Oparin <oparin@me.com>` in curator, curator-spec and the peer
Python implementation. No `Co-Authored-By` lines and no AI attribution.

**Landing.** Product changes land through a pull request per phase, with the full gate set
green. Board state is committed separately by the orchestrator and never inside a product
pull request.

**Stop the line.** If a phase cannot be completed cleanly, do not add a flag, a stub, a
mock, a priority rule or a test that avoids the real behaviour. Two specific traps here:
declaring a control `available` that a probe will contradict on real hosts, and skipping a
case on Linux without a classified reason and a ledger row. Either would make the suite
report a coverage it does not have. Record the constraint, the evidence, the approaches
tried and the decision needed, then stop.

---

## 6. Where findings go

`curator/LOGBOOK.md`, newest entries first, under a `## YYYY-MM-DD` heading with an
`### HHMM — <what was learned>` title. Record what was measured, not what was done: the
falsification results from §3, any Darwin/Linux semantic difference you had to accommodate,
and the qualification evidence with exact host facts. Cross-reference entries with
`[[HHMM]]`.

Board resources go under `.task-board/.resources/<ID>/`. Note that `*.log`, archives and
binaries there are gitignored by design — markdown evidence is what the repository keeps.

---

## 7. First three commands

```
git -C curator rev-parse origin/main
python3 -c "import json;d=json.load(open('curator-spec/conformance/v1/vectors/go-host-execution-policy.json'));print(json.dumps(d['native_control_inventory'],indent=2))"
sed -n '1,60p' curator/internal/godriver/controls_darwin.go
```

Start by reproducing the refusal on a Linux host and reading its exact diagnostic. If you
cannot reach a Linux host at all, say so before writing any code — Phases 1 to 3 are still
worth doing, but you will be writing an implementation you cannot execute, and that must be
stated in every handoff rather than discovered at Phase 4.
