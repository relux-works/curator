# TASK-260910-5nrmtt revision 5 — CHANGES_REQUESTED

Candidate tree: 2990523057576c695bff9c62f1a3c5dd1965c87d.
Base: 38e62cb8f4b0a427625e17dc6a02cf825f4407a4.
All 12 changed working-tree files independently byte-compared with the candidate blobs: match. Review scope is the binding rework-3 / review-rev5 note and repository-transport revision 1. No candidate files modified; probes and mutants ran in a temporary git-archive copy of the exact tree.

## P1: preprocessing bypasses the required whole-output closed grammar

internal/buildrepo/transport.go:311–312 applies the unanchored framingGlue replacement before splitting stderr. This invents line boundaries in malformed input. A single original line that matches no closed-table entry can therefore authorize fallback.

Reproduced through AcquireNetworkResolved with this first fetch stderr:

```
fatal: unable to access 'https://fixture.test/repository.git/': The requested URL returned error: 503fatal: the remote end hung up unexpectedly
```

Expected: one fetch, refusal (unconsumed tail; no original full-line match).
Observed: TWO fetches, err=nil, alternate succeeds.
TestReviewRev5GluedLineRefuses exits 1 with `malformed full line produced 2 fetches, err=<nil>` (1.786 s). Reproducer attached as TASK-260910-5nrmtt_review-reproducers-rev5.patch.

This violates the binding requirement to split captured output into lines and require every non-empty line to match the closed full-line table, plus the contract's malformed-response refusal. It is not a request to support additional diagnostic shapes. Remove framingGlue repair from classification; preserve only actual newline boundaries. A helper that emits a malformed unterminated diagnostic must fail closed. Update the SSH fixture to emit properly terminated diagnostics when testing a positive availability case; retain the new glued-line refusal regression. Keep Windows refusal and unrelated accepted behavior unchanged.

## Independent verification

Shell: zsh. Commands ran to completion; exit codes are direct process results.

- `go test -p 1 -count=1 ./internal/buildrepo -run 'Test(Review|ResolvedTransport|Classify|TransportResolution|SemanticFallback)'`: exit 0, 69.464 s. Includes prior reviewer probes, production-entry 14/14 accepted shape rows and 42/42 malformed table variants, admission, timeout/child pipes, locked-content, legacy and sanitization cases.
- `go test -p 1 -count=1 ./internal/gitcred`: exit 0, 3.823 s.
- `go vet ./internal/buildrepo ./internal/gitcred`: exit 0.
- `gofmt -l internal/buildrepo internal/gitcred`: exit 0, no output.
- `git diff --check`: exit 0.
- New malformed-line probe: exit 1, intended failure proving the finding (0/1 refusal rows satisfied).
- Mutant A: unmatched-line return changed from `(FailureUnknown,true)` to `(FailureUnknown,false)`; `go test -p 1 -count=1 ./internal/buildrepo -run '^TestReviewRev4ClosedGrammar$'`: exit 1, audit-facility case made 2 fetches with nil error. KILLED.
- Mutant B: HTTPS 50x pattern loosened with `.*` before `$`; same test command: exit 1, `503; object verification failed` made 2 fetches with nil error. KILLED. Required mutants killed: 2/2.
- Restored transport.go verified byte-identical to candidate blob. Same rev4 probe command after restore: exit 0, 2.008 s.

Attached rev5 validation log inspected, not rerun: GitHub run 35125542536 reports exit 0 and successful lint, conformance, Ubuntu/macOS/Windows tests, Ubuntu/macOS race, naming and gate self-tests. Rose-air and candidate-suite rows skipped. Local Windows execution and rose-air are unverified here; accepted Windows refusal is not reopened. Full module suite was not run locally. Producer evidence alone was not used for acceptance.

## Scope and bounds

Diff stays in buildrepo/gitcred plus draft documentation. Existing wrapper/broker, strict admission, shared deadline, locked raw-object verification and sanitized errors are retained. Draft label and no-caller-wiring boundary documented; no frozen schema changes. The whole-line table itself contains no `.*`/`.+`; the issue is the preprocessing path around that table. Existing known test rows pass, but do not prove arbitrary malformed output is refused, as the additional row demonstrates.

Run goal queried: no active goal; no directives. Verdict: CHANGES_REQUESTED, route to to-dev. No commits, runtime-home writes, live credentials, or candidate edits. LOGBOOK.md edits prohibited by campaign rules; this task-scoped verdict and board notes persist the finding instead.
