# Final Curator compiled-build CI execution map — revision 2

**Task:** TASK-260729-osjeay (read-only audit, rework cycle 1)
**Target task:** TASK-260720-1pvfj5 `enforce-cross-platform-ci-gates`
**Date:** 2026-07-29
**Supersedes:** revision 1, which carried three factual errors corrected in §1.1.

**Classification.** Read-only. No product, spec, CI, `Makefile`, pin, or `TASK-260720-1pvfj5` field
was modified. No Go command, build, test, install, download, or network fetch was executed.
**Every command in §5, §6, §7, §8 and §9 is a future producer gate, not evidence.** Nothing in this
document is a green CI result. The two `shasum` measurements in §4.2 and the `git`/filesystem reads
in §10 are the only commands this audit actually ran.

---

## 1. Executive summary

Implementation cannot start from the target task as currently worded. **Five** of its clauses are
unsatisfiable against the accepted rc.5 candidate, and four need a board-owner decision before a
producer writes a line of YAML. §3 is the decision packet with exact proposed wording.

The four blocking findings, in descending order of blast radius:

1. **The composite is red on every platform under the committed pin — before Linux even enters the
   picture.** Six pre-existing conformance tests hard-`t.Fatal` when `CURATOR_CONFORMANCE_ROOT` is
   set but the artifact is absent, and the committed pin publishes none of the rc.5 artifacts they
   read. The delta's *own new* tests already solved this with `t.Skipf`; the older six were never
   updated. §3.3, §4.4.
2. **`go test ./...` on the composite exits 1 at Go's default 10-minute timeout**, measured twice by
   `TASK-260729-2kaopg` in `cmd/curator`. Every gate needs an explicit `-timeout 30m`. §3.6.
3. **Linux cannot run `internal/godriver` at all.** `rc5-native-control-inventory-v1` covers macOS
   and Windows only; the driver fails closed with `build_execution_control_unavailable` before the
   worker starts. §4.3.
4. **The rc.5 candidate has no revision, so no hosted runner can consume it.** It exists only as an
   uncommitted `curator-spec` working tree. §5 selects one executable delivery mechanism and deletes
   revision 1's non-executable hosted candidate rows.

The producer's edit surface stays two files (`.github/workflows/ci.yml`, `Makefile`) **only if**
decision D3 resolves to the fallback. The recommended resolution of D3 adds six test files. §4.5.

### 1.1 Corrections to revision 1

| # | Revision 1 said | Verified truth | Why it matters |
|---|---|---|---|
| C1 | Repo HEAD `c06aa1a` is main; committed pin is `e72defe…`; worktree base `17804ce` is "stale" and composing on it would "silently revert the pin" (invariant I6) | **`main` = `origin/main` = `17804ce`**, pin **`00b1688a…`**. `c06aa1a` is the tip of the *divergent* branch `agent/link-curator-skill-registry`; neither is an ancestor of the other (merge base `ecb6c1a`). | **I6 was inverted.** Composing on `c06aa1a` is what would revert the pin — from `00b1688a` (14 vectors, incl. `manager-lifecycle.json`) down to `e72defe` (10 vectors), breaking `internal/closure`. The worktree base is correct. |
| C2 | rc.5 root is "3 modified + 18 untracked files" | `git status --short --untracked-files=all -- conformance/` → **3 modified, 354 untracked** (357 lines). Collapsed default → 3 modified, 91 untracked (94 lines). | The provenance claim was understated by 20×. Digest `b6f56aac…04c` was and remains correct. |
| C3 | Linux `go test ./...` is the load-bearing blocker | It is the *third* blocker. The pin-artifact `t.Fatal` set (§3.3) and the 10-minute timeout (§3.6) are both larger and hit all three platforms. | Revision 1's YAML shape would still have been red on macOS and Windows. |

---

## 2. Verified state of the world

| Fact | Value | Source (§10 row) |
|---|---|---|
| `main` / `origin/main` | `17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8` "Pin landed rc.3 protocol", 2026-07-14 05:20 | 1 |
| Current checkout | `c06aa1a` on `agent/link-curator-skill-registry`, 2026-07-13 17:17 — **divergent from main** | 1, 2 |
| **Committed suite pin (main)** | **`00b1688a9b2457ca397a0bb550acf47cad8ee967`** (`ci.yml:28`, `ci.yml:81` at `17804ce`) | 3 |
| Pin's tag position | `v1.0.0-rc.2-1-g00b1688` — **one commit past rc.2, not a release tag** | 4 |
| Branch-only pin | `e72defe…` at `c06aa1a` — untagged, ancestor of `00b1688a`, 10 vectors. Not main's pin. | 4, 5 |
| curator-spec tags | `v1.0.0-rc.1`, `-rc.2`, `-rc.3` only. **No rc.4, no rc.5 tag.** | 6 |
| Nearest real release | `v1.0.0-rc.3` = `57c1f56…`, same 14 vectors as `00b1688a` | 4, 7 |
| rc.5 candidate root | `.temp/TASK-260729-3nx97g/worktree/conformance/v1` — dirty tree at `57c1f56` (`v1.0.0-rc.3`) | 8 |
| rc.5 dirt | 3 modified, 354 untracked paths | 8 |
| rc.5 manifest digest | `b6f56aacc0e37dcc6692f73f641bff761e89b645adfe20a47a06d81c6fda204c` — **re-verified, exit 0** | 9 |
| rc.5 tree identity | 448 files, `e6a132157806bf747f0ad24a61bb5a9a4c8b915dac743d9465c889c0a1ad2fae` (snapshot taken this run) | 9 |
| Go version | `go 1.25.5` (`go.mod:3`); CI resolves via `actions/setup-go@v5` + `go-version-file: go.mod` | 10 |
| Package count | 40 buildable package dirs under `cmd` + `internal` in the candidate | 11 |
| Candidate delta | 23 product files, **all `_test.go`**, zero production drift | 12 |
| golangci-lint | action `@v7`, `version: latest` (**mutable**); `.golangci.yml` `version: "2"`, byte-identical main↔candidate | 13 |
| Race gate today | **none** in `ci.yml`, **none** in `Makefile` | 14 |
| Free disk | 25 GiB | 15 |

### 2.1 Current CI job inventory (`.github/workflows/ci.yml`)

| Job | Runner | Steps | `CURATOR_CONFORMANCE_ROOT`? |
|---|---|---|---|
| `test` | matrix `ubuntu-latest`, `macos-latest`, `windows-latest` | gofmt (skipped on Windows), `go vet ./...`, `go test ./...` | yes, on `go test` only |
| `lint` | `ubuntu-latest` | `golangci/golangci-lint-action@v7`, `version: latest` | no |
| `interop` | `ubuntu-latest` | `go test ./internal/interop/ -v` | yes |
| `naming-gate` | `ubuntu-latest` | inline `grep` gate | n/a |

No job passes `-timeout`. No job runs `-race`. `.github/workflows/release.yml` is out of 1pvfj5 scope.

### 2.2 Current `Makefile` targets

`build`, `test`, `fmt`, `vet`, `lint`, `check` (`= vet test` + `gofmt -l .`). No `race`, no timeout,
no conformance-root plumbing, no platform targets.

**Drift:** CI checks `gofmt -l cmd internal`; `make check` checks `gofmt -l .`, which additionally
walks the `agents/skills/skill-go-testing-tools` submodule. Not equivalent gates.

---

## 3. Target-task contract drift — board-owner decision packet

Six clauses of `TASK-260720-1pvfj5` conflict with the accepted rc.5 candidate. D1, D3 and D5 need a
board-owner decision; D2 needs a clarification only; D4 and D6 are mechanical. Exact proposed
wording is given for each.

### D1 — Scope demands the full candidate suite on Linux; Linux cannot run `internal/godriver`

**Current scope text:**
> "Run the full supplied candidate suite on Linux, macOS, and Windows using the repository Go version…"

