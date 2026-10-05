# TASK-261005-22yioq — Publish CIP drafts 0002–0006 into curator-spec cips/: revision 2 review

Verdict: **accepted**. Accept CR-TASK-261005-22yioq-2 revision 2 and route to **integrating**. This accepts the documentation publication; the five proposal statuses remain Draft and no operator design decision is implied. Integration is outstanding and belongs to the tracked producer.

Reviewer run: RUN-261005-c5b436. Base: `b0caf8db9bf14b7da2541cd729751631d05d8a26`. Exact candidate tree: `133e96f8f0c9740dc12bfb48cccf3de50eee046e`. All six working files match the candidate blobs before and after verification (6/6). Fresh `git ls-remote --symref origin HEAD refs/heads/main` returned this same base for main, exit 0: no upstream changes to converge. Read-only repository review; no staging, commits, regeneration, repository writes or Go tests.

`task-board spawn goal "$TASK_BOARD_RUN_ID"` was queried initially and immediately before recording the verdict: this run is not goal-bound. Safe-checkpoint directives query returned none.

## Previous findings resolved

Read TASK-261005-22yioq_review-verdict-rev1.md and the producer's TASK-261005-22yioq_rework1-validation.md in full, plus both attached regression/render probes and revision-2 validation log.

1. Revision-1 finding 1, research citations rendered as code spans: **resolved**. Source and companion are now real Markdown links in all five CIPs: 10/10. Each label is the repository-relative `.research/` path and each URL pins the full curator commit `fae2ff9cab17a031c26a2b4c776afd8dc5e8b4f6`. All ten targets exist at that commit and are byte-identical to curator origin/main; all ten URLs independently return HTTP 200, curl exit 0.
2. Revision-1 finding 2, unescaped GFM table pipes: **resolved**. CIP-0002 lines 77, 80, 81 and 82 contain escaped enum pipes. Actual markdown-it-py `gfm-like` rendering retains the full enum and trailing semantics in 4/4 affected cells; all seven body rows have two cells. An in-memory narrowing mutant unescapes only `enabled | disabled` and loses exactly that one semantic value (3/4 retained). Restoring all four original defects loses all four values (0/4 retained). No working file was mutated.

The exact rev1→rev2 Git diff contains only the five evidence paragraphs (one per CIP, including the corresponding short commit pin) and four pipe escapes. README and all remaining lines are byte-identical to revision 1. No mechanism repeats; the current findings set is empty. Previous verdict is legacy prose, so no fabricated finding IDs or repeat-of chains are assigned.

## Swept surfaces

| Surface | Result and evidence |
|---|---|
| Scope and frozen artifacts | Held: exactly five new CIPs plus README; no normative protocol, schema, conformance, released/frozen artifact, CHANGELOG or LOGBOOK change. Full base→candidate path inventory checked; candidate bytes 6/6 verified twice. |
| Template and status | Held: 25/25 metadata fields, 50/50 required major sections, 10/10 required Design subsections. Five Status: Draft values; no invented acceptance. |
| Source fidelity | Held: every published draft compared with curator origin/main via Git show. All seven substantive design/security/migration/spec/implementation/test/questions sections per CIP match exactly after normalizing only the four pipe escapes (35/35). Remaining diffs inspected: title alignment, Draft normalization and relocation/expansion of provenance citations only. All alternatives and 31/31 operator questions retained (7, 6, 5, 8, 5). |
| README | Held: 5/5 unique linked Draft rows with existing targets and titles equal to the respective H1. |
| Evidence provenance | Held: 10/10 pinned repository-relative-labeled source/companion links; source Git blobs exist and match origin/main. Evidence remains in curator, with bounds and source locators preserved; no wholesale companion copying. |
| Architecture/process | Held: publication fits CIP-0001's proposal/adoption distinction. Proposed normative sketches remain inside Draft CIPs; normative artifacts and runtime behavior are untouched. |
| Formatting/rendering | Held: complete scan across the six changed Markdown files: 31 tables / 205 body rows, 0 inconsistent widths. GFM renderer and narrowing probe confirm the four repaired values. Whitespace checks pass. |
| Public-data hygiene | Held within stated grep bounds: personal paths 0 matching lines, private hosts 0, machine user 0, employer labels 0 across all six documents. Manual inspection of publication deltas finds no personal data. |
| Links | Held: repository local-link validator passes; 5/5 README targets exist; 24/24 distinct external Markdown URLs return HTTP 200 after redirects, including 10/10 research URLs. |
| Tests/docs gates | Held: reviewer reran validator, Skillfile independence, formatting/rendering, source/template/index/candidate checks, privacy grep and external HTTP checks. Attached Python-suite results accepted as existing evidence; details and limits below. |

No configured `surface-table.json`/`.md` precondition is attached; this sweep merges the task's review instructions, AC and prior verdict surfaces. Free hunt found no additional defect.

