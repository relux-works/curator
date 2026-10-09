# TASK-261008-1j34ro — N8 audit content identity review

Verdict: changes_requested. Revision: 1, CR-TASK-261008-1j34ro-1.
Candidate tree: ecf22dd17f8ede7d8eb36659a8cfced12e91b643.
Base: 3d395ffe72ec979e2ef1d3d792655a7f405785cc.
Route: to-dev; ordinary implementation and regression rework, no external blocker.

## Blocking findings

### F1 — Explicit empty --allow bypasses validation before filesystem access

At cmd/curator/main.go:2141 and :2157, flag presence is inferred from `*allow != ""`. Both `audit --allow "" --reason "synthetic approval"` and `audit --allow= --reason "synthetic approval"` parse successfully but skip ParseDigest and enter loadConfig at :2153. parseInterspersed preserves the explicit empty value (flag.Parse accepts it). The CLI then enters ordinary target selection: nearestProjectRoot calls os.Stat and auditTarget loads a manifest. Thus an explicitly supplied non-digest reaches configuration/filesystem access rather than the required usage refusal. Pin itself correctly rejects the empty string, but the CLI never reaches it on this path.

This is established by source/control-flow inspection, not a claimed newly executed test or a new path-escape exploit. A deterministic regression can drive production run with a config source whose Load records calls or returns a sentinel error: explicit empty --allow must return exitUsage and call Load zero times. Preserve ordinary `audit` behavior when --allow is absent.

Requested fix: detect whether --allow was supplied (for example via flag visitation), validate that supplied value even when empty, and use the same presence decision for the pin branch. Cover both split and equals spellings through run.

### F2 — Current regression does not prove refusal precedes configuration access

cmd/curator/audit_allow_test.go:45-55 uses a stubConfigSource that returns an in-memory config and asserts only nonzero exit and no trust.json. It cannot observe configuration reads, nor distinguish usage refusal from a configuration failure. Reordering loadConfig ahead of ParseDigest while retaining the library refusal is therefore outside what these assertions measure. The task explicitly requires parsing before any filesystem access. This ordering needs a production-entry assertion that configuration Load is never called for each refused value, including empty, traversal, nested path, short digest and non-hex digest; require exitUsage.

Also make assertNoPinState propagate WalkDir read errors at :65-66. Currently every read failure is ignored, making an unreadable subtree look like an absence of records.

Requested proof: hosted green on the new frozen candidate and hosted red with only production fixes reverted; additionally a focused ordering mutant that moves loadConfig ahead of the CLI parse must fail the new no-Load assertion. A delete-only mutant cannot establish ordering. No new research task or human decision is needed.

## Swept surfaces

| Surface | Evidence and result |
| --- | --- |
| CLI flag parsing, explicit presence and refusal before config | All cmdAudit branches and parseInterspersed read. F1 bypass and F2 ordering-proof gap above. |
| Supported grammar and directory normalization | ParseDigest accepts 64 ASCII hex digits, optional lowercase sha256: prefix, surrounding whitespace consistent with existing Normalize; canonical directory is lowercase bare hex. 3/3 positive CLI forms and 4/4 malformed nonempty CLI forms pass in hosted evidence. |
| Pin and PinAtVersion production callers | Pin delegates to PinAtVersion; CLI calls PinAtVersion. ParseDigest executes before filesystem calls for both versions; invalid version is still refused. Direct-library regression passes. |
| Containment before pin writes | pinDir runs before MkdirAll/WriteFile and checks cleaned lexical containment. Parsed bare hex makes escape impossible through supplied identity. Physical symlink containment and filesystem exchange races are outside this N8 path-input finding and are not claimed proven. |
| Pin carrier and trust decision compatibility | Existing version carriers, v1/v2 separation, old-schema pin requirement, preserved pin, revocation and fresh/cached strict-finding controls inspected in hosted ledger; all matched rows pass. No changed read or decision-precedence code. |
| Regression quality | Real run entry used; refusal rows detect original production revert. Explicit empty flag and pre-config ordering absent; WalkDir read failures swallowed (F1/F2). |
| Operator documentation and scope | CHANGELOG under Unreleased/Fixed accurately describes the intended N8 correction, subject to F1. N9 creation-time work is separate; no request to add it here. Six changed paths reviewed, no unrelated product delta found. |
| Free hunt | Checked flag split/equal spellings, empty versus absent, parser normalization, alternate pin entry, version refusal, containment and test observation. No additional blocking findings beyond F1/F2. |

## Validation evidence and provenance

Reused producer evidence after independently checking exact identity; no local Go tests, test binaries, or product code changes in this reviewer run.

- Green: https://github.com/relux-works/curator/actions/runs/37869043464 (head 7090e6309ad6b5a83baae3d4cfc048bea4a8acd1). Its tree is exactly ecf22dd17f8ede7d8eb36659a8cfced12e91b643, not merely a similar patch. API conclusion success. Downloaded test-evidence-ubuntu-latest and inspected JSON pass rows: CLI parent plus 7 subtests = 8/8 pass, library regression = 1/1 pass. Existing pin/version/revocation/strict-finding controls matched above pass. Enumerated coverage does not include F1/F2.
- Producer red: https://github.com/relux-works/curator/actions/runs/37873727909 (head 662727550da457ac8c8095fd0aa23ec5250b4476). Independently diffed against green: only the three production files reverted; new test and changelog bytes retained. Named mutant: original-unvalidated-pin-writer-and-CLI. Accepted producer disposable-clone execution as evidence, not claimed run by this reviewer. API conclusion failure. Independently downloaded test-evidence-ubuntu-latest: fail rows are the four TestAuditAllowPinsOnlySupportedContentIdentity refusal subtests and their parent, TestPinRefusesNonDigest, and the corresponding two package summaries. No other failing test is present. The three positive CLI cases pass on red. This verifies the original bug detection, but does not establish missing empty-input or pre-config ordering coverage. No reviewer-run narrowed/ordering mutant was executed; that proof is required for F2 on the next revision.
- Handoff gate: https://github.com/relux-works/curator/actions/runs/37878233932 (head a905273976ff99421f823c1bedb143d2309470ab). Tree also equals this candidate. Attached CR validation log reports success: 20 jobs success, 2 trigger-inapplicable skips, exit 0. It reports case coverage unknown; it is not proof of the missing input/order cases.
- Accepted local producer compile-only evidence from TASK-261008-1j34ro_results.txt: go vet ./... and go build ./... exit 0, gofmt clean, changed-package lint 0 issues. Reviewer personally reran git diff --check (exit 0), examined all six changed paths and exact green/red tree identities, downloaded hosted evidence and read board artifacts. Did not rerun vet/build or any test suite.
- Fresh origin main advertised 3d395ffe72ec979e2ef1d3d792655a7f405785cc, matching CR base at review time; no upstream delta observed.

## Next producer and reviewer cycle

Resolve F1/F2 together, retain existing supported-identity controls and carrier/version semantics, attach the next candidate's hosted evidence and ordering-mutant test name, then hand off a new CR revision for review. No LOGBOOK edits made, following the task brief; all review findings are persisted in this task-scoped outcome.
