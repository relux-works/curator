# TASK-261005-22yioq \u2014 Publish CIP drafts 0002-0006 into curator-spec cips/: review verdict, revision 1

Verdict: **changes_requested**. Route: **to-dev** for bounded documentation rework and another reviewer cycle.

Reviewed CR-TASK-261005-22yioq-1 revision 1, base `b0caf8db9bf14b7da2541cd729751631d05d8a26`, candidate tree `fc65f0a36ad537935b27a60b9577123e2c1c8abb`. All six working files match the candidate blobs (6/6), including after checks. Fresh `git ls-remote --symref origin HEAD refs/heads/main` reports that same base for upstream main: no upstream overlap to converge. Review was read-only for repository content; no commits, staging, regeneration, Go tests, or LOGBOOK edits.

The run-goal query reports this run is not goal-bound. No directives were recorded at the safe checkpoint.

## Numbered findings

1. **[P2] Restore navigable, commit-pinned research links in every CIP.** Current state in CIP-0002 line 23, CIP-0003 line 23, CIP-0004 line 25, CIP-0005 line 23 and CIP-0006 line 24 replaces the source's Markdown companion-evidence link with code spans. Consequently all five published CIPs lack a link to either their source draft or their companion evidence: **0/10 research references are Markdown links**. A reader cannot follow the E/P/C/L identifiers into the evidence; the local-link validator ignores these code spans, so its passing result does not establish that this required evidence is linked. The publication brief explicitly requires links by repository-relative path and commit. The producer's cited security-audit convention supplies plain citations, but does not override this task's link requirement. Use a `.research/<filename>` label with a commit-pinned cross-repository blob URL for each source and companion, retaining the stated evidence bounds and keeping evidence in curator. All 10 files exist at the cited full commit `3d9aa98768158cd5bc64ee884896e7a7064a1b55`, match current curator origin/main byte-for-byte, and their full-commit GitHub blob URLs were independently checked (10/10 HTTP 200, curl exit 0). No source relocation, evidence duplication or decision changes are needed.

2. **[P2] Escape four inline pipes in CIP-0002's operator-control table.** `cips/CIP-0002-project-context-in-managed-launches.md` lines 77, 80, 81 and 82 contain unescaped `off | admitted`, `enabled | disabled`, `deny | review` and `off | private`. GitHub-flavored table parsing treats those pipes as cell separators even inside backticks. The two-column table renders the second cells as only `Closed enum `off`, `` `enabled``, `` `deny`` and `` `off``; the mode/default, approval, fleet-ceiling and memory text after the pipe is discarded. This defect existed in the research Markdown, but publication must preserve the substance in its rendered form as well. Escape the four pipes as `\|` inside their code spans (or express each choice as separate code spans without an internal separator). A markdown-it-py `gfm-like` rendering reproduced all four missing semantics; an in-memory escaping-only variant restored all four (4/4). No repository files were changed by that probe. The complete changed-doc table scan inspected 31 tables / 267 rows; these are the only four inconsistent-width rows (4/267).

## Swept review surfaces and merged checklist evidence

