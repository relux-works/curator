TASK-261008-1pg0pd (N9) — results
===============================

Finding (wave-2 report N9): a successful audit pin omits its required
creation time. The pinned manager source-audit contract requires a pin to
record content identity, operator identity, reason, and creation time;
Pin wrote no timestamp field.

Fix (production boundary: the audit pin writer)
------------------------------------------------
internal/audit/audit.go — PinAtVersion now stamps every pin record with
  "created_at": time.Now().UTC().Format(time.RFC3339)
for both the v1 (schema 1) and v2 (schema 2) carriers. The member name
and RFC3339-UTC-Z encoding mirror the existing created_at convention in
internal/audit/sourceaudit.go. Read-side (isPinned) intentionally
unchanged: legacy pins without the member keep authorizing, attested by
TestLegacySchemalessPinReadsAsV1. Doc comments updated; the v1 carrier
keeps schema_version 1 with no hash_version member.

Regression test (through the production entry)
----------------------------------------------
cmd/curator/audit_pin_creation_time_test.go —
TestAuditAllowRecordsPinCreationTime drives the real `run` CLI
dispatcher (`audit --allow <digest> --reason ...`, USER pinned to a
synthetic operator) under both the v1 and v2 hash writers, then asserts
the stored trust.json carries content_sha256, pinned=true, pinned_by,
reason, AND a created_at that is present, ends in Z, parses as RFC3339,
and falls inside the pin window. No "wave2" in the name.

Also touched:
- internal/audit/opaque_v1_pins_test.go: comment-only refresh (the v1
  carrier is no longer byte-identical after created_at; assertions
  unchanged).
- CHANGELOG.md: one Unreleased/Fixed line, operator-visible wording.

Compile-only tail (local, R223 — no `go test` on the mini)
----------------------------------------------------------
gofmt -l <3 touched go files>  -> clean, exit 0
go vet ./internal/audit/ ./cmd/curator/ -> exit 0
go build ./... -> exit 0

Hosted red/green proof (GitHub CI, full matrix)
-----------------------------------------------
GREEN (fix + test):
  https://github.com/relux-works/curator/actions/runs/37883305445
  conclusion: success. Lint, naming, interop, all gate self-tests, all
  Go driver lanes, Test (ubuntu/macos/windows), Race (ubuntu/macos):
  success. Candidate-suite and rose-air lanes: skipped by design.
  test-evidence-ubuntu-latest observed-cases.tsv:
    pass cmd/curator TestAuditAllowRecordsPinCreationTime (+v1/v2 subtests)
    pass TestAuditAllowWritesWriterVersionPinCarrier
    pass internal/audit TestPinAtVersionWritesVersionedCarrier,
      TestPinVersionMatchAuthorizes, TestV2RejectsLegacyV1Pin,
      TestV1RejectsV2Pin, TestLegacySchemalessPinReadsAsV1
  go-test.json: zero "fail" actions.

RED (same candidate, ONLY the two production fix lines reverted —
"time" import and the created_at member; test/changelog/comments kept):
  https://github.com/relux-works/curator/actions/runs/37886218996
  conclusion: failure, confined to the five Test/Race lanes; every other
  job success/skipped. Failing test (ubuntu AND macos evidence, the only
  failures in observed-cases.tsv):
    fail cmd/curator TestAuditAllowRecordsPinCreationTime (+v1/v2 subtests)
  Failure text: `pin created_at = <nil>, want a defined creation timestamp`.

Branch note: the hosted-evidence brief suggests scratch/<task-id>-* branches,
but ci.yml's push trigger covers only main and gate/** (plus PRs), so the
snapshots were pushed as gate/STORY-261009-3246nu/n9-green-* and
.../n9-red-* per scripts/remote-gate.sh mechanics. Both branches deleted
after recording; run URLs above remain.

Probe note: the N9-probe_test.go attachment to this task contains
TestWave2AuditDecisionTable — the report's 20/20-green decision-precedence
table (negative controls, command 13), not the expected-red N9 probe
(TestWave2AuditPinCreationTime, CLI-driven per the report's N9 table). The
permanent regression above was therefore written fresh through the same
production entry (real CLI dispatcher) asserting the N9 invariant plus
timestamp validity; the decision-table controls are covered by the
existing pin/gate tests, all green in the hosted run.

Worktree left UNCOMMITTED with exactly:
  M CHANGELOG.md
  M internal/audit/audit.go
  M internal/audit/opaque_v1_pins_test.go
  ?? cmd/curator/audit_pin_creation_time_test.go
