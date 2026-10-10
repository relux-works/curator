# TASK-261010-1ypla7 — research-platform-contract-table: rev3 evidence

Ready for review. Candidate: `.research/261010_platform-contract-table.md`, 78,890 bytes (limit 81,920); SHA-256 `b50becef762053c8b40b6b11853546ec1ba3deee2652ceca97f1e94a4ecb29b6`.

## Response to the rev2 deciding verdict

| Finding | Change and source |
|---|---|
| DB-F1 (P1) | P8 now says OWNER DECISION NEEDED: same version only versus qualified version range with adapter probe. Neither is canonical; A07.5 remains open. AR L413, L1814–1825, L1973–1977 establish proposed/open status; AR L474–488 and B07 L71–77 establish the options. Other continuity prerequisites remain. |
| DB-N1 (P2) | Removed blank lines before I5 and P7; all 52 contract rows now remain in six-column tables. |
| DB-N2 (P2) | W4 owning section corrected to CB §§4.3/6.6; heading at CB L233, normative text L235. |
| DB-N3 (P2), clarified by owner rev3 brief | P9 records the real 2026-10-10 owner-session acceptance of board code appearing in public CI logs, confirmed by the task precondition and B07 L119. Private lanes remain as specified in SG L12–17. Acceptance does not relocate lanes; no Git-pinned session record claimed. |
| DB-N4 (P2) | P11 treats the DP L211 service-account timer as an internal service-only CLI entry outside caller dispatch-op/1. This default needs no new decision; exposing a public operation would. |

The open-decision register reflects P8 and P11; a short rev3 changes section records the delta. Other 48/48 contract rows are byte-identical to the starting rev2 document. Only the research file is a repository change. No LOGBOOK edits, product tests, builds or commits.

## Direct reads and command results

Five fresh authenticated `gh api` contents reads, each a standalone process, exited 0: AR and B07 at wiki `2e30edbd54977b306ab53497663fcd94a5ca6cb6`; CB at broker `f910e68f882900f03ffd036ecc2481f35bf8988b`; DP at dispatcher `9cec4115b8666aa1f3325a792359135a9df5a2f7`; SG at session-host `2bb9fda8ed923266532f331c9e692ba88b5c4cb6`. Exact file names and line anchors are in the table. Relevant ranges were read directly. Other source verification remains historical rev1/rev2 evidence, not rerun or claimed current.

| Command / inspection | Real exit | Result |
|---|---:|---|
| Initial source-range display (inline Python) | 1 | Requested nonexistent AR line 1978 after displaying through 1977; partial read was not an absence or a verification pass. Follow-up bounded display of B07/CB/DP/SG exited 0. All cited AR lines were already displayed. |
| `diff -u` of retained rev2 and rev3 documents | 1 | Expected difference-reporting status; diff shows requested corrections, corresponding register wording and rev3 section. |
| Initial document checker, candidate | 1 | Checker defect: source aliases counted as rows and status cell indexed incorrectly. No pass claimed. |
| Initial document checker, narrowing mode | 1 | Same checker defects; not counted as negative evidence. |
| Corrected `python3 .temp/contract-rev3/TASK-261010-1ypla7_rev3-document-check.py` | 0 | Named `rev3_rejection_regression`: 8/8 document checks. |
| Same corrected command with `--narrow-version-options` | 1 | Expected-red: 7/8. Keeps P8's open-status gate but narrows its options to same-version-only; DB-F1 check rejects the missing qualified-range option. In-memory mutation only. |
| Standalone row/delta inspection (inline Python) | 0 | 4 changed rows (P8/P9/P11/W4), other 48/48 unchanged; 52/52 continuous six-column rows. |
| `git diff --check` | 0 | Tracked whitespace only; research file is untracked. |
| `git status --short` | 0 | Only the requested research document is untracked. |

No command was piped through tee or a pipeline. Named regression check is a document inspection, not a product test or CI gate. Actual input is the research artifact; a runtime production call site is inapplicable. The narrowing mutant leaves the candidate bytes unchanged, as shown by identical SHA-256 output. The attached checker runs from the project root without private inputs.

Measured scope: 8/8 bounded textual checks (two DB-F1 checks, DB-N1–N4, size/row packaging, limited path/token hygiene). Manual source review establishes the cited status distinction; substring checks cannot prove semantics of every unchanged contract, rendering in every Markdown engine, runtime behavior, or any CI qualification. No such broader proof is claimed. Public text contains no personal names, personal paths, session links or source corpus copies; generic platform filesystem contracts remain intentional.

The explicit no-tests/no-LOGBOOK brief supersedes generic checklist items; those two inapplicable items are removed, never checked as green. Findings and the P9 session-provenance exception are recorded in the table, this evidence and board notes. The deciding rev2 rejection remains attached; researcher handoff routes the revised candidate to independent review.
