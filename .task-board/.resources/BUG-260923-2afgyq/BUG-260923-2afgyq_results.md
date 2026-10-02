# BUG-260923-2afgyq — marker-reader-cross-field-validation

Ready for review. Changes remain uncommitted in the assigned Story worktree.

## Change

Moved the existing core-v5 external cross-field binding into `validV3Build`,
used by marker v3/v4 and core v5. Unsubstituted builds must keep the declared
identity kind, identity value, object format, and commit. Local substitutions
require the local identity kind and no ref; network substitutions require the
network identity kind and a typed ref whose revision width matches SHA-1 or
SHA-256. The separately frozen package-v5 path is unchanged. No released schema
changes. One CHANGELOG line under Unreleased. No LOGBOOK, per the task brief.

Production call chain: `Read` -> `ReadState` -> `validMarker` -> `validBuildState`
-> `validV3Build`. Tests write foreign-writer bytes directly, bypassing `Write`,
and assert both `Read` refusal and `ReadState` present/invalid behavior.

## Exact cases and ledger

All five published invalid cases are rejected in v4, and their v3 counterparts
are rejected, for all three pinned suites (30/30 direct published regressions).

- `invalid-external-declared-effective-mismatch.json` (the published fixture
  changes the effective commit, not the identity text).
- `invalid-marker-local-identity-kind-mismatch.json`.
- `invalid-marker-network-identity-kind-mismatch.json`.
- `invalid-marker-sha1-effective-revision-width.json`.
- `invalid-marker-sha256-effective-revision-width.json`.

The ledger has 20 physical rows for these five distinct cases: five v4 rows
in each of rc.13, content-hash-v2, and Muse, plus five rc.13 v3 counterparts.
Exactly those 20 owned rows were removed after their exact published cases
passed; no unrelated rows or case-count pins changed. The removed rows are
included verbatim in `removed-ledger-rows.tsv`.

| Suite | v2 driven/total | v3 driven/total | v4 driven/total | v5 driven/total | Accepted valid fixtures |
| --- | --- | --- | --- | --- | --- |
| rc.13 | 14/14 | 27/27 | 27/27 | not published | 26/26 across v1–v4 |
| content-hash-v2 | 14/14 | 28/28 | 28/28 | 29/29 | 37/37 across v1–v5 |
| Muse | 14/14 | 28/28 | 28/28 | existing 29 known-gap classification | 37/37 across v1–v5 |

No gaps, bounds, or skips remain in the tested v2/v3/v4 schema families.
Valid-only fixture tests explicitly cover the frozen v1 family (2 published
cases, 1 valid) and every published later marker version. Total valid fixture
acceptance is 100/100 across the three suites. Muse's existing v5 coverage
classification is owned by a different task; this change proves its valid
fixtures but does not claim to retire that full family's historical rows.

## Validation run directly

Every command was a standalone process; log redirection preserved its real
exit code. No gate was piped through `tee`. All verification was run in this
session; no previously attached validation was accepted as a substitute.
Exact commands, log names, and exit codes for initial, intermediate, mutant,
and review-candidate runs are in `commands.json`.

- `go test ./internal/marker ./internal/conformancecoverage -count=1 -v`, with
  each of the three pinned `CURATOR_CONFORMANCE_ROOT` values: exit **0**, **0**,
  **0** on the review candidate. Logs: `verified-rc13-v1.log`,
  `verified-hashv2-v1.log`, `verified-muse-v1.log`.
- Relevant `./cmd/curator` build classification, marker refusal, schema-band
  status, concurrent-state status, and repair notice rows: exit **0**.
  The exact anchored `-run` mask is in `commands.json`; log
  `cmd-curator-review.log`. The earlier broader status mask also exited **0**.
- `golangci-lint run ./internal/marker/... ./internal/conformancecoverage/...`:
  exit **0**, **0 issues** (`lint-review.log`).
- `go vet ./internal/marker ./internal/conformancecoverage ./cmd/curator`:
  exit **0** (`vet-review.log`).
- `go build -o /tmp/BUG-260923-2afgyq-evidence/curator ./cmd/curator`:
  exit **0** (`build-review.log`).
- `gofmt -w internal/marker/marker_external_cross_fields_test.go` and
  `git diff --check`: exit **0** each.

Full repository/platform/race gates were not run; validation was scoped to the
reader, its conformance ledger, and relevant CLI consumers. Host syspolicyd
checks reported `state = running`, `successive crashes = 349`; no down-state
gate was started.

## Negative evidence

The original production source plus the new regressions exited **1**
(`baseline.log`): every named v3/v4 invalid fixture was admitted. This is a
failing expected-red run, not a passing gate.

Five per-check mutants each exited **1** and failed their exact matching v4
published case: declared binding, local kind, network kind, SHA-1 width, and
SHA-256 width. Mutant kill ratio: **5/5**. Three further narrowed binding
mutants retained the other checks while dropping only identity value, identity
kind, or commit equality; each exited **1** in the matching v3/v4 negative
tests. Narrowed binding mutant kill ratio: **3/3**. Logs and the preparation
script are attached. The production source was restored byte-for-byte before
review-candidate verification.