**Current AC text:**
> "…runs every compiled-build case on ubuntu, macos, and windows with no case silently skipped
> except protocol-defined unsupported platform controls."

**Conflict.** `internal/godriver` fails closed on Linux *before the worker starts* (§4.3). This is
not a "protocol-defined unsupported platform control" — those are individual `t.Skip`ped cases
(`boundary_test.go:268` Windows-only, `build_test.go:433` macOS-only). It is a whole-package
platform exclusion declared normatively in
`conformance/v1/vectors/conformance-claim-v3-qualification.json` as
`{"name":"linux","status":"excluded","until_task":"TASK-260728-1skseh"}`.

**Proposed scope wording (replace that sentence):**
> "Run the full supplied candidate suite on macOS and Windows using the repository Go version. On
> Linux run every package outside `rc5-native-control-inventory-v1` plus the inventory's own
> fail-closed rejection case; `internal/godriver` execution on Linux is deferred to
> TASK-260728-1skseh, which the rc.5 qualification vector names as the `until_task` for the linux
> exclusion."

**Proposed AC wording (replace that clause):**
> "…runs every compiled-build case on macos and windows with no case silently skipped except
> protocol-defined unsupported platform controls, and on ubuntu runs every package the rc.5
> qualification vector does not exclude, with the exclusion asserted by the inventory rejection
> test rather than by omission."

### D2 — AC demands `go test -race ./...`; scope demands scoped race on Linux

**Current AC text:**
> "`go test -race ./...` passes on the selected supported runner"

**Current scope text:**
> "add a supported race job on at least Linux covering transaction, cache, install, and conformance
> packages"

**Conflict.** These two clauses cannot both be met on one runner. `-race ./...` on Linux includes
`internal/godriver` → red (D1). The scope's Linux package list is by construction *not* `./...`.

**Resolution — no AC change needed if two race jobs are run.** `macos-latest` is an inventory
platform and the Go race detector supports `darwin/arm64`, so `go test -race ./...` is executable
there verbatim and satisfies the AC's "selected supported runner". A second, scoped
`race (ubuntu-latest)` satisfies the scope's "at least Linux" clause. Both are in §6.3.

**Proposed scope clarification (append):**
> "The AC's `go test -race ./...` gate is satisfied on macos-latest, the selected supported runner.
> The Linux race job is additionally required and is scoped to the packages named above."

D2 is the only one of the six a producer can implement without a wording change. The clarification
exists so a later reviewer does not read the two clauses as contradictory.

### D3 — The committed pin cannot serve six hard-`t.Fatal` conformance reads (**largest**)

**Current scope text:**
> "Keep the committed curator-spec checkout at the currently qualified released revision during
> candidate development."

**Conflict.** Six test sites read an rc.5-only artifact and call `t.Fatal` when it is missing. They
skip only when `CURATOR_CONFORMANCE_ROOT` is *unset*, not when the artifact is absent. The committed
pin `00b1688a` publishes none of those artifacts (§4.4). CI exports the root for `go test ./...`.
The composite's default `test` job is therefore **statically predicted red on ubuntu, macOS and
Windows alike**, in six packages.

The candidate's own new tests already solved this. Five of the 20 new files guard with
`t.Skipf("%s publishes no build-drivers vector", root)` before touching the artifact
(`internal/skillcheck/builddriver_context_conformance_test.go:24`,
`internal/whitelist/builddriver_context_conformance_test.go:25`,
`internal/skillspec/builddriver_conformance_test.go:39` and `:301`,
`internal/buildcache/builddriver_positive_conformance_test.go:31`). The six older sites were never
updated to that pattern. This is an internal inconsistency in the accepted composite, not a
1pvfj5 defect.

**The six sites, exactly:**

| # | File:line | Missing input | Provenance |
|---|---|---|---|
| 1 | `internal/buildsource/conformance_test.go:16-19` | `vectors/build-drivers.json` | accepted composite |
| 2 | `internal/buildcache/conformance_test.go:15` → `readJSONObject` `:63-66` | `vectors/build-drivers.json` | modified by the jrrgw9 delta |
| 3 | `internal/scopes/gc_conformance_test.go:38-41` | `vectors/external-repository-lifecycle.json` | **new in the jrrgw9 delta** |
| 4 | `internal/marker/marker_v2_test.go:37-41` | `schema-cases/install-marker-v2/` | accepted composite |
| 5 | `internal/whitelist/conformance_test.go:20-24` and `:25-28` | `fixtures/go-build-skill`, `expected/build-driver/context_files.json` | accepted composite |
| 6 | `internal/skillspec/conformance_test.go:106-109` | `fixtures/go-build-skill` | accepted composite |

**Recommended resolution (P1).** Convert those six reads to the delta's own `t.Skipf` pattern. It
removes the maintained list entirely, makes any pin/root combination safe permanently, and is six
mechanically identical edits. It requires a board-owner decision because it widens 1pvfj5's file
surface beyond `ci.yml` + `Makefile`.

**Proposed scope wording (append):**
> "1pvfj5 additionally owns the six conformance test sites that hard-fail on an artifact the
> committed pin does not publish (`internal/buildsource/conformance_test.go`,
> `internal/buildcache/conformance_test.go`, `internal/scopes/gc_conformance_test.go`,
> `internal/marker/marker_v2_test.go`, `internal/whitelist/conformance_test.go`,
> `internal/skillspec/conformance_test.go`). Each is changed to `t.Skipf` on a missing artifact,
> matching the pattern the accepted candidate already uses in its own build-driver conformance
> tests. No assertion is weakened: with the candidate root supplied, every one of them still runs."

Site 3 is a file `TASK-260720-jrrgw9` has not yet handed to acceptance, so that one edit can instead
be routed to jrrgw9 before it lands. Flagged, not actioned — this audit does not write to jrrgw9.

**Fallback (P3), no test-file edits, implementable inside `ci.yml` + `Makefile` only.** Split the
pin lane: export `CURATOR_CONFORMANCE_ROOT` only for a maintained "pin-servable" package list and
run the remaining packages with the variable unset. Cost: a second maintained list that drifts every
time a package gains a conformance read, with no guard that can detect the drift before it goes red.
**Not recommended**, but it is the only path that needs no board decision.

### D4 — Stale rc.4 wording (mechanical)

| Where | Stale text | Correct text |
|---|---|---|
| `checklist[0]` | "CI pins the reviewed immutable **rc.4** protocol commit" | "CI keeps the committed suite pin on the qualified revision; the rc.5 candidate enters only through an explicit non-default input" |
| `ac` | "No README release wording or committed suite pin claims **rc.4** before TASK-260720-25d05o" | `rc.5` / `1.0.0-rc.5` |
| `scope` | "Provide an explicit caller-supplied full candidate revision **or** `CURATOR_CONFORMANCE_ROOT` path" | The "full candidate revision" branch is unavailable — no rc.5 revision exists (§5.1). Only the path branch is executable today. |
| `notes` ¶1 | "candidate **rc.4** tests may use an explicitly supplied `CURATOR_CONFORMANCE_ROOT`" | `rc.5` |

`description` and the two attached `…candidate-release-ci-gates.puml/.svg` resources carry no rc.4
claim and need no change.

### D5 — "one immutable currently released curator-spec pin" is factually false

**Current AC text:**
> "Normal Curator CI keeps one immutable currently released curator-spec pin…"

**Conflict.** The committed pin `00b1688a` is `v1.0.0-rc.2-1-g00b1688` — one commit *past* the rc.2
tag, described by no release tag. It is immutable; it is not a release. The nearest actual release
with identical vector coverage is `v1.0.0-rc.3` = `57c1f56…`.

**Two options, board owner's call:**

- **D5-a (no code change):** soften the AC to "one immutable committed pin at the currently
  qualified revision", and record for `TASK-260720-38l1sy` that the pin it will audit is untagged.
