# BUG-260906-1bdotx — registry snapshot future-timestamp flake on Windows

Repository `curator`, branch `fix/snapshot-timestamp-flake`, base `919e2e9c`, worktree
`/Users/iv/Developer/ReluxWorks/.worktrees/curator-snapshot-flake`.
Three signed commits, not pushed, no PR: `a8a64197`, `e8b7129c`, `879b884b`.

---

## 1. The mechanism, and how it was established

**Established from evidence, and reproduced end to end.** The brief's arithmetic was right and its
premise was wrong: nothing needs to be five minutes ahead, because **the 300-second default never
reaches this code path**.

Two facts compose. Neither fires the gate alone.

### Fact 1 — the e2e config carries a literal zero skew, not 300 seconds

`internal/config/config.go:709` (`parseAudit`) is the *only* place the audit defaults are
established. `newEnv` (`internal/install/install_test.go:32`) and `newEnvIn`
(`internal/install/commit_test.go:1067`) build `&config.Config{...}` as struct literals and never go
through it, so the whole `Audit` block is the zero value.

Measured, by driving `newEnv` and printing what it holds:

```
PROBE env audit policy: SnapshotClockSkewSeconds=0 SnapshotMaxAgeSeconds=0 RegistryPolicy=""
PROBE production default: DefaultSnapshotClockSkewSeconds=300
```

`internal/registry/snapshot.go:62-64` documents the reading this produces, and it is deliberate:

> A zero clock skew is literal; a zero max age retains the historical default for callers of
> `CheckSnapshots`.

So `maxAge` falls back to seven days, and `clockSkew` **stays zero**. In this test's path the gate at
`snapshot.go:158` is effectively `parsed.CreatedAt.After(now)` — no tolerance whatsoever. The error
text says "too far in the future"; the actual excess is sub-millisecond.

### Fact 2 — the fixture minted `created_at` after the checker sampled `now`

`resolveRegistries` evaluates `time.Now()` as a *call argument* at `internal/install/install.go:1192`,
so `now` is fixed **before** `registry.HTTPGetSnapshot` runs. The fixture then stamped `created_at`
inside the HTTP handler (`internal/install/registry_e2e_test.go:44`), at a strictly later instant.

RFC3339 truncates downward, which is exactly what makes this bite. Let `t₁` be the checker's `now`,
`t₂` the handler's clock reading, `Δ = t₂ − t₁`. The served timestamp is `floor(t₂)`, and

```
floor(t₂) > t₁   ⟺   frac(t₁) + Δ ≥ 1s   ⟺   the fetch crossed a whole-second boundary
```

With `frac(t₁)` uniform, **P(flake) = Δ per snapshot fetch**. The snapshot lands ahead of `now` by
less than `Δ` — micro- to milliseconds — and under a literal zero tolerance that is tampering.

The brief's candidate list is therefore resolved: the wall clock did **not** step backwards; there is
**no** timezone or parse asymmetry (`parseSnapshot` round-trips `created_at` against
`"2006-01-02T15:04:05Z"` and rejects anything inexact); the timestamp **is** produced in the handler,
per fetch; and **the configured skew is not the default in this test's path** — that last candidate
is the one that holds.

### The reproduction

A probe fixture identical to the real one except that the handler waits past the next whole-second
boundary before responding, driven through the production `e.install(Options{})` path:

```
PROBE cross=true  status="failed"
  messages=[... test: registry: registry test-reg snapshot timestamp is too far in the future]
  errors=[every trusted audit registry served a tampered snapshot]
PROBE cross=false status="ok" errors=[]
```

Both reported strings, verbatim, from the production entry point. The control passes.

### The window, measured

`install.go:1192` was temporarily instrumented to publish the `now` it hands the checker, and the
fixture handler to publish its own reading. Twelve real installs, darwin/arm64, then repeated under
`-race`. The instrumentation was reverted before any committed change.

