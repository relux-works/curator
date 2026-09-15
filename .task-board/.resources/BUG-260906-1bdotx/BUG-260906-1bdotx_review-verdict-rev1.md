# BUG-260906-1bdotx — review verdict, Change Request rev 1

**ACCEPT.** Reviewed head `879b884b`, branch `fix/snapshot-timestamp-flake`, base `919e2e9c`.
Every claim below was reproduced by the reviewer in an independent throwaway copy
(`/tmp/BUG-260906-1bdotx-review`, tar of the producer head, submodules excluded). The producer's
worktree and the story worktree were never written to; both verified clean at the same heads
afterwards.

---

## 0. Why `repository_delta=empty`, and why that is not the failure here

The Change Request snapshotted `curator-spec/.temp/STORY-260907-3hsndz/worktree`. **That is a
different repository.** It has no `internal/install`, no `internal/registry`, no `internal/config` —
`find . -name '*.go'` returns a vector generator and conformance fixtures, and `git remote -v` is
`relux-works/curator-spec`. The task's own Scope field names the `curator` repository and the
worktree `/Users/iv/Developer/ReluxWorks/.worktrees/curator-snapshot-flake`, and the producer brief
names the same. The work could not have been performed in the tree that was snapshotted.

The work exists, in the location both the task description and the brief name: three commits, all
`%G? = G`, author `Ivan Oparin <oparin@me.com>`, tree clean —

```
879b884b G  Make the windows lane run the snapshot clock bound by name
e8b7129c G  Pin the future-timestamp bound at every skew a config can carry
a8a64197 G  Stop the registry fixtures minting a timestamp during the fetch
```

`git diff 919e2e9c..879b884b --stat` → 4 files, +200/−5. The producer flagged the discrepancy itself
in §6 of its report rather than letting an empty snapshot pass silently.

So `repository_delta=empty` here is a **workspace-routing artifact, not an absence of work**.
Accepting is the right call: the delivered change is correct, complete and green on hosted CI, and
routing it to `to-dev` would ask a producer to redo correct work in a repository where it cannot be
done. **The routing itself needs the orchestrator's attention** — the CR snapshot does not cover the
accepted artefact, so integration must take `879b884b` from the `curator` branch, not from this
candidate tree.

---

## 1. The mechanism — independently confirmed, and reproduced deterministically

Both composing facts verified **by construction**, not by reading.

**Fact 1 — the e2e config carries a literal zero skew.** Driving `newEnv` and `newEnvIn` and printing
what they hold:

```
PROBE newEnv:   SnapshotClockSkewSeconds=0 SnapshotMaxAgeSeconds=0 RegistryPolicy=""
PROBE newEnvIn: SnapshotClockSkewSeconds=0 SnapshotMaxAgeSeconds=0 RegistryPolicy=""
PROBE production default via parseAudit: DefaultSnapshotClockSkewSeconds=300
```

Confirmed structurally too: `config.go:709` (`parseAudit`) is the **only** production writer of the
default — `grep -rn SnapshotClockSkewSeconds --include=*.go` outside tests returns exactly
`install.go:1194`, `install.go:1200` (readers), `config.go:41` (the constant), `config.go:126` (the
field), `config.go:709` (`parseAudit`), `config.go:784` (the JSON binding). And in
`checkSnapshotsWithPolicy` **only `maxAge` has a zero-fallback**; `clockSkew` has none. In this path
the gate is literally `CreatedAt.After(now)`.

**Fact 2 — `now` is sampled before the fetch.** `time.Now()` is an *argument expression* at
`install.go:1191-1192`, so Go evaluates it before the callee runs, and the callee is what invokes
`registry.HTTPGetSnapshot`. The ordering is a language guarantee, not an empirical hope. A probe
confirmed the handler stamps afterwards (`delta=338ms` on a warm run).

**The reproduction.** Reverting *only* the fixture to its pre-fix per-request `created_at` — the gate
untouched — and forcing the boundary crossing:

```
--- FAIL: TestRegistrySnapshotSurvivesASecondBoundaryDuringFetch (0.69s)
  Messages: [... test: registry: registry test-reg snapshot timestamp is too far in the future]
  Errors:   [every trusted audit registry served a tampered snapshot]
exit 1, === RUN lines: 1
```

Both reported strings, verbatim, **on darwin/arm64**. That settles the last open question in the
brief: **the defect is not Windows-specific.** Windows was simply the only lane slow enough for the
fetch window to cross a whole-second boundary on its own. The producer says exactly this and does not
overclaim it.

---

## 2. The gate is intact — checked adversarially, not read