- **D5-b (pin promotion):** move the pin to `57c1f56846d221ecc55786bd3c2467ec32f11730`
  (`v1.0.0-rc.3`), an actual published release with a byte-identical vector set. This would make the
  AC true for the first time. **But pin promotion is explicitly owned by `TASK-260720-38l1sy` after
  `TASK-260720-25d05o`**, so 1pvfj5 must not do it unilaterally.

**Recommendation: D5-a.** It is a wording fix inside the task, changes no gate, and leaves promotion
where the story put it. D5-b is a real improvement but belongs to `38l1sy`.

### D6 — Every `go test` gate needs `-timeout 30m` (mechanical, no decision)

`TASK-260729-2kaopg` measured `go test -count=1 ./...` on this composite exiting **1 at Go's fixed
10-minute default timeout in `cmd/curator`**, twice, with a different late test active each time;
the two alleged victims passed in isolation at 40.0 s and 9.3 s, and `internal/install/atomicity`
alone takes 371.7 s. Its recorded recommendation is an explicit 30-minute package timeout, and its
recorded rule is: never report the default-timeout command as green — its real exit is 1 until the
suite is split or the gate timeout is raised. Every gate in §6, §7 and §9 carries `-timeout 30m` for
that reason. Without it the composite CI is red regardless of D1–D5.

---

## 4. Source evidence for the structural constraints

### 4.1 Dependency state

```
TASK-260720-1pvfj5   status=backlog   isBlocked=true
  blockedBy: TASK-260720-2qqq0w   status=done          ✔ satisfied
             TASK-260720-jrrgw9   status=development   ✘ NOT accepted
  blocks:    TASK-260720-38l1sy, -3pvihp, -z2z795, -z9j4c9
```

`TASK-260720-jrrgw9`'s own blockers (`-2284br`, `-1ljev5`, `-1nlmvv`) are all `done`; jrrgw9 itself
is not. `TASK-260728-1skseh` (`run-linux-native-external-repository-qualification`) is `backlog`.
The NEXT-STEP directive requires both 1pvfj5 blockers independently accepted. **1pvfj5 is not
startable today.** This audit is the pre-work; it does not unblock it.

### 4.2 rc.5 candidate identity — re-verified this run

```
$ shasum -a 256 manifest.json
b6f56aacc0e37dcc6692f73f641bff761e89b645adfe20a47a06d81c6fda204c  manifest.json     # exit 0
$ find . -type f -print0 | LC_ALL=C sort -z | xargs -0 shasum -a 256 | shasum -a 256
e6a132157806bf747f0ad24a61bb5a9a4c8b915dac743d9465c889c0a1ad2fae  -                 # exit 0, 448 files
```

The manifest digest matches the value named in this task's scope brief. The tree digest is a
**snapshot of a live, mutable worktree taken at audit time** — `TASK-260729-1y7okj` recorded a
candidate worktree file moving three times in one afternoon under another producer. The producer
must freeze before measuring (§5.2 C1).

### 4.3 Linux is outside `rc5-native-control-inventory-v1`

```go
// internal/godriver/controls.go:75
var nativeControlPlatforms = map[string]map[string]inventoryRecord{
    PlatformMacOS:   { … },   // "macos"
    PlatformWindows: { … },   // "windows"
}
// internal/godriver/controls.go:200
func InventoryPlatform(goos string) string {
    switch goos {
    case "darwin":  return PlatformMacOS
    case "windows": return PlatformWindows
    default:        return ""            // ← linux
    }
}
// internal/godriver/controls.go:241-245
if platform == "" || nativeControlPlatforms[platform] == nil {
    return "", nil, diagnostic(CodeControlUnavailable, …)
}
```

`internal/godriver/controls_other.go` (`//go:build !darwin && !windows`) is a fail-closed stub whose
own comment says its entry points "must never be reached". `Build` probes at
`internal/godriver/build.go:161`. The protocol side agrees:
`conformance/v1/vectors/conformance-claim-v3-qualification.json` declares
`linux: excluded, until_task TASK-260728-1skseh`.

The godriver tests that drive `Build` / `prepareControlDomain` / `probeNativeControls`
(`build_test.go`, `worker_test.go`, `boundary_test.go`, `executor_test.go`,
`fingerprint_equivalence_test.go`, `graph_test.go`) carry **no build constraint and no Linux skip**;
`newSnapshotFixture` (`main_test.go:134`) has no platform guard. Only five godriver test files carry
constraints, all `_unix`/`_windows` helper files.

**Blast radius is exactly one package.** The only importers of `internal/godriver` are `cmd/curator`
(`builds.go`, `main.go`; tests use `TestedFamilies`, `SelectionCuratorGo`, `SelectionGOROOT`,
`WorkerMode`, `RunWorker`) and `internal/install` (`builddeps.go`, `global.go`, `install.go`,
`plan.go`; `stage_test.go:1288` uses `&godriver.Diagnostic{}` as a fixture value). Neither test set
calls `godriver.Build`. So the scope's named Linux race packages — transaction, cache, install,
conformance — are all Linux-safe, and `internal/interop` imports no godriver either.

### 4.4 Pin capability matrix — which conformance inputs each revision publishes

`✔` present, `✘` absent, each verified with `git cat-file -e <rev>:<path>`.

| Conformance input | `00b1688a` (**committed pin**) | `57c1f56` (`v1.0.0-rc.3`) | `e72defe` (branch-only) | rc.5 candidate | Consumer behaviour when absent |
|---|:--:|:--:|:--:|:--:|---|
| `vectors/closures.json` | ✔ | ✔ | ✔ | ✔ | — |
| `vectors/portable-paths.json` | ✔ | ✔ | ✔ | ✔ | — |
| `vectors/manager-lifecycle.json` | ✔ | ✔ | **✘** | ✔ | `internal/closure` **FATAL** |
| `schema-cases/index.json` | ✔ | ✔ | ✔ | ✔ | — |
| `vectors/build-drivers.json` | ✘ | ✘ | ✘ | ✔ | `buildsource`, `buildcache` **FATAL**; 5 new tests skip |
| `vectors/external-repository-lifecycle.json` | ✘ | ✘ | ✘ | ✔ | `internal/scopes` **FATAL** |
| `schema-cases/install-marker-v2/` | ✘ | ✘ | ✘ | ✔ | `internal/marker` **FATAL** |
| `fixtures/go-build-skill` | ✘ | ✘ | ✘ | ✔ | `whitelist`, `skillspec` **FATAL** |
| `expected/build-driver/` | ✘ | ✘ | ✘ | ✔ | `whitelist` **FATAL**; `buildmeta`, `godriver` skip |
| `schema-cases/agent-skill-v6/valid.json` | ✘ | ✘ | ✘ | ✔ | `skillspec` builddriver test skips earlier |
| `vectors/go-host-execution-policy.json` | ✘ | ✘ | ✘ | ✔ | `godriver` skips |
| `schema-cases/build-receipt-v1/valid.json` | ✘ | ✘ | ✘ | ✔ | `buildmeta` skips |

**`build-drivers.json` exists at no published curator-spec revision** — it is untracked even in the
rc.5 worktree. The build-driver conformance suite can therefore run only under the candidate root.
This restates, with the exact per-revision matrix, what `TASK-260720-2qqq0w` recorded and explicitly
routed to 1pvfj5: "pin/gate reconciliation belongs to TASK-260720-1pvfj5".

Main's own CI is green today because main's tree predates all of this: `internal/godriver` does not
exist at `17804ce`, and main's `internal/closure/conformance_test.go` and
`internal/skillspec/conformance_test.go` read only `closures.json` and `portable-paths.json`, both
of which the pin publishes. The six FATAL sites arrive with the accepted `2kaopg` composite and the
`jrrgw9` delta.

### 4.5 Candidate delta and edit-ownership conflict check

