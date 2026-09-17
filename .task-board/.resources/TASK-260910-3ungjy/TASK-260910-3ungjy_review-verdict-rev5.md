# TASK-260910-3ungjy — independent review, revision 5

Verdict: **accepted**. No blocking findings; R1–R3 remain closed and the package-lock combination is coherent. Record through accept_cr(revision=5), routing to integrating; reviewer does not mark done.

Candidate: `9d863b928d2f60d3f7e6dbee8fd38b6043bc7ebd`; base `b92bf5ea03840276cd87bcc122ee8e47c7fb41ab`; refreshed sibling checkpoint `d15e2d5282c0732056c9b4485c21f71bb66533ac`.
Patch SHA-256: `083bcf20569fcaf2d838a81bb9ea56f25e2f538711fcef0ee230294bf6389285` (verified).

## Review findings

| Requirement | Assessment and production evidence |
|---|---|
| Trunk combination | `main.go:71,1102,1138,1152` preserves trunk's refresh help, usage, shared resolve/refresh case and `cmdProjectResolve` call. Hook help/dispatch at `:81,197` and status at `:732,778,812` remain separate, with no shadowed case, duplicate registration or lost flag. Compared `git log -1 6645b9b -- cmd/curator/main.go`, its patch and current source. Rev4-to-rev5 main.go delta is exactly the four trunk hunks. Both CR patches have identical file headers and added/removed lines modulo patch context/index headers. |
| Approve | `cmd/curator/hook.go:53`: shared `Canonicalize` (`hookapproval.go:131`), current-byte read before state write, distinct absent/unreadable errors, operator SHA-256/RFC3339 record, re-record after change and byte-preserving matching-record idempotence. Reuses `Upsert` and atomic `writeRecords` (`hookapproval.go:583`). |
| Listing | `hook.go:97`, `hookapproval.go:390`: sorted tolerant Scan; malformed lines reported and skipped; no mutation or repair. State read failure returns error, not empty approval list. |
| Revoke | `hook.go:123`, `hookapproval.go:351`: existing canonical Revoke helper; missing record returns before publication, reports nothing to revoke, exit 0. Command test checks byte-identical state. |
| Posture and R1 | `hookapproval.go:476,496,501`: associate record before stat; absent recorded files retain approved_by and file=missing; optional absent unrecorded files omitted. `NonCurrent` at :455 and `hook.go:176` fail --check without claiming a digest comparison. |
| R2 | `hookapproval.go:510,515,542`, `hook.go:193`: unreadable candidates retain record/approved_by and explicit unreadable qualifier. Deterministic directory-as-file fixtures; both production commands, text and JSON, with otherwise-current env matrix. |
| R3 | `hookapproval.go:567`, `main.go:785,812,824`, `envprofile/status.go:132,216`: state read failure travels as JSON warnings and non-current; truly absent state remains ordinary warning. Both command entry points have paired absent/unreadable tests. |
| JSON contract | `cmd/curator/status_test.go:1745`: exact four top-level keys for normal no-build report; explicit no-builds assertion; two approved/manager rows with exactly three keys. Conditional warnings key is intentional and documented. |
| Windows harness | `hook_posture_test.go:64` decodes JSON and compares exact paths; no added folding/GOOS skip. State path itself is a directory for portable unreadability. |
| Scope and documentation | `docs/cli.md:220,640`, `CHANGELOG.md:28`: three commands, posture, JSON keys and --check behavior. `internal/shell/shell.go:58` remains A-warning. No leaf changes to hook text/trust decision, SPEC_PIN, ledger or vendored spec. Existing sibling symlink skip wording changes only, already reviewed in prior rounds. |

Review ran from a frozen archive under `/tmp/3ungjy-review5.GiQ1xD`; mutants use a separate `/tmp/3ungjy-review5-mutant` copy. Verified 885/885 non-board archive blobs against the tree and 12/12 live leaf paths against the candidate. No candidate changes, commits, branch operations or user configuration writes.

## Validation bounds

Shell: bash with `set -o pipefail`; Go 1.26.0 darwin/amd64. All tests set `CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1`. Detailed commands and complete outputs are in `TASK-260910-3ungjy_review-transcripts-rev5.md`.

Full unmasked CLI/envprofile suites and native Windows execution are accepted from existing revision-5 hosted validation, not claimed as independently rerun. Its gate run 35266735589 reports exit 0 and Linux/macOS/Windows test, race, lint and gate lanes green (optional candidate/rose-air lanes skipped). Verified hosted commit `824ccea90c6991b833ff0f6f2861ed774bc03f07` has the exact candidate tree above. No full landing gate rerun.

Known bound: direct MSYS CLI operand limitation documented by the sibling and prior reviews is unchanged; native Windows spelling/case folding uses the shared canonicalizer. Local Windows execution is not claimed. Mutation coverage is limited to the two specified classes, not every predicate.

Validation setup anomaly: the first build/vet ran before archive extraction completed and failed on missing local packages (exit 1). These are not candidate failures or passes. The archive was subsequently blob-verified and build/vet rerun successfully. The failed log is retained transparently. Task-scoped review logbook records this; repository LOGBOOK.md remains untouched per campaign rules.

## Independent results

- Build, vet and gofmt: exit 0; formatting output empty.
- Full hookapproval, shell and envfiles packages: exit 0 (3.038s, 36.203s, 4.489s).
- Combined CLI command: exit 0, 401.734s; 49/49 top-level tests passed, comprising 25/25 project resolve/refresh tests and 24/24 hook/status tests. R1–R3 exercised through both production commands in text/JSON and --check modes.
- Envprofile TestStatusShellHookTrust mask: 2/2 passed, exit 0, 28.114s.
- Targeted golangci-lint using repository config: exit 0, 0 issues.
- Trust vectors with /tmp/pwsh/app on PATH: 14/14 cases passed, zero skips, exit 0; covers POSIX shells and PowerShell and both rollout profiles.
- Fresh binary plus generated-hook E2E: 6/6 activation checks passed (sh/bash/PowerShell × approve/revoke), every invocation exit 0. Approval sources silently; revoke sources with shell_hook_env_unapproved under A-warning.
- M1 narrows digest-sensitive reapproval to reuse any operator record: committed TestHookApproveReRecordsAfterChange fails at hook_test.go:150, stale digest (exit 1). Its 215.477s includes shared host-lock waiting, which is not counted as evidence; the named assertion failure is.
- M2 narrows non-current checking to manager-approved rows: committed TestStatusReportsShellHookTrustPosture fails at hook_test.go:488, changed operator file wrongly exits 0 (exit 1, 1.942s).
- Mutation coverage: 2/2 targeted class-narrowing mutants detected, zero survivors. Original bytes restored; cmp exit 0. Both committed tests pass after restoration (exit 0, 3.720s). Mutant runner itself exits 0.

## Checklist and lifecycle

All task DoD items reviewed and satisfied: command semantics, posture, regression/E2E/vector checks, documentation, architecture/scope, build/vet/format/lint, and durable task-scoped evidence. Existing checklist remains ticked; conditional changes-requested item is not applicable. Candidate remains untouched. Run goal queried: not goal-bound. Evidence attached before acceptance: this verdict, TASK-260910-3ungjy_review-transcripts-rev5.md, and TASK-260910-3ungjy_review-logbook-rev5.md. Hosted full-suite evidence is TASK-260910-3ungjy_change-request_rev5-validation.log.
