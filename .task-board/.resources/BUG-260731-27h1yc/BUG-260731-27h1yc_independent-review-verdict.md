# BUG-260731-27h1yc — independent review verdict: ACCEPTED

Reviewer run `RUN-260731-a902cb` (Claude Opus 5), 2026-07-31. Read-only; no code
modified. Every claim below was re-derived from primary sources (CI artifacts,
`git`, the GitHub API, local execution) rather than taken from the implementer's
notes.

Subject: PR https://github.com/relux-works/curator/pull/12,
`task/BUG-260731-27h1yc-windows-lane` → `main`, head `c7bc890`.

---

## 1. Acceptance criterion

> Curator Test (windows-latest) reports no required-case failures for
> internal/buildsource, internal/install or internal/install/atomicity, with the
> platform-case ledger unchanged or strengthened rather than relaxed.

**Met.**

### 1.1 No failures in the three packages, independently parsed

Both runs' `test-evidence-windows-latest` artifacts downloaded and parsed here
(unique top-level cases, deduplicated across `go-test.json`,
`go-test-deferred.json` and `go-test-served.json`):

| package | baseline `main` @ `3a047d5`, run `30624569953` | PR head `c7bc890`, run `30626331508` |
|---|---|---|
| `internal/install` | 60 failing | **0 failing** (100 pass, 7 skip) |
| `internal/install/atomicity` | 8 failing | **0 failing** (8 pass) |
| `internal/buildsource` | 2 failing | **0 failing** (7 pass, 2 skip) |
| repo-wide total | **91** | **19** |

The packages ran — they are not absent-by-omission: pass counts are recorded
above, and `internal/identity` executes 6 passing cases on the Windows runner too.

### 1.2 Every required ledger row reports `ok`

All 8 non-wildcard `platform-cases.tsv` rows for these packages report `ok` in
the gate's own report on the Windows runner. The 7 remaining
`FAIL required case failed` lines are all `cmd/curator`, owned by
BUG-260731-33v6zz.

### 1.3 The ledger is unchanged, verified by diff rather than by claim

- `git diff origin/main...c7bc890 -- .github/ci/platform-cases.tsv .github/ci/skip-classes.tsv` → **empty**. Both files byte-identical to `main`.
- `test/skips-observed.tsv` on the Windows runner vs the `main` baseline: 64 rows
  each, one textual difference, a `t.TempDir` nonce inside an unrelated
  `internal/scopes` reason string. **No skip added.**
- No test file deleted (`--diff-filter=D` empty); no `t.Skip`, no
  `runtime.GOOS`-based exclusion added anywhere in the diff.

### 1.4 Strengthened, not relaxed

Two `runtime.GOOS != "windows"` guards were **removed**, so `TestGlobalInstall`
and the global half of
`TestFailureAtEveryTargetClassRestoresPriorStateInReverseOrder` now assert the
PATH-visible forwarding shim and the user-bin mirror ledger on Windows.
`TestAdapterMirrorLinksAreJournaledAndRestoredExactly` reports `ok` rather than
`tol` on Windows, so its host-capability tolerance was not consumed.

### 1.5 Attached artifacts are authentic

`BUG-260731-27h1yc_windows-platform-case-gate.txt` and
`BUG-260731-27h1yc_windows-skips-observed.tsv` are **byte-identical** to
`test/platform-cases.txt` and `test/skips-observed.tsv` inside the CI artifact
for run `30626331508`. Nothing was hand-authored.

## 2. Definition of Done

| item | verdict | evidence |
|---|---|---|
| classify the masked failures | done | root causes A–E in the results artifact match what the evidence shows |
| fix without skips or ledger weakening | done | §1.3, §1.4 |
| focused Windows regression tests, macOS/Linux preserved | done | §3; `Test`/`Race` on ubuntu and macos green on `c7bc890` |
| signed PR to `main` with native Windows CI evidence | done | `2a02da9`, `a164dca`, `c7bc890` all `verification.verified=true, reason=valid` via the GitHub API; base `main`; merge-base == `origin/main`; `mergeable` |
| lint clean | done | `Lint` job pass; locally `gofmt -l internal cmd` clean, `go vet` and `GOOS=windows go vet` clean on the touched packages |
| tests green | done | §4 |
| outcome artifacts attached | done | §1.5 |

## 3. Implementation review

**Deriving the platform from the host is correct, not a weakened expectation.**
Production never passes a non-host platform: `install.Options.Platform` is
documented `"" resolves to the current platform`, and both `projectAttempt`
(`install.go:191`) and `globalAttempt` (`global.go:63`) fall through to
`runtimestore.Platform()`. No production caller sets it. The fixtures pinning
`"unix"` were therefore asserting a shape production never builds on that host —
`installPlatform()` restores fidelity and removes no legitimate coverage.