`diff -rq` accepted (`TASK-260729-2kaopg/worktree`) vs candidate (`TASK-260720-jrrgw9/worktree`),
excluding `.git`/`.task-board`: **24 entries — 21 `Only in`, 3 `Files … differ`.** One `Only in`
is the non-product `.temp` directory. Product delta = **20 new + 3 modified = 23 files, all
`_test.go`, zero production drift.**

| Package | New files |
|---|---|
| `cmd/curator` | `lifecycle_conformance_test.go` |
| `internal/buildcache` | `builddriver_positive_conformance_test.go`, `builddriver_rejection_conformance_test.go`, `builddriver_conformance_unix_test.go`, `builddriver_conformance_other_test.go` |
| `internal/buildmeta` | `builddriver_policy_conformance_test.go` |
| `internal/buildsource` | `builddriver_conformance_test.go`, `builddriver_conformance_unix_test.go`, `builddriver_conformance_other_test.go` |
| `internal/godriver` | `builddriver_positive_conformance_test.go`, `builddriver_rejection_conformance_test.go` |
| `internal/install` | `cache_conformance_test.go`, `dryrun_conformance_test.go` |
| `internal/runtimestore` | `launcher_conformance_test.go` |
| `internal/scopes` | `gc_conformance_test.go` |
| `internal/skillcheck` | `builddriver_context_conformance_test.go` |
| `internal/skillspec` | `builddriver_conformance_test.go`, `builddriver_conformance_unix_test.go`, `builddriver_conformance_other_test.go` |
| `internal/whitelist` | `builddriver_context_conformance_test.go` |

Modified: `cmd/curator/status_test.go`, `internal/buildcache/conformance_test.go`,
`internal/closure/conformance_test.go`.

The three `_unix_test.go` / `_other_test.go` pairs use complementary constraints
(`aix||darwin||dragonfly||freebsd||linux||netbsd||openbsd||solaris` and its negation), so every host
compiles exactly one of each pair — no host is silently uncovered.

**Conflict check.** Under the recommended D3-P1 resolution 1pvfj5 owns `ci.yml`, `Makefile`, and six
`_test.go` files. Five of the six are in the accepted composite and are **not** in the candidate
delta — no conflict. The sixth, `internal/scopes/gc_conformance_test.go`, **is** in the delta and is
still owned by in-flight `jrrgw9`: route that one edit to jrrgw9, or apply it after jrrgw9 is
accepted. Under the D3-P3 fallback the surface stays two files and the intersection is empty.

**Explicitly not owned by 1pvfj5:** `README.md` (owned by `2qqq0w`, done), `.golangci.yml`
(byte-identical main↔candidate), `go.mod`/`go.sum`, any `conformance/` byte, 1pvfj5's own dependency
links, the committed pin value (owned by `38l1sy`).

---

## 5. Candidate delivery — one selected mechanism

Revision 1 left the producer choosing between a self-hosted workflow and local evidence. That choice
is made here, and revision 1's hosted `candidate (macos-latest)` / `candidate (windows-latest)` rows
are **deleted** as non-executable.

### 5.1 Why hosted candidate jobs cannot exist yet

The rc.5 candidate is an uncommitted `curator-spec` working tree (3 modified, 354 untracked) on top
of `v1.0.0-rc.3`. `vectors/build-drivers.json` — the file 15 test sites read — is itself untracked.
There is no ref for `actions/checkout` to fetch, and a Unix absolute path is meaningless to a hosted
runner. `TASK-260728-1jafds` independently recorded that `make regenerate-check` exits 2 and
`release_gate.py` exits 1 with "release gate requires a clean candidate checkout" for the same
reason.

Writing a hosted candidate job today means either fabricating a hash (forbidden by 1pvfj5 scope) or
shipping a job that can never fire. **Decision: 1pvfj5 adds no hosted candidate job.** The CI-side
candidate lane is deferred to `TASK-260720-38l1sy`, which already owns pin promotion and will have a
real revision to consume once rc.5 is committed to `relux-works/curator-spec`.

No self-hosted runner is assumed. This audit did not query GitHub (no network), so it makes no claim
about registered runner labels; the plan requires none.

### 5.2 Selected mechanism — frozen local snapshot, then native hosts

**C1 — freeze.** The authoritative root is live and has been observed moving under another producer.
Copy it once into a task-owned, read-only snapshot and never read the live tree again:

```bash
SRC=/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3nx97g/worktree/conformance/v1
DST=/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-1pvfj5/candidate/conformance/v1
mkdir -p "$(dirname "$DST")"
COPYFILE_DISABLE=1 cp -R "$SRC" "$DST"
chmod -R a-w "$DST"
```

**C2 — identity, on every host, before every candidate run.** Two independent checks: the manifest
digest named by the task brief, and a whole-tree digest that catches any other drift.

```bash
# macOS / Linux (use sha256sum on Linux)
cd "$DST"
shasum -a 256 manifest.json                      # must be b6f56aac…04c
find . -type f -print0 | LC_ALL=C sort -z | xargs -0 shasum -a 256 | shasum -a 256
                                                 # must be e6a13215…2fae
find . -type f | wc -l                           # must be 448
```

```powershell
# Windows, PowerShell
Set-Location $Root
(Get-FileHash manifest.json -Algorithm SHA256).Hash.ToLower()   # b6f56aac…04c
```

The producer **must re-measure both digests after the freeze** and record the measured values. If
they differ from the ones above, the live tree moved between this audit and the freeze: record the
new values as the task's candidate identity and say so — do not reuse this document's numbers as if
they had been re-measured.

**C3 — transport to native hosts.** Use the mechanism `TASK-260729-2sxx7k` already proved
reproducible: an explicit path manifest, a metadata-free archive, a SHA-256 checked remotely, and
extraction into a private mode-0700 directory. Its recorded trap: BSD `tar` embeds AppleDouble/xattr
records that GNU `tar` materializes as `._*` files, so the flags are mandatory.

```bash
COPYFILE_DISABLE=1 tar --no-mac-metadata --no-xattrs --no-acls --no-fflags \
  -cf .temp/TASK-260720-1pvfj5/candidate.tar -C "$(dirname "$DST")" conformance
shasum -a 256 .temp/TASK-260720-1pvfj5/candidate.tar
scp .temp/TASK-260720-1pvfj5/candidate.tar <host>:<private-dir>/
ssh <host> 'sha256sum <private-dir>/candidate.tar'     # must equal the local digest
ssh <host> 'tar -xf <private-dir>/candidate.tar -C <private-dir>'
# then re-run C2 remotely; only then run any gate
```

**C4 — the Windows-visible root.** `ssh win` is `cmd.exe`-fronted Windows 10 Pro 19045. The root
must be a native Windows path, and two recorded traps apply:

- `set VAR=1 && …` keeps the trailing space, so the value becomes `"1 "` and an env-gated test
  silently skips. Use the quoted form.
- `ssh win "powershell -EncodedCommand <b64>"` dies at exit 1 past the ~8191-char `cmd.exe` limit
  (8,660 chars failed, 6,372 worked). Transfer files; never inline them.

```
%LOCALAPPDATA%\Temp\TASK-260720-1pvfj5\conformance\v1
set "CURATOR_CONFORMANCE_ROOT=C:\Users\admin\AppData\Local\Temp\TASK-260720-1pvfj5\conformance\v1"
certutil -hashfile "%CURATOR_CONFORMANCE_ROOT%\manifest.json" SHA256
```

**Named preflight unknown the producer must resolve before relying on C3 for Windows.** `ssh win`
has no Git/Git Bash and no runnable Python, so a GNU `tar`/`scp` path is not guaranteed. Preflight,
in order: `ssh win "where tar"` (Windows 10 1803+ ships bsdtar as `C:\Windows\System32\tar.exe`),
then `ssh win "where sftp-server"` or a trial `scp`. If `scp` is unavailable, fall back to chunked
base64 writes kept under the 6,372-char encoded-command ceiling. Record which path was used.

