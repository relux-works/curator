# TASK-260823-lk8hxy — reviewer verdict: ACCEPTED

Run `RUN-260823-2172f5` (not goal-bound). Reviewed head: `d76fe4d` (merge of PR 30,
branch `fix/TASK-260823-lk8hxy-unicode-vector-payload`, head `7762807`).
`git diff 7762807 d76fe4d` is empty, so the branch head and the merge tree are identical.

## Verdict

Accepted. The acceptance criterion — "Both invalid-unicode candidate cases pass on
windows; merged to main with green CI" — is met, verified from CI artifacts I pulled
myself rather than from the implementer's report.

## 1. Both candidate cases pass on Windows (independently verified)

Source: `candidate-evidence-windows-latest/test/go-test.json`, candidate lane run
`32644459040` job `97206457840`, on the exact merged tree `7762807`.

| Case | Baseline `95ca5ae` (run 32638424105) | Merged tree `7762807` |
| --- | --- | --- |
| `internal/buildsource TestBuildSourceIdentityVectors/invalid-unicode-build-source-path` | fail | **pass** |
| `internal/godriver TestToolchainIdentityVectors/invalid-unicode-toolchain-path` | fail | **pass** |

Ubuntu candidate lane: `invalid-unicode-build-source-path` pass, zero failing tests.
(`internal/godriver` is in `plan-excluded.txt` on ubuntu, so the toolchain vector is
Windows/macOS-covered only — pre-existing lane shape, not a regression.)

Full failing-test delta on the Windows candidate lane:

| Failing test | Baseline `95ca5ae` | Merged `7762807` | Owner |
| --- | :---: | :---: | --- |
| `buildsource .../invalid-unicode-build-source-path` | FAIL | pass | this task — fixed |
| `godriver .../invalid-unicode-toolchain-path` | FAIL | pass | this task — fixed |
| `godriver .../unsorted-directories-files-and-internal-link` | FAIL | FAIL | TASK-260823-czs1cx |
| `godriver TestFixedEnvironmentAndFiveDirectArgvFormsVector/fixed_environment` | FAIL | FAIL | TASK-260823-czs1cx |
| `install TestDryRunEffectBindingsSeeWhatARealOperationWrites` | FAIL | pass | fixed elsewhere |

No case regressed.

## 2. Merged to main with green CI

`main` push run `32646215503` on `d76fe4d`: **success**, every job green —
Lint, Naming gate, Interop conformance gate, Gate self-test ×3, Test ×3
(including `Test (windows-latest)`), Race ×2.

## 3. Root-cause reasoning re-derived, not taken on trust

The claim is that the guard was already fail-closed and the laundering was in the
probe's write path. Checked at the merged head:

- `identifiers.PortablePath` (`internal/identifiers/identifiers.go:64`) already calls
  `utf8.ValidString`, and `buildsource.(*pathSet).add` calls it. Invalid UTF-8 was
  never admissible.
- `godriver.validToolchainPath` calls `utf8.ValidString` directly.

So a name that genuinely reaches the guard as invalid UTF-8 is refused on every
platform, and no platform-specific product rule is needed. The premise holds.

## 4. PR 28 is fully reverted; product side is a net no-op

- `git diff f761e50^1:internal/buildsource/buildsource.go d76fe4d:internal/buildsource/buildsource.go` → empty
- `git diff f761e50^1:internal/godriver/fingerprint.go d76fe4d:internal/godriver/fingerprint.go` → empty
- `git ls-tree -r d76fe4d | grep fsunicode` → nothing; `git grep fsunicode d76fe4d` → nothing.
  Only `LOGBOOK.md` and board artifacts still name it, as history.
- `git diff --stat f761e50^1 d76fe4d -- internal/` touches test files only.

PR 28 added a Windows rule refusing every U+FFFD in the identity path on a premise the
implementer then disproved by measurement (unpaired surrogates round-trip; the raw
`0xFF` payload does not). Removing it is right: that rule would have rejected literal
U+FFFD names that no laundering could produce and diverged Windows admission from POSIX.

## 5. Test-harness change reviewed

