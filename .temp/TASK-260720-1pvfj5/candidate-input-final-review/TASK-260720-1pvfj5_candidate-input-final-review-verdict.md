# TASK-260720-1pvfj5 candidate-input final review verdict

Date: 2026-07-30
Role: reviewer
Verdict: ACCEPTED
Route: done
Reviewed composite: `.temp/TASK-260720-1pvfj5/rework/composite`

## Verdict

The focused candidate-input provenance rework closes the P1 defect recorded in
`TASK-260720-1pvfj5_review-verdict.md`.

`candidate_ref` and `candidate_root` are now mutually exclusive. The candidate
job invokes the validation at workflow line 319, before the candidate checkout
at line 333 and candidate identity recording at line 359. An ambiguous
revision-plus-root invocation exits 1 with the intended mutual-exclusion
diagnostic. Ref-only and root-only invocations each exit 0.

The implementation meets the current task description, acceptance criteria,
mandatory focused-rework instructions, and project architecture. No changes are
requested.

## Independent focused checks

All checks were run against the live composite:

| Check | Exit | Evidence |
| --- | ---: | --- |
| `candidate-suite.sh verify-inputs <full-ref> /candidate/root` | 1 | Expected rejection: candidate revision and candidate root are mutually exclusive |
| `candidate-suite.sh verify-inputs <full-ref> ""` | 0 | Ref-only remains valid |
| `candidate-suite.sh verify-inputs "" /candidate/root` | 0 | Root-only remains valid |
| `bash .github/ci/gate-selftest.sh` | 0 | 74 passed, 0 failed |
| `bash -n` on the two changed shell scripts | 0 | Syntax valid |
| `actionlint .github/workflows/ci.yml` | 0 | Workflow valid |
| `shellcheck .github/ci/candidate-suite.sh` | 0 | No findings |
| warning/error ShellCheck on `gate-selftest.sh` | 0 | No warning/error findings |
| focused patch reverse-apply check | 0 | Patch exactly describes the live three-file change |
| live manifest versus focused rework manifest | 0 | `cmp` equality |
| attached manifest versus task-local manifest | 0 | `cmp` equality |
| staged-index check | 0 | Nothing staged |

The ordering check also verified that each named workflow step occurs exactly
once: validation line 319, candidate checkout line 333, identity recording line
359.

## Exact delta and preserved inputs

The accepted final-integration manifest and focused rework manifest each contain
374 entries. Their only changed paths are:

1. `.github/ci/candidate-suite.sh`
2. `.github/ci/gate-selftest.sh`
3. `.github/workflows/ci.yml`

There are no additions, removals, product changes, or unrelated overlay
changes. The seven independently accepted blocker-owned product paths retain
the exact hashes from the accepted final-integration manifest. The earlier
372-to-374 proof and its seven-path product delta therefore remain intact.

The live composite matches the attached 374-entry focused manifest
byte-for-byte. The focused patch and result artifact attached to the board also
match their task-local copies.

The workflow still declares Go 1.25.5 and exactly one released `SPEC_PIN`:

`00b1688a9b2457ca397a0bb550acf47cad8ee967`

The pin is unchanged. Candidate evidence remains explicit, non-default, and
stamped candidate-only with no release or conformance claim. No rc.4 release
wording was introduced.

## Reused accepted heavy evidence

No full Go, default-pin, candidate, or race suite was rerun, as required by the
focused final-review instruction. The attached final-integration evidence
packet is byte-identical to its task-local source, and the focused manifest
proves that product code, `test-gate.sh`, platform ledgers, Makefile, vectors,
timeouts, fixtures, and suppression policy did not change.

The previously accepted evidence therefore remains applicable:

- default released-pin test gate: exit 0, 33 served / 7 explicit deferred;
- explicit rc.5 candidate test gate: exit 0, 40 served / 0 deferred / 0 excluded;
- one serialized full race gate: exit 0, with no race diagnostic or `FAIL` token;
- pinned golangci-lint v2.12.2: exit 0, 0 issues;
- gofmt, vet, build, no-broad-suppression, and deterministic godriver
  cancellation gates: exit 0;
- ledger consistency: exit 0, 49 rows across linux/darwin/windows.

Static inspection confirms the workflow matrices still cover Ubuntu, macOS,
and Windows; the race matrix still includes Ubuntu; and the shipped platform
ledger still requires Windows DACL/reparse/`.cmd` cases plus Unix ownership,
no-follow, read-only-source, resource-policy, and executable behavior.

## Stale checklist item

Checklist item 1 says CI must pin an rc.4 protocol commit. That text is stale
and is explicitly superseded by the task description, scope, acceptance
criteria, mandatory review instructions, and the existing 2026-07-20 board
note. Truthful acceptance requires leaving the released pin at the prior
qualified revision and must not check or implement the stale rc.4 statement.
The unchecked stale item is therefore not an acceptance deficiency.

No code, workflow, evidence input, pin, or repository index was modified by
this review.