### 5.3 Host availability — measured, not assumed

| Host | Platform | Go present? | Usable for 1pvfj5 candidate evidence? |
|---|---|---|---|
| local | macOS arm64 | **yes**, `go1.25.5 darwin/arm64` | **yes — the only host verified usable today.** PATH resolves a 411-byte goenv shim ahead of two native launchers and the shell exports a goenv `GOROOT`; use `GOENV=off GOTOOLCHAIN=local` and an explicit launcher, or the roots are misidentified. |
| `ssh relux` | macOS amd64 | **unverified since 2026-07-28** | preflight required — it ran gates on 07-28; no 07-29 inventory exists |
| `ssh win` | Windows 10 Pro 19045 amd64 | **no** — 2026-07-29 04:40 read-only inventory found no `go.exe` on process/user/machine PATH, no Go MSI/uninstall registry entry, and no tree at conventional MSI/archive/package-manager roots | **blocked** on an operator-installed, manager-approved Go 1.25.x amd64 root |
| `ssh lev` | Ubuntu 26.04 x86_64 | **no** — absent from PATH and from `/usr/local/go`, `/opt/curator/toolchains/go1.25.12`, `/usr/lib/go`, `/snap/go/current`; distro offers 1.26, outside the accepted `1.25` family | **blocked**, and non-gating regardless |

**The named Linux prerequisite, precisely:** an operator-installed, manager-approved absolute
Go 1.25.x `GOROOT` with a trusted `curator-go-toolchain-v1` identity on `ssh lev`, recorded as a
stop-the-line blocker on `TASK-260729-2sxx7k`. Linux *platform qualification* is separately owned by
`TASK-260728-1skseh`, the `until_task` named in the rc.5 exclusion vector. Until both land, **native
Linux validation stays non-gating and must not be attempted** — no auto-install, no download, no
ambient PATH. This is distinct from the hosted `ubuntu-latest` GitHub runner, which does have Go and
remains the right host for vet, lint, and the scoped Linux test and race jobs.

**State this plainly at handoff:** with `ssh win` and `ssh lev` both lacking Go, the candidate-root
evidence 1pvfj5 can produce today is **macOS arm64 only** (plus macOS amd64 if relux preflights
clean). Windows candidate evidence is blocked on an operator action. That limitation belongs in the
handoff to `TASK-260720-38l1sy`, not hidden behind a green macOS run.

---

## 6. Exact `.github/workflows/ci.yml` changes

Job by job. `test` splits, `race` is new, `lint`/`interop`/`naming-gate` change minimally.

### 6.0 Workflow-level `env` — the pin lives in exactly one place

```yaml
env:
  SPEC_PIN: 00b1688a9b2457ca397a0bb550acf47cad8ee967
```

Replace the literal at the current `ci.yml:28` and `ci.yml:81` with `${{ env.SPEC_PIN }}`. Every new
checkout of the suite references the same value (invariant I5).

### 6.1 `test` — drop `ubuntu-latest`, add the timeout

```yaml
  test:
    name: Test (${{ matrix.os }})
    runs-on: ${{ matrix.os }}
    strategy:
      fail-fast: false
      matrix:
        os: [macos-latest, windows-latest]        # ubuntu moves to the test-linux job
    steps:
      - uses: actions/checkout@v4
        with:
          submodules: true

      - name: Checkout authoritative protocol suite
        uses: actions/checkout@v4
        with:
          repository: relux-works/curator-spec
          ref: ${{ env.SPEC_PIN }}
          path: protocol-spec

      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod

      - name: gofmt check
        if: runner.os != 'Windows'
        run: |
          unformatted="$(gofmt -l cmd internal)"
          if [ -n "$unformatted" ]; then
            echo "gofmt: files need formatting:"; echo "$unformatted"; exit 1
          fi

      - name: go vet
        run: go vet ./...

      - name: go test
        env:
          CURATOR_CONFORMANCE_ROOT: ${{ github.workspace }}/protocol-spec/conformance/v1
        run: go test -count=1 -timeout 30m ./...
```

### 6.2 `test-linux` — new job: scoped execution plus the inventory rejection assertion

```yaml
  test-linux:
    name: Test (ubuntu-latest, inventory-scoped)
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          submodules: true

      - name: Checkout authoritative protocol suite
        uses: actions/checkout@v4
        with:
          repository: relux-works/curator-spec
          ref: ${{ env.SPEC_PIN }}
          path: protocol-spec

      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod

      - name: gofmt check
        run: |
          unformatted="$(gofmt -l cmd internal)"
          if [ -n "$unformatted" ]; then
            echo "gofmt: files need formatting:"; echo "$unformatted"; exit 1
          fi

      # Full breadth on purpose: vet compiles every package, including
      # internal/godriver, without executing a single test. Linux still
      # type-checks the whole tree.
      - name: go vet
        run: go vet ./...

      - name: godriver exclusion guard
        run: make linux-package-guard

      - name: go test (packages outside rc5-native-control-inventory-v1)
        env:
          CURATOR_CONFORMANCE_ROOT: ${{ github.workspace }}/protocol-spec/conformance/v1
        run: make test-linux

      - name: inventory rejection is asserted, not assumed
        run: |
          go test -count=1 -timeout 30m \
            -run 'TestProbeRejectsAnUncoveredPlatformBeforeTheWorker' \
            ./internal/godriver/
```

The last step is what keeps the exclusion honest: Linux does not skip `internal/godriver`, it proves
the package fails closed there. **The producer must confirm that test name exists in the composite
before writing it** — a Go `-run` pattern matching nothing exits 0, which would make the gate
vacuous. `TASK-260729-1y7okj` recorded exactly that regression class after a test rename.

### 6.3 `race` — new job, two runners

```yaml
  race:
    name: Race (${{ matrix.os }})
    runs-on: ${{ matrix.os }}
    strategy:
      fail-fast: false
      matrix:
        include:
          - os: macos-latest      # AC clause: go test -race ./... on the selected supported runner
            target: race-full
          - os: ubuntu-latest     # scope clause: race on at least Linux, named packages
            target: race
    steps:
      - uses: actions/checkout@v4
        with:
          submodules: true

      - name: Checkout authoritative protocol suite
        uses: actions/checkout@v4
        with:
          repository: relux-works/curator-spec
          ref: ${{ env.SPEC_PIN }}
          path: protocol-spec

      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod

      - name: go test -race
        env:
          CURATOR_CONFORMANCE_ROOT: ${{ github.workspace }}/protocol-spec/conformance/v1
        run: make ${{ matrix.target }}
```

`windows-latest` is deliberately absent: the Go race detector on Windows requires a C toolchain and
this audit has no measurement that the hosted image satisfies it. If the producer measures that it
does, add the row; if not, **record the absence explicitly** rather than dropping it silently.

### 6.4 `lint` — pin the linter version

```yaml
      - uses: golangci/golangci-lint-action@v7
        with:
          version: v2.4.0        # was: latest
```

`version: latest` is a mutable supply-chain input: a new golangci-lint minor turns the `lint` job
red with no repository change. `v2.4.0` is the version a sibling task ran against this tree; the
producer must confirm it is compatible with `.golangci.yml`'s `version: "2"` schema before
committing the value. Recorded trap for that version: a sudden `#nosec`-suppression failure on files
byte-identical to the accepted base is a **stale golangci-lint cache**, not a code defect — run
`golangci-lint cache clean` before hunting the code.

The config itself needs no change. The existing `gosec` excludes (`G306`, `G301`, `G122`, `G703`)
are narrow and documented, and the `_test\.go` gosec exclusion is scoped to tests — the AC's "no
broad suppression for new security code" is satisfied as-is.

### 6.5 `interop` and `naming-gate`

