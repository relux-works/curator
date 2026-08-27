# TASK-260729-2kaopg — global-status gate rework: what changed and what it cost

Date: 2026-07-29
Role: developer (focused performance rework, RUN-260729-589a90)
Candidate tree: `.temp/TASK-260729-2kaopg/worktree`
Accepted currentness base: `.temp/TASK-260720-1nlmvv/worktree`
Accepted fingerprint patch: `TASK-260729-1zex8r_fingerprint-cycle3.patch`

Nothing is staged or committed. No host software was installed. No pin was
moved. No timeout was changed, no test was skipped, no `-run` exclusion was
used, and no cached result was accepted as evidence. Every gate below ran as a
standalone process with output redirected to a file — no `tee`, no pipe — and
every exit code reported is the real process exit code.

---

## 1. The problem this pass had to solve

The integrated candidate was semantically green but could not pass the required
default-timeout gate: `cmd/curator` reached the standard 10-minute package
timeout at 602.193s. `TASK-260729-2kaopg_timeout-diagnostic.md` attributed the
cost precisely — the global-status file drove 26 expensive read-only plans over
compiled scopes and three compiled installations that existed only to reset a
fixture, and every such plan fingerprints the whole trusted GOROOT.

This pass implements that diagnostic's recommendation. It changes no currentness
semantics and no assertion outcome.

---

## 2. Production change: `curator global status` is now three phases

`cmd/curator/main.go` only. The command is decomposed into the phases it already
performed in that order:

1. `parseGlobalStatusOptions(args) (globalStatusOptions, int)` — parse the
   request; the machine-wide scope still takes no positional target.
2. `globalStatusAcquire` — the acquisition phase type. `cmdGlobalStatus` always
   passes `globalStatusPlan`, so every real invocation still acquires exactly
   one fresh read-only plan of its own.
3. `reportGlobalStatus(cfg, opts, acquire) int` — fingerprint installed markers,
   acquire, bracket the classification, render the requested document, apply the
   fail-closed `--check`.

Behaviour is unchanged, including ordering. `before := markerDigests(...)` is
still taken *before* acquisition and re-read after classification, so the
race-window bracketing is byte-for-byte the same decision procedure. The only
new fact is that the classify/render/check phase can be driven from a plan the
caller already holds.

No change to `cmd/curator/builds.go`, to `statusReport`, to the project-scope
`cmdStatus`, to `internal/`, or to the documented contract. `README.md` needed
no edit: the `--json`/`--check` contract it documents at lines 232–260 is
unchanged by an internal phase split.

---

## 3. Test change: `cmd/curator/global_status_test.go`

Every case still proves the same complete contract — machine-readable row,
demoted declared-skill entry, fail-closed exit, operator line, and the zero exit
of the plain report. What changed is where the plan under classification comes
from, and that `--json` and `--check` are now proven by one invocation instead
of two.

| Case | Plan source | Why |
| --- | --- | --- |
| unchanged compiled installation | end-to-end CLI | headline contract; proves the whole command path |
| recorded build-source identity drifted | end-to-end CLI | pins the replayed group to what the CLI publishes |
| recorded logical key drifted | replay of the unchanged plan | a rewritten marker cannot change what a plan derives |
| marker records no build for the command | replay of the unchanged plan | same |
| build root reached agent-facing context | end-to-end CLI | exposure lives in the installed skill, not in the plan |
| protected cache entry cannot be interpreted | own acquisition while tampered | the plan itself reads protected cache evidence |
| protected cache holds no entry | own acquisition while tampered | same |
| trusted Go toolchain cannot be resolved | end-to-end CLI, reusing this fixture | refusal is what the acquisition phase must report |
| transitive compiled command, current and drifted | end-to-end CLI | proves resolution through a store no declaration names |

Replays go through `reportGlobalStatus` — the production classifier, the
production renderer, and the production fail-closed decision. Nothing is
re-implemented in the test.

Structural changes:

- `TestGlobalStatusReportsAnUnusableToolchainPerCompiledCommand` is now the
  `trusted go toolchain cannot be resolved` subtest of
  `TestGlobalStatusReportsCompiledCurrentnessAndFailsCheck`, reusing the default
  compiled installation instead of building a second identical one. Every
  assertion it carried is preserved verbatim; it gained one — that the plain
  report over an unusable toolchain still exits zero.
- `reinstallGlobal` is gone. The two cache-tamper cases now snapshot every
  protected cache byte and permission bit and restore them in cleanup
  (`snapshotBuildCacheAfter`). Running the real machine-wide reconciliation to
  repair a fixture compiled the command again and proved nothing this file owns;
  install, repair, and rollback semantics keep their own dedicated tests, which
  were not touched or reduced.
- One assertion was added, not removed: the human line must now also carry
  `cache=<outcome>` for every non-current state.

### Cost

| | Before | After |
| --- | ---: | ---: |
| expensive plans over a compiled scope | 26 | 8 |
| real compiled installations | 4 | 2 |
| top-level `TestGlobalStatus*` tests | 8 | 7 (one merged as a subtest) |

The target in the directive was six expensive plans. Eight is the floor that
keeps the required end-to-end coverage: the two protected-cache states and the
transitive drift state are observed by planning itself, so a replayed plan would
describe cache evidence that no longer exists, and each needs its own
acquisition or CLI run. The remaining six are the unchanged end-to-end run, one
shared immutable acquisition, and one end-to-end run for each of the three cases
the directive names that a replay could otherwise have covered.

---

## 4. Gate evidence

Working directory for every command: `.temp/TASK-260729-2kaopg/worktree`.

