# TASK-260916-1xib1x revision 2 — independent review

Verdict: changes_requested; route to `to-dev`.

Exact candidate tree: `f684c415e35e7c9ebd716cf23c124bbbbf2bcc92`; base `d275e2d9d23afdde75abecdc51971b663047ffd5`. Reviewed and tested a disposable archive inside the assigned Story worktree. Product working files were not modified. Revision 1 → 2 changes are confined to policy records, global-entry coverage, their tests and documentation.

## Required rework: existing cached verdicts never gain policy records

`internal/audit/audit.go:255–259` returns on a findings-cache hit before the only call to persist `script_policies` at lines 263–264. Cache identity/version was not changed. A valid verdict produced before this change has no `script_policies`; every subsequent production Gate call continues to use it without recording any command identity. The in-memory Report and CLI correctly compute the policies, but the stored audit record stays incomplete indefinitely. This violates the binding revision-2 requirement that the per-command identity appear in the stored verdict, including commands with no warning. It also contradicts the new documentation's stored-verdict claim for existing installations.

Reproduction attached as `TASK-260916-1xib1x_rev2-cache-regression_test.go`. Place it in internal/audit of a disposable exact candidate and run:

```sh
go test -count=1 ./internal/audit -run '^TestReviewerExistingVerdictGetsPolicyRecord$'
```

Independent result: exit 1, 1.002s. The test uses production Gate to create an enforced no-host verdict, removes only the new script_policies member to model the valid pre-R4 format, calls Gate again, then reads the stored file. Failure: `production Gate left existing verdict without per-command identity`. Existing TestReportCarriesScriptPoliciesOnEveryPath asserts only the in-memory Report on cache hits; TestStoredVerdictPersistsScriptPolicies covers only a fresh cache.

Refresh/backfill the persisted record on a writable cache hit, or deliberately invalidate old-format verdicts so they are regenerated. Preserve cached findings semantics and GateReadOnly's no-write contract. Add a regression for an existing old-format verdict, including mixed commands or an enforced no-warning command. This is ordinary implementation rework; no human decision is needed.

## Independent baseline verification

Darwin amd64, Go 1.26.0, zsh. Exact spec pin `87a0d0060bad64ab883d007dcdf35df7485368bf` extracted under this worktree.

- `CURATOR_CONFORMANCE_ROOT=<pin>/conformance/v1 go test -count=1 ./internal/scriptpolicy ./internal/audit ./internal/skillcheck`: exit 0, 1.037s / 2.105s / 2.333s.
- `go test -count=1 ./cmd/curator ./internal/install -run 'Test(CLIAudit.*|ScriptAuditLabelsAt.*|DeclaredOnlyWithHosts.*|DeclaredOnlySchema8ScriptCommandInstallsUnchanged)$'`: exit 0, CLI 1.527s, install 154.897s.
- `go build -o ../TASK-260916-1xib1x-curator-r2 ./cmd/curator`: exit 0.
- All four vector shapes are driven at CLI, project install and default global install. Global also drives declared-only-with-hosts and distinguishes validation from audit warning lines. Legacy launcher rows pass.
- Mixed CLI records include guarded → script-worker-v1 and tool → explicit (none), with exactly one declared-only warning. JSON uses an explicit empty policy string for absence. Fresh stored verdicts include both records. No-warning commands are retained.
- Warning semantics, strict/fail_on bypass, docs and Added entry remain correct; warning classes are absent from scriptworker applied-control code. No new runtime refusal/attestation qualification claim.

## Landing evidence accepted, not rerun

Independently queried https://github.com/relux-works/curator/actions/runs/35704285880: success. Head `c7146052169f01b8afef85cc0e739e99ff1cd308` resolves to exact candidate tree `f684c415e35e7c9ebd716cf23c124bbbbf2bcc92`. Hosted Ubuntu/macOS/Windows tests, all three gate self-tests, Ubuntu/macOS race, lint, interop and naming succeeded. Rose-air and candidate suite skipped; no ARM64 passing claim. These are job-status results; raw hosted artifacts were not downloaded/replayed. The full runtime-owned landing suite was not rerun manually.

## Logbook note

2026-09-22: revision 2 fixes the global coverage gap and fresh per-command records, but the old findings-cache fast path bypasses stored-record backfill. Green exact-tree CI misses this upgrade scenario. No logbook executable is available; campaign rules prohibit LOGBOOK.md edits. This task-scoped verdict and board note preserve the finding.

## Independent mutation attacks

All five revision-1 attacks rerun on the exact revision-2 archive; every mutation restored from original bytes. All 16 changed candidate files match the candidate tree after restoration.

| Mutation | Production entry and result |
|---|---|
| Also label enforced commands declared-only | CLI audit vector rows, exit 1 |
| Narrow network label to more than one host | CLI single-host network vector, exit 1 |
| Narrow declared-only to commands lacking win_path | CLI both declared-only vectors, exit 1 |
| Append script warnings to errors | CLI labelled rows refuse, exit 1 |
| Set only global.go Subject.Commands to nil | Default Global four vectors + negative row, exit 1, 100.949s; exact audit warning absent while validation warning survives |

Measured ratio: 5/5 killed, 0/5 survived. Global coverage is repaired: 4/4 named vectors plus the declared-only-with-hosts negative row. Local CLI/project/global rows cover 4/4 vector shapes each. The separate cache-upgrade regression fails on the unmutated candidate; mutant success does not discharge that defect. The vector-key classification remains consumed; the broader 12/12 classification is not a new claim of full worker/platform qualification by this review.