| Lane | Mean window | Worst window | P(flake) per snapshot fetch |
| --- | ---: | ---: | ---: |
| plain | 667 µs | 3.18 ms | 0.067 % (worst 0.32 %) |
| `-race` | 856 µs | 1.45 ms | 0.086 % (worst 0.14 %) |

Roughly 1 in 1500. Consistent with 0 failures in a 120-run `-race` stress of the unmodified
`TestRegistryAttestationLandsInMarker` on this host (P(0 failures) ≈ 92 %), and consistent with the
defect surfacing once, on the slowest lane in the matrix, in a 39m44s job.

**Bound — no Windows result is claimed.** The window magnitude on `windows-latest` was **not
measured**; nothing here was run on a hosted runner. The Windows-only observation is *explained* by
this mechanism (P scales linearly with a window that Defender-mediated filesystem calls, the Windows
loopback stack and matrix load all widen), but that explanation is an argument, not a measurement.
`ledger-consistency.sh` proves the new cases **compile** on `GOOS=windows` — also an argument, not a
measurement.

---

## 2. The fix, and why it does not weaken the gate

Test-only. **The gate is not touched.** `snapshot.go:158` is byte-identical to base, `clockSkew` is
not widened anywhere, no assertion was removed, and no tolerance was introduced.

| Change | File | Why |
| --- | --- | --- |
| Mint `created_at` once, at server construction | `internal/install/registry_e2e_test.go` (`fakeRegistry`) | A registry *publishes* a snapshot and then serves it. Stamping per request is what let the timestamp drift past a `now` already sampled. |
| Same, second fixture with the identical defect | `internal/install/dryrun_conformance_test.go` (`loopbackRegistry`) | Found by scanning for the shape; same race, same package. |
| Two fixture options + a boundary hook | `internal/install/registry_e2e_test.go` | Lets the regression test force the crossing, and lets a test stamp a genuinely future snapshot. |
| Four new cases in the platform ledger | `.github/ci/platform-cases.tsv` | The only lane that ever saw this was slow by luck. Now every runner must execute the bound by name. |

This is the shape the brief endorsed — "a deterministic clock the checker and the fixture share" —
rather than a tolerance. The frozen timestamp is minted before `e.install()` is called, so it is
unconditionally in the past relative to the checker's `now`, at any skew including zero.

**What was deliberately *not* changed.** Giving `newEnv` the production audit defaults would also make
the symptom disappear — by raising the effective tolerance in the test path from 0 s to 300 s. That is
a tolerance dressed as a config fix, it would silently absorb the next fixture with the same defect,
and it reads directly against the AC. Rejected. `TestSnapshotZeroClockSkewIsLiteral` now pins the
zero-skew semantics instead, so the policy the install fixtures actually run under is stated rather
than assumed.

**Stated bound (production, unchanged, reported not fixed).** With `snapshot_clock_skew_seconds: 0`,
a registry that mints `created_at` at *response* time will be reported as tampered with probability
≈ fetch-window ÷ 1 s, because one-way latency guarantees the snapshot can be minted after the client
sampled `now`. The shipped 300 s default absorbs this by five orders of magnitude. No production code
was changed for it: zero means zero is the documented contract, and a real registry serves a published
snapshot.

---

## 3. Mutants

Harness `.temp/BUG-260906-1bdotx/mutants.sh`. Each mutant edits the gate, runs the **behavioural**
suite of both packages as a standalone process, and restores the tree. No survivors.

