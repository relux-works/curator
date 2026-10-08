# Review verdict \u2014 TASK-261008-2uq6jo \u2014 N6 credential answer bound

Verdict: changes_requested. Reviewed CR-TASK-261008-2uq6jo-2 revision 2. Route to to-dev; acceptance is withheld.

Candidate identity: base 3b6481c15d61b08d3f0fae7c3269329555336709, candidate tree 7b2f1409d3e4e1ba1b9b5c77615b6c9651549cb6. Fresh origin HEAD advertisement and fetched main both equal the base; there is no upstream overlap to resolve. Hosted snapshot 20b9f345b7545dabd71e5e617beab50b4e440927 has exactly the candidate tree. Reviewed all three changed paths.

## Required rework

1. Windows coverage remains absent. internal/gitcred/credential_bound_test.go:15-16 and :51-52 unconditionally skip both new tests on Windows. Revision 2 merely changes the reason strings to match the broad existing `(is|are) exercised (on|by)` allow pattern. The steering instruction prefers portable fixtures and permits a skip only if portability is genuinely impossible; the results describe classification success but no platform constraint or failed portability attempt. The package already re-executes its own test binary through TestMain, fakeAccess and fakeGitMain on every platform, explicitly without a shell or build step. That provides a viable controlled Git transport for exact frames; a portable helper can preserve the real-Git reproduction. Implement portable fixtures so the Windows lane exercises oversized refusal and cap-1/cap/cap+1. Keep real-Git coverage and exact-byte acceptance assertions. A green skip classifier is not evidence of the invariant on Windows.

2. The requested mutation proof is missing. The evidence-basis resource accepts old report expected-red evidence and calls the green half asserted by construction. No disposable-clone mutant result is attached. I did not run a mutant in this round and make no mutation-coverage claim. After fixture rework, run targeted hosted checks from a disposable clone, with -count=1: disable only Access.call's overflow refusal to establish production-entry wiring; narrow overflow detection from len(accepted) > b.remaining to len(accepted) > b.remaining+1 and require TestCredentialExactFrameBound/1 to fail; change the detection to >= and require the exact-limit /0 positive control to fail. Name the mutants, source identity, hosted execution and observed failures in the task-scoped results. No local go test on the mini. These are diagnostic mutations only; do not hand off mutated source.

## Swept surfaces and evidence

| Surface | Finding |
| --- | --- |
| Shared production boundary | Access.ReadHost reaches Access.call; all credential operations use the same call. Overflow is checked after cmd.Run and before parseAnswer. No bypass found in this delta. |
| Bounded writer | Strict > comparison preserves exact-cap acceptance, sticky overflow records a later nonempty write after capacity reaches zero, retained bytes stay bounded, full write lengths continue draining stdout. Both added refusal clause sites inspected; dynamic mutation proof remains unrun. |
| Regression assertions | Real Git covers 32, 65456 and 65664 secret bytes; controlled Git covers frames 65535, 65536 and 65537 bytes through ReadHost. Unix positive cases preserve secret bytes; overflow cases require absence. Windows cases are skipped. |
| Report controls | Prompting, namespace, provider near miss and failed persistence all pass in the hosted artifacts, 4/4 per platform. |
| CHANGELOG | Unreleased Fixed entry accurately describes whole-answer refusal and no credential for the affected host. |
| Architecture and scope | Small change at the owning production boundary; three paths; no unrelated behavior or source change found. |
| Instruction compliance | R223 honored during review. Windows steer and mutation-proof requirement remain unresolved. |

Hosted evidence accepted from CR validation and independently inspected: https://github.com/relux-works/curator/actions/runs/37838299621 . Its overall conclusion is success; all default hosted test/race/lint jobs pass. Downloaded test-evidence-ubuntu-latest, test-evidence-macos-latest and test-evidence-windows-latest and read terminal events for the task cases in test/go-test.json. Size coverage is 6/6 on Linux, 6/6 on macOS, 0/6 on Windows (two top-level tests skipped). Four report controls are 4/4 on each of the three platforms. This evidence establishes Unix regression behavior, not Windows bound coverage. The self-hosted and candidate-conformance lanes were skipped as reported by this gate.

Reviewer reran locally: go vet ./... exit 0; go build ./... exit 0; gofmt -l on both changed Go files clean; git diff --check exit 0. No local go test, no source edits, no commits, no branch changes, no commit_ack. No full suite replay performed; green evidence was reused only after checking exact candidate identity. Mutation proof was not rerun and is not inferred from the old report.

Findings are recorded in this task-scoped outcome and board notes. No LOGBOOK edit, following the task instruction. This run is not goal-bound (spawn goal reports none). Ordinary implementation/evidence rework; no external blocker or human decision is required.