## Merged task checklist

- [x] Five CIP-0002–0006 files in TEMPLATE.md shape, Status Draft.
- [x] README index lists and links all five Draft CIPs.
- [x] Evidence linked by repository-relative path labels and full commit; not copied.
- [x] Scope restricted to cips/; no normative/schema/CHANGELOG/LOGBOOK edits or personal-path/host/employer findings.
- [x] Applicable docs/link/format checks run with exit codes recorded.
- [x] Requested documentation implementation matches task description and AC.
- [x] New task-scoped outcome: this revision-2 verdict, attached before acceptance.
- [x] Review decisions and prior-finding closure persisted on the board in this artifact; no repository LOGBOOK edit, per explicit task prohibition.
- [x] Implementation matches all acceptance criteria.
- [x] Solution fits the established CIP process and architecture.
- [x] Applicable checks green; existing Python-suite evidence accepted as specified below.
- [x] Explicit accepted verdict recorded with accept_cr; no reviewing or done terminal state from this reviewer.

## Reviewer-executed verification and exit codes

Python validation used the existing project development virtual environment, with `-B` to avoid bytecode writes.

| Command/probe | Exit and measurement |
|---|---|
| `python -B tools/validate.py` | 0; validated 73 schemas and 1,294 vector files, including repository-local Markdown links and frozen/released immutability. |
| `python -B tools/verify_skillfile_sources_independence.py` | 0; 89/89 schema refs byte-identical, 28/28 cited clauses present, documented conditional exception allowlisted. |
| `git diff --check <base> <candidate>` | 0; clean. |
| Source/template/index/candidate comparison via Python/Git | 0; sections 35/35, questions 31/31, research blobs 10/10, candidate bytes 6/6, index rows 5/5. |
| Full changed-doc table width scan and markdown-it-py `gfm-like` render | 0; 31 tables / 205 body rows, inconsistent 0; repaired semantics 4/4; narrowing mutant drops 1/4, all-unescaped mutant drops 4/4. |
| `grep -nE` privacy patterns, each across all six files | Each exit 1 (no matches): personal paths 0, private hosts 0, machine user 0, employer labels 0. |
| `grep -nE '[[:blank:]]+$'` across all six files | Exit 1; 0 matching lines. |
| Per-URL `curl --location --silent --show-error --output /dev/null --write-out '%{http_code}' --max-time 30 <url>` | All 24 commands exit 0, HTTP 200; research 10/10. |
| Final candidate SHA-256/blob comparison | 0; files still match accepted candidate 6/6. |

Privacy regexes:

```text
personal-paths: /Users/[^[:space:]`]+|/home/[^[:space:]`]+|[A-Za-z]:[\\/](Users|Documents and Settings)[\\/]|~[A-Za-z][A-Za-z0-9_.-]*/
private-hosts: ([A-Za-z0-9_-]+\.)+(local|lan|internal|corp)([[:space:]"`/:]|$)|relux-(entrypoint|host)
machine-user: \badministrator\b
employer-labels: \bemployer\b|\bReluxWorks\b
```

These patterns are bounded disclosure detectors, not universal named-entity detection. Template Owner handle, public vendor/repository domains, standard file names and illustrative reverse-domain service IDs are appropriate public context.

## Accepted prior evidence and limits

No Python implementation or validation input outside the six proposal docs changed since the green revision-1 review. Accepted the independently executed revision-1 Python suite (672/672, exit 0, 487.398 s) and the revision-2 attached validation log (672/672, exit 0, 437.632 s). This run did not rerun the full Python suite. Revision-2 log additionally contains a passing Go generator shard; this reviewer did not execute or independently adjudicate Go tests, per the explicit instruction, and does not claim a full make validate run. Producer table probes were inspected, but their counters were replaced by the independent reviewer scan/render measurements above.

No lychee execution, cross-platform CI result, native-provider launch, proposed-feature behavioral qualification, or comprehensive external-fragment/content attestation is claimed. curl checks establish current HTTP reachability; the Git comparison establishes the pinned research content. Proposal source claims are preserved, not re-researched or treated as implemented guarantees.

## Machine-readable verdict

```verdict-findings
{
  "findings": [],
  "notes": [],
  "surface_results": [
    {"row": "Scope and frozen artifacts", "result": "held"},
    {"row": "Template and status", "result": "held"},
    {"row": "Source fidelity", "result": "held"},
    {"row": "README", "result": "held"},
    {"row": "Evidence provenance", "result": "held"},
    {"row": "Architecture/process", "result": "held"},
    {"row": "Formatting/rendering", "result": "held"},
    {"row": "Public-data hygiene", "result": "held"},
    {"row": "Links", "result": "held"},
    {"row": "Tests/docs gates", "result": "held"}
  ],
  "free_hunt": []
}
```