| Surface | Coverage and result |
|---|---|
| Scope and frozen artifacts | Exactly five new CIP files and one README modification; all other paths, normative protocols, schemas, conformance, released/frozen artifacts, CHANGELOG and LOGBOOK untouched. Pass. |
| Template and status | Required metadata 25/25; major sections 50/50; Options considered and Recommendation subsections 10/10. Every new CIP and index row is Draft, with no operator acceptance invented. Pass. |
| Source fidelity | Compared every draft with current curator origin/main `fae2ff9cab17a031c26a2b4c776afd8dc5e8b4f6`. All seven design/security/migration/specification/implementation/test/questions sections are byte-identical per file (35/35). Only titles/status normalization and evidence-location prose differ. All alternatives and 31/31 numbered operator questions remain (7, 6, 5, 8, 5). Pass for text; rendered fidelity fails finding 2. |
| README | Five unique, linked Draft rows; filenames exist and titles agree with each CIP H1 (5/5). Pass. |
| Evidence provenance | Source plus companion files present at the cited commit and unchanged at current curator main (10/10). Companion evidence was not copied into spec. Navigable links missing (0/10): finding 1. |
| Architecture/process | Appropriate docs-only proposal publication; no runtime/schema behavior or adoption decision introduced. Binding evidence-link and rendered-document requirements need the two fixes. |
| Formatting/rendering | Whitespace checks pass. Table scan and real GFM-style rendering expose the four rows in finding 2. |
| Public-data hygiene | Grep over all six changed documents: personal-path lines 0, private-host lines 0, machine-user lines 0, employer-label lines 0. No personal host/employer disclosure found. The template-mandated Owner handle, public repository/doc domains, provider names and illustrative reverse-domain service ID remain. Pattern scope is recorded below; this is not a universal named-entity detector. |
| Existing links | README local targets 5/5 exist; repository local-link validator passes; all 14 distinct external Markdown URLs return HTTP 200 after redirects, curl exit 0. These counts exclude the missing research links. External fragment/content semantics were not exhaustively attested. |
| Tests/docs gates | Independently reran applicable Python validation, Python unit tests, Skillfile-source independence, whitespace, table rendering and HTTP reachability. Results below. No Go tests run, as expressly instructed. |

## Reviewer-executed commands and exit codes

Python commands used the repository's existing development virtual environment, with `-B` to avoid bytecode writes.

- `python -B tools/validate.py`: **0** \u2014 73 schemas, 1,294 vector files; includes repository-local Markdown links.
- `python -B -m unittest discover -s tools -p 'test_*.py'`: **0** \u2014 **672/672**, 487.398 seconds, OK. The command remained attached and was polled in bounded calls until completion; no background work is left running.
- `python -B tools/verify_skillfile_sources_independence.py`: **0** \u2014 89/89 schema references byte-identical to rc.10; 28/28 cited clauses present; documented conditional exception allowlisted.
- `git diff --check <base> <candidate>`: **0**.
- Trailing-whitespace grep on all six documents: **1**, **0 matching lines** (grep's no-match result is success for this absence check).
- Template/index/candidate/source comparison scripts: **0**, measured coverage in the table above.
- Table-width probe and markdown-it-py `gfm-like` renderer: **0** as diagnostic commands; **format property FAIL**, four malformed rows and four missing semantic passages. An in-memory escaping variant restored 4/4. This is not claimed as a passing formatting gate.
- HTTP reachability: **14/14 existing external Markdown links** and **10/10 proposed full-commit research blob URLs**, each curl exit **0** / HTTP **200** after redirects.
- Privacy grep: personal paths **1 / 0 matches**; private hosts **1 / 0 matches**; machine user **1 / 0 matches**; employer labels **1 / 0 matches** (exit / matching-line count).

Privacy patterns: `/Users/[^[:space:]`]+|/home/[^[:space:]`]+|[A-Za-z]:[\\/](Users|Documents and Settings)[\\/]|~[A-Za-z][A-Za-z0-9_.-]*/`; `([A-Za-z0-9_-]+\.)+(local|lan|internal|corp)([[:space:]"`/:]|$)|relux-(entrypoint|host)`; `\badministrator\b`; `\bemployer\b|\bReluxWorks\b`. A broader exploratory host scan matched the documented filename `CLAUDE.local.md`, which is not a host; the host-boundary scan above has zero matches.

## Prior evidence and limits

Read the attached producer `TASK-261005-22yioq_results.md` and revision-1 validation log. The results artifact describes an earlier pre-existing Go pin failure; the later CR log records a passing Go generator shard. Neither Go result was rerun or independently adjudicated in this reviewer run, and neither is used to excuse the documentation findings. No full `make validate`, lychee binary run, behavioral implementation qualification or native-provider launch is claimed. HTTP reachability and the existing local-link validator are the applicable link checks rerun here.

## Required next producer cycle

Add the 10 pinned research links and escape the four table separators, preserving Draft status and source decisions. Rerun applicable docs/link checks and GFM rendering, attach fresh task-scoped outcome evidence, then hand off a new CR revision for review. Findings and anomalous rendering are recorded here and in board notes; the task-specific no-LOGBOOK rule prevents repository logbook edits. This is ordinary documentation rework, not an external blocker or a human decision.