`interop`: replace the literal ref with `${{ env.SPEC_PIN }}` and add `-timeout 30m`.
`internal/interop` is a single test-only package (`golden_test.go`) that imports no godriver, so it
stays on `ubuntu-latest`, otherwise unchanged.

`naming-gate`: unchanged. Verified clean against the candidate tree — the assembled pattern matches
nothing outside `README.md`.

---

## 7. Exact `Makefile` recipes

Additions, plus one changed line in `test`. Existing targets otherwise keep their bodies.

```make
GO ?= go
VERSION ?= dev
LDFLAGS := -X github.com/relux-works/curator/internal/version.value=$(VERSION)

# --- TASK-260720-1pvfj5 additions -------------------------------------------
MODULE          := github.com/relux-works/curator
GODRIVER_PKG    := $(MODULE)/internal/godriver
# rc5-native-control-inventory-v1 covers macOS and Windows only. On every other
# host internal/godriver fails closed before the worker starts, so Linux runs
# every package except that one. The list is DERIVED at run time and never
# hand-maintained: a new package is included automatically.
LINUX_PKGS       = $(shell $(GO) list ./... | grep -v -x '$(GODRIVER_PKG)')
GODRIVER_IMPORTERS := $(MODULE)/cmd/curator $(MODULE)/internal/install
RACE_PKGS       := ./internal/transaction/... ./internal/buildcache/... \
                   ./internal/install/... ./internal/closure/... \
                   ./internal/skillspec/... ./internal/whitelist/... \
                   ./internal/scopes/... ./internal/runtimestore/...
# TASK-260729-2kaopg measured `go test ./...` exiting 1 at Go's 10-minute
# default in cmd/curator, twice. 30m is that task's recorded recommendation.
GOTESTFLAGS     := -count=1 -timeout 30m

.PHONY: build test fmt lint vet check \
        race race-full test-linux linux-package-guard \
        candidate-digest conformance-candidate check-ci

race:
	$(GO) test -race $(GOTESTFLAGS) $(RACE_PKGS)

race-full:
	$(GO) test -race $(GOTESTFLAGS) ./...

test-linux: linux-package-guard
	$(GO) test $(GOTESTFLAGS) $(LINUX_PKGS)

# Two invariants, both executable, both fail closed.
#  1. the excluded package still exists under the path the exclusion names
#  2. the set of packages that import it has not grown a third member that
#     might drive godriver.Build from a Linux test
linux-package-guard:
	@$(GO) list ./... | grep -q -x '$(GODRIVER_PKG)' \
	  || { echo 'guard: $(GODRIVER_PKG) no longer exists; the Linux exclusion is stale'; exit 1; }
	@got="$$($(GO) list -f '{{.ImportPath}} {{join .Imports " "}} {{join .TestImports " "}} {{join .XTestImports " "}}' ./... \
	          | grep '$(GODRIVER_PKG)' | awk '{print $$1}' \
	          | grep -v -x '$(GODRIVER_PKG)' | LC_ALL=C sort | tr '\n' ' ')"; \
	  want="$$(printf '%s\n' $(GODRIVER_IMPORTERS) | LC_ALL=C sort | tr '\n' ' ')"; \
	  [ "$$got" = "$$want" ] \
	    || { echo 'guard: godriver importer set drifted'; echo "  got:  $$got"; echo "  want: $$want"; exit 1; }
	@echo 'linux-package-guard: ok'

candidate-digest:
	@test -n '$(CANDIDATE_ROOT)' || { echo 'CANDIDATE_ROOT is required'; exit 1; }
	@test -n '$(CANDIDATE_MANIFEST_SHA256)' || { echo 'CANDIDATE_MANIFEST_SHA256 is required'; exit 1; }
	@got="$$(shasum -a 256 '$(CANDIDATE_ROOT)/manifest.json' | cut -d' ' -f1)"; \
	  [ "$$got" = '$(CANDIDATE_MANIFEST_SHA256)' ] \
	    || { echo "candidate manifest sha256 $$got != $(CANDIDATE_MANIFEST_SHA256)"; exit 1; }; \
	  echo "candidate manifest sha256 $$got"

# Candidate evidence. Never a release claim: the target name and every log line
# say "candidate".
conformance-candidate: candidate-digest
	CURATOR_CONFORMANCE_ROOT='$(CANDIDATE_ROOT)' $(GO) test $(GOTESTFLAGS) ./...

# Mirrors the CI gate exactly, unlike `check`, which walks the submodule.
check-ci: vet linux-package-guard
	@test -z "$$(gofmt -l cmd internal)" \
	  || { echo 'gofmt: files need formatting:'; gofmt -l cmd internal; exit 1; }
	$(GO) test $(GOTESTFLAGS) ./...
```

**One existing line changes:** `test:` becomes `$(GO) test $(GOTESTFLAGS) ./...`. Without the
timeout it is red for the reason in D6.

**Dependency graph:** `test-linux` → `linux-package-guard`; `conformance-candidate` →
`candidate-digest`; `check-ci` → `vet`, `linux-package-guard`. `race` and `race-full` have no
prerequisites. Nothing depends on `build`.

**Recipe → CI correspondence:**

| Make target | CI job / step |
|---|---|
| `test-linux` | `test-linux` → "go test (packages outside rc5-native-control-inventory-v1)" |
| `linux-package-guard` | `test-linux` → "godriver exclusion guard" |
| `race` | `race` matrix row `ubuntu-latest` |
| `race-full` | `race` matrix row `macos-latest` |
| `check-ci` | local mirror of the `test` job; not itself a CI step |
| `candidate-digest`, `conformance-candidate` | **no CI job** — native-only, §5 |

**Platform note.** This `Makefile` is Unix-shell only (`$$(…)`, `grep`, `awk`, `shasum`). The Windows
lane in §8 invokes `go test` directly and never calls `make`. `shasum` is macOS; on Linux use
`sha256sum`, or branch on `uname -s` — otherwise `candidate-digest` is macOS-only and should be
documented as such.

### 7.1 Linux package inventory (39 of 40)

Listed for review. The `Makefile` derives this at run time and does **not** hard-code it, which is
why new-package drift is structurally impossible; `linux-package-guard` covers the drift that *can*
happen — the godriver importer set growing a third member.

`./cmd/curator`, `./internal/adapters`, `./internal/audit`, `./internal/buildcache`,
`./internal/buildmeta`, `./internal/buildsource`, `./internal/capabilities`, `./internal/closure`,
`./internal/config`, `./internal/devsub`, `./internal/envfiles`, `./internal/gitignore`,
`./internal/gitops`, `./internal/globalbins`, `./internal/hashing`, `./internal/identifiers`,
`./internal/identity`, `./internal/install`, `./internal/install/atomicity`, `./internal/interop`,
`./internal/locale`, `./internal/managerlock`, `./internal/manifest`, `./internal/marker`,
`./internal/mcp`, `./internal/protocoljson`, `./internal/registry`, `./internal/runtimestore`,
`./internal/scopes`, `./internal/shell`, `./internal/skillcheck`, `./internal/skillspec`,
`./internal/snapshot`, `./internal/staging`, `./internal/transaction`, `./internal/ui`,
`./internal/verr`, `./internal/version`, `./internal/whitelist`.

**Excluded: `./internal/godriver` only.** `./internal/interop` is test-only (`golden_test.go`); it is
still a package `go list` reports and `go test` runs.

---

## 8. Executable platform matrix

**G** = gating required check, **N** = non-gating. Every row is a *future* producer gate; none has
been run.