- `git diff 919e2e9c..HEAD -- internal/registry/snapshot.go internal/config/config.go
  internal/install/install.go` → **0 lines**. No production file changed at all; the 4 changed paths
  are two test files, one test-support file and `.github/ci/platform-cases.tsv`.
- `clockSkew` is not widened anywhere; no assertion removed; no tolerance introduced.
- **No bypass path.** All three exported entry points (`CheckSnapshots`, `CheckSnapshotsWithPolicy`,
  `CheckSnapshotsWithPolicyReadOnly`) funnel into `checkSnapshotsWithPolicy`, and the future gate sits
  **before** the `if persist` branch — so the persisting *and* the read-only install paths both carry
  it. `install.go:1191` and `install.go:1197` are the only production callers.
- No test tolerates a genuinely future-dated snapshot; the one new test that accepts
  (`…AtTheSkewBoundIsAcceptedThroughInstall`) uses a **past**-dated snapshot.

### Mutants — reproduced by the reviewer, not taken on report

| Mutant | Gate becomes | At BASE | At HEAD | Named tests that fail at head |
| --- | --- | ---: | ---: | --- |
| **M1 narrowing** | `After(now.Add(clockSkew + time.Second))` | exit 1 | **exit 1** | `TestSnapshotFutureBoundIsExactAtEveryConfiguredSkew/{0s,30s,5m0s}/one_second_past_the_bound`, `TestSnapshotZeroClockSkewIsLiteral`, `TestSnapshotRequiresCompleteShapeAndRejectsEquivocation` |
| **M2 narrowing** | `After(now.Add(max(clockSkew, time.Second)))` | **exit 0 — SURVIVED** | **exit 1** | `TestSnapshotFutureBoundIsExactAtEveryConfiguredSkew/0s/one_second_past_the_bound`, `TestSnapshotZeroClockSkewIsLiteral` |
| **M3 delete** (`false &&`) | assertion gone | install **exit 0 — SURVIVED** | registry exit 1, **install exit 1** | + `TestRegistryFutureSnapshotDeniesInstallThroughResolveRegistries` |
| **M0 fix-revert** | gate untouched, old fixture | — | exit 1 | `TestRegistrySnapshotSurvivesASecondBoundaryDuringFetch`, exact reported text |

Two rows carry the weight, and both are **measured coverage gains, not restatements**:

- **M2 survived the entire pre-existing suite and now dies by name.** A one-second floor under the
  configured skew is precisely the careless fix this ticket invites; nothing in the tree caught it
  before. The producer called M2 "the mutant that matters most here" — that is correct.
- **M3 left `internal/install` completely green at base** (`exit 0`, full package). There was **zero**
  production-path coverage of this gate. `TestRegistryFutureSnapshotDeniesInstallThroughResolveRegistries`
  now carries it through `Project → resolveRegistries → registry.CheckSnapshotsWithPolicy`.

M1 was already killed at base inside `internal/registry` by the pre-existing
`TestSnapshotRequiresCompleteShapeAndRejectsEquivocation`. The report lists that test among M1's
killers, so it is not passing pre-existing coverage off as new; its "before this change, neither
mutant was visible to `internal/install`" is scoped to install and is exactly right (M1 install:
exit 0 at head, as reported).

DoD item 12 (token-preserving mutant) genuinely does not apply: this gate compares two `time.Time`
values and reads no source text.

---

## 3. Windows — measured, on a hosted runner, at the exact accepted head

The producer correctly claimed no Windows result (it had no dispatch). The measurement now exists and
is recorded here.

Run **34130394522**, `event=pull_request`, `headSha=879b884b…` — **the exact head accepted**. All 12
checks pass; `Test (windows-latest)` pass in 43m24s. The Windows job log shows the platform-case gate
reporting all four ledger cases:

```
ok  internal/install  :: TestRegistrySnapshotSurvivesASecondBoundaryDuringFetch
ok  internal/install  :: TestRegistryFutureSnapshotDeniesInstallThroughResolveRegistries
ok  internal/registry :: TestSnapshotFutureBoundIsExactAtEveryConfiguredSkew
ok  internal/registry :: TestSnapshotZeroClockSkewIsLiteral
```

That `ok` is **execution evidence, not declaration**: `platform-case-gate.sh:270` prints it only on
`seen["pass\t"key]` from the `go test -json` stream, and :285 has a distinct `FAIL required case never
ran` branch. So the Windows lane genuinely ran the bound by name.

---

## 4. Blast radius — enumerated