| Gate | Exit | Result |
| --- | ---: | --- |
| `gofmt -l .` | 0 | no output |
| `git diff --check` | 0 | no output |
| `go build ./...` | 0 | — |
| `go vet ./...` | 0 | — |
| `go test -count=1 -timeout 10m -run '^TestGlobalStatus' ./cmd/curator` | 0 | **43.134s** |
| `go test -count=1 -timeout 10m -run '^TestGlobalStatus' -coverprofile ./cmd/curator` | 0 | **43.022s**, package coverage 28.9% |
| `go tool cover -func` | 0 | see below |
| `go test -count=1 ./cmd/curator` (standard default timeout) | PENDING | PENDING |
| `command -v golangci-lint` | 1 | executable absent; not installed per directive |

Focused-run comparison against the same command on this candidate before the
rework, as recorded by tester RUN-260729-2655bc: **110.541s → 43.022s**, a
67.5s reduction on the coverage run.

An earlier `go test -count=1 ./cmd/curator` was started against an intermediate
revision and cancelled at 159s before it produced a result. It is not evidence
and its partial artifacts were deleted; the row above is a fresh run against the
final source.

**The host was not quiet.** During the package gate the 1-minute load average
was 26.2–26.6, with `gramdrive-agent` at ~100% CPU, `suggestd` at ~97%, and
Spotlight `mds`/`mds_stores` at ~80% — none of them this task's processes. The
package duration below is therefore an upper bound taken under contention, not a
clean-host figure. Raw: `logs/cmdcurator-gate-01-host-load.txt`. Every focused
number above was taken under the same kind of load, so the 110.541s → 43.022s
comparison is not flattered by a quieter machine.

Owned-surface coverage after the rework:

| Function | Before | After |
| --- | ---: | ---: |
| `cmdGlobalStatus` | 96.9% | 100.0% |
| `parseGlobalStatusOptions` | — | 90.0% (the flag-parse-error branch is reached by other CLI tests) |
| `reportGlobalStatus` | — | 100.0% |
| `globalStatusPlan` | 100.0% | 100.0% |
| `globalStatusScope` | 100.0% | 100.0% |
| `statusReport` | 86.7% | 86.7% |
| `classifySkillBuilds` | 81.8% | 81.8% |
| `installedSkillDir` | 80.0% | 80.0% |
| `checkFailed`, `factsOf`, `Describe`, `plannedRows`, `demoteSkill`, `markerMoved` | 100.0% | 100.0% |

`recheckBuildCache` (66.7%) and `changedDuringCheck` (66.7%) are shared with the
project-scope status path and are exercised by `TestStatus*`, which is outside
this focused selection.

Raw logs, all under `.temp/TASK-260729-2kaopg/logs/`:
`rework-globalstatus-01.log`, `rework-globalstatus-cover.log`,
`rework-globalstatus-cover-func.log`, `rework-globalstatus.cover`,
`cmdcurator-gate-01.log`, `cmdcurator-gate-01-exit.txt`,
`cmdcurator-gate-01-disk-before.txt`, `cmdcurator-gate-01-disk-after.txt`.

---

## 5. Provenance after the rework

`diff -rq` between the accepted currentness base and the candidate, excluding
VCS/board/scratch trees and the base's stray compiled `curator` binary
(`logs/provenance-diffrq-rework.txt`, exit 1 — files differ, the expected
result):

```text
README.md                                          differ   (owned)
cmd/curator/builds_test.go                         differ   (owned, call-site rewire)
cmd/curator/main.go                                differ   (owned)
cmd/curator/status_test.go                         differ   (owned, call-site rewire)
cmd/curator/global_status_test.go                  candidate only (owned, new)
internal/godriver/fingerprint.go                   differ   (accepted patch)
internal/godriver/fingerprint_equivalence_test.go  candidate only (accepted patch)
```

The file set is exactly the one the previous integration cycle proved. No
accepted currentness file is reverted or absent.

The accepted fingerprint delta is untouched by this pass:

| File | SHA256 | Matches |
| --- | --- | --- |
| `internal/godriver/fingerprint.go` | `560d0c98c665a5a83c3a6989a7b0cdcc9f26c4fb513c7688d9b1bd6e42552d1d` | the reconstruction from the accepted patch |
| `internal/godriver/fingerprint_equivalence_test.go` | `6390e75c9848f575f2f4b50217ebd1d53481a58d349073fb0e819491b5fed484` | the reconstruction from the accepted patch |
| `TASK-260729-1zex8r_fingerprint-cycle3.patch` | `a7e0906612ce6f007bfdb3776de632dd9c7a673e9b501443be5fb3eced8f1beb` | the cycle-3 reviewer verdict |

Owned delta after the rework: `logs/owned-delta-rework.patch`, 1193 lines,
sha256 `5becface29bdc22cb82f14efb95264a9590b09c4530a5665b1e663dc5ca028bd`.
`cmd/curator/builds_test.go` and `cmd/curator/status_test.go` still carry only
the call-site rewires forced by the owned API change; no assertion in either
file was altered by this pass or by any earlier one. Their complete non-context
delta against the accepted base is seven lines removed and eight added, all of
the form `statusStores(cfg, project)` → `scope.stores` and
`statusReport(cfg, project, …)` → `statusReport(cfg, scope, …)`, with
`scope := projectStatusScope(cfg, project, "app")` introduced where needed.

---

## 6. Not run in this pass, and why

- The two consecutive literal `go test -count=1 ./...` gates with the standard
  default timeout. The directive reserves them for an independent Codex tester,
  and the host is not in a fit state to run them honestly: `/System/Volumes/Data`
  reports 100% capacity with roughly 6.6 GiB available, which is the low-disk
  condition that produced the earlier runaway-testlog incident.
- `golangci-lint`. The executable is absent (`command -v` exit 1) and the
  directive forbids installing host software. The preserved implementer lint
  artifact reports 0 issues, and that artifact predates this pass.