| # | Lane | Runner label / host | Command | Root | Gate | Status |
|---|---|---|---|---|---|---|
| 1 | `test` | `macos-latest` (hosted) | `go vet ./...`, `gofmt -l cmd internal`, `go test -count=1 -timeout 30m ./...` | committed pin `00b1688a` | **G** | executable once D3 resolves |
| 2 | `test` | `windows-latest` (hosted) | `go vet ./...`, `go test -count=1 -timeout 30m ./...` | committed pin | **G** | executable once D3 resolves; see §8.2 risk |
| 3 | `test-linux` | `ubuntu-latest` (hosted) | `go vet ./...` (full breadth), `make linux-package-guard`, `make test-linux`, inventory rejection test | committed pin | **G** | executable once D1 + D3 resolve |
| 4 | `race` | `macos-latest` (hosted) | `make race-full` → `go test -race -count=1 -timeout 30m ./...` | committed pin | **G** | satisfies the AC verbatim |
| 5 | `race` | `ubuntu-latest` (hosted) | `make race` → 8 named package groups | committed pin | **G** | satisfies the scope verbatim |
| 6 | `lint` | `ubuntu-latest` (hosted) | `golangci-lint run` at a pinned version | — | **G** | vet-class only, Linux-safe |
| 7 | `interop` | `ubuntu-latest` (hosted) | `go test ./internal/interop/ -v -timeout 30m` | committed pin | **G** | no godriver import |
| 8 | `naming-gate` | `ubuntu-latest` (hosted) | inline `grep` | — | **G** | unchanged |
| 9 | candidate | **local macOS arm64** | `make conformance-candidate CANDIDATE_ROOT=… CANDIDATE_MANIFEST_SHA256=b6f56aac…04c` | frozen rc.5 snapshot | **N** | **the only candidate lane executable today** |
| 10 | candidate | `ssh relux` (macOS amd64) | same, after §5.2 C3 transport | frozen rc.5 snapshot | **N** | Go presence unverified since 2026-07-28 — preflight |
| 11 | candidate | `ssh win` (Windows amd64) | direct `go test`, §5.2 C4 root | frozen rc.5 snapshot | **N** | **blocked** — no Go on host as of 2026-07-29 04:40 |
| 12 | native Linux | `ssh lev` (Ubuntu 26.04) | — | — | **deferred** | **blocked** — no Go; also gated on TASK-260728-1skseh |

Rows 9–12 are **not** CI jobs. There is no hosted candidate job (§5.1).

### 8.1 Why macOS and Windows are both mandatory

Two protocol-defined platform-exclusive controls exist nowhere else:

- `TestBuildFailsClosedWhenTheGoChildCannotStart` — `boundary_test.go:267`, Windows only.
- `TestPerFileSizeLimitIsReallyApplied` — `build_test.go:433`, macOS only.

Windows additionally holds the only DACL / reparse / `.cmd` coverage:
`internal/buildcache/collect_windows_test.go`, `internal/buildcache/protection_windows_test.go`,
`internal/runtimestore/targets_windows_test.go`, `internal/scopes/gc_conservative_windows_test.go`,
`internal/scopes/gc_integration_windows_test.go`, `internal/managerlock/identity_windows_test.go`,
`internal/transaction/validation_windows_test.go`.

Unix holds the only ownership / no-follow / readonly-source / executable coverage:
`internal/buildsource/buildsource_special_unix_test.go`, `internal/buildcache/collect_unix_test.go`,
`internal/buildcache/protection_unix_test.go`, `internal/godriver/fingerprint_unix_test.go`,
`internal/transaction/root_metadata_unix_test.go`, `internal/adapters/fifo_unix_test.go`,
`internal/runtimestore/targets_unix_test.go`, `internal/scopes/gc_conservative_unix_test.go`.

Dropping either runner loses a class the AC explicitly requires.

### 8.2 Known Windows risk the producer must triage early

On 2026-07-28 a native `ssh win` full-suite run exited non-zero in **five packages this task does
not own** — `buildcache` (`owner does not match the effective user` on temp roots), `buildsource`
(Windows rename/path semantics), `globalbins` and `runtimestore` (`script command is not
executable`, POSIX exec-bit assumption), `shell` (PowerShell hook). All passed on both macOS hosts.
Recorded as "Reported, not fixed."

Three of those packages have since been reworked by tasks that are now `done` (`2284br`, `1ljev5`,
`1nlmvv`), but **no post-fix Windows measurement exists** — `ssh win` lost its Go before one could
be taken. Row 2 is therefore at real risk of being red for reasons outside 1pvfj5's ownership.
Measure it early (§9.1 step 4). If it is still red, that is a separate stop-the-line, not something
to absorb into the CI task.

---

## 9. Producer gate commands

**None of these has been executed. Every one is a future gate.** Record the real exit code of each.

### 9.1 Order — cheapest disproof first

1. **Compose on the right base.** New worktree from `origin/main` = `17804ce`, **not** `c06aa1a`
   (§1.1 C1). Apply the accepted `2kaopg` composite, then the 23-file `jrrgw9` delta. Then:
   ```bash
   grep -n 'ref:' .github/workflows/ci.yml     # must show 00b1688a… twice, unchanged
   ```
2. **Freeze and identify the candidate** — §5.2 C1 + C2. Record both measured digests.
3. **Disprove or confirm D3 (largest, cheapest).** One command, macOS, seconds:
   ```bash
   CURATOR_CONFORMANCE_ROOT=<pin-checkout>/conformance/v1 \
     go test -count=1 -timeout 30m \
       ./internal/buildsource/ ./internal/buildcache/ ./internal/scopes/ \
       ./internal/marker/ ./internal/whitelist/ ./internal/skillspec/
   echo "exit=$?"
   ```
   Predicted non-zero with `no such file or directory` in each. **A non-zero exit here is the
   expected-red confirmation of §3.3, not a defect** — report it with that rationale and do not
   present it as passing. If it exits 0, §3.3 is wrong and D3 evaporates; say so plainly.
4. **Measure Windows early** (§8.2), if and only if an operator has supplied Go on `ssh win`.
5. **Disprove or confirm D1.** Requires an approved Go 1.25.x GOROOT on `ssh lev`:
   ```bash
   CURATOR_CONFORMANCE_ROOT=<root> go test ./internal/godriver/ -count=1 -timeout 30m
   echo "exit=$?"
   ```
   Expected on macOS: 0. Expected on Linux: non-zero with `build_execution_control_unavailable` —
   again an **expected-red** result, to be reported as such. Until a Go root exists this step cannot
   run; say that rather than inferring the outcome.
6. **Take the D1–D6 decisions** with the board owner. Record each choice and its rationale.
7. **Edit.** `ci.yml` + `Makefile` (+ the six test files if D3 → P1). Apply I1–I8.
8. **Validate** §9.2 locally, §9.3 natively.
9. **Attach** exact commands, real exit codes, elapsed times, both candidate digests, per-platform
   evidence, and cleanup evidence. No release claim.
10. **Hand off** to review with the candidate CI evidence `TASK-260720-38l1sy` needs, including an
    explicit statement of which platforms could not be measured and why.

### 9.2 Local macOS gate set

```bash
go vet ./...                                        # exit 0 required
gofmt -l cmd internal                               # empty output required
make linux-package-guard                            # exit 0 required
golangci-lint run                                   # exit 0 required — see caveat
CURATOR_CONFORMANCE_ROOT=<pin-root> go test -count=1 -timeout 30m ./...
make conformance-candidate CANDIDATE_ROOT=<frozen-rc5-root> \
     CANDIDATE_MANIFEST_SHA256=b6f56aacc0e37dcc6692f73f641bff761e89b645adfe20a47a06d81c6fda204c
make race-full
make race
```

**`golangci-lint` caveat:** it was **not installed on this machine** at audit time
(`golangci-lint --version` → command not found). The producer must either install it at the pinned
version or state plainly that the local lint gate did not run. **Do not check the lint checklist
item off a CI-only green.**

### 9.3 Native supplemental

Same commands, native, after §5.2 C3/C4 transport and identity verification. `ssh win` is blocked
until an operator installs Go — **do not install it from the agent, do not accept ambient PATH.**
Windows `-race` additionally needs a C toolchain; if absent, run Windows without `-race` and record
the absence explicitly rather than omitting the row.

