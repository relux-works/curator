# TASK-261010-1ypla7 — research-platform-contract-table: rev2 evidence

Ready for review. Only `.research/261010_platform-contract-table.md` changes in the candidate. No tests, builds, LOGBOOK edits, commits, provider logins or runtime changes were performed.

## Rejection response

| Review finding | Response | Verification bound |
|---|---|---|
| F1 / A-F1, A03.9 | P7 records archive/delete request, operator-only keep policy, each caller role, retention defaults and explicit owner options. | UM §§3, 4.4–4.4.1 at the existing full pin, read afresh. The inconsistent enum is not silently resolved. |
| F2 / A-F2, A07.5 | P8 records qualified version range plus adapter probe and continuity prerequisites. | AR L474–488; B07 L71–77, pinned and read afresh. No actual version qualification claimed. |
| F2 / A-F2, A07.8 | P9 separates two private lanes and public module CI; current invented fixtures prove tooling only. | SG L3–20; SH L69–97; B07 L119–121, pinned and read afresh. |
| F2 / A-F2, A07.10 | P10 records public tagged-module disposition, retained composition root, operator cutover and provenance tense. | SH L3–5, L50–52, L78–97, L202–209; SG L3–6; B12 L49, pinned and read afresh. |
| A-F3 / P2 | I5 ledger locator and refusal semantics; W4 key classes and adapter migration; P11 sweep cadence/ownership and unresolved invocation contract. | UM, CB, DP and AR pinned ranges. Sweep is not invented as a public operation. |
| Alias, goals, baseline notes | LR defined; G1 calls for optional generic-interface rewrite; audit outcome retrieved again (rev3 content, current task status to-review). | No assumption of audit acceptance. Owner precondition provenance exception remains unchanged. |

New source reads: 12/12 authenticated `gh api repos/relux-works/<repository>/contents/<file>?ref=<full-commit>` calls exited 0: UM, SH, SG, SHR, AR, CB, DP, B03, B07, B12, B14, LR. Exact pins and line ranges are in the table. Other rev1 reads and verification counts are explicitly historical. Review A and the deciding rev1 verdict were both read; mandatory repairs and recommended additions are addressed.

## Command evidence

All checks ran as standalone processes without pipes or tee. These are document inspection/packaging checks, not product tests or CI qualification.

| Command / invocation | Real exit | Result |
|---|---:|---|
| `python3 .temp/contract-rev2-check.py .research/261010_platform-contract-table.md` (initial) | 1 | Checker bug: source-index aliases were counted as contract rows (63). No document pass claimed. |
| Same initial checker with `--narrow-retire` | 1 | Same row-counting bug; not counted as mutant evidence. |
| Corrected checker, normal invocation | 0 | `rev2_rejection_regression`: 4/4 rejected surfaces; 52/52 unique rows; 152/152 pinned ranges in bounds; 191/191 pinned file links; 9 consistent tables; 76,851 / 81,920 bytes; bounded hygiene/whitespace checks. |
| Corrected checker with `--narrow-retire` | 1 | Expected failure: P7 loses the admitted `delete` value while retaining its row and citations; named regression check refuses that narrowed set. No source file is mutated. |
| `git diff --check` | 0 | Tracked whitespace check; new untracked research file separately checked above. |
| `git status --short` | 0 | Only the requested research file is untracked; no other repository changes. |

Named regression check: `rev2_rejection_regression`. Production/runtime call site: inapplicable, documentation-only rework. Actual reviewed entry point is the research document supplied to the checker. The narrowing mutant changes its in-memory P7 request set from archive/delete to archive-only. Measured rejection coverage is 4/4 requested surfaces; semantic fact-checking is manual against cited sources. Blind spots: textual checks cannot prove runtime behavior, actual CI success, qualification of a binary version, or correctness of every unchanged rev1 claim. No such property is inferred.

The checker is attached as a task-scoped outcome for reproduction; it expects the pinned source files listed in the table under its temporary source cache ; it derives the mapping from the document source index. It is not added to the repository candidate.

Generic checklist requirements for tests-green and LOGBOOK conflict with the explicit rework brief and are removed as inapplicable, not checked as satisfied. Findings and provenance exceptions are recorded here, in the table, and in board notes. The prior rejection remains attached and this revision routes back through the researcher handoff for independent review.
