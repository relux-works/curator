# TASK-260910-3ungjy — independent review of CR revision 4

Verdict: **accepted**. R1–R3 closed; no new blocking findings. Route through accept_cr(revision=4) to integrating; reviewer does not mark done.

Candidate: `6987eaa1d028317b2d27e51b2b5fb00561a3fbf7`, base `0f0ae61766026cf2389ec4b91c09134cdb8aac62`, first-leaf checkpoint `dc5675e`. Patch SHA-256 independently verified: `99a330e632d54531d9d5e6beefb777e3f4f640ae07f4c2434342db1b310bc703`.

Reviewed from a frozen git archive under `/tmp/curator-review4.ibq8gp`; mutations only in `/tmp/curator-review4.ibq8gp-mutant`. No candidate, control-root, user configuration, or LOGBOOK.md writes. The 12 leaf paths in the live workspace matched the candidate blobs byte-for-byte. The 20-path CR includes the checkpointed sibling; leaf scope is the 12-path delta from dc5675e.

| Item | Evidence and assessment |
|---|---|
| §8.3 approve | `cmd/curator/hook.go:53-90`: shared Canonicalize, current bytes read before publication, distinct absent/unreadable errors, operator record, digest-sensitive re-record, unchanged matching operator record. Atomic Upsert helper reused. Command tests cover absent, unreadable, symlink alias, idempotence and re-record. |
| §8.3 approvals | `hook.go:97-116`, `internal/hookapproval/hookapproval.go:391-420`: tolerant read-only Scan, stable path sort, malformed lines reported/skipped without repair. |
| §8.3 revoke | `hook.go:123-145`, existing `hookapproval.go:350-377`: reuse Revoke, no-record returns before publication and reports nothing to revoke, exit 0. Committed command tests assert byte-identical absent-record state. |
| R1 closed | `hookapproval.go:496-508`: associate record before stat; recorded missing row retains approved_by with file=missing, no fabricated digest comparison. `NonCurrent` at :455 and `hook.go:176` fail --check. Production tests `hook_posture_test.go:92` and :353 cover both status commands, text/JSON, normal/--check. |
| R2 closed | `hookapproval.go:510-542`: stat/read failures retain recorder plus unreadable qualifier. Both production commands tested with deterministic directory-as-file fixture and an otherwise-current env matrix; approvals read failures named at `hook_posture_test.go:513`. |
| R3 closed | `hookapproval.go:567-580`, `main.go:781-826`, `envprofile/status.go:129-132,217-225`: unavailable approval state explicitly reported in JSON warnings and fails --check; absent state remains non-failing. Both command surfaces have paired controls. |
| JSON shape | `cmd/curator/status_test.go:1745-1776`: exactly four top-level keys, no builds, two approved/manager rows with exactly path/state/approved_by. Conditional warnings key documented. |
| Windows repair | Independently compared rev3/rev4 patch resources: only hook_posture_test.go changes, adding JSON decoding helper and replacing three raw-path assertions. No folding or GOOS skip added. Directory at state-file path proves unreadability portably. |
| Scope/docs/profile | `docs/cli.md:220,640`, `CHANGELOG.md:38`: all three commands, status rows, JSON keys, --check behavior. A-warning remains `internal/shell/shell.go:58`. No leaf hook logic, SPEC_PIN, ledger, or vendored-spec changes. One inherited symlink-test skip message clarified; behavior unchanged. |

Independent validation uses bash with `set -o pipefail`, Go 1.26.0 darwin/amd64, and `CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1`. Exact commands and outputs are attached in the review transcripts. Full unmasked CLI/envprofile suites and native Windows execution are accepted from the attached revision-4 hosted validation, not claimed as rerun locally. Hosted run 35257888444 reports exit 0 and Linux/macOS/Windows test, lint, race and gate lanes success; optional rose-air/candidate lanes skipped. No full landing suite rerun.

Known bound: native Windows/MSYS behavior is supported by the unchanged shared canonicalizer and sibling evidence plus hosted Windows lane; this macOS review does not independently execute Windows. Scope remains native Windows operands; the sibling-documented direct MSYS CLI operand limitation is unchanged.

Validation anomaly: first CLI invocation exited 1 before any test with `acquire package host GOROOT test lock: context deadline exceeded` (301.014s). The first M1 invocation was killed by the Go wrapper at 180 seconds while its stack was in TestMain → AcquireHostGOROOT → AcquireHomeOnly and was explicitly NOT counted as a killed mutant. Another process held the shared lock. Bounded retries retain the unmodified lock and test harness. No lock bypass or host-process termination.

The CLI bounded retry passed (24 selected top-level tests, 256.322s, exit 0). A second two-minute mutant attempt again reached only the host lock and was not counted; retry uses an eight-minute wrapper bound without changing TestMain or lock semantics.

Validation results: build + vet + gofmt exit 0 (no formatting output); full hookapproval/shell/envfiles packages exit 0; envprofile TestStatus mask exit 0 (29.880s); targeted golangci-lint exit 0 (0 issues). POSIX-only initial vector run passed with six explicit PowerShell capability skips. After finding sibling-documented `/tmp/pwsh/app/pwsh`, reran vectors with that PATH: 14/14 cases passed, zero skips (62.257s). Fresh CLI binary + emitted hooks: sh, bash and pwsh each source silently after approve and emit shell_hook_env_unapproved after revoke (all exit 0). Full original shell package was also green before adding pwsh to PATH.

## Adversarial result and coverage bounds

Two class-narrowing mutants detected out of two attempted (2/2, zero survivors):
- M1: retain recorded-but-missing inventory only for manager records; operator records disappear. `TestStatusRecordedButMissingStaysInInventory` fails at hook_posture_test.go:107 (exit 1).
- M2: retain unreadable record association only for manager records; operator records lose approved_by. `TestStatusUnreadableCandidatesKeepRecord/recorded` fails at :168 (exit 1).

Both mutate production `internal/hookapproval/hookapproval.go` only in the disposable mutant copy, run committed command-entry tests with -count=1, and restore from original bytes. The restored pair passes (exit 0, 3.314s); cmp against the frozen source succeeds. These prove the two narrowed classes, not an exhaustive mutation audit of every predicate. Ordinary command/reapproval/revoke/listing and R3 coverage is the passing baseline suite, not additional mutant coverage.

## Checklist conclusion

All task DoD items assessed: command semantics, posture, tests/E2E/vectors, docs/changelog, scope, architecture, build/vet/format/lint, and task-scoped evidence satisfied. Reviewer logbook attached; no repository logbook edit under campaign rules. Relevant tests pass, with original pre-test lock failures retained transparently. The conditional changes-requested checklist item is not applicable to this accepted verdict.

Evidence: TASK-260910-3ungjy_review-transcripts-rev4.md (independent commands and results), TASK-260910-3ungjy_review-logbook-rev4.md (anomalies), and existing TASK-260910-3ungjy_change-request_rev4-validation.log (hosted lanes). Goal queried immediately before verdict: run is not goal-bound.