**The Windows frozen-root fixture genuinely asserts the property.** `Recheck`
(`buildsource.go:153`) detects replacement via `os.Lstat(token.path)` +
`os.SameFile`, and `scan` walks the *already-open* `*os.Root`. Both instances
carry byte-identical content, so if the reparse-point repoint were not observed,
`scan` would still see instance-1, identity and state would match, and `Recheck`
would return `nil` — failing the case. It passes on Windows, so the
root-replacement detection genuinely fired. The fixture `t.Fatal`s rather than
skips when no reparse point can be created, which is right for a row whose
`skip_allowed_on` is `-`.

**`internal/identity` opens no new bypass class.** The new `driveRE` is
`^[A-Za-z]:[\\/]` — strictly *narrower* than the single-letter-host rule already
in `Parse` (`identity.go:83`), which treated `C:` as a drive for any following
path. A real hostname before a colon still reaches `scpRE` and stays subject to
the allowlist; the new `Parse`-level test pins that half. Local sources were
already unrestricted by `Allowed`, so nothing that was gated becomes ungated.
The change is out of the named package set, but justified: two in-scope
`internal/install` dry-run cases fail *inside* it, and the defect is real —
`curator` could not install from a local path on Windows at all, which
`TestCanonical` masked because `Canonical` discards `Parse`'s error.

**`shimName` duplicating the `.cmd` rule** that `runtimestore` already encodes is
acceptable here: the task forbids modifying `internal/runtimestore`
(BUG-260731-11bpa4), so an exported helper was not available. Worth folding into
`runtimestore` once that bug closes.

## 4. Local re-execution (darwin/arm64, go1.25.5)

`internal/install`, `internal/install/atomicity`, `internal/buildsource`,
`internal/identity`, `internal/runtimestore`, `internal/globalbins` — all `ok` at
`c7bc890`. `go vet` and `GOOS=windows go vet` clean on the touched packages;
`gofmt -l internal cmd` clean.

## 5. Non-blocking observations for the coordinator

1. **Pre-existing flake, not caused by this PR.** Under heavy local parallel load
   `internal/install TestRegistryRevocationDeniesInstall` failed once with
   `registry test-reg snapshot timestamp is too far in the future`. Mechanism:
   `registry.checkSnapshotsWithPolicy` (`snapshot.go:75`) defaults `maxAge` when
   zero but **not** `clockSkew`; the install test env builds `&config.Config{}`
   as a literal, so `Audit.SnapshotClockSkewSeconds` is 0 and `clockSkew` is 0,
   while `fakeRegistry` serves `created_at` as second-precision `time.Now()` at
   serve time — strictly after the `now` captured at `install.go:1134`. Any
   second-boundary crossing between capture and fetch trips
   `parsed.CreatedAt.After(now)`. It did not reproduce on the PR head run alone,
   nor on `main` under the same parallel command, and this diff touches none of
   `registry_e2e_test.go`, `install.go`, `snapshot.go` or `config.go`. Worth its
   own board item, against `internal/registry`.
2. **`TestReadDocumentBindsGenerationToBytesReplacedByRename` is weaker on
   Windows by construction.** Since the kernel refuses the replacement during the
   read window, the "read stayed on the old inode" assertion is trivially true
   there. The other two properties — the generation belongs to the returned
   bytes, and a later path read differs so the run restarts — are still genuinely
   asserted, and the code comments state the limitation honestly. Not a required
   ledger row.
3. **The authoritative suite did not cross-check the identity change.**
   `internal/buildsource TestBuildSourceIdentityVectors` and
   `TestBuildSourceConformanceVectors` skip as `root-unset`
   (`CURATOR_CONFORMANCE_ROOT is not set`) on *every* lane including ubuntu, so
   the identity fix rests on this repo's own tests. Pre-existing gate gap, not
   introduced here.
4. **No CHANGELOG entry** for a user-facing fix (`## Unreleased` is empty).
   PRs 9, 10 and 11 also skipped it and no CI gate enforces it, so this is a note
   for the release owner rather than a defect.
5. **BUG-260731-33v6zz's description is wrong about the ledger.** It states "None
   of the 14 appears in `.github/ci/platform-cases.tsv`". `main` carries 7
   `cmd/curator` rows (lines 149–155) and the Windows gate reports all 7 as
   required-case failures. That owner's scope should be corrected before it
   starts.

## 6. Verdict

**Accepted.** The acceptance criterion is met on a native `windows-latest`
runner, the ledger is byte-identical to `main` and materially strengthened in
behaviour, macOS and Linux lanes are green, and every artifact on the board
matches its CI source exactly. `Test (windows-latest)` remains red only on
`cmd/curator`, which the ownership map assigns to BUG-260731-33v6zz.

Acceptance evidence recorded for the commit-owning mover; no `commit_ack`
supplied by this reviewer run.