Eight tests consume the touched fixtures: `TestRegistryRevocationDeniesInstall`,
`TestRegistryAttestationLandsInMarker`, `TestStrictRegistryPolicyFailsUnknown`,
`TestGlobalMarkersCarryMcpAndAttestationEvidence` (all via `registryEnv`/`fakeRegistry`), the three
new ones, and `TestGoToolchainRemedyReachesTheOperatorIntact` (via `loopbackRegistry`). None asserts
on a timestamp. Freezing `created_at` cannot trip the neighbouring checks: equivocation keys on
version/head/merkle-root/log-size, all constant in these fixtures, and staleness has 7 days of room
via the `maxAge` zero-fallback. All eight are green in the full-package `-race` run below.

---

## 5. Gates re-run by the reviewer — standalone processes, observed exit codes

| Gate | Exit | Notes |
| --- | ---: | --- |
| `go build ./...` | 0 | |
| `go vet ./...` | 0 | |
| `gofmt -l cmd internal` | 0 | no output |
| `bash .github/ci/gate-selftest.sh` | 0 | `130 passed, 0 failed` |
| `go test -count=1 -race ./internal/install/ ./internal/registry/` | 0 | 70.1 s / 3.2 s |
| `go test -count=25 -race -run '^TestRegistrySnapshotSurvives…$'` | 0 | 25 `=== RUN`, 25 PASS, 0 FAIL |
| `go test -count=1 -run '^(TestSnapshotFutureBound…\|TestSnapshotZeroClockSkewIsLiteral)$' ./internal/registry/` | 0 | 14 `=== RUN` (2 top-level + 12 subtests), 14 PASS |
| `gh pr checks 64` | — | 12/12 pass at `879b884b`, incl. `Test (windows-latest)` |

Every `-run` anchored `^(...)$` at a single level and its `=== RUN` lines counted — the discipline
paid off once: the first M0 attempt showed `RUN_LINES=0` on a build failure that would otherwise have
read as a pass.

`ledger-consistency.sh` was **not** re-run by me (it requires an evidence dir argument; exit 2 on
usage). It is not needed for the Windows question — it proves cross-`GOOS` *compilation*, and I have
the stronger thing: hosted *execution*.

---

## 6. Findings — all minor, none blocking

- **F1 (accuracy, prose overclaims a measured set).** The ledger row and the test's own doc comment
  say the bound is pinned "at every skew a manager config can carry". The test covers **three**
  skews — `{0, 30s, 300s}`; a config accepts any integer up to `maximumDurationSeconds`. The sample is
  a fair one (the gate is linear in `clockSkew`, and both skews that matter — the literal `0` the e2e
  path uses and the shipped default — are covered), but the prose asserts universality over an
  unbounded set from three points. The honest phrasing is "at the zero the e2e path uses, an
  arbitrary middle, and the shipped default". Same wording recurs in §3/§4 of the drafting report.
- **F2 (naming).** `TestRegistrySnapshotAtTheSkewBoundIsAcceptedThroughInstall` is not at the skew
  bound: it uses `-time.Minute` under skew `0`, i.e. an ordinarily past-dated snapshot. It is a valid
  positive control — it stops the refusal being satisfied by a reject-everything gate — but the name
  describes an edge it does not exercise. It is also the one new test absent from the ledger, so
  Windows does not run it by name (it does run in-package).
- **F3 (report staleness, not a defect).** §6/§7 say "not pushed, no PR" and "no hosted-runner
  dispatch". Both were true when written; by review time the orchestrator had pushed and opened
  PR #64 at the exact head. §3 of this verdict supplies the measurement the report correctly declined
  to claim.

None of these touches the gate, the fix, or the evidence. F1 and F2 are wording; fold them into a
later pass rather than spending a rework cycle.

---

## 7. AC — 4 of 4 rows driven

| AC row | Verdict | Reviewer evidence |
| --- | --- | --- |
| mechanism established from evidence and stated | **met** | Both facts re-derived by construction; flake reproduced deterministically on darwin with the exact two strings, gate untouched (§1) |
| gate still refuses a future-dated snapshot; narrowing mutant fails a named test; test exists after the change | **met** | M1/M2 reproduced by the reviewer; **M2 survived at base and dies at head**; M3 left install green at base and now dies through `resolveRegistries` (§2) |
| skew not widened, assertion not removed | **met** | `snapshot.go` 0-line diff; no production file changed; no bypass path across either install branch (§2) |
| no unmeasured Windows result claimed | **met** | Producer claimed none. Reviewer supplies a real hosted measurement at the exact head, with execution — not declaration — evidence (§3) |

**Accepted at `879b884b`. Windows was green on the exact head accepted.** Integration must take that
commit from the `curator` branch `fix/snapshot-timestamp-flake` — the CR candidate tree does not
contain it (§0).