| Mutant | What it narrows the gate to | Named test that fails | Result |
| --- | --- | --- | --- |
| **M1 narrowing** `After(now.Add(clockSkew + time.Second))` | Gate present, threshold one second out — admits exactly the one-second-past-the-bound class, still refuses everything further | `TestSnapshotFutureBoundIsExactAtEveryConfiguredSkew/{0s,30s,5m0s}/one_second_past_the_bound`, `TestSnapshotZeroClockSkewIsLiteral`, `TestSnapshotRequiresCompleteShapeAndRejectsEquivocation` | killed, exit 1 |
| **M2 narrowing** `After(now.Add(max(clockSkew, time.Second)))` | Gate present, a one-second floor under the configured skew — admits exactly the sub-second/one-second future class when skew is 0, identical behaviour at every larger skew | `TestSnapshotFutureBoundIsExactAtEveryConfiguredSkew/0s/one_second_past_the_bound`, `TestSnapshotZeroClockSkewIsLiteral` | killed, exit 1 |
| **M3 delete** `if false` | Assertion gone | the above, plus `TestRegistryFutureSnapshotDeniesInstallThroughResolveRegistries` | killed, exit 1 |
| **M0 fix-revert** `created_at` back to per-request, gate untouched | The pre-fix fixture | `TestRegistrySnapshotSurvivesASecondBoundaryDuringFetch`, failing with the exact reported text | killed, exit 1 |
| baseline | — | — | exit 0 |

`created_at` is an exact UTC seconds timestamp by contract (`parseSnapshot` rejects anything else), so
**one second is the smallest future step the schema can express** — M1 and M2 admit exactly one class
in the schema's own granularity.

M2 is the mutant that matters most here: it is precisely the shape a careless fix for this ticket
would have taken.

**Bound — where the narrowing kill lands.** M1 and M2 are killed by `internal/registry` only;
`internal/install` stays green under both. The narrowing bound is proven at
`registry.CheckSnapshotsWithPolicy`, which is the exact function `install.resolveRegistries` calls at
`install.go:1191`. `TestRegistryFutureSnapshotDeniesInstallThroughResolveRegistries` carries the
production-call-site burden and kills the delete-only mutant; it uses a +1 h offset and does not kill
a one-second narrowing, because a test outside the process cannot see the `now` the checker sampled
and any attempt to hit a one-second edge from there would itself be timing-dependent. Before this
change, **neither** mutant was visible to `internal/install` at all.

`m.2` of the DoD (a gate inspecting source text, attacked by a token-preserving mutant) does not apply:
this gate compares two `time.Time` values and reads no source text.

---

## 4. Tests

| Test | Package | Drives | Proves |
| --- | --- | --- | --- |
| `TestRegistrySnapshotSurvivesASecondBoundaryDuringFetch` | `internal/install` | `Project` → `resolveRegistries` | the regression: a snapshot minted before the run installs even when the fetch crosses a boundary |
| `TestRegistryFutureSnapshotDeniesInstallThroughResolveRegistries` | `internal/install` | `Project` → `resolveRegistries` | **negative**: a genuinely future-dated snapshot *denies* the install, naming both strings |
| `TestRegistrySnapshotAtTheSkewBoundIsAcceptedThroughInstall` | `internal/install` | `Project` → `resolveRegistries` | the refusal above is not a gate that rejects everything |
| `TestSnapshotFutureBoundIsExactAtEveryConfiguredSkew` | `internal/registry` | `CheckSnapshotsWithPolicy` | **negative, narrowing target**: the bound is exactly the configured skew, at 0 / 30 s / 5 m, walking one second inside → on → one second past → an hour past |
| `TestSnapshotZeroClockSkewIsLiteral` | `internal/registry` | `CheckSnapshotsWithPolicy` | **negative**: zero is literal, not a fallback; and the shipped default absorbs sub-second drift |

**AC coverage: 4 of 4 rows driven.**