`invalidUnicodeNames` / `writeInvalidUnicodeMember` / `hasInvalidUnicodeMember`, added
to both conformance test files. The probe now writes a spelling, reads the directory
back, and only asserts on the guard once the host actually presents a name that fails
`utf8.ValidString`; otherwise it removes the laundered member and tries the next
spelling, and skips as a host-capability limit when none survives. That is strictly
stronger than the old code, which asserted on the guard without ever checking that the
probe had created what it claimed to create — the exact hole that produced the original
false failure.

Checked and fine: the remove-on-laundering path resolves through the same UTF-16
conversion as the write, so `os.Remove` targets the U+FFFD member it just created;
the godriver call site scans `root/bin`, whose only other member `go` is valid, so a
true result cannot come from an unrelated entry.

The helper trio is duplicated verbatim across `internal/buildsource` and
`internal/godriver`. Not a finding: the repo has no shared test-helper package and
these files already duplicate `writeTestFile` / `decodeVectorBytes`-class helpers.

## 6. Local re-validation at the merged tree

Worktree `.temp/TASK-260823-lk8hxy/fixforward` @ `7762807`, darwin/arm64.

| Command | Exit |
| --- | ---: |
| `go build ./...` | 0 |
| `go vet ./...` | 0 |
| `GOOS=windows go vet ./...` | 0 |
| `go test ./internal/buildsource ./internal/godriver ./internal/identifiers -count=1` | 0 |

## 7. Corrections to the delivery record (non-blocking)

**7.1 — "every lane verified green pre-merge" is not accurate.** The board note says the
merge waited for every lane green. The `Candidate suite (windows-latest)` job on the
merged SHA (`32644459040` job `97206457840`) was **red**, and the evidence artifact does
not mention that lane at all. I confirmed the redness is pre-existing and unrelated:

- the platform-case gate failed on `internal/install/atomicity ::
  TestAdapterMirrorLinksAreJournaledAndRestoredExactly` — "required case never ran on
  windows". The same gate failure is present on `fbca886` (run `32642340559`) and
  `062d89b` before this change, where it failed on *two* atomicity cases. It looks
  order- or timing-dependent and is unowned by this task.
- the two remaining `godriver` failures are the TASK-260823-czs1cx pair.

What was actually verified green pre-merge was `Test (windows-latest)` and the whole
`pull_request` run `32644449835`. That is enough for this AC, but the note overstates it
and the next reader should not treat the candidate lane on `d76fe4d` as clean.

**7.2 — `Test (windows-latest)` does not exercise these vectors.** In that lane both
`TestBuildSourceIdentityVectors` and `TestToolchainIdentityVectors` skip wholesale
(`CURATOR_CONFORMANCE_ROOT` deferral); confirmed from
`test-evidence-windows-latest/test/go-test.json`, where both report `skip` with no
subtests. The candidate lane is the only Windows coverage for this AC. A green
`Test (windows-latest)` alone would not have proven the fix.

**7.3 — spec follow-up, not rework here.** On Windows the probe asserts the guard against
a *substitute* payload (unpaired surrogate `ED A0 80`) rather than the vector's published
`path_bytes_base64: "/w=="`. That is defensible: this is a reject-only case with no digest
dependency, the substitution is in the same equivalence class (not valid Unicode), the
code documents why, and the probe proves the host really presents an invalid name before
asserting. It is still a divergence from "materialize the exact published vector", which
the neighbouring symlink case enforces with a `t.Fatalf`. The durable fix belongs in the
conformance root: publish a platform-representable spelling, or mark platform
applicability, so each implementation harness does not have to invent its own substitute.
Worth raising against curator-spec; it does not block this task.

## Definition of Done

| Item | State |
| --- | --- |
| Merged to main with green CI; candidate case verified | yes — `d76fe4d`, run `32646215503` success; candidate cases verified from artifacts |
| Code written per task description and AC | yes |
| Relevant tests written for new/changed behavior and passing | yes — probe now proves its own precondition |
| Lint clean | yes — CI `Lint` success |
| Build/validation run after changes | yes — §6 |
| New outcome artifact attached with task-scoped name | yes — `TASK-260823-lk8hxy_windows-invalid-unicode-evidence.md`, plus this verdict |
| Findings recorded in logbook | yes — `LOGBOOK.md` records the measurement table, the PR 28 regression, and the PR 30 fix |
| Implementation matches AC | yes |
| Solution fits project architecture | yes — product side is a net no-op; the fix lives in the conformance harness, which is where the defect was |
| Tests green | yes, with the two czs1cx failures explicitly out of scope |