An intermediate attempt to widen the existing Muse v5 conformance test exited
**1** because its existing historical known-gap rows become stale under direct
reader classification. That harness widening was reverted; valid-only tests
were added instead, and the scoped suites subsequently exited **0**. No Muse
v5 ledger rows were removed.

Toolchain versions, exact suite revisions/manifest digests, source base, and
candidate file SHA-256 digests are in `metadata.json`. The evidence archive
contains logs and references, not duplicate corpora or the built binary.

## Handoff retry

The first `task-board handoff BUG-260923-2afgyq --role developer` exited **1**:
the generic logbook checklist item was unchecked. The current task brief
explicitly says **"No LOGBOOK."** The obsolete generic item was replaced
through the board CLI with the task-specific requirement to record findings in
the attached outcomes and honor the no-LOGBOOK instruction. No logbook claim
was checked as satisfied.

The first attempt was also delayed by the host and board snapshot processing.
Read-only service checks subsequently exited **0**, reporting syspolicyd
running and successive crashes 351. A read-only `sample` command exited **0**
and showed the handoff inside `writeboundary.TakeSnapshot`, scanning files.
The captured diagnostic is attached in the evidence archive. Initial host
execution attribution was corrected after inspecting that sample. The code
and validation candidate have not changed during the handoff retry.

## Revision 2 (refresh)

Base refresh only, from `2cb29dac8a4c82c5a07d7ca2d107aa6e6e2c93e7`
to trunk `f40b77c19c01746bda8b9a610358d860f2ad20c5`. No new marker
implementation or test changes. The actual revision 1 verdict is **ACCEPTED**;
the generated review-round text describing a rejection does not match that
artifact. No rejection finding requires a new regression or mutant here.

`task-board worktree refresh-candidate BUG-260923-2afgyq` exited **0**, outcome
`refresh_advanced`. It preserves candidate files exactly, so the incoming source
was then combined explicitly. Verified the preserved files equaled revision 1
before combining; restored only unrelated incoming source paths to trunk, removed
only the exact revision 1 marker rows from trunk's ledger, and retained both
Unreleased CHANGELOG lines. A second refresh exited **0**, outcome
`refresh_already_current`. No replay conflict or hand commit. No board files were
edited directly; no LOGBOOK changes.

Proof against accepted revision 1 tree
`4eded38359c54177b468b13d7c4934b4c52e4f7b`:
- Marker source and regression tests are byte-identical to revision 1.
- All unrelated incoming source paths are byte-identical to new trunk.
- Ledger: old base **138**, trunk **136**, combined candidate **116** rows.
  Trunk's **2** hardlink rows remain removed. This candidate removes the same
  **20 physical rows / 5 distinct case names** as accepted revision 1: hashv2
  v4 **5**, rc13 v3 **5**, rc13 v4 **5**, Muse v4 **5**. No unrelated removals,
  additions, or case-count pin changes.
- CHANGELOG equals trunk plus the single unchanged marker line.
- Candidate source digests were checked again after validation; no drift.

Directly rerun for this refresh, standalone commands with real exit codes:
- `go test ./internal/marker ./internal/conformancecoverage -count=1`
  with rc13, hashv2, and Muse roots: **0 / 0 / 0**. These run the named
  production-reader regressions and every published valid marker version.
- Relevant cmd/curator rows, same anchored mask as revision 1: **0**.
- `go build -o /tmp/BUG-260923-2afgyq-refresh/curator ./cmd/curator`: **0**.
- `golangci-lint run ./internal/marker/... ./internal/conformancecoverage/...`:
  **0**, 0 issues.
- `go vet ./internal/marker ./internal/conformancecoverage ./cmd/curator`: **0**.
- `git diff --check -- . ':!.task-board'`: **0**.

The revision 1 producer/reviewer mutant evidence and focused install evidence
are accepted as existing evidence, not claimed as rerun in this refresh.
No full install-package, full repository, race, or platform suite was rerun:
the current brief requires the marker/coverage refresh gate, and marker code
and tests have not changed. Existing full-install green remains UNKNOWN per
the accepted review's recorded timeouts. No new gate failure in this refresh.
Before gates, syspolicyd checks exited **0**, reporting running and successive
crashes **355**. No gate was started while that service was down.

New outcome `BUG-260923-2afgyq_refresh-evidence.tar.gz` contains the exact commands,
raw logs, file SHA-256 identities, and removed marker rows. The executable is
excluded. Ready for review as revision 2; normal independent review is required.

### Revision 2 handoff retry

The first refresh handoff exited **1**: unchecked item **12**, the generic
LOGBOOK requirement, had been added alongside the already-checked task-specific
item **7** (findings in task-scoped outcomes; no LOGBOOK). The current brief
explicitly forbids LOGBOOK. Inspected the live checklist and removed only the
obsolete conflicting item through the CLI, retaining checked item 7. No
LOGBOOK evidence was invented and no repository logbook was edited.
The warning about comparing board content recorded no command progress;
the source gates and attached resources remain independently evidenced.
Read-only handoff process sampling exited **0**. This retry changes only
board checklist/evidence, not the proved repository source candidate.