| AC row | Driven by | Production call site |
| --- | --- | --- |
| mechanism established from evidence and stated | §1 above — probe reproduction + instrumented window measurement | `install.go:1192` → `snapshot.go:158` |
| gate still refuses a future-dated snapshot; narrowing mutant fails a named test; the test exists after this change | M1/M2 vs `TestSnapshotFutureBoundIsExactAtEveryConfiguredSkew` (new); M3 vs `TestRegistryFutureSnapshotDeniesInstallThroughResolveRegistries` (new) | `registry.CheckSnapshotsWithPolicy`, called from `install.go:1191` |
| skew not widened, assertion not removed | `snapshot.go` byte-identical to base; `git diff 919e2e9c..HEAD -- internal/registry/snapshot.go` is empty | — |
| no unmeasured Windows result claimed | stated as a bound in §1 | — |

---

## 5. Gates

Every command a standalone process; exit code observed directly, no pipe.

| Gate | Exit | Notes |
| --- | ---: | --- |
| `go build ./...` | 0 | |
| `go vet ./...` | 0 | |
| `gofmt -l cmd internal` | 0 | no output |
| `golangci-lint run ./...` | 0 | `0 issues.` |
| `bash .github/ci/gate-selftest.sh` | 0 | `130 passed, 0 failed` |
| `bash .github/ci/ledger-consistency.sh <evidence>` | 0 | `235 rows checked across linux darwin windows` (was 231) |
| `go test -count=1 -race ./internal/install/ ./internal/registry/` | 0 | 82.6 s / 3.9 s |
| `bash .github/ci/test-gate.sh` — plain lane | 0 | `go test exit=0, platform-case gate exit=0`; served=69 deferred=4 excluded=0 |
| `bash .github/ci/test-gate.sh` — `-race` lane | 0 | `go test exit=0, platform-case gate exit=0` |
| `go test -count=25 -race -run '^TestRegistrySnapshotSurvives…$'` | 0 | 25 `=== RUN`, 25 PASS, 0 FAIL — 25 forced boundary crossings |

Both `test-gate` lanes ran sequentially against the pinned conformance root, materialized as a plain
`git clone --no-checkout` + `git checkout --detach 0ed5c691e9208eea52f21db2fc05e226ce3516fd` of
`curator-spec` (`SPEC_PIN` from `ci.yml:44`) — **not** `git archive` — and verified against its own
`manifest.json`: **691 files, 0 mismatches, 0 missing**.

Both lanes observed all four new cases by name through the real `go test -json` stream:

```
ok    internal/install :: TestRegistrySnapshotSurvivesASecondBoundaryDuringFetch
ok    internal/install :: TestRegistryFutureSnapshotDeniesInstallThroughResolveRegistries
ok    internal/registry :: TestSnapshotFutureBoundIsExactAtEveryConfiguredSkew
ok    internal/registry :: TestSnapshotZeroClockSkewIsLiteral
```

`-run` anchoring was checked as the brief requires: every scoped invocation was anchored `^(...)$`
and its `=== RUN` lines counted (registry set: 16 = 4 top-level + 12 subtests; install set: 6).

Logs under `.temp/BUG-260906-1bdotx/`: `gate-*.log`, `mutant-{M1,M2,M3,baseline}.log`,
`stress-boundary.log`, `mutants.sh`.

---

## 6. Workspace discrepancy — needs the orchestrator's attention

The spawn placed this run in the **curator-spec** story worktree
(`curator-spec/.temp/STORY-260907-3hsndz/worktree`), whose boilerplate asks for uncommitted work
there. That repository contains no Go code — no `internal/install`, no `internal/registry` — so this
task cannot be performed in it.

The work is in the location the task description and the producer brief both name:
`/Users/iv/Developer/ReluxWorks/.worktrees/curator-snapshot-flake`, branch
`fix/snapshot-timestamp-flake`, base `919e2e9c`, three signed commits by `Ivan Oparin
<oparin@me.com>` (all `%G? = G`). **The story worktree is unchanged and the handoff snapshot against
its checkpoint will be empty.** Not pushed, no PR opened, per the brief.

## 7. Not done

- No hosted-runner dispatch. The orchestrator owns that, and the Windows confirmation this fix
  deserves is a `windows-latest` run of the two `test-gate` lanes on this head.