### 9.4 Gate hygiene (evidence-honesty contract)

Run each gate as a standalone foreground process. Never pipe it through `tee` or an unguarded pipe —
the pipe's exit status is not the gate's. To capture output use
`go test … > .temp/TASK-260720-1pvfj5/gate-NN.log 2>&1; echo "exit=$?"`, or enable `set -o pipefail`
and read `PIPESTATUS`. On `ssh win`, `cmd … & echo %ERRORLEVEL%` expands at parse time and **always
prints 0** — a recorded trap that invalidated a whole Windows evidence set; use runtime
`&& echo EXIT_0 || echo EXIT_NONZERO` instead.

---

## 10. Fact-check ledger

Every claim traces to one of these reads. All are read-only and all ran during this audit.

| # | Claim | Command | Result |
|---|---|---|---|
| 1 | `main` = `origin/main` = `17804ce`; `c06aa1a` is divergent | `git rev-parse origin/main`; `git merge-base --is-ancestor` both directions; `git merge-base` | `17804cea…`; **NO** in both directions; merge base `ecb6c1a` |
| 2 | `c06aa1a` is a branch tip, not main | `git log --oneline -1`; `git branch -a --contains 17804ce` | tip of `agent/link-curator-skill-registry`; `17804ce` on `main` and `origin/main` |
| 3 | Committed pin is `00b1688a…` | `git show 17804ce:.github/workflows/ci.yml \| grep 'ref:'` | `00b1688a…` at lines 28 and 81 |
| 4 | Pin/tag positions | `git describe --tags` on four revisions | `00b1688a`→`v1.0.0-rc.2-1-g00b1688`; `57c1f56`→`v1.0.0-rc.3`; `e72defe` and `6c9b1cf`→"No tags can describe" |
| 5 | `e72defe` is the branch-only pin, ancestor of `00b1688a` | `git log -L 28,28:.github/workflows/ci.yml`; `git merge-base --is-ancestor e72defe 00b1688a` | `6c9b1cf`→`e72defe` at commit `0f63a8d`; ancestor **YES** |
| 6 | No rc.4 / rc.5 tag | `git for-each-ref refs/tags` in curator-spec | `v1.0.0-rc.1`, `-rc.2`, `-rc.3` only |
| 7 | Pin capability matrix (§4.4) | `git cat-file -e <rev>:<path>` over 12 paths × 3 revisions; `git ls-tree <rev> conformance/v1/vectors/` | as tabulated; `build-drivers.json` absent at all three |
| 8 | **rc.5 root dirt (corrects revision 1)** | `git status --short --untracked-files=all -- conformance/` | **3 modified, 354 untracked** (357 lines). Modified: `conformance/README.md`, `conformance/v1/manifest.json`, `conformance/v1/schema-cases/index.json`. Collapsed default: 94 lines = 3 M + 91 `??`. `build-drivers.json` is itself `??` |
| 9 | rc.5 identity | `shasum -a 256 manifest.json`; `find … \| sort -z \| xargs -0 shasum -a 256 \| shasum -a 256`; `find . -type f \| wc -l` | `b6f56aac…04c` **exit 0**; `e6a13215…2fae` **exit 0**; 448 files |
| 10 | Go 1.25.5 | `head -3 go.mod`; `grep 'go-version-file' ci.yml` | `go 1.25.5`; `go-version-file: go.mod` |
| 11 | 40 package dirs; only `interop` is test-only | `find cmd internal -name '*.go' -not -path '*/testdata/*'` deduped to dirs; per-dir non-test `.go` count | 40 dirs; `internal/interop` has 0 non-test `.go` files |
| 12 | Candidate delta = 23 product files, all `_test.go` | `diff -rq` accepted vs candidate, excluding `.git`/`.task-board` | 24 entries: 21 `Only in` (one is `.temp`) + 3 `Files differ`; every product path ends `_test.go` |
| 13 | `.golangci.yml` byte-identical, schema v2 | `diff -q`; `head -3` | IDENTICAL; `version: "2"` |
| 14 | No race gate today | `grep -n race .github/workflows/ci.yml Makefile` | no matches |
| 15 | Disk headroom | `df -h /Users/iv` | 25 GiB free |
| 16 | Linux outside the inventory | `controls.go:75,200,241`; `controls_other.go`; `conformance-claim-v3-qualification.json` | exhaustive over macos+windows; linux `excluded`, `until_task TASK-260728-1skseh` |
| 17 | godriver importers are exactly two | `grep -rl 'internal/godriver' --include='*_test.go' cmd internal`; same without `_test.go` | tests: `cmd/curator` (4 files), `internal/install/stage_test.go`. prod: `cmd/curator` (2), `internal/install` (4) |
| 18 | Six hard-`t.Fatal` sites | direct read of each file at the cited lines | `buildsource:16-19`, `buildcache:15` + `readJSONObject:63-66`, `scopes:38-41`, `marker:37-41`, `whitelist:20-28`, `skillspec:106-109` — all `t.Fatal`, none guarded by an artifact-presence check |
| 19 | Five new delta tests DO guard | `grep -n 'Skipf' <5 files>` | `t.Skipf("%s publishes no build-drivers vector", root)` at `skillcheck:24`, `whitelist:25`, `skillspec:39` and `:301`, `buildcache/builddriver_positive:31` |
| 20 | Main is green because its tree predates this | `ls internal/godriver` at main (absent); `grep 'filepath.Join(root' internal/{closure,skillspec}/conformance_test.go` at main | `internal/godriver` does not exist at main; main reads only `closures.json` and `portable-paths.json`, both present at the pin |
| 21 | Dependency states | `task-board q 'get(<id>) { full }'` × 6 | `1pvfj5` backlog/blocked; `2qqq0w` done; `jrrgw9` development (its own three blockers all done); `1skseh` backlog |
| 22 | 10-minute timeout; `ssh win`/`ssh lev` have no Go; transport mechanism; Windows quoting/length traps; five red Windows packages | `LOGBOOK.md` entries 0607, 0440, 0510, 0912, 1247 | quoted inline in §3.6, §5.2, §5.3, §8.2, §9.4 |

**Not verified, by design:** no Go compile, test, vet, gofmt, lint, or build was executed; no network
fetch; no host mutation; no board mutation outside this task's own status and resources. Every
red-prediction in §3.3, §3.6, §4.3 and §8.2 is static analysis or a cited prior measurement, each
carrying a named confirmation command in §9.1. **No claim in this document is a green CI result.**

---

## 11. Invariants the producer must preserve

- **I1.** The candidate enters only through a non-default input, defaulting to empty. With it empty,
  CI uses the committed pin and nothing else.
- **I2.** Every candidate run asserts the manifest digest **and** the tree digest before executing a
  case, and emits both into the log. A mismatch fails the run. The digest is never written as a `ref:`.
- **I3.** No job, target, or log line labels candidate output a release, a published conformance
  claim, or "rc.5 qualified". They say *candidate*.
- **I4.** 1pvfj5 does not move the committed pin. `TASK-260720-38l1sy` owns promotion after
  `TASK-260720-25d05o`.
- **I5.** The pin appears in exactly one place (`env.SPEC_PIN`). Every suite checkout references it.
- **I6 (corrected).** The composite is built on `origin/main` = `17804ce`, **not** on `c06aa1a`.
  Building on `c06aa1a` would revert the pin from `00b1688a` to `e72defe` and break
  `internal/closure`, which needs `manager-lifecycle.json`.
- **I7.** Every `go test` invocation carries `-timeout 30m`. A default-timeout run is red (D6) and
  must never be reported as green.
- **I8.** No `-run` selector is added without confirming it matches a test that exists in the
  composite. A Go `-run` matching nothing exits 0 — a vacuous gate reads exactly like a passing one.
